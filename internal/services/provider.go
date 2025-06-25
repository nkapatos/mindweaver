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
func (s *ProviderService) CreateProvider(ctx context.Context, name, providerType string, isActive bool) error {
	s.logger.Info("Creating new provider",
		"name", name,
		"type", providerType,
		"is_active", isActive)

	var isActiveNull sql.NullBool
	if isActive {
		isActiveNull.Bool = true
		isActiveNull.Valid = true
	}

	params := store.CreateProviderParams{
		Name:     name,
		Type:     providerType,
		IsActive: isActiveNull,
	}

	if _, err := s.providerStore.CreateProvider(ctx, params); err != nil {
		s.logger.Error("Failed to create provider",
			"name", name,
			"type", providerType,
			"is_active", isActive,
			"error", err)
		return err
	}

	s.logger.Info("Provider created successfully", "name", name, "type", providerType, "is_active", isActive)
	return nil
}

// GetProviderByID retrieves a provider by its ID
func (s *ProviderService) GetProviderByID(ctx context.Context, id int64) (store.Provider, error) {
	s.logger.Debug("Getting provider by ID", "id", id)

	provider, err := s.providerStore.GetProviderByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get provider by ID", "id", id, "error", err)
		return store.Provider{}, err
	}

	s.logger.Debug("Provider retrieved successfully", "id", id, "name", provider.Name)
	return provider, nil
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

// UpdateProvider updates a provider by its ID
func (s *ProviderService) UpdateProvider(ctx context.Context, id int64, name, providerType string, isActive bool) error {
	s.logger.Info("Updating provider",
		"id", id,
		"name", name,
		"type", providerType,
		"is_active", isActive)

	var isActiveNull sql.NullBool
	if isActive {
		isActiveNull.Bool = true
		isActiveNull.Valid = true
	}

	params := store.UpdateProviderParams{
		ID:       id,
		Name:     name,
		Type:     providerType,
		IsActive: isActiveNull,
	}

	if err := s.providerStore.UpdateProvider(ctx, params); err != nil {
		s.logger.Error("Failed to update provider",
			"id", id,
			"name", name,
			"type", providerType,
			"is_active", isActive,
			"error", err)
		return err
	}

	s.logger.Info("Provider updated successfully", "id", id, "name", name, "type", providerType, "is_active", isActive)
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
