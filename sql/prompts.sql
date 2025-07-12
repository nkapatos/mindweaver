-- name: CreatePrompt :exec
INSERT INTO prompts (actor_id, title, content, is_system, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetPromptById :one
SELECT id, actor_id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE id = ? LIMIT 1;

-- name: GetAllPrompts :many
SELECT id, actor_id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts;

-- name: GetPromptsByActorID :many
SELECT id, actor_id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE actor_id = ?;

-- name: GetSystemPrompts :many
SELECT id, actor_id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE is_system = 1;

-- name: UpdatePrompt :exec
UPDATE prompts SET actor_id = ?, title = ?, content = ?, is_system = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? WHERE id = ?;

-- name: DeletePrompt :exec
DELETE FROM prompts WHERE id = ?;