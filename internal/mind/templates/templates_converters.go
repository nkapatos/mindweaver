package templates

import (
	"fmt"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/shared/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StoreTemplateToProto(t store.Template) *mindv3.Template {
	proto := &mindv3.Template{
		Name:          fmt.Sprintf("templates/%d", t.ID),
		Id:            t.ID,
		DisplayName:   t.Name,
		StarterNoteId: t.StarterNoteID,
		Description:   utils.FromNullString(t.Description),
		NoteTypeId:    utils.FromNullInt64(t.NoteTypeID),
	}

	if t.CreatedAt.Valid {
		proto.CreateTime = timestamppb.New(t.CreatedAt.Time)
	}
	if t.UpdatedAt.Valid {
		proto.UpdateTime = timestamppb.New(t.UpdatedAt.Time)
	}

	return proto
}

func StoreTemplatesToProto(templates []store.Template) []*mindv3.Template {
	result := make([]*mindv3.Template, len(templates))
	for i, t := range templates {
		result[i] = StoreTemplateToProto(t)
	}
	return result
}

func ProtoCreateTemplateToStore(req *mindv3.CreateTemplateRequest) store.CreateTemplateParams {
	return store.CreateTemplateParams{
		Name:          req.DisplayName,
		Description:   utils.ToNullString(req.Description),
		StarterNoteID: req.StarterNoteId,
		NoteTypeID:    utils.ToNullInt64(req.NoteTypeId),
	}
}

func ProtoUpdateTemplateToStore(req *mindv3.UpdateTemplateRequest) store.UpdateTemplateByIDParams {
	return store.UpdateTemplateByIDParams{
		ID:            req.Id,
		Name:          req.DisplayName,
		Description:   utils.ToNullString(req.Description),
		StarterNoteID: req.StarterNoteId,
		NoteTypeID:    utils.ToNullInt64(req.NoteTypeId),
	}
}
