package notetypes

import (
	"fmt"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/shared/utils"
)

func StoreNoteTypeToProto(nt store.NoteType) *mindv3.NoteType {
	return &mindv3.NoteType{
		Name:        fmt.Sprintf("note_types/%d", nt.ID),
		Id:          nt.ID,
		Type:        nt.Type,
		DisplayName: nt.Name,
		IsSystem:    nt.IsSystem,
		Description: utils.FromNullString(nt.Description),
		Icon:        utils.FromNullString(nt.Icon),
		Color:       utils.FromNullString(nt.Color),
	}
}

func StoreNoteTypesToProto(noteTypes []store.NoteType) []*mindv3.NoteType {
	result := make([]*mindv3.NoteType, len(noteTypes))
	for i, nt := range noteTypes {
		result[i] = StoreNoteTypeToProto(nt)
	}
	return result
}

func ProtoCreateNoteTypeToStore(req *mindv3.CreateNoteTypeRequest) store.CreateNoteTypeParams {
	return store.CreateNoteTypeParams{
		Type:        req.Type,
		Name:        req.DisplayName,
		Description: utils.ToNullString(req.Description),
		Icon:        utils.ToNullString(req.Icon),
		Color:       utils.ToNullString(req.Color),
	}
}

// ProtoUpdateNoteTypeToStore converts an UpdateNoteTypeRequest to store params.
// Note: IsSystem is intentionally set to false here as a safe default.
// The service layer fetches the current record and preserves the actual IsSystem value,
// and also prevents updates to system note types entirely.
func ProtoUpdateNoteTypeToStore(req *mindv3.UpdateNoteTypeRequest) store.UpdateNoteTypeByIDParams {
	return store.UpdateNoteTypeByIDParams{
		ID:          req.Id,
		Type:        req.Type,
		Name:        req.DisplayName,
		Description: utils.ToNullString(req.Description),
		Icon:        utils.ToNullString(req.Icon),
		Color:       utils.ToNullString(req.Color),
		IsSystem:    false, // Service layer will preserve actual value
	}
}
