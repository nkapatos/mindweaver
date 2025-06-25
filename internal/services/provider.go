package services

import (
	"context"
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
