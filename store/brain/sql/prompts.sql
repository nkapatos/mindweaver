-- prompts.sql
-- CRUD operations for prompts table
-- Manages reusable system and user prompts
-- Timestamps managed by DB.

-- name: CreatePrompt :execlastid
INSERT INTO prompts (title, content, category, is_system)
VALUES (:title, :content, :category, :is_system);

-- name: GetPromptByID :one
SELECT * FROM prompts WHERE id = :id;

-- name: ListPrompts :many
SELECT * FROM prompts ORDER BY created_at DESC;

-- name: ListSystemPrompts :many
SELECT * FROM prompts 
WHERE is_system = 1 
ORDER BY created_at DESC;

-- name: ListUserPrompts :many
SELECT * FROM prompts 
WHERE is_system = 0 
ORDER BY created_at DESC;

-- name: ListPromptsByCategory :many
SELECT * FROM prompts 
WHERE category = :category 
ORDER BY created_at DESC;

-- name: GetPromptByTitle :one
SELECT * FROM prompts WHERE title = :title LIMIT 1;

-- name: UpdatePromptByID :exec
UPDATE prompts
SET title = :title,
    content = :content,
    category = :category,
    is_system = :is_system,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeletePromptByID :exec
DELETE FROM prompts WHERE id = :id;
