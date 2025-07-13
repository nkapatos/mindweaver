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
func (s *PromptService) CreatePrompt(ctx context.Context, title, content string, isSystem bool, createdBy, updatedBy int64) error {
	s.logger.Info("Creating new prompt",
		"title", title,
		"created_by", createdBy,
		"is_system", isSystem,
		"content_length", len(content))

	var isSystemNull sql.NullInt64
	if isSystem {
		isSystemNull.Int64 = 1
		isSystemNull.Valid = true
	}

	params := store.CreatePromptParams{
		Title:     title,
		Content:   content,
		IsSystem:  isSystemNull,
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
	}

	if err := s.promptStore.CreatePrompt(ctx, params); err != nil {
		s.logger.Error("Failed to create prompt",
			"title", title,
			"created_by", createdBy,
			"is_system", isSystem,
			"error", err)
		return err
	}

	s.logger.Info("Prompt created successfully", "title", title, "created_by", createdBy, "is_system", isSystem)
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

// GetPromptsByCreatedBy retrieves all prompts for a specific actor
func (s *PromptService) GetPromptsByCreatedBy(ctx context.Context, createdBy int64) ([]store.Prompt, error) {
	s.logger.Debug("Getting prompts by created_by", "created_by", createdBy)

	prompts, err := s.promptStore.GetPromptsByActorID(ctx, createdBy) // SQL query filters by created_by
	if err != nil {
		s.logger.Error("Failed to get prompts by created_by", "created_by", createdBy, "error", err)
		return nil, err
	}

	s.logger.Debug("Prompts retrieved successfully", "created_by", createdBy, "count", len(prompts))
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
func (s *PromptService) UpdatePrompt(ctx context.Context, id int64, createdBy *int64, title, content string, isSystem bool, updatedBy int64) error {
	s.logger.Info("Updating prompt",
		"id", id,
		"created_by", createdBy,
		"title", title,
		"is_system", isSystem,
		"content_length", len(content))

	var createdByNull sql.NullInt64
	if createdBy != nil {
		createdByNull.Int64 = *createdBy
		createdByNull.Valid = true
	}

	var isSystemNull sql.NullInt64
	if isSystem {
		isSystemNull.Int64 = 1
		isSystemNull.Valid = true
	}

	params := store.UpdatePromptParams{
		Title:     title,
		Content:   content,
		IsSystem:  isSystemNull,
		UpdatedBy: updatedBy,
		ID:        id,
	}

	if err := s.promptStore.UpdatePrompt(ctx, params); err != nil {
		s.logger.Error("Failed to update prompt",
			"id", id,
			"created_by", createdBy,
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
// 5. Indexing: Ensure proper indexes on created_by for user prompts
func (s *PromptService) GetPromptWithActor(ctx context.Context, promptID int64) (*store.Prompt, *store.Actor, error) {
	s.logger.Debug("Getting prompt with actor", "prompt_id", promptID)

	// Get the prompt
	prompt, err := s.promptStore.GetPromptById(ctx, promptID)
	if err != nil {
		s.logger.Error("Failed to get prompt", "prompt_id", promptID, "error", err)
		return nil, nil, err
	}

	// Check if this is a user prompt (has a created_by)
	if prompt.CreatedBy == 0 {
		s.logger.Debug("Prompt is a system prompt, no actor relationship", "prompt_id", promptID)
		return &prompt, nil, nil
	}

	// Get the related actor
	actor, err := s.promptStore.GetActorByID(ctx, prompt.CreatedBy)
	if err != nil {
		s.logger.Error("Failed to get actor for prompt", "prompt_id", promptID, "created_by", prompt.CreatedBy, "error", err)
		return &prompt, nil, err
	}

	s.logger.Debug("Prompt with actor retrieved successfully", "prompt_id", promptID)
	return &prompt, &actor, nil
}

// GetAllPromptsWithRelations retrieves all prompts along with their related actors
// This is optimized for UI display where we need to show meaningful names instead of IDs
func (s *PromptService) GetAllPromptsWithRelations(ctx context.Context) ([]struct {
	Prompt store.Prompt
	Actor  *store.Actor
}, error) {
	s.logger.Debug("Getting all prompts with relations")

	// Get all prompts
	prompts, err := s.promptStore.GetAllPrompts(ctx)
	if err != nil {
		s.logger.Error("Failed to get all prompts for relations", "error", err)
		return nil, err
	}

	// Get all actors for efficient lookup
	allActors, err := s.promptStore.GetAllActors(ctx)
	if err != nil {
		s.logger.Error("Failed to get all actors for prompt relations", "error", err)
		return nil, err
	}

	// Create a map for efficient actor lookup
	actorMap := make(map[int64]store.Actor)
	for _, actor := range allActors {
		actorMap[actor.ID] = actor
	}

	// Build the result with relations
	var result []struct {
		Prompt store.Prompt
		Actor  *store.Actor
	}

	for _, prompt := range prompts {
		// Get the related actor (if any)
		var actor *store.Actor
		if prompt.CreatedBy != 0 { // Changed from ActorID.Valid to CreatedBy != 0
			if a, exists := actorMap[prompt.CreatedBy]; exists { // Changed from ActorID.Int64 to CreatedBy
				actor = &a
			} else {
				s.logger.Warn("Actor not found for prompt", "prompt_id", prompt.ID, "created_by", prompt.CreatedBy) // Changed from ActorID.Int64 to CreatedBy
			}
		}

		result = append(result, struct {
			Prompt store.Prompt
			Actor  *store.Actor
		}{
			Prompt: prompt,
			Actor:  actor,
		})
	}

	s.logger.Debug("All prompts with relations retrieved successfully", "count", len(result))
	return result, nil
}
