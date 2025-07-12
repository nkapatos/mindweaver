-- name: CreateMessage :one
INSERT INTO messages (conversation_id, actor_id, uuid, content, message_type, metadata, created_by, updated_by) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?) 
RETURNING id, conversation_id, actor_id, uuid, content, message_type, metadata, created_at, updated_at, created_by, updated_by;

-- name: GetMessageByUUID :one
SELECT id, conversation_id, actor_id, uuid, content, message_type, metadata, created_at, updated_at, created_by, updated_by 
FROM messages 
WHERE uuid = ? 
LIMIT 1;

-- name: GetMessageByID :one
SELECT id, conversation_id, actor_id, uuid, content, message_type, metadata, created_at, updated_at, created_by, updated_by 
FROM messages 
WHERE id = ? 
LIMIT 1;

-- name: GetMessagesByConversationID :many
SELECT id, conversation_id, actor_id, uuid, content, message_type, metadata, created_at, updated_at, created_by, updated_by 
FROM messages 
WHERE conversation_id = ? 
ORDER BY uuid ASC;

-- name: GetMessagesByActorID :many
SELECT id, conversation_id, actor_id, uuid, content, message_type, metadata, created_at, updated_at, created_by, updated_by 
FROM messages 
WHERE actor_id = ? 
ORDER BY uuid DESC;

-- name: UpdateMessage :exec
UPDATE messages 
SET content = ?, message_type = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? 
WHERE id = ?;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = ?; 