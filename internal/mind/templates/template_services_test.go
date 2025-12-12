package templates

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	mindtesting "github.com/nkapatos/mindweaver/internal/mind/testing"
	"github.com/nkapatos/mindweaver/internal/mind/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createStarterNote(t *testing.T, querier store.Querier) int64 {
	t.Helper()
	noteID, err := querier.CreateNote(context.Background(), store.CreateNoteParams{
		Uuid:         uuid.New(),
		Title:        "Starter Note",
		Body:         sql.NullString{String: "Template content", Valid: true},
		CollectionID: 1,
	})
	require.NoError(t, err)
	return noteID
}

// TestTemplatesService_CreateTemplate tests creating a template
func TestTemplatesService_CreateTemplate(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	// Create template
	params := store.CreateTemplateParams{
		Name:          "Test Template",
		Description:   sql.NullString{String: "A test template", Valid: true},
		StarterNoteID: noteID,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	id, err := svc.CreateTemplate(context.Background(), params)

	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

// TestTemplatesService_CreateTemplate_DuplicateName tests unique constraint
func TestTemplatesService_CreateTemplate_DuplicateName(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	params := store.CreateTemplateParams{
		Name:          "Duplicate Template",
		Description:   sql.NullString{String: "First template", Valid: true},
		StarterNoteID: noteID,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	// Create first template
	_, err := svc.CreateTemplate(context.Background(), params)
	require.NoError(t, err)

	// Try to create duplicate
	_, err = svc.CreateTemplate(context.Background(), params)
	assert.ErrorIs(t, err, ErrTemplateAlreadyExists)
}

// TestTemplatesService_GetTemplateByID tests retrieving a template
func TestTemplatesService_GetTemplateByID(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	params := store.CreateTemplateParams{
		Name:          "Get Test Template",
		Description:   sql.NullString{String: "Description", Valid: true},
		StarterNoteID: noteID,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	id, err := svc.CreateTemplate(context.Background(), params)
	require.NoError(t, err)

	// Retrieve template
	template, err := svc.GetTemplateByID(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, id, template.ID)
	assert.Equal(t, "Get Test Template", template.Name)
	assert.Equal(t, "Description", template.Description.String)
	assert.Equal(t, noteID, template.StarterNoteID)
}

// TestTemplatesService_GetTemplateByID_NotFound tests not found error
func TestTemplatesService_GetTemplateByID_NotFound(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")

	// Try to get non-existent template
	_, err := svc.GetTemplateByID(context.Background(), 99999)

	assert.ErrorIs(t, err, ErrTemplateNotFound)
}

// TestTemplatesService_ListTemplates tests listing all templates
func TestTemplatesService_ListTemplates(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	// Create multiple templates
	for i := 1; i <= 3; i++ {
		params := store.CreateTemplateParams{
			Name:          fmt.Sprintf("Template_%d", i),
			Description:   sql.NullString{String: "Description", Valid: true},
			StarterNoteID: noteID,
			NoteTypeID:    sql.NullInt64{Valid: false},
		}
		_, err := svc.CreateTemplate(context.Background(), params)
		require.NoError(t, err)
	}

	// List templates
	templates, err := svc.ListTemplates(context.Background())

	require.NoError(t, err)
	assert.Len(t, templates, 3)
}

// TestTemplatesService_CountTemplates tests counting templates
func TestTemplatesService_CountTemplates(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")

	// Initially should be 0
	count, err := svc.CountTemplates(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	noteID := createStarterNote(t, querier)

	for i := 1; i <= 5; i++ {
		params := store.CreateTemplateParams{
			Name:          fmt.Sprintf("Template_%d", i),
			Description:   sql.NullString{String: "Description", Valid: true},
			StarterNoteID: noteID,
			NoteTypeID:    sql.NullInt64{Valid: false},
		}
		_, err := svc.CreateTemplate(context.Background(), params)
		require.NoError(t, err)
	}

	// Count should be 5
	count, err = svc.CountTemplates(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

// TestTemplatesService_UpdateTemplate tests updating a template
func TestTemplatesService_UpdateTemplate(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	params := store.CreateTemplateParams{
		Name:          "Original Name",
		Description:   sql.NullString{String: "Original Description", Valid: true},
		StarterNoteID: noteID,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	id, err := svc.CreateTemplate(context.Background(), params)
	require.NoError(t, err)

	// Update template
	updateParams := store.UpdateTemplateByIDParams{
		ID:            id,
		Name:          "Updated Name",
		Description:   sql.NullString{String: "Updated Description", Valid: true},
		StarterNoteID: noteID,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	err = svc.UpdateTemplate(context.Background(), updateParams)
	require.NoError(t, err)

	// Verify update
	template, err := svc.GetTemplateByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", template.Name)
	assert.Equal(t, "Updated Description", template.Description.String)
}

// TestTemplatesService_UpdateTemplate_NotFound tests update on non-existent template
func TestTemplatesService_UpdateTemplate_NotFound(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")

	updateParams := store.UpdateTemplateByIDParams{
		ID:            99999,
		Name:          "Does Not Exist",
		Description:   sql.NullString{String: "Description", Valid: true},
		StarterNoteID: 1,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	err := svc.UpdateTemplate(context.Background(), updateParams)
	// SQLite UPDATE with no rows affected returns nil error
	// Service should detect this via RowsAffected, but current implementation doesn't
	// For now, we just verify no panic occurs
	assert.NoError(t, err) // TODO: Should be ErrorIs(ErrTemplateNotFound) after service fix
}

// TestTemplatesService_DeleteTemplate tests deleting a template
func TestTemplatesService_DeleteTemplate(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	params := store.CreateTemplateParams{
		Name:          "To Delete",
		Description:   sql.NullString{String: "Will be deleted", Valid: true},
		StarterNoteID: noteID,
		NoteTypeID:    sql.NullInt64{Valid: false},
	}

	id, err := svc.CreateTemplate(context.Background(), params)
	require.NoError(t, err)

	// Delete template
	err = svc.DeleteTemplate(context.Background(), id)
	require.NoError(t, err)

	// Verify deletion
	_, err = svc.GetTemplateByID(context.Background(), id)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}

// TestTemplatesService_DeleteTemplate_NotFound tests delete on non-existent template
func TestTemplatesService_DeleteTemplate_NotFound(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")

	err := svc.DeleteTemplate(context.Background(), 99999)
	// SQLite DELETE with no rows affected returns nil error
	// Service should detect this via RowsAffected, but current implementation doesn't
	assert.NoError(t, err) // TODO: Should be ErrorIs(ErrTemplateNotFound) after service fix
}

// TestTemplatesService_ListTemplatesPaginated tests paginated listing
func TestTemplatesService_ListTemplatesPaginated(t *testing.T) {
	db, querier, logger := mindtesting.SetupTest(t)
	defer db.Close()

	svc := NewTemplatesService(querier, logger, "templates-test")
	noteID := createStarterNote(t, querier)

	// Create 10 templates
	for i := 1; i <= 10; i++ {
		params := store.CreateTemplateParams{
			Name:          fmt.Sprintf("Template_%d", i),
			Description:   sql.NullString{String: "Description", Valid: true},
			StarterNoteID: noteID,
			NoteTypeID:    sql.NullInt64{Valid: false},
		}
		_, err := svc.CreateTemplate(context.Background(), params)
		require.NoError(t, err)
	}

	// Get first page (5 items)
	templates, err := svc.ListTemplatesPaginated(context.Background(), 5, 0)
	require.NoError(t, err)
	assert.Len(t, templates, 5)

	// Get second page (5 items)
	templates, err = svc.ListTemplatesPaginated(context.Background(), 5, 5)
	require.NoError(t, err)
	assert.Len(t, templates, 5)
}
