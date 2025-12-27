-- Note Types: CRUD and filters (SQLite/sqlc)

-- name: CreateNoteType :execlastid
INSERT INTO note_types (type, name, description, icon, color, is_system, created_at, updated_at)
VALUES (:type, :name, :description, :icon, :color, :is_system, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: GetNoteTypeByID :one
SELECT * FROM note_types WHERE id = :id;

-- name: ListNoteTypes :many
SELECT * FROM note_types ORDER BY id;

-- name: GetNoteTypeByType :one
SELECT * FROM note_types WHERE type = :type;

-- name: GetNoteTypesByTypes :many
SELECT * FROM note_types WHERE type IN (:types);

-- name: UpdateNoteTypeByID :exec
UPDATE note_types
SET type = :type,
    name = :name,
    description = :description,
    icon = :icon,
    color = :color,
    is_system = :is_system,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteNoteTypeByID :exec
DELETE FROM note_types WHERE id = :id;

-- ========================================
-- Composite Queries - Note Types with Notes
-- ========================================

-- name: GetNoteTypeWithNotesCount :one
SELECT 
    nt.*,
    COUNT(n.id) as notes_count
FROM note_types nt
LEFT JOIN notes n ON nt.id = n.note_type_id
WHERE nt.id = :id
GROUP BY nt.id;

-- name: ListNoteTypesWithCount :many
SELECT 
    nt.*,
    COUNT(n.id) as notes_count
FROM note_types nt
LEFT JOIN notes n ON nt.id = n.note_type_id
GROUP BY nt.id
ORDER BY nt.id;

-- ========================================
-- System Type Protection
-- ========================================

-- name: CheckIfSystemType :one
SELECT is_system FROM note_types WHERE id = :id;

-- ========================================
-- Paginated Queries (AIP-158)
-- ========================================

-- name: ListNoteTypesPaginated :many
SELECT * FROM note_types 
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountNoteTypes :one
SELECT COUNT(*) FROM note_types;
