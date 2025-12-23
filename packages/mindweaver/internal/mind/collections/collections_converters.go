package collections

import (
	"strconv"

	mindv3 "github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/sqlcext"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StoreCollectionToProto(c store.Collection) *mindv3.Collection {
	proto := &mindv3.Collection{
		Name:        "collections/" + strconv.FormatInt(c.ID, 10),
		Id:          c.ID,
		DisplayName: c.Name,
		Path:        c.Path,
		IsSystem:    c.IsSystem,
	}

	if parentID := utils.FromInterface(c.ParentID); parentID != nil {
		proto.ParentId = parentID
	}

	if c.Description.Valid {
		proto.Description = &c.Description.String
	}

	if c.Position.Valid {
		proto.Position = &c.Position.Int64
	}

	if c.CreatedAt.Valid {
		proto.CreateTime = timestamppb.New(c.CreatedAt.Time)
	}
	if c.UpdatedAt.Valid {
		proto.UpdateTime = timestamppb.New(c.UpdatedAt.Time)
	}

	return proto
}

func StoreCollectionsToProto(collections []store.Collection) []*mindv3.Collection {
	result := make([]*mindv3.Collection, len(collections))
	for i, c := range collections {
		result[i] = StoreCollectionToProto(c)
	}
	return result
}

// Path must be generated separately by the service
func ProtoCreateCollectionToStore(req *mindv3.CreateCollectionRequest, path string) store.CreateCollectionParams {
	var parentID interface{}
	if req.ParentId != nil {
		parentID = *req.ParentId
	}

	return store.CreateCollectionParams{
		Name:        req.DisplayName,
		ParentID:    parentID,
		Path:        path,
		Description: utils.ToNullString(req.Description),
		Position:    utils.ToNullInt64(req.Position),
		IsSystem:    false,
	}
}

// Path must be generated separately by the service
func ProtoUpdateCollectionToStore(req *mindv3.UpdateCollectionRequest, path string, isSystem bool) store.UpdateCollectionParams {
	var parentID interface{}
	if req.ParentId != nil {
		parentID = *req.ParentId
	}

	return store.UpdateCollectionParams{
		ID:          req.Id,
		Name:        req.DisplayName,
		ParentID:    parentID,
		Path:        path,
		Description: utils.ToNullString(req.Description),
		Position:    utils.ToNullInt64(req.Position),
		IsSystem:    isSystem,
	}
}

// CollectionTreeRowToProto converts a sqlcext.CollectionTreeRow (from CTE queries) to protobuf Collection.
// This is used for tree/hierarchy queries where we need the collection data without timestamps.
func CollectionTreeRowToProto(row sqlcext.CollectionTreeRow) *mindv3.Collection {
	// Convert the tree row to a store.Collection first to reuse existing converter
	collection := store.Collection{
		ID:          row.ID,
		Name:        row.Name,
		ParentID:    row.ParentID,
		Path:        row.Path,
		Description: row.Description,
		Position:    row.Position,
		IsSystem:    row.IsSystem,
		// Note: CreatedAt and UpdatedAt are not available in tree queries
		// They will be omitted from the proto response
	}
	return StoreCollectionToProto(collection)
}

// CollectionTreeRowsToProto converts multiple CollectionTreeRow to protobuf Collections.
func CollectionTreeRowsToProto(rows []sqlcext.CollectionTreeRow) []*mindv3.Collection {
	result := make([]*mindv3.Collection, len(rows))
	for i, row := range rows {
		result[i] = CollectionTreeRowToProto(row)
	}
	return result
}
