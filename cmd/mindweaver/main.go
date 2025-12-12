package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// brainadapters "github.com/nkapatos/mindweaver/internal/brain/adapters"
	// brainbootstrap "github.com/nkapatos/mindweaver/internal/brain/bootstrap"
	mindbootstrap "github.com/nkapatos/mindweaver/internal/mind/bootstrap"
	mindnotes "github.com/nkapatos/mindweaver/internal/mind/notes"
	mindscheduler "github.com/nkapatos/mindweaver/internal/mind/scheduler"
	"github.com/nkapatos/mindweaver/pkg/config"
	"github.com/nkapatos/mindweaver/pkg/logging"
	nvmwmw "github.com/nkapatos/mindweaver/pkg/middleware"
	"github.com/nkapatos/mindweaver/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Mindweaver - unified binary for Mind and Brain services
func main() {
	// Parse runtime mode flag
	mode := flag.String("mode", "combined", "Runtime mode: combined, mind, or brain")
	flag.Parse()

	// Validate mode and load config
	var cfg *config.Config
	var enableMind, enableBrain bool

	switch *mode {
	case "combined":
		cfg = config.LoadConfig(config.ModeCombined)
		enableMind = true
		enableBrain = true
	case "mind":
		cfg = config.LoadConfig(config.ModeStandalone)
		enableMind = true
		enableBrain = false
	case "brain":
		cfg = config.LoadConfig(config.ModeStandalone)
		enableMind = false
		enableBrain = true
	default:
		log.Fatalf("Invalid mode: %s (must be: combined, mind, or brain)", *mode)
	}

	// Setup structured logging with module context
	var logModule string
	switch *mode {
	case "combined":
		logModule = logging.ModuleNVMW
	case "mind":
		logModule = logging.ModuleMind
	case "brain":
		logModule = logging.ModuleBrain
	}
	logger := logging.NewModuleLogger(logModule, cfg.Logging.Level, cfg.Logging.Format)
	slog.SetDefault(logger)

	// Initialize ETag salt for hashed ETag generation
	utils.InitETagSalt(cfg.Security.ETagSalt)

	logger.Info("🎸 Starting Mindweaver", "mode", *mode)

	// Declare database connection variables
	var notesDB *sql.DB
	var assistantDB *sql.DB

	// Create Echo instance (needed for bootstrap)
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true

	// Global middleware (applied to all routes)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(nvmwmw.ErrorHandlerMiddleware)
	e.Use(nvmwmw.RequestIDMiddleware)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		services := ""
		if enableMind && enableBrain {
			services = "mind+brain"
		} else if enableMind {
			services = "mind"
		} else if enableBrain {
			services = "brain"
		}
		return c.JSON(200, map[string]string{
			"status":   "healthy",
			"mode":     *mode,
			"services": services,
		})
	})

	// Create /api group for all services
	api := e.Group("/api")

	// Initialize Mind service if needed
	var mindNotesService *mindnotes.NotesService
	if enableMind {
		db, notesSvc, err := mindbootstrap.Initialize(e, api, cfg.Mind.DBPath, logger)
		if err != nil {
			logger.Error("Failed to initialize mind service", "error", err)
			os.Exit(1)
		}
		notesDB = db
		mindNotesService = notesSvc
		defer func() {
			if err := notesDB.Close(); err != nil {
				logger.Error("Failed to close notes database", "error", err)
			}
		}()
	}

	// Initialize Brain service if needed
	// if enableBrain {
	// 	// Create Mind adapter using factory (automatically selects Local or Remote based on mode)
	// 	mindOps, err := brainadapters.NewMindOperations(cfg, mindNotesService, logger)
	// 	if err != nil {
	// 		logger.Error("Failed to initialize Mind adapter", "error", err)
	// 		os.Exit(1)
	// 	}
	// 	logger.Info("Brain initialized with Mind adapter", "adapter_type", mindOps.GetAdapterType())
	//
	// 	db, err := brainbootstrap.Initialize(api, cfg.Brain.DBPath, mindOps, logger)
	// 	if err != nil {
	// 		logger.Error("Failed to initialize brain service", "error", err)
	// 		os.Exit(1)
	// 	}
	// 	assistantDB = db
	// 	defer func() {
	// 		if err := assistantDB.Close(); err != nil {
	// 			logger.Error("Failed to close assistant database", "error", err)
	// 		}
	// 	}()
	// }

	// Goroutine to periodically checkpoint WAL files
	if notesDB != nil || assistantDB != nil {
		go func() {
			ticker := time.NewTicker(120 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if notesDB != nil {
					if _, err := notesDB.Exec("PRAGMA wal_checkpoint(FULL);"); err != nil {
						logger.Error("notes db wal checkpoint failed", "error", err)
					}
				}
				if assistantDB != nil {
					if _, err := assistantDB.Exec("PRAGMA wal_checkpoint(FULL);"); err != nil {
						logger.Error("assistant db wal checkpoint failed", "error", err)
					}
				}
			}
		}()
	}

	// Graceful shutdown - checkpoint WAL files before exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		logger.Info("Shutdown signal received, stopping services...")

		// Checkpoint databases
		logger.Info("Checkpointing databases...")
		if notesDB != nil {
			if _, err := notesDB.Exec("PRAGMA wal_checkpoint(FULL);"); err != nil {
				logger.Error("Failed to checkpoint notes DB on shutdown", "error", err)
			}
		}
		if assistantDB != nil {
			if _, err := assistantDB.Exec("PRAGMA wal_checkpoint(FULL);"); err != nil {
				logger.Error("Failed to checkpoint assistant DB on shutdown", "error", err)
			}
		}
		time.Sleep(200 * time.Millisecond) // Give kernel time to flush
		os.Exit(0)
	}()

	// Initialize scheduler (Mind → Brain sync) if both services enabled
	var changeScheduler *mindscheduler.ChangeAccumulator
	if enableMind && enableBrain && mindNotesService != nil {
		logger.Info("🔄 Initializing Mind→Brain scheduler")

		// Use localhost for combined mode
		port := cfg.GetCombinedPort()
		schedulerCfg := mindscheduler.Config{
			BrainURL:      fmt.Sprintf("http://localhost:%d", port),
			FlushInterval: 5 * time.Minute, // Batch changes every 5 minutes
			BatchSize:     100,             // Max 100 changes per batch
		}

		changeScheduler = mindscheduler.NewChangeAccumulator(schedulerCfg, logger)
		mindNotesService.SetScheduler(changeScheduler)
		changeScheduler.Start()

		logger.Info("✅ Scheduler started - Mind will sync changes to Brain")

		// Ensure scheduler stops on shutdown
		defer func() {
			logger.Info("Stopping scheduler...")
			if err := changeScheduler.Stop(); err != nil {
				logger.Error("Failed to stop scheduler", "error", err)
			}
		}()
	}

	// Start the server
	var port int
	switch *mode {
	case "combined":
		port = cfg.GetCombinedPort()
	case "mind":
		port = cfg.Mind.Port
	case "brain":
		port = cfg.Brain.Port
	}
	addr := fmt.Sprintf(":%d", port)

	logger.Info("🔥 Mindweaver is LIVE!",
		"address", addr,
		"mode", *mode)

	if err := e.Start(addr); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
