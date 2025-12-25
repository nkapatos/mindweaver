package meta

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/shared/middleware"
)

// NoteMetaService provides business logic for note metadata operations.
type NoteMetaService struct {
	store  store.Querier
	db     *sql.DB
	logger *slog.Logger
}

// NewNoteMetaService creates a new NoteMetaService.
func NewNoteMetaService(store store.Querier, db *sql.DB, logger *slog.Logger, serviceName string) *NoteMetaService {
	return &NoteMetaService{
		store:  store,
		db:     db,
		logger: logger.With("service", serviceName),
	}
}

// GetNoteMetaByNoteID retrieves all metadata for a given note.
func (s *NoteMetaService) GetNoteMetaByNoteID(ctx context.Context, noteID int64) ([]store.NoteMetum, error) {
	items, err := s.store.GetNoteMetaByNoteID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get note metadata", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return items, nil
}

// CreateNoteMeta creates metadata for a note within a transaction.
func (s *NoteMetaService) CreateNoteMeta(ctx context.Context, params store.CreateNoteMetaParams) (int64, error) {
	id, err := s.store.CreateNoteMeta(ctx, params)
	if err != nil {
		s.logger.Error("failed to create note metadata", "note_id", params.NoteID, "key", params.Key, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	return id, nil
}

// DeleteNoteMetaByNoteID deletes all metadata for a note within a transaction.
func (s *NoteMetaService) DeleteNoteMetaByNoteID(ctx context.Context, noteID int64) error {
	err := s.store.DeleteNoteMetaByNoteID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to delete note metadata", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return err
}
