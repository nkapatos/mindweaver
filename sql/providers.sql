-- name: GetProviderByName :one
SELECT id, llm_service_id, system_prompt_id, name, description, created_at
FROM providers
WHERE name = ?
LIMIT 1;

-- name: GetProviderByID :one
SELECT id, llm_service_id, system_prompt_id, name, description, created_at
FROM providers
WHERE id = ?;

-- name: GetAllProviders :many
SELECT id, llm_service_id, system_prompt_id, name, description, created_at
FROM providers
ORDER BY name;

-- name: CreateProvider :one
INSERT INTO providers (llm_service_id, system_prompt_id, name, description) 
VALUES (?, ?, ?, ?) 
RETURNING id, llm_service_id, system_prompt_id, name, description, created_at;

-- name: UpdateProvider :exec
UPDATE providers 
SET llm_service_id = ?, system_prompt_id = ?, name = ?, description = ? 
WHERE id = ?;

-- name: DeleteProvider :exec
DELETE FROM providers 
WHERE id = ?;

-- name: GetProvidersByLLMService :many
SELECT id, llm_service_id, system_prompt_id, name, description, created_at
FROM providers
WHERE llm_service_id = ?
ORDER BY name;