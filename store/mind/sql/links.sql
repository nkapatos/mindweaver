-- name: CreateLink :execlastid
INSERT INTO links (src_id, dest_id, display_text, is_embed)
VALUES (:src_id, :dest_id, :display_text, :is_embed);

-- name: CreateUnresolvedLink :execlastid
INSERT INTO links (
    src_id, dest_id, dest_title, display_text, is_embed, resolved
)
VALUES (:src_id, NULL, :dest_title, :display_text, :is_embed, 0);

-- name: GetLinkByID :one
SELECT * FROM links WHERE id = :id;

-- name: ListLinks :many
SELECT * FROM links ORDER BY id;

-- name: ListLinksBySrcID :many
SELECT * FROM links WHERE src_id = :src_id;

-- name: ListLinksByDestID :many
SELECT * FROM links WHERE dest_id = :dest_id;

-- name: SearchLinksByDisplayText :many
SELECT * FROM links WHERE display_text LIKE :display_text_pattern;

-- name: DeleteLinksBySrcID :exec
DELETE FROM links WHERE src_id = :src_id;

-- ========================================
-- WikiLink Resolution Queries
-- ========================================

-- name: ListUnresolvedLinks :many
-- Include pending (0) and broken (-1) links
SELECT * FROM links
WHERE resolved IN (0, -1)
ORDER BY id
LIMIT :limit;

-- name: FindUnresolvedLinksByDestTitle :many
-- Unresolved links by destination title (for resolution)
SELECT * FROM links
WHERE
    resolved = 0
    AND dest_id IS NULL
    AND dest_title = :dest_title;

-- name: CountUnresolvedLinks :one
SELECT COUNT(*) FROM links WHERE resolved = 0;

-- name: ResolveLink :exec
-- Set destination and mark resolved
UPDATE links
SET dest_id = :dest_id,
    dest_title = NULL,
    resolved = 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: MarkLinkBroken :exec
UPDATE links
SET resolved = -1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: ListBrokenLinks :many
SELECT * FROM links WHERE resolved = -1 ORDER BY id;

-- name: CountBrokenLinks :one
SELECT COUNT(*) FROM links WHERE resolved = -1;

-- name: ListOrphanedLinks :many
-- Destination note missing (dest_id IS NULL)
SELECT * FROM links
WHERE dest_id IS NULL
ORDER BY src_id, dest_title ;
