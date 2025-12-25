-- notes_find.sql
-- Find notes by metadata/properties (title, tags, collection, type)
-- Used by UI pickers and Brain service for structured queries
-- For content-based search, see notes_search.sql (FTS5)
--
-- Design: Supports AIP-136 :find custom method pattern
-- - Global search by default (no collection filter)
-- - Optional filters narrow scope (collection, type, template, tags)
-- - Always includes collection_path for "where is it?" UX
--
-- Future extensions:
-- - Tag filtering (AND/OR logic)
-- - Date range filtering (created_after, updated_before, etc.)
-- - Fuzzy matching mode

-- name: FindNotes :many
-- Find notes by title and optional filters, with collection path
-- Default behavior: searches globally across ALL collections
-- Optional filters: collection_id (scope narrowing), note_type_id, is_template
-- Title filter: case-insensitive substring match (SQLite LIKE is case-insensitive)
-- Returns: All note fields + collection_path from JOIN
-- Sort: exact title matches first (better UX), then by recency
SELECT 
  n.id,
  n.uuid,
  n.title,
  n.body,
  n.description,
  n.frontmatter,
  n.note_type_id,
  n.collection_id,
  n.is_template,
  n.version,
  n.created_at,
  n.updated_at,
  c.path as collection_path
FROM notes n
LEFT JOIN collections c ON n.collection_id = c.id
WHERE 
  -- Title filter (optional, empty/null = no filter)
  -- SQLite LIKE is case-insensitive by default
  (sqlc.narg(title) IS NULL OR sqlc.narg(title) = '' OR n.title LIKE '%' || sqlc.narg(title) || '%')
  
  -- Collection filter (optional, enables scope narrowing when context known)
  -- NULL = global search across all collections (default)
  AND (sqlc.narg(collection_id) IS NULL OR n.collection_id = sqlc.narg(collection_id))
  
  -- Note type filter (optional)
  AND (sqlc.narg(note_type_id) IS NULL OR n.note_type_id = sqlc.narg(note_type_id))
  
  -- Template filter (optional)
  AND (sqlc.narg(is_template) IS NULL OR n.is_template = sqlc.narg(is_template))
  
ORDER BY 
  n.updated_at DESC
LIMIT sqlc.arg(limit) 
OFFSET sqlc.arg(offset);

-- name: CountFindNotes :one
-- Count notes matching find criteria (same WHERE conditions as FindNotes)
-- Used for pagination total_size calculation
SELECT COUNT(*)
FROM notes n
WHERE 
  (sqlc.narg(title) IS NULL OR sqlc.narg(title) = '' OR n.title LIKE '%' || sqlc.narg(title) || '%')
  AND (sqlc.narg(collection_id) IS NULL OR n.collection_id = sqlc.narg(collection_id))
  AND (sqlc.narg(note_type_id) IS NULL OR n.note_type_id = sqlc.narg(note_type_id))
  AND (sqlc.narg(is_template) IS NULL OR n.is_template = sqlc.narg(is_template));
