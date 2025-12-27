package notes

import (
	"context"
	"errors"
	"strconv"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/internal/mind/links"
	"github.com/nkapatos/mindweaver/internal/mind/meta"
	"github.com/nkapatos/mindweaver/internal/mind/tags"
	apierrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/pagination"
	"github.com/nkapatos/mindweaver/shared/utils"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NotesHandler implements the Connect-RPC NotesService handlers.
// Orchestrates multiple services (notes, meta, links, tags) for sub-resource endpoints.
type NotesHandler struct {
	mindv3connect.UnimplementedNotesServiceHandler
	service      *NotesService
	metaService  *meta.NoteMetaService
	linksService *links.LinksService
	tagsSvc      *tags.TagsService
}

// NewNotesHandler creates a new NotesHandler.
func NewNotesHandler(service *NotesService, metaService *meta.NoteMetaService, linksService *links.LinksService, tagsSvc *tags.TagsService) *NotesHandler {
	return &NotesHandler{
		service:      service,
		metaService:  metaService,
		linksService: linksService,
		tagsSvc:      tagsSvc,
	}
}

func (h *NotesHandler) CreateNote(
	ctx context.Context,
	req *connect.Request[mindv3.CreateNoteRequest],
) (*connect.Response[mindv3.Note], error) {
	params := ProtoCreateNoteToStore(req.Msg)

	noteID, err := h.service.CreateNote(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNoteAlreadyExists) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "notes", "title", req.Msg.Title)
		}
		if apierrors.IsForeignKeyConstraintError(err) {
			return nil, apierrors.NewInvalidArgumentError("collection_id or note_type_id", "referenced resource does not exist")
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to create note", err)
	}

	note, err := h.service.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve created note", err)
	}

	return connect.NewResponse(StoreNoteToProto(note)), nil
}

func (h *NotesHandler) GetNote(
	ctx context.Context,
	req *connect.Request[mindv3.GetNoteRequest],
) (*connect.Response[mindv3.Note], error) {
	note, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get note", err)
	}

	return connect.NewResponse(StoreNoteToProto(note)), nil
}

func (h *NotesHandler) ReplaceNote(
	ctx context.Context,
	req *connect.Request[mindv3.ReplaceNoteRequest],
) (*connect.Response[mindv3.Note], error) {
	current, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get note", err)
	}

	// Optimistic locking via ETag
	if req.Header().Get("If-Match") != "" {
		currentETag := utils.ComputeHashedETag(current.Version)
		if req.Header().Get("If-Match") != currentETag {
			metadata := map[string]string{
				"provided_etag": req.Header().Get("If-Match"),
				"current_etag":  currentETag,
				"header":        "If-Match",
			}
			return nil, apierrors.NewFailedPreconditionError(apierrors.MindDomain, "ETAG_MISMATCH", metadata)
		}
	}

	params := ProtoReplaceNoteToStore(req.Msg, current)

	err = h.service.UpdateNote(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNoteAlreadyExists) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "notes", "title", req.Msg.Title)
		}
		if apierrors.IsForeignKeyConstraintError(err) {
			return nil, apierrors.NewInvalidArgumentError("collection_id or note_type_id", "referenced resource does not exist")
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to replace note", err)
	}

	updated, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve replaced note", err)
	}

	return connect.NewResponse(StoreNoteToProto(updated)), nil
}

func (h *NotesHandler) DeleteNote(
	ctx context.Context,
	req *connect.Request[mindv3.DeleteNoteRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get note", err)
	}

	err = h.service.DeleteNote(ctx, req.Msg.Id)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to delete note", err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *NotesHandler) ListNotes(
	ctx context.Context,
	req *connect.Request[mindv3.ListNotesRequest],
) (*connect.Response[mindv3.ListNotesResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.PageSize, req.Msg.PageToken)
	params := pageReq.ToParams()

	var notes []store.Note
	var totalCount int64
	var err error
	var countErr error

	// Determine which query to use based on filters
	// Priority: collection_id > note_type_id > is_template > all
	if req.Msg.CollectionId != nil {
		notes, err = h.service.ListNotesByCollectionIDPaginated(ctx, *req.Msg.CollectionId, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, countErr = h.service.CountNotesByCollectionID(ctx, *req.Msg.CollectionId)
		}
	} else if req.Msg.NoteTypeId != nil {
		noteTypeID := utils.NullInt64(*req.Msg.NoteTypeId)
		notes, err = h.service.ListNotesByNoteTypeIDPaginated(ctx, noteTypeID, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, countErr = h.service.CountNotesByNoteTypeID(ctx, noteTypeID)
		}
	} else if req.Msg.IsTemplate != nil {
		isTemplate := utils.NullBool(*req.Msg.IsTemplate)
		notes, err = h.service.ListNotesByIsTemplatePaginated(ctx, isTemplate, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, countErr = h.service.CountNotesByIsTemplate(ctx, isTemplate)
		}
	} else {
		notes, err = h.service.ListNotesPaginated(ctx, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, countErr = h.service.CountNotes(ctx)
		}
	}

	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to list notes", err)
	}

	// Count errors are logged in service but don't fail the request
	// totalCount will be 0 if count failed, which is acceptable for pagination
	_ = countErr

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(notes), totalCount)
	notes = pagination.TrimResults(notes, pageReq.PageSize)

	// Convert to proto
	protoNotes := StoreNotesToProto(notes)

	// Apply field mask if provided
	if req.Msg.FieldMask != nil && *req.Msg.FieldMask != "" {
		protoNotes = ApplyFieldMaskToNotes(protoNotes, *req.Msg.FieldMask)
	}

	resp := &mindv3.ListNotesResponse{
		Notes:         protoNotes,
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *NotesHandler) GetNoteMeta(
	ctx context.Context,
	req *connect.Request[mindv3.GetNoteMetaRequest],
) (*connect.Response[mindv3.GetNoteMetaResponse], error) {
	metadata, err := h.service.GetNoteMeta(ctx, req.Msg.NoteId, h.metaService)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note", strconv.FormatInt(req.Msg.NoteId, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get note metadata", err)
	}

	resp := &mindv3.GetNoteMetaResponse{
		Metadata: metadata,
	}

	return connect.NewResponse(resp), nil
}

func (h *NotesHandler) GetNoteRelationships(
	ctx context.Context,
	req *connect.Request[mindv3.GetNoteRelationshipsRequest],
) (*connect.Response[mindv3.GetNoteRelationshipsResponse], error) {
	outgoingLinks, incomingLinks, tagIDs, err := h.service.GetNoteRelationships(ctx, req.Msg.NoteId, h.linksService, h.tagsSvc)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note", strconv.FormatInt(req.Msg.NoteId, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get note relationships", err)
	}

	resp := &mindv3.GetNoteRelationshipsResponse{
		OutgoingLinks: outgoingLinks,
		IncomingLinks: incomingLinks,
		TagIds:        tagIDs,
	}

	return connect.NewResponse(resp), nil
}

func (h *NotesHandler) NewNote(
	ctx context.Context,
	req *connect.Request[mindv3.NewNoteRequest],
) (*connect.Response[mindv3.Note], error) {
	collectionID, templateID := ProtoNewNoteToParams(req.Msg)
	noteID, err := h.service.NewNoteCreation(ctx, collectionID, templateID)
	if err != nil {
		if errors.Is(err, ErrNoteAlreadyExists) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "notes", "title", "auto-generated title")
		}
		if errors.Is(err, ErrNoteNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "template", strconv.FormatInt(templateID, 10))
		}
		if apierrors.IsForeignKeyConstraintError(err) {
			return nil, apierrors.NewInvalidArgumentError("collection_id or template_id", "referenced resource does not exist")
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to create new note", err)
	}
	note, err := h.service.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve created note", err)
	}
	return connect.NewResponse(StoreNoteToProto(note)), nil
}

// FindNotes implements the AIP-136 :find custom method for notes.
// Searches notes by title and optional filters (collection, type, template).
// Default behavior: global search across all collections.
// Always includes collection_path in results for "where is it?" UX.
func (h *NotesHandler) FindNotes(
	ctx context.Context,
	req *connect.Request[mindv3.FindNotesRequest],
) (*connect.Response[mindv3.FindNotesResponse], error) {
	// Parse pagination request (with defaults for optional fields)
	pageSize := int32(0)
	if req.Msg.PageSize != nil {
		pageSize = *req.Msg.PageSize
	}
	pageToken := ""
	if req.Msg.PageToken != nil {
		pageToken = *req.Msg.PageToken
	}
	pageReq := pagination.ParseRequest(pageSize, pageToken)
	params := pageReq.ToParams()

	// Build find parameters (all filters are optional)
	findParams := store.FindNotesParams{
		Title:        req.Msg.Title,
		CollectionID: req.Msg.CollectionId,
		NoteTypeID:   req.Msg.NoteTypeId,
		IsTemplate:   req.Msg.IsTemplate,
		Limit:        int64(params.Limit),
		Offset:       int64(params.Offset),
	}

	// Execute find query
	rows, err := h.service.FindNotesPaginated(ctx, findParams)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to find notes", err)
	}

	// Get total count (only on first page)
	var totalCount int64
	var countErr error
	if pageReq.IsFirstPage() {
		countParams := store.CountFindNotesParams{
			Title:        req.Msg.Title,
			CollectionID: req.Msg.CollectionId,
			NoteTypeID:   req.Msg.NoteTypeId,
			IsTemplate:   req.Msg.IsTemplate,
		}
		totalCount, countErr = h.service.CountFindNotes(ctx, countParams)
		// Count errors are logged in service but don't fail the request
		_ = countErr
	}

	// Convert rows to proto notes
	protoNotes := make([]*mindv3.Note, 0, len(rows))
	for _, row := range rows {
		protoNotes = append(protoNotes, FindNotesRowToProto(row))
	}

	// Apply field mask if specified
	if req.Msg.FieldMask != nil && *req.Msg.FieldMask != "" {
		protoNotes = ApplyFieldMaskToNotes(protoNotes, *req.Msg.FieldMask)
	}

	// Build pagination response
	pagResp := pageReq.BuildResponse(len(rows), totalCount)
	protoNotes = pagination.TrimResults(protoNotes, pageReq.PageSize)

	resp := &mindv3.FindNotesResponse{
		Notes:         protoNotes,
		NextPageToken: pagResp.NextPageToken,
	}

	// Include total size only on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(totalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}
