-- tags.sql
-- CRUD operations for tags table
-- Tags that assistant can use to organize notes

-- name: CreateTag :execlastid
INSERT INTO tags (name, created_at, updated_at) 
VALUES (:name, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = :id;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = :name LIMIT 1;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: DeleteTagByID :exec
DELETE FROM tags WHERE id = :id;

-- ========================================
-- Tag Assignment Operations
-- ========================================

-- name: AddTagToAssistantNote :exec
INSERT INTO assistant_note_tags (assistant_note_id, tag_id)
VALUES (:assistant_note_id, :tag_id)
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromAssistantNote :exec
DELETE FROM assistant_note_tags 
WHERE assistant_note_id = :assistant_note_id AND tag_id = :tag_id;

-- name: ListTagsByAssistantNote :many
SELECT t.* FROM tags t
JOIN assistant_note_tags ant ON t.id = ant.tag_id
WHERE ant.assistant_note_id = :assistant_note_id
ORDER BY t.name;

-- name: ListAssistantNotesByTag :many
SELECT an.* FROM assistant_notes an
JOIN assistant_note_tags ant ON an.id = ant.assistant_note_id
WHERE ant.tag_id = :tag_id
ORDER BY an.created_at DESC;
