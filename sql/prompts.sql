-- name: CreatePrompt :exec
INSERT INTO prompts (user_id, title, content, is_system) VALUES (?, ?, ?, ?);

-- name: GetPromptById :one
SELECT * FROM prompts WHERE id = ? LIMIT 1;

-- name: GetAllPrompts :many
SELECT * FROM prompts;

-- name: GetPromptsByUserId :many
SELECT * FROM prompts WHERE user_id = ?;

-- name: GetSystemPrompts :many
SELECT * FROM prompts WHERE is_system = 1;

-- name: UpdatePrompt :exec
UPDATE prompts SET title = ?, content = ?, is_system = ?, updated_at = (datetime('now')) WHERE id = ?;

-- name: DeletePrompt :exec
DELETE FROM prompts WHERE id = ?;