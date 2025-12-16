package search

import (
	"context"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3/mindv3connect"
)

type SearchHandlerV3 struct {
	mindv3connect.UnimplementedSearchServiceHandler
	service *SearchService
}

func NewSearchHandlerV3(service *SearchService) *SearchHandlerV3 {
	return &SearchHandlerV3{service: service}
}

func (h *SearchHandlerV3) SearchNotes(
	ctx context.Context,
	req *connect.Request[mindv3.SearchNotesRequest],
) (*connect.Response[mindv3.SearchNotesResponse], error) {
	serviceQuery := ProtoSearchRequestToQuery(req.Msg)

	resp, err := h.service.Search(ctx, serviceQuery)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoResults := SearchResultsToProto(resp.Results)

	protoResp := &mindv3.SearchNotesResponse{
		Results:    protoResults,
		Total:      int32(resp.Total),
		Query:      resp.Query,
		DurationMs: resp.Duration,
		Limit:      int32(serviceQuery.Limit),
		Offset:     int32(serviceQuery.Offset),
	}

	return connect.NewResponse(protoResp), nil
}
