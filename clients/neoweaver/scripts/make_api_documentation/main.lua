--- Generate API documentation for neoweaver using mini.doc
---
--- This script extracts LuaLS annotations from the neoweaver codebase
--- and generates Vim help documentation files.

local minidoc = require("mini.doc")

--- Get the directory where this script is located
---@return string # Script directory path
local function get_script_directory()
	local path = debug.getinfo(1, "S").source:sub(2) -- Remove '@' prefix
	return path:match("(.*)/")
end

--- Main documentation generation function
local function main()
	local script_dir = get_script_directory()
	local root = vim.fs.normalize(vim.fs.joinpath(script_dir, "..", ".."))

	print("Generating neoweaver API documentation...")
	print("Root directory: " .. root)

	-- Generate API documentation from public modules
	local api_source = vim.fs.joinpath(root, "lua", "neoweaver", "init.lua")
	local api_dest = vim.fs.joinpath(root, "doc", "neoweaver_api.txt")

	minidoc.generate({
		api_source,
	}, api_dest, {
		annotation_extractor = minidoc.default_annotation_extractor,
	})

	print("✓ Generated doc/neoweaver_api.txt")

	-- Generate types documentation
	local types_source = vim.fs.joinpath(root, "lua", "neoweaver", "types.lua")
	local types_dest = vim.fs.joinpath(root, "doc", "neoweaver_types.txt")

	minidoc.generate({
		types_source,
	}, types_dest, {
		annotation_extractor = minidoc.default_annotation_extractor,
	})

	print("✓ Generated doc/neoweaver_types.txt")
	print("")
	print("Documentation generation complete!")
	print("Run :helptags doc to generate tags for :help navigation")
end

main()
