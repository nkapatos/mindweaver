--- Explorer module for neoweaver
--- Provides a tree view of collections and notes
---
local M = {}

local window = require("neoweaver._internal.explorer.window")
local tree = require("neoweaver._internal.explorer.tree")

--- Open the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.open(opts)
  local split = window.open(opts)
  
  -- Initialize and render tree if we just opened the window
  if split and split.bufnr then
    tree.init(split.bufnr)
  end
end

--- Close the explorer sidebar
function M.close()
  window.close()
end

--- Toggle the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.toggle(opts)
  if window.is_open() then
    window.close()
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
