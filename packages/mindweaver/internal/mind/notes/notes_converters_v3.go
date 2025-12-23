package notes

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DefaultCollectionID is the root collection ID used when none is specified.
const DefaultCollectionID = int64(1)

// StoreNoteToProto converts a store.Note to the proto Note message.
func StoreNoteToProto(note store.Note) *mindv3.Note {
	var body *string
	if note.Body.Valid {
		fullBody := note.Body.String
		body = &fullBody
	}

	name := fmt.Sprintf("notes/%d", note.ID)
	etag := utils.ComputeHashedETag(note.Version)

	return &mindv3.Note{
		Id:           note.ID,
		Uuid:         note.Uuid.String(),
		Name:         name,
		Title:        note.Title,
		Body:         body,
		Description:  utils.FromNullString(note.Description),
		NoteTypeId:   utils.FromNullInt64(note.NoteTypeID),
		CollectionId: note.CollectionID,
		IsTemplate:   utils.FromNullBool(note.IsTemplate),
		Etag:         etag,
		CreateTime:   timestamppb.New(note.CreatedAt.Time),
		UpdateTime:   timestamppb.New(note.UpdatedAt.Time),
	}
}

// StoreNotesToProto converts a slice of store.Note to proto Note messages.
func StoreNotesToProto(notes []store.Note) []*mindv3.Note {
	result := make([]*mindv3.Note, len(notes))
	for i, note := range notes {
		result[i] = StoreNoteToProto(note)
	}
	return result
}

// ProtoCreateNoteToStore converts a CreateNoteRequest to store params.
// Generates a new UUID for the note. Defaults collectionID to DefaultCollectionID if not specified.
func ProtoCreateNoteToStore(req *mindv3.CreateNoteRequest) store.CreateNoteParams {
	collectionID := DefaultCollectionID
	if req.CollectionId != nil {
		collectionID = *req.CollectionId
	}

	return store.CreateNoteParams{
		Uuid:         uuid.New(),
		Title:        req.Title,
		Body:         utils.NullStringFrom(req.GetBody(), true),
		Description:  utils.ToNullString(req.Description),
		NoteTypeID:   utils.ToNullInt64(req.NoteTypeId),
		IsTemplate:   utils.ToNullBool(req.IsTemplate),
		CollectionID: collectionID,
	}
}

// ProtoReplaceNoteToStore converts a ReplaceNoteRequest to store params.
// Preserves UUID and version from current note for optimistic locking.
// Defaults collectionID to DefaultCollectionID if not specified.
func ProtoReplaceNoteToStore(req *mindv3.ReplaceNoteRequest, current store.Note) store.UpdateNoteByIDParams {
	collectionID := DefaultCollectionID
	if req.CollectionId != nil {
		collectionID = *req.CollectionId
	}

	return store.UpdateNoteByIDParams{
		ID:           req.Id,
		Uuid:         current.Uuid,
		Title:        req.Title,
		Body:         utils.NullStringFrom(req.GetBody(), true),
		Description:  utils.ToNullString(req.Description),
		NoteTypeID:   utils.ToNullInt64(req.NoteTypeId),
		IsTemplate:   utils.ToNullBool(req.IsTemplate),
		CollectionID: collectionID,
		Version:      current.Version,
	}
}

// ProtoNewNoteToParams extracts collection_id and template_id from NewNoteRequest.
// Returns defaulted values: collection_id defaults to 1, template_id defaults to 1.
func ProtoNewNoteToParams(req *mindv3.NewNoteRequest) (collectionID int64, templateID int64) {
	collectionID = DefaultCollectionID
	if req.CollectionId != nil {
		collectionID = *req.CollectionId
	}

	templateID = DefaultCollectionID // Template ID 1 is system default empty template
	if req.TemplateId != nil {
		templateID = *req.TemplateId
	}

	return collectionID, templateID
}

// ApplyFieldMask applies field masking to a Note proto message.
// If fieldMask is empty, returns the note unchanged (all fields).
// Otherwise, returns a new Note with only the requested fields populated.
// Field names are comma-separated (e.g., "id,title,collectionId").
func ApplyFieldMask(note *mindv3.Note, fieldMask string) *mindv3.Note {
	if fieldMask == "" {
		return note
	}

	// Parse field mask into set
	fields := make(map[string]bool)
	for _, field := range strings.Split(fieldMask, ",") {
		fields[strings.TrimSpace(field)] = true
	}

	// Create new note with only requested fields
	masked := &mindv3.Note{}

	if fields["id"] {
		masked.Id = note.Id
	}
	if fields["uuid"] {
		masked.Uuid = note.Uuid
	}
	if fields["name"] {
		masked.Name = note.Name
	}
	if fields["title"] {
		masked.Title = note.Title
	}
	if fields["body"] {
		masked.Body = note.Body
	}
	if fields["description"] {
		masked.Description = note.Description
	}
	if fields["noteTypeId"] || fields["note_type_id"] {
		masked.NoteTypeId = note.NoteTypeId
	}
	if fields["collectionId"] || fields["collection_id"] {
		masked.CollectionId = note.CollectionId
	}
	if fields["isTemplate"] || fields["is_template"] {
		masked.IsTemplate = note.IsTemplate
	}
	if fields["etag"] {
		masked.Etag = note.Etag
	}
	if fields["createTime"] || fields["create_time"] {
		masked.CreateTime = note.CreateTime
	}
	if fields["updateTime"] || fields["update_time"] {
		masked.UpdateTime = note.UpdateTime
	}
	if fields["metadata"] {
		masked.Metadata = note.Metadata
	}

	return masked
}

// ApplyFieldMaskToNotes applies field masking to a slice of Note proto messages.
func ApplyFieldMaskToNotes(notes []*mindv3.Note, fieldMask string) []*mindv3.Note {
	if fieldMask == "" {
		return notes
	}

	masked := make([]*mindv3.Note, len(notes))
	for i, note := range notes {
		masked[i] = ApplyFieldMask(note, fieldMask)
	}
	return masked
}
