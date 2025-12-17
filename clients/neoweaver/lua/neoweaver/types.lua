--- Type definitions for MindWeaver API
---
--- AUTO-GENERATED from Protocol Buffer definitions (proto/mind/v3/*.proto).
--- Do not edit manually.
---
--- To regenerate these types:
---   task neoweaver:types:generate
---
--- This file provides LuaLS type annotations for the MindWeaver v3 API,
--- enabling autocomplete and type checking in Neovim plugin development.
---
---@module neoweaver.types
local M = {}

-- From collections.proto

---@class mind.v3.Collection
---@field name string
---@field id integer
---@field displayName string
---@field parentId? integer
---@field path string
---@field description? string
---@field position? integer
---@field isSystem boolean
---@field createTime? string
---@field updateTime? string

---@class mind.v3.CreateCollectionRequest
---@field displayName string
---@field parentId? integer
---@field description? string
---@field position? integer

---@class mind.v3.GetCollectionRequest
---@field id integer

---@class mind.v3.UpdateCollectionRequest
---@field id integer
---@field displayName string
---@field parentId? integer
---@field description? string
---@field position? integer

---@class mind.v3.DeleteCollectionRequest
---@field id integer

---@class mind.v3.ListCollectionsRequest
---@field pageSize number
---@field pageToken string
---@field parentId? integer

---@class mind.v3.ListCollectionsResponse
---@field collections Collection[]
---@field nextPageToken string
---@field totalSize? number

---@class mind.v3.ListCollectionChildrenRequest
---@field parentId integer
---@field pageSize number
---@field pageToken string

---@class mind.v3.GetCollectionTreeRequest
---@field rootId integer
---@field maxDepth number

---@class mind.v3.GetCollectionTreeResponse
---@field root? Collection
---@field descendants Collection[]

-- From links.proto

---@class mind.v3.Link
---@field name string
---@field id integer
---@field srcId integer
---@field destId? integer
---@field destTitle? string
---@field displayText? string
---@field isEmbed? boolean
---@field context? string
---@field resolved? integer
---@field createTime? string
---@field updateTime? string

---@class mind.v3.ListLinksRequest
---@field pageSize number
---@field pageToken string

---@class mind.v3.ListLinksResponse
---@field links Link[]
---@field nextPageToken string
---@field totalSize? number

-- From note_meta.proto

---@class mind.v3.NoteMeta
---@field key string
---@field value string
---@field createTime? string
---@field updateTime? string

---@class mind.v3.ListMetaRequest
---@field noteId integer

---@class mind.v3.ListMetaResponse
---@field items NoteMeta[]
---@field total number

-- From note_types.proto

---@class mind.v3.NoteType
---@field name string
---@field id integer
---@field type string
---@field displayName string
---@field description? string
---@field icon? string
---@field color? string
---@field isSystem boolean
---@field createTime? string
---@field updateTime? string

---@class mind.v3.CreateNoteTypeRequest
---@field type string
---@field displayName string
---@field description? string
---@field icon? string
---@field color? string

---@class mind.v3.GetNoteTypeRequest
---@field id integer

---@class mind.v3.UpdateNoteTypeRequest
---@field id integer
---@field type string
---@field displayName string
---@field description? string
---@field icon? string
---@field color? string

---@class mind.v3.DeleteNoteTypeRequest
---@field id integer

---@class mind.v3.ListNoteTypesRequest
---@field pageSize number
---@field pageToken string

---@class mind.v3.ListNoteTypesResponse
---@field noteTypes NoteType[]
---@field nextPageToken string
---@field totalSize? number

-- From notes.proto

---@class mind.v3.Note
---@field name string
---@field id integer
---@field uuid string
---@field title string
---@field body? string
---@field description? string
---@field noteTypeId? integer
---@field collectionId integer
---@field isTemplate? boolean
---@field etag string
---@field createTime? string
---@field updateTime? string
---@field metadata table<string, string>

---@class mind.v3.CreateNoteRequest
---@field title string
---@field body? string
---@field description? string
---@field noteTypeId? integer
---@field collectionId? integer
---@field isTemplate? boolean
---@field metadata table<string, string>

---@class mind.v3.GetNoteRequest
---@field id integer

---@class mind.v3.ReplaceNoteRequest
---@field id integer
---@field title string
---@field body? string
---@field description? string
---@field noteTypeId? integer
---@field collectionId? integer
---@field isTemplate? boolean
---@field metadata table<string, string>

---@class mind.v3.DeleteNoteRequest
---@field id integer

---@class mind.v3.ListNotesRequest
---@field pageSize number
---@field pageToken string
---@field collectionId? integer
---@field noteTypeId? integer
---@field isTemplate? boolean

---@class mind.v3.ListNotesResponse
---@field notes Note[]
---@field nextPageToken string
---@field totalSize? number

---@class mind.v3.GetNoteMetaRequest
---@field noteId integer

---@class mind.v3.GetNoteMetaResponse
---@field metadata table<string, string>

---@class mind.v3.GetNoteRelationshipsRequest
---@field noteId integer

---@class mind.v3.GetNoteRelationshipsResponse
---@field outgoingLinks integer[]
---@field incomingLinks integer[]
---@field tagIds integer[]

-- From search.proto

---@class mind.v3.SearchNotesRequest
---@field query string
---@field limit? number
---@field offset? number
---@field includeBody? boolean
---@field minScore? number

---@class mind.v3.SearchResult
---@field id integer
---@field title string
---@field snippet string
---@field score number
---@field createTime? string

---@class mind.v3.SearchNotesResponse
---@field results SearchResult[]
---@field total number
---@field query string
---@field durationMs integer
---@field limit number
---@field offset number

-- From tags.proto

---@class mind.v3.Tag
---@field name string
---@field id integer
---@field displayName string
---@field createTime? string
---@field updateTime? string

---@class mind.v3.ListTagsRequest
---@field pageSize number
---@field pageToken string
---@field noteId? integer

---@class mind.v3.ListTagsResponse
---@field tags Tag[]
---@field nextPageToken string
---@field totalSize? number

-- From templates.proto

---@class mind.v3.Template
---@field name string
---@field id integer
---@field displayName string
---@field description? string
---@field starterNoteId integer
---@field noteTypeId? integer
---@field createTime? string
---@field updateTime? string

---@class mind.v3.CreateTemplateRequest
---@field displayName string
---@field description? string
---@field starterNoteId integer
---@field noteTypeId? integer

---@class mind.v3.ListTemplatesRequest
---@field pageSize number
---@field pageToken string

---@class mind.v3.ListTemplatesResponse
---@field templates Template[]
---@field nextPageToken string
---@field totalSize? number

---@class mind.v3.GetTemplateRequest
---@field id integer

---@class mind.v3.UpdateTemplateRequest
---@field id integer
---@field displayName string
---@field description? string
---@field starterNoteId integer
---@field noteTypeId? integer

---@class mind.v3.DeleteTemplateRequest
---@field id integer

return M
