--- Tree management for the neoweaver explorer
--- Handles collection hierarchy rendering using NuiTree
---
local NuiTree = require("nui.tree")
local NuiLine = require("nui.line")
local collections = require("neoweaver._internal.collections")

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

--- Build tree nodes recursively from flat collection list
---@param collections table[] Flat list of collections
---@param parent_id string|nil Parent collection ID (nil for roots)
---@return NuiTree.Node[]
local function build_nodes(collections, parent_id)
  local nodes = {}

  -- Find all collections with the given parent_id
  for _, collection in ipairs(collections) do
    if collection.parentId == parent_id then
      -- Build children recursively
      local children = build_nodes(collections, collection.id)

      -- Create node with collection data
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

--- Prepare a node for rendering
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

  -- Icon for collection
  if node.is_system then
    line:append("󰉖 ", "Special") -- System collection icon
  else
    line:append("󰉋 ", "Directory") -- Regular collection icon
  end

  -- Collection name
  local name_hl = node.is_system and "Comment" or "Directory"
  line:append(node.name, name_hl)

  return line
end

--- Build the tree from collections data
---@param bufnr number Buffer to render tree in
---@param collections_data table[] Array of collection objects
---@return NuiTree
function M.build_tree(bufnr, collections_data)
  local root_nodes = build_nodes(collections_data, nil)

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

  -- Toggle expand/collapse on Enter or 'l'
  vim.keymap.set("n", "<CR>", function()
    local node = tree:get_node()
    if node and node:has_children() then
      if node:is_expanded() then
        node:collapse()
      else
        node:expand()
      end
      tree:render()
    end
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Toggle expand/collapse" }))

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

  -- Refresh tree (for future API integration)
  vim.keymap.set("n", "R", function()
    M.refresh(bufnr)
  end, vim.tbl_extend("force", map_opts, { buffer = bufnr, desc = "Refresh tree" }))
end

--- Load collections and render tree
--- Async function that fetches collections from API
---@param bufnr number Buffer to render tree in
function M.load_and_render(bufnr)
  state.bufnr = bufnr
  
  -- Show loading indicator
  vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, { "Loading collections..." })
  
  -- Fetch collections from API
  collections.list_collections({}, function(collections_data, err)
    if err then
      -- Clear loading message
      vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, {})
      -- Show error notification only (like notes module)
      vim.notify("Failed to load collections: " .. vim.inspect(err), vim.log.levels.ERROR)
      return
    end
    
    -- Handle empty collections
    if not collections_data or #collections_data == 0 then
      vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, { "No collections found" })
      vim.notify("No collections found", vim.log.levels.INFO)
      return
    end
    
    -- Store collections data
    state.collections = collections_data
    
    -- Build and render tree
    state.tree = M.build_tree(bufnr, collections_data)
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
--- Re-fetches collections from API
---@param bufnr number
function M.refresh(bufnr)
  if state.bufnr == bufnr then
    -- Show loading indicator
    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, { "Refreshing collections..." })
    
    -- Re-fetch collections from API
    collections.list_collections({}, function(collections_data, err)
      if err then
        vim.notify("Failed to refresh collections: " .. (err.message or "unknown error"), vim.log.levels.ERROR)
        return
      end
      
      -- TODO: Store expanded state before rebuild and restore after
      state.collections = collections_data
      state.tree = M.build_tree(bufnr, collections_data)
      setup_keymaps(bufnr, state.tree)
      state.tree:render()
      
      vim.notify("Collections refreshed", vim.log.levels.INFO)
    end)
  end
end

--- Get current tree instance
---@return NuiTree|nil
function M.get_tree()
  return state.tree
end

return M
