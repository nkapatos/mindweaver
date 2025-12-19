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
---@param name string Collection display name
---@param parent_id? number Parent collection ID (nil for root)
---@param cb fun(collection: table|nil, error: table|nil) Callback with created collection or error
function M.create_collection(name, parent_id, cb)
  ---@type mind.v3.CreateCollectionRequest
  local req = {
    displayName = name,
  }
  
  if parent_id then
    req.parentId = parent_id
  end
  
  api.collections.create(req, function(res)
    if res.error then
      cb(nil, res.error)
      return
    end
    
    ---@type mind.v3.Collection
    local collection = res.data
    cb(collection, nil)
  end)
end

--- Delete a collection
---@param collection_id number Collection ID to delete
---@param cb fun(success: boolean, error: table|nil) Callback
function M.delete_collection(collection_id, cb)
  ---@type mind.v3.DeleteCollectionRequest
  local req = {
    id = collection_id,
  }
  
  api.collections.delete(req, function(res)
    if res.error then
      cb(false, res.error)
      return
    end
    
    cb(true, nil)
  end)
end

--- Update a collection (rename and/or move)
---@param collection_id number Collection ID to update
---@param opts { displayName?: string, parentId?: number, description?: string, position?: number }
---@param cb fun(collection: table|nil, error: table|nil) Callback
function M.update_collection(collection_id, opts, cb)
  -- First fetch current collection to get current values
  api.collections.get({ id = collection_id }, function(get_res)
    if get_res.error then
      cb(nil, get_res.error)
      return
    end
    
    local current = get_res.data
    
    ---@type mind.v3.UpdateCollectionRequest
    local req = {
      id = collection_id,
      displayName = opts.displayName or current.displayName,
    }
    
    -- Optional fields
    if opts.parentId ~= nil then
      req.parentId = opts.parentId
    elseif current.parentId then
      req.parentId = current.parentId
    end
    
    if opts.description ~= nil then
      req.description = opts.description
    elseif current.description then
      req.description = current.description
    end
    
    if opts.position ~= nil then
      req.position = opts.position
    elseif current.position then
      req.position = current.position
    end
    
    api.collections.update(req, function(res)
      if res.error then
        cb(nil, res.error)
        return
      end
      
      ---@type mind.v3.Collection
      local collection = res.data
      cb(collection, nil)
    end)
  end)
end

--- Rename a collection (convenience wrapper around update_collection)
---@param collection_id number Collection ID to rename
---@param new_name string New display name
---@param cb fun(collection: table|nil, error: table|nil) Callback
function M.rename_collection(collection_id, new_name, cb)
  M.update_collection(collection_id, { displayName = new_name }, cb)
end

function M.setup(opts)
  opts = opts or {}
  -- Future: Configuration options for collections
end

return M
