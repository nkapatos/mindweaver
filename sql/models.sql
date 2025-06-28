-- name: CreateModel :one
INSERT INTO models (
    llm_service_id, model_id, name, provider, description, created_at, owned_by
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetModelsByLLMServiceID :many
SELECT * FROM models 
WHERE llm_service_id = ? 
ORDER BY name;

-- name: GetModelByID :one
SELECT * FROM models WHERE id = ?;

-- name: GetModelByServiceAndModelID :one
SELECT * FROM models 
WHERE llm_service_id = ? AND model_id = ?;

-- name: UpdateModel :one
UPDATE models SET
    name = ?,
    provider = ?,
    description = ?,
    created_at = ?,
    owned_by = ?,
    last_fetched_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteModel :exec
DELETE FROM models WHERE id = ?;

-- name: DeleteModelsByLLMServiceID :exec
DELETE FROM models WHERE llm_service_id = ?;

-- name: GetModelsLastFetchedBefore :many
SELECT * FROM models 
WHERE llm_service_id = ? AND last_fetched_at < ?
ORDER BY last_fetched_at;

-- name: UpdateLastFetched :exec
UPDATE models SET last_fetched_at = CURRENT_TIMESTAMP WHERE id = ?; 