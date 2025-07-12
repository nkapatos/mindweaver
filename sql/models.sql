-- name: CreateModel :one
INSERT INTO models (
    llm_service_id, model_id, name, provider, description, created_at, owned_by, created_by, updated_by
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, llm_service_id, model_id, name, provider, description, created_at, owned_by, last_fetched_at, created_by, updated_by;

-- name: GetModelsByLLMServiceID :many
SELECT id, llm_service_id, model_id, name, provider, description, created_at, owned_by, last_fetched_at, created_by, updated_by FROM models 
WHERE llm_service_id = ? 
ORDER BY name;

-- name: GetModelByID :one
SELECT id, llm_service_id, model_id, name, provider, description, created_at, owned_by, last_fetched_at, created_by, updated_by FROM models WHERE id = ?;

-- name: GetModelByServiceAndModelID :one
SELECT id, llm_service_id, model_id, name, provider, description, created_at, owned_by, last_fetched_at, created_by, updated_by FROM models 
WHERE llm_service_id = ? AND model_id = ?;

-- name: UpdateModel :one
UPDATE models SET
    name = ?,
    provider = ?,
    description = ?,
    created_at = ?,
    owned_by = ?,
    last_fetched_at = CURRENT_TIMESTAMP,
    updated_by = ?
WHERE id = ?
RETURNING id, llm_service_id, model_id, name, provider, description, created_at, owned_by, last_fetched_at, created_by, updated_by;

-- name: DeleteModel :exec
DELETE FROM models WHERE id = ?;

-- name: DeleteModelsByLLMServiceID :exec
DELETE FROM models WHERE llm_service_id = ?;

-- name: GetModelsLastFetchedBefore :many
SELECT id, llm_service_id, model_id, name, provider, description, created_at, owned_by, last_fetched_at, created_by, updated_by FROM models 
WHERE llm_service_id = ? AND last_fetched_at < ?
ORDER BY last_fetched_at;

-- name: UpdateLastFetched :exec
UPDATE models SET last_fetched_at = CURRENT_TIMESTAMP WHERE id = ?; 