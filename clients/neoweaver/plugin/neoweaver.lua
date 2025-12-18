-- neoweaver.nvim plugin entry point
-- This file is automatically sourced by Neovim on startup
-- Commands are defined here following Neovim plugin best practices

if vim.g.loaded_neoweaver then
  return
end
vim.g.loaded_neoweaver = true

-- Create commands
-- Note: The plugin must be setup via require('neoweaver').setup() before these work

-- Notes commands
vim.api.nvim_create_user_command("NeoweaverNotesList", function()
  require("neoweaver._internal.notes").list_notes()
end, { desc = "List notes" })

vim.api.nvim_create_user_command("NeoweaverNotesOpen", function(opts)
  local id = tonumber(opts.args)
  if id then
    require("neoweaver._internal.notes").open_note(id)
  else
    vim.notify("Usage: :NeoweaverNotesOpen <note_id>", vim.log.levels.WARN)
  end
end, { nargs = 1, desc = "Open note by ID" })

vim.api.nvim_create_user_command("NeoweaverNotesNew", function()
  require("neoweaver._internal.notes").create_note()
end, { desc = "Create new note" })

vim.api.nvim_create_user_command("NeoweaverNotesNewWithTitle", function()
  require("neoweaver._internal.notes").create_note_with_title()
end, { desc = "Create note with title prompt" })

vim.api.nvim_create_user_command("NeoweaverNotesTitle", function()
  require("neoweaver._internal.notes").edit_title()
end, { desc = "Edit current note title" })

vim.api.nvim_create_user_command("NeoweaverNotesDelete", function(opts)
  local id = tonumber(opts.args)
  if id then
    require("neoweaver._internal.notes").delete_note(id)
  else
    vim.notify("Usage: :NeoweaverNotesDelete <note_id>", vim.log.levels.WARN)
  end
end, { nargs = 1, desc = "Delete note by ID" })

vim.api.nvim_create_user_command("NeoweaverNotesMeta", function(opts)
  local id = opts.args ~= "" and tonumber(opts.args) or nil
  require("neoweaver._internal.notes").edit_metadata(id)
end, { nargs = "?", desc = "Edit note metadata (TODO: not implemented)" })

vim.api.nvim_create_user_command("NeoweaverNotesQuick", function()
  require("neoweaver._internal.notes").create_quicknote()
end, { desc = "Create quicknote (TODO: not implemented)" })

vim.api.nvim_create_user_command("NeoweaverNotesQuickList", function()
  require("neoweaver._internal.notes").list_quicknotes()
end, { desc = "List quicknotes (TODO: not implemented)" })

vim.api.nvim_create_user_command("NeoweaverNotesQuickAmend", function()
  require("neoweaver._internal.notes").amend_quicknote()
end, { desc = "Amend quicknote (TODO: not implemented)" })

-- API/Server commands
vim.api.nvim_create_user_command("NeoweaverServerUse", function(opts)
  require("neoweaver._internal.api").set_current_server(opts.args)
end, {
  nargs = 1,
  complete = function(ArgLead)
    local api = require("neoweaver._internal.api")
    local matches = {}
    for _, name in ipairs(api.list_server_names()) do
      if name:find("^" .. vim.pesc(ArgLead)) then
        table.insert(matches, name)
      end
    end
    return matches
  end,
  desc = "Select Neoweaver server by name",
})

vim.api.nvim_create_user_command("NeoweaverToggleDebug", function()
  require("neoweaver._internal.api").toggle_debug()
end, { desc = "Toggle debug logging" })
