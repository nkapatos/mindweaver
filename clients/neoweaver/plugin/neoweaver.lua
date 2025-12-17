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
vim.api.nvim_create_user_command("NotesList", function()
	require("neoweaver.notes").list_notes()
end, { desc = "List all notes" })

vim.api.nvim_create_user_command("NotesOpen", function(opts)
	local id = tonumber(opts.args)
	if id then
		require("neoweaver.notes").open_note(id)
	else
		vim.notify("Usage: :NotesOpen <note_id>", vim.log.levels.WARN)
	end
end, { nargs = 1, desc = "Open note by ID" })

vim.api.nvim_create_user_command("NotesNew", function()
	require("neoweaver.notes").create_note()
end, { desc = "Create new note" })

-- NotesEdit is an alias for NotesOpen (for compatibility with mw)
vim.api.nvim_create_user_command("NotesEdit", function(opts)
	local id = tonumber(opts.args)
	if id then
		require("neoweaver.notes").open_note(id)
	else
		vim.notify("Usage: :NotesEdit <note_id>", vim.log.levels.WARN)
	end
end, { nargs = 1, desc = "Edit note by ID (alias for NotesOpen)" })

vim.api.nvim_create_user_command("NotesDelete", function(opts)
	local id = tonumber(opts.args)
	if id then
		require("neoweaver.notes").delete_note(id)
	else
		vim.notify("Usage: :NotesDelete <note_id>", vim.log.levels.WARN)
	end
end, { nargs = 1, desc = "Delete note by ID" })

-- TODO: Metadata editing (not yet implemented in v3)
vim.api.nvim_create_user_command("NotesMeta", function(opts)
	local id = opts.args ~= "" and tonumber(opts.args) or nil
	require("neoweaver.notes").edit_metadata(id)
end, { nargs = "?", desc = "Edit note metadata (TODO: not implemented)" })

-- TODO: Quicknotes commands (not yet implemented in v3)
vim.api.nvim_create_user_command("NotesQuick", function()
	require("neoweaver.notes").create_quicknote()
end, { desc = "Create quicknote (TODO: not implemented)" })

vim.api.nvim_create_user_command("NotesQuickList", function()
	require("neoweaver.notes").list_quicknotes()
end, { desc = "List quicknotes (TODO: not implemented)" })

vim.api.nvim_create_user_command("NotesQuickAmend", function()
	require("neoweaver.notes").amend_quicknote()
end, { desc = "Amend last quicknote (TODO: not implemented)" })

-- API/Server commands
vim.api.nvim_create_user_command("NeoweaverServerUse", function(opts)
	require("neoweaver.api").set_current_server(opts.args)
end, {
	nargs = 1,
	complete = function(ArgLead)
		local api = require("neoweaver.api")
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
	require("neoweaver.api").toggle_debug()
end, { desc = "Toggle debug logging" })
