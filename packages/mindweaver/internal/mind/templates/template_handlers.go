package templates

import (
	"context"
	"errors"
	"strconv"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TemplatesHandler struct {
	mindv3connect.UnimplementedTemplatesServiceHandler
	service *TemplatesService
}

func NewTemplatesHandler(service *TemplatesService) *TemplatesHandler {
	return &TemplatesHandler{service: service}
}

func (h *TemplatesHandler) CreateTemplate(
	ctx context.Context,
	req *connect.Request[mindv3.CreateTemplateRequest],
) (*connect.Response[mindv3.Template], error) {
	params := ProtoCreateTemplateToStore(req.Msg)

	templateID, err := h.service.CreateTemplate(ctx, params)
	if err != nil {
		if errors.Is(err, ErrTemplateAlreadyExists) {
			return nil, newAlreadyExistsError("templates", "display_name", req.Msg.DisplayName)
		}
		return nil, newInternalError("failed to create template", err)
	}

	template, err := h.service.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, newInternalError("failed to retrieve created template", err)
	}

	return connect.NewResponse(StoreTemplateToProto(template)), nil
}

func (h *TemplatesHandler) ListTemplates(
	ctx context.Context,
	req *connect.Request[mindv3.ListTemplatesRequest],
) (*connect.Response[mindv3.ListTemplatesResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.GetPageSize(), req.Msg.GetPageToken())
	params := pageReq.ToParams()

	templates, err := h.service.ListTemplatesPaginated(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, newInternalError("failed to list templates", err)
	}

	var totalCount int64
	if pageReq.IsFirstPage() {
		totalCount, _ = h.service.CountTemplates(ctx)
	}

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(templates), totalCount)
	templates = pagination.TrimResults(templates, pageReq.PageSize)

	resp := &mindv3.ListTemplatesResponse{
		Templates:     StoreTemplatesToProto(templates),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *TemplatesHandler) GetTemplate(
	ctx context.Context,
	req *connect.Request[mindv3.GetTemplateRequest],
) (*connect.Response[mindv3.Template], error) {
	template, err := h.service.GetTemplateByID(ctx, req.Msg.GetId())
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, newNotFoundError("templates", strconv.FormatInt(req.Msg.GetId(), 10))
		}
		return nil, newInternalError("failed to get template", err)
	}

	return connect.NewResponse(StoreTemplateToProto(template)), nil
}

func (h *TemplatesHandler) UpdateTemplate(
	ctx context.Context,
	req *connect.Request[mindv3.UpdateTemplateRequest],
) (*connect.Response[mindv3.Template], error) {
	params := ProtoUpdateTemplateToStore(req.Msg)

	err := h.service.UpdateTemplate(ctx, params)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, newNotFoundError("templates", strconv.FormatInt(req.Msg.GetId(), 10))
		}
		if errors.Is(err, ErrTemplateAlreadyExists) {
			return nil, newAlreadyExistsError("templates", "display_name", req.Msg.DisplayName)
		}
		return nil, newInternalError("failed to update template", err)
	}

	template, err := h.service.GetTemplateByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, newInternalError("failed to retrieve updated template", err)
	}

	return connect.NewResponse(StoreTemplateToProto(template)), nil
}

func (h *TemplatesHandler) DeleteTemplate(
	ctx context.Context,
	req *connect.Request[mindv3.DeleteTemplateRequest],
) (*connect.Response[emptypb.Empty], error) {
	err := h.service.DeleteTemplate(ctx, req.Msg.GetId())
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, newNotFoundError("templates", strconv.FormatInt(req.Msg.GetId(), 10))
		}
		return nil, newInternalError("failed to delete template", err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
