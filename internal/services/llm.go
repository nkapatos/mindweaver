package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/adapters"
	"github.com/nkapatos/mindweaver/internal/store"
)

// LLMService handles LLM service operations using the unified configuration system
// This service wraps adapters with metadata and provides configuration management
type LLMService struct {
	store           store.Querier
	modelCache      *modelCacheService
	providerService *ProviderService
	logger          *slog.Logger
}

// NewLLMService creates a new LLM service
func NewLLMService(store store.Querier, providerService *ProviderService) *LLMService {
	return &LLMService{
		store:           store,
		modelCache:      newModelCacheService(store),
		providerService: providerService,
		logger:          slog.Default(),
	}
}

// CreateLLMService creates a new LLM service (connection info only)
// This creates the internal wrapper around an adapter with metadata
func (s *LLMService) CreateLLMService(ctx context.Context, name, description, adapter, apiKey, baseURL, organization string, createdBy, updatedBy int64) (*store.LlmService, error) {
	s.logger.Info("Creating LLM service", "name", name, "adapter", adapter, "created_by", createdBy, "updated_by", updatedBy)

	// Validate that the adapter is supported
	if !s.isAdapterSupported(adapter) {
		return nil, fmt.Errorf("unsupported adapter: %s", adapter)
	}

	params := store.CreateLLMServiceParams{
		Name:         name,
		Description:  sql.NullString{String: description, Valid: description != ""},
		Adapter:      adapter,
		ApiKey:       apiKey,
		BaseUrl:      baseURL,
		Organization: sql.NullString{String: organization, Valid: organization != ""},
		CreatedBy:    createdBy,
		UpdatedBy:    updatedBy,
	}

	llmService, err := s.store.CreateLLMService(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create LLM service", "error", err, "name", name, "created_by", createdBy, "updated_by", updatedBy)
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	s.logger.Info("LLM service created successfully", "id", llmService.ID, "name", name)
	return &llmService, nil
}

// CreateLLMServiceConfig creates a new configuration for an LLM service using the unified config system
// The configuration is adapter-agnostic and validated at the service layer
func (s *LLMService) CreateLLMServiceConfig(ctx context.Context, llmServiceID int64, name, description string, config *LLMConfiguration, createdBy, updatedBy int64) (*store.LlmServiceConfig, error) {
	s.logger.Info("Creating LLM service config", "service_id", llmServiceID, "name", name, "created_by", createdBy)

	// Get the LLM service to determine the provider/adapter
	llmService, err := s.store.GetLLMServiceByID(ctx, llmServiceID)
	if err != nil {
		s.logger.Error("Failed to get LLM service", "service_id", llmServiceID, "error", err)
		return nil, fmt.Errorf("failed to get LLM service: %w", err)
	}

	// Set provider and service metadata
	config.Provider = llmService.Adapter
	config.ServiceID = llmServiceID

	// Validate configuration with provider-specific rules
	configMap := config.ToAdapterOptions()

	// Create adapter to validate configuration
	adapter, err := adapters.NewAdapter(llmService.Adapter, "", "")
	if err != nil {
		s.logger.Error("Failed to create adapter for validation", "error", err, "adapter", llmService.Adapter)
		return nil, fmt.Errorf("unsupported adapter: %s", llmService.Adapter)
	}

	// Use adapter's validation
	if err := adapter.ValidateConfig(configMap); err != nil {
		s.logger.Error("Configuration validation failed", "error", err, "service_id", llmServiceID)
		return nil, fmt.Errorf("configuration validation failed: %w", err)
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
		CreatedBy:     createdBy,
		UpdatedBy:     updatedBy,
	}

	llmServiceConfig, err := s.store.CreateLLMServiceConfig(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create LLM service config", "error", err, "service_id", llmServiceID)
		return nil, fmt.Errorf("failed to create LLM service config: %w", err)
	}

	// Set the config ID in the configuration
	config.ConfigID = llmServiceConfig.ID

	s.logger.Info("LLM service config created successfully", "id", llmServiceConfig.ID, "name", name)
	return &llmServiceConfig, nil
}

// CreateLLMServiceWithConfig creates a new LLM service with a default configuration
// This demonstrates the adapter providing default configuration for a model
func (s *LLMService) CreateLLMServiceWithConfig(ctx context.Context, name, description, adapter, apiKey, baseURL, organization, configName, configDescription, model string) (*store.LlmService, *store.LlmServiceConfig, error) {
	// First create the LLM service (adapter wrapper)
	llmService, err := s.CreateLLMService(ctx, name, description, adapter, apiKey, baseURL, organization, 0, 0)
	if err != nil {
		return nil, nil, err
	}

	// Get default configuration from the adapter for this model
	defaultConfig := s.getDefaultConfigFromAdapter(adapter, model)

	// Then create the configuration
	llmServiceConfig, err := s.CreateLLMServiceConfig(ctx, llmService.ID, configName, configDescription, defaultConfig, 0, 0)
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

// GetLLMServiceConfig retrieves and parses an LLM service configuration
func (s *LLMService) GetLLMServiceConfig(ctx context.Context, id int64) (*LLMConfiguration, error) {
	s.logger.Debug("Getting LLM service config", "id", id)

	llmServiceConfig, err := s.store.GetLLMServiceConfigByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get LLM service config", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get LLM service config: %w", err)
	}

	// Parse the configuration
	config, err := FromJSON(llmServiceConfig.Configuration)
	if err != nil {
		s.logger.Error("Failed to parse LLM service config", "id", id, "error", err)
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Set metadata
	config.ConfigID = llmServiceConfig.ID
	config.ServiceID = llmServiceConfig.LlmServiceID

	// Get provider from LLM service
	llmService, err := s.store.GetLLMServiceByID(ctx, llmServiceConfig.LlmServiceID)
	if err == nil {
		config.Provider = llmService.Adapter
	}

	return config, nil
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
func (s *LLMService) UpdateLLMService(ctx context.Context, id int64, name, description, adapter, apiKey, baseURL, organization string, updatedBy int64) error {
	s.logger.Info("Updating LLM service", "id", id, "name", name, "updated_by", updatedBy)

	// Validate that the adapter is supported
	if !s.isAdapterSupported(adapter) {
		return fmt.Errorf("unsupported adapter: %s", adapter)
	}

	params := store.UpdateLLMServiceParams{
		ID:           id,
		Name:         name,
		Description:  sql.NullString{String: description, Valid: description != ""},
		Adapter:      adapter,
		ApiKey:       apiKey,
		BaseUrl:      baseURL,
		Organization: sql.NullString{String: organization, Valid: organization != ""},
		UpdatedBy:    updatedBy,
	}

	err := s.store.UpdateLLMService(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update LLM service", "error", err, "id", id, "updated_by", updatedBy)
		return fmt.Errorf("failed to update LLM service: %w", err)
	}

	s.logger.Info("LLM service updated successfully", "id", id)
	return nil
}

// UpdateLLMServiceConfig updates an existing LLM service configuration using the unified config system
func (s *LLMService) UpdateLLMServiceConfig(ctx context.Context, id int64, name, description string, config *LLMConfiguration, updatedBy int64) error {
	s.logger.Info("Updating LLM service config", "id", id, "name", name, "updated_by", updatedBy)

	// Get the existing config to determine the provider
	existingConfig, err := s.store.GetLLMServiceConfigByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get existing LLM service config", "id", id, "error", err)
		return fmt.Errorf("failed to get existing config: %w", err)
	}

	// Get the LLM service to determine the provider
	llmService, err := s.store.GetLLMServiceByID(ctx, existingConfig.LlmServiceID)
	if err != nil {
		s.logger.Error("Failed to get LLM service", "service_id", existingConfig.LlmServiceID, "error", err)
		return fmt.Errorf("failed to get LLM service: %w", err)
	}

	// Set provider and service metadata
	config.Provider = llmService.Adapter
	config.ServiceID = llmService.ID
	config.ConfigID = id

	// Validate configuration with provider-specific rules
	configMap := config.ToAdapterOptions()

	// Create adapter to validate configuration
	adapter, err := adapters.NewAdapter(llmService.Adapter, "", "")
	if err != nil {
		s.logger.Error("Failed to create adapter for validation", "error", err, "adapter", llmService.Adapter)
		return fmt.Errorf("unsupported adapter: %s", llmService.Adapter)
	}

	// Use adapter's validation
	if err := adapter.ValidateConfig(configMap); err != nil {
		s.logger.Error("Configuration validation failed", "error", err, "service_id", existingConfig.LlmServiceID)
		return fmt.Errorf("configuration validation failed: %w", err)
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
		UpdatedBy:     updatedBy,
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

// DeleteLLMServiceConfig deletes an LLM service configuration
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

// GetAvailableModels retrieves available models from an adapter
// This demonstrates the adapter providing model information
func (s *LLMService) GetAvailableModels(ctx context.Context, adapter, apiKey, baseURL string) ([]adapters.Model, error) {
	s.logger.Info("Getting available models", "adapter", adapter)

	// Create adapter instance
	adapterInstance, err := adapters.NewAdapter(adapter, apiKey, baseURL)
	if err != nil {
		s.logger.Error("Failed to create adapter", "error", err, "adapter", adapter)
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Get models from adapter
	models, err := adapterInstance.GetModels(ctx)
	if err != nil {
		s.logger.Error("Failed to get models from adapter", "error", err, "adapter", adapter)
		return nil, fmt.Errorf("failed to get models: %w", err)
	}

	s.logger.Info("Retrieved models from adapter", "adapter", adapter, "count", len(models))
	return models, nil
}

// GetAvailableModelsForService retrieves available models for a specific LLM service
func (s *LLMService) GetAvailableModelsForService(ctx context.Context, llmServiceID int64) ([]adapters.Model, error) {
	s.logger.Info("Getting available models for service", "service_id", llmServiceID)

	// Get the LLM service
	llmService, err := s.GetLLMServiceByID(ctx, llmServiceID)
	if err != nil {
		return nil, err
	}

	// Get models from the adapter
	return s.GetAvailableModels(ctx, llmService.Adapter, llmService.ApiKey, llmService.BaseUrl)
}

// RefreshModelsForService refreshes the model cache for a specific LLM service
func (s *LLMService) RefreshModelsForService(ctx context.Context, llmServiceID int64) ([]adapters.Model, error) {
	s.logger.Info("Refreshing models for service", "service_id", llmServiceID)

	// Use the model cache service to refresh models
	return s.modelCache.RefreshModels(ctx, llmServiceID)
}

// GetCachedModels retrieves cached models for a service
func (s *LLMService) GetCachedModels(ctx context.Context, llmServiceID int64, forceRefresh bool) ([]adapters.Model, error) {
	s.logger.Info("Getting cached models for service", "service_id", llmServiceID, "force_refresh", forceRefresh)

	return s.modelCache.GetCachedModels(ctx, llmServiceID, forceRefresh)
}

// GetModelByServiceAndModelID retrieves a specific model by service and model ID
func (s *LLMService) GetModelByServiceAndModelID(ctx context.Context, llmServiceID int64, modelID string) (store.Model, error) {
	s.logger.Info("Getting model by service and model ID", "service_id", llmServiceID, "model_id", modelID)

	return s.modelCache.GetModelByServiceAndModelID(ctx, llmServiceID, modelID)
}

// GenerateResponse generates a response using a provider
// This demonstrates the unified interface regardless of underlying provider
func (s *LLMService) GenerateResponse(ctx context.Context, providerID int64, prompt string, systemPrompt string) (string, error) {
	s.logger.Info("Generating response", "provider_id", providerID)

	// Get provider with relations
	provider, llmServiceConfig, llmService, _, err := s.providerService.GetProviderWithRelations(ctx, providerID)
	if err != nil {
		s.logger.Error("Failed to get provider with relations", "error", err, "provider_id", providerID)
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Get configuration
	config, err := s.GetLLMServiceConfig(ctx, llmServiceConfig.ID)
	if err != nil {
		s.logger.Error("Failed to get configuration", "error", err, "config_id", llmServiceConfig.ID)
		return "", fmt.Errorf("failed to get configuration: %w", err)
	}

	// Create adapter
	adapter, err := adapters.NewAdapter(llmService.Adapter, llmService.ApiKey, llmService.BaseUrl)
	if err != nil {
		s.logger.Error("Failed to create adapter", "error", err, "adapter", llmService.Adapter)
		return "", fmt.Errorf("failed to create adapter: %w", err)
	}

	// Convert LLMConfiguration to map for adapter
	configMap := config.ToAdapterOptions()

	// Generate response using unified interface
	response, err := adapters.GenerateResponseWithConfig(ctx, adapter, prompt, systemPrompt, configMap)
	if err != nil {
		s.logger.Error("Failed to generate response", "error", err, "provider_id", providerID)
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	s.logger.Info("Response generated successfully", "provider_id", providerID, "provider_name", provider.Name)
	return response, nil
}

// ValidateConfiguration validates a configuration for a specific provider
func (s *LLMService) ValidateConfiguration(provider string, config *LLMConfiguration) ValidationResult {
	// Convert to map for adapter validation
	configMap := config.ToAdapterOptions()

	// Create adapter to validate configuration
	adapter, err := adapters.NewAdapter(provider, "", "")
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("unsupported provider: %s", provider)},
		}
	}

	// Use adapter's validation
	if err := adapter.ValidateConfig(configMap); err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{err.Error()},
		}
	}

	return ValidationResult{Valid: true}
}

// GetSupportedModels returns supported models for a provider
func (s *LLMService) GetSupportedModels(provider string) []string {
	// TODO: Implement this when adapters provide model lists
	return []string{}
}

// GetDefaultConfig returns a default configuration for a provider and model
func (s *LLMService) GetDefaultConfig(provider, model string) *LLMConfiguration {
	// Create adapter to get default config
	adapter, err := adapters.NewAdapter(provider, "", "")
	if err != nil {
		return DefaultConfiguration(model)
	}

	// Get default config from adapter
	defaultConfigMap := adapter.GetDefaultConfig(model)

	// Convert map to LLMConfiguration
	config := &LLMConfiguration{
		Model: model,
	}

	if temp, exists := defaultConfigMap["temperature"]; exists {
		if tempFloat, ok := temp.(float64); ok {
			config.Temperature = &tempFloat
		}
	}

	if maxTokens, exists := defaultConfigMap["max_tokens"]; exists {
		if maxInt, ok := maxTokens.(int); ok {
			config.MaxTokens = &maxInt
		}
	}

	if topP, exists := defaultConfigMap["top_p"]; exists {
		if topPFloat, ok := topP.(float64); ok {
			config.TopP = &topPFloat
		}
	}

	return config
}

// isAdapterSupported checks if an adapter is supported
func (s *LLMService) isAdapterSupported(adapter string) bool {
	supportedAdapters := []string{"openai", "openrouter", "ollama"}
	for _, supported := range supportedAdapters {
		if adapter == supported {
			return true
		}
	}
	return false
}

// getDefaultConfigFromAdapter gets default configuration from adapter for a model
func (s *LLMService) getDefaultConfigFromAdapter(adapter, model string) *LLMConfiguration {
	// Get default config from adapter
	return s.GetDefaultConfig(adapter, model)
}

// GetSupportedAdapters returns a list of available adapters for the template
func (s *LLMService) GetSupportedAdapters() []string {
	return []string{
		"openai",
		"openrouter",
		"ollama",
	}
}

// TestLLMServiceConnection tests the connection to an LLM service
func (s *LLMService) TestLLMServiceConnection(ctx context.Context, adapter, apiKey, baseURL string) error {
	s.logger.Info("Testing LLM service connection", "adapter", adapter, "base_url", baseURL)

	// Create adapter instance
	adapterInstance, err := adapters.NewAdapter(adapter, apiKey, baseURL)
	if err != nil {
		s.logger.Error("Failed to create adapter for testing", "adapter", adapter, "error", err)
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	// Try to get models to test the connection
	models, err := adapterInstance.GetModels(ctx)
	if err != nil {
		s.logger.Error("Failed to get models for connection test", "adapter", adapter, "error", err)
		return fmt.Errorf("connection test failed: %w", err)
	}

	s.logger.Info("LLM service connection test successful", "adapter", adapter, "models_count", len(models))
	return nil
}
