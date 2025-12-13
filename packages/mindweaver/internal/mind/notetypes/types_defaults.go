// Package notes provides default data initialization for note types.
package notetypes

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/store"
)

// DefaultNoteType defines a system note type that must exist for the app to function.
type DefaultNoteType struct {
	Type        string
	Name        string
	Description string
}

// defaultNoteTypes are the required system note types.
var defaultNoteTypes = []DefaultNoteType{
	{
		Type:        "default",
		Name:        "Default",
		Description: "Standard notes without a specific type",
	},
	{
		Type:        "quicknote",
		Name:        "Quick Note",
		Description: "High-volume frictionless capture notes",
	},
}

// EnsureDefaultNoteTypes ensures all required system note types exist.
// This is idempotent - safe to call multiple times.
func EnsureDefaultNoteTypes(ctx context.Context, q *store.Queries, logger *slog.Logger) error {
	for _, dt := range defaultNoteTypes {
		// Check if note type already exists by type
		_, err := q.GetNoteTypeByType(ctx, dt.Type)
		if err == nil {
			// Already exists, skip
			continue
		}
		if err != sql.ErrNoRows {
			// Unexpected error
			return err
		}

		// Create the note type
		_, err = q.CreateNoteType(ctx, store.CreateNoteTypeParams{
			Type:        dt.Type,
			Name:        dt.Name,
			Description: sql.NullString{String: dt.Description, Valid: true},
			Icon:        sql.NullString{},
			Color:       sql.NullString{},
			IsSystem:    true,
		})
		if err != nil {
			return err
		}
		logger.Info("Created default note type", "type", dt.Type)
	}
	return nil
}
