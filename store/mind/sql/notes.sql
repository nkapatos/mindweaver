-- Notes: CRUD and composite queries (SQLite/sqlc)
-- NOTE: uuid uses uuidv7; ordering by uuid
-- name: CreateNote :execlastid
INSERT INTO notes (uuid, title, body, description, frontmatter, note_type_id, is_template, collection_id)
VALUES (:uuid, :title, :body, :description, :frontmatter, :note_type_id, :is_template, :collection_id);

-- name: GetNoteByID :one
SELECT * FROM notes WHERE id = :id;

-- name: GetNoteByUUID :one
SELECT * FROM notes WHERE uuid = :uuid;

-- name: GetNoteByTitle :one
SELECT * FROM notes WHERE title = :title AND collection_id = :collection_id LIMIT 1;

-- name: GetNoteByTitleGlobal :one
-- Global title lookup across collections
SELECT * FROM notes WHERE title = :title LIMIT 1;

-- name: ListNotes :many
SELECT * FROM notes ORDER BY uuid;

-- name: UpdateNoteByID :exec
UPDATE notes
SET uuid = :uuid,
    title = :title,
    body = :body,
    description = :description,
    frontmatter = :frontmatter,
    updated_at = CURRENT_TIMESTAMP,
    note_type_id = :note_type_id,
    is_template = :is_template,
    collection_id = :collection_id,
    version = version + 1
WHERE id = :id and version = :version;

-- name: DeleteNoteByID :exec
DELETE FROM notes WHERE id = :id;

-- ========================================
-- Composite Queries - Notes with Relations
-- ========================================

-- name: GetNoteWithMetaByID :one
SELECT 
    n.*,
    GROUP_CONCAT(nm.key || ':' || nm.value, '|') as meta_pairs
FROM notes n
LEFT JOIN note_meta nm ON n.id = nm.note_id
WHERE n.id = :id
GROUP BY n.id;

-- name: GetNoteWithTypeByID :one
SELECT 
    n.*,
    nt.id as type_id,
    nt.type as type_type,
    nt.name as type_name,
    nt.description as type_description,
    nt.icon as type_icon,
    nt.color as type_color
FROM notes n
LEFT JOIN note_types nt ON n.note_type_id = nt.id
WHERE n.id = :id;

-- ========================================
-- Filter Queries - Notes by Relations
-- ========================================

-- name: ListNotesByTagIDs :many
SELECT DISTINCT n.* FROM notes n
JOIN note_tags nt ON n.id = nt.note_id
WHERE nt.tag_id = ?1
ORDER BY n.uuid;

-- name: ListNotesByMetaKeys :many
SELECT DISTINCT n.* FROM notes n
JOIN note_meta nm ON n.id = nm.note_id
WHERE nm.key = ?1
ORDER BY n.uuid;

-- name: ListNotesByNoteTypeID :many
SELECT * FROM notes 
WHERE note_type_id = ?1
ORDER BY updated_at DESC;

-- ========================================
-- Composite Query - Note with All Relations
-- ========================================

-- name: GetNoteWithAllRelationsByID :one
SELECT 
    n.*,
    nt.id as type_id,
    nt.type as type_type,
    nt.name as type_name,
    nt.description as type_description,
    nt.icon as type_icon,
    nt.color as type_color,
    -- Meta as JSON object (SQLite specific)
    (
        SELECT GROUP_CONCAT('"' || nm.key || '":' || json_quote(nm.value), ',')
        FROM note_meta nm WHERE nm.note_id = n.id
    ) as meta_json,
    -- Tags as JSON array
    (
        SELECT GROUP_CONCAT(json_quote(t.name), ',')
        FROM tags t
        JOIN note_tags ntg ON t.id = ntg.tag_id
        WHERE ntg.note_id = n.id
    ) as tags_json
FROM notes n
LEFT JOIN note_types nt ON n.note_type_id = nt.id
WHERE n.id = :id;

-- ========================================
-- Collection Queries
-- ========================================

-- name: ListNotesByCollectionID :many
SELECT * FROM notes 
WHERE collection_id = :collection_id
ORDER BY title;

-- name: ListNotesByCollectionPath :many
SELECT n.* FROM notes n
INNER JOIN collections c ON n.collection_id = c.id
WHERE c.path = :path
ORDER BY n.title;

-- name: CountNotesByCollectionID :one
SELECT COUNT(*) FROM notes 
WHERE collection_id = :collection_id;

-- ========================================
-- Multi-Tag Filtering (FR-TAGS-02)
-- ========================================

-- name: ListNotesByTagIDsAND :many
-- Notes having ALL specified tags
SELECT n.* FROM notes n
JOIN note_tags nt ON n.id = nt.note_id
WHERE nt.tag_id IN (sqlc.slice('tag_ids'))
GROUP BY n.id
HAVING COUNT(DISTINCT nt.tag_id) = sqlc.arg('tag_count')
ORDER BY n.uuid;

-- name: ListNotesByTagIDsOR :many
-- Notes having ANY of the specified tags
SELECT DISTINCT n.* FROM notes n
JOIN note_tags nt ON n.id = nt.note_id
WHERE nt.tag_id IN (sqlc.slice('tag_ids'))
ORDER BY n.uuid;

-- name: GetNoteByTitleInCollection :one
-- Lookup note by title within a specific collection (titles unique per collection)
SELECT * FROM notes 
WHERE title = :title AND collection_id = :collection_id 
LIMIT 1;

-- ========================================
-- Paginated Queries (AIP-158)
-- ========================================

-- name: ListNotesPaginated :many
SELECT * FROM notes 
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountNotes :one
SELECT COUNT(*) FROM notes;

-- name: ListNotesByCollectionIDPaginated :many
SELECT * FROM notes 
WHERE collection_id = :collection_id
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: ListNotesByNoteTypeIDPaginated :many
SELECT * FROM notes 
WHERE note_type_id = :note_type_id
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountNotesByNoteTypeID :one
SELECT COUNT(*) FROM notes 
WHERE note_type_id = :note_type_id;

-- name: ListNotesByIsTemplatePaginated :many
SELECT * FROM notes 
WHERE is_template = :is_template
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountNotesByIsTemplate :one
SELECT COUNT(*) FROM notes 
WHERE is_template = :is_template;
