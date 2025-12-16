package links

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/middleware"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/utils"
)

// LinksService provides business logic for links resource operations.
// This service handles ONLY the links resource as defined in the links.proto.
// Note-specific link operations (creating/updating links when notes change)
// should be handled by the notes service.
type LinksService struct {
	store  store.Querier
	logger *slog.Logger
}

// NewLinksService creates a new LinksService.
func NewLinksService(store store.Querier, logger *slog.Logger, serviceName string) *LinksService {
	return &LinksService{
		store:  store,
		logger: logger.With("service", serviceName),
	}
}

// ============================================================================
// Basic CRUD Operations
// ============================================================================

// CreateLink creates a resolved link between two notes.
func (s *LinksService) CreateLink(ctx context.Context, params store.CreateLinkParams) (int64, error) {
	id, err := s.store.CreateLink(ctx, params)
	if err != nil {
		s.logger.Error("failed to create link", "src_id", params.SrcID, "dest_id", params.DestID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	s.logger.Info("link created", "id", id, "src_id", params.SrcID, "dest_id", params.DestID, "request_id", middleware.GetRequestID(ctx))
	return id, nil
}

// CreateUnresolvedLink creates an unresolved link (target note doesn't exist yet).
func (s *LinksService) CreateUnresolvedLink(ctx context.Context, params store.CreateUnresolvedLinkParams) (int64, error) {
	id, err := s.store.CreateUnresolvedLink(ctx, params)
	if err != nil {
		s.logger.Error("failed to create unresolved link", "src_id", params.SrcID, "dest_title", params.DestTitle, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	s.logger.Info("unresolved link created", "id", id, "src_id", params.SrcID, "dest_title", params.DestTitle, "request_id", middleware.GetRequestID(ctx))
	return id, nil
}

// GetLinkByID returns a link by ID.
func (s *LinksService) GetLinkByID(ctx context.Context, id int64) (store.Link, error) {
	link, err := s.store.GetLinkByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("link not found", "id", id, "request_id", middleware.GetRequestID(ctx))
		} else {
			s.logger.Error("failed to get link by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		}
		return store.Link{}, err
	}
	return link, nil
}

// ListLinks returns all links.
func (s *LinksService) ListLinks(ctx context.Context) ([]store.Link, error) {
	links, err := s.store.ListLinks(ctx)
	if err != nil {
		s.logger.Error("failed to list links", "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// DeleteLinksBySrcID deletes all outgoing links from a note.
func (s *LinksService) DeleteLinksBySrcID(ctx context.Context, srcID int64) error {
	err := s.store.DeleteLinksBySrcID(ctx, srcID)
	if err != nil {
		s.logger.Error("failed to delete links by src_id", "src_id", srcID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("links deleted by src_id", "src_id", srcID, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// ============================================================================
// Query Operations
// ============================================================================

// ListLinksBySrcID returns all outgoing links from a note (forward links).
func (s *LinksService) ListLinksBySrcID(ctx context.Context, srcID int64) ([]store.Link, error) {
	links, err := s.store.ListLinksBySrcID(ctx, srcID)
	if err != nil {
		s.logger.Error("failed to list links by src_id", "src_id", srcID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// ListLinksByDestID returns all incoming links to a note (backlinks).
func (s *LinksService) ListLinksByDestID(ctx context.Context, destID sql.NullInt64) ([]store.Link, error) {
	links, err := s.store.ListLinksByDestID(ctx, destID)
	if err != nil {
		s.logger.Error("failed to list links by dest_id", "dest_id", destID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// SearchLinksByDisplayText searches for links by display text pattern.
func (s *LinksService) SearchLinksByDisplayText(ctx context.Context, pattern string) ([]store.Link, error) {
	links, err := s.store.SearchLinksByDisplayText(ctx, utils.NullString(pattern))
	if err != nil {
		s.logger.Error("failed to search links by display text", "pattern", pattern, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// ============================================================================
// WikiLink Resolution Operations
// ============================================================================

// ListUnresolvedLinks returns pending and broken links for resolution.
func (s *LinksService) ListUnresolvedLinks(ctx context.Context, limit int64) ([]store.Link, error) {
	links, err := s.store.ListUnresolvedLinks(ctx, limit)
	if err != nil {
		s.logger.Error("failed to list unresolved links", "limit", limit, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// FindUnresolvedLinksByDestTitle finds unresolved links pointing to a specific note title.
func (s *LinksService) FindUnresolvedLinksByDestTitle(ctx context.Context, destTitle sql.NullString) ([]store.Link, error) {
	links, err := s.store.FindUnresolvedLinksByDestTitle(ctx, destTitle)
	if err != nil {
		s.logger.Error("failed to find unresolved links by dest_title", "dest_title", destTitle, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// CountUnresolvedLinks returns the count of pending links.
func (s *LinksService) CountUnresolvedLinks(ctx context.Context) (int64, error) {
	count, err := s.store.CountUnresolvedLinks(ctx)
	if err != nil {
		s.logger.Error("failed to count unresolved links", "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	return count, nil
}

// ResolveLink resolves a pending link by setting the destination note ID.
func (s *LinksService) ResolveLink(ctx context.Context, params store.ResolveLinkParams) error {
	err := s.store.ResolveLink(ctx, params)
	if err != nil {
		s.logger.Error("failed to resolve link", "link_id", params.ID, "dest_id", params.DestID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("link resolved", "link_id", params.ID, "dest_id", params.DestID, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// MarkLinkBroken marks a link as broken (resolved = -1).
func (s *LinksService) MarkLinkBroken(ctx context.Context, id int64) error {
	err := s.store.MarkLinkBroken(ctx, id)
	if err != nil {
		s.logger.Error("failed to mark link broken", "link_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("link marked as broken", "link_id", id, "request_id", middleware.GetRequestID(ctx))
	return nil
}

// ============================================================================
// Broken/Orphaned Links Operations
// ============================================================================

// ListBrokenLinks returns all broken links (resolved = -1).
func (s *LinksService) ListBrokenLinks(ctx context.Context) ([]store.Link, error) {
	links, err := s.store.ListBrokenLinks(ctx)
	if err != nil {
		s.logger.Error("failed to list broken links", "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}

// CountBrokenLinks returns the count of broken links.
func (s *LinksService) CountBrokenLinks(ctx context.Context) (int64, error) {
	count, err := s.store.CountBrokenLinks(ctx)
	if err != nil {
		s.logger.Error("failed to count broken links", "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	return count, nil
}

// ListOrphanedLinks returns links where destination note no longer exists.
func (s *LinksService) ListOrphanedLinks(ctx context.Context) ([]store.Link, error) {
	links, err := s.store.ListOrphanedLinks(ctx)
	if err != nil {
		s.logger.Error("failed to list orphaned links", "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	return links, nil
}
