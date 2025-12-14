---
--- init.lua - Neoweaver entry point (v3)
---
local M = {}

function M.setup(opts)
	opts = opts or {}
	
	-- Setup API layer
	local api = require("neoweaver.api")
	api.setup(opts.api or {})
	
	-- Setup notes module
	local notes = require("neoweaver.notes")
	notes.setup()
	
	vim.notify("Neoweaver v3 loaded!", vim.log.levels.INFO)
end

return M
