-- name: CreatePrompt :exec
INSERT INTO prompts (actor_id, title, content, is_system) VALUES (?, ?, ?, ?);

-- name: GetPromptById :one
SELECT id, actor_id, title, content, is_system, created_at, updated_at FROM prompts WHERE id = ? LIMIT 1;

-- name: GetAllPrompts :many
SELECT id, actor_id, title, content, is_system, created_at, updated_at FROM prompts;

-- name: GetPromptsByActorID :many
SELECT id, actor_id, title, content, is_system, created_at, updated_at FROM prompts WHERE actor_id = ?;

-- name: GetSystemPrompts :many
SELECT id, actor_id, title, content, is_system, created_at, updated_at FROM prompts WHERE is_system = 1;

-- name: UpdatePrompt :exec
UPDATE prompts SET actor_id = ?, title = ?, content = ?, is_system = ?, updated_at = (datetime('now')) WHERE id = ?;

-- name: DeletePrompt :exec
DELETE FROM prompts WHERE id = ?;