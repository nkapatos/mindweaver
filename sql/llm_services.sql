-- name: GetLLMServiceByID :one
SELECT id, name, description, adapter, api_key, base_url, organization, created_at, updated_at, created_by, updated_by
FROM llm_services 
WHERE id = ?;

-- name: GetLLMServiceByName :one
SELECT id, name, description, adapter, api_key, base_url, organization, created_at, updated_at, created_by, updated_by
FROM llm_services 
WHERE name = ?
LIMIT 1;

-- name: GetAllLLMServices :many
SELECT id, name, description, adapter, api_key, base_url, organization, created_at, updated_at, created_by, updated_by
FROM llm_services 
ORDER BY name;

-- name: CreateLLMService :one
INSERT INTO llm_services (name, description, adapter, api_key, base_url, organization, created_by, updated_by) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?) 
RETURNING id, name, description, adapter, api_key, base_url, organization, created_at, updated_at, created_by, updated_by;

-- name: UpdateLLMService :exec
UPDATE llm_services 
SET name = ?, description = ?, adapter = ?, api_key = ?, base_url = ?, organization = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? 
WHERE id = ?;

-- name: DeleteLLMService :exec
DELETE FROM llm_services 
WHERE id = ?; 