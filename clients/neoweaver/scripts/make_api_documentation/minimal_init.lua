--- Bootstrap mini.doc for documentation generation
--- This script sets up a minimal Neovim environment with mini.nvim installed

local root = vim.fs.dirname(vim.fn.tempname())
vim.env.XDG_DATA_HOME = root .. "/data"

-- Clone mini.nvim (contains mini.doc)
local minidoc_path = root .. "/plugins/mini.nvim"
if not vim.uv.fs_stat(minidoc_path) then
	print("Cloning mini.nvim...")
	vim.fn.system({
		"git",
		"clone",
		"--filter=blob:none",
		"--branch",
		"stable",
		"https://github.com/echasnovski/mini.nvim.git",
		minidoc_path,
	})
end

vim.opt.runtimepath:prepend(minidoc_path)

-- Add neoweaver to runtimepath
local script_dir = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.expand("<sfile>:p")), ":h")
local neoweaver_root = vim.fs.dirname(vim.fs.dirname(script_dir))
vim.opt.runtimepath:prepend(neoweaver_root)

print("Bootstrap complete. mini.doc loaded from: " .. minidoc_path)
