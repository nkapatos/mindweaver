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

    -- Note: Automatic pagination not implemented - See issue #52
    -- API includes: nextPageToken, totalSize
    -- Could recursively fetch remaining pages in background

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

--- Build generic tree nodes from collections and notes data
--- Recursive function that builds collection hierarchy with notes
---@param collections_data table[] Flat list of collections
---@param notes_by_collection table<number, table[]> Notes grouped by collection_id
---@param parent_id number|nil Parent collection ID (nil for roots)
---@return table[] Array of generic tree nodes
local function build_collection_nodes_recursive(collections_data, notes_by_collection, parent_id)
  local nodes = {}

  -- Find all collections with the given parent_id
  for _, collection in ipairs(collections_data) do
    if collection.parentId == parent_id then
      local children = {}

      -- Add note children first
      local collection_notes = notes_by_collection[collection.id] or {}
      for _, note in ipairs(collection_notes) do
        table.insert(children, {
          id = "note:" .. note.id,
          type = "note",
          name = note.title,
          icon = "󰈙",
          highlight = "String",
          note_id = note.id,
          collection_id = note.collectionId,
          data = note,
        })
      end

      -- Then recursively add child collections
      local child_collections = build_collection_nodes_recursive(collections_data, notes_by_collection, collection.id)
      vim.list_extend(children, child_collections)

      -- Create collection node
      local node = {
        id = "collection:" .. collection.id,
        type = "collection",
        name = collection.displayName,
        icon = collection.isSystem and "󰉖" or "󰉋",
        highlight = collection.isSystem and "Special" or "Directory",
        collection_id = collection.id,
        is_system = collection.isSystem or false,
        data = collection,
        children = children,
      }

      table.insert(nodes, node)
    end
  end

  return nodes
end

--- Build tree nodes for collections mode (with server root)
--- Fetches collections and notes, then builds generic tree node structure
---@param callback fun(nodes: table[]|nil, error: table|nil, stats: { collections: number, notes: number }|nil)
function M.build_tree_nodes(callback)
  -- Fetch collections with notes
  M.list_collections_with_notes({}, function(data, err)
    if err then
      callback(nil, err, nil)
      return
    end

    -- Handle empty collections
    if not data or not data.collections or #data.collections == 0 then
      callback({}, nil, { collections = 0, notes = 0 })
      return
    end

    -- Build collection hierarchy
    local collection_nodes = build_collection_nodes_recursive(data.collections, data.notes_by_collection or {}, nil)

    -- Wrap in server node
    local api_mod = require("neoweaver._internal.api")
    local servers = api_mod.config.servers
    local current_server = api_mod.config.current_server
    local root_nodes = {}

    if current_server and servers[current_server] then
      local server_node = {
        id = "server:" .. current_server,
        type = "server",
        name = current_server,
        icon = "󰒋",
        highlight = "Title",
        server_name = current_server,
        server_url = servers[current_server].url,
        is_default = true,
        children = collection_nodes,
      }
      table.insert(root_nodes, server_node)
    else
      -- Fallback: show collections directly
      root_nodes = collection_nodes
    end

    -- Count notes
    local note_count = 0
    if data.notes_by_collection then
      for _, note_list in pairs(data.notes_by_collection) do
        note_count = note_count + #note_list
      end
    end

    -- Return nodes and stats
    callback(root_nodes, nil, {
      collections = #data.collections,
      notes = note_count,
    })
  end)
end

--- Handle collection create action
---@param node table Tree node
---@param refresh_callback fun() Callback to refresh tree after action
function M.handle_create(node, refresh_callback)
  -- Allow creating collections under:
  -- 1. Server nodes (creates root-level collection with no parent)
  -- 2. Collection nodes (creates child collection)
  if node.type ~= "server" and node.type ~= "collection" then
    vim.notify("Can only create collections under servers or other collections", vim.log.levels.WARN)
    return
  end

  -- Determine parent_id: nil for server nodes, collection_id for collection nodes
  local parent_id = nil
  if node.type == "collection" then
    parent_id = node.collection_id
  end

  -- Prompt for collection name
  vim.ui.input({ prompt = "New collection name: " }, function(name)
    if not name or name == "" then
      return
    end

    M.create_collection(name, parent_id, function(collection, err)
      if err then
        vim.notify("Failed to create collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
        return
      end

      vim.notify("Created collection: " .. collection.displayName, vim.log.levels.INFO)
      refresh_callback()
    end)
  end)
end

--- Handle collection rename action
---@param node table Tree node
---@param refresh_callback fun() Callback to refresh tree after action
function M.handle_rename(node, refresh_callback)
  -- Only allow renaming collections
  if node.type ~= "collection" then
    vim.notify("Can only rename collections", vim.log.levels.WARN)
    return
  end

  -- Don't allow renaming system collections
  if node.is_system then
    vim.notify("Cannot rename system collections", vim.log.levels.WARN)
    return
  end

  -- Prompt for new name
  vim.ui.input({ prompt = "Rename to: ", default = node.name }, function(new_name)
    if not new_name or new_name == "" or new_name == node.name then
      return
    end

    M.rename_collection(node.collection_id, new_name, function(collection, err)
      if err then
        vim.notify("Failed to rename collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
        return
      end

      vim.notify("Renamed collection to: " .. collection.displayName, vim.log.levels.INFO)
      refresh_callback()
    end)
  end)
end

--- Handle collection delete action
---@param node table Tree node
---@param refresh_callback fun() Callback to refresh tree after action
function M.handle_delete(node, refresh_callback)
  -- Only allow deleting collections
  if node.type ~= "collection" then
    vim.notify("Can only delete collections", vim.log.levels.WARN)
    return
  end

  -- Don't allow deleting system collections
  if node.is_system then
    vim.notify("Cannot delete system collections", vim.log.levels.WARN)
    return
  end

  -- Confirm deletion
  vim.ui.input({
    prompt = "Delete collection '" .. node.name .. "'? (y/N): ",
  }, function(confirm)
    if confirm ~= "y" and confirm ~= "Y" then
      return
    end

    M.delete_collection(node.collection_id, function(_, err)
      if err then
        vim.notify("Failed to delete collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
        return
      end

      vim.notify("Deleted collection: " .. node.name, vim.log.levels.INFO)
      refresh_callback()
    end)
  end)
end

function M.setup(opts)
  opts = opts or {}
  -- Future: Configuration options for collections
end

return M
