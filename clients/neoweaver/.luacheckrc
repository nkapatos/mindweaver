-- Luacheck configuration for neoweaver
-- https://luacheck.readthedocs.io/en/stable/config.html

cache = true
std = "luajit"
codes = true

-- Allow vim.wo.spell = true style assignments
ignore = {
  "122", -- Setting read-only field of global
}

-- Neovim global
read_globals = {
  "vim",
}

-- Exclude generated files and dependencies
exclude_files = {
  ".luarocks/",
  ".dependencies/",
}

-- Allow longer lines (matching stylua config)
max_line_length = 120
max_code_line_length = 120
max_comment_line_length = 120
