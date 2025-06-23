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
func (s *PromptService) CreatePrompt(ctx context.Context, userID *int64, title, content string, isSystem bool) error {
	s.logger.Info("Creating new prompt",
		"title", title,
		"user_id", userID,
		"is_system", isSystem,
		"content_length", len(content))

	var userIDNull sql.NullInt64
	if userID != nil {
		userIDNull.Int64 = *userID
		userIDNull.Valid = true
	}

	var isSystemNull sql.NullInt64
	if isSystem {
		isSystemNull.Int64 = 1
		isSystemNull.Valid = true
	}

	params := store.CreatePromptParams{
		UserID:   userIDNull,
		Title:    title,
		Content:  content,
		IsSystem: isSystemNull,
	}

	if err := s.promptStore.CreatePrompt(ctx, params); err != nil {
		s.logger.Error("Failed to create prompt",
			"title", title,
			"user_id", userID,
			"is_system", isSystem,
			"error", err)
		return err
	}

	s.logger.Info("Prompt created successfully", "title", title, "user_id", userID, "is_system", isSystem)
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

// GetPromptsByUserID retrieves all prompts for a specific user
func (s *PromptService) GetPromptsByUserID(ctx context.Context, userID int64) ([]store.Prompt, error) {
	s.logger.Debug("Getting prompts by user ID", "user_id", userID)

	userIDNull := sql.NullInt64{Int64: userID, Valid: true}
	prompts, err := s.promptStore.GetPromptsByUserId(ctx, userIDNull)
	if err != nil {
		s.logger.Error("Failed to get prompts by user ID", "user_id", userID, "error", err)
		return nil, err
	}

	s.logger.Debug("Prompts retrieved successfully", "user_id", userID, "count", len(prompts))
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
func (s *PromptService) UpdatePrompt(ctx context.Context, id int64, title, content string, isSystem bool) error {
	s.logger.Info("Updating prompt",
		"id", id,
		"title", title,
		"is_system", isSystem,
		"content_length", len(content))

	var isSystemNull sql.NullInt64
	if isSystem {
		isSystemNull.Int64 = 1
		isSystemNull.Valid = true
	}

	params := store.UpdatePromptParams{
		ID:       id,
		Title:    title,
		Content:  content,
		IsSystem: isSystemNull,
	}

	if err := s.promptStore.UpdatePrompt(ctx, params); err != nil {
		s.logger.Error("Failed to update prompt",
			"id", id,
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
