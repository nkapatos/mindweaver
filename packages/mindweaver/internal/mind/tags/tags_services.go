package tags

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	sharedErrors "github.com/nkapatos/mindweaver/packages/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/middleware"
)

// TagsService provides business logic for tags (CRUD + search only).
type TagsService struct {
	store  store.Querier
	logger *slog.Logger
}

// NewTagsService creates a new TagsService.
func NewTagsService(store store.Querier, logger *slog.Logger, serviceName string) *TagsService {
	return &TagsService{
		store:  store,
		logger: logger.With("service", serviceName),
	}
}

// ListTags returns all tags.
func (s *TagsService) ListTags(ctx context.Context) ([]store.Tag, error) {
	tags, err := s.store.ListTags(ctx)
	if err != nil {
		s.logger.Error("failed to list tags", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return tags, err
}

// ListTagsPaginated returns tags with pagination.
func (s *TagsService) ListTagsPaginated(ctx context.Context, limit, offset int32) ([]store.Tag, error) {
	tags, err := s.store.ListTagsPaginated(ctx, store.ListTagsPaginatedParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list tags paginated", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return tags, err
}

// CountTags returns the total number of tags.
func (s *TagsService) CountTags(ctx context.Context) (int64, error) {
	count, err := s.store.CountTags(ctx)
	if err != nil {
		s.logger.Error("failed to count tags", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// GetTagByID returns a tag by ID.
func (s *TagsService) GetTagByID(ctx context.Context, id int64) (store.Tag, error) {
	tag, err := s.store.GetTagByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Tag{}, ErrTagNotFound
		}
		s.logger.Error("failed to get tag by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Tag{}, err
	}
	return tag, nil
}

// CreateTag creates a new tag.
func (s *TagsService) CreateTag(ctx context.Context, name string) (int64, error) {
	id, err := s.store.CreateTag(ctx, name)
	if err != nil {
		if sharedErrors.IsUniqueConstraintError(err) {
			return 0, ErrTagAlreadyExists
		}
		s.logger.Error("failed to create tag", "name", name, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	s.logger.Info("tag created", "id", id, "request_id", middleware.GetRequestID(ctx))
	return id, nil
}

// UpdateTag updates an existing tag.
func (s *TagsService) UpdateTag(ctx context.Context, id int64, name string) error {
	params := store.UpdateTagByIDParams{ID: id, Name: name}
	err := s.store.UpdateTagByID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTagNotFound
		}
		if sharedErrors.IsUniqueConstraintError(err) {
			return ErrTagAlreadyExists
		}
		s.logger.Error("failed to update tag", "id", id, "name", name, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("tag updated", "id", id, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// DeleteTag deletes a tag by ID.
func (s *TagsService) DeleteTag(ctx context.Context, id int64) error {
	err := s.store.DeleteTagByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTagNotFound
		}
		s.logger.Error("failed to delete tag", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("tag deleted", "id", id, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// SearchTagsByName returns tags matching a name pattern.
func (s *TagsService) SearchTagsByName(ctx context.Context, namePattern string) ([]store.Tag, error) {
	tags, err := s.store.SearchTagsByName(ctx, namePattern)
	if err != nil {
		s.logger.Error("failed to search tags by name", "pattern", namePattern, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return tags, err
}

// ListTagsForNote returns all tags for a given note.
func (s *TagsService) ListTagsForNote(ctx context.Context, noteID int64) ([]store.Tag, error) {
	tags, err := s.store.ListTagsForNote(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to list tags for note", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return tags, err
}

// ListTagsForNotePaginated returns tags for a note with pagination.
func (s *TagsService) ListTagsForNotePaginated(ctx context.Context, noteID int64, limit, offset int32) ([]store.Tag, error) {
	tags, err := s.store.ListTagsForNotePaginated(ctx, store.ListTagsForNotePaginatedParams{
		NoteID: noteID,
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list tags for note paginated", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return tags, err
}

// CountTagsForNote returns the total number of tags for a note.
func (s *TagsService) CountTagsForNote(ctx context.Context, noteID int64) (int64, error) {
	count, err := s.store.CountTagsForNote(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to count tags for note", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// ListNotesForTag returns all notes for a given tag.
func (s *TagsService) ListNotesForTag(ctx context.Context, tagID int64) ([]store.Note, error) {
	notes, err := s.store.ListNotesForTag(ctx, tagID)
	if err != nil {
		s.logger.Error("failed to list notes for tag", "tag_id", tagID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}
