package collections

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/internal/mind/events"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	sharederrors "github.com/nkapatos/mindweaver/shared/errors"
	"github.com/nkapatos/mindweaver/shared/middleware"
	"github.com/nkapatos/mindweaver/shared/sqlcext"
	"github.com/nkapatos/mindweaver/shared/utils"
)

type CollectionsService struct {
	store      store.Querier
	cteQuerier *sqlcext.CTEQuerier
	logger     *slog.Logger
	eventHub   events.Hub
}

func NewCollectionsService(db sqlcext.DB, store store.Querier, logger *slog.Logger, serviceName string) *CollectionsService {
	return &CollectionsService{
		store:      store,
		cteQuerier: sqlcext.NewCTEQuerier(db),
		logger:     logger.With("service", serviceName),
	}
}

// SetEventHub sets the event hub for SSE notifications.
func (s *CollectionsService) SetEventHub(hub events.Hub) {
	s.eventHub = hub
	s.logger.Info("event hub enabled for collections service")
}

// ListCollections returns all collections.
func (s *CollectionsService) ListCollections(ctx context.Context) ([]store.Collection, error) {
	collections, err := s.store.ListCollections(ctx)
	if err != nil {
		s.logger.Error("failed to list collections", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return collections, err
}

// ListCollectionsPaginated returns collections with pagination.
func (s *CollectionsService) ListCollectionsPaginated(ctx context.Context, limit, offset int32) ([]store.Collection, error) {
	collections, err := s.store.ListCollectionsPaginated(ctx, store.ListCollectionsPaginatedParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list collections paginated", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return collections, err
}

// CountCollections returns the total number of collections.
func (s *CollectionsService) CountCollections(ctx context.Context) (int64, error) {
	count, err := s.store.CountCollections(ctx)
	if err != nil {
		s.logger.Error("failed to count collections", "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// GetCollectionByID returns a collection by ID.
func (s *CollectionsService) GetCollectionByID(ctx context.Context, id int64) (store.Collection, error) {
	collection, err := s.store.GetCollectionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Collection{}, ErrCollectionNotFound
		}
		s.logger.Error("failed to get collection by id", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Collection{}, err
	}
	return collection, nil
}

// GetCollectionByPath returns a collection by its path.
func (s *CollectionsService) GetCollectionByPath(ctx context.Context, path string) (store.Collection, error) {
	collection, err := s.store.GetCollectionByPath(ctx, path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Collection{}, ErrCollectionNotFound
		}
		s.logger.Error("failed to get collection by path", "path", path, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Collection{}, err
	}
	return collection, nil
}

func (s *CollectionsService) CreateCollection(ctx context.Context, params store.CreateCollectionParams) (store.Collection, error) {
	id, err := s.store.CreateCollection(ctx, params)
	if err != nil {
		if sharederrors.IsUniqueConstraintError(err) {
			return store.Collection{}, ErrCollectionAlreadyExists
		}
		if sharederrors.IsForeignKeyConstraintError(err) {
			return store.Collection{}, ErrInvalidParentCollection
		}
		s.logger.Error("failed to create collection", "params", params, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Collection{}, err
	}

	collection, err := s.store.GetCollectionByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to fetch created collection", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return store.Collection{}, err
	}

	s.logger.Info("collection created", "id", id, "path", params.Path, "request_id", middleware.GetRequestID(ctx))

	if s.eventHub != nil {
		s.eventHub.Publish(ctx, mindv3.EventDomain_EVENT_DOMAIN_COLLECTION, mindv3.EventType_EVENT_TYPE_CREATED, id)
	}

	return collection, nil
}

// UpdateCollection updates an existing collection.
func (s *CollectionsService) UpdateCollection(ctx context.Context, params store.UpdateCollectionParams) error {
	err := s.store.UpdateCollection(ctx, params)
	if err != nil {
		if sharederrors.IsUniqueConstraintError(err) {
			return ErrCollectionAlreadyExists
		}
		if sharederrors.IsForeignKeyConstraintError(err) {
			return ErrInvalidParentCollection
		}
		s.logger.Error("failed to update collection", "id", params.ID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("collection updated", "id", params.ID, "request_id", middleware.GetRequestID(ctx))

	if s.eventHub != nil {
		s.eventHub.Publish(ctx, mindv3.EventDomain_EVENT_DOMAIN_COLLECTION, mindv3.EventType_EVENT_TYPE_UPDATED, params.ID)
	}

	return nil
}

func (s *CollectionsService) GetCollectionTree(ctx context.Context, maxDepth int) ([]sqlcext.CollectionTreeRow, error) {
	tree, err := s.cteQuerier.GetCollectionTree(ctx, maxDepth)
	if err != nil {
		s.logger.Error("failed to get collection tree", "max_depth", maxDepth, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	s.logger.Debug("collection tree retrieved", "count", len(tree), "max_depth", maxDepth, "request_id", middleware.GetRequestID(ctx))
	return tree, nil
}

func (s *CollectionsService) GetCollectionSubtree(ctx context.Context, collectionID int64, maxDepth int) ([]sqlcext.CollectionTreeRow, error) {
	subtree, err := s.cteQuerier.GetCollectionSubtree(ctx, collectionID, maxDepth)
	if err != nil {
		s.logger.Error("failed to get collection subtree", "collection_id", collectionID, "max_depth", maxDepth, "err", err, "request_id", middleware.GetRequestID(ctx))
		return nil, err
	}
	s.logger.Debug("collection subtree retrieved", "count", len(subtree), "collection_id", collectionID, "max_depth", maxDepth, "request_id", middleware.GetRequestID(ctx))
	return subtree, nil
}

// DeleteCollection deletes a collection by ID.
// Note: This may fail if there are notes in the collection (FK constraint).
func (s *CollectionsService) DeleteCollection(ctx context.Context, id int64) error {
	err := s.store.DeleteCollection(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete collection", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}
	s.logger.Info("collection deleted", "id", id, "request_id", middleware.GetRequestID(ctx))

	if s.eventHub != nil {
		s.eventHub.Publish(ctx, mindv3.EventDomain_EVENT_DOMAIN_COLLECTION, mindv3.EventType_EVENT_TYPE_DELETED, id)
	}

	return nil
}

// GetCollectionAncestors returns all ancestors of a collection (parent, grandparent, etc).
func (s *CollectionsService) GetCollectionAncestors(ctx context.Context, id int64) ([]store.GetCollectionAncestorsRow, error) {
	ancestors, err := s.store.GetCollectionAncestors(ctx, id)
	if err != nil {
		s.logger.Error("failed to get collection ancestors", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return ancestors, err
}

// GetCollectionDescendants returns all descendants of a collection (children, grandchildren, etc).
func (s *CollectionsService) GetCollectionDescendants(ctx context.Context, id int64) ([]store.GetCollectionDescendantsRow, error) {
	descendants, err := s.store.GetCollectionDescendants(ctx, id)
	if err != nil {
		s.logger.Error("failed to get collection descendants", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return descendants, err
}

// ListCollectionsByParent returns direct children of a collection.
func (s *CollectionsService) ListCollectionsByParent(ctx context.Context, parentID sql.NullInt64) ([]store.Collection, error) {
	collections, err := s.store.ListCollectionsByParent(ctx, parentID)
	if err != nil {
		s.logger.Error("failed to list collections by parent", "parent_id", parentID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return collections, err
}

// ListCollectionsByParentPaginated returns direct children of a collection with pagination.
func (s *CollectionsService) ListCollectionsByParentPaginated(ctx context.Context, parentID sql.NullInt64, limit, offset int32) ([]store.Collection, error) {
	collections, err := s.store.ListCollectionsByParentPaginated(ctx, store.ListCollectionsByParentPaginatedParams{
		ParentID: parentID,
		Limit:    int64(limit),
		Offset:   int64(offset),
	})
	if err != nil {
		s.logger.Error("failed to list collections by parent paginated", "parent_id", parentID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return collections, err
}

// CountCollectionsByParent returns the total number of direct children of a collection.
func (s *CollectionsService) CountCollectionsByParent(ctx context.Context, parentID sql.NullInt64) (int64, error) {
	count, err := s.store.CountCollectionsByParent(ctx, parentID)
	if err != nil {
		s.logger.Error("failed to count collections by parent", "parent_id", parentID, "err", err, "request_id", middleware.GetRequestID(ctx))
	}
	return count, err
}

// CountNotesInCollection returns the number of notes in a collection.
func (s *CollectionsService) CountNotesInCollection(ctx context.Context, id int64) (int64, error) {
	count, err := s.store.CountNotesInCollection(ctx, id)
	if err != nil {
		s.logger.Error("failed to count notes in collection", "id", id, "err", err, "request_id", middleware.GetRequestID(ctx))
		return 0, err
	}
	return count, nil
}

// ============================================================================
// Path Management
// ============================================================================

// GenerateCollectionPath generates a path for a collection based on its name and parent.
// For root collections (parent_id = NULL), the path is just the slug of the name.
// For child collections, the path is parent_path/slug.
func (s *CollectionsService) GenerateCollectionPath(ctx context.Context, name string, parentID interface{}) (string, error) {
	slug := utils.GenerateSlug(name)

	// If no parent, this is a root collection
	if parentID == nil {
		return slug, nil
	}

	// Get parent to build path
	var parent store.Collection
	var err error

	switch v := parentID.(type) {
	case int64:
		parent, err = s.store.GetCollectionByID(ctx, v)
	case sql.NullInt64:
		if !v.Valid {
			return slug, nil
		}
		parent, err = s.store.GetCollectionByID(ctx, v.Int64)
	default:
		return "", fmt.Errorf("invalid parent_id type: %T", parentID)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidParentCollection
		}
		s.logger.Error("failed to get parent collection", "parent_id", parentID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return "", err
	}

	// Build path: parent_path/slug
	return fmt.Sprintf("%s/%s", parent.Path, slug), nil
}

// UpdateDescendantPaths recursively updates paths for all descendants when a collection is moved or renamed.
// This is called after updating a collection's path to keep the tree consistent.
func (s *CollectionsService) UpdateDescendantPaths(ctx context.Context, collectionID int64, newPath string) error {
	// Get all direct children
	children, err := s.store.GetCollectionChildren(ctx, utils.NullInt64(collectionID))
	if err != nil {
		s.logger.Error("failed to get children for path update", "collection_id", collectionID, "err", err, "request_id", middleware.GetRequestID(ctx))
		return err
	}

	// Update each child's path
	for _, child := range children {
		// Generate new path for child: parent_path/child_slug
		childSlug := utils.GenerateSlug(child.Name)
		newChildPath := fmt.Sprintf("%s/%s", newPath, childSlug)

		// Update child's path
		err := s.store.UpdateCollection(ctx, store.UpdateCollectionParams{
			ID:          child.ID,
			Name:        child.Name,
			ParentID:    child.ParentID,
			Path:        newChildPath,
			Description: child.Description,
			Position:    child.Position,
		})
		if err != nil {
			s.logger.Error("failed to update child path", "child_id", child.ID, "new_path", newChildPath, "err", err, "request_id", middleware.GetRequestID(ctx))
			return err
		}

		// Recursively update descendants
		if err := s.UpdateDescendantPaths(ctx, child.ID, newChildPath); err != nil {
			return err
		}
	}

	return nil
}
