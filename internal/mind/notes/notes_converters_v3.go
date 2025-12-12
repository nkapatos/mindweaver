package notes

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	mindv3 "github.com/nkapatos/mindweaver/internal/mind/gen/v3"
	"github.com/nkapatos/mindweaver/internal/mind/store"
	"github.com/nkapatos/mindweaver/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Default collection ID for notes when not specified.
// This is the root/default collection created during DB initialization.
const DefaultCollectionID = int64(1)

func StoreNoteToProto(note store.Note) *mindv3.Note {
	// Reconstruct full markdown with frontmatter for API response
	var body *string
	if note.Body.Valid {
		var fullBody string
		// FIXME: when the mind api stabilises just add a helper reconstruct function here. it just combines the frontmatter with the body with the necesary markers before and after teh frontmatter ---
		// if note.Frontmatter.Valid && note.Frontmatter.String != "" {
		// 	fullBody = markdown.ReconstructFullMarkdown(note.Body.String, note.Frontmatter.String)
		// } else {
		// 	fullBody = note.Body.String
		// }
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

func StoreNotesToProto(notes []store.Note) []*mindv3.Note {
	result := make([]*mindv3.Note, len(notes))
	for i, note := range notes {
		result[i] = StoreNoteToProto(note)
	}
	return result
}

// ProtoCreateNoteToStore converts a CreateNoteRequest to store params.
// Note: UUID is generated here as it's required for every new note.
// CollectionID defaults to DefaultCollectionID if not specified.
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
// Current note is needed for UUID preservation and optimistic locking via version.
// CollectionID defaults to DefaultCollectionID if not specified.
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
