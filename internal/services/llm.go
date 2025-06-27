package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

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

// GetLLMServiceWithProviders retrieves an LLM service along with all providers that use it
//
// FUTURE OPTIMIZATION: If this becomes a performance bottleneck, consider:
// 1. SQL JOIN approach: Single query with LEFT JOIN to providers table
// 2. Stored procedure: Complex LLM service loading with provider counts/relationships
// 3. Caching: Cache LLM service configurations and provider relationships
// 4. Batch loading: Load multiple LLM services with their providers in one operation
// 5. Aggregation: Include provider counts and usage statistics in the query
func (s *LLMService) GetLLMServiceWithProviders(ctx context.Context, llmServiceID int64) (*store.LlmService, []store.Provider, error) {
	s.logger.Debug("Getting LLM service with providers", "llm_service_id", llmServiceID)

	// Get the LLM service
	llmService, err := s.store.GetLLMServiceByID(ctx, llmServiceID)
	if err != nil {
		s.logger.Error("Failed to get LLM service", "llm_service_id", llmServiceID, "error", err)
		return nil, nil, err
	}

	// Get all providers and filter by LLM service ID
	allProviders, err := s.store.GetAllProviders(ctx)
	if err != nil {
		s.logger.Error("Failed to get all providers for LLM service filtering", "llm_service_id", llmServiceID, "error", err)
		return &llmService, nil, err
	}

	// Filter providers that use this LLM service
	var filteredProviders []store.Provider
	for _, provider := range allProviders {
		if provider.LlmServiceID == llmServiceID {
			filteredProviders = append(filteredProviders, provider)
		}
	}

	s.logger.Debug("LLM service with providers retrieved successfully", "llm_service_id", llmServiceID, "provider_count", len(filteredProviders))
	return &llmService, filteredProviders, nil
}
