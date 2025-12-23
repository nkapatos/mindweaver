package notes

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/links"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/meta"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/tags"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/dberrors"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/pagination"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/utils"
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
			return nil, newAlreadyExistsError("notes", "title", req.Msg.Title)
		}
		if dberrors.IsForeignKeyConstraintError(err) {
			return nil, newInvalidArgumentError("collection_id or note_type_id", "referenced resource does not exist")
		}
		return nil, newInternalError("failed to create note", err)
	}

	note, err := h.service.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, newInternalError("failed to retrieve created note", err)
	}

	return connect.NewResponse(StoreNoteToProto(note)), nil
}

func (h *NotesHandler) GetNote(
	ctx context.Context,
	req *connect.Request[mindv3.GetNoteRequest],
) (*connect.Response[mindv3.Note], error) {
	note, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("note", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, newInternalError("failed to get note", err)
	}

	return connect.NewResponse(StoreNoteToProto(note)), nil
}

func (h *NotesHandler) ReplaceNote(
	ctx context.Context,
	req *connect.Request[mindv3.ReplaceNoteRequest],
) (*connect.Response[mindv3.Note], error) {
	current, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("note", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, newInternalError("failed to get note", err)
	}

	// Optimistic locking via ETag
	if req.Header().Get("If-Match") != "" {
		currentETag := utils.ComputeHashedETag(current.Version)
		if req.Header().Get("If-Match") != currentETag {
			return nil, newETagMismatchError(req.Header().Get("If-Match"), currentETag)
		}
	}

	params := ProtoReplaceNoteToStore(req.Msg, current)

	err = h.service.UpdateNote(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNoteAlreadyExists) {
			return nil, newAlreadyExistsError("notes", "title", req.Msg.Title)
		}
		if dberrors.IsForeignKeyConstraintError(err) {
			return nil, newInvalidArgumentError("collection_id or note_type_id", "referenced resource does not exist")
		}
		return nil, newInternalError("failed to replace note", err)
	}

	updated, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, newInternalError("failed to retrieve replaced note", err)
	}

	return connect.NewResponse(StoreNoteToProto(updated)), nil
}

func (h *NotesHandler) DeleteNote(
	ctx context.Context,
	req *connect.Request[mindv3.DeleteNoteRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := h.service.GetNoteByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("note", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, newInternalError("failed to get note", err)
	}

	err = h.service.DeleteNote(ctx, req.Msg.Id)
	if err != nil {
		return nil, newInternalError("failed to delete note", err)
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

	// Determine which query to use based on filters
	// Priority: collection_id > note_type_id > is_template > all
	if req.Msg.CollectionId != nil {
		notes, err = h.service.ListNotesByCollectionIDPaginated(ctx, *req.Msg.CollectionId, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountNotesByCollectionID(ctx, *req.Msg.CollectionId)
		}
	} else if req.Msg.NoteTypeId != nil {
		noteTypeID := utils.NullInt64(*req.Msg.NoteTypeId)
		notes, err = h.service.ListNotesByNoteTypeIDPaginated(ctx, noteTypeID, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountNotesByNoteTypeID(ctx, noteTypeID)
		}
	} else if req.Msg.IsTemplate != nil {
		isTemplate := utils.NullBool(*req.Msg.IsTemplate)
		notes, err = h.service.ListNotesByIsTemplatePaginated(ctx, isTemplate, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountNotesByIsTemplate(ctx, isTemplate)
		}
	} else {
		notes, err = h.service.ListNotesPaginated(ctx, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountNotes(ctx)
		}
	}

	if err != nil {
		return nil, newInternalError("failed to list notes", err)
	}

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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("note", strconv.FormatInt(req.Msg.NoteId, 10))
		}
		return nil, newInternalError("failed to get note metadata", err)
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("note", strconv.FormatInt(req.Msg.NoteId, 10))
		}
		return nil, newInternalError("failed to get note relationships", err)
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
			return nil, newAlreadyExistsError("notes", "title", "auto-generated title")
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("template", strconv.FormatInt(templateID, 10))
		}
		if dberrors.IsForeignKeyConstraintError(err) {
			return nil, newInvalidArgumentError("collection_id or template_id", "referenced resource does not exist")
		}
		return nil, newInternalError("failed to create new note", err)
	}
	note, err := h.service.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, newInternalError("failed to retrieve created note", err)
	}
	return connect.NewResponse(StoreNoteToProto(note)), nil
}
