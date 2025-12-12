-- templates.sql
-- Pass 3: CRUD for templates (SQLite, sqlc compatible)
-- sqlc annotations added for code generation
-- Timestamps managed by DB. TODO: Revisit note_type_id FK.
-- Included: insert, select by id, select all, update by id, delete by id
-- Note: Search queries removed - use general search with filters instead
-- Next: Add advanced queries if needed (filter by note_type, etc.)

-- name: CreateTemplate :execlastid
INSERT INTO templates (name, description, starter_note_id, note_type_id)
VALUES (:name, :description, :starter_note_id, :note_type_id);

-- name: GetTemplateByID :one
SELECT * FROM templates WHERE id = :id;

-- name: ListTemplates :many
SELECT * FROM templates ORDER BY id;

-- name: UpdateTemplateByID :exec
UPDATE templates
SET name = :name,
    description = :description,
    starter_note_id = :starter_note_id,
    note_type_id = :note_type_id,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteTemplateByID :exec
DELETE FROM templates WHERE id = :id;

-- ========================================
-- Paginated Queries (AIP-158)
-- ========================================

-- name: ListTemplatesPaginated :many
SELECT * FROM templates 
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountTemplates :one
SELECT COUNT(*) FROM templates;
