package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

// ProviderService handles provider operations
// Providers use LLM service configurations to provide a unified interface regardless of underlying provider
//
// ARCHITECTURE NOTE: This service layer follows a simple query + relationship pattern:
// - Basic CRUD operations use direct SQL queries via the store layer
// - Relationship methods load related entities through separate queries
// - This approach prioritizes maintainability and flexibility over raw performance
//
// FUTURE OPTIMIZATION STRATEGY:
// If performance becomes a bottleneck, consider these approaches in order of complexity:
// 1. SQL JOINs: Single queries with JOINs for frequently accessed relationships
// 2. Stored procedures: Complex business logic in database for heavy operations
// 3. Caching: Redis/memory caching for frequently accessed data
// 4. Batch loading: Load multiple entities with relations in one operation
// 5. Database indexing: Ensure proper indexes on foreign keys and frequently queried fields
type ProviderService struct {
	providerStore store.Querier
	logger        *slog.Logger
}

func NewProviderService(providerStore store.Querier) *ProviderService {
	return &ProviderService{
		providerStore: providerStore,
		logger:        slog.Default(),
	}
}

// CreateProvider creates a new provider
// Providers use LLM service configurations to provide a unified interface
func (s *ProviderService) CreateProvider(ctx context.Context, name, description string, llmServiceConfigID int64, systemPromptID *int64, createdBy, updatedBy int64) (*store.Provider, error) {
	s.logger.Info("Creating new provider",
		"name", name,
		"description", description,
		"llm_service_config_id", llmServiceConfigID,
		"system_prompt_id", systemPromptID,
		"created_by", createdBy,
		"updated_by", updatedBy)

	params := store.CreateProviderParams{
		Name:               name,
		Description:        sql.NullString{String: description, Valid: description != ""},
		LlmServiceConfigID: llmServiceConfigID,
		CreatedBy:          createdBy,
		UpdatedBy:          updatedBy,
	}

	// Handle optional system_prompt_id
	if systemPromptID != nil {
		params.SystemPromptID = sql.NullInt64{
			Int64: *systemPromptID,
			Valid: true,
		}
	}

	provider, err := s.providerStore.CreateProvider(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create provider",
			"name", name,
			"description", description,
			"llm_service_config_id", llmServiceConfigID,
			"system_prompt_id", systemPromptID,
			"created_by", createdBy,
			"updated_by", updatedBy,
			"error", err)
		return nil, err
	}

	s.logger.Info("Provider created successfully", "id", provider.ID, "name", name)
	return &provider, nil
}

// GetProviderByID retrieves a provider by its ID
func (s *ProviderService) GetProviderByID(ctx context.Context, id int64) (*store.Provider, error) {
	s.logger.Debug("Getting provider by ID", "id", id)

	provider, err := s.providerStore.GetProviderByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get provider by ID", "id", id, "error", err)
		return nil, err
	}

	s.logger.Debug("Provider retrieved successfully", "id", id, "name", provider.Name)
	return &provider, nil
}

// GetProviderByName retrieves a provider by its name
func (s *ProviderService) GetProviderByName(ctx context.Context, name string) (*store.Provider, error) {
	s.logger.Debug("Getting provider by name", "name", name)

	provider, err := s.providerStore.GetProviderByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get provider by name", "name", name, "error", err)
		return nil, err
	}

	s.logger.Debug("Provider retrieved successfully", "id", provider.ID, "name", name)
	return &provider, nil
}

// GetAllProviders retrieves all providers
func (s *ProviderService) GetAllProviders(ctx context.Context) ([]store.Provider, error) {
	s.logger.Debug("Getting all providers")

	providers, err := s.providerStore.GetAllProviders(ctx)
	if err != nil {
		s.logger.Error("Failed to get all providers", "error", err)
		return nil, err
	}

	s.logger.Debug("All providers retrieved successfully", "count", len(providers))
	return providers, nil
}

// GetProvidersByLLMServiceConfig retrieves all providers for a specific LLM service configuration
func (s *ProviderService) GetProvidersByLLMServiceConfig(ctx context.Context, llmServiceConfigID int64) ([]store.Provider, error) {
	s.logger.Debug("Getting providers by LLM service config", "llm_service_config_id", llmServiceConfigID)

	providers, err := s.providerStore.GetProvidersByLLMServiceConfig(ctx, llmServiceConfigID)
	if err != nil {
		s.logger.Error("Failed to get providers by LLM service config", "llm_service_config_id", llmServiceConfigID, "error", err)
		return nil, err
	}

	s.logger.Debug("Providers retrieved successfully", "llm_service_config_id", llmServiceConfigID, "count", len(providers))
	return providers, nil
}

// GetProvidersByLLMService retrieves all providers that use configurations from a specific LLM service
func (s *ProviderService) GetProvidersByLLMService(ctx context.Context, llmServiceID int64) ([]store.Provider, error) {
	s.logger.Debug("Getting providers by LLM service", "llm_service_id", llmServiceID)

	// Get all providers
	allProviders, err := s.providerStore.GetAllProviders(ctx)
	if err != nil {
		s.logger.Error("Failed to get all providers for LLM service filtering", "llm_service_id", llmServiceID, "error", err)
		return nil, err
	}

	// Filter providers that use configurations from this LLM service
	var filteredProviders []store.Provider
	for _, provider := range allProviders {
		// Get the LLM service config for this provider
		llmServiceConfig, err := s.providerStore.GetLLMServiceConfigByID(ctx, provider.LlmServiceConfigID)
		if err != nil {
			s.logger.Error("Failed to get LLM service config for provider", "provider_id", provider.ID, "llm_service_config_id", provider.LlmServiceConfigID, "error", err)
			continue
		}

		// Check if this config belongs to the target LLM service
		if llmServiceConfig.LlmServiceID == llmServiceID {
			filteredProviders = append(filteredProviders, provider)
		}
	}

	s.logger.Debug("Providers retrieved successfully", "llm_service_id", llmServiceID, "count", len(filteredProviders))
	return filteredProviders, nil
}

// GetProvidersBySystemPrompt retrieves all providers that use a specific system prompt
func (s *ProviderService) GetProvidersBySystemPrompt(ctx context.Context, systemPromptID int64) ([]store.Provider, error) {
	s.logger.Debug("Getting providers by system prompt", "system_prompt_id", systemPromptID)

	providers, err := s.providerStore.GetProvidersBySystemPrompt(ctx, sql.NullInt64{Int64: systemPromptID, Valid: true})
	if err != nil {
		s.logger.Error("Failed to get providers by system prompt", "system_prompt_id", systemPromptID, "error", err)
		return nil, err
	}

	s.logger.Debug("Providers retrieved successfully", "system_prompt_id", systemPromptID, "count", len(providers))
	return providers, nil
}

// GetProviderWithRelations retrieves a provider along with its related LLM service config, LLM service, and system prompt
// This demonstrates the provider using LLM service configuration for unified interface
func (s *ProviderService) GetProviderWithRelations(ctx context.Context, providerID int64) (*store.Provider, *store.LlmServiceConfig, *store.LlmService, *store.Prompt, error) {
	s.logger.Debug("Getting provider with relations", "provider_id", providerID)

	// Get the provider
	provider, err := s.providerStore.GetProviderByID(ctx, providerID)
	if err != nil {
		s.logger.Error("Failed to get provider", "provider_id", providerID, "error", err)
		return nil, nil, nil, nil, err
	}

	// Get the related LLM service config
	llmServiceConfig, err := s.providerStore.GetLLMServiceConfigByID(ctx, provider.LlmServiceConfigID)
	if err != nil {
		s.logger.Error("Failed to get LLM service config for provider", "provider_id", providerID, "llm_service_config_id", provider.LlmServiceConfigID, "error", err)
		return &provider, nil, nil, nil, err
	}

	// Get the related LLM service
	llmService, err := s.providerStore.GetLLMServiceByID(ctx, llmServiceConfig.LlmServiceID)
	if err != nil {
		s.logger.Error("Failed to get LLM service for provider", "provider_id", providerID, "llm_service_id", llmServiceConfig.LlmServiceID, "error", err)
		return &provider, &llmServiceConfig, nil, nil, err
	}

	// Get the related system prompt (if any)
	var systemPrompt *store.Prompt
	if provider.SystemPromptID.Valid {
		prompt, err := s.providerStore.GetPromptById(ctx, provider.SystemPromptID.Int64)
		if err != nil {
			s.logger.Error("Failed to get system prompt for provider", "provider_id", providerID, "system_prompt_id", provider.SystemPromptID.Int64, "error", err)
			// Don't fail the entire operation for missing system prompt
		} else {
			systemPrompt = &prompt
		}
	}

	s.logger.Debug("Provider with relations retrieved successfully", "provider_id", providerID)
	return &provider, &llmServiceConfig, &llmService, systemPrompt, nil
}

// UpdateProvider updates an existing provider
func (s *ProviderService) UpdateProvider(ctx context.Context, id int64, name, description string, llmServiceConfigID int64, systemPromptID *int64, updatedBy int64) error {
	s.logger.Info("Updating provider", "id", id, "name", name, "updated_by", updatedBy)

	params := store.UpdateProviderParams{
		ID:                 id,
		Name:               name,
		Description:        sql.NullString{String: description, Valid: description != ""},
		LlmServiceConfigID: llmServiceConfigID,
		UpdatedBy:          updatedBy,
	}

	// Handle optional system_prompt_id
	if systemPromptID != nil {
		params.SystemPromptID = sql.NullInt64{
			Int64: *systemPromptID,
			Valid: true,
		}
	} else {
		params.SystemPromptID = sql.NullInt64{Valid: false}
	}

	err := s.providerStore.UpdateProvider(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update provider", "error", err, "id", id, "updated_by", updatedBy)
		return err
	}

	s.logger.Info("Provider updated successfully", "id", id)
	return nil
}

// DeleteProvider deletes a provider
func (s *ProviderService) DeleteProvider(ctx context.Context, id int64) error {
	s.logger.Info("Deleting provider", "id", id)

	err := s.providerStore.DeleteProvider(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete provider", "error", err, "id", id)
		return err
	}

	s.logger.Info("Provider deleted successfully", "id", id)
	return nil
}

// GetAllProvidersWithRelations retrieves all providers with their related entities
func (s *ProviderService) GetAllProvidersWithRelations(ctx context.Context) ([]struct {
	Provider         store.Provider
	LLMServiceConfig store.LlmServiceConfig
	LLMService       store.LlmService
	SystemPrompt     *store.Prompt
}, error) {
	s.logger.Debug("Getting all providers with relations")

	providers, err := s.providerStore.GetAllProviders(ctx)
	if err != nil {
		s.logger.Error("Failed to get all providers", "error", err)
		return nil, err
	}

	var result []struct {
		Provider         store.Provider
		LLMServiceConfig store.LlmServiceConfig
		LLMService       store.LlmService
		SystemPrompt     *store.Prompt
	}

	for _, provider := range providers {
		providerWithRelations, _, _, _, err := s.GetProviderWithRelations(ctx, provider.ID)
		if err != nil {
			s.logger.Error("Failed to get provider with relations", "provider_id", provider.ID, "error", err)
			continue
		}

		// Get the related entities
		llmServiceConfig, err := s.providerStore.GetLLMServiceConfigByID(ctx, provider.LlmServiceConfigID)
		if err != nil {
			s.logger.Error("Failed to get LLM service config", "provider_id", provider.ID, "error", err)
			continue
		}

		llmService, err := s.providerStore.GetLLMServiceByID(ctx, llmServiceConfig.LlmServiceID)
		if err != nil {
			s.logger.Error("Failed to get LLM service", "provider_id", provider.ID, "error", err)
			continue
		}

		var systemPrompt *store.Prompt
		if provider.SystemPromptID.Valid {
			prompt, err := s.providerStore.GetPromptById(ctx, provider.SystemPromptID.Int64)
			if err == nil {
				systemPrompt = &prompt
			}
		}

		result = append(result, struct {
			Provider         store.Provider
			LLMServiceConfig store.LlmServiceConfig
			LLMService       store.LlmService
			SystemPrompt     *store.Prompt
		}{
			Provider:         *providerWithRelations,
			LLMServiceConfig: llmServiceConfig,
			LLMService:       llmService,
			SystemPrompt:     systemPrompt,
		})
	}

	s.logger.Debug("All providers with relations retrieved successfully", "count", len(result))
	return result, nil
}
