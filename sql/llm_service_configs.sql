-- name: GetLLMServiceConfigByID :one
SELECT id, llm_service_id, name, description, configuration, created_at, updated_at
FROM llm_service_configs 
WHERE id = ?;

-- name: GetLLMServiceConfigsByServiceID :many
SELECT id, llm_service_id, name, description, configuration, created_at, updated_at
FROM llm_service_configs 
WHERE llm_service_id = ?
ORDER BY name;

-- name: GetAllLLMServiceConfigs :many
SELECT id, llm_service_id, name, description, configuration, created_at, updated_at
FROM llm_service_configs 
ORDER BY llm_service_id, name;

-- name: CreateLLMServiceConfig :one
INSERT INTO llm_service_configs (llm_service_id, name, description, configuration) 
VALUES (?, ?, ?, ?) 
RETURNING id, llm_service_id, name, description, configuration, created_at, updated_at;

-- name: UpdateLLMServiceConfig :exec
UPDATE llm_service_configs 
SET name = ?, description = ?, configuration = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteLLMServiceConfig :exec
DELETE FROM llm_service_configs 
WHERE id = ?;

-- name: GetLLMServiceConfigByName :one
SELECT id, llm_service_id, name, description, configuration, created_at, updated_at
FROM llm_service_configs 
WHERE llm_service_id = ? AND name = ?
LIMIT 1; 