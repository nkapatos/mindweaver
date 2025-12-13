-- note_meta.sql
-- Pass 4: CRUD + advanced queries for note_meta (SQLite, sqlc compatible)
-- sqlc annotations added for code generation
-- Timestamps managed by DB. Optimized with composite index (note_id, key).
-- Included: insert, select by id, select all, update by id, delete by id, advanced queries
-- UPSERT support added for bulk meta operations

-- name: ListNoteMetaByKey :many
SELECT * FROM note_meta WHERE key = :key;

-- name: ListNoteMetaByKeys :many
SELECT * FROM note_meta WHERE key IN (:keys);

-- name: ListNoteMetaByKeyValuePattern :many
SELECT * FROM note_meta WHERE key = :key AND value LIKE :value_pattern;

-- name: ListDistinctNoteMetaKeys :many
SELECT DISTINCT key FROM note_meta ORDER BY key;

-- name: CreateNoteMeta :execlastid
INSERT INTO note_meta (note_id, key, value)
VALUES (:note_id, :key, :value);

-- name: UpsertNoteMeta :exec
INSERT INTO note_meta (note_id, key, value)
VALUES (:note_id, :key, :value)
ON CONFLICT (note_id, key) 
DO UPDATE SET 
    value = excluded.value,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteNoteMetaByNoteID :exec
DELETE FROM note_meta WHERE note_id = :note_id;

-- name: GetNoteMetaByID :one
SELECT * FROM note_meta WHERE id = :id;

-- name: ListNoteMeta :many
SELECT * FROM note_meta ORDER BY id;

-- name: UpdateNoteMetaByID :exec
UPDATE note_meta
SET note_id = :note_id,
    key = :key,
    value = :value,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteNoteMetaByID :exec
DELETE FROM note_meta WHERE id = :id;

-- ========================================
-- Composite Queries - Note Meta with Notes
-- ========================================

-- name: GetNoteMetaByNoteID :many
SELECT * FROM note_meta WHERE note_id = :note_id ORDER BY key;

