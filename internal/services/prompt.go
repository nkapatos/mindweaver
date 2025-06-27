package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

type PromptService struct {
	promptStore store.Querier
	logger      *slog.Logger
}

func NewPromptService(promptStore store.Querier) *PromptService {
	return &PromptService{
		promptStore: promptStore,
		logger:      slog.Default(),
	}
}

// CreatePrompt creates a new prompt
func (s *PromptService) CreatePrompt(ctx context.Context, actorID *int64, title, content string, isSystem bool) error {
	s.logger.Info("Creating new prompt",
		"title", title,
		"actor_id", actorID,
		"is_system", isSystem,
		"content_length", len(content))

	var actorIDNull sql.NullInt64
	if actorID != nil {
		actorIDNull.Int64 = *actorID
		actorIDNull.Valid = true
	}

	var isSystemNull sql.NullInt64
	if isSystem {
		isSystemNull.Int64 = 1
		isSystemNull.Valid = true
	}

	params := store.CreatePromptParams{
		ActorID:  actorIDNull,
		Title:    title,
		Content:  content,
		IsSystem: isSystemNull,
	}

	if err := s.promptStore.CreatePrompt(ctx, params); err != nil {
		s.logger.Error("Failed to create prompt",
			"title", title,
			"actor_id", actorID,
			"is_system", isSystem,
			"error", err)
		return err
	}

	s.logger.Info("Prompt created successfully", "title", title, "actor_id", actorID, "is_system", isSystem)
	return nil
}

// GetPromptByID retrieves a prompt by its ID
func (s *PromptService) GetPromptByID(ctx context.Context, id int64) (store.Prompt, error) {
	s.logger.Debug("Getting prompt by ID", "id", id)

	prompt, err := s.promptStore.GetPromptById(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get prompt by ID", "id", id, "error", err)
		return store.Prompt{}, err
	}

	s.logger.Debug("Prompt retrieved successfully", "id", id, "title", prompt.Title)
	return prompt, nil
}

// GetAllPrompts retrieves all prompts
func (s *PromptService) GetAllPrompts(ctx context.Context) ([]store.Prompt, error) {
	s.logger.Debug("Getting all prompts")

	prompts, err := s.promptStore.GetAllPrompts(ctx)
	if err != nil {
		s.logger.Error("Failed to get all prompts", "error", err)
		return nil, err
	}

	s.logger.Debug("All prompts retrieved successfully", "count", len(prompts))
	return prompts, nil
}

// GetPromptsByActorID retrieves all prompts for a specific actor
func (s *PromptService) GetPromptsByActorID(ctx context.Context, actorID int64) ([]store.Prompt, error) {
	s.logger.Debug("Getting prompts by actor ID", "actor_id", actorID)

	actorIDNull := sql.NullInt64{Int64: actorID, Valid: true}
	prompts, err := s.promptStore.GetPromptsByActorID(ctx, actorIDNull)
	if err != nil {
		s.logger.Error("Failed to get prompts by actor ID", "actor_id", actorID, "error", err)
		return nil, err
	}

	s.logger.Debug("Prompts retrieved successfully", "actor_id", actorID, "count", len(prompts))
	return prompts, nil
}

// GetSystemPrompts retrieves all system prompts
func (s *PromptService) GetSystemPrompts(ctx context.Context) ([]store.Prompt, error) {
	s.logger.Debug("Getting system prompts")

	prompts, err := s.promptStore.GetSystemPrompts(ctx)
	if err != nil {
		s.logger.Error("Failed to get system prompts", "error", err)
		return nil, err
	}

	s.logger.Debug("System prompts retrieved successfully", "count", len(prompts))
	return prompts, nil
}

// UpdatePrompt updates a prompt by its ID
func (s *PromptService) UpdatePrompt(ctx context.Context, id int64, actorID *int64, title, content string, isSystem bool) error {
	s.logger.Info("Updating prompt",
		"id", id,
		"actor_id", actorID,
		"title", title,
		"is_system", isSystem,
		"content_length", len(content))

	var actorIDNull sql.NullInt64
	if actorID != nil {
		actorIDNull.Int64 = *actorID
		actorIDNull.Valid = true
	}

	var isSystemNull sql.NullInt64
	if isSystem {
		isSystemNull.Int64 = 1
		isSystemNull.Valid = true
	}

	params := store.UpdatePromptParams{
		ActorID:  actorIDNull,
		Title:    title,
		Content:  content,
		IsSystem: isSystemNull,
		ID:       id,
	}

	if err := s.promptStore.UpdatePrompt(ctx, params); err != nil {
		s.logger.Error("Failed to update prompt",
			"id", id,
			"actor_id", actorID,
			"title", title,
			"is_system", isSystem,
			"error", err)
		return err
	}

	s.logger.Info("Prompt updated successfully", "id", id, "title", title, "is_system", isSystem)
	return nil
}

// DeletePrompt deletes a prompt by its ID
func (s *PromptService) DeletePrompt(ctx context.Context, id int64) error {
	s.logger.Info("Deleting prompt", "id", id)

	if err := s.promptStore.DeletePrompt(ctx, id); err != nil {
		s.logger.Error("Failed to delete prompt", "id", id, "error", err)
		return err
	}

	s.logger.Info("Prompt deleted successfully", "id", id)
	return nil
}

// GetPromptWithActor retrieves a user prompt along with its owner actor (only for non-system prompts)
//
// FUTURE OPTIMIZATION: If this becomes a performance bottleneck, consider:
// 1. SQL JOIN approach: Single query with LEFT JOIN to actors for user prompts
// 2. Batch loading: Load multiple prompts with actors in one operation
// 3. Stored procedure: Complex prompt loading logic for prompt management interfaces
// 4. Caching: Cache prompt-actor relationships for frequently accessed prompts
// 5. Indexing: Ensure proper indexes on actor_id for user prompts
func (s *PromptService) GetPromptWithActor(ctx context.Context, promptID int64) (*store.Prompt, *store.Actor, error) {
	s.logger.Debug("Getting prompt with actor", "prompt_id", promptID)

	// Get the prompt
	prompt, err := s.promptStore.GetPromptById(ctx, promptID)
	if err != nil {
		s.logger.Error("Failed to get prompt", "prompt_id", promptID, "error", err)
		return nil, nil, err
	}

	// Check if this is a user prompt (has an actor_id)
	if !prompt.ActorID.Valid {
		s.logger.Debug("Prompt is a system prompt, no actor relationship", "prompt_id", promptID)
		return &prompt, nil, nil
	}

	// Get the related actor
	actor, err := s.promptStore.GetActorByID(ctx, prompt.ActorID.Int64)
	if err != nil {
		s.logger.Error("Failed to get actor for prompt", "prompt_id", promptID, "actor_id", prompt.ActorID.Int64, "error", err)
		return &prompt, nil, err
	}

	s.logger.Debug("Prompt with actor retrieved successfully", "prompt_id", promptID)
	return &prompt, &actor, nil
}
