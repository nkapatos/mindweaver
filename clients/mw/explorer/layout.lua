---
--- layout.lua - Window and buffer layout management for note explorer
--- Creates a 3-pane layout: input (top), list (bottom-left), preview (bottom-right)
---
--- Recent changes (2025-11-24):
--- - Initial implementation with fixed layout
---
--- Next steps:
--- - Add configurable layout options
--- - Add dynamic resizing support
---
local M = {}

--- Create the explorer layout with 3 panes
--- Layout: Input top, List bottom-left, Preview bottom-right
---@return table Layout state with buffers and windows
function M.create_layout()
	local width = vim.o.columns
	local height = vim.o.lines - 2 -- Leave room for statusline/cmdline
	
	-- Calculate dimensions
	local input_height = 1
	local list_width = math.floor(width * 0.5)
	local preview_width = width - list_width
	local content_height = height - input_height - 1
	
	-- Create buffers
	local input_buf = vim.api.nvim_create_buf(false, true)
	local list_buf = vim.api.nvim_create_buf(false, true)
	local preview_buf = vim.api.nvim_create_buf(false, true)
	
	-- Set buffer options
	vim.api.nvim_buf_set_option(input_buf, 'buftype', 'prompt')
	vim.api.nvim_buf_set_option(input_buf, 'bufhidden', 'wipe')
	
	vim.api.nvim_buf_set_option(list_buf, 'buftype', 'nofile')
	vim.api.nvim_buf_set_option(list_buf, 'bufhidden', 'wipe')
	vim.api.nvim_buf_set_option(list_buf, 'modifiable', false)
	
	vim.api.nvim_buf_set_option(preview_buf, 'buftype', 'nofile')
	vim.api.nvim_buf_set_option(preview_buf, 'bufhidden', 'wipe')
	vim.api.nvim_buf_set_option(preview_buf, 'modifiable', false)
	vim.api.nvim_buf_set_option(preview_buf, 'filetype', 'markdown')
	
	-- Create windows
	-- Input window (top)
	local input_win = vim.api.nvim_open_win(input_buf, true, {
		relative = 'editor',
		width = width,
		height = input_height,
		row = 0,
		col = 0,
		style = 'minimal',
		border = 'rounded',
		title = ' Search Notes ',
		title_pos = 'center',
	})
	
	-- List window (bottom-left)
	local list_win = vim.api.nvim_open_win(list_buf, false, {
		relative = 'editor',
		width = list_width,
		height = content_height,
		row = input_height + 1,
		col = 0,
		style = 'minimal',
		border = 'rounded',
		title = ' Notes ',
		title_pos = 'left',
	})
	
	-- Preview window (bottom-right)
	-- Add 1 column offset to account for list border
	local preview_win = vim.api.nvim_open_win(preview_buf, false, {
		relative = 'editor',
		width = preview_width - 1,
		height = content_height,
		row = input_height + 1,
		col = list_width + 1,
		style = 'minimal',
		border = 'rounded',
		title = ' Preview ',
		title_pos = 'left',
	})
	
	-- Set prompt for input buffer
	vim.fn.prompt_setprompt(input_buf, '> ')
	
	return {
		input = { buf = input_buf, win = input_win },
		list = { buf = list_buf, win = list_win },
		preview = { buf = preview_buf, win = preview_win },
	}
end

--- Close all windows in the layout
---@param layout table Layout state from create_layout()
function M.close_layout(layout)
	local windows = { layout.input.win, layout.list.win, layout.preview.win }
	for _, win in ipairs(windows) do
		if vim.api.nvim_win_is_valid(win) then
			vim.api.nvim_win_close(win, true)
		end
	end
end

--- Test function to verify layout
function M.test()
	local layout = M.create_layout()
	
	-- Add some test content
	vim.api.nvim_buf_set_option(layout.list.buf, 'modifiable', true)
	vim.api.nvim_buf_set_lines(layout.list.buf, 0, -1, false, {
		'> Fix auth bug',
		'  API design notes',
		'  Meeting notes',
	})
	vim.api.nvim_buf_set_option(layout.list.buf, 'modifiable', false)
	
	vim.api.nvim_buf_set_option(layout.preview.buf, 'modifiable', true)
	vim.api.nvim_buf_set_lines(layout.preview.buf, 0, -1, false, {
		'# Fix auth bug',
		'',
		'**Project:** my-app',
		'**Type:** note',
		'',
		'---',
		'',
		'The JWT token expires too quickly.',
		'',
		'Need to increase TTL from 1 hour to 24 hours.',
	})
	vim.api.nvim_buf_set_option(layout.preview.buf, 'modifiable', false)
	
	-- Setup close keymap
	vim.keymap.set('n', 'q', function()
		M.close_layout(layout)
	end, { buffer = layout.input.buf })
	
	print('Layout test created. Press q to close.')
	return layout
end

return M
