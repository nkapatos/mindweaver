---
--- explorer.lua - Note browsing and selection UI
--- Main entry point for explorer functionality - delegates to picker module
---
--- Architecture Decision:
--- - Uses vim.ui.select which respects user's picker plugin (Snacks/Telescope/fzf)
--- - Backend returns consistent schema (notes and quicknotes use same model)
--- - This module delegates to explorer/picker.lua for actual UI
---
--- Recent changes (2025-11-24):
--- - Integrated with explorer/picker.lua module
--- - Now uses real API data for listing
--- - Respects user's vim.ui.select configuration
---
--- Next steps:
--- - Add filtering options (tags, metadata)
--- - Add project switching
--- - Add search/FTS integration
---
local picker = require("mw.explorer.picker")

local M = {}

---Explorer configuration
---@class ExplorerConfig
M.config = {}

---Setup explorer configuration
---@param opts ExplorerConfig? Configuration options to override defaults
function M.setup(opts)
	M.config = vim.tbl_deep_extend("force", M.config, opts or {})
end

---Show note picker - delegates to picker module
---This is the main entry point for showing the note explorer
---@param notes table[] List of notes from API
---@param opts table Options (project, note_type, on_select)
function M.show_picker(notes, opts)
	return picker.show_picker(notes, opts)
end

---Legacy format_list for backward compatibility
---@deprecated Use picker module directly
---@param items table List of items from API
---@return table Items (passed through)
function M.format_list(items)
	-- Just pass through - picker handles formatting now
	return items
end

---Placeholder for pagination support in future
---@param items table List of items
---@param page integer Page number
---@param page_size integer Items per page
---@return table Items (passed through for now)
function M.paginate(items, page, page_size)
	-- Not implemented yet
	return items
end

return M
