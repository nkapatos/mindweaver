package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
	"github.com/nkapatos/mindweaver/internal/router"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
)

// ActorAuthMetadata represents authentication information stored in actor metadata
//
// Supported strategies (for future struct streamlining):
//   - "system":
//     Credentials: { "role": "system", "permissions": "all" }
//   - "password":
//     Credentials: { "username": "...", "password": "..." }
//
// Additional strategies (e.g., "oauth", "api_key") can be added as needed.
type ActorAuthMetadata struct {
	AuthStrategy string            `json:"auth_strategy"` // "password", "oauth", "api_key", etc.
	Credentials  map[string]string `json:"credentials"`   // Strategy-specific credentials
	LastLogin    *string           `json:"last_login,omitempty"`
	IsActive     bool              `json:"is_active"`
}

var db *sql.DB
var logger *slog.Logger

func init() {
	// Setup structured logging
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Quick check: if the database file does not exist, create it (touch)
	dbPath := "mw.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Info("Database file does not exist, creating...", "file", dbPath)
		file, err := os.Create(dbPath)
		if err != nil {
			logger.Error("Failed to create database file", "error", err)
			os.Exit(1)
		}
		file.Close()
	}

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

	// --- SETUP FLOW ---
	// 1. On init, check if db exists in the path. If not, create it (handled in init()).
	// 2. Always check if the system actor exists. If not, create it.
	//    NOTE: The system actor's ID may not always be 1 (auto-increment).
	//    If the system actor is manually deleted, the next created actor will have a higher ID.
	//    In the future, consider a more robust way to always identify the system actor (e.g., by a unique field or flag).
	if err := initializeSystemActor(actorService); err != nil {
		logger.Error("Failed to initialize system actor", "error", err)
		os.Exit(1)
	}

	// 3. If there is at least one user actor, continue. If not, show setup wizard to create one.
	needsSetup := checkIfSetupNeeded()
	if needsSetup {
		logger.Info("Application setup required - no test user will be created")
	}
	// Note: Test user creation has been removed. Users are now created through the setup wizard.

	// Initialize auth service
	authService := services.NewAuthService(actorService)

	promptService := services.NewPromptService(querier)
	providerService := services.NewProviderService(querier)
	llmService := services.NewLLMService(querier, providerService)
	conversationService := services.NewConversationService(querier)
	messageService := services.NewMessageService(querier)

	// API handlers (only the ones that work with our services)
	actorHandler := api.NewActorHandler(actorService)
	promptHandler := api.NewPromptHandler(promptService)
	providerHandler := api.NewProvidersHandler(providerService)
	llmServiceHandler := api.NewLLMServicesHandler(llmService)
	llmServiceConfigHandler := api.NewLLMServiceConfigsHandler(llmService)
	modelsHandler := api.NewModelsHandler(llmService)
	conversationHandler := api.NewConversationHandler(conversationService, messageService, providerService, llmService)

	// Web handlers (our main focus)
	authHandler := web.NewAuthHandler(authService)
	homeHandler := web.NewHomeHandler()
	notFoundHandler := web.NewNotFoundHandler()
	promptsHandler := web.NewPromptsHandler(promptService)
	providersHandler := web.NewProvidersHandler(providerService, llmService, promptService)
	llmServicesHandler := web.NewLLMServicesHandler(llmService)
	llmServiceConfigsHandler := web.NewLLMServiceConfigsHandler(llmService)
	settingsHandler := web.NewSettingsHandler()
	webConversationHandler := web.NewConversationHandler(conversationService, providerService)
	setupHandler := web.NewSetupHandler(actorService)

	logger.Info("Application dependencies initialized")

	// Initialize router
	router := router.New()

	// Setup all routes
	router.SetupRoutes(
		authService,
		authHandler,
		actorHandler,
		promptHandler,
		nil, // No LLM API handler for now - focus on web handlers
		conversationHandler,
		providerHandler,
		llmServiceHandler,
		llmServiceConfigHandler,
		modelsHandler,
		homeHandler,
		promptsHandler,
		providersHandler,
		llmServicesHandler,
		llmServiceConfigsHandler,
		settingsHandler,
		webConversationHandler,
		notFoundHandler,
		setupHandler,
	)

	logger.Info("Starting Echo server", "port", "8080")
	if err := router.Echo().Start(":8080"); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

// initializeSystemActor creates the system actor if it doesn't exist
func initializeSystemActor(actorService *services.ActorService) error {
	logger.Info("Initializing system actor")

	// Try to get the system actor by name
	_, err := actorService.GetActorByName(context.Background(), "System", "system")
	if err != nil {
		// System actor doesn't exist, create it
		logger.Info("System actor not found, creating...")

		// Create system metadata
		systemMetadata := ActorAuthMetadata{
			AuthStrategy: "system",
			Credentials: map[string]string{
				"role":        "system",
				"permissions": "all",
			},
			IsActive: true,
		}

		// Serialize system metadata to JSON
		metadataJSON, err := json.Marshal(systemMetadata)
		if err != nil {
			return err
		}

		// For the first system actor, we'll use ID 1 as created_by and updated_by
		// This is a special case for the initial system actor
		err = actorService.CreateActor(
			context.Background(),
			"system",
			"System",
			"System",
			"",
			string(metadataJSON), // Store system info in metadata
			true,
			1, // created_by - will be updated after creation
			1, // updated_by - will be updated after creation
		)
		if err != nil {
			return err
		}

		logger.Info("System actor created successfully with metadata")
	} else {
		logger.Info("System actor already exists")
	}

	return nil
}

// checkIfSetupNeeded checks if the application needs initial setup
func checkIfSetupNeeded() bool {
	// Check if setup marker file exists
	if _, err := os.Stat("setup_completed"); err == nil {
		// Setup marker exists, setup is not needed
		return false
	}

	// Check if database file exists
	if _, err := os.Stat("mw.db"); os.IsNotExist(err) {
		// Database doesn't exist, setup is needed
		return true
	}

	// If database exists but no setup marker, setup is needed
	return true
}
