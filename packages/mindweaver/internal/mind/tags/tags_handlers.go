package tags

import (
	"context"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/pagination"
)

// TODO: V1 to V3 Migration - Missing Endpoints
// The following V1 read endpoints are not yet implemented in V3:
// - GetTagByID: Get single tag by ID
// - GetTagNotes: List all notes that have a specific tag (IMPORTANT - core PKM feature for tag navigation)
// - SearchTagsByName: Search tags by name pattern
// Note: Tags are derived from note content, so Create/Update/Delete are not needed.

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

	if req.Msg.NoteId != nil {
		tags, err = h.service.ListTagsForNotePaginated(ctx, *req.Msg.NoteId, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountTagsForNote(ctx, *req.Msg.NoteId)
		}
	} else {
		tags, err = h.service.ListTagsPaginated(ctx, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountTags(ctx)
		}
	}

	if err != nil {
		return nil, newInternalError("failed to list tags", err)
	}

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
