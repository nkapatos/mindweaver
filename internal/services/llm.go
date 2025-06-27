package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"log/slog"

// 	"github.com/labstack/echo/v4"
// 	"github.com/nkapatos/mindweaver/internal/adapters"
// 	"github.com/nkapatos/mindweaver/internal/store"
// )

// LLMService handles LLM service operations
type LLMService struct {
	store  store.Querier
	logger *slog.Logger
}

// NewLLMService creates a new LLM service
func NewLLMService(store store.Querier) *LLMService {
	return &LLMService{
		store:  store,
		logger: slog.Default(),
	}
}

// CreateLLMService creates a new LLM service
func (s *LLMService) CreateLLMService(ctx context.Context, name, description, adapter, apiKey, baseURL, organization, configuration string) (*store.LlmService, error) {
	s.logger.Info("Creating LLM service", "name", name, "adapter", adapter)

	params := store.CreateLLMServiceParams{
		Name:          name,
		Description:   sql.NullString{String: description, Valid: description != ""},
		Adapter:       adapter,
		ApiKey:        apiKey,
		BaseUrl:       baseURL,
		Organization:  sql.NullString{String: organization, Valid: organization != ""},
		Configuration: configuration,
	}

	llmService, err := s.store.CreateLLMService(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create LLM service", "error", err, "name", name)
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	s.logger.Info("LLM service created successfully", "id", llmService.ID, "name", name)
	return &llmService, nil
}

// GetLLMServiceByID retrieves an LLM service by ID
func (s *LLMService) GetLLMServiceByID(ctx context.Context, id int64) (*store.LlmService, error) {
	s.logger.Info("Getting LLM service by ID", "id", id)

	llmService, err := s.store.GetLLMServiceByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get LLM service by ID", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get LLM service: %w", err)
	}

	return &llmService, nil
}

// GetLLMServiceByName retrieves an LLM service by name
func (s *LLMService) GetLLMServiceByName(ctx context.Context, name string) (*store.LlmService, error) {
	s.logger.Info("Getting LLM service by name", "name", name)

	llmService, err := s.store.GetLLMServiceByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get LLM service by name", "error", err, "name", name)
		return nil, fmt.Errorf("failed to get LLM service: %w", err)
	}

	return &llmService, nil
}

// GetAllLLMServices retrieves all LLM services
func (s *LLMService) GetAllLLMServices(ctx context.Context) ([]store.LlmService, error) {
	s.logger.Info("Getting all LLM services")

	llmServices, err := s.store.GetAllLLMServices(ctx)
	if err != nil {
		s.logger.Error("Failed to get all LLM services", "error", err)
		return nil, fmt.Errorf("failed to get LLM services: %w", err)
	}

	s.logger.Info("Retrieved LLM services", "count", len(llmServices))
	return llmServices, nil
}

// UpdateLLMService updates an existing LLM service
func (s *LLMService) UpdateLLMService(ctx context.Context, id int64, name, description, adapter, apiKey, baseURL, organization, configuration string) error {
	s.logger.Info("Updating LLM service", "id", id, "name", name)

	params := store.UpdateLLMServiceParams{
		ID:            id,
		Name:          name,
		Description:   sql.NullString{String: description, Valid: description != ""},
		Adapter:       adapter,
		ApiKey:        apiKey,
		BaseUrl:       baseURL,
		Organization:  sql.NullString{String: organization, Valid: organization != ""},
		Configuration: configuration,
	}

	err := s.store.UpdateLLMService(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update LLM service", "error", err, "id", id)
		return fmt.Errorf("failed to update LLM service: %w", err)
	}

	s.logger.Info("LLM service updated successfully", "id", id)
	return nil
}

// DeleteLLMService deletes an LLM service
func (s *LLMService) DeleteLLMService(ctx context.Context, id int64) error {
	s.logger.Info("Deleting LLM service", "id", id)

	err := s.store.DeleteLLMService(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete LLM service", "error", err, "id", id)
		return fmt.Errorf("failed to delete LLM service: %w", err)
	}

	s.logger.Info("LLM service deleted successfully", "id", id)
	return nil
}

// type LLMService struct {
// 	providerStore store.Querier
// 	logger        *slog.Logger
// 	// Cache for active LLM providers
// 	providers map[string]adapters.LLMProvider
// }

// // NewLLMService creates a new LLM service
// func NewLLMService(providerStore store.Querier) *LLMService {
// 	return &LLMService{
// 		providerStore: providerStore,
// 		logger:        slog.Default(),
// 		providers:     make(map[string]adapters.LLMProvider),
// 	}
// }

// // int64ToNullInt64 converts int64 to sql.NullInt64
// func int64ToNullInt64(id int64) sql.NullInt64 {
// 	return sql.NullInt64{
// 		Int64: id,
// 		Valid: true,
// 	}
// }

// // getOrCreateProvider gets an LLM provider from cache or creates a new one
// func (s *LLMService) getOrCreateProvider(providerName string) (adapters.LLMProvider, error) {
// 	// Check cache first
// 	if provider, exists := s.providers[providerName]; exists {
// 		return provider, nil
// 	}

// 	// Get provider from database
// 	provider, err := s.providerStore.GetProviderByName(context.Background(), providerName)
// 	if err != nil {
// 		return nil, fmt.Errorf("provider not found: %w", err)
// 	}

// 	// Get provider settings
// 	settings, err := s.providerStore.GetProviderSettings(context.Background(), int64ToNullInt64(provider.ID))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get provider settings: %w", err)
// 	}

// 	// Convert settings to map
// 	settingsMap := make(map[string]string)
// 	for _, setting := range settings {
// 		settingsMap[setting.SettingKey] = setting.SettingValue
// 	}

// 	// Create adapter config
// 	config := adapters.AdapterConfig{
// 		Name:     provider.Name,
// 		Settings: settingsMap,
// 	}

// 	// Extract common settings
// 	if apiKey, exists := settingsMap["api_key"]; exists {
// 		config.APIKey = apiKey
// 	}
// 	if baseURL, exists := settingsMap["base_url"]; exists {
// 		config.BaseURL = baseURL
// 	}

// 	// Create adapter
// 	adapter, err := adapters.NewAdapter(config)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create adapter: %w", err)
// 	}

// 	// Cache the provider
// 	s.providers[providerName] = adapter

// 	return adapter, nil
// }

// // Generate handles text generation requests
// func (s *LLMService) Generate(ctx echo.Context, prompt string, options adapters.GenerateOptions) (*adapters.GenerateResponse, error) {
// 	s.logger.Info("Generating text", "prompt_length", len(prompt))

// 	// For now, use a default provider - this can be enhanced to select based on options
// 	providerName := "openai"
// 	if options.Model != "" {
// 		// Extract provider from model name (e.g., "gpt-4" -> "openai")
// 		// This is a simple heuristic and can be improved
// 		providerName = "openai"
// 	}

// 	provider, err := s.getOrCreateProvider(providerName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get provider: %w", err)
// 	}

// 	response, err := provider.Generate(ctx.Request().Context(), prompt, options)
// 	if err != nil {
// 		s.logger.Error("Generation failed", "error", err)
// 		return nil, fmt.Errorf("generation failed: %w", err)
// 	}

// 	s.logger.Info("Generation completed", "response_length", len(response.Content))
// 	return response, nil
// }

// // Chat handles chat completion requests
// func (s *LLMService) Chat(ctx echo.Context, messages []adapters.ChatMessage, options adapters.ChatOptions) (*adapters.ChatResponse, error) {
// 	s.logger.Info("Processing chat", "message_count", len(messages))

// 	// For now, return an error since Chat method is not implemented yet
// 	// TODO: Implement Chat method in adapters
// 	return nil, fmt.Errorf("chat functionality not yet implemented")
// }
