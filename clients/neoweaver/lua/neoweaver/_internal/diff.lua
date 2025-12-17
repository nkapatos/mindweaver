---
--- diff.lua - Diff view for conflict resolution (placeholder)
---
--- This is a placeholder module for future diff functionality.
--- When a 412 Precondition Failed error occurs (etag conflict),
--- this module will display the server version vs local version
--- and allow resolution via diff hunks.
---
--- Reference: clients/mw/diff.lua (full implementation)
---
local M = {}

--- Setup highlight groups for diff overlays
function M.setup()
  -- TODO: Implement highlight group setup
  -- Reference: mw/diff.lua:setup_highlight()
end

--- Enable diff overlay for a buffer
---@param bufnr integer Buffer number
function M.enable(bufnr)
  vim.notify(
    string.format("Diff view enabled for buffer %d (placeholder - not yet implemented)", bufnr),
    vim.log.levels.INFO
  )
  -- TODO: Implement diff overlay enabling
  -- Reference: mw/diff.lua:enable()
end

--- Disable diff overlay for a buffer
---@param bufnr integer Buffer number
function M.disable(bufnr)
  vim.notify(
    string.format("Diff view disabled for buffer %d (placeholder)", bufnr),
    vim.log.levels.INFO
  )
  -- TODO: Implement diff overlay cleanup
  -- Reference: mw/diff.lua:disable()
end

--- Set reference text (server version) for diffing
---@param bufnr integer Buffer number
---@param ref_lines string[] Array of lines from server
function M.set_ref_text(bufnr, ref_lines)
  vim.notify(
    string.format(
      "Diff ref text set for buffer %d (%d lines from server) - placeholder",
      bufnr,
      #ref_lines
    ),
    vim.log.levels.INFO
  )
  -- TODO: Implement reference text storage and diff computation
  -- Reference: mw/diff.lua:set_ref_text()
end

--- Map diff navigation keys for a buffer
---@param bufnr integer Buffer number
function M.map_keys(bufnr)
  vim.notify(
    string.format("Diff keymaps set for buffer %d ([c/]c/gh) - placeholder", bufnr),
    vim.log.levels.INFO
  )
  -- TODO: Implement keymap setup
  -- Reference: mw/diff.lua:map_keys()
  -- Keys: ]c (next hunk), [c (prev hunk), gh (apply hunk)
end

--- Check if buffer has unresolved diff hunks
---@param _ integer Buffer number (unused in placeholder)
---@return boolean has_hunks True if unresolved hunks exist
function M.has_unresolved_hunks(_)
  -- TODO: Implement hunk checking
  -- Reference: mw/diff.lua:has_unresolved_hunks()
  return false
end

return M
