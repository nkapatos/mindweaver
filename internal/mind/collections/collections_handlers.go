package collections

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"connectrpc.com/connect"
	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	apierrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/pagination"
	"github.com/nkapatos/mindweaver/shared/utils"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Note: Some V1 endpoints not yet ported to V3 - See issue #38

type CollectionsHandler struct {
	service *CollectionsService
}

func NewCollectionsHandler(service *CollectionsService) *CollectionsHandler {
	return &CollectionsHandler{
		service: service,
	}
}

func (h *CollectionsHandler) CreateCollection(
	ctx context.Context,
	req *connect.Request[mindv3.CreateCollectionRequest],
) (*connect.Response[mindv3.Collection], error) {
	var parentID interface{}
	if req.Msg.ParentId != nil {
		parentID = *req.Msg.ParentId
	}

	path, err := h.service.GenerateCollectionPath(ctx, req.Msg.DisplayName, parentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewInvalidArgumentError("parent_id", ErrInvalidParentCollection.Error())
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to generate collection path", err)
	}

	params := ProtoCreateCollectionToStore(req.Msg, path)

	collection, err := h.service.CreateCollection(ctx, params)
	if err != nil {
		if apierrors.IsUniqueConstraintError(err) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "collection", "path", path)
		}
		if apierrors.IsForeignKeyConstraintError(err) {
			return nil, apierrors.NewInvalidArgumentError("parent_id", ErrInvalidParentCollection.Error())
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to create collection", err)
	}

	return connect.NewResponse(StoreCollectionToProto(collection)), nil
}

func (h *CollectionsHandler) GetCollection(
	ctx context.Context,
	req *connect.Request[mindv3.GetCollectionRequest],
) (*connect.Response[mindv3.Collection], error) {
	collection, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "collection", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get collection", err)
	}

	return connect.NewResponse(StoreCollectionToProto(collection)), nil
}

func (h *CollectionsHandler) UpdateCollection(
	ctx context.Context,
	req *connect.Request[mindv3.UpdateCollectionRequest],
) (*connect.Response[mindv3.Collection], error) {
	current, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "collection", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get collection", err)
	}

	if current.IsSystem {
		return nil, apierrors.NewPermissionDeniedError(apierrors.MindDomain, ErrCollectionIsSystem.Error())
	}

	// Regenerate path if name or parent changed
	var parentID interface{}
	if req.Msg.ParentId != nil {
		parentID = *req.Msg.ParentId
	}

	path, err := h.service.GenerateCollectionPath(ctx, req.Msg.DisplayName, parentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewInvalidArgumentError("parent_id", ErrInvalidParentCollection.Error())
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to generate collection path", err)
	}

	params := ProtoUpdateCollectionToStore(req.Msg, path, current.IsSystem)

	err = h.service.UpdateCollection(ctx, params)
	if err != nil {
		if apierrors.IsUniqueConstraintError(err) {
			return nil, apierrors.NewAlreadyExistsError(apierrors.MindDomain, "collection", "path", path)
		}
		if apierrors.IsForeignKeyConstraintError(err) {
			return nil, apierrors.NewInvalidArgumentError("parent_id", ErrInvalidParentCollection.Error())
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to update collection", err)
	}

	// If path changed, update descendant paths
	if path != current.Path {
		err = h.service.UpdateDescendantPaths(ctx, req.Msg.Id, path)
		if err != nil {
			return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to update descendant paths", err)
		}
	}

	updated, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to retrieve updated collection", err)
	}

	return connect.NewResponse(StoreCollectionToProto(updated)), nil
}

func (h *CollectionsHandler) DeleteCollection(
	ctx context.Context,
	req *connect.Request[mindv3.DeleteCollectionRequest],
) (*connect.Response[emptypb.Empty], error) {
	collection, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "collection", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get collection", err)
	}

	if collection.IsSystem {
		return nil, apierrors.NewPermissionDeniedError(apierrors.MindDomain, ErrCollectionIsSystem.Error())
	}

	err = h.service.DeleteCollection(ctx, req.Msg.Id)
	if err != nil {
		// Check for foreign key violation (collection has notes)
		if apierrors.IsForeignKeyConstraintError(err) {
			metadata := map[string]string{
				"collection_id": strconv.FormatInt(req.Msg.Id, 10),
			}
			return nil, apierrors.NewFailedPreconditionError(apierrors.MindDomain, "COLLECTION_HAS_NOTES", metadata)
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to delete collection", err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *CollectionsHandler) ListCollections(
	ctx context.Context,
	req *connect.Request[mindv3.ListCollectionsRequest],
) (*connect.Response[mindv3.ListCollectionsResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.PageSize, req.Msg.PageToken)
	params := pageReq.ToParams()

	var collections []store.Collection
	var totalCount int64
	var err error

	if req.Msg.ParentId != nil {
		parentID := utils.NullInt64(*req.Msg.ParentId)
		collections, err = h.service.ListCollectionsByParentPaginated(ctx, parentID, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountCollectionsByParent(ctx, parentID)
		}
	} else {
		collections, err = h.service.ListCollectionsPaginated(ctx, params.Limit, params.Offset)
		if err == nil && pageReq.IsFirstPage() {
			totalCount, _ = h.service.CountCollections(ctx)
		}
	}

	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to list collections", err)
	}

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(collections), totalCount)
	collections = pagination.TrimResults(collections, pageReq.PageSize)

	resp := &mindv3.ListCollectionsResponse{
		Collections:   StoreCollectionsToProto(collections),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *CollectionsHandler) ListCollectionChildren(
	ctx context.Context,
	req *connect.Request[mindv3.ListCollectionChildrenRequest],
) (*connect.Response[mindv3.ListCollectionsResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.PageSize, req.Msg.PageToken)
	params := pageReq.ToParams()

	parentID := utils.NullInt64(req.Msg.ParentId)
	collections, err := h.service.ListCollectionsByParentPaginated(ctx, parentID, params.Limit, params.Offset)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to list collection children", err)
	}

	var totalCount int64
	if pageReq.IsFirstPage() {
		totalCount, _ = h.service.CountCollectionsByParent(ctx, parentID)
	}

	// Build pagination response
	pageResp := pageReq.BuildResponse(len(collections), totalCount)
	collections = pagination.TrimResults(collections, pageReq.PageSize)

	resp := &mindv3.ListCollectionsResponse{
		Collections:   StoreCollectionsToProto(collections),
		NextPageToken: pageResp.NextPageToken,
	}

	// Only include total_size on first page
	if pageReq.IsFirstPage() {
		totalSize := int32(pageResp.TotalCount)
		resp.TotalSize = &totalSize
	}

	return connect.NewResponse(resp), nil
}

func (h *CollectionsHandler) GetCollectionTree(
	ctx context.Context,
	req *connect.Request[mindv3.GetCollectionTreeRequest],
) (*connect.Response[mindv3.GetCollectionTreeResponse], error) {
	root, err := h.service.GetCollectionByID(ctx, req.Msg.RootId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewNotFoundError(apierrors.MindDomain, "collection", strconv.FormatInt(req.Msg.RootId, 10))
		}
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get collection", err)
	}

	maxDepth := int(req.Msg.MaxDepth)
	descendants, err := h.service.GetCollectionSubtree(ctx, req.Msg.RootId, maxDepth)
	if err != nil {
		return nil, apierrors.NewInternalError(apierrors.MindDomain, "failed to get collection subtree", err)
	}

	resp := &mindv3.GetCollectionTreeResponse{
		Root:        StoreCollectionToProto(root),
		Descendants: CollectionTreeRowsToProto(descendants),
	}

	return connect.NewResponse(resp), nil
}
