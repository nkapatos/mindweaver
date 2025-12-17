---
--- init.lua - Neoweaver entry point (v3)
--- Public API for the neoweaver plugin
---
local M = {}

--- Setup the neoweaver plugin
---
--- Configures the API client, note handlers, and optional keymaps.
--- Must be called before using any plugin functionality.
---
---@param opts? table Configuration options
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

	-- Standard notes keymaps
	if km_notes.list then
		vim.keymap.set("n", km_notes.list, notes.list_notes, { desc = "List notes" })
	end

	if km_notes.open then
		vim.keymap.set("n", km_notes.open, function()
			vim.ui.input({ prompt = "Note ID: " }, function(input)
				local id = tonumber(input)
				if id then
					notes.open_note(id)
				end
			end)
		end, { desc = "Open note by ID" })
	end

	if km_notes.edit then
		vim.keymap.set("n", km_notes.edit, function()
			vim.ui.input({ prompt = "Note ID: " }, function(input)
				local id = tonumber(input)
				if id then
					notes.open_note(id)
				end
			end)
		end, { desc = "Edit note by ID" })
	end

	if km_notes.new then
		vim.keymap.set("n", km_notes.new, notes.create_note, { desc = "Create new note" })
	end

	if km_notes.delete then
		vim.keymap.set("n", km_notes.delete, function()
			vim.ui.input({ prompt = "Note ID to delete: " }, function(input)
				local id = tonumber(input)
				if id then
					notes.delete_note(id)
				end
			end)
		end, { desc = "Delete note by ID" })
	end

	if km_notes.meta then
		vim.keymap.set("n", km_notes.meta, function()
			notes.edit_metadata()
		end, { desc = "Edit note metadata (TODO: not implemented)" })
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
		vim.keymap.set("n", km_quick.new_fast, notes.create_quicknote, { desc = "New quicknote (fast) (TODO: not implemented)" })
	end

	if km_quick.amend_fast then
		vim.keymap.set("n", km_quick.amend_fast, notes.amend_quicknote, { desc = "Amend quicknote (fast) (TODO: not implemented)" })
	end

	if km_quick.list_fast then
		vim.keymap.set("n", km_quick.list_fast, notes.list_quicknotes, { desc = "List quicknotes (fast) (TODO: not implemented)" })
	end
end

--- Get current configuration
---
--- @return table Current configuration
function M.get_config()
	return require("neoweaver._internal.config").get()
end

return M
