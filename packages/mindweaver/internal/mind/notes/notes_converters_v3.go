package notes

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/store"
	mindv3 "github.com/nkapatos/mindweaver/pkg/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/pkg/utils"
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
		Body:         sql.NullString{String: req.GetBody(), Valid: req.Body != nil && *req.Body != ""},
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
		Body:         sql.NullString{String: req.GetBody(), Valid: req.Body != nil && *req.Body != ""},
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
