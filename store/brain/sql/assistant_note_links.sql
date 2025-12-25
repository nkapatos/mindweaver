-- assistant_note_links.sql
-- CRUD operations for assistant_note_links table
-- Enables Brain to create knowledge graphs by linking its own notes

-- name: CreateAssistantNoteLink :one
INSERT INTO assistant_note_links (
    src_note_id,
    dest_note_id,
    link_type,
    context,
    created_by_assistant_id,
    strength
) VALUES (
    :src_note_id,
    :dest_note_id,
    :link_type,
    :context,
    :created_by_assistant_id,
    :strength
)
RETURNING *;

-- name: GetAssistantNoteLinkByID :one
SELECT * FROM assistant_note_links WHERE id = :id;

-- name: GetForwardLinks :many
-- Get all notes linked FROM this note
SELECT 
    anl.*,
    an.id as dest_note_id,
    an.title as dest_note_title,
    an.note_type as dest_note_type
FROM assistant_note_links anl
JOIN assistant_notes an ON anl.dest_note_id = an.id
WHERE anl.src_note_id = :src_note_id
ORDER BY anl.created_at DESC;

-- name: GetBacklinks :many
-- Get all notes linked TO this note
SELECT 
    anl.*,
    an.id as src_note_id,
    an.title as src_note_title,
    an.note_type as src_note_type
FROM assistant_note_links anl
JOIN assistant_notes an ON anl.src_note_id = an.id
WHERE anl.dest_note_id = :dest_note_id
ORDER BY anl.created_at DESC;

-- name: GetBidirectionalLinks :many
-- Get all notes connected to this note (both directions)
SELECT DISTINCT
    CASE 
        WHEN anl.src_note_id = :note_id THEN anl.dest_note_id
        ELSE anl.src_note_id
    END as connected_note_id,
    an.title as connected_note_title,
    an.note_type as connected_note_type,
    anl.link_type,
    anl.context,
    anl.strength,
    CASE 
        WHEN anl.src_note_id = :note_id THEN 'forward'
        ELSE 'backward'
    END as direction
FROM assistant_note_links anl
JOIN assistant_notes an ON (
    CASE 
        WHEN anl.src_note_id = :note_id THEN anl.dest_note_id
        ELSE anl.src_note_id
    END = an.id
)
WHERE anl.src_note_id = :note_id OR anl.dest_note_id = :note_id
ORDER BY anl.created_at DESC;

-- name: GetLinksByType :many
-- Get all links of a specific type
SELECT 
    anl.*,
    src.title as src_note_title,
    dest.title as dest_note_title
FROM assistant_note_links anl
JOIN assistant_notes src ON anl.src_note_id = src.id
JOIN assistant_notes dest ON anl.dest_note_id = dest.id
WHERE anl.link_type = :link_type
ORDER BY anl.created_at DESC;

-- name: GetLinksByAssistant :many
-- Get all links created by a specific assistant
SELECT 
    anl.*,
    src.title as src_note_title,
    dest.title as dest_note_title
FROM assistant_note_links anl
JOIN assistant_notes src ON anl.src_note_id = src.id
JOIN assistant_notes dest ON anl.dest_note_id = dest.id
WHERE anl.created_by_assistant_id = :created_by_assistant_id
ORDER BY anl.created_at DESC;

-- name: UpdateAssistantNoteLink :exec
UPDATE assistant_note_links
SET link_type = :link_type,
    context = :context,
    strength = :strength,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteAssistantNoteLink :exec
DELETE FROM assistant_note_links WHERE id = :id;

-- name: DeleteAssistantNoteLinksBetween :exec
-- Delete all links between two specific notes
DELETE FROM assistant_note_links 
WHERE (src_note_id = :note_id_1 AND dest_note_id = :note_id_2)
   OR (src_note_id = :note_id_2 AND dest_note_id = :note_id_1);

-- name: DeleteAllLinksForNote :exec
-- Delete all links involving a note (used when deleting a note)
DELETE FROM assistant_note_links 
WHERE src_note_id = :note_id OR dest_note_id = :note_id;

-- name: CountLinksForNote :one
-- Count how many links a note has (useful for finding "hub" notes)
SELECT COUNT(*) as link_count
FROM assistant_note_links
WHERE src_note_id = :note_id OR dest_note_id = :note_id;

-- name: FindStronglyConnectedNotes :many
-- Find notes with many connections (knowledge hubs)
SELECT 
    an.id,
    an.title,
    an.note_type,
    COUNT(anl.id) as link_count
FROM assistant_notes an
LEFT JOIN assistant_note_links anl ON (an.id = anl.src_note_id OR an.id = anl.dest_note_id)
GROUP BY an.id
HAVING link_count >= :min_links
ORDER BY link_count DESC
LIMIT :limit_count;
