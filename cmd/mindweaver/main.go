package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/templ"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/pages"
)

var db *sql.DB
var logger *slog.Logger

func init() {
	// Setup structured logging
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	var err error
	db, err = sql.Open("sqlite3", "file:mw.db?cache=shared&mode=rwc")
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		os.Exit(1)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	logger.Info("Database connection established successfully")
}

func main() {
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()

	// Initialize dependencies
	querier := store.New(db)
	userService := services.NewUserService(querier)
	promptService := services.NewPromptService(querier)
	userHandler := api.NewUserHandler(userService)
	promptHandler := api.NewPromptHandler(promptService)

	logger.Info("Application dependencies initialized")

	// Setup routes
	http.Handle("/", templ.Handler(pages.Home()))

	// API routes
	http.HandleFunc("/api/users", userHandler.CreateUser)
	http.HandleFunc("/api/prompts", promptHandler.CreatePrompt)

	// Static assets
	fs := http.FileServer(http.Dir("dist"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	logger.Info("Starting HTTP server", "port", "8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
