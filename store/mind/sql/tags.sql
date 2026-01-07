-- Tags: CRUD and relations; consider advanced suggestions and analytics
-- name: CreateTag :execlastid
INSERT INTO tags (name, created_at, updated_at)
VALUES (:name, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = :id;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = :name;

-- name: ListTags :many
SELECT * FROM tags ORDER BY id;

-- name: UpdateTagByID :exec
UPDATE tags
SET name = :name,
updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteTagByID :exec
DELETE FROM tags WHERE id = :id;

-- name: SearchTagsByName :many
SELECT * FROM tags WHERE name LIKE :name_pattern;

-- name: ListTagsForNote :many
SELECT tags.* FROM tags
JOIN note_tags ON tags.id = note_tags.tag_id
WHERE note_tags.note_id = :note_id;

-- name: ListNotesForTag :many
SELECT notes.* FROM notes
JOIN note_tags ON notes.id = note_tags.note_id
WHERE note_tags.tag_id = :tag_id;

-- name: TagUsageCount :many
SELECT tag_id,
COUNT(*) as usage_count FROM note_tags GROUP BY tag_id ORDER BY usage_count DESC;

-- name: CreateNoteTag :exec
INSERT INTO note_tags (note_id, tag_id)
VALUES (:note_id, :tag_id);

-- name: ListNoteTagsByNoteID :many
SELECT * FROM note_tags WHERE note_id = :note_id;

-- name: ListNoteTagsByTagID :many
SELECT * FROM note_tags WHERE tag_id = :tag_id;

-- name: ListNoteTags :many
SELECT * FROM note_tags ORDER BY note_id, tag_id ;

-- name: DeleteNoteTag :exec
DELETE FROM note_tags WHERE note_id = :note_id AND tag_id = :tag_id;

-- name: DeleteNoteTagsByNoteID :exec
DELETE FROM note_tags WHERE note_id = :note_id;

-- ========================================
-- Paginated Queries (AIP-158)
-- ========================================

-- name: ListTagsPaginated :many
SELECT * FROM tags
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountTags :one
SELECT COUNT(*) FROM tags;

-- name: ListTagsForNotePaginated :many
SELECT tags.* FROM tags
JOIN note_tags ON tags.id = note_tags.tag_id
WHERE note_tags.note_id = :note_id
ORDER BY tags.id
LIMIT :limit OFFSET :offset;

-- name: CountTagsForNote :one
SELECT COUNT(*) FROM tags
JOIN note_tags ON tags.id = note_tags.tag_id
WHERE note_tags.note_id = :note_id;

-- name: ListNotesForTagPaginated :many
SELECT notes.* FROM notes
JOIN note_tags ON notes.id = note_tags.note_id
WHERE note_tags.tag_id = :tag_id
ORDER BY notes.id
LIMIT :limit OFFSET :offset;

-- name: CountNotesForTag :one
SELECT COUNT(*) FROM notes
JOIN note_tags ON notes.id = note_tags.note_id
WHERE note_tags.tag_id = :tag_id;

-- name: FindTagsPaginated :many
SELECT * FROM tags
WHERE name LIKE :pattern
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountFindTags :one
SELECT COUNT(*) FROM tags
WHERE name LIKE :pattern;

