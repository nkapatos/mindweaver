package collections

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"connectrpc.com/connect"
	"github.com/nkapatos/mindweaver/pkg/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/store"
	"github.com/nkapatos/mindweaver/pkg/dberrors"
	"github.com/nkapatos/mindweaver/pkg/pagination"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TODO: V1 to V3 Migration - Missing Endpoints
// The following V1 endpoints are not yet implemented in V3:
// - GetCollectionByPath: Get collection by file path (alternative to ID-based lookup)
// - GetCollectionAncestors: Get parent chain for breadcrumb navigation
// - GetCollectionDescendants: Get all descendants recursively
// - GetCollectionSubtree: Get subtree starting at a specific collection
// - ListCollectionNotes: List notes in a collection (may belong in Notes service)
// - CountNotesInCollection: Count notes in a collection (may belong in Notes service)
// Consider adding these as needed based on client requirements.

type CollectionsHandlerV3 struct {
	service *CollectionsService
}

func NewCollectionsHandlerV3(service *CollectionsService) *CollectionsHandlerV3 {
	return &CollectionsHandlerV3{
		service: service,
	}
}

func (h *CollectionsHandlerV3) CreateCollection(
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
			return nil, connect.NewError(connect.CodeNotFound, ErrInvalidParentCollection)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	params := ProtoCreateCollectionToStore(req.Msg, path)

	collection, err := h.service.CreateCollection(ctx, params)
	if err != nil {
		if dberrors.IsUniqueConstraintError(err) {
			return nil, newAlreadyExistsError("collection", "path", path)
		}
		if dberrors.IsForeignKeyConstraintError(err) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrInvalidParentCollection)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(StoreCollectionToProto(collection)), nil
}

func (h *CollectionsHandlerV3) GetCollection(
	ctx context.Context,
	req *connect.Request[mindv3.GetCollectionRequest],
) (*connect.Response[mindv3.Collection], error) {
	collection, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("collection", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(StoreCollectionToProto(collection)), nil
}

func (h *CollectionsHandlerV3) UpdateCollection(
	ctx context.Context,
	req *connect.Request[mindv3.UpdateCollectionRequest],
) (*connect.Response[mindv3.Collection], error) {
	current, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("collection", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if current.IsSystem {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrCollectionIsSystem)
	}

	// Regenerate path if name or parent changed
	var parentID interface{}
	if req.Msg.ParentId != nil {
		parentID = *req.Msg.ParentId
	}

	path, err := h.service.GenerateCollectionPath(ctx, req.Msg.DisplayName, parentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, ErrInvalidParentCollection)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	params := ProtoUpdateCollectionToStore(req.Msg, path, current.IsSystem)

	err = h.service.UpdateCollection(ctx, params)
	if err != nil {
		if dberrors.IsUniqueConstraintError(err) {
			return nil, newAlreadyExistsError("collection", "path", path)
		}
		if dberrors.IsForeignKeyConstraintError(err) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrInvalidParentCollection)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// If path changed, update descendant paths
	if path != current.Path {
		err = h.service.UpdateDescendantPaths(ctx, req.Msg.Id, path)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	updated, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(StoreCollectionToProto(updated)), nil
}

func (h *CollectionsHandlerV3) DeleteCollection(
	ctx context.Context,
	req *connect.Request[mindv3.DeleteCollectionRequest],
) (*connect.Response[emptypb.Empty], error) {
	collection, err := h.service.GetCollectionByID(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("collection", strconv.FormatInt(req.Msg.Id, 10))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if collection.IsSystem {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrCollectionIsSystem)
	}

	err = h.service.DeleteCollection(ctx, req.Msg.Id)
	if err != nil {
		// Check for foreign key violation (collection has notes)
		if dberrors.IsForeignKeyConstraintError(err) {
			errInfo := &errdetails.ErrorInfo{
				Reason: "COLLECTION_HAS_NOTES",
				Domain: "mind.v3",
				Metadata: map[string]string{
					"collection_id": strconv.FormatInt(req.Msg.Id, 10),
				},
			}
			connErr := connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot delete collection with notes"))
			if detail, err := connect.NewErrorDetail(errInfo); err == nil {
				connErr.AddDetail(detail)
			}
			return nil, connErr
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *CollectionsHandlerV3) ListCollections(
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
		parentID := sql.NullInt64{Int64: *req.Msg.ParentId, Valid: true}
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
		return nil, connect.NewError(connect.CodeInternal, err)
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

func (h *CollectionsHandlerV3) ListCollectionChildren(
	ctx context.Context,
	req *connect.Request[mindv3.ListCollectionChildrenRequest],
) (*connect.Response[mindv3.ListCollectionsResponse], error) {
	// Parse pagination request
	pageReq := pagination.ParseRequest(req.Msg.PageSize, req.Msg.PageToken)
	params := pageReq.ToParams()

	parentID := sql.NullInt64{Int64: req.Msg.ParentId, Valid: true}
	collections, err := h.service.ListCollectionsByParentPaginated(ctx, parentID, params.Limit, params.Offset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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

func (h *CollectionsHandlerV3) GetCollectionTree(
	ctx context.Context,
	req *connect.Request[mindv3.GetCollectionTreeRequest],
) (*connect.Response[mindv3.GetCollectionTreeResponse], error) {
	root, err := h.service.GetCollectionByID(ctx, req.Msg.RootId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newNotFoundError("collection", strconv.FormatInt(req.Msg.RootId, 10))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	maxDepth := int(req.Msg.MaxDepth)
	descendants, err := h.service.GetCollectionSubtree(ctx, req.Msg.RootId, maxDepth)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &mindv3.GetCollectionTreeResponse{
		Root:        StoreCollectionToProto(root),
		Descendants: CollectionTreeRowsToProto(descendants),
	}

	return connect.NewResponse(resp), nil
}
