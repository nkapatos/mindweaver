package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nkapatos/mindweaver/internal/store"
)

// MessageService handles message operations
//
// MESSAGE ORDERING OPTIMIZATION:
// Messages use UUID v7 for optimal performance and ordering:
// - UUID v7 includes a timestamp component that provides natural chronological ordering
// - Messages are ordered by UUID ASC in SQL queries, leveraging the timestamp component
// - This eliminates the need for additional ORDER BY created_at clauses
// - UUID v7 provides both uniqueness and timestamp-based sorting in a single field
// - Perfect for high-volume message systems where chronological order is critical
//
// FUTURE OPTIMIZATION: If message volume becomes extremely high, consider:
// 1. Partitioning: Partition messages by conversation_id or date ranges
// 2. Indexing: Ensure proper indexes on (conversation_id, uuid) for fast retrieval
// 3. Caching: Cache recent messages for active conversations
// 4. Pagination: Implement cursor-based pagination using UUID v7 timestamps
type MessageService struct {
	messageStore store.Querier
	logger       *slog.Logger
}

func NewMessageService(messageStore store.Querier) *MessageService {
	return &MessageService{
		messageStore: messageStore,
		logger:       slog.Default(),
	}
}

// CreateMessage creates a new message with an automatically generated UUID v7
// UUID v7 provides both uniqueness and timestamp-based ordering for optimal message sequencing
func (s *MessageService) CreateMessage(ctx context.Context, conversationID, createdBy int64, content, messageType, metadata string, createdByParam, updatedBy int64) (*store.Message, error) {
	// Generate UUID v7 for optimal message ordering and uniqueness
	messageUUID := uuid.Must(uuid.NewV7())

	s.logger.Info("Creating new message", "conversation_id", conversationID, "created_by", createdBy, "uuid", messageUUID)

	params := store.CreateMessageParams{
		ConversationID: conversationID,
		Uuid:           messageUUID.String(),
		Content:        content,
		MessageType:    sql.NullString{String: messageType, Valid: messageType != ""},
		Metadata:       sql.NullString{String: metadata, Valid: metadata != ""},
		CreatedBy:      createdByParam,
		UpdatedBy:      updatedBy,
	}

	message, err := s.messageStore.CreateMessage(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create message", "conversation_id", conversationID, "created_by", createdBy, "uuid", messageUUID, "error", err)
		return nil, err
	}

	s.logger.Info("Message created successfully", "id", message.ID, "uuid", messageUUID)
	return &message, nil
}

// GetMessageByID retrieves a message by its ID
func (s *MessageService) GetMessageByID(ctx context.Context, id int64) (*store.Message, error) {
	s.logger.Debug("Getting message by ID", "id", id)

	message, err := s.messageStore.GetMessageByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get message by ID", "id", id, "error", err)
		return nil, err
	}

	s.logger.Debug("Message retrieved successfully", "id", id, "uuid", message.Uuid)
	return &message, nil
}

// GetMessageByUUID retrieves a message by its UUID
func (s *MessageService) GetMessageByUUID(ctx context.Context, uuid string) (*store.Message, error) {
	s.logger.Debug("Getting message by UUID", "uuid", uuid)

	message, err := s.messageStore.GetMessageByUUID(ctx, uuid)
	if err != nil {
		s.logger.Error("Failed to get message by UUID", "uuid", uuid, "error", err)
		return nil, err
	}

	s.logger.Debug("Message retrieved successfully", "id", message.ID, "uuid", uuid)
	return &message, nil
}

// GetMessagesByConversationID retrieves all messages for a specific conversation
func (s *MessageService) GetMessagesByConversationID(ctx context.Context, conversationID int64) ([]store.Message, error) {
	s.logger.Debug("Getting messages by conversation ID", "conversation_id", conversationID)

	messages, err := s.messageStore.GetMessagesByConversationID(ctx, conversationID)
	if err != nil {
		s.logger.Error("Failed to get messages by conversation ID", "conversation_id", conversationID, "error", err)
		return nil, err
	}

	s.logger.Debug("Messages retrieved successfully", "conversation_id", conversationID, "count", len(messages))
	return messages, nil
}

// GetMessagesByCreatedBy retrieves all messages sent by a specific actor
func (s *MessageService) GetMessagesByCreatedBy(ctx context.Context, createdBy int64) ([]store.Message, error) {
	s.logger.Debug("Getting messages by created_by", "created_by", createdBy)

	messages, err := s.messageStore.GetMessagesByActorID(ctx, createdBy)
	if err != nil {
		s.logger.Error("Failed to get messages by created_by", "created_by", createdBy, "error", err)
		return nil, err
	}

	s.logger.Debug("Messages retrieved successfully", "created_by", createdBy, "count", len(messages))
	return messages, nil
}

// UpdateMessage updates a message by its ID
func (s *MessageService) UpdateMessage(ctx context.Context, id int64, content, messageType, metadata string, updatedBy int64) error {
	s.logger.Info("Updating message", "id", id)

	params := store.UpdateMessageParams{
		ID:          id,
		Content:     content,
		MessageType: sql.NullString{String: messageType, Valid: messageType != ""},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
		UpdatedBy:   updatedBy,
	}

	if err := s.messageStore.UpdateMessage(ctx, params); err != nil {
		s.logger.Error("Failed to update message", "id", id, "error", err)
		return err
	}

	s.logger.Info("Message updated successfully", "id", id)
	return nil
}

// DeleteMessage deletes a message by its ID
func (s *MessageService) DeleteMessage(ctx context.Context, id int64) error {
	s.logger.Info("Deleting message", "id", id)

	if err := s.messageStore.DeleteMessage(ctx, id); err != nil {
		s.logger.Error("Failed to delete message", "id", id, "error", err)
		return err
	}

	s.logger.Info("Message deleted successfully", "id", id)
	return nil
}

// GetMessageWithConversation retrieves a message along with its conversation
func (s *MessageService) GetMessageWithConversation(ctx context.Context, messageID int64) (*store.Message, *store.Conversation, error) {
	s.logger.Debug("Getting message with conversation", "message_id", messageID)

	// Get the message
	message, err := s.messageStore.GetMessageByID(ctx, messageID)
	if err != nil {
		s.logger.Error("Failed to get message", "message_id", messageID, "error", err)
		return nil, nil, err
	}

	// Get the related conversation
	conversation, err := s.messageStore.GetConversationByID(ctx, message.ConversationID)
	if err != nil {
		s.logger.Error("Failed to get conversation for message", "message_id", messageID, "conversation_id", message.ConversationID, "error", err)
		return &message, nil, err
	}

	s.logger.Debug("Message with conversation retrieved successfully", "message_id", messageID)
	return &message, &conversation, nil
}

// GetMessageWithActor retrieves a message along with its actor
func (s *MessageService) GetMessageWithActor(ctx context.Context, messageID int64) (*store.Message, *store.Actor, error) {
	s.logger.Debug("Getting message with actor", "message_id", messageID)

	// Get the message
	message, err := s.messageStore.GetMessageByID(ctx, messageID)
	if err != nil {
		s.logger.Error("Failed to get message", "message_id", messageID, "error", err)
		return nil, nil, err
	}

	// Get the related actor
	actor, err := s.messageStore.GetActorByID(ctx, message.CreatedBy)
	if err != nil {
		s.logger.Error("Failed to get actor for message", "message_id", messageID, "created_by", message.CreatedBy, "error", err)
		return &message, nil, err
	}

	s.logger.Debug("Message with actor retrieved successfully", "message_id", messageID)
	return &message, &actor, nil
}

// GetMessageWithRelations retrieves a message along with both its conversation and actor
//
// FUTURE OPTIMIZATION: If this becomes a performance bottleneck, consider:
// 1. SQL JOIN approach: Single query with JOINs to conversations and actors
// 2. Batch loading: Load multiple messages with relations in one operation
// 3. Stored procedure: Complex message loading logic for chat interfaces
// 4. Caching: Cache message context for frequently accessed messages
// 5. Streaming: For real-time chat, consider streaming message updates
func (s *MessageService) GetMessageWithRelations(ctx context.Context, messageID int64) (*store.Message, *store.Conversation, *store.Actor, error) {
	s.logger.Debug("Getting message with relations", "message_id", messageID)

	// Get the message
	message, err := s.messageStore.GetMessageByID(ctx, messageID)
	if err != nil {
		s.logger.Error("Failed to get message", "message_id", messageID, "error", err)
		return nil, nil, nil, err
	}

	// Get the related conversation
	conversation, err := s.messageStore.GetConversationByID(ctx, message.ConversationID)
	if err != nil {
		s.logger.Error("Failed to get conversation for message", "message_id", messageID, "conversation_id", message.ConversationID, "error", err)
		return &message, nil, nil, err
	}

	// Get the related actor
	actor, err := s.messageStore.GetActorByID(ctx, message.CreatedBy)
	if err != nil {
		s.logger.Error("Failed to get actor for message", "message_id", messageID, "created_by", message.CreatedBy, "error", err)
		return &message, &conversation, nil, err
	}

	s.logger.Debug("Message with relations retrieved successfully", "message_id", messageID)
	return &message, &conversation, &actor, nil
}
