package notetypes

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	sharedErrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/middleware"
	"github.com/nkapatos/mindweaver/shared/utils"
)

// NoteTypesService provides business logic for note_types (CRUD only).
type NoteTypesService struct {
	store  store.Querier
	logger *slog.Logger
}

// NewNoteTypesService creates a new NoteTypesService.
func NewNoteTypesService(store store.Querier, logger *slog.Logger, serviceName string) *NoteTypesService {
	return &NoteTypesService{
		store:  store,
		logger: logger.With("service", serviceName),
	}
}

// ListNoteTypes returns all note_types.
func (s *NoteTypesService) ListNoteTypes(ctx context.Context) ([]store.NoteType, error) {
	items, err := s.store.ListNoteTypes(ctx)
	if err != nil {
		s.logger.Error("failed to list note_types", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return items, err
}

// ListNoteTypesPaginated returns note_types with pagination.
func (s *NoteTypesService) ListNoteTypesPaginated(ctx context.Context, limit, offset int32) ([]store.NoteType, error) {
	items, err := s.store.ListNoteTypesPaginated(ctx, store.ListNoteTypesPaginatedParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list note_types paginated", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return items, err
}

// CountNoteTypes returns the total number of note_types.
func (s *NoteTypesService) CountNoteTypes(ctx context.Context) (int64, error) {
	count, err := s.store.CountNoteTypes(ctx)
	if err != nil {
		s.logger.Error("failed to count note_types", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// GetNoteTypeByID returns a note_type by ID.
func (s *NoteTypesService) GetNoteTypeByID(ctx context.Context, id int64) (store.NoteType, error) {
	item, err := s.store.GetNoteTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.NoteType{}, ErrNoteTypeNotFound
		}
		s.logger.Error("failed to get note_type by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.NoteType{}, err
	}
	return item, nil
}

// GetNoteTypeByType returns a note_type by its type string.
func (s *NoteTypesService) GetNoteTypeByType(ctx context.Context, typeStr string) (store.NoteType, error) {
	item, err := s.store.GetNoteTypeByType(ctx, typeStr)
	if err != nil {
		s.logger.Error("failed to get note_type by type", "type", typeStr, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return item, err
}

// GetNoteTypesByTypes returns note_types by a list of type strings.
func (s *NoteTypesService) GetNoteTypesByTypes(ctx context.Context, types []string) ([]store.NoteType, error) {
	typesStr := strings.Join(types, ",")
	items, err := s.store.GetNoteTypesByTypes(ctx, typesStr)
	if err != nil {
		s.logger.Error("failed to get note_types by types", "types", types, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return items, err
}

// CreateNoteType creates a new note_type.
func (s *NoteTypesService) CreateNoteType(ctx context.Context, params store.CreateNoteTypeParams) (int64, error) {
	id, err := s.store.CreateNoteType(ctx, params)
	if err != nil {
		if sharedErrors.IsUniqueConstraintError(err) {
			return 0, ErrNoteTypeAlreadyExists
		}
		s.logger.Error("failed to create note_type", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	s.logger.Info("note_type created", "id", id, "request_id", middleware.GetRequestID(ctx))
	return id, nil
}

// UpdateNoteType updates an existing note_type.
// System note types cannot be updated.
func (s *NoteTypesService) UpdateNoteType(ctx context.Context, params store.UpdateNoteTypeByIDParams) error {
	// First check if it's a system type
	noteType, err := s.store.GetNoteTypeByID(ctx, params.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoteTypeNotFound
		}
		s.logger.Error("failed to get note_type for update check", "id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	if noteType.IsSystem {
		return ErrNoteTypeIsSystem
	}

	// Preserve IsSystem from existing record
	params.IsSystem = noteType.IsSystem

	err = s.store.UpdateNoteTypeByID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoteTypeNotFound
		}
		if sharedErrors.IsUniqueConstraintError(err) {
			return ErrNoteTypeAlreadyExists
		}
		s.logger.Error("failed to update note_type", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("note_type updated", "id", params.ID, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// DeleteNoteType deletes a note_type by ID.
func (s *NoteTypesService) DeleteNoteType(ctx context.Context, id int64) error {
	// First check if it's a system type
	noteType, err := s.store.GetNoteTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoteTypeNotFound
		}
		s.logger.Error("failed to get note_type for deletion check", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	if noteType.IsSystem {
		return ErrNoteTypeIsSystem
	}

	err = s.store.DeleteNoteTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoteTypeNotFound
		}
		s.logger.Error("failed to delete note_type", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("note_type deleted", "id", id, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// ListNotesByNoteTypeID returns all notes for a given note_type.
func (s *NoteTypesService) ListNotesByNoteTypeID(ctx context.Context, noteTypeID int64) ([]store.Note, error) {
	nullTypeID := utils.ToNullInt64(&noteTypeID)
	notes, err := s.store.ListNotesByNoteTypeID(ctx, nullTypeID)
	if err != nil {
		s.logger.Error("failed to list notes by note_type_id", "note_type_id", noteTypeID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}
