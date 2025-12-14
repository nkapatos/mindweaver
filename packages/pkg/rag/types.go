// Package rag provides shared types and interfaces for Retrieval-Augmented Generation.
// Both Mind and Brain implement these interfaces with their respective data sources.
package rag

import "time"

// SearchResult represents a single search result from FTS or semantic search.
type SearchResult struct {
	ID       int64     `json:"id"`        // Note or observation ID
	Title    string    `json:"title"`     // Document title
	Snippet  string    `json:"snippet"`   // Matched text snippet (excerpt)
	Score    float64   `json:"score"`     // Relevance score (FTS rank or similarity)
	Source   string    `json:"source"`    // "mind" or "brain"
	NoteType string    `json:"note_type"` // Type of note/observation
	Created  time.Time `json:"created"`   // Creation timestamp
}

// SearchQuery contains parameters for a search request.
type SearchQuery struct {
	Query       string   `json:"query"`        // Search query text
	Limit       int      `json:"limit"`        // Max results to return
	Offset      int      `json:"offset"`       // Pagination offset
	Source      string   `json:"source"`       // "mind", "brain", or "both"
	NoteTypes   []string `json:"note_types"`   // Filter by note types (optional)
	MinScore    float64  `json:"min_score"`    // Minimum relevance score (optional)
	IncludeBody bool     `json:"include_body"` // Include full body in results
}

// SearchResponse contains search results and metadata.
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`       // Total matching documents
	Query    string         `json:"query"`       // Echo back the query
	Source   string         `json:"source"`      // Which source(s) were searched
	Duration int64          `json:"duration_ms"` // Search duration in milliseconds
}

// Searcher is the interface that both Mind and Brain RAG services implement.
// This allows the chat service to work with either data source transparently.
type Searcher interface {
	// Search performs full-text search on the data source
	Search(query SearchQuery) (SearchResponse, error)

	// GetByID retrieves a specific document by ID (for context expansion)
	GetByID(id int64) (*SearchResult, error)

	// GetRelated finds documents related to the given document ID
	// (e.g., via links, tags, or similar themes)
	GetRelated(id int64, limit int) ([]SearchResult, error)
}
