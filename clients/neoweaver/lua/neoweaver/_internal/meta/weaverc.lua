---
--- weaverc.lua - Project-level configuration loader for MindWeaver
---
--- Loads .weaverc.json from project root to provide project context:
--- - collection_id: Default collection for notes created in this project
--- - note_type_id: Default note type for notes created in this project
--- - project: Project name (overrides auto-detection)
--- - context: Additional project metadata (arbitrary key-value pairs)
---
--- This module is read-only: it loads and caches .weaverc.json but does not modify it.
--- Users must manually edit .weaverc.json files.
---
--- EXPERIMENTAL: This feature is disabled by default. Enable with:
---   require('neoweaver').setup({ metadata = { enabled = true } })
---
---@module 'neoweaver._internal.meta.weaverc'
local M = {}

---@class WeaverConfig
---@field project? string Project name (overrides auto-detection)
---@field collection_id? integer Default collection ID for notes created in this project
---@field note_type_id? integer Default note type ID for notes created in this project
---@field context? table<string, string> Additional project context (arbitrary key-value pairs)
---@field version? string Schema version (e.g., "1.0")

---@class WeaverConfigCache
---@field data WeaverConfig|nil Parsed .weaverc.json data
---@field mtime number|nil File modification time (seconds since epoch)
---@field root_dir string|nil Project root directory

-- Cache for loaded .weaverc.json per project root
-- Structure: { [root_dir] = WeaverConfigCache }
---@type table<string, WeaverConfigCache>
local cache = {}

--- Find project root by walking up directory tree looking for .weaverc.json
--- Falls back to other common project markers if .weaverc.json not found
---
--- Priority order:
--- 1. .weaverc.json (primary marker)
--- 2. .git/ (git repository root)
--- 3. Common project files (package.json, go.mod, Cargo.toml, etc.)
---
--- @param start_dir? string Starting directory (defaults to cwd)
--- @return string|nil root_dir Absolute path to project root, or nil if not found
function M.find_project_root(start_dir)
  start_dir = start_dir or vim.fn.getcwd()
  local seen = {} -- Prevent infinite loops on symlinks

  -- Project markers in priority order
  local markers = {
    ".weaverc.json", -- Primary marker
    ".git", -- Git repository root
    "package.json", -- Node.js/JavaScript projects
    "go.mod", -- Go projects
    "Cargo.toml", -- Rust projects
    "pyproject.toml", -- Python projects
    "composer.json", -- PHP projects
  }

  local function check_dir(dir)
    -- Prevent infinite loops
    if seen[dir] or dir == "/" then
      return nil
    end
    seen[dir] = true

    -- Check each marker
    for _, marker in ipairs(markers) do
      local path = dir .. "/" .. marker
      if vim.uv.fs_stat(path) then
        return dir
      end
    end

    -- Walk up to parent
    local parent = vim.fn.fnamemodify(dir, ":h")
    if parent == dir then
      return nil -- Reached filesystem root
    end

    return check_dir(parent)
  end

  return check_dir(start_dir)
end

--- Load .weaverc.json from project root
--- Returns cached result if file hasn't changed (mtime-based invalidation)
--- Returns nil if metadata extraction is disabled in config
---
--- @param root_dir? string Project root directory (defaults to auto-detected root)
--- @return WeaverConfig|nil config Parsed .weaverc.json, or nil if not found/invalid/disabled
function M.load(root_dir)
  -- Check if metadata extraction is enabled
  local config = require("neoweaver._internal.config")
  if not config.get().metadata.enabled then
    return nil
  end

  -- Auto-detect root if not provided
  root_dir = root_dir or M.find_project_root()

  if not root_dir then
    return nil -- No project root found
  end

  local weaverc_path = root_dir .. "/.weaverc.json"
  local stat = vim.uv.fs_stat(weaverc_path)

  -- File doesn't exist
  if not stat then
    -- Clear cache if file was deleted
    if cache[root_dir] then
      cache[root_dir] = nil
    end
    return nil
  end

  -- Check cache validity
  local cached = cache[root_dir]
  if cached and cached.mtime == stat.mtime.sec then
    return vim.deepcopy(cached.data) -- Return copy to prevent mutation
  end

  -- Load and parse file
  local content = vim.fn.readfile(weaverc_path)
  if not content or #content == 0 then
    vim.notify(string.format("[weaverc] Empty file: %s", weaverc_path), vim.log.levels.WARN)
    return nil
  end

  local content_str = table.concat(content, "\n")
  local ok, decoded = pcall(vim.json.decode, content_str)

  if not ok then
    vim.notify(string.format("[weaverc] Invalid JSON in %s: %s", weaverc_path, tostring(decoded)), vim.log.levels.ERROR)
    return nil
  end

  -- Validate decoded data is a table
  if type(decoded) ~= "table" then
    vim.notify(
      string.format("[weaverc] Expected object in %s, got %s", weaverc_path, type(decoded)),
      vim.log.levels.ERROR
    )
    return nil
  end

  -- Update cache
  cache[root_dir] = {
    data = decoded,
    mtime = stat.mtime.sec,
    root_dir = root_dir,
  }

  return vim.deepcopy(decoded)
end

--- Clear cache for a specific project root or all cached data
--- Useful for testing or manual refresh
---
--- @param root_dir? string Project root to clear (if nil, clears all cache)
function M.clear_cache(root_dir)
  if root_dir then
    cache[root_dir] = nil
  else
    cache = {}
  end
end

--- Get current cache state (for debugging/inspection)
--- @return table<string, WeaverConfigCache> Current cache contents
function M._get_cache()
  return cache
end

return M
