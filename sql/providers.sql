-- name: GetProviderByName :one
SELECT id, name, type, is_active, created_at 
FROM providers 
WHERE name = ? AND is_active = true
LIMIT 1;

-- name: GetAllProviders :many
SELECT id, name, type, is_active, created_at 
FROM providers 
WHERE is_active = true
ORDER BY name;

-- name: GetProviderSettings :many
SELECT setting_key, setting_value, is_secret 
FROM provider_settings 
WHERE provider_id = ?;

-- name: GetModelByName :one
SELECT id, provider_id, name, capabilities, default_params, is_active, created_at
FROM models 
WHERE provider_id = ? AND name = ? AND is_active = true
LIMIT 1;

-- name: GetModelsByProvider :many
SELECT id, provider_id, name, capabilities, default_params, is_active, created_at
FROM models 
WHERE provider_id = ? AND is_active = true
ORDER BY name;

-- name: UpdateProviderSetting :exec
UPDATE provider_settings 
SET setting_value = ? 
WHERE provider_id = ? AND setting_key = ?;

-- name: CreateProviderSetting :exec
INSERT INTO provider_settings (provider_id, setting_key, setting_value, is_secret) 
VALUES (?, ?, ?, ?);