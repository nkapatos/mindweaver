package notetypes

import (
	"context"
	"errors"
	"strconv"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	apierrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NoteTypesHandler struct {
	mindv3connect.UnimplementedNoteTypesServiceHandler
	service *NoteTypesService
}

func NewNoteTypesHandler(service *NoteTypesService) *NoteTypesHandler {
	return &NoteTypesHandler{service: service}
}

func (h *NoteTypesHandler) CreateNoteType(
	ctx context.Context,
	req *connect.Request[mindv3.CreateNoteTypeRequest],
) (*connect.Response[mindv3.NoteType], error) {
	params := ProtoCreateNoteTypeToStore(req.Msg)

	noteTypeID, err := h.service.CreateNoteType(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNoteTypeAlreadyExists) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "note_types", "type", req.Msg.Type)
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to create note type", err)
	}

	noteType, err := h.service.GetNoteTypeByID(ctx, noteTypeID)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve created note type", err)
	}

	return connect.NewResponse(StoreNoteTypeToProto(noteType)), nil
}

func (h *NoteTypesHandler) ListNoteTypes(
	ctx context.Context,
	req *connect.Request[mindv3.ListNoteTypesRequest],
) (*connect.Response[mindv3.ListNoteTypesResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.GetPageSize(), req.Msg.GetPageToken())
	params := pageReq.ToParams()

	noteTypes, err := h.service.ListNoteTypesPaginated(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to list note types", err)
	}

	var totalCount int64
	var countErr error
	if pageReq.IsFirstPage() {
		totalCount, countErr = h.service.CountNoteTypes(ctx)
		// Count errors are logged in service but don't fail the request
		_ = countErr
	}

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(noteTypes), totalCount)
	noteTypes = pagination.TrimResults(noteTypes, pageReq.PageSize)

	resp := &mindv3.ListNoteTypesResponse{
		NoteTypes:     StoreNoteTypesToProto(noteTypes),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *NoteTypesHandler) GetNoteType(
	ctx context.Context,
	req *connect.Request[mindv3.GetNoteTypeRequest],
) (*connect.Response[mindv3.NoteType], error) {
	noteType, err := h.service.GetNoteTypeByID(ctx, req.Msg.GetId())
	if err != nil {
		if errors.Is(err, ErrNoteTypeNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note_types", strconv.FormatInt(req.Msg.GetId(), 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get note type", err)
	}

	return connect.NewResponse(StoreNoteTypeToProto(noteType)), nil
}

func (h *NoteTypesHandler) UpdateNoteType(
	ctx context.Context,
	req *connect.Request[mindv3.UpdateNoteTypeRequest],
) (*connect.Response[mindv3.NoteType], error) {
	params := ProtoUpdateNoteTypeToStore(req.Msg)

	err := h.service.UpdateNoteType(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNoteTypeNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note_types", strconv.FormatInt(req.Msg.GetId(), 10))
		}
		if errors.Is(err, ErrNoteTypeAlreadyExists) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "note_types", "type", req.Msg.Type)
		}
		if errors.Is(err, ErrNoteTypeIsSystem) {
			return nil, apierrors.NewPermissionDeniedError(apierrors.MindDomain, "cannot update system note type")
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to update note type", err)
	}

	noteType, err := h.service.GetNoteTypeByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve updated note type", err)
	}

	return connect.NewResponse(StoreNoteTypeToProto(noteType)), nil
}

func (h *NoteTypesHandler) DeleteNoteType(
	ctx context.Context,
	req *connect.Request[mindv3.DeleteNoteTypeRequest],
) (*connect.Response[emptypb.Empty], error) {
	err := h.service.DeleteNoteType(ctx, req.Msg.GetId())
	if err != nil {
		if errors.Is(err, ErrNoteTypeNotFound) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "note_types", strconv.FormatInt(req.Msg.GetId(), 10))
		}
		if errors.Is(err, ErrNoteTypeIsSystem) {
			return nil, apierrors.NewPermissionDeniedError(apierrors.MindDomain, "cannot delete system note type")
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to delete note type", err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
