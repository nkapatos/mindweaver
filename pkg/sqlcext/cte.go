package sqlcext

import (
	"context"
	"fmt"
)

type CTEQuerier struct {
	db           DB
	treeQuery    string
	subtreeQuery string
}

func NewCTEQuerier(db DB) *CTEQuerier {
	q := &CTEQuerier{
		db: db,
	}

	q.treeQuery = `
WITH RECURSIVE tree(id, name, parent_id, path, description, position, is_system, depth) AS (
  SELECT c.id, c.name, c.parent_id, c.path, c.description, c.position, c.is_system, 0
  FROM collections c
  WHERE c.parent_id IS NULL
  
  UNION ALL
  
  SELECT c.id, c.name, c.parent_id, c.path, c.description, c.position, c.is_system, tree.depth + 1
  FROM collections c, tree
  WHERE c.parent_id = tree.id AND tree.depth < ?
)
SELECT id, name, parent_id, path, description, position, is_system, depth FROM tree ORDER BY path`

	q.subtreeQuery = `
WITH RECURSIVE subtree(id, name, parent_id, path, description, position, is_system, depth) AS (
  SELECT c.id, c.name, c.parent_id, c.path, c.description, c.position, c.is_system, 0
  FROM collections c
  WHERE c.id = ?
  
  UNION ALL
  
  SELECT c.id, c.name, c.parent_id, c.path, c.description, c.position, c.is_system, subtree.depth + 1
  FROM collections c, subtree
  WHERE c.parent_id = subtree.id AND subtree.depth < ?
)
SELECT id, name, parent_id, path, description, position, is_system, depth FROM subtree ORDER BY path`

	return q
}

func (q *CTEQuerier) GetCollectionTree(ctx context.Context, maxDepth int) ([]CollectionTreeRow, error) {
	rows, err := q.db.QueryContext(ctx, q.treeQuery, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("collection tree query failed: %w", err)
	}
	defer rows.Close()

	var results []CollectionTreeRow
	for rows.Next() {
		var r CollectionTreeRow
		if err := rows.Scan(&r.ID, &r.Name, &r.ParentID, &r.Path, &r.Description, &r.Position, &r.IsSystem, &r.Depth); err != nil {
			return nil, fmt.Errorf("failed to scan collection tree row: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("collection tree iteration failed: %w", err)
	}

	return results, nil
}

func (q *CTEQuerier) GetCollectionSubtree(ctx context.Context, collectionID int64, maxDepth int) ([]CollectionTreeRow, error) {
	rows, err := q.db.QueryContext(ctx, q.subtreeQuery, collectionID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("collection subtree query failed: %w", err)
	}
	defer rows.Close()

	var results []CollectionTreeRow
	for rows.Next() {
		var r CollectionTreeRow
		if err := rows.Scan(&r.ID, &r.Name, &r.ParentID, &r.Path, &r.Description, &r.Position, &r.IsSystem, &r.Depth); err != nil {
			return nil, fmt.Errorf("failed to scan collection subtree row: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("collection subtree iteration failed: %w", err)
	}

	return results, nil
}
