-- name: CreatePrompt :exec
INSERT INTO prompts (title, content, is_system, created_by, updated_by) VALUES (?, ?, ?, ?, ?);

-- name: GetPromptById :one
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE id = ? LIMIT 1;

-- name: GetAllPrompts :many
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts;

-- name: GetPromptsByActorID :many
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE created_by = ?;

-- name: GetSystemPrompts :many
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE is_system = 1;

-- name: UpdatePrompt :exec
UPDATE prompts SET title = ?, content = ?, is_system = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? WHERE id = ?;

-- name: DeletePrompt :exec
DELETE FROM prompts WHERE id = ?;