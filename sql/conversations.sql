-- name: CreateConversation :one
INSERT INTO conversations (actor_id, title, description, metadata, is_active) 
VALUES (?, ?, ?, ?, ?) 
RETURNING id, actor_id, title, description, metadata, is_active, created_at, updated_at;

-- name: GetConversationByID :one
SELECT id, actor_id, title, description, metadata, is_active, created_at, updated_at 
FROM conversations 
WHERE id = ? 
LIMIT 1;

-- name: GetConversationsByActorID :many
SELECT id, actor_id, title, description, metadata, is_active, created_at, updated_at 
FROM conversations 
WHERE actor_id = ? AND is_active = true 
ORDER BY created_at DESC;

-- name: UpdateConversation :exec
UPDATE conversations 
SET title = ?, description = ?, metadata = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteConversation :exec
DELETE FROM conversations WHERE id = ?;
