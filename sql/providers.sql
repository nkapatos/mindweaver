-- name: GetProviderByName :one
SELECT id, llm_service_config_id, system_prompt_id, name, description, created_at, updated_at, created_by, updated_by
FROM providers
WHERE name = ?
LIMIT 1;

-- name: GetProviderByID :one
SELECT id, llm_service_config_id, system_prompt_id, name, description, created_at, updated_at, created_by, updated_by
FROM providers
WHERE id = ?;

-- name: GetAllProviders :many
SELECT id, llm_service_config_id, system_prompt_id, name, description, created_at, updated_at, created_by, updated_by
FROM providers
ORDER BY name;

-- name: CreateProvider :one
INSERT INTO providers (llm_service_config_id, system_prompt_id, name, description, created_by, updated_by) 
VALUES (?, ?, ?, ?, ?, ?) 
RETURNING id, llm_service_config_id, system_prompt_id, name, description, created_at, updated_at, created_by, updated_by;

-- name: UpdateProvider :exec
UPDATE providers 
SET llm_service_config_id = ?, system_prompt_id = ?, name = ?, description = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? 
WHERE id = ?;

-- name: DeleteProvider :exec
DELETE FROM providers 
WHERE id = ?;

-- name: GetProvidersByLLMServiceConfig :many
SELECT id, llm_service_config_id, system_prompt_id, name, description, created_at, updated_at, created_by, updated_by
FROM providers
WHERE llm_service_config_id = ?
ORDER BY name;

-- name: GetProvidersBySystemPrompt :many
SELECT id, llm_service_config_id, system_prompt_id, name, description, created_at, updated_at, created_by, updated_by
FROM providers
WHERE system_prompt_id = ?
ORDER BY name;