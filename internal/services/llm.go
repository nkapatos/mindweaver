package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/adapters"
	"github.com/nkapatos/mindweaver/internal/store"
)

type LLMService struct {
	providerStore store.Querier
	logger        *slog.Logger
	// Cache for active LLM providers
	providers map[string]adapters.LLMProvider
}

// NewLLMService creates a new LLM service
func NewLLMService(providerStore store.Querier) *LLMService {
	return &LLMService{
		providerStore: providerStore,
		logger:        slog.Default(),
		providers:     make(map[string]adapters.LLMProvider),
	}
}

// int64ToNullInt64 converts int64 to sql.NullInt64
func int64ToNullInt64(id int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: id,
		Valid: true,
	}
}

// getOrCreateProvider gets an LLM provider from cache or creates a new one
func (s *LLMService) getOrCreateProvider(providerName string) (adapters.LLMProvider, error) {
	// Check cache first
	if provider, exists := s.providers[providerName]; exists {
		return provider, nil
	}

	// Get provider from database
	provider, err := s.providerStore.GetProviderByName(context.Background(), providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Get provider settings
	settings, err := s.providerStore.GetProviderSettings(context.Background(), int64ToNullInt64(provider.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get provider settings: %w", err)
	}

	// Convert settings to map
	settingsMap := make(map[string]string)
	for _, setting := range settings {
		settingsMap[setting.SettingKey] = setting.SettingValue
	}

	// Create adapter config
	config := adapters.AdapterConfig{
		Name:     provider.Name,
		Settings: settingsMap,
	}

	// Extract common settings
	if apiKey, exists := settingsMap["api_key"]; exists {
		config.APIKey = apiKey
	}
	if baseURL, exists := settingsMap["base_url"]; exists {
		config.BaseURL = baseURL
	}

	// Create adapter
	adapter, err := adapters.NewAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Cache the provider
	s.providers[providerName] = adapter

	return adapter, nil
}

// Generate handles text generation requests
func (s *LLMService) Generate(ctx echo.Context, prompt string, options adapters.GenerateOptions) (*adapters.GenerateResponse, error) {
	s.logger.Info("Generating text", "prompt_length", len(prompt))

	// For now, use a default provider - this can be enhanced to select based on options
	providerName := "openai"
	if options.Model != "" {
		// Extract provider from model name (e.g., "gpt-4" -> "openai")
		// This is a simple heuristic and can be improved
		providerName = "openai"
	}

	provider, err := s.getOrCreateProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	response, err := provider.Generate(ctx.Request().Context(), prompt, options)
	if err != nil {
		s.logger.Error("Generation failed", "error", err)
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	s.logger.Info("Generation completed", "response_length", len(response.Content))
	return response, nil
}

// Chat handles chat completion requests
func (s *LLMService) Chat(ctx echo.Context, messages []adapters.ChatMessage, options adapters.ChatOptions) (*adapters.ChatResponse, error) {
	s.logger.Info("Processing chat", "message_count", len(messages))

	// For now, return an error since Chat method is not implemented yet
	// TODO: Implement Chat method in adapters
	return nil, fmt.Errorf("chat functionality not yet implemented")
}
