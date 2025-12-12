package notetypes

import "encoding/json"

// CreateNoteReq is the API request DTO for creating a note.
type CreateNoteReq struct {
	Title        string           `json:"title"`          // Required
	Body         string           `json:"body"`           // Required
	Description  *string          `json:"description"`    // Optional
	NoteTypeID   *int64           `json:"note_type_id"`   // Optional FK to note_types
	IsTemplate   *bool            `json:"is_template"`    // Optional, defaults to false in DB
	CollectionID *int64           `json:"collection_id"`  // Optional FK to collections, defaults to 1 (root) in DB
	Meta         *json.RawMessage `json:"meta,omitempty"` // Optional metadata (saved separately)
}

// UpdateNoteReq is the API request DTO for updating a note.
// Optimistic locking is handled via If-Match ETag header, not in request body.
type UpdateNoteReq struct {
	Title        *string          `json:"title"`          // Optional
	Body         *string          `json:"body"`           // Optional
	Description  *string          `json:"description"`    // Optional
	NoteTypeID   *int64           `json:"note_type_id"`   // Optional FK to note_types
	IsTemplate   *bool            `json:"is_template"`    // Optional
	CollectionID *int64           `json:"collection_id"`  // Optional FK to collections (for moving notes)
	Meta         *json.RawMessage `json:"meta,omitempty"` // Optional metadata (updated separately)
}

// NoteRes is the standard note response for all operations (GET, CREATE, UPDATE, LIST).
// Per AIP-133/134, CREATE and UPDATE return the full resource, not minimal metadata.
// Maps directly to store.Note with API-friendly types.
type NoteRes struct {
	ID           int64            `json:"id"`
	UUID         string           `json:"uuid"`
	Title        string           `json:"title"`
	Body         *string          `json:"body,omitempty"`         // Nullable in DB
	Description  *string          `json:"description,omitempty"`  // Nullable in DB
	CreatedAt    string           `json:"created_at"`             // RFC3339 format
	UpdatedAt    string           `json:"updated_at"`             // RFC3339 format
	NoteTypeID   *int64           `json:"note_type_id,omitempty"` // Nullable FK
	IsTemplate   *bool            `json:"is_template,omitempty"`  // Nullable boolean
	CollectionID int64            `json:"collection_id"`          // FK to collections
	Version      int64            `json:"version"`                // Version for optimistic locking
	Meta         *json.RawMessage `json:"meta,omitempty"`         // Optional embedded meta
}
