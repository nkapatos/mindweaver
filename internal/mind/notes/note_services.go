package notes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/events"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/internal/mind/links"
	"github.com/nkapatos/mindweaver/internal/mind/meta"
	"github.com/nkapatos/mindweaver/internal/mind/scheduler"
	"github.com/nkapatos/mindweaver/internal/mind/tags"
	sharederrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/markdown"
	"github.com/nkapatos/mindweaver/shared/middleware"
	"github.com/nkapatos/mindweaver/shared/utils"
)

// NotesService provides business logic for notes CRUD operations.
// Handles note creation, updates, and deletion with automatic extraction
// of derived data (wiki-links, tags, metadata) from markdown content.
type NotesService struct {
	store     store.Querier
	db        *sql.DB
	logger    *slog.Logger
	scheduler *scheduler.ChangeAccumulator // Optional: notifies Brain of note changes
	eventHub  events.Hub                   // Optional: publishes events for SSE clients
	parser    *markdown.Parser
}

var untitledCounter int64 = 0

// NewNotesService creates a new NotesService.
func NewNotesService(db *sql.DB, store store.Querier, logger *slog.Logger, serviceName string) *NotesService {
	return &NotesService{
		store:     store,
		db:        db,
		logger:    logger.With("service", serviceName),
		scheduler: nil,
		parser:    markdown.NewParser(),
	}
}

// SetScheduler sets the change scheduler for Brain synchronization.
func (s *NotesService) SetScheduler(scheduler *scheduler.ChangeAccumulator) {
	s.scheduler = scheduler
	s.logger.Info("scheduler enabled for note service")
}

// SetEventHub sets the event hub for SSE notifications.
func (s *NotesService) SetEventHub(hub events.Hub) {
	s.eventHub = hub
	s.logger.Info("event hub enabled for note service")
}

// GetMarkdownParser returns the markdown parser instance.
func (s *NotesService) GetMarkdownParser() *markdown.Parser {
	return s.parser
}

// ListNotesPaginated returns notes with pagination.
func (s *NotesService) ListNotesPaginated(ctx context.Context, limit, offset int32) ([]store.Note, error) {
	notes, err := s.store.ListNotesPaginated(ctx, store.ListNotesPaginatedParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list notes paginated", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}

// CountNotes returns the total number of notes.
func (s *NotesService) CountNotes(ctx context.Context) (int64, error) {
	count, err := s.store.CountNotes(ctx)
	if err != nil {
		s.logger.Error("failed to count notes", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// GetNoteByID returns a note by ID.
func (s *NotesService) GetNoteByID(ctx context.Context, id int64) (store.Note, error) {
	note, err := s.store.GetNoteByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Note{}, ErrNoteNotFound
		}
		s.logger.Error("failed to get note by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Note{}, err
	}
	return note, nil
}

// CreateNote creates a new note with derived data (links, tags) atomically.
// All operations are wrapped in a transaction to ensure consistency.
func (s *NotesService) CreateNote(ctx context.Context, params store.CreateNoteParams) (int64, error) {
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("failed to begin transaction", "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	defer tx.Rollback()

	txStore := store.New(tx)

	id, err := txStore.CreateNote(ctx, params)
	if err != nil {
		if sharederrors.IsUniqueConstraintError(err) {
			return 0, ErrNoteAlreadyExists
		}
		s.logger.Error("failed to create note", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}

	// Extract and store derived data from note body (wiki-links, tags, metadata)
	if params.Body.Valid && params.Body.String != "" {
		parsed, err := s.parser.Parse([]byte(params.Body.String))
		if err != nil {
			s.logger.Error("failed to parse note body", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err
		}

		allTags := s.extractAndMergeTags(parsed)

		if err := s.insertWikiLinksWithStore(ctx, txStore, id, parsed); err != nil {
			s.logger.Error("failed to insert wiki-links", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err
		}

		if err := s.insertTagsWithStore(ctx, txStore, id, allTags); err != nil {
			s.logger.Error("failed to insert tags", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err
		}

		// Note: 'tags'/'tag' frontmatter keys are filtered out here (handled above)
		if err := s.insertMetadataWithStore(ctx, txStore, id, parsed, nil); err != nil {
			s.logger.Error("failed to insert metadata", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}

	s.logger.Info("note created", "id", id, "request_id", middleware.GetRequestID(ctx))

	if s.scheduler != nil {
		s.scheduler.TrackChange("note_created", id)
	}

	if s.eventHub != nil {
		s.eventHub.Publish(ctx, mindv3.EventDomain_EVENT_DOMAIN_NOTE, mindv3.EventType_EVENT_TYPE_CREATED, id)
	}

	return id, nil
}

// NewNoteCreation creates a new note with auto-generated title and optional template content.
func (s *NotesService) NewNoteCreation(ctx context.Context, collectionID, templateID int64) (int64, error) {
	// Generate auto-incremented title
	untitledCounter++
	title := fmt.Sprintf("Untitled %d", untitledCounter)

	// Get template body (template_id 1 is empty by default)
	body := ""
	if templateID != 1 {
		templateNote, err := s.GetNoteByID(ctx, templateID)
		if err != nil {
			s.logger.Error("failed to get template note", "template_id", templateID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err
		}
		if templateNote.Body.Valid {
			body = templateNote.Body.String
		}
	}

	// Build params for CreateNote
	params := store.CreateNoteParams{
		Uuid:         uuid.New(),
		Title:        title,
		Body:         utils.NullStringFrom(body, true),
		CollectionID: collectionID,
	}

	// Delegate to existing CreateNote logic
	noteID, err := s.CreateNote(ctx, params)
	if err != nil {
		s.logger.Error("failed to create new note", "title", title, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}

	return noteID, nil
}

// UpdateNote updates an existing note and re-extracts all derived data.
// Replaces all links, tags, and metadata from the new note body.
func (s *NotesService) UpdateNote(ctx context.Context, params store.UpdateNoteByIDParams) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("failed to begin transaction", "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	defer tx.Rollback()

	txStore := store.New(tx)

	// Clear existing derived data before re-extracting from updated body
	if delErr := txStore.DeleteLinksBySrcID(ctx, params.ID); delErr != nil {
		s.logger.Error("failed to delete existing links", "note_id", params.ID, "err", delErr, "request_id", middleware.GetRequestID(ctx))
		return delErr
	}

	if delErr := txStore.DeleteNoteTagsByNoteID(ctx, params.ID); delErr != nil {
		s.logger.Error("failed to delete existing tags", "note_id", params.ID, "err", delErr, "request_id", middleware.GetRequestID(ctx))
		return delErr
	}

	if delErr := txStore.DeleteNoteMetaByNoteID(ctx, params.ID); delErr != nil {
		s.logger.Error("failed to delete existing metadata", "note_id", params.ID, "err", delErr, "request_id", middleware.GetRequestID(ctx))
		return delErr
	}

	err = txStore.UpdateNoteByID(ctx, params)
	if err != nil {
		if sharederrors.IsUniqueConstraintError(err) {
			return ErrNoteAlreadyExists
		}
		s.logger.Error("failed to update note", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	// Re-extract derived data from updated body
	if params.Body.Valid && params.Body.String != "" {
		parsed, err := s.parser.Parse([]byte(params.Body.String))
		if err != nil {
			s.logger.Error("failed to parse note body", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err
		}

		if err := s.insertWikiLinksWithStore(ctx, txStore, params.ID, parsed); err != nil {
			s.logger.Error("failed to insert wiki-links", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err
		}

		allTags := s.extractAndMergeTags(parsed)
		if err := s.insertTagsWithStore(ctx, txStore, params.ID, allTags); err != nil {
			s.logger.Error("failed to insert tags", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err
		}

		if err := s.insertMetadataWithStore(ctx, txStore, params.ID, parsed, nil); err != nil {
			s.logger.Error("failed to insert metadata", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	s.logger.Info("note updated", "id", params.ID, "request_id", middleware.GetRequestID(ctx))

	if s.scheduler != nil {
		s.scheduler.TrackChange("note_updated", params.ID)
	}

	if s.eventHub != nil {
		s.eventHub.Publish(ctx, mindv3.EventDomain_EVENT_DOMAIN_NOTE, mindv3.EventType_EVENT_TYPE_UPDATED, params.ID)
	}

	return nil
}

// DeleteNote deletes a note by ID.
// Associated links, tags, and metadata are cascade-deleted by database constraints.
func (s *NotesService) DeleteNote(ctx context.Context, id int64) error {
	err := s.store.DeleteNoteByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete note", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("note deleted", "id", id, "request_id", middleware.GetRequestID(ctx))

	if s.scheduler != nil {
		s.scheduler.TrackChange("note_deleted", id)
	}

	if s.eventHub != nil {
		s.eventHub.Publish(ctx, mindv3.EventDomain_EVENT_DOMAIN_NOTE, mindv3.EventType_EVENT_TYPE_DELETED, id)
	}

	return nil
}

// ============================================================================
// Query Methods - List and Count with Filters
// ============================================================================

func (s *NotesService) ListNotesByCollectionID(ctx context.Context, collectionID int64) ([]store.Note, error) {
	notes, err := s.store.ListNotesByCollectionID(ctx, collectionID)
	if err != nil {
		s.logger.Error("failed to list notes by collection", "collection_id", collectionID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}

// ListNotesByCollectionIDPaginated returns notes in a collection with pagination.
func (s *NotesService) ListNotesByCollectionIDPaginated(ctx context.Context, collectionID int64, limit, offset int32) ([]store.Note, error) {
	notes, err := s.store.ListNotesByCollectionIDPaginated(ctx, store.ListNotesByCollectionIDPaginatedParams{
		CollectionID: collectionID,
		Limit:        int64(limit),
		Offset:       int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list notes by collection paginated", "collection_id", collectionID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}

// CountNotesByCollectionID returns the total number of notes in a collection.
func (s *NotesService) CountNotesByCollectionID(ctx context.Context, collectionID int64) (int64, error) {
	count, err := s.store.CountNotesByCollectionID(ctx, collectionID)
	if err != nil {
		s.logger.Error("failed to count notes by collection", "collection_id", collectionID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// ListNotesByNoteTypeIDPaginated returns notes of a specific type with pagination.
func (s *NotesService) ListNotesByNoteTypeIDPaginated(ctx context.Context, noteTypeID sql.NullInt64, limit, offset int32) ([]store.Note, error) {
	notes, err := s.store.ListNotesByNoteTypeIDPaginated(ctx, store.ListNotesByNoteTypeIDPaginatedParams{
		NoteTypeID: noteTypeID,
		Limit:      int64(limit),
		Offset:     int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list notes by type paginated", "note_type_id", noteTypeID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}

// CountNotesByNoteTypeID returns the total number of notes of a specific type.
func (s *NotesService) CountNotesByNoteTypeID(ctx context.Context, noteTypeID sql.NullInt64) (int64, error) {
	count, err := s.store.CountNotesByNoteTypeID(ctx, noteTypeID)
	if err != nil {
		s.logger.Error("failed to count notes by type", "note_type_id", noteTypeID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// ListNotesByIsTemplatePaginated returns notes filtered by template flag with pagination.
func (s *NotesService) ListNotesByIsTemplatePaginated(ctx context.Context, isTemplate sql.NullBool, limit, offset int32) ([]store.Note, error) {
	notes, err := s.store.ListNotesByIsTemplatePaginated(ctx, store.ListNotesByIsTemplatePaginatedParams{
		IsTemplate: isTemplate,
		Limit:      int64(limit),
		Offset:     int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list notes by template paginated", "is_template", isTemplate, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
}

// CountNotesByIsTemplate returns the total number of notes matching template flag.
func (s *NotesService) CountNotesByIsTemplate(ctx context.Context, isTemplate sql.NullBool) (int64, error) {
	count, err := s.store.CountNotesByIsTemplate(ctx, isTemplate)
	if err != nil {
		s.logger.Error("failed to count notes by template", "is_template", isTemplate, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// FindNotesPaginated finds notes by title and optional filters with pagination.
// Default behavior: searches globally across all collections.
// Filters can narrow scope (collection_id, note_type_id, is_template).
// Returns notes with collection_path populated from JOIN for "where is it?" UX.
// Used by UI pickers and Brain service for structured metadata queries.
func (s *NotesService) FindNotesPaginated(ctx context.Context, params store.FindNotesParams) ([]store.FindNotesRow, error) {
	notes, err := s.store.FindNotes(ctx, params)
	if err != nil {
		s.logger.Error(
			"failed to find notes",
			"title", params.Title,
			"collection_id", params.CollectionID,
			"err", err,
			"request_id", middleware.GetRequestID(ctx),
		)
	}
	return notes, err
}

// CountFindNotes counts notes matching find criteria.
// Uses same filter logic as FindNotesPaginated for consistency.
func (s *NotesService) CountFindNotes(ctx context.Context, params store.CountFindNotesParams) (int64, error) {
	count, err := s.store.CountFindNotes(ctx, params)
	if err != nil {
		s.logger.Error(
			"failed to count find notes",
			"title", params.Title,
			"collection_id", params.CollectionID,
			"err", err,
			"request_id", middleware.GetRequestID(ctx),
		)
	}
	return count, err
}

// ============================================================================
// Internal Helper Methods - Parsing and Data Extraction
// ============================================================================

// extractAndMergeTags merges tags from frontmatter ('tags'/'tag' keys) and body hashtags.
// Returns deduplicated list of all tags.
func (s *NotesService) extractAndMergeTags(parsed *markdown.ParseResult) []string {
	tagSet := make(map[string]bool)

	for _, tag := range parsed.Hashtags {
		tagSet[tag] = true
	}

	if parsed.Metadata != nil {
		for _, key := range []string{"tags", "tag"} {
			if tagsVal, exists := parsed.Metadata[key]; exists {
				switch v := tagsVal.(type) {
				case string:
					tagSet[v] = true
				case []string:
					for _, tag := range v {
						tagSet[tag] = true
					}
				case []any:
					for _, item := range v {
						if tagStr, ok := item.(string); ok {
							tagSet[tagStr] = true
						}
					}
				}
			}
		}
	}

	result := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}

	return result
}

// insertWikiLinksWithStore creates link records for all wiki-links found in the note body.
// Only creates links to existing notes - missing targets are skipped.
func (s *NotesService) insertWikiLinksWithStore(ctx context.Context, querier store.Querier, sourceNoteID int64, parsed *markdown.ParseResult) error {
	if len(parsed.WikiLinks) == 0 {
		return nil
	}

	for _, link := range parsed.WikiLinks {
		targetNote, err := querier.GetNoteByTitleGlobal(ctx, link.Target)
		if err != nil {
			s.logger.Debug("wiki-link target not found", "title", link.Target, "source_note_id", sourceNoteID)
			continue
		}

		params := store.CreateLinkParams{
			SrcID:   sourceNoteID,
			DestID:  utils.NullInt64(targetNote.ID),
			IsEmbed: utils.NullBool(link.Embed),
		}

		if link.DisplayText != "" && link.DisplayText != link.Target {
			params.DisplayText = utils.NullString(link.DisplayText)
		}

		if _, err := querier.CreateLink(ctx, params); err != nil {
			return err
		}
	}

	return nil
}

// insertTagsWithStore creates or reuses tags and associates them with the note.
// Creates new tags if they don't exist. Tags are already deduplicated by extractAndMergeTags.
func (s *NotesService) insertTagsWithStore(ctx context.Context, querier store.Querier, noteID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// TODO: optimize by using the helper bulk insert methods in the sqlcext package
	for _, tagName := range tags {
		tag, getErr := querier.GetTagByName(ctx, tagName)
		if getErr != nil {
			if errors.Is(getErr, sql.ErrNoRows) {
				tagID, createErr := querier.CreateTag(ctx, tagName)
				if createErr != nil {
					return createErr
				}
				tag.ID = tagID
				s.logger.Debug("created new tag", "name", tagName, "tag_id", tagID)
			} else {
				return getErr
			}
		}

		noteTagErr := querier.CreateNoteTag(ctx, store.CreateNoteTagParams{
			NoteID: noteID,
			TagID:  tag.ID,
		})
		if noteTagErr != nil {
			return noteTagErr
		}
	}

	return nil
}

// insertMetadataWithStore stores metadata key-value pairs from frontmatter.
// Merges with optional system metadata (frontmatter wins on conflicts).
// Filters out 'tags'/'tag' keys which are handled separately.
func (s *NotesService) insertMetadataWithStore(ctx context.Context, querier store.Querier, noteID int64, parsed *markdown.ParseResult, systemMeta map[string]string) error {
	mergedMeta := make(map[string]string)

	for k, v := range systemMeta {
		mergedMeta[k] = v
	}

	if parsed.Metadata != nil {
		for k, v := range parsed.Metadata {
			if k == "tags" || k == "tag" {
				continue
			}
			mergedMeta[k] = fmt.Sprint(v)
		}
	}

	for key, value := range mergedMeta {
		params := store.CreateNoteMetaParams{
			NoteID: noteID,
			Key:    key,
			Value:  utils.NullString(value),
		}
		_, err := querier.CreateNoteMeta(ctx, params)
		if err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Sub-Resource Methods - Metadata and Relationships
// ============================================================================

// GetNoteMeta retrieves all metadata for a note as a key-value map.
// Orchestrates NotesService and NoteMetaService to verify note existence and fetch metadata.
func (s *NotesService) GetNoteMeta(ctx context.Context, noteID int64, metaService *meta.NoteMetaService) (map[string]string, error) {
	_, err := s.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	metaItems, err := metaService.GetNoteMetaByNoteID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get note metadata", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}

	metadata := make(map[string]string)
	for _, item := range metaItems {
		if item.Value.Valid {
			metadata[item.Key] = item.Value.String
		}
	}

	return metadata, nil
}

// GetNoteRelationships retrieves relationship IDs for a note.
// Orchestrates NotesService, LinksService, and TagsService.
// Returns three slices: outgoing link IDs, incoming link IDs (backlinks), and tag IDs.
func (s *NotesService) GetNoteRelationships(ctx context.Context, noteID int64, linksService *links.LinksService, tagsService *tags.TagsService) ([]int64, []int64, []int64, error) {
	_, err := s.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, nil, nil, err
	}

	outgoingLinks := make([]int64, 0)
	incomingLinks := make([]int64, 0)
	tagIDs := make([]int64, 0)

	outgoingLinkRows, err := linksService.ListLinksBySrcID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get outgoing links", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, nil, nil, err
	}

	for _, link := range outgoingLinkRows {
		if link.DestID.Valid {
			outgoingLinks = append(outgoingLinks, link.DestID.Int64)
		}
	}

	incomingLinkRows, err := linksService.ListLinksByDestID(ctx, utils.NullInt64(noteID))
	if err != nil {
		s.logger.Error("failed to get incoming links", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, nil, nil, err
	}

	for _, link := range incomingLinkRows {
		incomingLinks = append(incomingLinks, link.SrcID)
	}

	tagRows, err := tagsService.ListTagsForNote(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get tags", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, nil, nil, err
	}

	for _, tag := range tagRows {
		tagIDs = append(tagIDs, tag.ID)
	}

	return outgoingLinks, incomingLinks, tagIDs, nil
}
