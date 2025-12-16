package links

import (
	"context"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/pagination"
)

type LinksHandlerV3 struct {
	mindv3connect.UnimplementedLinksServiceHandler
	service *LinksService
}

func NewLinksHandlerV3(service *LinksService) *LinksHandlerV3 {
	return &LinksHandlerV3{service: service}
}

// Links are derived from wikilinks in note body - read-only List operation only

func (h *LinksHandlerV3) ListLinks(
	ctx context.Context,
	req *connect.Request[mindv3.ListLinksRequest],
) (*connect.Response[mindv3.ListLinksResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.GetPageSize(), req.Msg.GetPageToken())

	// For now, just get all links (pagination not yet implemented in service)
	links, err := h.service.ListLinks(ctx)
	if err != nil {
		return nil, newInternalError("failed to list links", err)
	}

	// Apply pagination to results
	totalCount := int64(len(links))
	pageResp := pageReq.BuildResponse(len(links), totalCount)
	links = pagination.TrimResults(links, pageReq.PageSize)

	resp := &mindv3.ListLinksResponse{
		Links:         StoreLinksToProto(links),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}
