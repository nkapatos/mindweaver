---
--- notes.lua - Note management for Neoweaver (v3)
--- Handles note listing, opening, editing, and saving
---
--- Reference: clients/mw/notes.lua (v1 implementation)
---
local api = require("neoweaver.api")
local buffer_manager = require("neoweaver.buffer.manager")

local M = {}

-- Register note type handlers with buffer manager
-- This is called once during setup
local function register_handlers()
	buffer_manager.register_type("note", {
		on_save = function(bufnr, id)
			M.save_note(bufnr, id)
		end,
		on_close = function(bufnr, id)
			-- Cleanup if needed in future
		end,
	})
end

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
			-- Open note for editing
			M.open_note(tonumber(selected_note.id))
		end)
	end)
end

-- Debounce state for create_note
local last_create_time = 0
local DEBOUNCE_MS = 500

--- Create a new note (server-first approach with auto-generated title)
--- Server generates "Untitled 0", "Untitled 1", etc. via NewNote endpoint
function M.create_note()
	-- Debounce rapid calls to prevent accidental spam
	local now = vim.loop.now()
	if now - last_create_time < DEBOUNCE_MS then
		vim.notify("Please wait before creating another note", vim.log.levels.WARN)
		return
	end
	last_create_time = now
	
	-- Call NewNote endpoint - server generates title automatically
	---@type mind.v3.NewNoteRequest
	local req = {
		collectionId = 1, -- Default collection (optional, server defaults to 1)
	}
	
	api.notes.new(req, function(res)
		if res.error then
			vim.notify("Failed to create note: " .. res.error.message, vim.log.levels.ERROR)
			return
		end
		
		-- Note created with auto-generated title "Untitled 0", "Untitled 1", etc.
		---@type mind.v3.Note
		local note = res.data
		local note_id = tonumber(note.id)
		
		-- Create managed buffer via buffer_manager
		local bufnr = buffer_manager.create({
			type = "note",
			id = note_id,
			name = note.title, -- Server-generated "Untitled N"
			filetype = "markdown",
			modifiable = true,
		})
		
		-- Buffer starts empty (note.body = "")
		vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, {})
		vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
		
		-- Store note metadata
		vim.b[bufnr].note_id = note_id
		vim.b[bufnr].note_title = note.title
		vim.b[bufnr].note_etag = note.etag
		vim.b[bufnr].note_collection_id = note.collectionId
		vim.b[bufnr].note_type_id = note.noteTypeId
		vim.b[bufnr].note_metadata = note.metadata or {}
		
		vim.notify("Note created: " .. note.title, vim.log.levels.INFO)
	end)
end

--- Open a note in a buffer for editing
--- If buffer already exists, switch to it; otherwise fetch and create
---@param note_id integer The note ID to open
function M.open_note(note_id)
	if not note_id then
		vim.notify("Invalid note ID", vim.log.levels.ERROR)
		return
	end

	-- Check if buffer already exists
	local existing = buffer_manager.get("note", note_id)
	if existing and vim.api.nvim_buf_is_valid(existing) then
		vim.api.nvim_set_current_buf(existing)
		return
	end

	-- Fetch note from API
	---@type mind.v3.GetNoteRequest
	local req = { id = note_id }
	
	api.notes.get(req, function(res)
		if res.error then
			vim.notify("Error loading note: " .. res.error.message, vim.log.levels.ERROR)
			return
		end

		-- v3 API: Response is mind.v3.Note directly
		---@type mind.v3.Note
		local note = res.data

		-- Create buffer via buffer_manager
		local bufnr = buffer_manager.create({
			type = "note",
			id = note_id,
			name = note.title or "Untitled",
			filetype = "markdown",
			modifiable = true,
		})

		-- Load content into buffer
		local lines = vim.split(note.body or "", "\n")
		vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
		vim.api.nvim_set_option_value("modified", false, { buf = bufnr })

		-- Store note data in buffer variables
		-- Using note_* prefix for domain-specific storage
		vim.b[bufnr].note_id = note_id
		vim.b[bufnr].note_title = note.title
		vim.b[bufnr].note_etag = note.etag
		vim.b[bufnr].note_collection_id = note.collectionId
		vim.b[bufnr].note_type_id = note.noteTypeId
		vim.b[bufnr].note_metadata = note.metadata or {}
	end)
end

--- Save note buffer content to server
--- Called by buffer_manager when buffer is saved (:w)
--- Always updates existing note (server-first approach ensures ID exists)
---@param bufnr integer Buffer number
---@param id integer Note ID
function M.save_note(bufnr, id)
	-- Extract buffer content
	local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
	local body = table.concat(lines, "\n")

	-- Get stored note data
	local title = vim.b[bufnr].note_title or "Untitled"
	local etag = vim.b[bufnr].note_etag
	local collection_id = vim.b[bufnr].note_collection_id or 1
	local note_type_id = vim.b[bufnr].note_type_id
	local metadata = vim.b[bufnr].note_metadata or {}

	-- Build update request
	---@type mind.v3.ReplaceNoteRequest
	local req = {
		id = id,
		title = title,
		body = body,
		collectionId = collection_id,
	}
	
	-- Add optional fields only if they have values
	if note_type_id then
		req.noteTypeId = note_type_id
	end
	
	if metadata and next(metadata) ~= nil then
		req.metadata = metadata
	end

	-- Call API with etag for optimistic locking
	api.notes.update(req, etag, function(res)
		if res.error then
			vim.notify("Save failed: " .. res.error.message, vim.log.levels.ERROR)
			-- TODO: Handle conflict (412) - requires diff.lua integration
			return
		end

		-- Update etag and mark buffer as unmodified
		---@type mind.v3.Note
		local updated_note = res.data
		vim.b[bufnr].note_etag = updated_note.etag
		vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
		vim.notify("Note saved successfully", vim.log.levels.INFO)
	end)
end

function M.setup()
	-- Register note type handlers with buffer manager
	register_handlers()

	-- Create commands
	vim.api.nvim_create_user_command("NotesList", M.list_notes, { desc = "List all notes (v3)" })
	vim.api.nvim_create_user_command("NotesOpen", function(opts)
		local id = tonumber(opts.args)
		if id then
			M.open_note(id)
		else
			vim.notify("Usage: :NotesOpen <note_id>", vim.log.levels.WARN)
		end
	end, { nargs = 1, desc = "Open note by ID (v3)" })
	vim.api.nvim_create_user_command("NotesNew", M.create_note, { desc = "Create new note (v3)" })

	-- Create keymaps
	vim.keymap.set("n", "<leader>nl", M.list_notes, { desc = "List notes (v3)" })
	vim.keymap.set("n", "<leader>no", function()
		vim.ui.input({ prompt = "Note ID: " }, function(input)
			local id = tonumber(input)
			if id then
				M.open_note(id)
			end
		end)
	end, { desc = "Open note by ID (v3)" })
	vim.keymap.set("n", "<leader>nn", M.create_note, { desc = "Create new note (v3)" })
end

return M
