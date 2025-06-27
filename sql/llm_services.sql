-- name: GetLLMServiceByID :one
SELECT id, name, description, adapter, created_at, api_key, base_url, organization, configuration
FROM llm_services 
WHERE id = ?;

-- name: GetLLMServiceByName :one
SELECT id, name, description, adapter, created_at, api_key, base_url, organization, configuration
FROM llm_services 
WHERE name = ?
LIMIT 1;

-- name: GetAllLLMServices :many
SELECT id, name, description, adapter, created_at, api_key, base_url, organization, configuration
FROM llm_services 
ORDER BY name;

-- name: CreateLLMService :one
INSERT INTO llm_services (name, description, adapter, api_key, base_url, organization, configuration) 
VALUES (?, ?, ?, ?, ?, ?, ?) 
RETURNING id, name, description, adapter, created_at, api_key, base_url, organization, configuration;

-- name: UpdateLLMService :exec
UPDATE llm_services 
SET name = ?, description = ?, adapter = ?, api_key = ?, base_url = ?, organization = ?, configuration = ? 
WHERE id = ?;

-- name: DeleteLLMService :exec
DELETE FROM llm_services 
WHERE id = ?; 