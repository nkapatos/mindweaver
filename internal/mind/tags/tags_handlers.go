package tags

import (
	"context"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	apierrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/pagination"
)

type TagsHandler struct {
	mindv3connect.UnimplementedTagsServiceHandler
	service *TagsService
}

func NewTagsHandler(service *TagsService) *TagsHandler {
	return &TagsHandler{service: service}
}

// Tags are derived from note content (frontmatter/body) - read-only List operation only

func (h *TagsHandler) ListTags(
	ctx context.Context,
	req *connect.Request[mindv3.ListTagsRequest],
) (*connect.Response[mindv3.ListTagsResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.GetPageSize(), req.Msg.GetPageToken())
	params := pageReq.ToParams()

	var tags []store.Tag
	var totalCount int64
	var err error
	var countErr error

	if req.Msg.NoteId != nil {
		tags, err = h.service.ListTagsForNotePaginated(ctx, *req.Msg.NoteId, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, countErr = h.service.CountTagsForNote(ctx, *req.Msg.NoteId)
		}
	} else {
		tags, err = h.service.ListTagsPaginated(ctx, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, countErr = h.service.CountTags(ctx)
		}
	}

	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to list tags", err)
	}

	// Count errors are logged in service but don't fail the request
	_ = countErr

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(tags), totalCount)
	tags = pagination.TrimResults(tags, pageReq.PageSize)

	resp := &mindv3.ListTagsResponse{
		Tags:          StoreTagsToProto(tags),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *TagsHandler) ListNotesForTag(
	ctx context.Context,
	req *connect.Request[mindv3.ListNotesForTagRequest],
) (*connect.Response[mindv3.ListNotesForTagResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.GetPageSize(), req.Msg.GetPageToken())
	params := pageReq.ToParams()

	notes, err := h.service.ListNotesForTagPaginated(ctx, req.Msg.TagId, params.Limit, params.Offset)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to list notes for tag", err)
	}

	// Get total count (only on first page)
	var totalCount int64
	var countErr error
	if pageReq.IsFirstPage() {
		totalCount, countErr = h.service.CountNotesForTag(ctx, req.Msg.TagId)
	}
	// Count errors are logged in service but don't fail the request
	_ = countErr

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(notes), totalCount)
	notes = pagination.TrimResults(notes, pageReq.PageSize)

	resp := &mindv3.ListNotesForTagResponse{
		Notes:         StoreNotesToProto(notes),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *TagsHandler) FindTags(
	ctx context.Context,
	req *connect.Request[mindv3.FindTagsRequest],
) (*connect.Response[mindv3.FindTagsResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.GetPageSize(), req.Msg.GetPageToken())
	params := pageReq.ToParams()

	tags, err := h.service.FindTagsPaginated(ctx, req.Msg.Name, params.Limit, params.Offset)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to find tags", err)
	}

	// Get total count (only on first page)
	var totalCount int64
	var countErr error
	if pageReq.IsFirstPage() {
		totalCount, countErr = h.service.CountFindTags(ctx, req.Msg.Name)
	}
	// Count errors are logged in service but don't fail the request
	_ = countErr

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(tags), totalCount)
	tags = pagination.TrimResults(tags, pageReq.PageSize)

	resp := &mindv3.FindTagsResponse{
		Tags:          StoreTagsToProto(tags),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}
