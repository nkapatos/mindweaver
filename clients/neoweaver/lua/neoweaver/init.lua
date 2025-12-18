--- Neoweaver - Neovim client for MindWeaver
---
--- A plugin for managing notes via the MindWeaver server API.
--- Provides commands for creating, editing, and organizing notes directly
--- within Neovim using markdown buffers.
---
--- Requires:
--- - Neovim 0.11+
--- - plenary.nvim
--- - MindWeaver server running
---
---@module neoweaver
local M = {}

--- Setup the neoweaver plugin
---
--- Configures the API client, note handlers, and optional keymaps.
--- Must be called before using any plugin functionality.
---
---@param opts? table Configuration options
---@field opts.allow_multiple_empty_notes? boolean Allow multiple untitled notes (default: false)
---@field opts.api? table API configuration (servers, debug_info)
---@field opts.api.servers table Server configurations (required)
---@field opts.api.debug_info? boolean Enable debug logging (default: true)
---@field opts.keymaps? table Keymap configuration
---@field opts.keymaps.enabled? boolean Enable default keymaps (default: false)
---
---@usage [[
--- Basic setup with local server:
--- require('neoweaver').setup({
---   api = {
---     servers = {
---       local = { url = "http://localhost:9421", default = true }
---     }
---   }
--- })
---
--- With multiple servers and keymaps:
--- require('neoweaver').setup({
---   allow_multiple_empty_notes = true,
---   api = {
---     servers = {
---       local = { url = "http://localhost:9421", default = true },
---       remote = { url = "https://api.example.com" }
---     },
---     debug_info = true
---   },
---   keymaps = {
---     enabled = true,
---     notes = {
---       list = "<leader>nl",
---       open = "<leader>no",
---       new = "<leader>nn"
---     }
---   }
--- })
---]]
function M.setup(opts)
  opts = opts or {}

  -- Apply configuration
  local config = require("neoweaver._internal.config")
  config.apply(opts)

  -- Setup API layer
  local api = require("neoweaver._internal.api")
  api.setup(opts.api or {})

  -- Setup notes module
  local notes = require("neoweaver._internal.notes")
  notes.setup({
    allow_multiple_empty_notes = config.get().allow_multiple_empty_notes,
  })

  -- Setup keymaps if enabled
  if config.get().keymaps.enabled then
    M.setup_keymaps()
  end

  vim.notify("Neoweaver v3 loaded!", vim.log.levels.INFO)
end

--- Setup keymaps for note operations
--- @private
function M.setup_keymaps()
  local notes = require("neoweaver._internal.notes")
  local config = require("neoweaver._internal.config").get()
  local km_notes = config.keymaps.notes
  local km_quick = config.keymaps.quicknotes

  local function prompt_note_id(prompt, action)
    vim.ui.input({ prompt = prompt }, function(input)
      if input == nil or input == "" then
        return
      end

      local id = tonumber(input)
      if not id then
        vim.notify("Invalid note ID", vim.log.levels.WARN)
        return
      end

      action(id)
    end)
  end

  -- Standard notes keymaps
  if km_notes.list then
    vim.keymap.set("n", km_notes.list, notes.list_notes, { desc = "List notes" })
  end

  if km_notes.open then
    vim.keymap.set("n", km_notes.open, function()
      prompt_note_id("Note ID:", notes.open_note)
    end, { desc = "Open note by ID" })
  end

  if km_notes.edit then
    vim.keymap.set("n", km_notes.edit, function()
      prompt_note_id("Note ID:", notes.open_note)
    end, { desc = "Edit note by ID" })
  end

  if km_notes.title then
    vim.keymap.set("n", km_notes.title, notes.edit_title, { desc = "Edit current note title" })
  end

  if km_notes.new then
    vim.keymap.set("n", km_notes.new, notes.create_note, { desc = "Create new note" })
  end

  if km_notes.new_with_title then
    vim.keymap.set("n", km_notes.new_with_title, notes.create_note_with_title, { desc = "Create note with title" })
  end

  if km_notes.delete then
    vim.keymap.set("n", km_notes.delete, function()
      prompt_note_id("Note ID to delete:", notes.delete_note)
    end, { desc = "Delete note by ID" })
  end

  if km_notes.meta then
    vim.keymap.set("n", km_notes.meta, notes.edit_metadata, { desc = "Edit note metadata (TODO: not implemented)" })
  end

  -- Quicknotes keymaps
  if km_quick.new then
    vim.keymap.set("n", km_quick.new, notes.create_quicknote, { desc = "New quicknote (TODO: not implemented)" })
  end

  if km_quick.list then
    vim.keymap.set("n", km_quick.list, notes.list_quicknotes, { desc = "List quicknotes (TODO: not implemented)" })
  end

  if km_quick.amend then
    vim.keymap.set("n", km_quick.amend, notes.amend_quicknote, { desc = "Amend quicknote (TODO: not implemented)" })
  end

  -- Fast access quicknotes keymaps
  if km_quick.new_fast then
    vim.keymap.set(
      "n",
      km_quick.new_fast,
      notes.create_quicknote,
      { desc = "New quicknote (fast) (TODO: not implemented)" }
    )
  end

  if km_quick.amend_fast then
    vim.keymap.set(
      "n",
      km_quick.amend_fast,
      notes.amend_quicknote,
      { desc = "Amend quicknote (fast) (TODO: not implemented)" }
    )
  end

  if km_quick.list_fast then
    vim.keymap.set(
      "n",
      km_quick.list_fast,
      notes.list_quicknotes,
      { desc = "List quicknotes (fast) (TODO: not implemented)" }
    )
  end
end

--- Get current configuration
---
--- @return table Current configuration
function M.get_config()
  return require("neoweaver._internal.config").get()
end

--- Explorer module for browsing collections and notes
M.explorer = require("neoweaver._internal.explorer")

return M
