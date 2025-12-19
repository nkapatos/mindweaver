--- Generic tree renderer for the neoweaver explorer
--- Accepts pre-built node structures and renders them using NuiTree
--- No domain knowledge - works with any hierarchical data
---
local NuiTree = require("nui.tree")
local NuiLine = require("nui.line")

local M = {}

---@class GenericTreeNode
---@field id string|number Unique identifier
---@field type string Node type (for identification)
---@field name string Display name
---@field icon? string Icon to display (optional)
---@field highlight? string Highlight group for name (optional)
---@field data? table Domain-specific data (optional)
---@field children? GenericTreeNode[] Child nodes (optional)

---@class ExplorerTreeState
---@field tree NuiTree|nil

---@type ExplorerTreeState
local state = {
  tree = nil,
}

--- Convert generic tree nodes to NuiTree.Node recursively
---@param generic_nodes GenericTreeNode[] Array of generic nodes
---@return NuiTree.Node[]
local function build_nui_nodes(generic_nodes)
  local nodes = {}
  
  for _, gnode in ipairs(generic_nodes) do
    local children = {}
    
    -- Recursively build children
    if gnode.children and #gnode.children > 0 then
      children = build_nui_nodes(gnode.children)
    end
    
    -- Create NuiTree.Node with all properties from generic node
    local nui_node = NuiTree.Node({
      id = gnode.id,
      type = gnode.type,
      name = gnode.name,
      icon = gnode.icon,
      highlight = gnode.highlight,
      data = gnode.data,
      -- Copy any extra properties
      is_default = gnode.is_default,
      is_system = gnode.is_system,
      note_id = gnode.note_id,
      collection_id = gnode.collection_id,
      server_name = gnode.server_name,
      server_url = gnode.server_url,
    }, children)
    
    table.insert(nodes, nui_node)
  end
  
  return nodes
end

--- Prepare a node for rendering (generic renderer)
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

  -- Icon (use provided icon or empty space)
  if node.icon then
    line:append(node.icon .. " ", node.highlight or "Normal")
  end
  
  -- Name with highlight
  line:append(node.name, node.highlight or "Normal")
  
  -- Special suffix for default items
  if node.is_default then
    line:append(" ", "Comment")
    line:append("(default)", "Comment")
  end

  return line
end

--- Get all expanded node IDs from current tree
---@return string[] Array of node IDs that are expanded
function M.get_expanded_nodes()
  local node_ids = {}
  
  if not state.tree then
    return node_ids
  end
  
  local function collect_expanded(node)
    if node:is_expanded() then
      table.insert(node_ids, node:get_id())
    end
    if node:has_children() then
      for _, child in ipairs(state.tree:get_nodes(node:get_id())) do
        collect_expanded(child)
      end
    end
  end
  
  -- Walk all root nodes
  for _, node in ipairs(state.tree:get_nodes()) do
    collect_expanded(node)
  end
  
  return node_ids
end

--- Restore expanded state for nodes by ID
---@param node_ids string[] Array of node IDs to expand
function M.set_expanded_nodes(node_ids)
  if not state.tree then
    return
  end
  
  for _, id in ipairs(node_ids) do
    local node = state.tree:get_node(id)
    if node then
      node:expand()
    end
  end
end

--- Get the node ID at cursor position
---@return string|nil Node ID at cursor, or nil
function M.get_cursor_node_id()
  if not state.tree then
    return nil
  end
  
  -- Check if the tree's window is still valid before trying to get cursor position
  local winid = vim.fn.win_findbuf(state.tree.bufnr)[1]
  if not winid or not vim.api.nvim_win_is_valid(winid) then
    return nil
  end
  
  local node = state.tree:get_node()
  return node and node:get_id() or nil
end

--- Move cursor to a specific node by ID
---@param node_id string Node ID to focus
function M.set_cursor_to_node(node_id)
  if not state.tree or not node_id then
    return
  end
  
  local node, start_lnum, end_lnum = state.tree:get_node(node_id)
  if node and start_lnum then
    -- Get the window containing the tree buffer
    local winid = vim.fn.win_findbuf(state.tree.bufnr)[1]
    if winid then
      vim.api.nvim_win_set_cursor(winid, { start_lnum, 0 })
    end
  end
end

--- Build and render tree from generic node structure
---@param bufnr number Buffer to render tree in
---@param generic_nodes GenericTreeNode[] Array of pre-built generic nodes
function M.build_and_render(bufnr, generic_nodes)
  -- Convert generic nodes to NuiTree nodes
  local nui_nodes = build_nui_nodes(generic_nodes)
  
  -- Check if we need to create a new tree instance
  -- Create new tree if:
  --   1. No tree exists yet (first time)
  --   2. Buffer changed (after close/reopen - buffer was destroyed)
  local needs_new_tree = not state.tree or (state.tree.bufnr ~= bufnr)
  
  if needs_new_tree then
    -- Create new NuiTree instance
    -- NuiTree will manage modifiable/readonly during render
    state.tree = NuiTree({
      bufnr = bufnr,
      nodes = nui_nodes,
      prepare_node = prepare_node,
    })
  else
    -- Reuse existing tree, just update nodes
    -- This preserves tree's internal state (prev_linenr for ghost prevention)
    state.tree:set_nodes(nui_nodes)
  end
  
  -- Render the tree
  state.tree:render()
end

--- Build and render tree while preserving expanded state and cursor position
---@param bufnr number Buffer to render tree in
---@param generic_nodes GenericTreeNode[] Array of pre-built generic nodes
function M.build_and_render_with_state(bufnr, generic_nodes)
  -- Capture current state before rebuilding
  local expanded = M.get_expanded_nodes()
  local cursor_node_id = M.get_cursor_node_id()
  
  -- Rebuild tree (this replaces the old tree completely)
  M.build_and_render(bufnr, generic_nodes)
  
  -- Restore expanded state
  M.set_expanded_nodes(expanded)
  
  -- Restore cursor position (gracefully handles if node no longer exists)
  if cursor_node_id then
    M.set_cursor_to_node(cursor_node_id)
  end
  
  -- Re-render to reflect restored state
  M.render()
end



--- Get current tree instance
---@return NuiTree|nil
function M.get_tree()
  return state.tree
end

--- Get node at cursor in the current tree
---@param node_id? string Optional node ID (defaults to node under cursor)
---@return NuiTree.Node|nil
function M.get_node(node_id)
  if not state.tree then
    return nil
  end
  return state.tree:get_node(node_id)
end

--- Render the current tree
function M.render()
  if state.tree then
    state.tree:render()
  end
end

return M
