-- name: CreateActor :one
INSERT INTO actors (type, name, display_name, avatar_url, is_active, metadata) 
VALUES (?, ?, ?, ?, ?, ?) 
RETURNING id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at;

-- name: GetActorByID :one
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at 
FROM actors 
WHERE id = ? 
LIMIT 1;

-- name: GetActorByName :one
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at 
FROM actors 
WHERE name = ? AND type = ? 
LIMIT 1;

-- name: GetActorsByType :many
SELECT id, type, name, display_name, avatar_url, is_active, metadata, created_at, updated_at 
FROM actors 
WHERE type = ? AND is_active = true 
ORDER BY name;

-- name: UpdateActor :exec
UPDATE actors 
SET type = ?, name = ?, display_name = ?, avatar_url = ?, is_active = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteActor :exec
DELETE FROM actors WHERE id = ?;
