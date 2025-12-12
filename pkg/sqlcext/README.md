# sqlcext - Manual SQL Queries for sqlc Limitations

## Purpose

This package contains **hand-written SQL queries** for database features that **sqlc cannot parse or generate code for**.

## Critical: sqlc Limitations

### What sqlc CANNOT Handle

1. **FTS5 (Full-Text Search) Virtual Tables**
   - sqlc fails to parse `CREATE VIRTUAL TABLE ... USING fts5(...)`
   - sqlc cannot generate queries against FTS5 tables
   - **Solution**: Manual queries in `fts.go`

2. **Recursive CTEs (Common Table Expressions)**
   - sqlc cannot parse `WITH RECURSIVE ...` syntax
   - **Solution**: Manual queries in `cte.go`

3. **Complex SQLite-specific syntax**
   - Some advanced SQLite features are not supported by sqlc's parser

## When to Use This Package

**Use sqlcext when:**
- You need FTS5 full-text search
- You need recursive CTEs for hierarchical data
- sqlc fails with "unsupported" or "parse error"

**Use sqlc (normal SQL files) when:**
- Standard CRUD operations
- Simple SELECT/INSERT/UPDATE/DELETE
- JOINs, WHERE clauses, GROUP BY, ORDER BY
- Any query that sqlc can successfully parse

## Files in This Package

### `fts.go`
- **Purpose**: Full-text search queries for FTS5 virtual tables
- **Tables**: `notes_fts`, `assistant_notes_fts`
- **Queries**:
  - `SearchNotes(query string, limit, offset int)` - Search notes by content
  - Returns `[]FTSResult` with id, title, body, rank

### `cte.go`
- **Purpose**: Recursive CTE queries for hierarchical collections
- **Tables**: `collections` (tree structure)
- **Queries**:
  - `GetCollectionTree(maxDepth int)` - Full tree from all roots
  - `GetCollectionSubtree(rootID, maxDepth int)` - Subtree from specific node
  - Returns `[]CollectionTreeRow` with id, name, parent_id, path, depth

### `types.go`
- **Purpose**: Common types used across manual queries
- **Types**:
  - `DB` interface - for `*sql.DB`, `*sql.Tx`, or sqlc.DBTX
  - `FTSResult` - FTS search result row
  - `CollectionTreeRow` - CTE tree result row

### `bulk.go`
- **Purpose**: Bulk insert operations (performance optimization)
- **Operations**:
  - `BulkInsertNotes()` - Batch insert notes
  - `BulkInsertMeta()` - Batch insert metadata
  - `BulkInsertTags()` - Batch insert tags

### `sanitize.go`
- **Purpose**: FTS5 query sanitization (security)
- **Functions**:
  - `SanitizeFTSQuery(query string)` - Escape FTS5 special characters
  - Prevents FTS5 syntax errors from user input

## Security

**ALL queries in this package use parameterized statements:**
- ✅ Correct: `db.QueryContext(ctx, query, userInput)` with `?` placeholders
- ❌ NEVER: `fmt.Sprintf("SELECT * FROM x WHERE y = '%s'", userInput)`

## Testing

Each file has a corresponding `_test.go` file:
- `fts_test.go` - FTS search tests with in-memory SQLite
- `cte_test.go` - CTE tree/subtree tests
- `bulk_test.go` - Bulk operation tests
- `sanitize_test.go` - Query sanitization tests

Run tests: `go test ./pkg/sqlcext/...`

## Integration Pattern

### Example: Using FTS in a Service

```go
import "github.com/nkapatos/mindweaver/pkg/sqlcext"

type SearchService struct {
    db          *sql.DB
    ftsQuerier  *sqlcext.FTSQuerier
}

func NewSearchService(db *sql.DB) *SearchService {
    return &SearchService{
        db:         db,
        ftsQuerier: sqlcext.NewFTSQuerier(db, sqlcext.FTSConfig{
            TableName: "notes_fts",
            Columns:   []string{"title", "body"},
        }),
    }
}

func (s *SearchService) Search(ctx context.Context, query string) ([]sqlcext.FTSResult, error) {
    return s.ftsQuerier.SearchNotes(ctx, query, 50, 0)
}
```

### Example: Using CTE in a Service

```go
import "github.com/nkapatos/mindweaver/pkg/sqlcext"

type CollectionsService struct {
    store      store.Querier  // sqlc-generated queries
    cteQuerier *sqlcext.CTEQuerier  // manual CTE queries
}

func NewCollectionsService(db sqlcext.DB, store store.Querier) *CollectionsService {
    return &CollectionsService{
        store:      store,
        cteQuerier: sqlcext.NewCTEQuerier(db),
    }
}

func (s *CollectionsService) GetTree(ctx context.Context) ([]sqlcext.CollectionTreeRow, error) {
    return s.cteQuerier.GetCollectionTree(ctx, 10)  // max depth 10
}
```

## Migration Guide

### Moving a Query from sqlc to sqlcext

If sqlc fails to parse your SQL:

1. **Identify the unsupported feature** (FTS5? Recursive CTE?)
2. **Move query to appropriate file** (`fts.go` or `cte.go`)
3. **Create a method** on the querier struct
4. **Use parameterized queries** (`?` placeholders)
5. **Write tests** in `*_test.go`
6. **Update service** to use manual querier instead of store

### Example: Moving an FTS Query

**Before** (won't work in sqlc):
```sql
-- sql/notes_search.sql
-- name: SearchNotes :many
SELECT * FROM notes_fts WHERE notes_fts MATCH ?1;
```

**After** (in `fts.go`):
```go
func (q *FTSQuerier) SearchNotes(ctx context.Context, query string, limit, offset int) ([]FTSResult, error) {
    sanitized := SanitizeFTSQuery(query)
    sql := fmt.Sprintf(`
        SELECT rowid, title, body, rank
        FROM %s
        WHERE %s MATCH ?
        ORDER BY rank
        LIMIT ? OFFSET ?
    `, q.config.TableName, q.config.TableName)
    
    rows, err := q.db.QueryContext(ctx, sql, sanitized, limit, offset)
    // ... handle results
}
```

## Maintenance

### Adding New Manual Queries

1. **Check if sqlc can handle it first** - try generating with `sqlc generate`
2. **If sqlc fails**, add to this package
3. **Follow existing patterns** (parameterized queries, tests)
4. **Document in this README**

### Updating sqlc Version

When updating sqlc:
- Check if FTS5 or CTE support has been added
- If yes, migrate queries back to SQL files
- Remove from sqlcext if no longer needed

## References

- [sqlc Documentation](https://docs.sqlc.dev/)
- [SQLite FTS5 Documentation](https://www.sqlite.org/fts5.html)
- [SQLite Recursive CTEs](https://www.sqlite.org/lang_with.html)

## Change Log

- **2025-12-08**: Added CTE queries for collection tree/subtree
- **2025-12-04**: Added FTS5 queries for full-text search
- **2025-11-29**: Initial package creation
