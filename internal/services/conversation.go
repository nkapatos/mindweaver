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
func (s *ConversationService) CreateConversation(ctx context.Context, actorID int64, title string, description string, isActive bool, metadata string, createdBy, updatedBy int64) (*store.Conversation, error) {
	s.logger.Info("Creating new conversation", "actor_id", actorID, "title", title)

	params := store.CreateConversationParams{
		ActorID:     actorID,
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
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
func (s *ConversationService) UpdateConversation(ctx context.Context, id int64, actorID int64, title string, description string, isActive bool, metadata string, updatedBy int64) error {
	s.logger.Info("Updating conversation", "id", id, "title", title)

	params := store.UpdateConversationParams{
		ID:          id,
		ActorID:     actorID,
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
		UpdatedBy:   updatedBy,
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

// GetConversationWithActor retrieves a conversation along with its owner actor
func (s *ConversationService) GetConversationWithActor(ctx context.Context, conversationID int64) (*store.Conversation, *store.Actor, error) {
	s.logger.Debug("Getting conversation with actor", "conversation_id", conversationID)

	// Get the conversation
	conversation, err := s.conversationStore.GetConversationByID(ctx, conversationID)
	if err != nil {
		s.logger.Error("Failed to get conversation", "conversation_id", conversationID, "error", err)
		return nil, nil, err
	}

	// Get the related actor (conversation owner)
	actor, err := s.conversationStore.GetActorByID(ctx, conversation.ActorID)
	if err != nil {
		s.logger.Error("Failed to get actor for conversation", "conversation_id", conversationID, "actor_id", conversation.ActorID, "error", err)
		return &conversation, nil, err
	}

	s.logger.Debug("Conversation with actor retrieved successfully", "conversation_id", conversationID)
	return &conversation, &actor, nil
}

// GetConversationWithMessages retrieves a conversation along with all its messages
func (s *ConversationService) GetConversationWithMessages(ctx context.Context, conversationID int64) (*store.Conversation, []store.Message, error) {
	s.logger.Debug("Getting conversation with messages", "conversation_id", conversationID)

	// Get the conversation
	conversation, err := s.conversationStore.GetConversationByID(ctx, conversationID)
	if err != nil {
		s.logger.Error("Failed to get conversation", "conversation_id", conversationID, "error", err)
		return nil, nil, err
	}

	// Get all messages for this conversation
	messages, err := s.conversationStore.GetMessagesByConversationID(ctx, conversationID)
	if err != nil {
		s.logger.Error("Failed to get messages for conversation", "conversation_id", conversationID, "error", err)
		return &conversation, nil, err
	}

	s.logger.Debug("Conversation with messages retrieved successfully", "conversation_id", conversationID, "message_count", len(messages))
	return &conversation, messages, nil
}

// GetConversationWithActorAndMessages retrieves a conversation with both its actor and messages
//
// FUTURE OPTIMIZATION: If this becomes a performance bottleneck, consider:
// 1. SQL JOIN approach: Single query with JOINs to actors and LEFT JOIN to messages
// 2. Pagination: For conversations with many messages, implement message pagination
// 3. Stored procedure: Complex conversation loading logic in database
// 4. Caching: Cache conversation context for frequently accessed conversations
// 5. Lazy loading: Load messages only when needed (e.g., on scroll)
func (s *ConversationService) GetConversationWithActorAndMessages(ctx context.Context, conversationID int64) (*store.Conversation, *store.Actor, []store.Message, error) {
	s.logger.Debug("Getting conversation with actor and messages", "conversation_id", conversationID)

	// Get the conversation
	conversation, err := s.conversationStore.GetConversationByID(ctx, conversationID)
	if err != nil {
		s.logger.Error("Failed to get conversation", "conversation_id", conversationID, "error", err)
		return nil, nil, nil, err
	}

	// Get the related actor (conversation owner)
	actor, err := s.conversationStore.GetActorByID(ctx, conversation.ActorID)
	if err != nil {
		s.logger.Error("Failed to get actor for conversation", "conversation_id", conversationID, "actor_id", conversation.ActorID, "error", err)
		return &conversation, nil, nil, err
	}

	// Get all messages for this conversation
	messages, err := s.conversationStore.GetMessagesByConversationID(ctx, conversationID)
	if err != nil {
		s.logger.Error("Failed to get messages for conversation", "conversation_id", conversationID, "error", err)
		return &conversation, &actor, nil, err
	}

	s.logger.Debug("Conversation with actor and messages retrieved successfully", "conversation_id", conversationID, "message_count", len(messages))
	return &conversation, &actor, messages, nil
}
