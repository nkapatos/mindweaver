package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

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
