// Package collections provides default data initialization for collections.
package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
)

// DefaultCollection defines a system collection that must exist for the app to function.
type DefaultCollection struct {
	Name        string
	Path        string
	Description string
	Position    int64
}

// defaultCollections are the required system collections.
var defaultCollections = []DefaultCollection{
	{
		Name:        "default",
		Path:        "default",
		Description: "Default collection for uncategorized notes",
		Position:    0,
	},
	{
		Name:        "inbox",
		Path:        "inbox",
		Description: "Staging area for imports and refinement",
		Position:    1,
	},
}

// EnsureDefaultCollections ensures all required system collections exist.
// This is idempotent - safe to call multiple times.
func EnsureDefaultCollections(ctx context.Context, q *store.Queries, logger *slog.Logger) error {
	for _, dc := range defaultCollections {
		// Check if collection already exists by path
		_, err := q.GetCollectionByPath(ctx, dc.Path)
		if err == nil {
			// Already exists, skip
			continue
		}
		if err != sql.ErrNoRows {
			// Unexpected error
			return err
		}

		// Create the collection
		_, err = q.CreateCollection(ctx, store.CreateCollectionParams{
			Name:        dc.Name,
			ParentID:    nil, // Top-level
			Path:        dc.Path,
			Description: sql.NullString{String: dc.Description, Valid: true},
			Position:    sql.NullInt64{Int64: dc.Position, Valid: true},
			IsSystem:    true,
		})
		if err != nil {
			return err
		}
		logger.Info("Created default collection", "path", dc.Path)
	}
	return nil
}
