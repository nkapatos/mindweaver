---@class ExplorerStatus
---@field state "idle" | "loading" | "error"
---@field stats { collections: integer, notes: integer }
---@field error_msg string?

---@type ExplorerStatus
local status = {
  state = "idle",
  stats = { collections = 0, notes = 0 },
  error_msg = nil,
}

local M = {}

--- Set status to loading state
function M.set_loading()
  status.state = "loading"
  status.error_msg = nil
  vim.cmd("redrawstatus")
end

--- Set status to ready state with statistics
---@param collections_count integer Number of collections
---@param notes_count integer Number of notes
function M.set_ready(collections_count, notes_count)
  status.state = "idle"
  status.stats.collections = collections_count
  status.stats.notes = notes_count
  status.error_msg = nil
  vim.cmd("redrawstatus")
end

--- Set status to error state
---@param msg string Error message
function M.set_error(msg)
  status.state = "error"
  status.error_msg = msg
  vim.cmd("redrawstatus")
end

--- Get formatted status string for statusline
---@return string
function M.get_status()
  if status.state == "loading" then
    return "⟳ Refreshing..."
  elseif status.state == "error" then
    return "✗ Error: " .. (status.error_msg or "Unknown error")
  else
    -- idle state
    return string.format(
      "📁 Collections: %d | 📄 Notes: %d",
      status.stats.collections,
      status.stats.notes
    )
  end
end

return M
