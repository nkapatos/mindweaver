---
--- notes.lua - Note management for Neoweaver (v3)
--- Handles note listing, opening, editing, and saving
---
--- Reference: clients/mw/notes.lua (v1 implementation)
---
local api = require("neoweaver.api")
local buffer_manager = require("neoweaver.buffer.manager")

local M = {}

-- Debounce state for create_note
local last_create_time = 0
local DEBOUNCE_MS = 500

local allow_multiple_empty_notes = false

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

--- Create a new note (server-first approach with auto-generated title)
--- Server generates "Untitled 0", "Untitled 1", etc. via NewNote endpoint
function M.create_note()
	-- Debounce rapid calls unless feature explicitly allows multiple empty notes
	local now = vim.loop.now()
	if not allow_multiple_empty_notes and (now - last_create_time < DEBOUNCE_MS) then
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
		vim.api.nvim_set_option_value("modified", allow_multiple_empty_notes, { buf = bufnr })
		
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

--- Edit note metadata (frontmatter)
-- TODO: Implement metadata editing functionality
-- This will allow editing YAML frontmatter (tags, custom fields, etc.)
---@param note_id? integer Optional note ID (if nil, uses current buffer)
function M.edit_metadata(note_id)
	vim.notify("Metadata editing not yet implemented in v3", vim.log.levels.WARN)
	-- TODO: Implementation steps:
	-- 1. Get note_id from current buffer if not provided
	-- 2. Fetch note from API
	-- 3. Parse frontmatter from note.body
	-- 4. Open floating window with editable YAML
	-- 5. On save, update note with new frontmatter
end

--- Create a new quicknote in a floating window
-- TODO: Implement quicknotes functionality
-- Quicknotes are ephemeral floating windows for rapid note capture
function M.create_quicknote()
	vim.notify("Quicknotes not yet implemented in v3", vim.log.levels.WARN)
	-- TODO: Implementation steps:
	-- 1. Create floating window with configured dimensions
	-- 2. Create buffer with auto-generated title (timestamp-based)
	-- 3. On save, call NewNote API with quicknote note_type
	-- 4. Store reference for amend functionality
	-- Reference: clients/mw/notes.lua:handler__new_quicknote
end

--- List all quicknotes
-- TODO: Implement quicknotes listing
function M.list_quicknotes()
	vim.notify("Quicknotes not yet implemented in v3", vim.log.levels.WARN)
	-- TODO: Implementation steps:
	-- 1. Call ListNotes API with note_type filter for quicknote
	-- 2. Display in picker
	-- 3. On select, open in floating window
	-- Reference: clients/mw/notes.lua:handler__list_quicknotes
end

--- Amend the last created quicknote
-- TODO: Implement quicknote amend functionality
function M.amend_quicknote()
	vim.notify("Quicknote amend not yet implemented in v3", vim.log.levels.WARN)
	-- TODO: Implementation steps:
	-- 1. Retrieve last quicknote ID from state
	-- 2. Fetch note from API
	-- 3. Open in floating window with existing content
	-- 4. Allow editing and save
	-- Reference: clients/mw/notes.lua:handler__amend_quicknote
end

--- Delete a note by ID
---@param note_id integer The note ID to delete
function M.delete_note(note_id)
	if not note_id then
		vim.notify("Invalid note ID", vim.log.levels.ERROR)
		return
	end

	-- Ask for confirmation
	vim.ui.input({
		prompt = string.format("Delete note %d? (y/N): ", note_id),
	}, function(input)
		if not input or (input:lower() ~= "y" and input:lower() ~= "yes") then
			vim.notify("Delete cancelled", vim.log.levels.INFO)
			return
		end

		-- Call delete API
		---@type mind.v3.DeleteNoteRequest
		local req = { id = note_id }

		api.notes.delete(req, function(res)
			if res.error then
				vim.notify("Failed to delete note: " .. res.error.message, vim.log.levels.ERROR)
				return
			end

			-- Close buffer if it's open
			local bufnr = buffer_manager.get("note", note_id)
			if bufnr and vim.api.nvim_buf_is_valid(bufnr) then
				vim.api.nvim_buf_delete(bufnr, { force = true })
			end

			vim.notify("Note deleted successfully", vim.log.levels.INFO)
		end)
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

function M.setup(opts)
	opts = opts or {}
	allow_multiple_empty_notes = opts.allow_multiple_empty_notes == true

	-- Register note type handlers with buffer manager
	register_handlers()
end

return M
