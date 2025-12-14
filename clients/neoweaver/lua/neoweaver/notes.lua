---
--- notes.lua - Minimal note management for Neoweaver (v3)
--- Start simple: List notes only
---
local api = require("neoweaver.api")

local M = {}

--- List all notes using vim.ui.select
function M.list_notes()
	---@type mind.v3.ListNotesRequest
	local req = {
		pageSize = 100,
		pageToken = "",
	}

	api.notes.list(req, function(res)
		if res.error then
			vim.notify("Error listing notes: " .. vim.inspect(res.error), vim.log.levels.ERROR)
			return
		end

		-- v3 API: Response is mind.v3.ListNotesResponse directly
		---@type mind.v3.ListNotesResponse
		local list_res = res.data
		local notes = list_res.notes or {}

		if #notes == 0 then
			vim.notify("No notes found!", vim.log.levels.INFO)
			return
		end

		-- Format notes for display
		local items = {}
		for _, note in ipairs(notes) do
			table.insert(items, string.format("[%d] %s", note.id, note.title))
		end

		-- Show picker
		vim.ui.select(items, {
			prompt = "Select a note:",
		}, function(choice, idx)
			if not choice then
				return
			end
			local selected_note = notes[idx]
			vim.notify(
				string.format("Selected: [%d] %s", selected_note.id, selected_note.title),
				vim.log.levels.INFO
			)
			-- TODO: Open note for editing
		end)
	end)
end

function M.setup()
	-- Create command
	vim.api.nvim_create_user_command("NotesList", M.list_notes, { desc = "List all notes (v3)" })

	-- Create keymap
	vim.keymap.set("n", "<leader>nl", M.list_notes, { desc = "List notes (v3)" })
	
	vim.notify("Neoweaver (v3) notes module loaded", vim.log.levels.INFO)
end

return M
