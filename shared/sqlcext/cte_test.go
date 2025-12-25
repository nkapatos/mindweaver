package sqlcext

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupCTETestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
		CREATE TABLE collections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			parent_id INTEGER NULL,
			path TEXT NOT NULL UNIQUE,
			description TEXT,
			position INTEGER DEFAULT 0,
			is_system BOOLEAN DEFAULT 0 NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES collections(id) ON DELETE CASCADE,
			UNIQUE (parent_id, name)
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func insertTestCollection(t *testing.T, db *sql.DB, name, path string, parentID sql.NullInt64, isSystem bool) int64 {
	t.Helper()

	result, err := db.Exec(
		"INSERT INTO collections (name, path, parent_id, is_system) VALUES (?, ?, ?, ?)",
		name, path, parentID, isSystem,
	)
	if err != nil {
		t.Fatalf("failed to insert test collection: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	return id
}

func createTestCollectionHierarchy(t *testing.T, db *sql.DB) map[string]int64 {
	t.Helper()

	ids := make(map[string]int64)

	ids["root1"] = insertTestCollection(t, db, "root1", "root1", sql.NullInt64{}, false)
	ids["root2"] = insertTestCollection(t, db, "root2", "root2", sql.NullInt64{}, false)

	ids["root1_child1"] = insertTestCollection(t, db, "child1", "root1/child1", sql.NullInt64{Int64: ids["root1"], Valid: true}, false)
	ids["root1_child2"] = insertTestCollection(t, db, "child2", "root1/child2", sql.NullInt64{Int64: ids["root1"], Valid: true}, false)

	ids["root2_child1"] = insertTestCollection(t, db, "child1", "root2/child1", sql.NullInt64{Int64: ids["root2"], Valid: true}, false)

	ids["root1_child1_grandchild1"] = insertTestCollection(t, db, "grandchild1", "root1/child1/grandchild1", sql.NullInt64{Int64: ids["root1_child1"], Valid: true}, false)

	ids["root1_child1_grandchild1_greatgrandchild1"] = insertTestCollection(t, db, "greatgrandchild1", "root1/child1/grandchild1/greatgrandchild1", sql.NullInt64{Int64: ids["root1_child1_grandchild1"], Valid: true}, false)

	return ids
}

func TestGetCollectionTree_EmptyDB(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tree, err := querier.GetCollectionTree(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(tree) != 0 {
		t.Errorf("expected empty tree, got %d items", len(tree))
	}
}

func TestGetCollectionTree_SingleRoot(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	insertTestCollection(t, db, "root", "root", sql.NullInt64{}, false)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tree, err := querier.GetCollectionTree(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(tree))
	}

	if tree[0].Name != "root" {
		t.Errorf("expected name 'root', got '%s'", tree[0].Name)
	}

	if tree[0].Depth != 0 {
		t.Errorf("expected depth 0, got %d", tree[0].Depth)
	}

	if tree[0].ParentID.Valid {
		t.Errorf("expected null parent_id, got %d", tree[0].ParentID.Int64)
	}
}

func TestGetCollectionTree_MultipleRoots(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	insertTestCollection(t, db, "root1", "root1", sql.NullInt64{}, false)
	insertTestCollection(t, db, "root2", "root2", sql.NullInt64{}, false)
	insertTestCollection(t, db, "root3", "root3", sql.NullInt64{}, false)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tree, err := querier.GetCollectionTree(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(tree) != 3 {
		t.Fatalf("expected 3 collections, got %d", len(tree))
	}

	for _, col := range tree {
		if col.Depth != 0 {
			t.Errorf("expected all roots to have depth 0, got %d for %s", col.Depth, col.Name)
		}
		if col.ParentID.Valid {
			t.Errorf("expected null parent_id for root %s, got %d", col.Name, col.ParentID.Int64)
		}
	}
}

func TestGetCollectionTree_DeepHierarchy(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	ids := createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tree, err := querier.GetCollectionTree(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(tree) != 7 {
		t.Fatalf("expected 7 collections, got %d", len(tree))
	}

	depthMap := make(map[int64]int)
	for _, col := range tree {
		depthMap[col.ID] = col.Depth
	}

	if depthMap[ids["root1"]] != 0 {
		t.Errorf("expected root1 depth 0, got %d", depthMap[ids["root1"]])
	}
	if depthMap[ids["root1_child1"]] != 1 {
		t.Errorf("expected root1_child1 depth 1, got %d", depthMap[ids["root1_child1"]])
	}
	if depthMap[ids["root1_child1_grandchild1"]] != 2 {
		t.Errorf("expected root1_child1_grandchild1 depth 2, got %d", depthMap[ids["root1_child1_grandchild1"]])
	}
	if depthMap[ids["root1_child1_grandchild1_greatgrandchild1"]] != 3 {
		t.Errorf("expected greatgrandchild1 depth 3, got %d", depthMap[ids["root1_child1_grandchild1_greatgrandchild1"]])
	}
}

func TestGetCollectionTree_DepthLimit(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tests := []struct {
		maxDepth int
		expected int
	}{
		{0, 2},
		{1, 5},
		{2, 6},
		{3, 7},
		{10, 7},
	}

	for _, tt := range tests {
		t.Run("depth_"+string(rune(tt.maxDepth+'0')), func(t *testing.T) {
			tree, err := querier.GetCollectionTree(ctx, tt.maxDepth)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(tree) != tt.expected {
				t.Errorf("maxDepth=%d: expected %d collections, got %d", tt.maxDepth, tt.expected, len(tree))
			}

			for _, col := range tree {
				if col.Depth > tt.maxDepth {
					t.Errorf("maxDepth=%d: found collection at depth %d", tt.maxDepth, col.Depth)
				}
			}
		})
	}
}

func TestGetCollectionTree_OrderByPath(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tree, err := querier.GetCollectionTree(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for i := 1; i < len(tree); i++ {
		if tree[i-1].Path >= tree[i].Path {
			t.Errorf("collections not ordered by path: %s >= %s", tree[i-1].Path, tree[i].Path)
		}
	}
}

func TestGetCollectionSubtree_LeafNode(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	ids := createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	subtree, err := querier.GetCollectionSubtree(ctx, ids["root1_child1_grandchild1_greatgrandchild1"], 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(subtree) != 1 {
		t.Fatalf("expected 1 collection (leaf), got %d", len(subtree))
	}

	if subtree[0].Depth != 0 {
		t.Errorf("expected depth 0 for subtree root, got %d", subtree[0].Depth)
	}
}

func TestGetCollectionSubtree_WithChildren(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	ids := createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	subtree, err := querier.GetCollectionSubtree(ctx, ids["root1"], 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(subtree) != 5 {
		t.Fatalf("expected 5 collections, got %d", len(subtree))
	}

	if subtree[0].ID != ids["root1"] {
		t.Errorf("expected first item to be root1")
	}
	if subtree[0].Depth != 0 {
		t.Errorf("expected depth 0 for subtree root, got %d", subtree[0].Depth)
	}
}

func TestGetCollectionSubtree_DeepBranch(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	ids := createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	subtree, err := querier.GetCollectionSubtree(ctx, ids["root1_child1"], 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(subtree) != 3 {
		t.Fatalf("expected 3 collections, got %d", len(subtree))
	}

	depthMap := make(map[int64]int)
	for _, col := range subtree {
		depthMap[col.ID] = col.Depth
	}

	if depthMap[ids["root1_child1"]] != 0 {
		t.Errorf("expected root1_child1 depth 0 (subtree root), got %d", depthMap[ids["root1_child1"]])
	}
	if depthMap[ids["root1_child1_grandchild1"]] != 1 {
		t.Errorf("expected grandchild1 depth 1, got %d", depthMap[ids["root1_child1_grandchild1"]])
	}
	if depthMap[ids["root1_child1_grandchild1_greatgrandchild1"]] != 2 {
		t.Errorf("expected greatgrandchild1 depth 2, got %d", depthMap[ids["root1_child1_grandchild1_greatgrandchild1"]])
	}
}

func TestGetCollectionSubtree_DepthLimit(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	ids := createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	tests := []struct {
		maxDepth int
		expected int
	}{
		{0, 1},
		{1, 2},
		{2, 3},
		{10, 3},
	}

	for _, tt := range tests {
		t.Run("depth_"+string(rune(tt.maxDepth+'0')), func(t *testing.T) {
			subtree, err := querier.GetCollectionSubtree(ctx, ids["root1_child1"], tt.maxDepth)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(subtree) != tt.expected {
				t.Errorf("maxDepth=%d: expected %d collections, got %d", tt.maxDepth, tt.expected, len(subtree))
			}

			for _, col := range subtree {
				if col.Depth > tt.maxDepth {
					t.Errorf("maxDepth=%d: found collection at depth %d", tt.maxDepth, col.Depth)
				}
			}
		})
	}
}

func TestGetCollectionSubtree_InvalidID(t *testing.T) {
	db := setupCTETestDB(t)
	defer db.Close()

	createTestCollectionHierarchy(t, db)

	querier := NewCTEQuerier(db)
	ctx := context.Background()

	subtree, err := querier.GetCollectionSubtree(ctx, 99999, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(subtree) != 0 {
		t.Errorf("expected empty subtree for invalid ID, got %d items", len(subtree))
	}
}
