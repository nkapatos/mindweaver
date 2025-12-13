-- AUTO-GENERATED from pkg/proto/mind/v3/*.proto
-- Do not edit manually. Run: task neoweaver:types:generate

local M = {}

-- From collections.proto

---@class mind.v3.Collection
---@field name string
---@field display_name string
---@field path string
---@field description? string
---@field is_system boolean
---@field create_time string
---@field update_time string

---@class mind.v3.CreateCollectionRequest
---@field display_name string
---@field description? string

---@class mind.v3.GetCollectionRequest

---@class mind.v3.UpdateCollectionRequest
---@field display_name string
---@field description? string

---@class mind.v3.DeleteCollectionRequest

---@class mind.v3.ListCollectionsRequest
---@field page_token string

---@class mind.v3.ListCollectionsResponse
---@field collections Collection[]
---@field next_page_token string

---@class mind.v3.ListCollectionChildrenRequest
---@field page_token string

---@class mind.v3.GetCollectionTreeRequest

---@class mind.v3.GetCollectionTreeResponse
---@field root Collection
---@field descendants Collection[]

-- From links.proto

---@class mind.v3.Link
---@field name string
---@field dest_title? string
---@field display_text? string
---@field is_embed? boolean
---@field context? string
---@field create_time string
---@field update_time string

---@class mind.v3.ListLinksRequest

---@class mind.v3.ListLinksResponse
---@field links Link[]
---@field next_page_token string

-- From note_meta.proto

---@class mind.v3.NoteMeta
---@field key string
---@field value string
---@field create_time string
---@field update_time string

---@class mind.v3.ListMetaRequest

---@class mind.v3.ListMetaResponse
---@field items NoteMeta[]

-- From note_types.proto

---@class mind.v3.NoteType
---@field name string
---@field type string
---@field display_name string
---@field description? string
---@field icon? string
---@field color? string
---@field is_system boolean
---@field create_time string
---@field update_time string

---@class mind.v3.CreateNoteTypeRequest
---@field type string
---@field display_name string
---@field description? string
---@field icon? string
---@field color? string

---@class mind.v3.GetNoteTypeRequest

---@class mind.v3.UpdateNoteTypeRequest
---@field type string
---@field display_name string
---@field description? string
---@field icon? string
---@field color? string

---@class mind.v3.DeleteNoteTypeRequest

---@class mind.v3.ListNoteTypesRequest
---@field page_token string

---@class mind.v3.ListNoteTypesResponse
---@field note_types NoteType[]
---@field next_page_token string

-- From notes.proto

---@class mind.v3.Note
---@field name string
---@field uuid string
---@field title string
---@field body? string
---@field description? string
---@field is_template? boolean
---@field etag string
---@field create_time string
---@field update_time string
---@field metadata table<string, string>

---@class mind.v3.CreateNoteRequest
---@field title string
---@field body? string
---@field description? string
---@field is_template? boolean
---@field metadata table<string, string>

---@class mind.v3.GetNoteRequest

---@class mind.v3.ReplaceNoteRequest
---@field title string
---@field body? string
---@field description? string
---@field is_template? boolean
---@field metadata table<string, string>

---@class mind.v3.DeleteNoteRequest

---@class mind.v3.ListNotesRequest
---@field page_token string
---@field is_template? boolean

---@class mind.v3.ListNotesResponse
---@field notes Note[]
---@field next_page_token string

---@class mind.v3.GetNoteMetaRequest

---@class mind.v3.GetNoteMetaResponse
---@field metadata table<string, string>

---@class mind.v3.GetNoteRelationshipsRequest

---@class mind.v3.GetNoteRelationshipsResponse

-- From search.proto

---@class mind.v3.SearchNotesRequest
---@field query string
---@field include_body? boolean
---@field min_score? number

---@class mind.v3.SearchResult
---@field title string
---@field snippet string
---@field score number
---@field create_time string

---@class mind.v3.SearchNotesResponse
---@field results SearchResult[]
---@field query string

-- From tags.proto

---@class mind.v3.Tag
---@field name string
---@field display_name string
---@field create_time string
---@field update_time string

---@class mind.v3.ListTagsRequest

---@class mind.v3.ListTagsResponse
---@field tags Tag[]
---@field next_page_token string

-- From templates.proto

---@class mind.v3.Template
---@field name string
---@field display_name string

---@class mind.v3.CreateTemplateRequest
---@field display_name string

---@class mind.v3.ListTemplatesRequest

---@class mind.v3.ListTemplatesResponse
---@field templates Template[]
---@field next_page_token string

---@class mind.v3.GetTemplateRequest

---@class mind.v3.UpdateTemplateRequest
---@field display_name string

---@class mind.v3.DeleteTemplateRequest

return M
