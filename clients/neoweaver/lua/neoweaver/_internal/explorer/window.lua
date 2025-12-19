--- Window management for the neoweaver explorer
--- Manages the NuiSplit sidebar
--- Simple approach: let NuiSplit manage buffer lifecycle
---
local Split = require("nui.split")

local M = {}

---@class ExplorerWindowState
---@field split NuiSplit|nil
---@field is_open boolean

---@type ExplorerWindowState
local state = {
  split = nil,
  is_open = false,
}

---@class ExplorerWindowConfig
---@field position "left"|"right"
---@field size number

---@type ExplorerWindowConfig
local default_config = {
  position = "left",
  size = 30,
}

--- Create and mount the explorer split
---@param config? ExplorerWindowConfig
---@return NuiSplit|nil
function M.open(config)
  config = vim.tbl_deep_extend("force", default_config, config or {})

  if state.is_open and state.split then
    return state.split
  end

  -- Let NuiSplit create and manage everything
  state.split = Split({
    relative = "editor",
    position = config.position,
    size = config.size,
    buf_options = {
      buftype = "nofile",
      swapfile = false,
      filetype = "neoweaver_explorer",
    },
    win_options = {
      number = false,
      relativenumber = false,
      cursorline = true,
      signcolumn = "no",
      wrap = false,
    },
  })

  state.split:mount()
  state.is_open = true

  -- Set up status line
  vim.wo[state.split.winid].statusline = "%{%v:lua.require'neoweaver._internal.explorer.statusline'.get_status()%}"

  -- Close on q
  state.split:map("n", "q", function()
    M.close()
  end, { noremap = true })

  return state.split
end

--- Close the explorer split
function M.close()
  if state.split then
    state.split:unmount()
    state.split = nil
  end
  state.is_open = false
end

--- Toggle the explorer split
---@param config? ExplorerWindowConfig
function M.toggle(config)
  if state.is_open then
    M.close()
  else
    M.open(config)
  end
end

--- Check if explorer is open
---@return boolean
function M.is_open()
  return state.is_open
end

--- Get the split's buffer number
---@return number|nil
function M.get_bufnr()
  if state.split then
    return state.split.bufnr
  end
  return nil
end

--- Focus the explorer window
function M.focus()
  if state.split and state.is_open then
    vim.api.nvim_set_current_win(state.split.winid)
  end
end

return M
