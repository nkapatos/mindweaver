---
--- metadata.lua
--- Extract project metadata from common root marker files (package.json, pyproject.toml, Chart.yaml, etc.)
--- Part of MW Neovim plugin
---
--- Recent changes (2025-11-24):
--- - Removed debug print statements
--- - Cleaned up production logging
--- - Added caching mechanism with mtime-based invalidation
--- - Removed explicit 'type' field - now auto-detected from file extension
--- - Simplified: type detection based on extension (.json, .yaml, .toml) or filename (go.mod)
--- - Returns nil for unsupported file types (skips parsing)
---
--- Supported formats:
--- - JSON (.json, .jsonc)
--- - YAML (.yaml, .yml)
--- - TOML (.toml)
--- - Go modules (go.mod)
---
--- Next steps:
--- - Add support for nested .nvmwmeta files merging (see NEXT_TASK.md)
---
local M = {}

-- Cache for extracted metadata with file modification time tracking
local cache = {
	data = nil,
	mtime = {}, -- [marker_name] = mtime_sec
	root_dir = nil,
}

-- Default configuration
M.config = {
	-- List of root markers to look for (order matters — first match wins)
	-- Type is auto-detected from file extension/name - no need to specify
	markers = {
		{ name = "package.json", fields = { "name", "version", "description" } },
		{ name = "deno.json", fields = { "name", "version" } },
		{ name = "deno.jsonc", fields = { "name", "version" } },
		{
			name = "pyproject.toml",
			fields = {
				"project.name",
				"project.version",
				"project.description",
				"tool.poetry.name",
				"tool.poetry.version",
			},
		},
		{
			name = "Cargo.toml",
			fields = { "package.name", "package.version", "package.description" },
		},
		{ name = "Chart.yaml", fields = { "name", "version", "description" } },
		{ name = "pubspec.yaml", fields = { "name", "version", "description" } },
		{ name = "go.mod", fields = { "module" } },
		{ name = ".nvmwmeta.json", fields = {} },
	},

	-- Allow users to register completely custom extractors
	custom_extractors = {}, -- [marker_name] = function(root_dir, full_path, requested_fields) -> table
}

-- Allow user to override config (e.g. from setup{})
function M.setup(user_config)
	M.config = vim.tbl_deep_extend("force", M.config, user_config or {})
	-- Invalidate cache when config changes
	cache.data = nil
	cache.mtime = {}
end

--- Auto-detect filetype from filename using file extension
--- @param filename string The marker filename
--- @return string|nil Detected filetype ("json", "yaml", "toml", "gomod") or nil if unsupported
local function get_marker_type(filename)
	-- Extract file extension
	local ext = filename:match("%.([^%.]+)$")
	
	if ext == "json" or ext == "jsonc" then
		return "json"
	elseif ext == "yaml" or ext == "yml" then
		return "yaml"
	elseif ext == "toml" then
		return "toml"
	end

	-- Special case: go.mod has no extension
	if filename == "go.mod" then
		return "gomod"
	end

	-- Return nil for unsupported types (will skip parsing)
	return nil
end

-- Resolve a possibly dotted key in a table (e.g. "tool.poetry.name" -> value)
local function resolve_dotted(table, key)
	if not key:find(".", 1, true) then
		return table[key]
	end
	local value = table
	for part in key:gmatch("[^%.]+") do
		if type(value) ~= "table" then
			return nil
		end
		value = value[part]
	end
	return value
end

-- Generic extractor for JSON / YAML / TOML using built-in vim.* decoders
local function extract_structured(path, filetype, wanted_fields)
	if M.config.debug_info then
		vim.notify(string.format("[metadata] Extracting fields from %s as %s", path, filetype), vim.log.levels.INFO)
	end
	local content = vim.fn.readfile(path)
	if not content then
		return nil
	end
	content = table.concat(content, "\n")

	local data
	if filetype == "json" then
		local ok, decoded = pcall(vim.json.decode, content)
		if not ok then
			vim.notify(
				string.format("[metadata] Failed to decode JSON in %s: %s", path, tostring(decoded)),
				vim.log.levels.ERROR
			)
			return nil
		end
		data = decoded
	elseif filetype == "yaml" or filetype == "yml" then
		local ok, yaml_mod = pcall(require, "vim.yaml")
		if not ok then
			vim.notify(string.format("[metadata] Could not load vim.yaml for %s", path), vim.log.levels.ERROR)
			return nil
		end
		local ok2, decoded = pcall(yaml_mod.decode, content)
		if not ok2 then
			vim.notify(string.format("[metadata] Failed to decode YAML in %s", path), vim.log.levels.ERROR)
			return nil
		end
		data = decoded
	elseif filetype == "toml" then
		local ok, toml_mod = pcall(require, "vim.toml")
		if not ok then
			vim.notify(string.format("[metadata] Could not load vim.toml for %s", path), vim.log.levels.ERROR)
			return nil
		end
		local ok2, decoded = pcall(toml_mod.decode, content)
		if not ok2 then
			vim.notify(string.format("[metadata] Failed to decode TOML in %s", path), vim.log.levels.ERROR)
			return nil
		end
		data = decoded
	else
		vim.notify(string.format("[metadata] Unsupported filetype '%s' for %s", filetype, path), vim.log.levels.WARN)
		return nil
	end

	if type(data) ~= "table" then
		vim.notify(string.format("[metadata] Decoded data in %s is not a table", path), vim.log.levels.ERROR)
		return nil
	end

	local result = {}
	if #wanted_fields == 0 then
		for k, v in pairs(data) do
			if type(v) == "table" then
				-- flatten simple tables like { name = "foo" }
				v = v.name or v[1] or vim.inspect(v)
			end
			local key = tostring(k):gsub("[^%w_]", "_")
			result[key] = tostring(v)
		end
	else
		for _, field in ipairs(wanted_fields) do
			local raw_value = resolve_dotted(data, field)
			if raw_value ~= nil then
				local value = raw_value
				if type(value) == "table" then
					value = value.name or vim.inspect(value):gsub("\n", " ")
				end
				-- Use a clean key name: "project.name" → "project_name"
				local key = field:gsub("%.", "_"):gsub("[^%w_]", "_")
				result[key] = tostring(value)
			end
		end
	end
	return result
end

-- Special case: go.mod (very small, no need for heavy parser)
local function extract_gomod(path)
	local lines = vim.fn.readfile(path)
	for _, line in ipairs(lines) do
		local mod = line:match("^module%s+([%S]+)")
		if mod then
			return { module = mod }
		end
	end
	return nil
end

-- Find project root by walking upwards looking for any configured marker
local function find_project_root()
	local cwd = vim.fn.getcwd()
	local seen = {} -- prevent infinite loops on symlinks

	local function check(dir)
		if seen[dir] or dir == "/" then
			return nil
		end
		seen[dir] = true

		for _, marker in ipairs(M.config.markers) do
			local full = dir .. "/" .. marker.name
			if vim.uv.fs_stat(full) then
				return dir
			end
		end

		local parent = vim.fn.fnamemodify(dir, ":h")
		if parent == dir then
			return nil
		end
		return check(parent)
	end

	local lsp_root = nil
	if vim.lsp and vim.lsp.buf_is_attached and #vim.lsp.get_clients() > 0 then
		local roots = vim.lsp.buf.list_workspace_folders()
		if roots and #roots > 0 then
			lsp_root = roots[1]
		end
	end

	return lsp_root or check(cwd) or cwd
end

--- Main public function with caching support
--- Caches extracted metadata and only re-parses when marker files change
function M.extract_metadata()
	local root_dir = find_project_root()

	-- Check if cache is valid
	local needs_refresh = false
	
	-- Invalidate if root directory changed
	if cache.root_dir ~= root_dir then
		needs_refresh = true
		cache.root_dir = root_dir
	end

	-- Check if any marker files have been modified
	if not needs_refresh then
		for _, marker in ipairs(M.config.markers) do
			local full_path = root_dir .. "/" .. marker.name
			local stat = vim.uv.fs_stat(full_path)
			
			if stat then
				local current_mtime = stat.mtime.sec
				local cached_mtime = cache.mtime[marker.name]
				
				if not cached_mtime or current_mtime > cached_mtime then
					needs_refresh = true
					cache.mtime[marker.name] = current_mtime
				end
			elseif cache.mtime[marker.name] then
				-- File was deleted
				needs_refresh = true
				cache.mtime[marker.name] = nil
			end
		end
	end

	-- Return cached data if still valid
	if cache.data and not needs_refresh then
		return vim.deepcopy(cache.data)
	end

	-- Extract fresh metadata
	local meta = {
		project = vim.g.quicknote_project or vim.fn.fnamemodify(vim.fn.getcwd(), ":t"),
		cwd = vim.fn.getcwd(),
		commit_hash = vim.fn.systemlist("git rev-parse --short HEAD 2>/dev/null")[1] or nil,
	}

	meta.project_root = root_dir

	for _, marker in ipairs(M.config.markers) do
		local full_path = root_dir .. "/" .. marker.name
		local stat = vim.uv.fs_stat(full_path)
		
		if not stat then
			goto continue
		end

		-- Update mtime cache
		cache.mtime[marker.name] = stat.mtime.sec

		local extracted
		local marker_type = get_marker_type(marker.name)

		if M.config.custom_extractors[marker.name] then
			-- Custom extractor takes precedence
			local custom = M.config.custom_extractors[marker.name]
			local ok, res = pcall(custom, root_dir, full_path, marker.fields)
			extracted = ok and res or nil
		elseif marker_type == "gomod" then
			extracted = extract_gomod(full_path)
		elseif marker_type then
			-- marker_type is non-nil, so we can parse it
			extracted = extract_structured(full_path, marker_type, marker.fields)
		end
		-- If marker_type is nil, skip this marker (unsupported file type)

		if extracted then
			for k, v in pairs(extracted) do
				meta[k] = v
			end
			-- Optional: stop after first successful marker (most specific wins)
			-- Remove the `break` if you want to merge multiple markers
			-- break
		end

		::continue::
	end

	-- Cache the result
	cache.data = vim.deepcopy(meta)
	return meta
end

return M
