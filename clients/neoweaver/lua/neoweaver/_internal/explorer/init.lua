--- Explorer module for neoweaver
--- Orchestrates tree display and domain actions (collections, notes, tags)
---
--- Simple approach following neo-tree pattern:
---   - NuiSplit manages window + buffer lifecycle
---   - Explorer coordinates between window, tree, and domain modules
---   - Keymaps rebound on each open (fast, not a problem)
---
local M = {}

local window = require("neoweaver._internal.explorer.window")
local tree = require("neoweaver._internal.explorer.tree")
local collections = require("neoweaver._internal.collections")
local notes = require("neoweaver._internal.notes")
local statusline = require("neoweaver._internal.explorer.statusline")
local config = require("neoweaver._internal.config")

---@class ExplorerState
---@field current_mode string Current display mode ("collections", "tags", etc.)

---@type ExplorerState
local state = {
  current_mode = "collections", -- Default mode
}

--- Mode registry - defines available modes and their behaviors
---@type table<string, { load_data: fun(callback: fun(nodes, err, stats)), handle_create: fun(node, refresh_callback), handle_rename: fun(node, refresh_callback), handle_delete: fun(node, refresh_callback) }>
local modes = {
  collections = {
    load_data = collections.build_tree_nodes,
    handle_create = collections.handle_create,
    handle_rename = collections.handle_rename,
    handle_delete = collections.handle_delete,
  },
  -- Future modes can be added here:
  -- tags = { ... },
  -- search = { ... },
}

--- Setup keymaps for the explorer buffer
---@param bufnr number Buffer number
local function setup_keymaps(bufnr)
  local map_opts = { noremap = true, nowait = true, buffer = bufnr }
  
  -- Open/expand/collapse (Enter)
  vim.keymap.set("n", "<CR>", function()
    local node = tree.get_node()
    if not node then return end
    
    -- Handle note nodes - open for editing
    if node.type == "note" then
      notes.open_note(node.note_id)
      return
    end
    
    -- Handle server/collection nodes - expand/collapse
    if node:has_children() then
      if node:is_expanded() then
        node:collapse()
      else
        node:expand()
      end
      tree.render()
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Open note or toggle node" }))
  
  -- Open note (o)
  vim.keymap.set("n", "o", function()
    local node = tree.get_node()
    if node and node.type == "note" then
      notes.open_note(node.note_id)
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Open note" }))
  
  -- Expand (l)
  vim.keymap.set("n", "l", function()
    local node = tree.get_node()
    if node and node:has_children() and not node:is_expanded() then
      node:expand()
      tree.render()
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Expand node" }))
  
  -- Collapse or go to parent (h)
  vim.keymap.set("n", "h", function()
    local node = tree.get_node()
    if not node then return end
    
    if node:is_expanded() and node:has_children() then
      node:collapse()
      tree.render()
    else
      -- Navigate to parent
      local parent = node:get_parent_id()
      if parent then
        local parent_node = tree.get_node(parent)
        if parent_node then
          local tree_instance = tree.get_tree()
          if tree_instance then
            tree_instance:set_node(parent)
            tree.render()
          end
        end
      end
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Collapse or go to parent" }))
  
  -- Navigation (j/k)
  vim.keymap.set("n", "j", function()
    vim.cmd("normal! j")
  end, vim.tbl_extend("force", map_opts, { desc = "Move down" }))
  
  vim.keymap.set("n", "k", function()
    vim.cmd("normal! k")
  end, vim.tbl_extend("force", map_opts, { desc = "Move up" }))
  
  -- Refresh (R)
  vim.keymap.set("n", "R", function()
    M.refresh()
  end, vim.tbl_extend("force", map_opts, { desc = "Refresh tree" }))
  
  -- Generic actions - delegate to mode-specific handlers
  
  -- Create (a)
  vim.keymap.set("n", "a", function()
    local node = tree.get_node()
    if node then
      M.handle_create(node)
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Create item" }))
  
  -- Rename (r)
  vim.keymap.set("n", "r", function()
    local node = tree.get_node()
    if node then
      M.handle_rename(node)
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Rename item" }))
  
  -- Delete (d)
  vim.keymap.set("n", "d", function()
    local node = tree.get_node()
    if node then
      M.handle_delete(node)
    end
  end, vim.tbl_extend("force", map_opts, { desc = "Delete item" }))
end

--- Handle create action (delegates to current mode)
---@param node table Tree node
function M.handle_create(node)
  local mode = modes[state.current_mode]
  if mode and mode.handle_create then
    mode.handle_create(node, M.refresh)
  else
    vim.notify("Create action not supported in " .. state.current_mode .. " mode", vim.log.levels.WARN)
  end
end

--- Handle rename action (delegates to current mode)
---@param node table Tree node
function M.handle_rename(node)
  local mode = modes[state.current_mode]
  if mode and mode.handle_rename then
    mode.handle_rename(node, M.refresh)
  else
    vim.notify("Rename action not supported in " .. state.current_mode .. " mode", vim.log.levels.WARN)
  end
end

--- Handle delete action (delegates to current mode)
---@param node table Tree node
function M.handle_delete(node)
  local mode = modes[state.current_mode]
  if mode and mode.handle_delete then
    mode.handle_delete(node, M.refresh)
  else
    vim.notify("Delete action not supported in " .. state.current_mode .. " mode", vim.log.levels.WARN)
  end
end

--- Load and render tree data (async) using current mode
---@param show_notification? boolean Show notification when complete
local function load_and_render_tree(show_notification)
  -- Get buffer from window
  local bufnr = window.get_bufnr()
  if not bufnr then
    vim.notify("Explorer window not open", vim.log.levels.ERROR)
    return
  end
  
  local cfg = config.get()
  
  -- Default to config setting if not explicitly specified
  if show_notification == nil then
    show_notification = cfg.explorer.show_notifications
  end
  
  -- Get current mode
  local mode = modes[state.current_mode]
  if not mode or not mode.load_data then
    vim.notify("Invalid mode: " .. state.current_mode, vim.log.levels.ERROR)
    return
  end
  
  -- Set loading state
  statusline.set_loading()
  
  -- Load data using mode-specific loader
  mode.load_data(function(nodes, err, stats)
    if err then
      statusline.set_error(err.message or "Failed to load")
      vim.notify("Failed to load data: " .. vim.inspect(err), vim.log.levels.ERROR)
      return
    end
    
    -- Handle empty result
    if not nodes or #nodes == 0 then
      statusline.set_ready(0, 0)
      vim.notify("No data found", vim.log.levels.INFO)
      return
    end
    
    -- Update statusline with stats
    if stats then
      statusline.set_ready(stats.collections or 0, stats.notes or 0)
    end
    
    -- Build and render tree with state preservation
    tree.build_and_render_with_state(bufnr, nodes)
    
    -- Show notification if enabled
    if show_notification and stats then
      vim.notify(
        string.format("Loaded %d collections, %d notes", stats.collections or 0, stats.notes or 0),
        vim.log.levels.INFO
      )
    end
  end)
end

--- Refresh the tree (reloads data using current mode)
function M.refresh()
  load_and_render_tree(false)
end

--- Switch to a different display mode
---@param mode_name string Name of the mode to switch to ("collections", "tags", etc.)
function M.switch_mode(mode_name)
  if not modes[mode_name] then
    vim.notify("Unknown mode: " .. mode_name, vim.log.levels.ERROR)
    return
  end
  
  state.current_mode = mode_name
  M.refresh()
end

--- Open the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.open(opts)
  -- Open window (NuiSplit creates window + buffer)
  local split = window.open(opts)
  
  if not split then
    vim.notify("Failed to open explorer window", vim.log.levels.ERROR)
    return
  end
  
  -- Setup keymaps on the new buffer
  setup_keymaps(split.bufnr)
  
  -- Load and render tree data
  load_and_render_tree(true)
end

--- Close the explorer sidebar
function M.close()
  window.close()
end

--- Toggle the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.toggle(opts)
  if window.is_open() then
    M.close()  -- Use M.close() to properly clean up tree state
  else
    M.open(opts)
  end
end

--- Focus the explorer window (if open)
function M.focus()
  window.focus()
end

--- Check if explorer is currently open
---@return boolean
function M.is_open()
  return window.is_open()
end

return M
