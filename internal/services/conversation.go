package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

type ConversationService struct {
	conversationStore store.Querier
	logger            *slog.Logger
}

func NewConversationService(conversationStore store.Querier) *ConversationService {
	return &ConversationService{
		conversationStore: conversationStore,
		logger:            slog.Default(),
	}
}

// CreateConversation creates a new conversation
func (s *ConversationService) CreateConversation(ctx context.Context, actorID int64, title string, description string, isActive bool, metadata string) (*store.Conversation, error) {
	s.logger.Info("Creating new conversation", "actor_id", actorID, "title", title)

	params := store.CreateConversationParams{
		ActorID:     actorID,
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
	}

	conversation, err := s.conversationStore.CreateConversation(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create conversation", "actor_id", actorID, "title", title, "error", err)
		return nil, err
	}

	s.logger.Info("Conversation created successfully", "id", conversation.ID, "title", title)
	return &conversation, nil
}

// GetConversationByID retrieves a conversation by its ID
func (s *ConversationService) GetConversationByID(ctx context.Context, id int64) (*store.Conversation, error) {
	s.logger.Debug("Getting conversation by ID", "id", id)

	conversation, err := s.conversationStore.GetConversationByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get conversation by ID", "id", id, "error", err)
		return nil, err
	}

	s.logger.Debug("Conversation retrieved successfully", "id", id, "title", conversation.Title)
	return &conversation, nil
}

// GetConversationsByActorID retrieves all conversations for a specific actor
func (s *ConversationService) GetConversationsByActorID(ctx context.Context, actorID int64) ([]store.Conversation, error) {
	s.logger.Debug("Getting conversations by actor ID", "actor_id", actorID)

	conversations, err := s.conversationStore.GetConversationsByActorID(ctx, actorID)
	if err != nil {
		s.logger.Error("Failed to get conversations by actor ID", "actor_id", actorID, "error", err)
		return nil, err
	}

	s.logger.Debug("Conversations retrieved successfully", "actor_id", actorID, "count", len(conversations))
	return conversations, nil
}

// UpdateConversation updates a conversation by its ID
func (s *ConversationService) UpdateConversation(ctx context.Context, id int64, actorID int64, title string, description string, isActive bool, metadata string) error {
	s.logger.Info("Updating conversation", "id", id, "title", title)

	params := store.UpdateConversationParams{
		ID:          id,
		ActorID:     actorID,
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
	}

	if err := s.conversationStore.UpdateConversation(ctx, params); err != nil {
		s.logger.Error("Failed to update conversation", "id", id, "title", title, "error", err)
		return err
	}

	s.logger.Info("Conversation updated successfully", "id", id, "title", title)
	return nil
}

// DeleteConversation deletes a conversation by its ID
func (s *ConversationService) DeleteConversation(ctx context.Context, id int64) error {
	s.logger.Info("Deleting conversation", "id", id)

	if err := s.conversationStore.DeleteConversation(ctx, id); err != nil {
		s.logger.Error("Failed to delete conversation", "id", id, "error", err)
		return err
	}

	s.logger.Info("Conversation deleted successfully", "id", id)
	return nil
}
