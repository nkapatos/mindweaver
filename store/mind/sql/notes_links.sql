-- TODO: Add composite queries with note titles for display

-- name: ListNotesLinksBySrcID :many
SELECT * FROM notes_links WHERE src_id = :src_id;

-- name: ListNotesLinksByDestID :many
SELECT * FROM notes_links WHERE dest_id = :dest_id;

-- name: SearchNotesLinksByDisplayText :many
SELECT * FROM notes_links WHERE display_text LIKE :display_text_pattern;

-- name: CreateNotesLink :execlastid
INSERT INTO notes_links (src_id, dest_id, display_text, is_embed)
VALUES (:src_id, :dest_id, :display_text, :is_embed);

-- name: CreateUnresolvedNotesLink :execlastid
INSERT INTO notes_links (
    src_id, dest_id, dest_title, display_text, is_embed, resolved
)
VALUES (:src_id, NULL, :dest_title, :display_text, :is_embed, 0);

-- name: GetNotesLinkByID :one
SELECT * FROM notes_links WHERE id = :id;

-- name: ListNotesLinks :many
SELECT * FROM notes_links ORDER BY id;

-- name: DeleteNotesLinksBySrcID :exec
DELETE FROM notes_links WHERE src_id = :src_id;

-- ========================================
-- Composite Queries - Notes Links with Note Details
-- ========================================

-- name: GetNotesLinkWithNoteTitles :one
SELECT
    nl.*,
    src_note.title AS src_title,
    dest_note.title AS dest_title
FROM notes_links nl
JOIN notes src_note ON nl.src_id = src_note.id
JOIN notes dest_note ON nl.dest_id = dest_note.id
WHERE nl.id = :id;

-- ========================================
-- WikiLink Resolution Queries
-- ========================================

-- name: ListUnresolvedLinks :many
-- Gets both pending (0) and broken (-1) links for resolution
-- Broken links can become resolved if the target note is created later
SELECT * FROM notes_links
WHERE resolved IN (0, -1)
ORDER BY id
LIMIT :limit;

-- name: FindUnresolvedLinksByDestTitle :many
-- Finds unresolved links that point to a specific note title
-- Used when creating a note to resolve pending links
SELECT * FROM notes_links
WHERE
    resolved = 0
    AND dest_id IS NULL
    AND dest_title = :dest_title;

-- name: CountUnresolvedLinks :one
SELECT COUNT(*) FROM notes_links WHERE resolved = 0;

-- name: ResolveLink :exec
-- Resolves a pending link by setting dest_id and clearing dest_title
-- dest_title is only kept for unresolved (0) and broken (-1) links
UPDATE notes_links
SET dest_id = :dest_id,
    dest_title = NULL,
    resolved = 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: MarkLinkBroken :exec
UPDATE notes_links
SET resolved = -1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: ListBrokenLinks :many
SELECT * FROM notes_links WHERE resolved = -1 ORDER BY id;

-- name: CountBrokenLinks :one
SELECT COUNT(*) FROM notes_links WHERE resolved = -1;

-- ========================================
-- Additional Query Patterns (FR-LINKS-02)
-- ========================================

-- name: ListOrphanedLinks :many
-- Returns links where destination note no longer exists (dest_id IS NULL)
-- Used for "broken links" UI view (BR-03: Knowledge Preservation)
SELECT * FROM notes_links
WHERE dest_id IS NULL
ORDER BY src_id, dest_title ;

-- name: ListNotesLinksByNoteIDs :many
-- Batch query for multiple notes (graph view construction)
-- Example usage: GetLinks for notes 1,2,3 to build subgraph
SELECT * FROM notes_links
WHERE src_id IN (sqlc.slice ('note_ids'))
OR dest_id IN (sqlc.slice ('note_ids'))
ORDER BY src_id, dest_id ;

