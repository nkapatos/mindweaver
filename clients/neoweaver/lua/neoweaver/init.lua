---
--- init.lua - Neoweaver entry point (v3)
---
local M = {}

local config = {
	allow_multiple_empty_notes = false,
}

function M.setup(opts)
	opts = opts or {}

	if opts.allow_multiple_empty_notes ~= nil then
		config.allow_multiple_empty_notes = opts.allow_multiple_empty_notes == true
	else
		config.allow_multiple_empty_notes = false
	end

	-- Setup API layer
	local api = require("neoweaver.api")
	api.setup(opts.api or {})

	-- Setup notes module
	local notes = require("neoweaver.notes")
	notes.setup({
		allow_multiple_empty_notes = config.allow_multiple_empty_notes,
	})

	vim.notify("Neoweaver v3 loaded!", vim.log.levels.INFO)
end

function M.get_config()
	return config
end

return M
