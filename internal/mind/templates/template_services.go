package templates

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	sharedErrors "github.com/nkapatos/mindweaver/packages/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/middleware"
)

// TemplatesService provides business logic for templates (CRUD + search only).
type TemplatesService struct {
	store  store.Querier
	logger *slog.Logger
}

// NewTemplatesService creates a new TemplatesService.
func NewTemplatesService(store store.Querier, logger *slog.Logger, serviceName string) *TemplatesService {
	return &TemplatesService{
		store:  store,
		logger: logger.With("service", serviceName),
	}
}

// ListTemplates returns all templates.
func (s *TemplatesService) ListTemplates(ctx context.Context) ([]store.Template, error) {
	templates, err := s.store.ListTemplates(ctx)
	if err != nil {
		s.logger.Error("failed to list templates", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return templates, err
}

// ListTemplatesPaginated returns templates with pagination.
func (s *TemplatesService) ListTemplatesPaginated(ctx context.Context, limit, offset int32) ([]store.Template, error) {
	templates, err := s.store.ListTemplatesPaginated(ctx, store.ListTemplatesPaginatedParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list templates paginated", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return templates, err
}

// CountTemplates returns the total number of templates.
func (s *TemplatesService) CountTemplates(ctx context.Context) (int64, error) {
	count, err := s.store.CountTemplates(ctx)
	if err != nil {
		s.logger.Error("failed to count templates", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// GetTemplateByID returns a template by ID.
func (s *TemplatesService) GetTemplateByID(ctx context.Context, id int64) (store.Template, error) {
	template, err := s.store.GetTemplateByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Template{}, ErrTemplateNotFound
		}
		s.logger.Error("failed to get template by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Template{}, err
	}
	return template, nil
}

// CreateTemplate creates a new template.
func (s *TemplatesService) CreateTemplate(ctx context.Context, params store.CreateTemplateParams) (int64, error) {
	id, err := s.store.CreateTemplate(ctx, params)
	if err != nil {
		if sharedErrors.IsUniqueConstraintError(err) {
			return 0, ErrTemplateAlreadyExists
		}
		s.logger.Error("failed to create template", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	s.logger.Info("template created", "id", id, "request_id", middleware.GetRequestID(ctx))
	return id, nil
}

// UpdateTemplate updates an existing template.
func (s *TemplatesService) UpdateTemplate(ctx context.Context, params store.UpdateTemplateByIDParams) error {
	err := s.store.UpdateTemplateByID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTemplateNotFound
		}
		if sharedErrors.IsUniqueConstraintError(err) {
			return ErrTemplateAlreadyExists
		}
		s.logger.Error("failed to update template", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("template updated", "id", params.ID, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// DeleteTemplate deletes a template by ID.
func (s *TemplatesService) DeleteTemplate(ctx context.Context, id int64) error {
	err := s.store.DeleteTemplateByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTemplateNotFound
		}
		s.logger.Error("failed to delete template", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("template deleted", "id", id, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// Note: Template search methods removed - use general search with filters instead
