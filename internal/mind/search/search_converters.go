package search

// ToServiceQuery converts API request to service layer query.
func ToServiceQuery(req SearchRequest) SearchQuery {
	// Default values
	limit := 10
	offset := 0
	includeBody := false
	minScore := 0.0

	// Apply provided values
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}
	if req.IncludeBody != nil {
		includeBody = *req.IncludeBody
	}
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

// ToAPIResponse converts service response to API response.
func ToAPIResponse(resp SearchResponse, limit, offset int) SearchAPIResponse {
	results := make([]SearchResultResponse, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, SearchResultResponse{
			ID:        r.ID,
			Title:     r.Title,
			Snippet:   r.Snippet,
			Score:     r.Score,
			CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"), // RFC3339
		})
	}

	return SearchAPIResponse{
		Results:  results,
		Total:    resp.Total,
		Query:    resp.Query,
		Duration: resp.Duration,
		Limit:    limit,
		Offset:   offset,
	}
}
