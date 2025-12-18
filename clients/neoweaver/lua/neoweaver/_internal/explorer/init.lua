--- Explorer module for neoweaver
--- Provides a tree view of collections and notes
---
local M = {}

local window = require("neoweaver._internal.explorer.window")

--- Open the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.open(opts)
  window.open(opts)
  -- TODO: Initialize tree and render
end

--- Close the explorer sidebar
function M.close()
  window.close()
end

--- Toggle the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.toggle(opts)
  window.toggle(opts)
  -- TODO: Initialize tree if opening
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
