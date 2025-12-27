-- Notes search and related queries
-- NOTE: FTS5 queries are in internal/mind/store/fts_queries.go (sqlc can't handle virtual tables)

-- name: GetNoteByIDForRAG :one
-- Get a single note with minimal fields for RAG context
SELECT 
    id,
    title,
    body,
    note_type_id,
    created_at
FROM notes
WHERE id = :id;

-- name: GetRelatedNotesByForwardLinks :many
-- Notes linked from this note (forward)
SELECT DISTINCT
    n.id,
    n.title,
    substr(n.body, 1, 200) as snippet,
    n.note_type_id,
    n.created_at
FROM notes n
JOIN links nl ON n.id = nl.dest_id
WHERE nl.src_id = :note_id
LIMIT :limit_count;

-- name: GetRelatedNotesByBackwardLinks :many
-- Notes linking to this note (backward)
SELECT DISTINCT
    n.id,
    n.title,
    substr(n.body, 1, 200) as snippet,
    n.note_type_id,
    n.created_at
FROM notes n
JOIN links nl ON n.id = nl.src_id
WHERE nl.dest_id = :note_id
LIMIT :limit_count;

-- name: GetRelatedNotesByTags :many
-- Notes sharing tags with the given note
SELECT DISTINCT
    n.id,
    n.title,
    substr(n.body, 1, 200) as snippet,
    n.note_type_id,
    n.created_at,
    COUNT(DISTINCT nt2.tag_id) as shared_tags
FROM notes n
JOIN note_tags nt2 ON n.id = nt2.note_id
JOIN note_tags nt1 ON nt1.tag_id = nt2.tag_id
WHERE nt1.note_id = :note_id
AND n.id != :note_id
GROUP BY n.id, n.title, n.body, n.note_type_id, n.created_at
ORDER BY shared_tags DESC
LIMIT :limit_count;
