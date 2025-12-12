-- messages.sql
-- CRUD operations for messages table
-- Manages chat messages within conversations
-- Uses uuid v7 for ordering. Timestamps managed by DB.

-- name: CreateMessage :execlastid
INSERT INTO messages (conversation_id, uuid, role, content, metadata, created_at, updated_at)
VALUES (:conversation_id, :uuid, :role, :content, :metadata, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = :id;

-- name: GetMessageByUUID :one
SELECT * FROM messages WHERE uuid = :uuid;

-- name: ListMessages :many
SELECT * FROM messages ORDER BY uuid;

-- name: ListMessagesByConversation :many
SELECT * FROM messages 
WHERE conversation_id = :conversation_id 
ORDER BY uuid ASC;

-- name: ListMessagesByRole :many
SELECT * FROM messages 
WHERE role = :role 
ORDER BY uuid;

-- name: UpdateMessageByID :exec
UPDATE messages
SET conversation_id = :conversation_id,
    uuid = :uuid,
    role = :role,
    content = :content,
    metadata = :metadata,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteMessageByID :exec
DELETE FROM messages WHERE id = :id;

-- name: DeleteMessagesByConversation :exec
DELETE FROM messages WHERE conversation_id = :conversation_id;

-- ========================================
-- Composite Queries - Messages with Relations
-- ========================================

-- name: GetMessageWithConversation :one
SELECT 
    m.*,
    c.id as conversation_id,
    c.title as conversation_title
FROM messages m
LEFT JOIN conversations c ON m.conversation_id = c.id
WHERE m.id = :id;

-- name: GetConversationHistory :many
SELECT 
    m.*,
    c.title as conversation_title
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE m.conversation_id = :conversation_id
ORDER BY m.uuid ASC;

-- name: GetLatestMessageByConversation :one
SELECT * FROM messages 
WHERE conversation_id = :conversation_id 
ORDER BY uuid DESC 
LIMIT 1;

-- name: CountMessagesByConversation :one
SELECT COUNT(*) as count 
FROM messages 
WHERE conversation_id = :conversation_id;
