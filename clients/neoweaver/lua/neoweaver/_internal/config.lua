---
--- config.lua - Configuration management for neoweaver
--- Internal module - not part of public API
---
local M = {}

M.defaults = {
  allow_multiple_empty_notes = false,
  metadata = {
    enabled = false, -- EXPERIMENTAL: Enable metadata extraction from .weaverc.json
  },
  explorer = {
    show_notifications = true, -- Show notifications on explorer refresh
  },
  quicknotes = {
    title_template = "%Y%m%d%H%M",
    collection_id = 2,
    note_type_id = 2,
    -- Note: Quicknote payload configuration - See issue #46
    popup = {
      relative = "editor",
      position = "50%",
      size = {
        width = "40%",
        height = "20%",
      },
      border = {
        style = "rounded",
        text = {
          top = "Quick Note",
          top_align = "center",
        },
      },
    },
  },

  picker = {
    size = {
      width = 60,
      height = 20,
    },
    position = "50%",
    border = {
      style = "rounded",
    },
  },

  keymaps = {

    enabled = false, -- Keymaps are opt-in
    notes = {
      -- Standard notes (using <leader>n* for "notes")
      list = "<leader>nl",
      find = "<leader>nf", -- Find notes by title (search picker)
      open = "<leader>no",
      edit = "<leader>ne", -- Alias for open
      new = "<leader>nn",
      new_with_title = "<leader>nN",
      title = "<leader>nt",
      delete = "<leader>nd",
      meta = "<leader>nm", -- Note: Not implemented - See issue #44
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

  if opts.quicknotes ~= nil then
    M.current.quicknotes = vim.tbl_deep_extend("force", M.current.quicknotes, opts.quicknotes)
  end

  -- Merge metadata configuration
  if opts.metadata ~= nil then
    M.current.metadata = vim.tbl_extend("force", M.current.metadata, opts.metadata)
  end

  -- Merge picker configuration
  if opts.picker ~= nil then
    M.current.picker = vim.tbl_deep_extend("force", M.current.picker, opts.picker)
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
