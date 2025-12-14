---
--- buffer.lua - Buffer management for notes
--- Handles buffer creation, saving, and state tracking
---
local M = {}

-- State management for open buffers
M.state = {
	-- Standard notes: map of note IDs to buffer numbers
	open_buffers = {},
}

---Sets a buffer to an item id key
---@param id integer The note's ID from the server
---@param bufnr integer
function M.set_buf_for_id(id, bufnr)
	if not id or id == 0 then
		vim.notify("Can't set buffer in the table without a key", vim.log.levels.ERROR)
		return
	end
	M.state.open_buffers[tostring(id)] = bufnr
end

---Returns the bufnr(int) associated with the item id if it exists, otherwise nil
---@param id integer The note's ID as retrieved from the server
---@return integer | nil
function M.get_buf_for_id(id)
	return M.state.open_buffers[tostring(id)] or nil
end

---Set the buffers table key value to nil. IE remove the bufnr associated with that id
---@param id integer The note's ID as retrieved from the server
function M.remove_buf_for_id(id)
	M.state.open_buffers[tostring(id)] = nil
end

--- Creates a new buffer, ft = markdown and attaches a buffer write autocommand to trigger the note save
---@param name string
---@param on_save fun(bufnr: integer) Callback when buffer is saved
---@return integer bufnr The created buffer number
function M.create_buffer(name, on_save)
	local bufnr = vim.fn.bufnr(name, true)
	vim.api.nvim_set_option_value("bufhidden", "wipe", { buf = bufnr })
	vim.api.nvim_set_option_value("filetype", "markdown", { buf = bufnr })

	vim.api.nvim_set_current_buf(bufnr)

	local group = vim.api.nvim_create_augroup("NwBufWrite_" .. tostring(bufnr), { clear = true })

	vim.api.nvim_create_autocmd("BufWriteCmd", {
		group = group,
		buffer = bufnr,
		callback = function()
			vim.notify("saving note " .. bufnr)
			on_save(bufnr)
		end,
	})

	return bufnr
end

--- Extract buffer data for saving
---@param bufnr integer
---@return table data Buffer data including text and metadata
function M.get_buffer_data(bufnr)
	local text = table.concat(vim.api.nvim_buf_get_lines(bufnr, 0, -1, false), "\n")
	
	return {
		id = vim.b[bufnr].mw_id,
		text = text,
		title = vim.b[bufnr].mw_title,
		description = vim.b[bufnr].mw_description,
		noteTypeId = vim.b[bufnr].mw_note_type_id,
		collectionId = vim.b[bufnr].mw_collection_id or 1,
		isTemplate = vim.b[bufnr].mw_is_template,
		meta = vim.b[bufnr].mw_meta or {},
		etag = vim.b[bufnr].etag,
	}
end

--- Set buffer data from a note
---@param bufnr integer
---@param note mind.v3.Note
function M.set_buffer_data(bufnr, note)
	vim.b[bufnr].mw_id = note.id
	vim.b[bufnr].etag = note.etag
	vim.b[bufnr].mw_title = note.title
	vim.b[bufnr].mw_description = note.description
	vim.b[bufnr].mw_note_type_id = note.noteTypeId
	vim.b[bufnr].mw_collection_id = note.collectionId
	vim.b[bufnr].mw_is_template = note.isTemplate
	vim.b[bufnr].mw_meta = note.metadata
end

--- Load note content into buffer
---@param bufnr integer
---@param note mind.v3.Note
function M.load_note_into_buffer(bufnr, note)
	local lines = vim.split(note.body or "", "\n")
	local name = note.title or "Untitled"

	vim.api.nvim_buf_set_name(bufnr, name)
	vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
	vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
	vim.api.nvim_set_current_buf(bufnr)

	M.set_buffer_data(bufnr, note)
end

return M
