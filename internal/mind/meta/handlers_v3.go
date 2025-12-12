// Metadata is managed via note CUD operations (CreateNote, ReplaceNote, DeleteNote)
package meta

import (
	"context"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/internal/mind/gen/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/v3/mindv3connect"
)

type NoteMetaHandlerV3 struct {
	mindv3connect.UnimplementedNoteMetaServiceHandler
	service *NoteMetaService
}

func NewNoteMetaHandlerV3(service *NoteMetaService) *NoteMetaHandlerV3 {
	return &NoteMetaHandlerV3{service: service}
}

func (h *NoteMetaHandlerV3) ListMeta(
	ctx context.Context,
	req *connect.Request[mindv3.ListMetaRequest],
) (*connect.Response[mindv3.ListMetaResponse], error) {
	metaItems, err := h.service.GetNoteMetaByNoteID(ctx, req.Msg.NoteId)
	if err != nil {
		return nil, newInternalError("failed to retrieve note metadata", err)
	}

	protoItems := StoreNoteMetasToProto(metaItems)

	response := &mindv3.ListMetaResponse{
		Items: protoItems,
		Total: int32(len(protoItems)),
	}

	return connect.NewResponse(response), nil
}
