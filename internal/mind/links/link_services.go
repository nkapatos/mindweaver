package links

import (
	"context"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/mind/store"
	"github.com/nkapatos/mindweaver/pkg/middleware"
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

// ListLinks returns all links (matches ListLinks in proto).
func (s *LinksService) ListLinks(ctx context.Context) ([]store.NotesLink, error) {
	items, err := s.store.ListNotesLinks(ctx)
	if err != nil {
		s.logger.Error("failed to list notes_links", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return items, err
}

// GetLinkByID returns a link by ID.
func (s *LinksService) GetLinkByID(ctx context.Context, id int64) (store.NotesLink, error) {
	item, err := s.store.GetNotesLinkByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get notes_link by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return item, err
}

// ============================================================================
// TODO: Move the following methods to notes service
// These operations are triggered by note creation/update, not link management
// ============================================================================

// // DeleteLinksBySourceID deletes all outgoing links from a note.
// // Used when updating a note's body to remove old links before adding new ones.
// func (s *LinksService) DeleteLinksBySourceID(ctx context.Context, srcID int64) error {
// 	err := s.store.DeleteNotesLinksBySrcID(ctx, srcID)
// 	if err != nil {
// 		s.logger.Error("failed to delete links by src_id", "src_id", srcID, "err", err, "request_id", middleware.GetRequestID(ctx))
// 		return err
// 	}
// 	s.logger.Info("links deleted by src_id", "src_id", srcID, "request_id", middleware.GetRequestID(ctx))
// 	return nil
// }

// // ListLinksBySrcID returns all outgoing links from a note (where note is the source).
// func (s *LinksService) ListLinksBySrcID(ctx context.Context, srcID int64) ([]store.NotesLink, error) {
// 	links, err := s.store.ListNotesLinksBySrcID(ctx, srcID)
// 	if err != nil {
// 		s.logger.Error("failed to list links by src_id", "src_id", srcID, "err", err, "request_id", middleware.GetRequestID(ctx))
// 	}
// 	return links, err
// }

// // ListLinksByDestID returns all incoming links to a note (backlinks, where note is the destination).
// func (s *LinksService) ListLinksByDestID(ctx context.Context, destID int64) ([]store.NotesLink, error) {
// 	links, err := s.store.ListNotesLinksByDestID(ctx, utils.ToNullInt64(&destID))
// 	if err != nil {
// 		s.logger.Error("failed to list links by dest_id", "dest_id", destID, "err", err, "request_id", middleware.GetRequestID(ctx))
// 	}
// 	return links, err
// }

// // CreateLinksForNote creates notes_links entries from parsed WikiLinks.
// // It attempts to resolve links to existing notes by title.
// // If a target note exists, creates a resolved link (dest_id set, dest_title NULL).
// // If target doesn't exist, creates an unresolved link (dest_id NULL, dest_title set).
// func (s *LinksService) CreateLinksForNote(ctx context.Context, noteID int64, wikiLinks []markdown.WikiLink) error {
// 	for _, link := range wikiLinks {
// 		// Try to find target note by title (search globally across all collections)
// 		targetNote, err := s.store.GetNoteByTitleGlobal(ctx, link.Target)

// 		if err == nil {
// 			// Target exists - create resolved link
// 			params := store.CreateNotesLinkParams{
// 				SrcID:       noteID,
// 				DestID:      utils.ToNullInt64(&targetNote.ID),
// 				DisplayText: utils.ToNullString(&link.DisplayText),
// 				IsEmbed:     sql.NullBool{Bool: link.Embed, Valid: true},
// 			}
// 			_, err = s.store.CreateNotesLink(ctx, params)
// 			if err != nil {
// 				s.logger.Error("failed to create resolved link", "src_id", noteID, "dest_id", targetNote.ID, "target", link.Target, "err", err, "request_id", middleware.GetRequestID(ctx))
// 				return err
// 			}
// 			s.logger.Info("resolved link created", "src_id", noteID, "dest_id", targetNote.ID, "target", link.Target, "request_id", middleware.GetRequestID(ctx))
// 		} else if err == sql.ErrNoRows {
// 			// Target doesn't exist - create unresolved link
// 			params := store.CreateUnresolvedNotesLinkParams{
// 				SrcID:       noteID,
// 				DestTitle:   utils.ToNullString(&link.Target),
// 				DisplayText: utils.ToNullString(&link.DisplayText),
// 				IsEmbed:     sql.NullBool{Bool: link.Embed, Valid: true},
// 			}
// 			_, err = s.store.CreateUnresolvedNotesLink(ctx, params)
// 			if err != nil {
// 				s.logger.Error("failed to create unresolved link", "src_id", noteID, "target", link.Target, "err", err, "request_id", middleware.GetRequestID(ctx))
// 				return err
// 			}
// 			s.logger.Info("unresolved link created", "src_id", noteID, "target", link.Target, "request_id", middleware.GetRequestID(ctx))
// 		} else {
// 			// Database error
// 			s.logger.Error("failed to check if target note exists", "target", link.Target, "err", err, "request_id", middleware.GetRequestID(ctx))
// 			return err
// 		}
// 	}
// 	return nil
// }

// // ResolveBacklinksForNote finds and resolves any unresolved links pointing to this note.
// // This should be called after creating a new note to resolve any pending WikiLinks.
// func (s *LinksService) ResolveBacklinksForNote(ctx context.Context, noteID int64, noteTitle string) error {
// 	// Find unresolved links with matching dest_title
// 	unresolvedLinks, err := s.store.FindUnresolvedLinksByDestTitle(ctx, utils.ToNullString(&noteTitle))
// 	if err != nil {
// 		s.logger.Error("failed to find unresolved links", "title", noteTitle, "err", err, "request_id", middleware.GetRequestID(ctx))
// 		return err
// 	}

// 	// Resolve each link
// 	for _, link := range unresolvedLinks {
// 		params := store.ResolveLinkParams{
// 			ID:     link.ID,
// 			DestID: utils.ToNullInt64(&noteID),
// 		}
// 		err = s.store.ResolveLink(ctx, params)
// 		if err != nil {
// 			s.logger.Error("failed to resolve link", "link_id", link.ID, "dest_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
// 			return err
// 		}
// 		s.logger.Info("link resolved", "link_id", link.ID, "src_id", link.SrcID, "dest_id", noteID, "request_id", middleware.GetRequestID(ctx))
// 	}

// 	if len(unresolvedLinks) > 0 {
// 		s.logger.Info("resolved backlinks for new note", "note_id", noteID, "title", noteTitle, "count", len(unresolvedLinks), "request_id", middleware.GetRequestID(ctx))
// 	}

// 	return nil
// }
