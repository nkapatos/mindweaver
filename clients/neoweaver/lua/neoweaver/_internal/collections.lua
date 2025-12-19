---
--- collections.lua - Collection management for Neoweaver (v3)
--- Handles collection listing, creation, deletion, and hierarchy management
---
local api = require("neoweaver._internal.api")

local M = {}

--- List all collections
--- Returns flat list with parentId for building hierarchy
---@param opts? { pageSize?: number, pageToken?: string }
---@param cb fun(collections: table[]|nil, error: table|nil) Callback with collections array or error
function M.list_collections(opts, cb)
  opts = opts or {}
  
  ---@type mind.v3.ListCollectionsRequest
  local req = {
    pageSize = opts.pageSize or 50, -- Default 50 collections (server max is 100)
    pageToken = opts.pageToken or "",
  }

  api.collections.list(req, function(res)
    if res.error then
      cb(nil, res.error)
      return
    end

    -- v3 API: Response is mind.v3.ListCollectionsResponse directly
    ---@type mind.v3.ListCollectionsResponse
    local list_res = res.data
    local collections = list_res.collections or {}

    -- TODO: Implement automatic pagination
    -- If nextPageToken is present, recursively fetch remaining pages in background
    -- API response includes: nextPageToken, totalSize
    -- Pagination is per-level (root collections only, not nested children)
    -- When all pages fetched, notify explorer to update tree
    -- Note: nextPageToken is empty string when no more pages available
    -- local next_token = list_res.nextPageToken
    -- local total = list_res.totalSize

    cb(collections, nil)
  end)
end

--- List collections with note titles
--- Orchestrates two API calls: collections + notes, then returns combined data
--- Returns: { collections: table[], notes_by_collection: table<number, table[]> }
---@param opts? { pageSize?: number }
---@param cb fun(data: { collections: table[], notes_by_collection: table }|nil, error: table|nil)
function M.list_collections_with_notes(opts, cb)
  opts = opts or {}
  
  -- Step 1: Fetch collections
  M.list_collections(opts, function(collections, err)
    if err then
      cb(nil, err)
      return
    end
    
    -- Step 2: Fetch all notes with field masking (only id, title, collectionId)
    ---@type mind.v3.ListNotesRequest
    local notes_req = {
      pageSize = opts.pageSize or 100, -- Fetch up to 100 notes (can adjust or paginate later)
      fieldMask = "id,title,collectionId", -- Only request fields needed for tree building
    }
    
    api.notes.list(notes_req, function(notes_res)
      if notes_res.error then
        cb(nil, notes_res.error)
        return
      end
      
      ---@type mind.v3.ListNotesResponse
      local notes_list = notes_res.data
      local notes = notes_list.notes or {}
      
      -- Step 3: Build hashmap - group notes by collection_id
      local notes_by_collection = {}
      for _, note in ipairs(notes) do
        local cid = note.collectionId
        if not notes_by_collection[cid] then
          notes_by_collection[cid] = {}
        end
        table.insert(notes_by_collection[cid], note)
      end
      
      -- Step 4: Sort notes alphabetically by title within each collection
      for _, note_list in pairs(notes_by_collection) do
        table.sort(note_list, function(a, b)
          return a.title < b.title
        end)
      end
      
      -- Step 5: Return combined data
      cb({
        collections = collections,
        notes_by_collection = notes_by_collection,
      }, nil)
    end)
  end)
end

--- Create a new collection
--- TODO: Implement when needed for explorer actions
---@param name string Collection display name
---@param parent_id? string Parent collection ID (nil for root)
---@param cb fun(collection: table|nil, error: table|nil) Callback with created collection or error
function M.create_collection(name, parent_id, cb)
  vim.notify("Collection creation not yet implemented", vim.log.levels.WARN)
  -- TODO: Implementation
  -- api.collections.create({ displayName = name, parentId = parent_id }, cb)
end

--- Delete a collection
--- TODO: Implement when needed for explorer actions
---@param collection_id string Collection ID to delete
---@param cb fun(success: boolean, error: table|nil) Callback
function M.delete_collection(collection_id, cb)
  vim.notify("Collection deletion not yet implemented", vim.log.levels.WARN)
  -- TODO: Implementation
  -- api.collections.delete({ id = collection_id }, cb)
end

--- Rename a collection
--- TODO: Implement when needed for explorer actions
---@param collection_id string Collection ID to rename
---@param new_name string New display name
---@param cb fun(collection: table|nil, error: table|nil) Callback
function M.rename_collection(collection_id, new_name, cb)
  vim.notify("Collection renaming not yet implemented", vim.log.levels.WARN)
  -- TODO: Implementation
  -- api.collections.update({ id = collection_id, displayName = new_name }, cb)
end

function M.setup(opts)
  opts = opts or {}
  -- Future: Configuration options for collections
end

return M
