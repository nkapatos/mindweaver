package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

type ActorService struct {
	actorStore store.Querier
	logger     *slog.Logger
}

func NewActorService(actorStore store.Querier) *ActorService {
	return &ActorService{
		actorStore: actorStore,
		logger:     slog.Default(),
	}
}

// CreateActor creates a new actor with the given type, name, and metadata
func (s *ActorService) CreateActor(ctx context.Context, actorType, name, displayName, avatarURL, metadata string, isActive bool, createdBy, updatedBy int64) error {
	s.logger.Info("Creating new actor", "type", actorType, "name", name)

	params := store.CreateActorParams{
		Type:        actorType,
		Name:        name,
		DisplayName: sql.NullString{String: displayName, Valid: displayName != ""},
		AvatarUrl:   sql.NullString{String: avatarURL, Valid: avatarURL != ""},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
	}

	if _, err := s.actorStore.CreateActor(ctx, params); err != nil {
		s.logger.Error("Failed to create actor", "type", actorType, "name", name, "error", err)
		return err
	}

	s.logger.Info("Actor created successfully", "type", actorType, "name", name)
	return nil
}

// GetActorByID retrieves an actor by their ID
func (s *ActorService) GetActorByID(ctx context.Context, id int64) (store.Actor, error) {
	s.logger.Debug("Getting actor by ID", "id", id)

	actor, err := s.actorStore.GetActorByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get actor by ID", "id", id, "error", err)
		return store.Actor{}, err
	}

	s.logger.Debug("Actor retrieved successfully", "id", id, "name", actor.Name)
	return actor, nil
}

// GetActorByName retrieves an actor by their name and type
func (s *ActorService) GetActorByName(ctx context.Context, name, actorType string) (store.Actor, error) {
	s.logger.Debug("Getting actor by name", "name", name, "type", actorType)

	params := store.GetActorByNameParams{
		Name: name,
		Type: actorType,
	}

	actor, err := s.actorStore.GetActorByName(ctx, params)
	if err != nil {
		s.logger.Error("Failed to get actor by name", "name", name, "type", actorType, "error", err)
		return store.Actor{}, err
	}

	s.logger.Debug("Actor retrieved successfully", "name", name, "type", actorType, "id", actor.ID)
	return actor, nil
}

// GetActorsByType retrieves all actors of a specific type
func (s *ActorService) GetActorsByType(ctx context.Context, actorType string) ([]store.Actor, error) {
	s.logger.Debug("Getting actors by type", "type", actorType)

	actors, err := s.actorStore.GetActorsByType(ctx, actorType)
	if err != nil {
		s.logger.Error("Failed to get actors by type", "type", actorType, "error", err)
		return nil, err
	}

	s.logger.Debug("Actors retrieved successfully", "type", actorType, "count", len(actors))
	return actors, nil
}

// UpdateActor updates an actor's information by their ID
func (s *ActorService) UpdateActor(ctx context.Context, id int64, actorType, name, displayName, avatarURL, metadata string, isActive bool, updatedBy int64) error {
	s.logger.Info("Updating actor", "id", id, "type", actorType, "name", name)

	params := store.UpdateActorParams{
		Type:        actorType,
		Name:        name,
		DisplayName: sql.NullString{String: displayName, Valid: displayName != ""},
		AvatarUrl:   sql.NullString{String: avatarURL, Valid: avatarURL != ""},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		UpdatedBy:   updatedBy,
		ID:          id,
	}

	if err := s.actorStore.UpdateActor(ctx, params); err != nil {
		s.logger.Error("Failed to update actor", "id", id, "type", actorType, "name", name, "error", err)
		return err
	}

	s.logger.Info("Actor updated successfully", "id", id, "name", name)
	return nil
}

// DeleteActor deletes an actor by their ID
func (s *ActorService) DeleteActor(ctx context.Context, id int64) error {
	s.logger.Info("Deleting actor", "id", id)

	if err := s.actorStore.DeleteActor(ctx, id); err != nil {
		s.logger.Error("Failed to delete actor", "id", id, "error", err)
		return err
	}

	s.logger.Info("Actor deleted successfully", "id", id)
	return nil
}
