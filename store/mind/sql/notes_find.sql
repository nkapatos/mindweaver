-- Find notes by properties (title, collection, type, template)
-- For content search, see notes_search.sql (FTS5)
-- Design: AIP-136 :find pattern; global by default, optional filters; includes collection_path
-- Future: tag filters (AND/OR), date ranges, fuzzy matching
-- name: FindNotes :many
-- Global by default; optional filters; includes collection_path
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
  (sqlc.narg(title) IS NULL OR sqlc.narg(title) = '' OR n.title LIKE '%' || sqlc.narg(title) || '%')
  AND (sqlc.narg(collection_id) IS NULL OR n.collection_id = sqlc.narg(collection_id))
  AND (sqlc.narg(note_type_id) IS NULL OR n.note_type_id = sqlc.narg(note_type_id))
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
