---@class BufferStatusline
---Buffer statusline management for entity buffers
---Displays server context and entity information

local api = require("neoweaver._internal.api")
local buffer_manager = require("neoweaver._internal.buffer.manager")

local M = {}

--- Get formatted statusline for an entity buffer
---@param bufnr integer Buffer number
---@return string
function M.get_status(bufnr)
  -- Get entity info
  local entity = buffer_manager.get_entity(bufnr)
  if not entity then
    return "Neoweaver"
  end
  
  -- Get server info
  local server_name = api.config.current_server or "No Server"
  local server_url = ""
  if api.config.current_server and api.config.servers[api.config.current_server] then
    server_url = api.config.servers[api.config.current_server].url
  end
  
  -- Build status line
  local parts = {}
  
  -- Server context
  table.insert(parts, string.format("󰒋 %s", server_name))
  if server_url ~= "" then
    table.insert(parts, string.format("(%s)", server_url))
  end
  
  -- Entity info
  table.insert(parts, "|")
  if entity.type == "note" then
    table.insert(parts, string.format("󰈙 Note #%s", tostring(entity.id)))
  else
    table.insert(parts, string.format("%s #%s", entity.type, tostring(entity.id)))
  end
  
  return table.concat(parts, " ")
end

--- Setup statusline for a buffer window
---@param bufnr integer Buffer number
---@param winid integer Window ID
function M.setup(bufnr, winid)
  -- Use window-local statusline with Lua callback
  vim.wo[winid].statusline = string.format(
    "%%{%%v:lua.require'neoweaver._internal.buffer.statusline'.get_status(%d)%%}",
    bufnr
  )
end

return M
