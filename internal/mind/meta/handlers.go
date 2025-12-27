// Metadata is managed via note CUD operations (CreateNote, ReplaceNote, DeleteNote)
package meta

import (
	"context"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	apierrors "github.com/nkapatos/mindweaver/shared/errors"
)

type NoteMetaHandler struct {
	mindv3connect.UnimplementedNoteMetaServiceHandler
	service *NoteMetaService
}

func NewNoteMetaHandler(service *NoteMetaService) *NoteMetaHandler {
	return &NoteMetaHandler{service: service}
}

func (h *NoteMetaHandler) ListMeta(
	ctx context.Context,
	req *connect.Request[mindv3.ListMetaRequest],
) (*connect.Response[mindv3.ListMetaResponse], error) {
	metaItems, err := h.service.GetNoteMetaByNoteID(ctx, req.Msg.NoteId)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve note metadata", err)
	}

	protoItems := StoreNoteMetasToProto(metaItems)

	response := &mindv3.ListMetaResponse{
		Items: protoItems,
		Total: int32(len(protoItems)),
	}

	return connect.NewResponse(response), nil
}
