package meta

import (
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StoreNoteMetaToProto(meta store.NoteMetum) *mindv3.NoteMeta {
	result := &mindv3.NoteMeta{
		Key:   meta.Key,
		Value: meta.Value.String,
	}

	if meta.CreatedAt.Valid {
		result.CreateTime = timestamppb.New(meta.CreatedAt.Time)
	}
	if meta.UpdatedAt.Valid {
		result.UpdateTime = timestamppb.New(meta.UpdatedAt.Time)
	}

	return result
}

func StoreNoteMetasToProto(metas []store.NoteMetum) []*mindv3.NoteMeta {
	result := make([]*mindv3.NoteMeta, 0, len(metas))
	for _, meta := range metas {
		result = append(result, StoreNoteMetaToProto(meta))
	}
	return result
}
