-- name: CreateActor :one
INSERT INTO actors (type, name, display_name, avatar_url, is_active, metadata, created_by, updated_by) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?) 
RETURNING id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at, created_by, updated_by;

-- name: GetActorByID :one
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM actors 
WHERE id = ? 
LIMIT 1;

-- name: GetAllActors :many
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM actors 
ORDER BY name;

-- name: GetActorByName :one
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM actors 
WHERE name = ? AND type = ? 
LIMIT 1;

-- name: GetActorsByType :many
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM actors 
WHERE type = ? AND is_active = true 
ORDER BY name;

-- name: UpdateActor :exec
UPDATE actors 
SET type = ?, name = ?, display_name = ?, avatar_url = ?, is_active = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? 
WHERE id = ?;

-- name: DeleteActor :exec
DELETE FROM actors WHERE id = ?;
