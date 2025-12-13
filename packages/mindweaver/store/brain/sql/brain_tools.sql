-- brain_tools stores available tools that Brain can execute
-- These are seeded with core operations and can be dynamically queried
-- by the small model to match user intent to executable tools

-- name: GetToolByName :one
SELECT * FROM brain_tools WHERE name = ?;

-- name: GetToolByUUID :one
SELECT * FROM brain_tools WHERE uuid = ?;

-- name: GetAllTools :many
SELECT * FROM brain_tools ORDER BY name;

-- name: GetToolsByType :many
SELECT * FROM brain_tools WHERE tool_type = ? ORDER BY name;

-- name: SearchToolsByDescription :many
-- Brain's LLM will use tool descriptions to select appropriate tools
SELECT * FROM brain_tools 
WHERE description LIKE '%' || ? || '%'
ORDER BY name;

-- name: CreateTool :one
INSERT INTO brain_tools (
    uuid,
    name,
    description,
    tool_type,
    endpoint,
    parameters
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTool :exec
UPDATE brain_tools 
SET description = ?,
    tool_type = ?,
    endpoint = ?,
    parameters = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = ?;

-- name: DeleteTool :exec
DELETE FROM brain_tools WHERE uuid = ?;

-- name: CountTools :one
SELECT COUNT(*) FROM brain_tools;
