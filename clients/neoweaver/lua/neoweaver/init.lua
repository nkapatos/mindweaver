-- neoweaver.nvim
-- Neovim client for MindWeaver

local M = {}

M.config = {
  -- Default configuration
}

function M.setup(opts)
  M.config = vim.tbl_deep_extend("force", M.config, opts or {})
end

return M
