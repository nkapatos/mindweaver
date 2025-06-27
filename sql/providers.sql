-- name: GetProviderByName :one
SELECT p.id, p.name, p.description, p.llm_service_id, p.system_prompt, p.created_at,
       ls.name as llm_service_name, ls.adapter, ls.base_url, ls.organization, ls.configuration
FROM providers p
JOIN llm_services ls ON p.llm_service_id = ls.id
WHERE p.name = ?
LIMIT 1;

-- name: GetProviderByID :one
SELECT p.id, p.name, p.description, p.llm_service_id, p.system_prompt, p.created_at,
       ls.name as llm_service_name, ls.adapter, ls.base_url, ls.organization, ls.configuration
FROM providers p
JOIN llm_services ls ON p.llm_service_id = ls.id
WHERE p.id = ?;

-- name: GetAllProviders :many
SELECT p.id, p.name, p.description, p.llm_service_id, p.system_prompt, p.created_at,
       ls.name as llm_service_name, ls.adapter, ls.base_url, ls.organization, ls.configuration
FROM providers p
JOIN llm_services ls ON p.llm_service_id = ls.id
ORDER BY p.name;

-- name: CreateProvider :one
INSERT INTO providers (name, description, llm_service_id, system_prompt) 
VALUES (?, ?, ?, ?) 
RETURNING id, name, description, llm_service_id, system_prompt, created_at;

-- name: UpdateProvider :exec
UPDATE providers 
SET name = ?, description = ?, llm_service_id = ?, system_prompt = ? 
WHERE id = ?;

-- name: DeleteProvider :exec
DELETE FROM providers 
WHERE id = ?;

-- name: GetProvidersByLLMService :many
SELECT p.id, p.name, p.description, p.llm_service_id, p.system_prompt, p.created_at,
       ls.name as llm_service_name, ls.adapter, ls.base_url, ls.organization, ls.configuration
FROM providers p
JOIN llm_services ls ON p.llm_service_id = ls.id
WHERE p.llm_service_id = ?
ORDER BY p.name;