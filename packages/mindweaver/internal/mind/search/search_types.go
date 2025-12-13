package search

// SearchRequest represents the API request for searching notes.
type SearchRequest struct {
	Query       string   `json:"query" query:"q"`                             // Search query text (required)
	Limit       *int     `json:"limit,omitempty" query:"limit"`               // Max results (default: 10, max: 100)
	Offset      *int     `json:"offset,omitempty" query:"offset"`             // Pagination offset (default: 0)
	IncludeBody *bool    `json:"include_body,omitempty" query:"include_body"` // Include full body (default: false, returns snippets)
	MinScore    *float64 `json:"min_score,omitempty" query:"min_score"`       // Minimum relevance score filter
}

// SearchResultResponse represents a single search result in the API response.
type SearchResultResponse struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Snippet   string  `json:"snippet"`    // Body snippet or full body
	Score     float64 `json:"score"`      // Relevance score (0-1)
	CreatedAt string  `json:"created_at"` // RFC3339 timestamp
}

// SearchAPIResponse represents the API response for search results.
type SearchAPIResponse struct {
	Results  []SearchResultResponse `json:"results"`
	Total    int                    `json:"total"`    // Total matches (for pagination)
	Query    string                 `json:"query"`    // Original query
	Duration int64                  `json:"duration"` // Search duration in ms
	Limit    int                    `json:"limit"`    // Applied limit
	Offset   int                    `json:"offset"`   // Applied offset
}
