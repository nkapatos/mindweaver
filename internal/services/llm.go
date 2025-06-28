package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/adapters"
	"github.com/nkapatos/mindweaver/internal/store"
)

// LLMService handles LLM service operations
type LLMService struct {
	store      store.Querier
	modelCache *modelCacheService
	logger     *slog.Logger
}

// NewLLMService creates a new LLM service
func NewLLMService(store store.Querier) *LLMService {
	return &LLMService{
		store:      store,
		modelCache: newModelCacheService(store),
		logger:     slog.Default(),
	}
}

// CreateLLMService creates a new LLM service (connection info only)
func (s *LLMService) CreateLLMService(ctx context.Context, name, description, adapter, apiKey, baseURL, organization string) (*store.LlmService, error) {
	s.logger.Info("Creating LLM service", "name", name, "adapter", adapter)

	params := store.CreateLLMServiceParams{
		Name:         name,
		Description:  sql.NullString{String: description, Valid: description != ""},
		Adapter:      adapter,
		ApiKey:       apiKey,
		BaseUrl:      baseURL,
		Organization: sql.NullString{String: organization, Valid: organization != ""},
	}

	llmService, err := s.store.CreateLLMService(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create LLM service", "error", err, "name", name)
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	s.logger.Info("LLM service created successfully", "id", llmService.ID, "name", name)
	return &llmService, nil
}

// CreateLLMServiceConfig creates a new configuration for an LLM service
func (s *LLMService) CreateLLMServiceConfig(ctx context.Context, llmServiceID int64, name, description string, config *LLMConfiguration) (*store.LlmServiceConfig, error) {
	s.logger.Info("Creating LLM service config", "service_id", llmServiceID, "name", name)

	// Validate configuration
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if err := config.Validate(); err != nil {
		s.logger.Error("Invalid LLM configuration", "error", err, "service_id", llmServiceID)
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Convert configuration to JSON for storage
	configJSON, err := config.ToJSON()
	if err != nil {
		s.logger.Error("Failed to serialize configuration", "error", err, "service_id", llmServiceID)
		return nil, fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Log configuration (without sensitive data)
	config.LogConfiguration(s.logger)

	params := store.CreateLLMServiceConfigParams{
		LlmServiceID:  llmServiceID,
		Name:          name,
		Description:   sql.NullString{String: description, Valid: description != ""},
		Configuration: configJSON,
	}

	llmServiceConfig, err := s.store.CreateLLMServiceConfig(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create LLM service config", "error", err, "service_id", llmServiceID)
		return nil, fmt.Errorf("failed to create LLM service config: %w", err)
	}

	s.logger.Info("LLM service config created successfully", "id", llmServiceConfig.ID, "name", name)
	return &llmServiceConfig, nil
}

// CreateLLMServiceWithConfig creates a new LLM service with a default configuration
func (s *LLMService) CreateLLMServiceWithConfig(ctx context.Context, name, description, adapter, apiKey, baseURL, organization, configName, configDescription, model string) (*store.LlmService, *store.LlmServiceConfig, error) {
	// First create the LLM service
	llmService, err := s.CreateLLMService(ctx, name, description, adapter, apiKey, baseURL, organization)
	if err != nil {
		return nil, nil, err
	}

	// Create default configuration
	config := DefaultConfiguration(model)

	// Then create the configuration
	llmServiceConfig, err := s.CreateLLMServiceConfig(ctx, llmService.ID, configName, configDescription, config)
	if err != nil {
		return nil, nil, err
	}

	return llmService, llmServiceConfig, nil
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

// GetLLMServiceConfigsByServiceID retrieves all configurations for an LLM service
func (s *LLMService) GetLLMServiceConfigsByServiceID(ctx context.Context, llmServiceID int64) ([]store.LlmServiceConfig, error) {
	s.logger.Info("Getting LLM service configs by service ID", "service_id", llmServiceID)

	configs, err := s.store.GetLLMServiceConfigsByServiceID(ctx, llmServiceID)
	if err != nil {
		s.logger.Error("Failed to get LLM service configs", "error", err, "service_id", llmServiceID)
		return nil, fmt.Errorf("failed to get LLM service configs: %w", err)
	}

	return configs, nil
}

// GetLLMServiceConfigByID retrieves a specific configuration by ID
func (s *LLMService) GetLLMServiceConfigByID(ctx context.Context, id int64) (*store.LlmServiceConfig, error) {
	s.logger.Info("Getting LLM service config by ID", "id", id)

	config, err := s.store.GetLLMServiceConfigByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get LLM service config by ID", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get LLM service config: %w", err)
	}

	return &config, nil
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
func (s *LLMService) UpdateLLMService(ctx context.Context, id int64, name, description, adapter, apiKey, baseURL, organization string) error {
	s.logger.Info("Updating LLM service", "id", id, "name", name)

	params := store.UpdateLLMServiceParams{
		ID:           id,
		Name:         name,
		Description:  sql.NullString{String: description, Valid: description != ""},
		Adapter:      adapter,
		ApiKey:       apiKey,
		BaseUrl:      baseURL,
		Organization: sql.NullString{String: organization, Valid: organization != ""},
	}

	err := s.store.UpdateLLMService(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update LLM service", "error", err, "id", id)
		return fmt.Errorf("failed to update LLM service: %w", err)
	}

	s.logger.Info("LLM service updated successfully", "id", id)
	return nil
}

// UpdateLLMServiceConfig updates an existing LLM service configuration
func (s *LLMService) UpdateLLMServiceConfig(ctx context.Context, id int64, name, description string, config *LLMConfiguration) error {
	s.logger.Info("Updating LLM service config", "id", id, "name", name)

	// Validate configuration
	if config == nil {
		return fmt.Errorf("configuration is required")
	}

	if err := config.Validate(); err != nil {
		s.logger.Error("Invalid LLM configuration", "error", err, "id", id)
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Convert configuration to JSON for storage
	configJSON, err := config.ToJSON()
	if err != nil {
		s.logger.Error("Failed to serialize configuration", "error", err, "id", id)
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Log configuration (without sensitive data)
	config.LogConfiguration(s.logger)

	params := store.UpdateLLMServiceConfigParams{
		ID:            id,
		Name:          name,
		Description:   sql.NullString{String: description, Valid: description != ""},
		Configuration: configJSON,
	}

	err = s.store.UpdateLLMServiceConfig(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update LLM service config", "error", err, "id", id)
		return fmt.Errorf("failed to update LLM service config: %w", err)
	}

	s.logger.Info("LLM service config updated successfully", "id", id)
	return nil
}

// DeleteLLMService deletes an LLM service (and all its configurations via CASCADE)
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

// DeleteLLMServiceConfig deletes a specific LLM service configuration
func (s *LLMService) DeleteLLMServiceConfig(ctx context.Context, id int64) error {
	s.logger.Info("Deleting LLM service config", "id", id)

	err := s.store.DeleteLLMServiceConfig(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete LLM service config", "error", err, "id", id)
		return fmt.Errorf("failed to delete LLM service config: %w", err)
	}

	s.logger.Info("LLM service config deleted successfully", "id", id)
	return nil
}

// GetAvailableModels gets available models for a service, using cache when possible
func (s *LLMService) GetAvailableModels(ctx context.Context, adapter, apiKey, baseURL string) ([]adapters.Model, error) {
	// For now, fall back to direct API call
	// TODO: Implement proper service lookup and caching
	adapterInstance, err := adapters.NewAdapter(adapters.AdapterConfig{
		Name:    adapter,
		BaseURL: baseURL,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, err
	}

	return adapterInstance.ListModels(ctx, apiKey, baseURL)
}

// GetAvailableModelsForService gets available models for a specific LLM service ID
// This method handles caching internally - the caller doesn't need to know about caching
func (s *LLMService) GetAvailableModelsForService(ctx context.Context, llmServiceID int64) ([]adapters.Model, error) {
	return s.modelCache.GetCachedModels(ctx, llmServiceID, false)
}

// RefreshModelsForService forces a refresh of models for a specific service
// This is useful for admin operations or when cache is known to be stale
func (s *LLMService) RefreshModelsForService(ctx context.Context, llmServiceID int64) ([]adapters.Model, error) {
	return s.modelCache.GetCachedModels(ctx, llmServiceID, true)
}
