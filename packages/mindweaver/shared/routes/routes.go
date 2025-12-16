// Package routes provides route path constants to ensure consistency
// and prevent typos across the application.
//
// All API routes are prefixed with /api to separate them from future
// web/template routes (e.g., for HTMX, Wails desktop app).
package routes

// APIPrefix is the base prefix for all API routes.
// This separates API endpoints from web/template routes.
const APIPrefix = "/api"

// Shared routes (no prefix)
const (
	Health = "/health"
)

// Mind service routes (Notes/PKM/Knowledge)
// These routes are full paths relative to the mind subgroup: /api/mind + route
// Example: mindGroup.GET(routes.MindNotesID, handler) registers /api/mind/notes/:id
const (
	// Notes
	MindNotes          = "/notes"
	MindNotesID        = "/notes/:id"
	MindNotesMeta      = "/notes/:id/meta"
	MindNotesLinks     = "/notes/:id/links"
	MindNotesBacklinks = "/notes/:id/backlinks"
	MindNotesTags      = "/notes/:id/tags"

	// Tags
	MindTags           = "/tags"
	MindTagsID         = "/tags/:id"
	MindTagsSearchName = "/tags/search/name"
	MindTagsNotes      = "/tags/:id/notes"

	// Templates
	MindTemplates   = "/templates"
	MindTemplatesID = "/templates/:id"
	// Note: Template search routes removed - use general search with filters instead

	// Note Meta
	MindNotesMeta2  = "/notes-meta"
	MindNotesMetaID = "/notes-meta/:id"

	// Note Links
	MindNoteLinks   = "/note-links"
	MindNoteLinksID = "/note-links/:id"

	// Note Types
	MindNoteTypes   = "/note-types"
	MindNoteTypesID = "/note-types/:id"

	// Collections (Folders/Hierarchy)
	MindCollections            = "/collections"
	MindCollectionsID          = "/collections/:id"
	MindCollectionsByPath      = "/collections/by-path"
	MindCollectionsTree        = "/collections/tree"
	MindCollectionsSubtree     = "/collections/:id/subtree"
	MindCollectionsAncestors   = "/collections/:id/ancestors"
	MindCollectionsDescendants = "/collections/:id/descendants"
	MindCollectionsStats       = "/collections/:id/stats"
	MindCollectionsNotes       = "/collections/:id/notes"

	// Import
	MindImport                 = "/import"
	MindImportBatch            = "/import/batch"
	MindImportOperationsID     = "/import/operations/:id"
	MindImportOperationsStream = "/import/operations/:id/stream"
	MindImportLinksResolve     = "/import/links/resolve"
	MindImportLinksStats       = "/import/links/stats"
)

// Brain service routes (AI/Assistant/Chat)
// These routes are full paths relative to the brain subgroup: /api/brain + route
// Example: brainGroup.GET(routes.BrainAssistantsID, handler) registers /api/brain/assistants/:id
const (
	// Assistants
	BrainAssistants          = "/assistants"
	BrainAssistantsID        = "/assistants/:id"
	BrainAssistantsActive    = "/assistants/active"
	BrainAssistantsSetActive = "/assistants/:id/active"

	// Chat
	BrainChat = "/chat"

	// Prompts
	BrainPrompts   = "/prompts"
	BrainPromptsID = "/prompts/:id"

	// Conversations
	BrainConversations   = "/conversations"
	BrainConversationsID = "/conversations/:id"

	// Messages
	BrainConversationMessages = "/conversations/:conversation_id/messages"
	BrainMessages             = "/messages"
	BrainMessagesID           = "/messages/:id"

	// Ingestion (Background Intelligence)
	BrainIngestFull   = "/ingest/full"
	BrainIngestBatch  = "/ingest/batch"
	BrainIngestNote   = "/ingest/note/:id"
	BrainIngestStatus = "/ingest/status"

	// Query Engine (3-Tier Intelligence)
	BrainQuery = "/query"
)
