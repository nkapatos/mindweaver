---
--- Minimal, buffer-agnostic diff overlay for Neovim buffers.
---
--- Provides diff hunks, highlights, and navigation for conflict resolution.
--- Designed for use in merge/conflict resolution workflows.
---
--- # Usage
---
--- local diff = require('codefupanda.mw.diff')
--- diff.setup()
--- diff.set_ref_text(bufnr, ref_lines)
--- diff.enable(bufnr)
--- diff.map_keys(bufnr)
---
---@class MinimalDiffOverlay
local M = {}

--- Namespace for line highlights
local ns_id = vim.api.nvim_create_namespace("MinimalDiffOverlay")
--- Namespace for virtual lines (deletions)
local overlay_ns_id = vim.api.nvim_create_namespace("MinimalDiffOverlayVirt")
--- Buffer-local state: stores ref_lines and hunks for each enabled buffer
---@type table<integer, {ref_lines?: string[], hunks?: table, enabled?: boolean}>
local bufstate = {}

--- Setup highlight groups for diff overlays
local function setup_highlight()
  vim.cmd("highlight default link MinimalDiffAdd DiffAdd")
  vim.cmd("highlight default link MinimalDiffChange DiffChange")
  vim.cmd("highlight default link MinimalDiffDelete DiffDelete")
  vim.cmd("highlight default link MinimalDiffText DiffText")
end

--- Clear all diff overlay extmarks from a buffer
---@param bufnr integer
local function clear(bufnr)
  vim.api.nvim_buf_clear_namespace(bufnr, ns_id, 0, -1)
  vim.api.nvim_buf_clear_namespace(bufnr, overlay_ns_id, 0, -1)
end

---@param ref_lines string[]
---@param buf_lines string[]
local function compute_hunks(ref_lines, buf_lines)
  assert(type(ref_lines) == "table", "ref_lines must be a table of lines")
  assert(type(buf_lines) == "table", "buf_lines must be a table of lines")
  local ref_str = table.concat(ref_lines, "\n")
  local buf_str = table.concat(buf_lines, "\n")
  ---@diagnostic disable-next-line: param-type-mismatch
  local diff = vim.diff(ref_str, buf_str, { result_type = "indices", ctxlen = 0, interhunkctxlen = 0 })
  local hunks = {}
  for _, d in ipairs(diff) do
    local n_ref, n_buf = d[2], d[4]
    local htype = n_ref == 0 and "add" or (n_buf == 0 and "delete" or "change")
    table.insert(hunks, {
      type = htype,
      ref_start = d[1],
      ref_count = n_ref,
      buf_start = d[3],
      buf_count = n_buf,
    })
  end
  return hunks
end

local function draw(bufnr)
  clear(bufnr)
  local state = bufstate[bufnr]
  if not state or not state.ref_lines then return end
  local buf_lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local hunks = compute_hunks(state.ref_lines, buf_lines)
  state.hunks = hunks
  for _, h in ipairs(hunks) do
    if h.type == "add" or h.type == "change" then
      -- Highlight local lines
      for l = h.buf_start, h.buf_start + math.max(h.buf_count, 1) - 1 do
        vim.api.nvim_buf_set_extmark(bufnr, ns_id, l - 1, 0, {
          end_row = l,
          hl_group = h.type == "add" and "MinimalDiffAdd" or "MinimalDiffChange",
          hl_eol = true,
        })
      end
      -- Show server (ref) lines as virtual lines if there are more ref lines than buf lines
      local extra = h.ref_count - h.buf_count
      if extra > 0 then
        local virt_lines = {}
        -- Git-style marker (start)
        table.insert(virt_lines, { { (h.type == "add" and "<<<<<<< SERVER ADD" or "<<<<<<< SERVER CHANGE"), "MinimalDiffText" } })
        for i = h.ref_start + h.buf_count, h.ref_start + h.ref_count - 1 do
          table.insert(virt_lines, { { state.ref_lines[i] or "", h.type == "add" and "MinimalDiffAdd" or "MinimalDiffChange" } })
        end
        -- Git-style marker (end)
        table.insert(virt_lines, { { (h.type == "add" and ">>>>>>> END ADD" or ">>>>>>> END CHANGE"), "MinimalDiffText" } })
        local lnum = h.buf_start + h.buf_count - 1
        if lnum < 0 then lnum = 0 end
        vim.api.nvim_buf_set_extmark(bufnr, overlay_ns_id, lnum, 0, {
          virt_lines = virt_lines,
          virt_lines_above = false,
        })
      end
      -- For "change" hunks, also show the server lines as virtual lines above the hunk for clarity
      if h.type == "change" and h.ref_count > 0 then
        local virt_lines = {}
        table.insert(virt_lines, { { "======= SERVER VERSION", "MinimalDiffText" } })
        for i = h.ref_start, h.ref_start + h.ref_count - 1 do
          table.insert(virt_lines, { { state.ref_lines[i] or "", "MinimalDiffChange" } })
        end
        table.insert(virt_lines, { { "======= END SERVER VERSION", "MinimalDiffText" } })
        local lnum = math.max(h.buf_start, 1) - 1
        vim.api.nvim_buf_set_extmark(bufnr, overlay_ns_id, lnum, 0, {
          virt_lines = virt_lines,
          virt_lines_above = true,
        })
      end
    elseif h.type == "delete" then
      local virt_lines = {}
      -- Git-style marker (start)
      table.insert(virt_lines, { { "<<<<<<< SERVER DELETE", "MinimalDiffText" } })
      for i = h.ref_start, h.ref_start + h.ref_count - 1 do
        table.insert(virt_lines, { { state.ref_lines[i] or "", "MinimalDiffDelete" } })
      end
      -- Git-style marker (end)
      table.insert(virt_lines, { { ">>>>>>> END DELETE", "MinimalDiffText" } })
      local lnum = math.max(h.buf_start, 1) - 1
      vim.api.nvim_buf_set_extmark(bufnr, overlay_ns_id, lnum, 0, {
        virt_lines = virt_lines,
        virt_lines_above = true,
      })
    end
  end
end

---
--- Sets up highlight groups for diff overlays. Call once at startup.
function M.setup()
  setup_highlight()
end

---
--- Enables diff overlay for a buffer. Does nothing if buffer is invalid.
--- @param bufnr integer Buffer number
function M.enable(bufnr)
  if not vim.api.nvim_buf_is_valid(bufnr) then return end
  bufstate[bufnr] = bufstate[bufnr] or { enabled = true }
  if bufstate[bufnr].ref_lines then
    draw(bufnr)
  end
end

---
--- Disables and clears the diff overlay for a buffer.
--- @param bufnr integer Buffer number
function M.disable(bufnr)
  bufstate[bufnr] = nil
  clear(bufnr)
end

---
--- Sets the reference text (lines) for diffing against the buffer.
--- @param bufnr integer Buffer number
--- @param ref_lines string[] Array of lines to use as reference
function M.set_ref_text(bufnr, ref_lines)
  if not vim.api.nvim_buf_is_valid(bufnr) then return end
  if type(ref_lines) == "string" then
    error("set_ref_text: ref_lines must be a table (array of lines), not a string.")
  end
  bufstate[bufnr] = bufstate[bufnr] or {}
  bufstate[bufnr].ref_lines = vim.deepcopy(ref_lines)
  draw(bufnr)
end

local function find_hunk_idx(hunks, line, dir)
  if not hunks or #hunks == 0 then return nil end
  if dir == 1 then
    for i, h in ipairs(hunks) do
      if h.buf_start >= line then return i end
    end
    return 1 -- wrap
  else
    for i = #hunks, 1, -1 do
      local h = hunks[i]
      local hunk_to = h.buf_start + math.max(h.buf_count, 1) - 1
      if hunk_to < line then return i end
    end
    return #hunks -- wrap
  end
end

---
--- Moves the cursor to the start of the next diff hunk in the buffer.
--- @param bufnr integer Buffer number
function M.goto_next_hunk(bufnr)
  local state = bufstate[bufnr]
  if not state or not state.hunks then return end
  local line = vim.api.nvim_win_get_cursor(0)[1]
  local idx = find_hunk_idx(state.hunks, line + 1, 1)
  if idx then
    local h = state.hunks[idx]
    vim.api.nvim_win_set_cursor(0, { h.buf_start, 0 })
  end
end

---
--- Moves the cursor to the start of the previous diff hunk in the buffer.
--- @param bufnr integer Buffer number
function M.goto_prev_hunk(bufnr)
  local state = bufstate[bufnr]
  if not state or not state.hunks then return end
  local line = vim.api.nvim_win_get_cursor(0)[1]
  local idx = find_hunk_idx(state.hunks, line, -1)
  if idx then
    local h = state.hunks[idx]
    vim.api.nvim_win_set_cursor(0, { h.buf_start, 0 })
  end
end

---
--- Applies the diff hunk under the cursor to the buffer.
--- For 'add'/'change', replaces lines with reference; for 'delete', inserts deleted lines.
--- @param bufnr integer Buffer number
function M.apply_hunk(bufnr)
  local state = bufstate[bufnr]
  if not state or not state.hunks then return end
  local line = vim.api.nvim_win_get_cursor(0)[1]
  for _, h in ipairs(state.hunks) do
    local from = h.buf_start
    local to = h.buf_start + math.max(h.buf_count, 1) - 1
    if line >= from and line <= to then
      if h.type == "add" or h.type == "change" then
        local new_lines = {}
        for i = h.ref_start, h.ref_start + h.ref_count - 1 do
          table.insert(new_lines, state.ref_lines[i] or "")
        end
        vim.api.nvim_buf_set_lines(bufnr, from - 1, to, false, new_lines)
      elseif h.type == "delete" then
        local new_lines = {}
        for i = h.ref_start, h.ref_start + h.ref_count - 1 do
          table.insert(new_lines, state.ref_lines[i] or "")
        end
        vim.api.nvim_buf_set_lines(bufnr, from - 1, from - 1, false, new_lines)
      end
      break
    end
  end
  draw(bufnr)
end

---
--- Sets up buffer-local keymaps for diff navigation and application.
--- [c / ]c: prev/next hunk, gh: apply hunk
--- @param bufnr integer Buffer number
function M.map_keys(bufnr)
  vim.keymap.set('n', ']c', function() M.goto_next_hunk(bufnr) end, { buffer = bufnr, desc = 'Next diff hunk' })
  vim.keymap.set('n', '[c', function() M.goto_prev_hunk(bufnr) end, { buffer = bufnr, desc = 'Prev diff hunk' })
  vim.keymap.set('n', 'gh', function() M.apply_hunk(bufnr) end, { buffer = bufnr, desc = 'Apply diff hunk' })
end

---
--- Returns true if the buffer has unresolved hunks (for save blocking)
---
function M.has_unresolved_hunks(bufnr)
  local state = bufstate[bufnr]
  return state and state.hunks and #state.hunks > 0
end

return M
