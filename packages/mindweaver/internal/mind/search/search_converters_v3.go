package search

import (
	mindv3 "github.com/nkapatos/mindweaver/internal/mind/gen/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SearchResultToProto(result SearchResult) *mindv3.SearchResult {
	return &mindv3.SearchResult{
		Id:         result.ID,
		Title:      result.Title,
		Snippet:    result.Snippet,
		Score:      result.Score,
		CreateTime: timestamppb.New(result.CreatedAt),
	}
}

func SearchResultsToProto(results []SearchResult) []*mindv3.SearchResult {
	protoResults := make([]*mindv3.SearchResult, len(results))
	for i, result := range results {
		protoResults[i] = SearchResultToProto(result)
	}
	return protoResults
}

func ProtoSearchRequestToQuery(req *mindv3.SearchNotesRequest) SearchQuery {
	limit := 10
	if req.Limit != nil {
		limit = int(*req.Limit)
	}

	offset := 0
	if req.Offset != nil {
		offset = int(*req.Offset)
	}

	includeBody := false
	if req.IncludeBody != nil {
		includeBody = *req.IncludeBody
	}

	minScore := 0.0
	if req.MinScore != nil {
		minScore = *req.MinScore
	}

	return SearchQuery{
		Query:       req.Query,
		Limit:       limit,
		Offset:      offset,
		IncludeBody: includeBody,
		MinScore:    minScore,
	}
}
