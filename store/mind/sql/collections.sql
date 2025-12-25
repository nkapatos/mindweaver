-- collections.sql
-- CRUD operations for collections (hierarchical folders/paths for notes)

-- name: CreateCollection :execlastid
INSERT INTO collections (name, parent_id, path, description, position, is_system)
VALUES (:name, :parent_id, :path, :description, :position, :is_system);

-- name: GetCollectionByID :one
SELECT * FROM collections WHERE id = :id;

-- name: GetCollectionByPath :one
SELECT * FROM collections WHERE path = :path LIMIT 1;

-- name: ListCollections :many
SELECT * FROM collections ORDER BY path;

-- name: ListCollectionsByParent :many
SELECT * FROM collections 
WHERE parent_id = :parent_id 
ORDER BY position, name;

-- name: ListRootCollections :many
SELECT * FROM collections 
WHERE parent_id IS NULL AND id != 1
ORDER BY position, name;

-- name: UpdateCollection :exec
UPDATE collections
SET name = :name,
    parent_id = :parent_id,
    path = :path,
    description = :description,
    position = :position,
    is_system = :is_system,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteCollection :exec
DELETE FROM collections WHERE id = :id;

-- name: GetCollectionChildren :many
SELECT * FROM collections
WHERE parent_id = :parent_id
ORDER BY position, name;

-- name: GetCollectionAncestors :many
-- Get all ancestors of a collection by walking up the parent_id chain
WITH RECURSIVE ancestors AS (
  -- Start with the collection itself
  SELECT c.id, c.name, c.parent_id, c.path, 0 as depth
  FROM collections c
  WHERE c.id = :collection_id
  
  UNION ALL
  
  -- Recursively get parent
  SELECT c.id, c.name, c.parent_id, c.path, a.depth + 1
  FROM collections c
  JOIN ancestors a ON c.id = a.parent_id
)
SELECT id, name, parent_id, path, depth FROM ancestors WHERE depth > 0 ORDER BY depth DESC;

-- name: GetCollectionDescendants :many
-- Get all descendants of a collection by walking down the parent_id chain
WITH RECURSIVE descendants AS (
  -- Start with the collection itself
  SELECT c.id, c.name, c.parent_id, c.path, 0 as depth
  FROM collections c
  WHERE c.id = :collection_id
  
  UNION ALL
  
  -- Recursively get children
  SELECT c.id, c.name, c.parent_id, c.path, d.depth + 1
  FROM collections c
  JOIN descendants d ON c.parent_id = d.id
)
SELECT id, name, parent_id, path, depth FROM descendants WHERE depth > 0 ORDER BY path;

-- name: CountNotesInCollection :one
SELECT COUNT(*) as count
FROM notes
WHERE collection_id = :collection_id;

-- name: FindOrCreateCollectionByPath :one
-- This is a helper query to find existing collection by path
-- or return NULL if not found (actual creation happens in Go code)
SELECT * FROM collections WHERE path = :path LIMIT 1;

-- ========================================
-- System Collection Protection
-- ========================================

-- name: CheckIfSystemCollection :one
SELECT is_system FROM collections WHERE id = :id;

-- name: GetCollectionStats :many
SELECT 
    c.id,
    c.name,
    c.path,
    c.parent_id,
    COUNT(n.id) as notes_count
FROM collections c
LEFT JOIN notes n ON c.id = n.collection_id
GROUP BY c.id
ORDER BY c.path;

-- ========================================
-- Paginated Queries (AIP-158)
-- ========================================

-- name: ListCollectionsPaginated :many
SELECT * FROM collections 
ORDER BY id
LIMIT :limit OFFSET :offset;

-- name: CountCollections :one
SELECT COUNT(*) FROM collections;

-- name: ListCollectionsByParentPaginated :many
SELECT * FROM collections 
WHERE parent_id = :parent_id 
ORDER BY position, name
LIMIT :limit OFFSET :offset;

-- name: CountCollectionsByParent :one
SELECT COUNT(*) FROM collections 
WHERE parent_id = :parent_id;
