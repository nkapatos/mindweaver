-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_actor_id, uuid, content, message_type, metadata) 
VALUES (?, ?, ?, ?, ?, ?) 
RETURNING id, conversation_id, sender_actor_id, uuid, content, message_type, metadata, created_at;

-- name: GetMessageByUUID :one
SELECT id, conversation_id, sender_actor_id, uuid, content, message_type, metadata, created_at 
FROM messages 
WHERE uuid = ? 
LIMIT 1;

-- name: GetMessageByID :one
SELECT id, conversation_id, sender_actor_id, uuid, content, message_type, metadata, created_at 
FROM messages 
WHERE id = ? 
LIMIT 1;

-- name: GetMessagesByConversationID :many
SELECT id, conversation_id, sender_actor_id, uuid, content, message_type, metadata, created_at 
FROM messages 
WHERE conversation_id = ? 
ORDER BY uuid ASC;

-- name: GetMessagesByActorID :many
SELECT id, conversation_id, sender_actor_id, uuid, content, message_type, metadata, created_at 
FROM messages 
WHERE sender_actor_id = ? 
ORDER BY created_at DESC;

-- name: UpdateMessage :exec
UPDATE messages 
SET content = ?, message_type = ?, metadata = ? 
WHERE id = ?;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = ?; 