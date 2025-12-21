--- Minimal Neovim environment for testing
---
--- Sets up mini.nvim (specifically mini.test) for running tests.
---
--- Dependencies managed via Task (see tasks/tasks.neoweaver.yml)
--- Run `task neoweaver:deps` to download mini.nvim to deps/mini.nvim
---
--- Usage:
---   nvim -u scripts/minimal_init_test.lua -l scripts/run_tests.lua

-- Get the project root (two levels up from scripts/)
local script_dir = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.expand("<sfile>:p")), ":h")
local project_root = vim.fs.dirname(script_dir)

-- Path to mini.nvim managed by Task
local mini_path = vim.fs.joinpath(project_root, "deps", "mini.nvim")

-- Verify mini.nvim exists (should be downloaded via task neoweaver:deps)
if not vim.uv.fs_stat(mini_path) then
	error(
		"mini.nvim not found at: "
			.. mini_path
			.. "\n"
			.. "Please run: task neoweaver:deps"
	)
end

-- Add mini.nvim to runtimepath for mini.test
vim.opt.runtimepath:prepend(mini_path)

-- Add neoweaver to runtimepath so we can test it
vim.opt.runtimepath:prepend(project_root)

print("✓ Minimal init (test) loaded")
print("  mini.nvim: " .. mini_path)
print("  neoweaver: " .. project_root)
