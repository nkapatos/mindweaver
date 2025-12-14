---
--- notes.lua - Unified note management for the MW plugin
--- Merges functionality from content.lua and quicknotes.lua
--- Supports both standard notes (normal buffers) and quicknotes (floating windows)
---
--- Recent changes (2025-11-24):
--- - Added note_type filtering to NotesList (excludes quicknotes)
--- - Prepared for backend API filtering support
--- - Integrated explorer picker for note listing
---
--- Next steps:
--- - Coordinate with backend for ?note_type and ?project query params
--- - Add filtering options to picker
---
local api = require("mw.api")
local diff = require("mw.diff")
local metadata = require("mw.metadata")
local picker = require("mw.explorer.picker")

local M = {}

-- Configuration defaults
local defaults = {
	quicknote_opts = {
		relative = "editor",
		width = 0.4, -- Fraction of editor width
		height = 0.2, -- Fraction of editor height
		row = 0.5, -- Centered vertically (adjusted after computation)
		col = 0.5, -- Centered horizontally (adjusted after computation)
		style = "minimal",
		border = "rounded",
		title = "Quick Note",
		title_pos = "center",
	},
	title_template = "%Y%m%d%H%M", -- Zettelkasten-style timestamp
}

M.config = defaults

-- State management
local state = {
	-- Standard notes: map of note IDs to buffer numbers
	open_buffers = {},
	-- Quicknote state
	quicknote = {
		bufnr = nil,
		winid = nil,
		recent = {
			etag = nil,
			id = nil,
		},
	},
}

---Sets a buffer to an item id key
---@param id string The note's ID from the server
---@param bufnr integer
local function set_buf_for_id(id, bufnr)
	if not id or id == "" then
		vim.notify("Can't set buffer in the table without a key", vim.log.levels.ERROR)
		return
	end
	state.open_buffers[id] = bufnr
end

---Returns the bufnr(int) associated with the item id if it exists, otherwise nil
---@param id string The note's ID as retrieved from the server
---@return integer | nil
local function get_buf_for_id(id)
	return state.open_buffers[id] or nil
end

---Set the buffers table key value to nil. IE remove the bufnr associated with that id
---@param id string The note's ID as retrieved from the server
local function remove_buf_for_id(id)
	state.open_buffers[id] = nil
end

-- Forward declarations
local create_buffer
local save_buffer
local note_create_cb_handler
local note_update_cb_handler
local note_update_conflict_resolver

--- Creates a new buffer, ft = markdown and attaches a buffer write autocommand to trigger the note save
---@param name string
---@return integer
function create_buffer(name)
	local bufnr = vim.fn.bufnr(name, true)
	vim.api.nvim_set_option_value("bufhidden", "wipe", { buf = bufnr })
	vim.api.nvim_set_option_value("filetype", "markdown", { buf = bufnr })

	vim.api.nvim_set_current_buf(bufnr)

	local group = vim.api.nvim_create_augroup("MwBufWrite_" .. tostring(bufnr), { clear = true })

	vim.api.nvim_create_autocmd("BufWriteCmd", {
		group = group,
		buffer = bufnr,
		callback = function()
			vim.notify("saving note " .. bufnr)
			save_buffer(bufnr)
		end,
	})

	return bufnr
end

function save_buffer(bufnr)
	local id = vim.b[bufnr].mw_id
	local text = table.concat(vim.api.nvim_buf_get_lines(bufnr, 0, -1, false), "\n")
	local payload = { body = text }

	-- Extract metadata (cached internally to avoid re-parsing on every save)
	local extracted_meta = metadata.extract_metadata()
	payload.meta = extracted_meta

	if vim.b[bufnr].mw_title then
		payload.title = vim.b[bufnr].mw_title
	end
	if vim.b[bufnr].mw_description then
		payload.description = vim.b[bufnr].mw_description
	end
	if vim.b[bufnr].mw_note_type then
		payload.note_type = vim.b[bufnr].mw_note_type
	end
	if vim.b[bufnr].mw_meta then
		payload.meta = vim.b[bufnr].mw_meta
	end

	if id then
		api.notes.update(id, payload, vim.b[bufnr].etag, function(res)
			note_update_cb_handler(res, bufnr)
		end)
	else
		api.notes.create(payload, function(res)
			note_create_cb_handler(res, bufnr)
		end)
	end
end

function note_create_cb_handler(res, bufnr)
	if res.error then
		vim.notify(vim.inspect(res.error), vim.log.levels.ERROR)
		return
	end
	-- API Response Format: Single item operations return flat data object (not items array)
	-- POST /api/mind/notes → { data: { id, etag, title, body, ... } }
	local id = res.data.id
	vim.b[bufnr].mw_id = id
	vim.b[bufnr].etag = res.data.etag
	vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
	set_buf_for_id(id, bufnr)
end

function note_update_cb_handler(res, bufnr)
	if res.error then
		vim.notify(vim.inspect(res.error), vim.log.levels.ERROR)
		if res.error.code == 412 then
			vim.notify("Server version doesn't match: Fetching latest...", vim.log.levels.WARN)
			api.notes.get(vim.b[bufnr].mw_id, function(latest_note_res)
				if latest_note_res.error then
					vim.notify(vim.inspect(latest_note_res.error), vim.log.levels.ERROR)
					return
				end
				note_update_conflict_resolver(latest_note_res, bufnr)
			end)
			return
		end
		return
	end
	vim.b[bufnr].etag = res.data.etag
	vim.api.nvim_buf_set_name(bufnr, vim.b[bufnr].mw_title)
	vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
end

function note_update_conflict_resolver(latest_note_res, bufnr)
	vim.notify("entering the conflict resolver")
	-- API Response Format: GET /api/mind/notes/:id → { data: { id, body, etag, ... } }
	local srv_lines = vim.split(latest_note_res.data.body, "\n")
	local srv_etag = latest_note_res.data.etag

	-- Ensure buffer is listed and modifiable
	if not vim.api.nvim_buf_is_loaded(bufnr) then
		vim.notify("Buffer not loaded, cannot enable diff overlays.", vim.log.levels.ERROR)
		return
	end

	if vim.api.nvim_get_option_value("modifiable", { buf = bufnr }) == false then
		vim.api.nvim_set_option_value("modifiable", true, { buf = bufnr })
	end
	if vim.api.nvim_get_option_value("buflisted", { buf = bufnr }) == false then
		vim.api.nvim_set_option_value("buflisted", true, { buf = bufnr })
	end

	diff.enable(bufnr)
	diff.map_keys(bufnr)

	vim.defer_fn(function()
		local ok, err = pcall(diff.set_ref_text, bufnr, srv_lines)
		if not ok then
			vim.notify("Diff overlay couldn't be enabled for this buffer: " .. tostring(err), vim.log.levels.ERROR)
			return
		end
	end, 50)

	vim.notify("Resolve conflicts with ]c/[c (navigate), gh (apply hunk). Save (:w) to retry.", vim.log.levels.INFO)

	-- Autocommand to clean up and retry save after resolution
	local group = vim.api.nvim_create_augroup("mw_conflict_resolution_" .. bufnr, { clear = true })
	vim.api.nvim_create_autocmd("BufWriteCmd", {
		group = group,
		buffer = bufnr,
		callback = function()
			diff.disable(bufnr)
			vim.api.nvim_del_augroup_by_id(group)
			-- Update etag and retry
			vim.b[bufnr].etag = srv_etag
			save_buffer(bufnr)
		end,
		once = true,
	})
end

local function handle_note_load(res)
	local bufnr
	-- API Response Format: GET /api/mind/notes/:id → { data: { id, title, body, etag, ... } }
	local item = res.data
	local etag = res.data.etag
	local existing_bufnr = state.open_buffers[item.id] or nil
	local name = item.title or vim.fn.strftime(M.config.title_template, item.created_at)

	if not existing_bufnr or not vim.api.nvim_buf_is_valid(existing_bufnr) then
		bufnr = create_buffer(name)
	else
		bufnr = existing_bufnr
	end

	local lines = vim.split(item.body, "\n")

	vim.api.nvim_buf_set_name(bufnr, name)
	vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
	vim.api.nvim_set_option_value("modified", false, { buf = bufnr })

	-- Open buffer in current window
	vim.api.nvim_set_current_buf(bufnr)
	vim.b[bufnr].mw_id = item.id
	vim.b[bufnr].etag = etag
	vim.b[bufnr].mw_title = item.title
	vim.b[bufnr].mw_description = item.description
	vim.b[bufnr].mw_note_type = item.note_type
	vim.b[bufnr].mw_meta = item.meta
end

-- ============================================================================
-- QUICKNOTE FUNCTIONALITY (Floating Window)
-- ============================================================================

local function save_quicknote_buffer(bufnr)
	local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
	local content = table.concat(lines, "\n")

	-- Extract metadata (cached internally to avoid re-parsing on every save)
	local extracted_meta = metadata.extract_metadata()
	if content == "" then
		return
	end
	local title = M.config.title_template and vim.fn.strftime(M.config.title_template) or nil
	local payload = {
		body = content,
		meta = extracted_meta,
		title = title,
		note_type = "quicknote",
	}

	if vim.b[state.quicknote.bufnr].id then
		api.notes.update(vim.b[state.quicknote.bufnr].id, payload, state.quicknote.recent.etag, function(res)
			if res.error then
				if res.error.status == 412 then
					vim.notify(
						"Quicknote update failed due to conflict: " .. vim.inspect(res.error),
						vim.log.levels.ERROR
					)
					return
				end
				vim.notify("Quicknote update failed: " .. vim.inspect(res.error), vim.log.levels.ERROR)
				return
			end
			vim.b[state.quicknote.bufnr].id = nil
			state.quicknote.recent.etag = res.data.etag
			vim.notify("Quicknote updated", vim.log.levels.INFO)
		end)
	else
		api.notes.create(payload, function(res)
			if res.error then
				vim.notify("Quicknote save failed: " .. vim.inspect(res.error), vim.log.levels.ERROR)
			else
				-- Single item response: data is a flat object (not items array)
				state.quicknote.recent.id = res.data.id
				state.quicknote.recent.etag = res.data.etag
				vim.notify("Quicknote saved", vim.log.levels.INFO)
			end
		end)
	end
end

-- Create single quicknote buffer
local function create_quicknotes_buffer()
	local bufnr = vim.api.nvim_create_buf(false, true)
	vim.api.nvim_set_option_value("buftype", "nofile", { buf = bufnr })
	vim.api.nvim_set_option_value("filetype", "markdown", { buf = bufnr })
	vim.api.nvim_buf_set_name(bufnr, "Quicknote")

	-- Set autocommands and keymaps once
	vim.api.nvim_create_autocmd("BufWinLeave", {
		buffer = bufnr,
		callback = function()
			save_quicknote_buffer(bufnr)
			if state.quicknote.bufnr == bufnr then
				state.quicknote.winid = nil -- Only clear window, keep buffer
			end
		end,
	})

	vim.keymap.set({ "n" }, "q", function()
		vim.cmd("close")
	end, { buffer = bufnr })

	return bufnr
end

-- Open buffer in floating window
local function open_in_window(bufnr)
	if state.quicknote.winid and vim.api.nvim_win_is_valid(state.quicknote.winid) then
		vim.api.nvim_win_set_buf(state.quicknote.winid, bufnr)
	else
		state.quicknote.winid = vim.api.nvim_open_win(bufnr, true, M.config.quicknote_opts)
	end
end

-- ============================================================================
-- COMMAND HANDLERS
-- ============================================================================

--- List all standard notes (excludes quicknotes)
--- Uses the explorer picker with vim.ui.select integration
---
--- If collection_id is present in .nvmwmeta.json, uses collection-filtered endpoint:
--- GET /api/mind/collections/:id/notes
---
--- Otherwise falls back to global notes endpoint:
--- GET /api/mind/notes
---
--- Backend API supports:
--- - ?fields=id,title,body (field selection)
--- - ?include=meta,tags (include relations)
--- - ?project_id=X (filter by project ID)
--- - ?tag_id=X (filter by tag)
--- - ?meta_key=X (filter by metadata key)
---
--- TODO: Add note_type param to backend for filtering
function M.handler__list_notes()
	-- Extract metadata to check for collection_id
	local meta = metadata.extract_metadata()
	local collection_id = meta.collection_id
	
	-- Request with metadata and tags included
	local query = { include = "meta,tags" }
	
	-- Choose endpoint based on collection_id presence
	local list_fn = function(cb)
		if collection_id then
			-- Use collection-filtered endpoint
			api.list_notes_by_collection(collection_id, query, cb)
		else
			-- Use global notes endpoint
			api.notes.list(query, cb)
		end
	end
	
	list_fn(function(res)
		if res.error then
			vim.notify(res.message or vim.inspect(res.error), vim.log.levels.ERROR)
			return
		end

		local items = res.data.items or {}

		-- Client-side fallback filtering if backend doesn't support note_type param
		-- TODO: Remove this once backend supports ?note_type filtering
		local filtered_items = {}
		for _, item in ipairs(items) do
			-- Only include items that are NOT quicknotes
			if not item.note_type or item.note_type ~= "quicknote" then
				table.insert(filtered_items, item)
			end
		end

		if #filtered_items == 0 then
			vim.notify("No notes found!", vim.log.levels.INFO)
			return
		end

		-- Use explorer picker with real data
		local current_project = meta.project or "unknown"
		picker.show_picker(filtered_items, {
			project = current_project,
			note_type = "notes",
			on_select = function(note)
				M.handler__edit_note(note.id)
			end,
		})
	end)
end

--- List all quicknotes only
--- Uses the explorer picker with vim.ui.select integration
--- TODO: Backend should support ?note_type=quicknote query param for filtering
function M.handler__list_quicknotes()
	-- Request with metadata and tags included
	api.notes.list({ include = "meta,tags" }, function(res)
		if res.error then
			vim.notify(vim.inspect(res.error), vim.log.levels.ERROR)
			return
		end
		local items = res.data.items or {}

		-- Client-side fallback filtering if backend doesn't support note_type param
		-- TODO: Remove this once backend supports ?note_type filtering
		local filtered_items = {}
		for _, item in ipairs(items) do
			if item.note_type == "quicknote" then
				table.insert(filtered_items, item)
			end
		end

		if #filtered_items == 0 then
			vim.notify("No quicknotes found", vim.log.levels.INFO)
			return
		end

		-- Use explorer picker with real data
		local current_project = metadata.extract_metadata().project or "unknown"
		picker.show_picker(filtered_items, {
			project = current_project,
			note_type = "quicknotes",
			on_select = function(note)
				-- Reuse the amend functionality by storing the selected id
				state.quicknote.recent.id = note.id
				state.quicknote.recent.etag = note.etag
				M.handler__amend_quicknote()
			end,
		})
	end)
end

function M.handler__new_note()
	local name = vim.fn.strftime(M.config.title_template) .. " - Untitled"
	local bufnr = create_buffer(name)
	vim.b[bufnr].mw_title = name
	vim.b[bufnr].mw_note_type = "note"
end

function M.handler__new_quicknote()
	-- Clear buffer for new note
	vim.api.nvim_buf_set_lines(state.quicknote.bufnr, 0, -1, false, {})
	open_in_window(state.quicknote.bufnr)
end

function M.handler__amend_quicknote()
	if state.quicknote.winid then
		return
	end
	if not state.quicknote.recent.id then
		vim.notify("There is no previous quicknote to amend. Creating a new one", vim.log.levels.INFO)
		return M.handler__new_quicknote()
	end

	api.notes.get(state.quicknote.recent.id, function(res)
		if res.error then
			vim.notify("hmm" .. vim.inspect(res.error), vim.log.levels.ERROR)
			return
		end
		-- Single item response: data is a flat object (not items array)
		local item = res.data
		vim.b[state.quicknote.bufnr].id = state.quicknote.recent.id
		vim.api.nvim_buf_set_lines(state.quicknote.bufnr, 0, -1, false, vim.split(item.body, "\n"))
		open_in_window(state.quicknote.bufnr)
	end)
end

function M.handler__edit_note(id)
	if not id or id == "" then
		vim.ui.input({ prompt = "Enter note ID to edit:" }, function(input_id)
			if input_id and input_id == "" then
				vim.notify("No ID provided.", vim.log.levels.WARN)
				return
			end
			M.handler__edit_note(input_id)
		end)
		return
	end

	api.notes.get(id, function(res)
		if res.error then
			vim.notify(vim.inspect(res.error), vim.log.levels.ERROR)
			return
		end
		handle_note_load(res)
	end)
end

function M.handler__delete_note(id)
	api.notes.delete(id, function(res)
		if res.error then
			vim.notify(vim.inspect(res.error), vim.log.levels.ERROR)
			return
		end
		vim.notify("Note with id " .. id .. " deleted successfully")
		local bufnr = get_buf_for_id(id)
		if bufnr and vim.api.nvim_buf_is_valid(bufnr) then
			vim.api.nvim_buf_delete(bufnr, { force = true })
		end
		remove_buf_for_id(id)
	end)
end

function M.handler__note_meta()
	local bufnr = vim.api.nvim_get_current_buf()

	if not vim.b[bufnr].mw_id and not vim.b[bufnr].mw_title then
		vim.notify("Not a managed note buffer.", vim.log.levels.WARN)
		return
	end

	vim.ui.input({ prompt = "Edit Title:", default = vim.b[bufnr].mw_title or "" }, function(title)
		if title == nil then
			return
		end
		vim.b[bufnr].mw_title = title
		vim.ui.input({ prompt = "Edit Description:", default = vim.b[bufnr].mw_description or "" }, function(desc)
			if desc == nil then
				return
			end
			vim.b[bufnr].mw_description = desc
			vim.notify("Metadata updated. Save buffer to persist changes.", vim.log.levels.INFO)
		end)
	end)
end

-- ============================================================================
-- SETUP
-- ============================================================================

---
-- Setup for notes module. Accepts user config and passes metadata config to metadata.setup().
-- Example: require('mw').setup({ metadata = { markers = { ... } } })
function M.setup(usr_opts)
	usr_opts = usr_opts or {}
	M.config = vim.tbl_deep_extend("force", defaults, usr_opts)

	-- Pass metadata config to metadata.setup if present
	if usr_opts.metadata then
		metadata.setup(usr_opts.metadata)
	end

	local opts = M.config.quicknote_opts

	opts.width = math.floor(vim.o.columns * (type(opts.width) == "number" and opts.width or 0.8))
	opts.height = math.floor(vim.o.lines * (type(opts.height) == "number" and opts.height or 0.6))
	opts.row = math.floor((vim.o.lines - opts.height) * (type(opts.row) == "number" and opts.row or 0.5))
	opts.col = math.floor((vim.o.columns - opts.width) * (type(opts.col) == "number" and opts.col or 0.5))

	-- Commands for standard notes
	vim.api.nvim_create_user_command("NotesList", M.handler__list_notes, { desc = "List all notes" })
	vim.api.nvim_create_user_command("NotesNew", M.handler__new_note, { desc = "Create new note" })
	vim.api.nvim_create_user_command("NotesEdit", function(cmd_opts)
		M.handler__edit_note(cmd_opts.args)
	end, { nargs = "?", desc = "Edit note by ID" })
	vim.api.nvim_create_user_command("NotesDelete", function(cmd_opts)
		M.handler__delete_note(cmd_opts.args)
	end, { nargs = 1, desc = "Delete note by ID" })
	vim.api.nvim_create_user_command("NotesMeta", M.handler__note_meta, { desc = "Edit note metadata" })

	-- Commands for quicknotes
	vim.api.nvim_create_user_command("NotesQuick", M.handler__new_quicknote, { desc = "Open new quicknote" })
	vim.api.nvim_create_user_command("NotesQuickList", M.handler__list_quicknotes, { desc = "List quicknotes" })
	vim.api.nvim_create_user_command("NotesQuickAmend", M.handler__amend_quicknote, { desc = "Amend last quicknote" })

	-- Keymaps for standard notes
	vim.keymap.set("n", "<leader>wl", M.handler__list_notes, { desc = "List notes" })
	vim.keymap.set("n", "<leader>wn", M.handler__new_note, { desc = "New note" })
	vim.keymap.set("n", "<leader>we", M.handler__edit_note, { desc = "Edit note by ID" })
	vim.keymap.set("n", "<leader>wd", function()
		vim.ui.input({ prompt = "Enter note ID to delete:" }, function(id)
			if id and id ~= "" then
				M.handler__delete_note(id)
			else
				vim.notify("No ID provided.", vim.log.levels.WARN)
			end
		end)
	end, { desc = "Delete note by ID" })
	vim.keymap.set("n", "<leader>wm", M.handler__note_meta, { desc = "Edit note metadata" })

	-- Keymaps for quicknotes (dual mapping: <leader>w* and <leader>.*)
	vim.keymap.set("n", "<leader>wq", M.handler__new_quicknote, { desc = "New quicknote" })
	vim.keymap.set("n", "<leader>wa", M.handler__amend_quicknote, { desc = "Amend quicknote" })
	vim.keymap.set("n", "<leader>wql", M.handler__list_quicknotes, { desc = "List quicknotes" })

	-- Quicknotes: <leader>.* for fast access
	vim.keymap.set("n", "<leader>.n", M.handler__new_quicknote, { desc = "New quicknote (fast)" })
	vim.keymap.set("n", "<leader>.a", M.handler__amend_quicknote, { desc = "Amend quicknote (fast)" })
	vim.keymap.set("n", "<leader>.l", M.handler__list_quicknotes, { desc = "List quicknotes (fast)" })

	-- Create the window buffer for the quicknotes
	if not state.quicknote.bufnr or not vim.api.nvim_buf_is_valid(state.quicknote.bufnr) then
		state.quicknote.bufnr = create_quicknotes_buffer()
	end
end

return M
