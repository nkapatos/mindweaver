package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
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
	homeHandler := web.NewHomeHandler()
	notFoundHandler := web.NewNotFoundHandler()

	logger.Info("Application dependencies initialized")

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", homeHandler.Home)

	// API routes
	e.POST("/api/users", func(c echo.Context) error {
		userHandler.CreateUser(c.Response().Writer, c.Request())
		return nil
	})
	e.POST("/api/prompts", func(c echo.Context) error {
		promptHandler.CreatePrompt(c.Response().Writer, c.Request())
		return nil
	})

	// Static assets
	e.Static("/assets", "dist")

	// 404 handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok && he.Code == 404 {
			notFoundHandler.NotFound(c)
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	logger.Info("Starting Echo server", "port", "8080")
	if err := e.Start(":8080"); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
