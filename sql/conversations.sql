-- name: CreateConversation :one
INSERT INTO conversations (actor_id, title, description, is_active, metadata, created_by, updated_by) 
VALUES (?, ?, ?, ?, ?, ?, ?) 
RETURNING id, actor_id, title, description, is_active, metadata, created_at, updated_at, created_by, updated_by;

-- name: GetConversationByID :one
SELECT id, actor_id, title, description, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM conversations 
WHERE id = ? 
LIMIT 1;

-- name: GetConversationsByActorID :many
SELECT id, actor_id, title, description, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM conversations 
WHERE actor_id = ? AND is_active = true 
ORDER BY created_at DESC;

-- name: UpdateConversation :exec
UPDATE conversations 
SET actor_id = ?, title = ?, description = ?, is_active = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? 
WHERE id = ?;

-- name: DeleteConversation :exec
DELETE FROM conversations WHERE id = ?;
