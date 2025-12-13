-- assistant_note_meta.sql
-- CRUD operations for assistant_note_meta table
-- EAV pattern for flexible key-value pairs on assistant notes

-- name: CreateAssistantNoteMeta :exec
INSERT INTO assistant_note_meta (assistant_note_id, key, value)
VALUES (?, ?, ?)
ON CONFLICT(assistant_note_id, key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP;

-- name: GetAssistantNoteMetaByID :one
SELECT * FROM assistant_note_meta WHERE id = :id;

-- name: GetAssistantNoteMetaByKey :one
SELECT * FROM assistant_note_meta 
WHERE assistant_note_id = :assistant_note_id AND key = :key 
LIMIT 1;

-- name: ListAssistantNoteMetaByNoteID :many
SELECT * FROM assistant_note_meta 
WHERE assistant_note_id = :assistant_note_id 
ORDER BY key;

-- name: ListAssistantNotesByMetaKey :many
SELECT DISTINCT an.* FROM assistant_notes an
JOIN assistant_note_meta anm ON an.id = anm.assistant_note_id
WHERE anm.key = :key
ORDER BY an.created_at DESC;

-- name: UpdateAssistantNoteMetaByID :exec
UPDATE assistant_note_meta
SET key = :key,
    value = :value,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteAssistantNoteMetaByID :exec
DELETE FROM assistant_note_meta WHERE id = :id;

-- name: DeleteAssistantNoteMetaByKey :exec
DELETE FROM assistant_note_meta 
WHERE assistant_note_id = :assistant_note_id AND key = :key;
