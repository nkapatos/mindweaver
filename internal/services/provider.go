package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

// ProviderService handles provider operations
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
func (s *ProviderService) CreateProvider(ctx context.Context, name, description string, llmServiceID int64, systemPromptID *int64) (*store.Provider, error) {
	s.logger.Info("Creating new provider",
		"name", name,
		"description", description,
		"llm_service_id", llmServiceID,
		"system_prompt_id", systemPromptID)

	params := store.CreateProviderParams{
		Name:         name,
		Description:  description,
		LlmServiceID: llmServiceID,
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
			"llm_service_id", llmServiceID,
			"system_prompt_id", systemPromptID,
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

// GetProvidersByLLMService retrieves all providers for a specific LLM service
func (s *ProviderService) GetProvidersByLLMService(ctx context.Context, llmServiceID int64) ([]store.Provider, error) {
	s.logger.Debug("Getting providers by LLM service", "llm_service_id", llmServiceID)

	providers, err := s.providerStore.GetProvidersByLLMService(ctx, llmServiceID)
	if err != nil {
		s.logger.Error("Failed to get providers by LLM service", "llm_service_id", llmServiceID, "error", err)
		return nil, err
	}

	s.logger.Debug("Providers retrieved successfully", "llm_service_id", llmServiceID, "count", len(providers))
	return providers, nil
}

// GetProvidersBySystemPrompt retrieves all providers that use a specific system prompt
func (s *ProviderService) GetProvidersBySystemPrompt(ctx context.Context, systemPromptID int64) ([]store.Provider, error) {
	s.logger.Debug("Getting providers by system prompt", "system_prompt_id", systemPromptID)

	// Get all providers and filter by system prompt ID
	allProviders, err := s.providerStore.GetAllProviders(ctx)
	if err != nil {
		s.logger.Error("Failed to get all providers for system prompt filtering", "system_prompt_id", systemPromptID, "error", err)
		return nil, err
	}

	// Filter providers that use the specified system prompt
	var filteredProviders []store.Provider
	for _, provider := range allProviders {
		if provider.SystemPromptID.Valid && provider.SystemPromptID.Int64 == systemPromptID {
			filteredProviders = append(filteredProviders, provider)
		}
	}

	s.logger.Debug("Providers retrieved successfully", "system_prompt_id", systemPromptID, "count", len(filteredProviders))
	return filteredProviders, nil
}

// GetProviderWithRelations retrieves a provider along with its related LLM service and system prompt
//
// FUTURE OPTIMIZATION: If this becomes a performance bottleneck, consider:
// 1. SQL JOIN approach: Single query with LEFT JOINs to llm_services and prompts
// 2. Stored procedure: Complex business logic in database for frequently accessed data
// 3. Caching: Cache provider relationships if they're accessed frequently
// 4. Batch loading: Load multiple providers with relations in one operation
func (s *ProviderService) GetProviderWithRelations(ctx context.Context, providerID int64) (*store.Provider, *store.LlmService, *store.Prompt, error) {
	s.logger.Debug("Getting provider with relations", "provider_id", providerID)

	// Get the provider
	provider, err := s.providerStore.GetProviderByID(ctx, providerID)
	if err != nil {
		s.logger.Error("Failed to get provider", "provider_id", providerID, "error", err)
		return nil, nil, nil, err
	}

	// Get the related LLM service
	llmService, err := s.providerStore.GetLLMServiceByID(ctx, provider.LlmServiceID)
	if err != nil {
		s.logger.Error("Failed to get LLM service for provider", "provider_id", providerID, "llm_service_id", provider.LlmServiceID, "error", err)
		return &provider, nil, nil, err
	}

	// Get the related system prompt (if any)
	var systemPrompt *store.Prompt
	if provider.SystemPromptID.Valid {
		prompt, err := s.providerStore.GetPromptById(ctx, provider.SystemPromptID.Int64)
		if err != nil {
			s.logger.Error("Failed to get system prompt for provider", "provider_id", providerID, "system_prompt_id", provider.SystemPromptID.Int64, "error", err)
			// Don't fail the entire operation if system prompt is missing
		} else {
			systemPrompt = &prompt
		}
	}

	s.logger.Debug("Provider with relations retrieved successfully", "provider_id", providerID)
	return &provider, &llmService, systemPrompt, nil
}

// UpdateProvider updates a provider by its ID
func (s *ProviderService) UpdateProvider(ctx context.Context, id int64, name, description string, llmServiceID int64, systemPromptID *int64) error {
	s.logger.Info("Updating provider",
		"id", id,
		"name", name,
		"description", description,
		"llm_service_id", llmServiceID,
		"system_prompt_id", systemPromptID)

	params := store.UpdateProviderParams{
		ID:           id,
		Name:         name,
		Description:  description,
		LlmServiceID: llmServiceID,
	}

	// Handle optional system_prompt_id
	if systemPromptID != nil {
		params.SystemPromptID = sql.NullInt64{
			Int64: *systemPromptID,
			Valid: true,
		}
	} else {
		params.SystemPromptID = sql.NullInt64{
			Valid: false,
		}
	}

	if err := s.providerStore.UpdateProvider(ctx, params); err != nil {
		s.logger.Error("Failed to update provider",
			"id", id,
			"name", name,
			"description", description,
			"llm_service_id", llmServiceID,
			"system_prompt_id", systemPromptID,
			"error", err)
		return err
	}

	s.logger.Info("Provider updated successfully", "id", id, "name", name)
	return nil
}

// DeleteProvider deletes a provider by its ID
func (s *ProviderService) DeleteProvider(ctx context.Context, id int64) error {
	s.logger.Info("Deleting provider", "id", id)

	if err := s.providerStore.DeleteProvider(ctx, id); err != nil {
		s.logger.Error("Failed to delete provider", "id", id, "error", err)
		return err
	}

	s.logger.Info("Provider deleted successfully", "id", id)
	return nil
}

// GetAllProvidersWithRelations retrieves all providers along with their related LLM services and system prompts
// This is optimized for UI display where we need to show meaningful names instead of IDs
func (s *ProviderService) GetAllProvidersWithRelations(ctx context.Context) ([]struct {
	Provider     store.Provider
	LLMService   store.LlmService
	SystemPrompt *store.Prompt
}, error) {
	s.logger.Debug("Getting all providers with relations")

	// Get all providers
	providers, err := s.providerStore.GetAllProviders(ctx)
	if err != nil {
		s.logger.Error("Failed to get all providers for relations", "error", err)
		return nil, err
	}

	// Get all LLM services for efficient lookup
	allLLMServices, err := s.providerStore.GetAllLLMServices(ctx)
	if err != nil {
		s.logger.Error("Failed to get all LLM services for provider relations", "error", err)
		return nil, err
	}

	// Create a map for efficient LLM service lookup
	llmServiceMap := make(map[int64]store.LlmService)
	for _, service := range allLLMServices {
		llmServiceMap[service.ID] = service
	}

	// Get all system prompts for efficient lookup
	allPrompts, err := s.providerStore.GetAllPrompts(ctx)
	if err != nil {
		s.logger.Error("Failed to get all prompts for provider relations", "error", err)
		return nil, err
	}

	// Create a map for efficient prompt lookup
	promptMap := make(map[int64]store.Prompt)
	for _, prompt := range allPrompts {
		promptMap[prompt.ID] = prompt
	}

	// Build the result with relations
	var result []struct {
		Provider     store.Provider
		LLMService   store.LlmService
		SystemPrompt *store.Prompt
	}

	for _, provider := range providers {
		// Get the LLM service
		llmService, exists := llmServiceMap[provider.LlmServiceID]
		if !exists {
			s.logger.Warn("LLM service not found for provider", "provider_id", provider.ID, "llm_service_id", provider.LlmServiceID)
			continue
		}

		// Get the system prompt (if any)
		var systemPrompt *store.Prompt
		if provider.SystemPromptID.Valid {
			if prompt, exists := promptMap[provider.SystemPromptID.Int64]; exists {
				systemPrompt = &prompt
			} else {
				s.logger.Warn("System prompt not found for provider", "provider_id", provider.ID, "system_prompt_id", provider.SystemPromptID.Int64)
			}
		}

		result = append(result, struct {
			Provider     store.Provider
			LLMService   store.LlmService
			SystemPrompt *store.Prompt
		}{
			Provider:     provider,
			LLMService:   llmService,
			SystemPrompt: systemPrompt,
		})
	}

	s.logger.Debug("All providers with relations retrieved successfully", "count", len(result))
	return result, nil
}
