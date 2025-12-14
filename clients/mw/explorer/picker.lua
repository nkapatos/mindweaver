---
--- picker.lua - Note picker using vim.ui.select with optional preview
--- Leverages Neovim's picker ecosystem (Telescope, fzf, etc.)
---
--- Architecture Decision:
--- - Use vim.ui.select as primary picker (respects user's picker plugin)
--- - Optional preview window shown alongside picker
--- - Users with Telescope/fzf get their enhanced UI automatically
---
--- Recent changes (2025-11-24):
--- - Refactored to use vim.ui.select instead of custom input
--- - Added optional split preview window
--- - Simpler, more Neovim-idiomatic approach
---
--- Next steps:
--- - Add filter toggles (note_type)
--- - Add project filtering
--- - Connect to real API
---
local M = {}

-- TODO: Preview support removed for now
-- vim.ui.select creates its own floating UI, can't be wrapped in custom layout
-- Telescope/Snacks handle their own preview automatically

--- Format a note for display in vim.ui.select
---@param note table Note from API
---@return string Formatted line
local function format_note_item(note)
	local title = note.title or "(untitled)"
	local project = (note.meta and note.meta.project) or "unknown"
	local type_label = note.note_type == "quicknote" and "Q" or "N"
	
	return string.format("[%s] %-50s  [%s]", type_label, title:sub(1, 50), project)
end

--- Create preview content for a note
---@param note table Note from API
---@return string[] Lines for preview
local function create_preview_lines(note)
	local lines = {}
	
	table.insert(lines, "# " .. (note.title or "(untitled)"))
	table.insert(lines, "")
	table.insert(lines, "**Project:** " .. ((note.meta and note.meta.project) or "unknown"))
	table.insert(lines, "**Type:** " .. (note.note_type or "note"))
	
	if note.meta and note.meta.commit_hash then
		table.insert(lines, "**Commit:** " .. note.meta.commit_hash)
	end
	if note.meta and note.meta.cwd then
		table.insert(lines, "**CWD:** " .. note.meta.cwd)
	end
	
	table.insert(lines, "")
	table.insert(lines, "---")
	table.insert(lines, "")
	
	if note.body and note.body ~= "" then
		local body_lines = vim.split(note.body, "\n")
		for _, line in ipairs(body_lines) do
			table.insert(lines, line)
		end
	else
		table.insert(lines, "(empty note)")
	end
	
	return lines
end

-- Preview functions removed - not compatible with vim.ui.select
-- User's picker plugin (Telescope/Snacks) handles preview automatically

--- Show note picker using vim.ui.select
--- Respects user's picker plugin (Telescope, fzf, etc.)
---@param notes table[] List of notes from API
---@param opts table Options (project, note_type, on_select, show_preview)
function M.show_picker(notes, opts)
	opts = opts or {}
	
	if not notes or #notes == 0 then
		vim.notify("No notes to display", vim.log.levels.INFO)
		return
	end
	
	-- Build prompt with context
	local project_name = opts.project or "all projects"
	local note_type_label = opts.note_type or "all"
	local prompt = string.format("Notes: %s [%s]", project_name, note_type_label)
	
	-- TODO: Preview integration with vim.ui.select
	-- Problem: vim.ui.select creates floating windows that appear on top of everything
	-- Can't wrap them in our custom layout (Snacks/Telescope handle their own UI)
	-- Options:
	--   1. Let Telescope/Snacks handle preview (they already do)
	--   2. Show preview AFTER selection in a split/vsplit
	--   3. Build fully custom picker (like we did before) but that defeats the purpose
	-- For now: Rely on user's picker plugin to handle preview
	
	-- Use Neovim's native picker (enhanced by user's plugins)
	vim.ui.select(notes, {
		prompt = prompt,
		format_item = format_note_item,
		
		-- Note: Telescope/Snacks will override this with their own UI + preview
		-- This is for basic fallback only
		kind = 'mw_notes', -- Custom kind for picker customization
		
	}, function(selected_note, idx)
		if not selected_note then
			-- User cancelled
			return
		end
		
		-- Call user's on_select callback
		if opts.on_select then
			opts.on_select(selected_note)
		else
			vim.notify("Selected: " .. (selected_note.title or selected_note.id), vim.log.levels.INFO)
		end
	end)
end

--- Test function with mock data
function M.test()
	local mock_notes = {
		{
			id = "1",
			title = "Fix authentication bug",
			note_type = "note",
			body = "The JWT token expires too quickly.\n\nNeed to increase TTL from 1 hour to 24 hours in the auth service.\n\nSteps:\n1. Update AUTH_TOKEN_TTL in config\n2. Test with staging\n3. Deploy to production",
			meta = { project = "my-app", commit_hash = "abc123", cwd = "/home/user/my-app" },
		},
		{
			id = "2",
			title = "API design discussion",
			note_type = "note",
			body = "RESTful API design principles:\n\n1. Use proper HTTP verbs (GET, POST, PUT, DELETE)\n2. Resource naming conventions (plural nouns)\n3. Versioning strategy (URL vs headers)\n4. Error responses with proper status codes\n5. Pagination for large datasets",
			meta = { project = "backend-api", cwd = "/home/user/backend" },
		},
		{
			id = "3",
			title = "Meeting notes - Sprint planning",
			note_type = "note",
			body = "Sprint 23 Planning Meeting\n\nAttendees: Alice, Bob, Charlie\n\nDiscussed:\n- User authentication refactor\n- Database migration strategy\n- Performance optimization goals\n\nAction items:\n- Alice: Research migration tools\n- Bob: Profile slow queries\n- Charlie: Document auth flow",
			meta = { project = "my-app" },
		},
		{
			id = "4",
			title = "Quick thought on deployment",
			note_type = "quicknote",
			body = "Maybe we should use blue-green deployment instead of rolling updates?\nLess risk, easier rollback.",
			meta = { project = "my-app", commit_hash = "def456" },
		},
		{
			id = "5",
			title = "Database indexing strategy",
			note_type = "note",
			body = "Need to add indexes on:\n- users.email\n- posts.created_at\n- comments.post_id\n\nExpected performance improvement: 50-70%",
			meta = { project = "backend-api" },
		},
	}
	
	print("Testing MW note picker with vim.ui.select...")
	print("Users with Telescope/fzf will see their enhanced picker")
	print("Basic users will see simple select menu")
	
	M.show_picker(mock_notes, {
		project = "my-app",
		note_type = "all",
		show_preview = true,
		on_select = function(note)
			print("✓ Selected: " .. note.title .. " (ID: " .. note.id .. ")")
			print("  Project: " .. (note.meta.project or "unknown"))
			print("  Type: " .. note.note_type)
		end,
	})
end

return M
