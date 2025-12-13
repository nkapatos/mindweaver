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
-- Gets both pending (0) and broken (-1) links for resolution
-- Broken links can become resolved if the target note is created later
SELECT * FROM links
WHERE resolved IN (0, -1)
ORDER BY id
LIMIT :limit;

-- name: FindUnresolvedLinksByDestTitle :many
-- Finds unresolved links that point to a specific note title
-- Used when creating a note to resolve pending links
SELECT * FROM links
WHERE
    resolved = 0
    AND dest_id IS NULL
    AND dest_title = :dest_title;

-- name: CountUnresolvedLinks :one
SELECT COUNT(*) FROM links WHERE resolved = 0;

-- name: ResolveLink :exec
-- Resolves a pending link by setting dest_id and clearing dest_title
-- dest_title is only kept for unresolved (0) and broken (-1) links
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
-- Returns links where destination note no longer exists (dest_id IS NULL)
-- Used for "broken links" UI view (BR-03: Knowledge Preservation)
SELECT * FROM links
WHERE dest_id IS NULL
ORDER BY src_id, dest_title ;
