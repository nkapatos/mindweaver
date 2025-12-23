---
--- extractor.lua - Multi-format project metadata extraction
---
--- Extracts metadata from common project files (package.json, pyproject.toml, etc.)
--- Supports JSON, YAML, TOML, and special formats like go.mod
---
--- Migrated from clients/mw/metadata.lua
---
--- EXPERIMENTAL: This feature is disabled by default. Enable with:
---   require('neoweaver').setup({ metadata = { enabled = true } })
---
---@module 'neoweaver._internal.meta.extractor'
local M = {}

---@class ExtractorConfig
---@field markers table[] List of marker files to scan with their field configurations
---@field custom_extractors table<string, function> Custom extractor functions per marker

---@class ProjectMetadata
---@field project? string Project name
---@field cwd? string Current working directory
---@field commit_hash? string Git commit hash
---@field project_root? string Detected project root directory
---@field [string] any Additional extracted fields from project files

---@class ExtractorCache
---@field data ProjectMetadata|nil Cached metadata
---@field mtime table<string, number> File modification times per marker
---@field root_dir string|nil Cached project root

-- Cache for extracted metadata with file modification time tracking
---@type ExtractorCache
local cache = {
  data = nil,
  mtime = {},
  root_dir = nil,
}

-- Default configuration
-- Note: Configuration via setup() not implemented - See issue #47
M.config = {
  -- List of root markers to look for (order matters — first match wins for root detection)
  -- Type is auto-detected from file extension/name - no need to specify
  markers = {
    { name = ".weaverc.json", fields = {} }, -- Explicit config (handled by weaverc.lua)
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
  },

  -- Note: Custom extractors not implemented - See issue #47
  -- Would allow users to register custom extraction functions
  -- custom_extractors = {}, -- [marker_name] = function(root_dir, full_path, fields) -> table
}

-- Note: Setup function not implemented - See issue #47
-- Would allow user to override config from setup()
-- function M.setup(user_config)
--   M.config = vim.tbl_deep_extend("force", M.config, user_config or {})
--   -- Invalidate cache when config changes
--   cache.data = nil
--   cache.mtime = {}
-- end

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

--- Resolve a possibly dotted key in a table (e.g. "tool.poetry.name" -> value)
--- @param tbl table The table to search
--- @param key string The key (possibly dotted like "project.name")
--- @return any|nil The resolved value or nil if not found
local function resolve_dotted(tbl, key)
  if not key:find(".", 1, true) then
    return tbl[key]
  end
  local value = tbl
  for part in key:gmatch("[^%.]+") do
    if type(value) ~= "table" then
      return nil
    end
    value = value[part]
  end
  return value
end

--- Generic extractor for JSON / YAML / TOML using built-in vim.* decoders
--- @param path string Absolute path to the file
--- @param filetype string File type ("json", "yaml", "toml")
--- @param wanted_fields string[] List of field names to extract (empty = extract all)
--- @return table|nil Extracted key-value pairs or nil on error
local function extract_structured(path, filetype, wanted_fields)
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

--- Special case: go.mod (very small, no need for heavy parser)
--- @param path string Absolute path to go.mod file
--- @return table|nil Extracted module name or nil
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

--- Find project root by walking upwards looking for any configured marker
--- @param start_dir? string Starting directory (defaults to cwd)
--- @return string|nil root_dir Absolute path to project root or nil
local function find_project_root(start_dir)
  start_dir = start_dir or vim.fn.getcwd()
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

  -- Note: LSP workspace integration not implemented - See issue #50
  -- File-based detection preferred for simplicity

  return check(start_dir) or start_dir
end

--- Main extraction function with caching support
--- Caches extracted metadata and only re-parses when marker files change
--- Returns nil if metadata extraction is disabled in config
---
--- @param root_dir? string Project root directory (defaults to auto-detected)
--- @return ProjectMetadata|nil Extracted metadata or nil if disabled/not found
function M.extract_metadata(root_dir)
  -- Check if metadata extraction is enabled
  local config = require("neoweaver._internal.config")
  if not config.get().metadata.enabled then
    return nil
  end

  root_dir = root_dir or find_project_root()

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
  -- Note: Git integration evaluation needed - See issue #51
  -- Currently extracts commit hash, usefulness unclear
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

    -- Note: Custom extractors not implemented - See issue #47
    -- Would allow user-registered parsing functions
    if marker_type == "gomod" then
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
      -- Note: Merge strategy needs discussion - See issue #48
      -- Currently merges all markers, could stop at first match
      -- break
    end

    ::continue::
  end

  -- Cache the result
  cache.data = vim.deepcopy(meta)
  return meta
end

--- Clear cache for all cached metadata
--- Useful for testing or manual refresh
function M.clear_cache()
  cache.data = nil
  cache.mtime = {}
  cache.root_dir = nil
end

--- Get current cache state (for debugging/inspection)
--- @return ExtractorCache Current cache contents
function M._get_cache()
  return cache
end

return M
