-- assistant_notes_search.sql
-- Search and related queries for assistant_notes
-- NOTE: FTS5 queries are in internal/brain/store/fts_queries.go (SQLC can't handle FTS5 virtual tables)

-- name: GetAssistantNoteByIDForRAG :one
-- Get a single assistant note with minimal fields for RAG context
SELECT 
    id,
    title,
    body,
    note_type,
    related_note_id,
    created_by_assistant_id,
    created_at
FROM assistant_notes
WHERE id = :id;

-- name: GetRelatedAssistantNotesByForwardLinks :many
-- Find assistant notes linked from this note (forward links)
SELECT DISTINCT
    an.id,
    an.title,
    substr(an.body, 1, 200) as snippet,
    an.note_type,
    an.created_at
FROM assistant_notes an
JOIN assistant_note_links anl ON an.id = anl.dest_note_id
WHERE anl.src_note_id = :note_id
LIMIT :limit_count;

-- name: GetRelatedAssistantNotesByBackwardLinks :many
-- Find assistant notes linking to this note (backward links)
SELECT DISTINCT
    an.id,
    an.title,
    substr(an.body, 1, 200) as snippet,
    an.note_type,
    an.created_at
FROM assistant_notes an
JOIN assistant_note_links anl ON an.id = anl.src_note_id
WHERE anl.dest_note_id = :note_id
LIMIT :limit_count;

-- name: GetRelatedAssistantNotesByTags :many
-- Find assistant notes with shared tags
SELECT DISTINCT
    an.id,
    an.title,
    substr(an.body, 1, 200) as snippet,
    an.note_type,
    an.created_at,
    COUNT(DISTINCT ant2.tag_id) as shared_tags
FROM assistant_notes an
JOIN assistant_note_tags ant2 ON an.id = ant2.assistant_note_id
JOIN assistant_note_tags ant1 ON ant1.tag_id = ant2.tag_id
WHERE ant1.assistant_note_id = :note_id
AND an.id != :note_id
GROUP BY an.id, an.title, an.body, an.note_type, an.created_at
ORDER BY shared_tags DESC
LIMIT :limit_count;

-- name: GetRelatedAssistantNotesBySourceNote :many
-- Find all observations related to the same source Mind note
SELECT 
    an.id,
    an.title,
    substr(an.body, 1, 200) as snippet,
    an.note_type,
    an.created_at
FROM assistant_notes an
WHERE an.related_note_id = :related_note_id
AND an.id != :note_id
LIMIT :limit_count;
