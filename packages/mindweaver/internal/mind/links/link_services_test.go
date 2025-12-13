package links

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/nkapatos/mindweaver/packages/mindweaver/internal/mind/store"
	mindmigrations "github.com/nkapatos/mindweaver/packages/mindweaver/migrations/mind"
	"github.com/nkapatos/mindweaver/pkg/testdb"
	"github.com/nkapatos/mindweaver/pkg/utils"
)

// setupTestService creates a LinksService with in-memory database for testing.
func setupTestService(t *testing.T) (*LinksService, *store.Queries) {
	t.Helper()

	db := testdb.SetupTestDB(t, mindmigrations.RunMigrations)
	t.Cleanup(func() { db.Close() })

	queries := store.New(db)
	logger := testdb.NewTestLogger(t)
	service := NewLinksService(queries, logger, "links-test")

	return service, queries
}

// createTestNote creates a note for testing links.
func createTestNote(t *testing.T, queries *store.Queries, title string) int64 {
	t.Helper()

	noteID, err := queries.CreateNote(context.Background(), store.CreateNoteParams{
		Uuid:         uuid.New(),
		Title:        title,
		Body:         utils.NullString("Test body"),
		CollectionID: 1, // Default collection
	})
	require.NoError(t, err)
	return noteID
}

// ============================================================================
// Basic CRUD Tests
// ============================================================================

func TestCreateLink(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create test notes
	srcID := createTestNote(t, queries, "Source Note")
	destID := createTestNote(t, queries, "Destination Note")

	// Test creating a resolved link
	params := store.CreateLinkParams{
		SrcID:       srcID,
		DestID:      utils.NullInt64(destID),
		DisplayText: utils.NullString("Link Text"),
		IsEmbed:     utils.NullBool(false),
	}

	linkID, err := service.CreateLink(ctx, params)
	require.NoError(t, err)
	require.NotZero(t, linkID)

	// Verify link was created
	link, err := queries.GetLinkByID(ctx, linkID)
	require.NoError(t, err)
	require.Equal(t, srcID, link.SrcID)
	require.Equal(t, destID, link.DestID.Int64)
	require.True(t, link.DestID.Valid)
	require.Equal(t, "Link Text", link.DisplayText.String)
}

func TestCreateUnresolvedLink(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create source note (destination doesn't exist yet)
	srcID := createTestNote(t, queries, "Source Note")

	// Test creating an unresolved link
	params := store.CreateUnresolvedLinkParams{
		SrcID:       srcID,
		DestTitle:   utils.NullString("Nonexistent Note"),
		DisplayText: utils.NullString("Broken Link"),
		IsEmbed:     utils.NullBool(false),
	}

	linkID, err := service.CreateUnresolvedLink(ctx, params)
	require.NoError(t, err)
	require.NotZero(t, linkID)

	// Verify unresolved link was created
	link, err := queries.GetLinkByID(ctx, linkID)
	require.NoError(t, err)
	require.Equal(t, srcID, link.SrcID)
	require.False(t, link.DestID.Valid) // No destination yet
	require.Equal(t, "Nonexistent Note", link.DestTitle.String)
	require.Equal(t, int64(0), link.Resolved.Int64) // Pending
}

func TestGetLinkByID(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create test link
	srcID := createTestNote(t, queries, "Source")
	destID := createTestNote(t, queries, "Dest")

	linkID, err := queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  srcID,
		DestID: utils.NullInt64(destID),
	})
	require.NoError(t, err)

	// Test getting existing link
	link, err := service.GetLinkByID(ctx, linkID)
	require.NoError(t, err)
	require.Equal(t, linkID, link.ID)
	require.Equal(t, srcID, link.SrcID)

	// Test getting non-existent link
	_, err = service.GetLinkByID(ctx, 99999)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestListLinks(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create multiple links
	srcID := createTestNote(t, queries, "Source")
	dest1ID := createTestNote(t, queries, "Dest1")
	dest2ID := createTestNote(t, queries, "Dest2")

	_, err := queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  srcID,
		DestID: sql.NullInt64{Int64: dest1ID, Valid: true},
	})
	require.NoError(t, err)

	_, err = queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  srcID,
		DestID: sql.NullInt64{Int64: dest2ID, Valid: true},
	})
	require.NoError(t, err)

	// Test listing all links
	links, err := service.ListLinks(ctx)
	require.NoError(t, err)
	require.Len(t, links, 2)
}

func TestDeleteLinksBySrcID(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create links from same source
	srcID := createTestNote(t, queries, "Source")
	dest1ID := createTestNote(t, queries, "Dest1")
	dest2ID := createTestNote(t, queries, "Dest2")

	_, err := queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  srcID,
		DestID: utils.NullInt64(dest1ID),
	})
	require.NoError(t, err)

	_, err = queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  srcID,
		DestID: utils.NullInt64(dest2ID),
	})
	require.NoError(t, err)

	// Delete all links from source
	err = service.DeleteLinksBySrcID(ctx, srcID)
	require.NoError(t, err)

	// Verify links were deleted
	links, err := queries.ListLinksBySrcID(ctx, srcID)
	require.NoError(t, err)
	require.Empty(t, links)
}

// ============================================================================
// Query Operation Tests
// ============================================================================

func TestListLinksBySrcID(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create notes and links
	src1ID := createTestNote(t, queries, "Source1")
	src2ID := createTestNote(t, queries, "Source2")
	dest1ID := createTestNote(t, queries, "Dest1")
	dest2ID := createTestNote(t, queries, "Dest2")

	// Links from src1
	_, err := queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  src1ID,
		DestID: utils.NullInt64(dest1ID),
	})
	require.NoError(t, err)

	_, err = queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  src1ID,
		DestID: utils.NullInt64(dest2ID),
	})
	require.NoError(t, err)

	// Link from src2
	_, err = queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  src2ID,
		DestID: utils.NullInt64(dest1ID),
	})
	require.NoError(t, err)

	// Test listing links from src1
	links, err := service.ListLinksBySrcID(ctx, src1ID)
	require.NoError(t, err)
	require.Len(t, links, 2)
	for _, link := range links {
		require.Equal(t, src1ID, link.SrcID)
	}
}

func TestListLinksByDestID(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create notes and links
	src1ID := createTestNote(t, queries, "Source1")
	src2ID := createTestNote(t, queries, "Source2")
	destID := createTestNote(t, queries, "Destination")

	// Multiple sources pointing to same destination
	_, err := queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  src1ID,
		DestID: utils.NullInt64(destID),
	})
	require.NoError(t, err)

	_, err = queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:  src2ID,
		DestID: utils.NullInt64(destID),
	})
	require.NoError(t, err)

	// Test listing backlinks
	links, err := service.ListLinksByDestID(ctx, utils.NullInt64(destID))
	require.NoError(t, err)
	require.Len(t, links, 2)
	for _, link := range links {
		require.Equal(t, destID, link.DestID.Int64)
	}
}

func TestSearchLinksByDisplayText(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create links with different display text
	srcID := createTestNote(t, queries, "Source")
	dest1ID := createTestNote(t, queries, "Dest1")
	dest2ID := createTestNote(t, queries, "Dest2")

	_, err := queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:       srcID,
		DestID:      utils.NullInt64(dest1ID),
		DisplayText: utils.NullString("API Documentation"),
	})
	require.NoError(t, err)

	_, err = queries.CreateLink(ctx, store.CreateLinkParams{
		SrcID:       srcID,
		DestID:      utils.NullInt64(dest2ID),
		DisplayText: utils.NullString("User Guide"),
	})
	require.NoError(t, err)

	// Search for links with "API" in display text
	links, err := service.SearchLinksByDisplayText(ctx, "%API%")
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, "API Documentation", links[0].DisplayText.String)
}

// ============================================================================
// WikiLink Resolution Tests
// ============================================================================

func TestListUnresolvedLinks(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	srcID := createTestNote(t, queries, "Source")

	// Create unresolved links
	_, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Future Note 1"),
	})
	require.NoError(t, err)

	_, err = queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Future Note 2"),
	})
	require.NoError(t, err)

	// List unresolved links
	links, err := service.ListUnresolvedLinks(ctx, 10)
	require.NoError(t, err)
	require.Len(t, links, 2)
}

func TestFindUnresolvedLinksByDestTitle(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	srcID := createTestNote(t, queries, "Source")

	// Create unresolved links with different titles
	_, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Future Note"),
	})
	require.NoError(t, err)

	_, err = queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Another Note"),
	})
	require.NoError(t, err)

	// Find unresolved links for specific title
	links, err := service.FindUnresolvedLinksByDestTitle(ctx, utils.NullString("Future Note"))
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, "Future Note", links[0].DestTitle.String)
}

func TestCountUnresolvedLinks(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	srcID := createTestNote(t, queries, "Source")

	// Initially zero
	count, err := service.CountUnresolvedLinks(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Create unresolved links
	_, err = queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Note 1"),
	})
	require.NoError(t, err)

	_, err = queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Note 2"),
	})
	require.NoError(t, err)

	// Count should be 2
	count, err = service.CountUnresolvedLinks(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestResolveLink(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create unresolved link
	srcID := createTestNote(t, queries, "Source")
	linkID, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Future Note"),
	})
	require.NoError(t, err)

	// Create the destination note
	destID := createTestNote(t, queries, "Future Note")

	// Resolve the link
	err = service.ResolveLink(ctx, store.ResolveLinkParams{
		ID:     linkID,
		DestID: utils.NullInt64(destID),
	})
	require.NoError(t, err)

	// Verify link is resolved
	link, err := queries.GetLinkByID(ctx, linkID)
	require.NoError(t, err)
	require.Equal(t, destID, link.DestID.Int64)
	require.True(t, link.DestID.Valid)
	require.Equal(t, int64(1), link.Resolved.Int64) // Resolved
	require.False(t, link.DestTitle.Valid)          // Title cleared
}

func TestMarkLinkBroken(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	// Create unresolved link
	srcID := createTestNote(t, queries, "Source")
	linkID, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Broken Note"),
	})
	require.NoError(t, err)

	// Mark as broken
	err = service.MarkLinkBroken(ctx, linkID)
	require.NoError(t, err)

	// Verify link is marked broken
	link, err := queries.GetLinkByID(ctx, linkID)
	require.NoError(t, err)
	require.Equal(t, int64(-1), link.Resolved.Int64) // Broken
}

// ============================================================================
// Broken/Orphaned Links Tests
// ============================================================================

func TestListBrokenLinks(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	srcID := createTestNote(t, queries, "Source")

	// Create unresolved link and mark as broken
	linkID, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Broken"),
	})
	require.NoError(t, err)

	err = queries.MarkLinkBroken(ctx, linkID)
	require.NoError(t, err)

	// List broken links
	links, err := service.ListBrokenLinks(ctx)
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, int64(-1), links[0].Resolved.Int64)
}

func TestCountBrokenLinks(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	srcID := createTestNote(t, queries, "Source")

	// Initially zero
	count, err := service.CountBrokenLinks(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Create and mark broken links
	link1ID, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Broken 1"),
	})
	require.NoError(t, err)
	err = queries.MarkLinkBroken(ctx, link1ID)
	require.NoError(t, err)

	link2ID, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Broken 2"),
	})
	require.NoError(t, err)
	err = queries.MarkLinkBroken(ctx, link2ID)
	require.NoError(t, err)

	// Count should be 2
	count, err = service.CountBrokenLinks(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestListOrphanedLinks(t *testing.T) {
	service, queries := setupTestService(t)
	ctx := context.Background()

	srcID := createTestNote(t, queries, "Source")

	// Create unresolved link (dest_id is NULL = orphaned)
	_, err := queries.CreateUnresolvedLink(ctx, store.CreateUnresolvedLinkParams{
		SrcID:     srcID,
		DestTitle: utils.NullString("Orphaned"),
	})
	require.NoError(t, err)

	// List orphaned links
	links, err := service.ListOrphanedLinks(ctx)
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.False(t, links[0].DestID.Valid) // NULL dest_id
}
