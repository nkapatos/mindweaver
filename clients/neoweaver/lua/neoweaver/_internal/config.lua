---
--- config.lua - Configuration management for neoweaver
--- Internal module - not part of public API
---
local M = {}

M.defaults = {
  allow_multiple_empty_notes = false,
  explorer = {
    show_notifications = true, -- Show notifications on explorer refresh
  },
  keymaps = {
    enabled = false, -- Keymaps are opt-in
    notes = {
      -- Standard notes (using <leader>n* for "notes")
      list = "<leader>nl",
      open = "<leader>no",
      edit = "<leader>ne", -- Alias for open
      new = "<leader>nn",
      new_with_title = "<leader>nN",
      title = "<leader>nt",
      delete = "<leader>nd",
      meta = "<leader>nm", -- TODO: not implemented
    },
    quicknotes = {
      -- Quicknotes (using <leader>q* for "quick")
      new = "<leader>qn",
      list = "<leader>ql",
      amend = "<leader>qa",
      -- Fast access alternatives (using <leader>.* for rapid capture)
      new_fast = "<leader>.n",
      amend_fast = "<leader>.a",
      list_fast = "<leader>.l",
    },
  },
}

M.current = vim.deepcopy(M.defaults)

--- Apply user configuration options
---@param opts table User configuration options
function M.apply(opts)
  opts = opts or {}

  if opts.allow_multiple_empty_notes ~= nil then
    M.current.allow_multiple_empty_notes = opts.allow_multiple_empty_notes == true
  end

  -- Merge explorer configuration
  if opts.explorer ~= nil then
    M.current.explorer = vim.tbl_extend("force", M.current.explorer, opts.explorer)
  end

  -- Merge keymap configuration
  if opts.keymaps ~= nil then
    if opts.keymaps.enabled ~= nil then
      M.current.keymaps.enabled = opts.keymaps.enabled
    end
    if opts.keymaps.notes ~= nil then
      M.current.keymaps.notes = vim.tbl_extend("force", M.current.keymaps.notes, opts.keymaps.notes)
    end
    if opts.keymaps.quicknotes ~= nil then
      M.current.keymaps.quicknotes = vim.tbl_extend("force", M.current.keymaps.quicknotes, opts.keymaps.quicknotes)
    end
  end
end

--- Get current configuration
---@return table Current configuration
function M.get()
  return M.current
end

return M
