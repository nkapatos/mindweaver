package notes

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/mind/scheduler"
	"github.com/nkapatos/mindweaver/internal/mind/store"
	"github.com/nkapatos/mindweaver/pkg/dberrors"
	"github.com/nkapatos/mindweaver/pkg/markdown"
	"github.com/nkapatos/mindweaver/pkg/middleware"
)

// NotesService provides business logic for notes (CRUD operations).
type NotesService struct {
	store     store.Querier
	db        *sql.DB // Database for transaction support
	logger    *slog.Logger
	scheduler *scheduler.ChangeAccumulator // Optional: may be nil if scheduler disabled
	parser    *markdown.Parser             // Markdown parser for wiki-link extraction
}

// NewNotesService creates a new NotesService.
func NewNotesService(db *sql.DB, store store.Querier, logger *slog.Logger, serviceName string) *NotesService {
	return &NotesService{
		store:     store,
		db:        db,
		logger:    logger.With("service", serviceName),
		scheduler: nil,                  // Will be set via SetScheduler() if enabled
		parser:    markdown.NewParser(), // Initialize markdown parser
	}
}

// SetScheduler sets the change scheduler for Brain synchronization.
// This is optional and can be called after service creation if scheduler is enabled.
func (s *NotesService) SetScheduler(scheduler *scheduler.ChangeAccumulator) {
	s.scheduler = scheduler
	s.logger.Info("scheduler enabled for note service")
}

// GetMarkdownParser returns the markdown parser instance for use by handlers
func (s *NotesService) GetMarkdownParser() *markdown.Parser {
	return s.parser
}

// ListNotes returns all notes.
func (s *NotesService) ListNotes(ctx context.Context) ([]store.Note, error) {
	notes, err := s.store.ListNotes(ctx)
	if err != nil {
		s.logger.Error("failed to list notes", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return notes, err
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
		s.logger.Error("failed to get note by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return note, err
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
	defer tx.Rollback() // Will be no-op if committed

	// Create transactional store
	txStore := store.New(tx)

	// Create note
	id, err := txStore.CreateNote(ctx, params)
	if err != nil {
		// Translate DB errors to domain errors
		if dberrors.IsUniqueConstraintError(err) {
			// Could be UUID collision or (collection_id, title) duplicate
			return 0, ErrNoteAlreadyExists
		}
		s.logger.Error("failed to create note", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}

	// Parse body once to extract all derived data (links, tags, metadata)
	if params.Body.Valid && params.Body.String != "" {
		parsed, err := s.parser.Parse([]byte(params.Body.String))
		if err != nil {
			s.logger.Error("failed to parse note body", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err
		}

		// Extract and merge tags from frontmatter and body
		allTags := s.extractAndMergeTags(parsed)

		// Insert wiki-links from parsed result
		if err := s.insertWikiLinksWithStore(ctx, txStore, id, parsed); err != nil {
			s.logger.Error("failed to insert wiki-links", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err // Fail the entire operation
		}

		// Insert merged tags (frontmatter + body hashtags)
		if err := s.insertTagsWithStore(ctx, txStore, id, allTags); err != nil {
			s.logger.Error("failed to insert tags", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err // Fail the entire operation
		}

		// Insert metadata from parsed result (with optional system meta merge)
		// Note: 'tags'/'tag' keys are filtered out and handled separately above
		if err := s.insertMetadataWithStore(ctx, txStore, id, parsed, nil); err != nil {
			s.logger.Error("failed to insert metadata", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
			return 0, err // Fail the entire operation
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}

	s.logger.Info("note created", "id", id, "request_id", middleware.GetRequestID(ctx))

	// Notify Brain of new note (after successful commit)
	if s.scheduler != nil {
		s.scheduler.TrackChange("note_created", id)
	}

	return id, nil
}

// UpdateNote updates an existing note with derived data (links, tags) atomically.
// All operations are wrapped in a transaction to ensure consistency.
func (s *NotesService) UpdateNote(ctx context.Context, params store.UpdateNoteByIDParams) error {
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("failed to begin transaction", "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	defer tx.Rollback() // Will be no-op if committed

	// Create transactional store
	txStore := store.New(tx)

	// Delete existing links before updating
	if err := txStore.DeleteNotesLinksBySrcID(ctx, params.ID); err != nil {
		s.logger.Error("failed to delete existing links", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	// Delete existing tags before updating
	if err := txStore.DeleteNoteTagsByNoteID(ctx, params.ID); err != nil {
		s.logger.Error("failed to delete existing tags", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	// Delete existing metadata before updating
	if err := txStore.DeleteNoteMetaByNoteID(ctx, params.ID); err != nil {
		s.logger.Error("failed to delete existing metadata", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	// Update note
	err = txStore.UpdateNoteByID(ctx, params)
	if err != nil {
		// Translate DB errors to domain errors
		if dberrors.IsUniqueConstraintError(err) {
			// Title conflict within same collection
			return ErrNoteAlreadyExists
		}
		s.logger.Error("failed to update note", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	// Parse body once to extract all derived data (links, tags, metadata)
	if params.Body.Valid && params.Body.String != "" {
		parsed, err := s.parser.Parse([]byte(params.Body.String))
		if err != nil {
			s.logger.Error("failed to parse note body", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err
		}

		// Insert wiki-links from parsed result
		if err := s.insertWikiLinksWithStore(ctx, txStore, params.ID, parsed); err != nil {
			s.logger.Error("failed to insert wiki-links", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err // Fail the entire operation
		}

		// Extract and merge tags from frontmatter and body hashtags
		allTags := s.extractAndMergeTags(parsed)

		// Insert all tags (merged from frontmatter + body hashtags)
		if err := s.insertTagsWithStore(ctx, txStore, params.ID, allTags); err != nil {
			s.logger.Error("failed to insert tags", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err // Fail the entire operation
		}

		// Insert metadata from parsed result (with optional system meta merge)
		if err := s.insertMetadataWithStore(ctx, txStore, params.ID, parsed, nil); err != nil {
			s.logger.Error("failed to insert metadata", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err // Fail the entire operation
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction", "note_id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	s.logger.Info("note updated", "id", params.ID, "request_id", middleware.GetRequestID(ctx))

	// Notify Brain of updated note (after successful commit)
	if s.scheduler != nil {
		s.scheduler.TrackChange("note_updated", params.ID)
	}

	return nil
}

// DeleteNote deletes a note by ID.
func (s *NotesService) DeleteNote(ctx context.Context, id int64) error {
	err := s.store.DeleteNoteByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete note", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("note deleted", "id", id, "request_id", middleware.GetRequestID(ctx))

	// Notify Brain of deleted note
	if s.scheduler != nil {
		s.scheduler.TrackChange("note_deleted", id)
	}

	return nil
}

// ============================================================================
// Enhanced Service Methods for Field Selection and Includes
// ============================================================================
// NoteWithRelations is defined in note_extended_types.go

// GetNoteWithRelations returns a note with specified relations included
func (s *NotesService) GetNoteWithRelations(ctx context.Context, id int64, includes []string) (*NoteWithRelations, error) {
	// Get base note
	note, err := s.store.GetNoteByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &NoteWithRelations{
		Note: note,
	}

	// Include relations based on request
	for _, include := range includes {
		switch include {
		case "type":
			if note.NoteTypeID.Valid {
				noteType, err := s.store.GetNoteTypeByID(ctx, note.NoteTypeID.Int64)
				if err == nil {
					result.Type = &noteType
				}
			}
		case "meta":
			meta, err := s.store.GetNoteMetaByNoteID(ctx, id)
			if err == nil {
				result.Meta = make(map[string]string)
				for _, m := range meta {
					if m.Value.Valid {
						result.Meta[m.Key] = m.Value.String
					}
				}
			}
		case "tags":
			tags, err := s.store.ListTagsForNote(ctx, id)
			if err == nil {
				result.Tags = make([]string, len(tags))
				for i, t := range tags {
					result.Tags[i] = t.Name
				}
			}
		}
	}

	return result, nil
}

// ListNotesByFilters returns notes filtered by various criteria
// Note: For now, only single filter supported due to sqlc limitations
func (s *NotesService) ListNotesByFilters(ctx context.Context, tagID int64, metaKey string) ([]store.Note, error) {
	if tagID > 0 {
		return s.store.ListNotesByTagIDs(ctx, tagID)
	}
	if metaKey != "" {
		return s.store.ListNotesByMetaKeys(ctx, metaKey)
	}

	// Default to all notes
	return s.store.ListNotes(ctx)
}

// ListNotesByCollectionID returns all notes in a specific collection.
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

// extractAndMergeTags extracts tags from frontmatter metadata and merges with body hashtags.
// Returns deduplicated slice of all tags.
func (s *NotesService) extractAndMergeTags(parsed *markdown.ParseResult) []string {
	tagSet := make(map[string]bool)

	// Add body hashtags
	for _, tag := range parsed.Hashtags {
		tagSet[tag] = true
	}

	// Extract tags from frontmatter metadata
	// Check for both 'tags' and 'tag' keys
	if parsed.Metadata != nil {
		for _, key := range []string{"tags", "tag"} {
			if tagsVal, exists := parsed.Metadata[key]; exists {
				// Handle different types: string, []string, []interface{}
				switch v := tagsVal.(type) {
				case string:
					// Single tag as string
					tagSet[v] = true
				case []string:
					// Array of strings
					for _, tag := range v {
						tagSet[tag] = true
					}
				case []interface{}:
					// Array of any (typical from YAML)
					for _, item := range v {
						if tagStr, ok := item.(string); ok {
							tagSet[tagStr] = true
						}
					}
				}
			}
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}

	return result
}

// insertWikiLinksWithStore inserts wiki-links from parsed result using the provided store
func (s *NotesService) insertWikiLinksWithStore(ctx context.Context, querier store.Querier, sourceNoteID int64, parsed *markdown.ParseResult) error {
	if len(parsed.WikiLinks) == 0 {
		return nil // No links to insert
	}

	// For each wiki-link, find the target note and create the link
	for _, link := range parsed.WikiLinks {
		// Try to find existing note by title (search globally across collections)
		targetNote, err := querier.GetNoteByTitleGlobal(ctx, link.Target)
		if err != nil {
			// Note doesn't exist yet - skip this link for now
			// In a real implementation, you might want to create a placeholder note
			s.logger.Debug("wiki-link target not found", "title", link.Target, "source_note_id", sourceNoteID)
			continue
		}

		// Create the link
		params := store.CreateNotesLinkParams{
			SrcID:   sourceNoteID,
			DestID:  sql.NullInt64{Int64: targetNote.ID, Valid: true},
			IsEmbed: sql.NullBool{Bool: link.Embed, Valid: true},
		}

		// Add display text if it differs from title
		if link.DisplayText != "" && link.DisplayText != link.Target {
			params.DisplayText = sql.NullString{String: link.DisplayText, Valid: true}
		}

		if _, err := querier.CreateNotesLink(ctx, params); err != nil {
			// Return error to roll back transaction
			return err
		}
	}

	return nil
}

// insertTagsWithStore inserts tags (merged from frontmatter and body hashtags) using the provided store.
// Tags are already deduplicated by extractAndMergeTags.
func (s *NotesService) insertTagsWithStore(ctx context.Context, querier store.Querier, noteID int64, tags []string) error {
	if len(tags) == 0 {
		return nil // No tags to insert
	}

	// Get or create tags and associate with note
	for _, tagName := range tags {
		// Try to get existing tag
		tag, err := querier.GetTagByName(ctx, tagName)
		if err != nil {
			if err == sql.ErrNoRows {
				// Tag doesn't exist, create it
				tagID, err := querier.CreateTag(ctx, tagName)
				if err != nil {
					return err // Return error to roll back transaction
				}
				tag.ID = tagID
				tag.Name = tagName
				s.logger.Debug("created new tag", "name", tagName, "tag_id", tagID)
			} else {
				return err // Return error to roll back transaction
			}
		}

		// Create note-tag association
		err = querier.CreateNoteTag(ctx, store.CreateNoteTagParams{
			NoteID: noteID,
			TagID:  tag.ID,
		})
		if err != nil {
			return err // Return error to roll back transaction
		}
	}

	return nil
}

// insertMetadataWithStore inserts metadata from parsed result using the provided store.
// systemMeta is optional metadata from system/plugins. In case of key conflicts, frontmatter takes precedence.
func (s *NotesService) insertMetadataWithStore(ctx context.Context, querier store.Querier, noteID int64, parsed *markdown.ParseResult, systemMeta map[string]string) error {
	// Merge system meta and frontmatter meta (frontmatter wins on conflicts)
	mergedMeta := make(map[string]string)

	// Start with system meta
	if systemMeta != nil {
		for k, v := range systemMeta {
			mergedMeta[k] = v
		}
	}

	// Overlay frontmatter meta (overwrites system meta on conflicts)
	if parsed.Metadata != nil {
		for k, v := range parsed.Metadata {
			// Skip 'tags' and 'tag' keys - these go to note_tags table instead
			if k == "tags" || k == "tag" {
				continue
			}
			// Convert any type to string
			mergedMeta[k] = fmt.Sprint(v)
		}
	}

	// Insert merged metadata
	for key, value := range mergedMeta {
		params := store.CreateNoteMetaParams{
			NoteID: noteID,
			Key:    key,
			Value:  sql.NullString{String: value, Valid: true},
		}
		_, err := querier.CreateNoteMeta(ctx, params)
		if err != nil {
			return err // Return error to roll back transaction
		}
	}

	return nil
}

// ============================================================================
// Extended Service Methods for Brain Adapter Integration
// ============================================================================

// GetNoteWithMetadata retrieves a note with its complete metadata, tags, and type.
// Returns the note, metadata map, and tags list.
// This is a convenience method for Brain adapters that need all related data.
func (s *NotesService) GetNoteWithMetadata(ctx context.Context, id int64) (store.Note, map[string]string, []store.Tag, error) {
	// Get base note
	note, err := s.store.GetNoteByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get note", "note_id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Note{}, nil, nil, err
	}

	// Get metadata
	metaList, err := s.store.GetNoteMetaByNoteID(ctx, id)
	metadata := make(map[string]string)
	if err == nil {
		for _, m := range metaList {
			if m.Value.Valid {
				metadata[m.Key] = m.Value.String
			}
		}
	} else {
		s.logger.Warn("failed to get note metadata", "note_id", id, "err", err)
	}

	// Get tags
	tags, err := s.store.ListTagsForNote(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get note tags", "note_id", id, "err", err)
		tags = []store.Tag{} // Return empty tags rather than failing
	}

	return note, metadata, tags, nil
}

// GetNoteMeta retrieves metadata for a note as a map.
// This orchestrates NotesService and NoteMetaService.
func (s *NotesService) GetNoteMeta(ctx context.Context, noteID int64, metaService *NoteMetaService) (map[string]string, error) {
	// Verify note exists
	_, err := s.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	// Get metadata
	metaItems, err := metaService.GetNoteMetaByNoteID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get note metadata", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}

	// Convert to map
	metadata := make(map[string]string)
	for _, item := range metaItems {
		if item.Value.Valid {
			metadata[item.Key] = item.Value.String
		}
	}

	return metadata, nil
}

// NoteRelationships holds the relationship data for a note
type NoteRelationships struct {
	OutgoingLinks []int64
	IncomingLinks []int64
	TagIDs        []int64
}

// GetNoteRelationships retrieves all relationships for a note.
// This orchestrates NotesService, LinksService, and TagsService.
func (s *NotesService) GetNoteRelationships(ctx context.Context, noteID int64, linksService *LinksService, tagsService *TagsService) (*NoteRelationships, error) {
	// Verify note exists
	_, err := s.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	result := &NoteRelationships{
		OutgoingLinks: make([]int64, 0),
		IncomingLinks: make([]int64, 0),
		TagIDs:        make([]int64, 0),
	}

	// Get outgoing links
	outgoingLinks, err := linksService.ListLinksBySrcID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get outgoing links", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}

	for _, link := range outgoingLinks {
		if link.DestID.Valid {
			result.OutgoingLinks = append(result.OutgoingLinks, link.DestID.Int64)
		}
	}

	// Get incoming links (backlinks)
	incomingLinks, err := linksService.ListLinksByDestID(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get incoming links", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}

	for _, link := range incomingLinks {
		result.IncomingLinks = append(result.IncomingLinks, link.SrcID)
	}

	// Get tags
	tags, err := tagsService.ListTagsForNote(ctx, noteID)
	if err != nil {
		s.logger.Error("failed to get tags", "note_id", noteID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}

	for _, tag := range tags {
		result.TagIDs = append(result.TagIDs, tag.ID)
	}

	return result, nil
}

// GetRelatedNotes finds notes related to the given note through links and shared tags.
// Combines forward links, backward links, and tag-based relations.
// Returns up to 'limit' related notes, deduplicated by note ID.
func (s *NotesService) GetRelatedNotes(ctx context.Context, noteID int64, limit int) ([]store.Note, error) {
	if limit <= 0 {
		limit = 10
	}

	// Track seen note IDs to deduplicate
	seenIDs := make(map[int64]bool)
	relatedNotes := make([]store.Note, 0, limit)

	// 1. Get forward links (notes this note links to)
	forwardLinksRows, err := s.store.GetRelatedNotesByForwardLinks(ctx, store.GetRelatedNotesByForwardLinksParams{
		NoteID:     noteID,
		LimitCount: int64(limit),
	})
	if err == nil {
		for _, row := range forwardLinksRows {
			if !seenIDs[row.ID] && len(relatedNotes) < limit {
				note, err := s.store.GetNoteByID(ctx, row.ID)
				if err == nil {
					relatedNotes = append(relatedNotes, note)
					seenIDs[row.ID] = true
				}
			}
		}
	}

	// 2. Get backward links (notes that link to this note)
	if len(relatedNotes) < limit {
		backwardLinksRows, err := s.store.GetRelatedNotesByBackwardLinks(ctx, store.GetRelatedNotesByBackwardLinksParams{
			NoteID:     sql.NullInt64{Int64: noteID, Valid: true},
			LimitCount: int64(limit),
		})
		if err == nil {
			for _, row := range backwardLinksRows {
				if !seenIDs[row.ID] && len(relatedNotes) < limit {
					note, err := s.store.GetNoteByID(ctx, row.ID)
					if err == nil {
						relatedNotes = append(relatedNotes, note)
						seenIDs[row.ID] = true
					}
				}
			}
		}
	}

	// 3. Get tag-based relations (notes with shared tags)
	if len(relatedNotes) < limit {
		tagRelatedRows, err := s.store.GetRelatedNotesByTags(ctx, store.GetRelatedNotesByTagsParams{
			NoteID:     noteID,
			LimitCount: int64(limit),
		})
		if err == nil {
			for _, row := range tagRelatedRows {
				if !seenIDs[row.ID] && len(relatedNotes) < limit {
					note, err := s.store.GetNoteByID(ctx, row.ID)
					if err == nil {
						relatedNotes = append(relatedNotes, note)
						seenIDs[row.ID] = true
					}
				}
			}
		}
	}

	s.logger.Info("get related notes completed", "note_id", noteID, "results", len(relatedNotes), "request_id", middleware.GetRequestID(ctx))
	return relatedNotes, nil
}
