--- Tree management for the neoweaver explorer
--- Handles collection hierarchy rendering using NuiTree
---
local NuiTree = require("nui.tree")
local NuiLine = require("nui.line")
local collections = require("neoweaver._internal.collections")
local notes = require("neoweaver._internal.notes")

local M = {}

---@class ExplorerTreeState
---@field tree NuiTree|nil
---@field bufnr number|nil
---@field collections table[]|nil

---@type ExplorerTreeState
local state = {
  tree = nil,
  bufnr = nil,
  collections = nil,
}

--- Build tree nodes recursively from flat collection list with notes
---@param collections table[] Flat list of collections
---@param notes_by_collection table<number, table[]> Hashmap of notes grouped by collection_id
---@param parent_id number|nil Parent collection ID (nil for roots)
---@return NuiTree.Node[]
local function build_nodes(collections, notes_by_collection, parent_id)
  local nodes = {}

  -- Find all collections with the given parent_id
  for _, collection in ipairs(collections) do
    if collection.parentId == parent_id then
      local children = {}
      
      -- Add note children first (already sorted alphabetically)
      local notes = notes_by_collection[collection.id] or {}
      for _, note in ipairs(notes) do
        local note_node = NuiTree.Node({
          id = "note:" .. note.id,
          type = "note",
          name = note.title,
          note_id = note.id,
          collection_id = note.collectionId,
          data = note,
        })
        table.insert(children, note_node)
      end
      
      -- Then recursively add child collections
      local child_collections = build_nodes(collections, notes_by_collection, collection.id)
      vim.list_extend(children, child_collections)

      -- Create collection node with all children
      local node = NuiTree.Node({
        id = "collection:" .. collection.id,
        type = "collection",
        name = collection.displayName,
        collection_id = collection.id,
        is_system = collection.isSystem or false,
        data = collection,
      }, children)

      table.insert(nodes, node)
    end
  end

  return nodes
end

--- Prepare a node for rendering (handles both collections and notes)
---@param node NuiTree.Node
---@return NuiLine
local function prepare_node(node)
  local line = NuiLine()

  -- Indentation (2 spaces per level)
  local indent = string.rep("  ", node:get_depth() - 1)
  line:append(indent)

  -- Expand/collapse indicator for nodes with children
  if node:has_children() then
    if node:is_expanded() then
      line:append("▾ ", "NeoTreeExpander")
    else
      line:append("▸ ", "NeoTreeExpander")
    end
  else
    line:append("  ")
  end

  -- Icon and name based on node type
  if node.type == "note" then
    -- Note node
    line:append("󰈙 ", "String")  -- Document icon
    line:append(node.name, "Normal")
  else
    -- Collection node
    if node.is_system then
      line:append("󰉖 ", "Special") -- System collection icon
    else
      line:append("󰉋 ", "Directory") -- Regular collection icon
    end
    local name_hl = node.is_system and "Comment" or "Directory"
    line:append(node.name, name_hl)
  end

  return line
end

--- Build the tree from collections and notes data
---@param bufnr number Buffer to render tree in
---@param collections_data table[] Array of collection objects
---@param notes_by_collection table<number, table[]> Hashmap of notes grouped by collection_id
---@return NuiTree
function M.build_tree(bufnr, collections_data, notes_by_collection)
  local root_nodes = build_nodes(collections_data, notes_by_collection or {}, nil)

  local tree = NuiTree({
    bufnr = bufnr,
    nodes = root_nodes,
    prepare_node = prepare_node,
  })

  return tree
end

--- Setup keymaps for tree navigation and actions
---@param bufnr number
---@param tree NuiTree
local function setup_keymaps(bufnr, tree)
  local map_opts = { noremap = true, nowait = true }

  -- Open note or toggle collection expand/collapse on Enter
  vim.keymap.set("n", "<CR>", function()
    local node = tree:get_node()
    if not node then return end
    
    -- Handle note nodes - open for editing
    if node.type == "note" then
      notes.open_note(node.note_id)
      return
    end
    
    -- Handle collection nodes - expand/collapse
    if node:has_children() then
      if node:is_expanded() then
        node:collapse()
      else
        node:expand()
      end
      tree:render()
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Open note or toggle collection" }))

  -- Open note with 'o' (alternative to Enter)
  vim.keymap.set("n", "o", function()
    local node = tree:get_node()
    if not node then return end
    
    if node.type == "note" then
      notes.open_note(node.note_id)
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Open note" }))

  vim.keymap.set("n", "l", function()
    local node = tree:get_node()
    if node and node:has_children() then
      if not node:is_expanded() then
        node:expand()
        tree:render()
      end
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Expand collection" }))

  -- Collapse on 'h'
  vim.keymap.set("n", "h", function()
    local node = tree:get_node()
    if node then
      if node:is_expanded() and node:has_children() then
        -- Collapse current node
        node:collapse()
        tree:render()
      else
        -- Navigate to parent
        local parent = node:get_parent_id()
        if parent then
          local parent_node = tree:get_node(parent)
          if parent_node then
            tree:set_node(parent)
            tree:render()
          end
        end
      end
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Collapse or go to parent" }))

  -- Navigation with j/k (default vim bindings work, but ensure proper cursor behavior)
  vim.keymap.set("n", "j", function()
    vim.cmd("normal! j")
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Move down" }))

  vim.keymap.set("n", "k", function()
    vim.cmd("normal! k")
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Move up" }))

  -- Refresh tree
  vim.keymap.set("n", "R", function()
    M.refresh(bufnr)
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Refresh tree" }))

  -- Generic action keybindings - delegate to action handlers
  -- These will be wired up by the explorer based on what's being displayed
  
  -- Create (a) - handled by action handler
  vim.keymap.set("n", "a", function()
    local node = tree:get_node()
    if M.on_create and node then
      M.on_create(node, bufnr)
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Create item" }))

  -- Rename (r) - handled by action handler
  vim.keymap.set("n", "r", function()
    local node = tree:get_node()
    if M.on_rename and node then
      M.on_rename(node, bufnr)
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Rename item" }))

  -- Delete (d) - handled by action handler
  vim.keymap.set("n", "d", function()
    local node = tree:get_node()
    if M.on_delete and node then
      M.on_delete(node, bufnr)
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Delete item" }))
end

--- Helper to set buffer lines (handles modifiable state)
---@param bufnr number
---@param lines string[]
local function set_buffer_lines(bufnr, lines)
  vim.api.nvim_set_option_value("modifiable", true, { buf = bufnr })
  vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
  vim.api.nvim_set_option_value("modifiable", false, { buf = bufnr })
end

--- Load collections with notes and render tree
--- Async function that fetches collections and notes from API
---@param bufnr number Buffer to render tree in
function M.load_and_render(bufnr)
  state.bufnr = bufnr
  
  -- Show loading indicator
  set_buffer_lines(bufnr, { "Loading collections and notes..." })
  
  -- Fetch collections with notes from API (orchestrated call)
  collections.list_collections_with_notes({}, function(data, err)
    if err then
      -- Clear loading message and show error
      set_buffer_lines(bufnr, {})
      vim.notify("Failed to load collections: " .. vim.inspect(err), vim.log.levels.ERROR)
      return
    end
    
    -- Handle empty collections
    if not data or not data.collections or #data.collections == 0 then
      set_buffer_lines(bufnr, { "No collections found" })
      vim.notify("No collections found", vim.log.levels.INFO)
      return
    end
    
    -- Store collections data
    state.collections = data.collections
    
    -- Build and render tree with notes
    -- NuiTree.render() handles modifiable state automatically
    state.tree = M.build_tree(bufnr, data.collections, data.notes_by_collection)
    setup_keymaps(bufnr, state.tree)
    state.tree:render()
  end)
end

--- Initialize and render the tree in the given buffer
--- Alias for load_and_render for backward compatibility
---@param bufnr number
function M.init(bufnr)
  M.load_and_render(bufnr)
end

--- Refresh the tree (rebuild and re-render)
--- Re-fetches collections and notes from API
---@param bufnr number
function M.refresh(bufnr)
  if state.bufnr == bufnr then
    -- Show loading indicator
    set_buffer_lines(bufnr, { "Refreshing collections and notes..." })
    
    -- Re-fetch collections with notes from API (orchestrated call)
    collections.list_collections_with_notes({}, function(data, err)
      if err then
        vim.notify("Failed to refresh collections: " .. (err.message or "unknown error"), vim.log.levels.ERROR)
        return
      end
      
      -- TODO: Store expanded state before rebuild and restore after
      state.collections = data.collections
      state.tree = M.build_tree(bufnr, data.collections, data.notes_by_collection)
      setup_keymaps(bufnr, state.tree)
      -- NuiTree.render() handles modifiable state automatically
      state.tree:render()
      
      vim.notify("Collections and notes refreshed", vim.log.levels.INFO)
    end)
  end
end

--- Get current tree instance
---@return NuiTree|nil
function M.get_tree()
  return state.tree
end

return M
