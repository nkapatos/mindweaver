package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/nkapatos/mindweaver/internal/adapters"
	"github.com/nkapatos/mindweaver/internal/store"
)

// modelCacheService handles caching of model data from LLM services
type modelCacheService struct {
	querier store.Querier
}

// newModelCacheService creates a new model cache service
func newModelCacheService(querier store.Querier) *modelCacheService {
	return &modelCacheService{
		querier: querier,
	}
}

// GetCachedModels retrieves cached models for a service, refreshing if needed
func (s *modelCacheService) GetCachedModels(ctx context.Context, llmServiceID int64, forceRefresh bool) ([]adapters.Model, error) {
	// If not forcing refresh, try to get from cache first
	if !forceRefresh {
		cachedModels, err := s.querier.GetModelsByLLMServiceID(ctx, llmServiceID)
		if err == nil && len(cachedModels) > 0 {
			// Check if cache is fresh (less than 1 hour old)
			oldestModel := cachedModels[0]
			if oldestModel.LastFetchedAt.Valid {
				lastFetched := oldestModel.LastFetchedAt.Time
				if time.Since(lastFetched) < time.Hour {
					// Cache is fresh, return cached models
					return s.convertToAdapterModels(cachedModels), nil
				}
			}
		}
	}

	// Cache is stale or empty, fetch fresh data
	return s.RefreshModels(ctx, llmServiceID)
}

// RefreshModels fetches fresh model data from the LLM service and updates cache
func (s *modelCacheService) RefreshModels(ctx context.Context, llmServiceID int64) ([]adapters.Model, error) {
	// Get the LLM service details
	llmService, err := s.querier.GetLLMServiceByID(ctx, llmServiceID)
	if err != nil {
		return nil, err
	}

	// Create adapter and fetch models
	adapter, err := adapters.NewAdapter(llmService.Adapter, llmService.ApiKey, llmService.BaseUrl)
	if err != nil {
		return nil, err
	}

	// Fetch models from the service
	freshModels, err := adapter.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	// Clear existing cache for this service
	if err := s.querier.DeleteModelsByLLMServiceID(ctx, llmServiceID); err != nil {
		return nil, err
	}

	// Cache the fresh models
	for _, model := range freshModels {
		// Convert adapter model fields to database types
		var description sql.NullString
		if model.Description != "" {
			description.String = model.Description
			description.Valid = true
		}

		var createdAt sql.NullInt64
		if model.Created > 0 {
			createdAt.Int64 = model.Created
			createdAt.Valid = true
		}

		var ownedBy sql.NullString
		if model.OwnedBy != "" {
			ownedBy.String = model.OwnedBy
			ownedBy.Valid = true
		}

		_, err := s.querier.CreateModel(ctx, store.CreateModelParams{
			LlmServiceID: llmServiceID,
			ModelID:      model.ID,
			Name:         model.Name,
			Provider:     model.Provider,
			Description:  description,
			CreatedAt:    createdAt,
			OwnedBy:      ownedBy,
		})
		if err != nil {
			// Log error but continue with other models
			continue
		}
	}

	return freshModels, nil
}

// GetModelByID retrieves a specific model from cache
func (s *modelCacheService) GetModelByID(ctx context.Context, modelID int64) (store.Model, error) {
	return s.querier.GetModelByID(ctx, modelID)
}

// GetModelByServiceAndModelID retrieves a model by service and model ID
func (s *modelCacheService) GetModelByServiceAndModelID(ctx context.Context, llmServiceID int64, modelID string) (store.Model, error) {
	return s.querier.GetModelByServiceAndModelID(ctx, store.GetModelByServiceAndModelIDParams{
		LlmServiceID: llmServiceID,
		ModelID:      modelID,
	})
}

// convertToAdapterModels converts store models to adapter models
func (s *modelCacheService) convertToAdapterModels(storeModels []store.Model) []adapters.Model {
	adapterModels := make([]adapters.Model, len(storeModels))
	for i, model := range storeModels {
		adapterModels[i] = adapters.Model{
			ID:          model.ModelID,
			Name:        model.Name,
			Provider:    model.Provider,
			Description: model.Description.String,
			Created:     model.CreatedAt.Int64,
			OwnedBy:     model.OwnedBy.String,
		}
	}
	return adapterModels
}
