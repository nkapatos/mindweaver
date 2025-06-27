package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

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

// CreateMessage creates a new message
func (s *MessageService) CreateMessage(ctx context.Context, conversationID, senderActorID int64, uuid, content, messageType, metadata string) (*store.Message, error) {
	s.logger.Info("Creating new message", "conversation_id", conversationID, "sender_actor_id", senderActorID, "uuid", uuid)

	params := store.CreateMessageParams{
		ConversationID: conversationID,
		SenderActorID:  senderActorID,
		Uuid:           uuid,
		Content:        content,
		MessageType:    sql.NullString{String: messageType, Valid: messageType != ""},
		Metadata:       sql.NullString{String: metadata, Valid: metadata != ""},
	}

	message, err := s.messageStore.CreateMessage(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create message", "conversation_id", conversationID, "sender_actor_id", senderActorID, "uuid", uuid, "error", err)
		return nil, err
	}

	s.logger.Info("Message created successfully", "id", message.ID, "uuid", uuid)
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

// GetMessagesByActorID retrieves all messages sent by a specific actor
func (s *MessageService) GetMessagesByActorID(ctx context.Context, senderActorID int64) ([]store.Message, error) {
	s.logger.Debug("Getting messages by actor ID", "sender_actor_id", senderActorID)

	messages, err := s.messageStore.GetMessagesByActorID(ctx, senderActorID)
	if err != nil {
		s.logger.Error("Failed to get messages by actor ID", "sender_actor_id", senderActorID, "error", err)
		return nil, err
	}

	s.logger.Debug("Messages retrieved successfully", "sender_actor_id", senderActorID, "count", len(messages))
	return messages, nil
}

// UpdateMessage updates a message by its ID
func (s *MessageService) UpdateMessage(ctx context.Context, id int64, content, messageType, metadata string) error {
	s.logger.Info("Updating message", "id", id)

	params := store.UpdateMessageParams{
		ID:          id,
		Content:     content,
		MessageType: sql.NullString{String: messageType, Valid: messageType != ""},
		Metadata:    sql.NullString{String: metadata, Valid: metadata != ""},
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
