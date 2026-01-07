package tags

import (
	"fmt"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/shared/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StoreTagToProto(tag store.Tag) *mindv3.Tag {
	proto := &mindv3.Tag{
		Name:        fmt.Sprintf("tags/%d", tag.ID),
		Id:          tag.ID,
		DisplayName: tag.Name,
	}

	if tag.CreatedAt.Valid {
		proto.CreateTime = timestamppb.New(tag.CreatedAt.Time)
	}
	if tag.UpdatedAt.Valid {
		proto.UpdateTime = timestamppb.New(tag.UpdatedAt.Time)
	}

	return proto
}

func StoreTagsToProto(tags []store.Tag) []*mindv3.Tag {
	result := make([]*mindv3.Tag, len(tags))
	for i, tag := range tags {
		result[i] = StoreTagToProto(tag)
	}
	return result
}

// StoreNoteToProto converts a store.Note to proto Note for ListNotesForTag.
// Note: This is a local copy to avoid import cycle with notes package.
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
