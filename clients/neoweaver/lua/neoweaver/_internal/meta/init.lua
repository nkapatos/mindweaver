---
--- init.lua - Public API for metadata extraction module
---
--- Main entry point for loading project metadata from .weaverc.json
--- and auto-detected project files (package.json, pyproject.toml, etc.)
---
--- EXPERIMENTAL: This feature is disabled by default. Enable with:
---   require('neoweaver').setup({ metadata = { enabled = true } })
---
--- Usage:
---   local meta = require('neoweaver._internal.meta')
---   local data = meta.load() -- Load complete metadata (weaverc + auto-detected)
---
---@module 'neoweaver._internal.meta'
local M = {}

local weaverc = require("neoweaver._internal.meta.weaverc")
local extractor = require("neoweaver._internal.meta.extractor")

---@class MergedMetadata
---@field weaverc? WeaverConfig Data from .weaverc.json
---@field project? ProjectMetadata Auto-detected project metadata
---@field [string] any Merged fields from both sources

--- Load complete metadata (weaverc + auto-detected project metadata)
--- Returns nil if metadata extraction is disabled in config
---
--- Note: Merge strategy needs discussion - See issue #48
--- - Currently returns separate namespaces (weaverc, project)
--- - Could flatten or use priority rules
---
--- @param root_dir? string Project root directory (defaults to auto-detected)
--- @return MergedMetadata|nil Complete metadata or nil if disabled
function M.load(root_dir)
  -- Check if metadata extraction is enabled
  local config = require("neoweaver._internal.config")
  if not config.get().metadata.enabled then
    return nil
  end

  -- Load both sources
  local weaverc_data = weaverc.load(root_dir)
  local project_data = extractor.extract_metadata(root_dir)

  -- Note: Using separate namespaces - See issue #48
  -- Avoids field conflicts, consumers decide how to handle overlaps
  return {
    weaverc = weaverc_data,
    project = project_data,
  }
end

--- Load only .weaverc.json configuration
--- @param root_dir? string Project root directory (defaults to auto-detected)
--- @return WeaverConfig|nil Weaverc data or nil if disabled/not found
function M.load_weaverc(root_dir)
  return weaverc.load(root_dir)
end

--- Load only auto-detected project metadata
--- @param root_dir? string Project root directory (defaults to auto-detected)
--- @return ProjectMetadata|nil Project metadata or nil if disabled/not found
function M.load_project(root_dir)
  return extractor.extract_metadata(root_dir)
end

--- Find project root directory
--- @param start_dir? string Starting directory (defaults to cwd)
--- @return string|nil root_dir Absolute path to project root or nil
function M.find_project_root(start_dir)
  return weaverc.find_project_root(start_dir)
end

--- Clear all metadata caches
--- Useful for testing or manual refresh
function M.clear_cache()
  weaverc.clear_cache()
  extractor.clear_cache()
end

--- Get current cache state (for debugging/inspection)
--- @return table Cache contents from both weaverc and extractor
function M._get_cache()
  return {
    weaverc = weaverc._get_cache(),
    extractor = extractor._get_cache(),
  }
end

return M
