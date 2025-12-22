---
--- quicknote.lua - Ephemeral quicknote capture
---
--- Quicknotes are short-lived floating buffers used for rapid capture.
--- They create notes server-side only when content exists on close.
---
local Popup = require("nui.popup")
local event = require("nui.utils.autocmd").event

local api = require("neoweaver._internal.api")
local config_module = require("neoweaver._internal.config")

local M = {}

local state = {
  popup = nil,
  closing = false,
  saved_buffers = {},
}

local function get_config()
  local cfg = config_module.get().quicknotes or {}

  return vim.tbl_deep_extend("force", {
    title_template = "%Y%m%d%H%M",
    collection_id = 2,
    --- TODO: make quicknote note type configurable via user preferences
    note_type_id = 2,
    popup = {
      relative = "editor",
      position = "50%",
      size = {
        width = "40%",
        height = "20%",
      },
      border = {
        style = "rounded",
        padding = {
          top = 1,
          bottom = 1,
          left = 2,
          right = 2,
        },
        text = {
          top = "Quick Note",
          top_align = "center",
        },
      },
      buf_options = {},
      win_options = {},
    },
  }, cfg)
end

local function build_popup_options(cfg)
  local options = vim.deepcopy(cfg.popup or {})

  options.enter = true
  options.focusable = true
  options.relative = options.relative or "editor"
  options.position = options.position or "50%"
  options.size = options.size or { width = "40%", height = "20%" }

  options.border = vim.tbl_deep_extend("force", {
    style = "rounded",
    padding = {
      top = 1,
      bottom = 1,
      left = 2,
      right = 2,
    },
    text = {
      top = "Quick Note",
      top_align = "center",
    },
  }, options.border or {})

  options.buf_options = vim.tbl_extend("force", {
    buftype = "nofile",
    bufhidden = "hide",
    swapfile = false,
    filetype = "markdown",
    modifiable = true,
    buflisted = false,
  }, options.buf_options or {})

  options.win_options = vim.tbl_extend("force", {}, options.win_options or {})

  return options
end

local function save_if_needed(bufnr)
  if not vim.api.nvim_buf_is_valid(bufnr) then
    return
  end

  if vim.b[bufnr].neoweaver_quicknote_saved then
    return
  end

  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local body = table.concat(lines, "\n")

  if vim.trim(body) == "" then
    return
  end

  vim.b[bufnr].neoweaver_quicknote_saved = true

  local cfg = get_config()
  local title = vim.fn.strftime(cfg.title_template or "%Y%m%d%H%M")

  ---@type mind.v3.CreateNoteRequest
  local payload = {
    title = title,
    body = body,
    collectionId = cfg.collection_id,
    noteTypeId = cfg.note_type_id,
  }

  -- TODO: allow configuring quicknote metadata/payload enrichment per editor prefs
  api.notes.create(payload, function(res)
    if res and res.error then
      vim.notify("Quicknote save failed: " .. (res.error.message or vim.inspect(res.error)), vim.log.levels.ERROR)
      return
    end

    if state.popup and state.popup.border then
      state.popup.border:set_text("bottom", "Saved as " .. title, "center")
    end

    vim.notify(string.format("Quicknote saved as %s", title), vim.log.levels.INFO)
  end)
end

local function close_popup(popup)
  popup = popup or state.popup
  if not popup or state.closing then
    return
  end

  state.closing = true

  local bufnr = popup.bufnr
  save_if_needed(bufnr)

  popup:unmount()
  if state.popup == popup then
    state.popup = nil
  end

  state.closing = false
end

local function create_popup()
  local cfg = get_config()
  local popup = Popup(build_popup_options(cfg))

  popup:map("n", "q", function()
    close_popup(popup)
  end, { nowait = true, noremap = true, desc = "Close quicknote" })

  popup:map("n", "<Esc>", function()
    close_popup(popup)
  end, { nowait = true, noremap = true, desc = "Close quicknote" })

  popup:on(event.BufWinLeave, function()
    save_if_needed(popup.bufnr)
  end)

  return popup
end

function M.open()
  if state.popup then
    close_popup(state.popup)
  end

  local cfg = get_config()
  local title = vim.fn.strftime(cfg.title_template or "%Y%m%d%H%M")

  local popup = create_popup()
  popup:mount()

  -- Update the title to show the generated note title
  popup.border:set_text("top", " " .. title .. " ", "center")

  vim.api.nvim_buf_set_name(popup.bufnr, "Quicknote")
  vim.api.nvim_buf_set_lines(popup.bufnr, 0, -1, false, {})
  vim.b[popup.bufnr].neoweaver_quicknote_saved = nil
  popup.border:set_text("bottom", "", "center")

  state.popup = popup

  -- Enter insert mode
  vim.cmd("startinsert")
end

function M.amend()
  vim.notify("TODO: Quicknote amend not yet implemented", vim.log.levels.INFO)
end

function M.list()
  vim.notify("TODO: Quicknote listing not yet implemented", vim.log.levels.INFO)
end

return M
