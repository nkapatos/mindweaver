---
--- health.lua - Health check for neoweaver plugin
--- Provides diagnostics via :checkhealth neoweaver
---
local M = {}

--- Run checkhealth diagnostics
function M.check()
  vim.health.start("neoweaver")

  -- Check Neovim version
  if vim.fn.has("nvim-0.11") == 1 then
    vim.health.ok("Neovim >= 0.11.0")
  else
    local version = vim.version()
    vim.health.error(
      string.format("Neovim 0.11+ required, found: %d.%d.%d", version.major, version.minor, version.patch)
    )
  end

  -- Check plenary.nvim
  local has_plenary, _ = pcall(require, "plenary.curl")
  if has_plenary then
    vim.health.ok("plenary.nvim is installed")
  else
    vim.health.error("plenary.nvim is required but not found", {
      "Install via your plugin manager:",
      "  { 'nvim-lua/plenary.nvim' }",
    })
  end

  -- Check nui.nvim
  local has_nui, _ = pcall(require, "nui.split")
  if has_nui then
    vim.health.ok("nui.nvim is installed")
  else
    vim.health.error("nui.nvim is required but not found", {
      "Install via your plugin manager:",
      "  { 'MunifTanjim/nui.nvim' }",
    })
  end

  -- Check if plugin is configured
  local config_ok, config = pcall(require, "neoweaver._internal.config")
  if not config_ok then
    vim.health.warn("Plugin not loaded yet (call setup() first)")
    return
  end

  -- Check server configuration
  local api_ok, api = pcall(require, "neoweaver._internal.api")
  if api_ok then
    if api.config.current_server and api.config.current_server ~= "" then
      local server_url = api.config.servers[api.config.current_server]
      if server_url then
        vim.health.ok(string.format("Server configured: %s (%s)", api.config.current_server, server_url.url))
      end
    else
      vim.health.warn("No server selected", {
        "Use :NeoweaverServerUse <name> to select a server",
        "Or set default = true in your setup config",
      })
    end

    -- List available servers
    if api.config.servers and next(api.config.servers) ~= nil then
      local server_names = {}
      for name, _ in pairs(api.config.servers) do
        table.insert(server_names, name)
      end
      vim.health.info("Available servers: " .. table.concat(server_names, ", "))
    end
  end

  -- Check configuration
  local current_config = config.get()
  if current_config.keymaps.enabled then
    vim.health.info("Keymaps are enabled")
  else
    vim.health.info("Keymaps are disabled (opt-in via config)")
  end
end

return M
