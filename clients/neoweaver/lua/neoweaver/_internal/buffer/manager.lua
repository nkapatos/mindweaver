---
--- buffer/manager.lua - Generic buffer management for Neoweaver
--- Provides buffer lifecycle management decoupled from domain logic
---
--- Purpose:
---   - Create and manage buffers for different entity types (notes, collections, etc.)
---   - Track buffer state with bidirectional lookup (entity → buffer, buffer → entity)
---   - Handle buffer lifecycle events (save, close)
---   - Provide reusable buffer infrastructure for all domain modules
---
--- Design Principles:
---   - Domain-agnostic: no notes-specific logic
---   - Type-based handlers: register once per entity type
---   - Clean separation: buffer lifecycle vs business logic
---   - Flexible storage: modules control their own buffer variables
---
--- Reference: clients/mw/notes.lua (lines 42-108, buffer creation and state)
---
local M = {}

---@class BufferEntity
---@field type string Entity type (e.g., "note", "collection")
---@field id any Entity identifier
---@field data? table Optional entity data

---@class BufferHandlers
---@field on_save? fun(bufnr: integer, id: any) Called when buffer is saved
---@field on_close? fun(bufnr: integer, id: any) Called when buffer is closed

---@class BufferOptions
---@field type string Entity type (e.g., "note", "collection")
---@field id any Entity identifier (number, string, etc.)
---@field name? string Buffer name (optional, will use "Unnamed" if not provided)
---@field filetype? string Buffer filetype (default: "markdown")
---@field modifiable? boolean Whether buffer is modifiable (default: true)
---@field buflisted? boolean Whether buffer appears in buffer list (default: true)
---@field bufhidden? string Buffer hide behavior (default: "wipe")
---@field data? table Initial entity data to store
---@field win? number Target window ID to display buffer in (optional, will find suitable window if not provided)

-- State management
M.state = {
  -- Forward lookup: bufnr → entity metadata
  ---@type table<integer, BufferEntity>
  buffers = {},

  -- Reverse lookup: "type:id" → bufnr
  ---@type table<string, integer>
  index = {},

  -- Type handlers: type → event callbacks
  ---@type table<string, BufferHandlers>
  handlers = {},
}

---Generate index key from type and id
---@param type string
---@param id any
---@return string
local function make_key(type, id)
  return type .. ":" .. tostring(id)
end

---Find a suitable window to display a buffer in
---Excludes special buffers like explorer, terminal, etc.
---@return number|nil window_id Window ID to use, or nil to use current
local function find_target_window()
  local current_win = vim.api.nvim_get_current_win()
  local current_buf = vim.api.nvim_win_get_buf(current_win)
  local current_buftype = vim.api.nvim_get_option_value("buftype", { buf = current_buf })
  local current_filetype = vim.api.nvim_get_option_value("filetype", { buf = current_buf })
  
  -- If current window is a normal file buffer (not special UI), use it
  if current_buftype == "" and current_filetype ~= "neoweaver_explorer" then
    return current_win
  end
  
  -- Otherwise, find a suitable window in the current tab
  local wins = vim.api.nvim_tabpage_list_wins(0)
  
  -- Sort by last used buffer (most recent first)
  table.sort(wins, function(a, b)
    local ba = vim.api.nvim_win_get_buf(a)
    local bb = vim.api.nvim_win_get_buf(b)
    local info_a = vim.fn.getbufinfo(ba)[1]
    local info_b = vim.fn.getbufinfo(bb)[1]
    return (info_a and info_a.lastused or 0) > (info_b and info_b.lastused or 0)
  end)
  
  -- Find first suitable window
  for _, win in ipairs(wins) do
    if vim.api.nvim_win_is_valid(win) then
      local buf = vim.api.nvim_win_get_buf(win)
      local buftype = vim.api.nvim_get_option_value("buftype", { buf = buf })
      local filetype = vim.api.nvim_get_option_value("filetype", { buf = buf })
      local win_config = vim.api.nvim_win_get_config(win)
      local is_float = win_config.relative ~= ""
      
      -- Skip floating windows, special buffers, and explorer
      if not is_float 
         and buftype == "" 
         and filetype ~= "neoweaver_explorer" then
        return win
      end
    end
  end
  
  -- No suitable window found - create a new split
  -- Find the explorer window to split from
  for _, win in ipairs(wins) do
    local buf = vim.api.nvim_win_get_buf(win)
    local filetype = vim.api.nvim_get_option_value("filetype", { buf = buf })
    if filetype == "neoweaver_explorer" then
      -- Focus explorer and create vertical split to the right
      vim.api.nvim_set_current_win(win)
      vim.cmd("vsplit")
      local new_win = vim.api.nvim_get_current_win()
      return new_win
    end
  end
  
  -- Fallback: use current window (will create split if needed)
  return nil
end

---Register event handlers for a buffer type
---Handlers are called when buffers of this type trigger events
---@param type string Entity type (e.g., "note", "collection")
---@param handlers BufferHandlers Event handlers { on_save, on_close }
function M.register_type(type, handlers)
  M.state.handlers[type] = handlers
end

---Switch to a buffer in a suitable window
---Intelligently finds or creates a target window (not explorer, terminal, etc.)
---@param bufnr integer Buffer number to switch to
function M.switch_to_buffer(bufnr)
  if not vim.api.nvim_buf_is_valid(bufnr) then
    return
  end
  
  local target_win = find_target_window()
  if target_win and vim.api.nvim_win_is_valid(target_win) then
    vim.api.nvim_set_current_win(target_win)
  end
  vim.api.nvim_set_current_buf(bufnr)
end

---Create a new managed buffer for an entity
---@param opts BufferOptions Buffer creation options
---@return integer bufnr The created buffer number
function M.create(opts)
  -- Validate required fields
  if not opts.type or not opts.id then
    error("buffer_manager.create: type and id are required")
  end

  -- Determine target window
  local target_win = opts.win or find_target_window()
  
  -- Check if buffer already exists
  local existing = M.get(opts.type, opts.id)
  if existing and vim.api.nvim_buf_is_valid(existing) then
    -- Buffer already exists, switch to target window and set buffer
    if target_win and vim.api.nvim_win_is_valid(target_win) then
      vim.api.nvim_set_current_win(target_win)
    end
    vim.api.nvim_set_current_buf(existing)
    return existing
  end

  -- Create buffer
  local name = opts.name or "Unnamed"
  local bufnr = vim.fn.bufnr(name, true)

  -- Set buffer options
  local buf_opts = {
    filetype = opts.filetype or "markdown",
    buflisted = opts.buflisted ~= false, -- default true
    modifiable = opts.modifiable ~= false, -- default true
    bufhidden = opts.bufhidden or "wipe",
  }

  for opt, value in pairs(buf_opts) do
    vim.api.nvim_set_option_value(opt, value, { buf = bufnr })
  end

  -- Track buffer in state
  local key = make_key(opts.type, opts.id)
  M.state.buffers[bufnr] = {
    type = opts.type,
    id = opts.id,
    data = opts.data or {},
  }
  M.state.index[key] = bufnr

  -- Register BufWriteCmd if save handler exists
  local handlers = M.state.handlers[opts.type]
  if handlers and handlers.on_save then
    local group = vim.api.nvim_create_augroup("NwBufWrite_" .. bufnr, { clear = true })
    vim.api.nvim_create_autocmd("BufWriteCmd", {
      group = group,
      buffer = bufnr,
      callback = function()
        handlers.on_save(bufnr, opts.id)
      end,
    })
  end

  -- Register BufWipeout to cleanup state
  local cleanup_group = vim.api.nvim_create_augroup("NwBufCleanup_" .. bufnr, { clear = true })
  vim.api.nvim_create_autocmd("BufWipeout", {
    group = cleanup_group,
    buffer = bufnr,
    callback = function()
      -- Call on_close handler if exists
      if handlers and handlers.on_close then
        handlers.on_close(bufnr, opts.id)
      end

      -- Cleanup state
      local entity = M.state.buffers[bufnr]
      if entity then
        local cleanup_key = make_key(entity.type, entity.id)
        M.state.index[cleanup_key] = nil
        M.state.buffers[bufnr] = nil
      end
    end,
  })

  -- Switch to target window and display buffer
  if target_win and vim.api.nvim_win_is_valid(target_win) then
    vim.api.nvim_set_current_win(target_win)
  end
  vim.api.nvim_set_current_buf(bufnr)
  
  -- Setup statusline for the buffer in current window
  local statusline = require("neoweaver._internal.buffer.statusline")
  local current_win = vim.api.nvim_get_current_win()
  statusline.setup(bufnr, current_win)

  return bufnr
end

---Get buffer number for an entity
---@param type string Entity type
---@param id any Entity identifier
---@return integer|nil bufnr Buffer number or nil if not found
function M.get(type, id)
  local key = make_key(type, id)
  return M.state.index[key]
end

---Get entity information from buffer number (reverse lookup)
---@param bufnr integer Buffer number
---@return BufferEntity|nil entity Entity metadata or nil if not managed
function M.get_entity(bufnr)
  return M.state.buffers[bufnr]
end

---Check if a buffer exists for an entity
---@param type string Entity type
---@param id any Entity identifier
---@return boolean exists True if buffer exists and is valid
function M.exists(type, id)
  local bufnr = M.get(type, id)
  return bufnr ~= nil and vim.api.nvim_buf_is_valid(bufnr)
end

---Check if a buffer is managed by buffer_manager
---@param bufnr integer Buffer number
---@return boolean managed True if buffer is managed
function M.is_managed(bufnr)
  return M.state.buffers[bufnr] ~= nil
end

---List all managed buffers, optionally filtered by type
---@param type? string Optional entity type filter
---@return table<integer, BufferEntity> buffers Map of bufnr → entity
function M.list(type)
  if not type then
    return M.state.buffers
  end

  local filtered = {}
  for bufnr, entity in pairs(M.state.buffers) do
    if entity.type == type then
      filtered[bufnr] = entity
    end
  end
  return filtered
end

---Close a managed buffer
---Calls on_close handler and removes from state
---@param bufnr integer Buffer number
function M.close(bufnr)
  if not M.is_managed(bufnr) then
    return
  end

  -- Buffer wipeout will trigger cleanup via autocmd
  if vim.api.nvim_buf_is_valid(bufnr) then
    vim.api.nvim_buf_delete(bufnr, { force = true })
  end
end

return M
