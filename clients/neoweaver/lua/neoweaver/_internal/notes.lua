---
--- notes.lua - Note management for Neoweaver (v3)
--- Handles note listing, opening, editing, and saving
---
--- Reference: clients/mw/notes.lua (v1 implementation)
---
local api = require("neoweaver._internal.api")
local buffer_manager = require("neoweaver._internal.buffer.manager")
local diff = require("neoweaver._internal.diff")
local picker = require("neoweaver._internal.ui.picker")
local config_module = require("neoweaver._internal.config")

-- Debounce state for create_note
local last_create_time = 0
local DEBOUNCE_MS = 500

local allow_multiple_empty_notes = false

-- Note: ConnectCode enum temporary - See issue #53
local ConnectCode = {
  OK = "ok",
  CANCELLED = "cancelled",
  UNKNOWN = "unknown",
  INVALID_ARGUMENT = "invalid_argument",
  DEADLINE_EXCEEDED = "deadline_exceeded",
  NOT_FOUND = "not_found",
  ALREADY_EXISTS = "already_exists",
  PERMISSION_DENIED = "permission_denied",
  RESOURCE_EXHAUSTED = "resource_exhausted",
  FAILED_PRECONDITION = "failed_precondition", -- Your ETag case
  ABORTED = "aborted",
  OUT_OF_RANGE = "out_of_range",
  UNIMPLEMENTED = "unimplemented",
  INTERNAL = "internal",
  UNAVAILABLE = "unavailable",
  DATA_LOSS = "data_loss",
  UNAUTHENTICATED = "unauthenticated",
}

local M = {}

-- Forward declarations
local handle_conflict

--- Handle ETag conflict by showing diff and resolving
---@param bufnr integer
---@param note_id integer
handle_conflict = function(bufnr, note_id)
  vim.notify("ETag conflict detected - fetching latest version from server...", vim.log.levels.WARN)

  ---@type mind.v3.GetNoteRequest
  local req = { id = note_id }

  api.notes.get(req, function(res)
    if res.error then
      vim.notify("Failed to fetch latest note: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    ---@type mind.v3.Note
    local latest_note = res.data
    local server_lines = vim.split(latest_note.body or "", "\n")

    vim.defer_fn(function()
      local conflict_count = diff.get_conflict_count(bufnr)
      local msg = table.concat({
        string.format("⚠️  %d conflict%s detected", conflict_count, conflict_count > 1 and "s" or ""),
        "Resolve: gh (server) | gl (local) | gb (both)",
        "Navigate: ]c (next) | [c (prev)",
        "Save: :w (retry)",
      }, "\n")
      vim.notify(msg, vim.log.levels.WARN)
    end, 100)

    if not vim.api.nvim_get_option_value("modifiable", { buf = bufnr }) then
      vim.api.nvim_set_option_value("modifiable", true, { buf = bufnr })
    end

    diff.enable(bufnr)
    diff.map_keys(bufnr)

    vim.defer_fn(function()
      local ok, err = pcall(diff.set_ref_text, bufnr, server_lines)
      if not ok then
        vim.notify("Failed to enable diff overlay: " .. tostring(err), vim.log.levels.ERROR)
        return
      end
    end, 50)

    local group = vim.api.nvim_create_augroup("neoweaver_conflict_" .. bufnr, { clear = true })
    vim.api.nvim_create_autocmd("BufWriteCmd", {
      group = group,
      buffer = bufnr,
      once = true,
      callback = function()
        local remaining = diff.get_conflict_count(bufnr)
        if remaining > 0 then
          vim.notify(
            string.format("⚠️  Saving with %d unresolved conflict%s", remaining, remaining > 1 and "s" or ""),
            vim.log.levels.WARN
          )
        else
          vim.notify("✓ All conflicts resolved - saving", vim.log.levels.INFO)
        end

        diff.disable(bufnr)
        vim.api.nvim_del_augroup_by_id(group)

        vim.b[bufnr].note_etag = latest_note.etag
        M.save_note(bufnr, note_id)
      end,
    })
  end)
end

--- List all notes using nui picker
function M.list_notes()
  ---@type mind.v3.ListNotesRequest
  local req = {
    pageSize = 100,
    pageToken = "",
  }

  api.notes.list(req, function(res)
    if res.error then
      vim.notify("Error listing notes: " .. vim.inspect(res.error), vim.log.levels.ERROR)
      return
    end

    ---@type mind.v3.ListNotesResponse
    local list_res = res.data
    local notes = list_res.notes or {}

    if #notes == 0 then
      vim.notify("No notes found!", vim.log.levels.INFO)
      return
    end

    local cfg = config_module.get().picker or {}

    picker.pick(notes, {
      prompt = "Select a note",
      format_item = function(note, _idx)
        return string.format("[%d] %s", note.id, note.title)
      end,
      on_submit = function(note, _idx)
        M.open_note(tonumber(note.id))
      end,
      size = cfg.size,
      position = cfg.position,
      border = cfg.border,
    })
  end)
end

--- Create a new note (server-first approach with auto-generated title)
function M.create_note()
  local now = vim.loop.now()
  if not allow_multiple_empty_notes and (now - last_create_time < DEBOUNCE_MS) then
    vim.notify("Please wait before creating another note", vim.log.levels.WARN)
    return
  end
  last_create_time = now

  local collection_id = 1
  local current_buf = vim.api.nvim_get_current_buf()
  local base_entity = buffer_manager.get_entity(current_buf)
  if base_entity and base_entity.type == "note" then
    collection_id = vim.b[current_buf].note_collection_id or collection_id
  end

  ---@type mind.v3.NewNoteRequest
  local req = {
    collectionId = collection_id,
  }

  api.notes.new(req, function(res)
    if res.error then
      vim.notify("Failed to create note: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    local note = res.data
    local note_id = tonumber(note.id)

    local bufnr = buffer_manager.create({
      type = "note",
      id = note_id,
      name = note.title,
      filetype = "markdown",
      modifiable = true,
    })

    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, {})
    vim.api.nvim_set_option_value("modified", allow_multiple_empty_notes, { buf = bufnr })

    vim.b[bufnr].note_id = note_id
    vim.b[bufnr].note_title = note.title
    vim.b[bufnr].note_etag = note.etag
    vim.b[bufnr].note_collection_id = note.collectionId
    vim.b[bufnr].note_type_id = note.noteTypeId
    vim.b[bufnr].note_metadata = note.metadata or {}

    vim.notify("Note created: " .. note.title, vim.log.levels.INFO)
  end)
end

--- Open a note in a buffer for editing
---@param note_id integer
function M.open_note(note_id)
  if not note_id then
    vim.notify("Invalid note ID", vim.log.levels.ERROR)
    return
  end

  local existing = buffer_manager.get("note", note_id)
  if existing and vim.api.nvim_buf_is_valid(existing) then
    buffer_manager.switch_to_buffer(existing)
    return
  end

  ---@type mind.v3.GetNoteRequest
  local req = { id = note_id }

  api.notes.get(req, function(res)
    if res.error then
      vim.notify("Error loading note: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    local note = res.data

    local bufnr = buffer_manager.create({
      type = "note",
      id = note_id,
      name = note.title or "Untitled",
      filetype = "markdown",
      modifiable = true,
    })

    local lines = vim.split(note.body or "", "\n")
    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
    vim.api.nvim_set_option_value("modified", false, { buf = bufnr })

    vim.b[bufnr].note_id = note_id
    vim.b[bufnr].note_title = note.title
    vim.b[bufnr].note_etag = note.etag
    vim.b[bufnr].note_collection_id = note.collectionId
    vim.b[bufnr].note_type_id = note.noteTypeId
    vim.b[bufnr].note_metadata = note.metadata or {}
  end)
end

--- Edit note metadata (placeholder)
function M.edit_metadata()
  vim.notify("Metadata editing not yet implemented in v3 - See issue #44", vim.log.levels.WARN)
end

--- Edit the current note title and persist immediately
function M.edit_title()
  local bufnr = vim.api.nvim_get_current_buf()
  local entity = buffer_manager.get_entity(bufnr)

  if not entity or entity.type ~= "note" then
    vim.notify("Current buffer is not a managed note", vim.log.levels.WARN)
    return
  end

  local note_id = entity.id
  local current_title = vim.b[bufnr].note_title or "Untitled"

  vim.ui.input({
    prompt = "Edit Title:",
    default = current_title,
  }, function(input)
    if input == nil then
      return
    end

    local new_title = vim.trim(input)

    if new_title == "" then
      vim.notify("Title cannot be empty", vim.log.levels.WARN)
      return
    end

    if new_title == current_title then
      return
    end

    vim.b[bufnr].note_title = new_title
    vim.api.nvim_buf_set_name(bufnr, new_title)
    M.save_note(bufnr, note_id)
  end)
end

--- Save note buffer content to server
---@param bufnr integer
---@param id integer
function M.save_note(bufnr, id)
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local body = table.concat(lines, "\n")

  local title = vim.b[bufnr].note_title or "Untitled"
  local etag = vim.b[bufnr].note_etag
  local collection_id = vim.b[bufnr].note_collection_id or 1
  local note_type_id = vim.b[bufnr].note_type_id
  local metadata = vim.b[bufnr].note_metadata or {}

  ---@type mind.v3.ReplaceNoteRequest
  local req = {
    id = id,
    title = title,
    body = body,
    collectionId = collection_id,
  }

  if note_type_id then
    req.noteTypeId = note_type_id
  end

  if metadata and next(metadata) ~= nil then
    req.metadata = metadata
  end

  api.notes.update(req, etag, function(res)
    if res.error then
      if res.error.code == ConnectCode.FAILED_PRECONDITION then
        handle_conflict(bufnr, id)
        return
      end

      vim.notify("Save failed: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    local updated_note = res.data
    vim.b[bufnr].note_etag = updated_note.etag
    vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
    vim.notify("Note saved successfully", vim.log.levels.INFO)
  end)
end

function M.delete_note(note_id)
  if not note_id then
    vim.notify("Invalid note ID", vim.log.levels.ERROR)
    return
  end

  vim.ui.input({
    prompt = string.format("Delete note %d? (y/N): ", note_id),
  }, function(input)
    if not input or (input:lower() ~= "y" and input:lower() ~= "yes") then
      vim.notify("Delete cancelled", vim.log.levels.INFO)
      return
    end

    ---@type mind.v3.DeleteNoteRequest
    local req = { id = note_id }

    api.notes.delete(req, function(res)
      if res.error then
        vim.notify("Failed to delete note: " .. res.error.message, vim.log.levels.ERROR)
        return
      end

      local bufnr = buffer_manager.get("note", note_id)
      if bufnr and vim.api.nvim_buf_is_valid(bufnr) then
        vim.api.nvim_buf_delete(bufnr, { force = true })
      end

      vim.notify("Note deleted successfully", vim.log.levels.INFO)
    end)
  end)
end

--- Find notes by title using interactive search picker
function M.find_notes()
  local search_picker = require("neoweaver._internal.ui.search_picker")

  --- Search function - calls FindNotes API
  ---@param query string Search query
  ---@param page_token string|nil Pagination token
  ---@param callback function Callback(items, error, has_more, next_token)
  local function search_fn(query, page_token, callback)
    ---@type mind.v3.FindNotesRequest
    local req = {
      title = query,
      pageSize = 100,
      pageToken = page_token,
      fieldMask = "id,title,collectionId,collectionPath",
    }

    api.notes.find(req, function(res)
      if res.error then
        callback(nil, res.error.message, false, nil)
        return
      end

      ---@type mind.v3.FindNotesResponse
      local find_res = res.data
      local notes = find_res.notes or {}
      local has_more = find_res.nextPageToken ~= nil and find_res.nextPageToken ~= ""

      callback(notes, nil, has_more, find_res.nextPageToken)
    end)
  end

  -- Show search picker
  search_picker.show({
    prompt = "Find notes:",
    min_query_length = 3,
    debounce_ms = 300,
    empty_message = "No notes found",
    search_fn = search_fn,
    format_item = function(note, _idx)
      -- Format: "Note Title          collection-path"
      local title = note.title or "Untitled"
      local path = note.collectionPath or ""
      return string.format("%-40s %s", title, path)
    end,
    on_select = function(note, _idx)
      M.open_note(tonumber(note.id))
    end,
    on_close = function()
      -- Optional: handle picker close
    end,
  })
end

function M.setup(opts)
  opts = opts or {}
  allow_multiple_empty_notes = opts.allow_multiple_empty_notes == true

  buffer_manager.register_type("note", {
    on_save = M.save_note,
    on_close = function() end,
  })
end

return M
