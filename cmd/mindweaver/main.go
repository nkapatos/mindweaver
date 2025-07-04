package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
	"github.com/nkapatos/mindweaver/internal/router"
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
	actorService := services.NewActorService(querier)
	promptService := services.NewPromptService(querier)
	providerService := services.NewProviderService(querier)
	llmService := services.NewLLMService(querier)
	conversationService := services.NewConversationService(querier)
	// API handlers (only the ones that work with our services)
	actorHandler := api.NewActorHandler(actorService)
	promptHandler := api.NewPromptHandler(promptService)

	// Web handlers (our main focus)
	homeHandler := web.NewHomeHandler()
	notFoundHandler := web.NewNotFoundHandler()
	promptsHandler := web.NewPromptsHandler(promptService)
	providersHandler := web.NewProvidersHandler(providerService, llmService, promptService)
	llmServicesHandler := web.NewLLMServicesHandler(llmService)
	llmServiceConfigsHandler := web.NewLLMServiceConfigsHandler(llmService)
	settingsHandler := web.NewSettingsHandler()
	conversationHandler := web.NewConversationHandler(conversationService, providerService)

	logger.Info("Application dependencies initialized")

	// Initialize router
	router := router.New()

	// Setup all routes
	router.SetupRoutes(
		actorHandler,
		promptHandler,
		nil, // No LLM API handler for now - focus on web handlers
		homeHandler,
		promptsHandler,
		providersHandler,
		llmServicesHandler,
		llmServiceConfigsHandler,
		settingsHandler,
		conversationHandler,
		notFoundHandler,
	)

	logger.Info("Starting Echo server", "port", "8080")
	if err := router.Echo().Start(":8080"); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
