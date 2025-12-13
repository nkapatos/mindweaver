// Package search provides full-text search functionality for Mind notes.
// This service uses FTS5 (SQLite Full-Text Search) for efficient text searching.
//
// TODO: Future enhancements:
// - Composable search filters (e.g., search within specific meta keys, tags, or collections)
// - Multi-resource search (search across notes + collections simultaneously)
// - Advanced query syntax support (phrases, boolean operators, field-specific search)
// - Search result highlighting and context snippets
package search

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nkapatos/mindweaver/internal/mind/store"
	"github.com/nkapatos/mindweaver/pkg/middleware"
	"github.com/nkapatos/mindweaver/pkg/sqlcext"
)

// SearchService provides full-text search functionality for Mind notes.
type SearchService struct {
	store      *store.Queries      // Use concrete type to access generated queries
	ftsQuerier *sqlcext.FTSQuerier // FTS5 querier for full-text search
	logger     *slog.Logger
}

// SearchQuery represents a search request.
type SearchQuery struct {
	Query       string  // The search query text
	Limit       int     // Maximum number of results to return
	Offset      int     // Number of results to skip (for pagination)
	IncludeBody bool    // Whether to include full note body in results
	MinScore    float64 // Minimum relevance score (0-1)
}

// SearchResult represents a single search result.
type SearchResult struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Snippet   string    `json:"snippet"` // Body snippet or full body
	Score     float64   `json:"score"`   // Relevance score
	CreatedAt time.Time `json:"created_at"`
}

// SearchResponse represents the search results and metadata.
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`    // Total number of matches (for pagination)
	Query    string         `json:"query"`    // The original query
	Duration int64          `json:"duration"` // Search duration in milliseconds
}

// NewSearchService creates a new SearchService for Mind notes.
func NewSearchService(db sqlcext.DB, store *store.Queries, logger *slog.Logger) *SearchService {
	// Configure FTS querier for Mind notes
	ftsConfig := sqlcext.FTSConfig{
		ContentTable: "notes",
		FTSTable:     "notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}

	return &SearchService{
		store:      store,
		ftsQuerier: sqlcext.NewFTSQuerier(db, ftsConfig),
		logger:     logger.With("service", "mind_search"),
	}
}

// Search performs full-text search on Mind notes.
func (s *SearchService) Search(ctx context.Context, query SearchQuery) (SearchResponse, error) {
	startTime := time.Now()

	s.logger.Info("searching notes",
		"query", query.Query,
		"limit", query.Limit,
		"offset", query.Offset,
		"request_id", middleware.GetRequestID(ctx),
	)

	// Validate query
	if query.Query == "" {
		return SearchResponse{}, fmt.Errorf("search query cannot be empty")
	}
	if query.Limit <= 0 {
		query.Limit = 10 // Default limit
	}
	if query.Limit > 100 {
		query.Limit = 100 // Max limit
	}

	// Perform FTS search using sqlcext
	var ftsResults []sqlcext.FTSSearchResult
	var err error

	ftsParams := sqlcext.FTSSearchParams{
		Query:       query.Query, // FTS querier handles sanitization
		LimitCount:  int64(query.Limit),
		OffsetCount: int64(query.Offset),
	}

	if query.IncludeBody {
		// Full body search
		ftsResults, err = s.ftsQuerier.Search(ctx, ftsParams)
		if err != nil {
			s.logger.Error("fts search failed", "err", err, "query", query.Query, "request_id", middleware.GetRequestID(ctx))
			return SearchResponse{}, fmt.Errorf("search failed: %w", err)
		}
	} else {
		// Snippet-only search
		ftsResults, err = s.ftsQuerier.SearchWithSnippet(ctx, ftsParams)
		if err != nil {
			s.logger.Error("fts search with snippet failed", "err", err, "query", query.Query, "request_id", middleware.GetRequestID(ctx))
			return SearchResponse{}, fmt.Errorf("search failed: %w", err)
		}
	}

	// Convert FTS results to search results
	results := s.convertFTSResults(ftsResults)

	// Get total count for pagination
	total, err := s.ftsQuerier.Count(ctx, query.Query)
	if err != nil {
		s.logger.Error("failed to count search results", "err", err, "query", query.Query, "request_id", middleware.GetRequestID(ctx))
		// Don't fail the request, just log the error
		total = int64(len(results))
	}

	// Filter by minimum score if specified
	if query.MinScore > 0 {
		filtered := make([]SearchResult, 0, len(results))
		for _, r := range results {
			if r.Score >= query.MinScore {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	duration := time.Since(startTime).Milliseconds()

	s.logger.Info("search completed",
		"query", query.Query,
		"results", len(results),
		"total", total,
		"duration_ms", duration,
		"request_id", middleware.GetRequestID(ctx),
	)

	return SearchResponse{
		Results:  results,
		Total:    int(total),
		Query:    query.Query,
		Duration: duration,
	}, nil
}

// convertFTSResults converts sqlcext FTS results to SearchResult
func (s *SearchService) convertFTSResults(ftsResults []sqlcext.FTSSearchResult) []SearchResult {
	results := make([]SearchResult, 0, len(ftsResults))
	for _, fts := range ftsResults {
		results = append(results, SearchResult{
			ID:        fts.ID,
			Title:     fts.Title,
			Snippet:   fts.Body, // Body or snippet depending on search type
			Score:     fts.Score,
			CreatedAt: fts.CreatedAt,
		})
	}
	return results
}
