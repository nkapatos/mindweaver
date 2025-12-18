---
--- notes.lua - Note management for Neoweaver (v3)
--- Handles note listing, opening, editing, and saving
---
--- Reference: clients/mw/notes.lua (v1 implementation)
---
local api = require("neoweaver._internal.api")
local buffer_manager = require("neoweaver._internal.buffer.manager")
local diff = require("neoweaver._internal.diff")

-- NOTE: This is temp here
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

-- Debounce state for create_note
local last_create_time = 0
local DEBOUNCE_MS = 500

local allow_multiple_empty_notes = false

-- Register note type handlers with buffer manager
-- This is called once during setup
local function register_handlers()
  buffer_manager.register_type("note", {
    on_save = M.save_note,
    on_close = function() end,
  })
end

--- List all notes using vim.ui.select
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

    -- v3 API: Response is mind.v3.ListNotesResponse directly
    ---@type mind.v3.ListNotesResponse
    local list_res = res.data
    local notes = list_res.notes or {}

    if #notes == 0 then
      vim.notify("No notes found!", vim.log.levels.INFO)
      return
    end

    -- Format notes for display
    local items = {}
    for _, note in ipairs(notes) do
      table.insert(items, string.format("[%d] %s", note.id, note.title))
    end

    -- Show picker
    vim.ui.select(items, {
      prompt = "Select a note:",
    }, function(choice, idx)
      if not choice then
        return
      end
      local selected_note = notes[idx]
      -- Open note for editing
      M.open_note(tonumber(selected_note.id))
    end)
  end)
end

--- Create a new note (server-first approach with auto-generated title)
--- Server generates "Untitled 0", "Untitled 1", etc. via NewNote endpoint
function M.create_note()
  -- Debounce rapid calls unless feature explicitly allows multiple empty notes
  local now = vim.loop.now()
  if not allow_multiple_empty_notes and (now - last_create_time < DEBOUNCE_MS) then
    vim.notify("Please wait before creating another note", vim.log.levels.WARN)
    return
  end
  last_create_time = now

  -- Attempt to reuse the current note's collection when available
  local collection_id = 1
  local current_buf = vim.api.nvim_get_current_buf()
  local base_entity = buffer_manager.get_entity(current_buf)
  if base_entity and base_entity.type == "note" then
    collection_id = vim.b[current_buf].note_collection_id or collection_id
  end

  -- Call NewNote endpoint - server generates title automatically
  ---@type mind.v3.NewNoteRequest
  local req = {
    collectionId = collection_id,
  }

  api.notes.new(req, function(res)
    if res.error then
      vim.notify("Failed to create note: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    -- Note created with auto-generated title "Untitled 0", "Untitled 1", etc.
    ---@type mind.v3.Note
    local note = res.data
    local note_id = tonumber(note.id)

    -- Create managed buffer via buffer_manager
    local bufnr = buffer_manager.create({
      type = "note",
      id = note_id,
      name = note.title, -- Server-generated "Untitled N"
      filetype = "markdown",
      modifiable = true,
    })

    -- Buffer starts empty (note.body = "")
    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, {})
    vim.api.nvim_set_option_value("modified", allow_multiple_empty_notes, { buf = bufnr })

    -- Store note metadata
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
--- If buffer already exists, switch to it; otherwise fetch and create
---@param note_id integer The note ID to open
function M.open_note(note_id)
  if not note_id then
    vim.notify("Invalid note ID", vim.log.levels.ERROR)
    return
  end

  -- Check if buffer already exists
  local existing = buffer_manager.get("note", note_id)
  if existing and vim.api.nvim_buf_is_valid(existing) then
    buffer_manager.switch_to_buffer(existing)
    return
  end

  -- Fetch note from API
  ---@type mind.v3.GetNoteRequest
  local req = { id = note_id }

  api.notes.get(req, function(res)
    if res.error then
      vim.notify("Error loading note: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    -- v3 API: Response is mind.v3.Note directly
    ---@type mind.v3.Note
    local note = res.data

    -- Create buffer via buffer_manager
    local bufnr = buffer_manager.create({
      type = "note",
      id = note_id,
      name = note.title or "Untitled",
      filetype = "markdown",
      modifiable = true,
    })

    -- Load content into buffer
    local lines = vim.split(note.body or "", "\n")
    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)
    vim.api.nvim_set_option_value("modified", false, { buf = bufnr })

    -- Store note data in buffer variables
    -- Using note_* prefix for domain-specific storage
    vim.b[bufnr].note_id = note_id
    vim.b[bufnr].note_title = note.title
    vim.b[bufnr].note_etag = note.etag
    vim.b[bufnr].note_collection_id = note.collectionId
    vim.b[bufnr].note_type_id = note.noteTypeId
    vim.b[bufnr].note_metadata = note.metadata or {}
  end)
end

--- Edit note metadata (frontmatter)
-- TODO: Implement metadata editing functionality
-- This will allow editing YAML frontmatter (tags, custom fields, etc.)
function M.edit_metadata()
  vim.notify("Metadata editing not yet implemented in v3", vim.log.levels.WARN)
  -- TODO: Implementation steps:
  -- 1. Get note_id from current buffer if not provided
  -- 2. Fetch note from API
  -- 3. Parse frontmatter from note.body
  -- 4. Open floating window with editable YAML
  -- 5. On save, update note with new frontmatter
end

--- Edit the current note title and persist immediately
--- Uses buffer_manager to ensure buffer is managed and saves body + title together
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

    -- Save immediately so server validates duplicates and updates etag/body
    M.save_note(bufnr, note_id)
  end)
end

--- Create a new quicknote in a floating window
-- TODO: Implement quicknotes functionality
-- Quicknotes are ephemeral floating windows for rapid note capture
function M.create_quicknote()
  vim.notify("Quicknotes not yet implemented in v3", vim.log.levels.WARN)
  -- TODO: Implementation steps:
  -- 1. Create floating window with configured dimensions
  -- 2. Create buffer with auto-generated title (timestamp-based)
  -- 3. On save, call NewNote API with quicknote note_type
  -- 4. Store reference for amend functionality
  -- Reference: clients/mw/notes.lua:handler__new_quicknote
end

--- List all quicknotes
-- TODO: Implement quicknotes listing
function M.list_quicknotes()
  vim.notify("Quicknotes not yet implemented in v3", vim.log.levels.WARN)
  -- TODO: Implementation steps:
  -- 1. Call ListNotes API with note_type filter for quicknote
  -- 2. Display in picker
  -- 3. On select, open in floating window
  -- Reference: clients/mw/notes.lua:handler__list_quicknotes
end

--- Amend the last created quicknote
-- TODO: Implement quicknote amend functionality
function M.amend_quicknote()
  vim.notify("Quicknote amend not yet implemented in v3", vim.log.levels.WARN)
  -- TODO: Implementation steps:
  -- 1. Retrieve last quicknote ID from state
  -- 2. Fetch note from API
  -- 3. Open in floating window with existing content
  -- 4. Allow editing and save
  -- Reference: clients/mw/notes.lua:handler__amend_quicknote
end

--- Handle etag conflict resolution
--- Called when save fails with 412 Precondition Failed
--- Fetches latest server version and enables diff view
---@param bufnr integer Buffer number
---@param note_id integer Note ID
local function handle_conflict(bufnr, note_id)
  -- NOTE: Due to how connect rpc is handling this and the connect error for precondition fail returns a 400
  -- a custom error has been addded for now to check against 409 with additional meta info
  -- This will chnge in the near future to match the 409 that the precondition etc should not be in the header but in the body
  vim.notify("ETag conflict detected - fetching latest version from server...", vim.log.levels.WARN)

  -- Fetch latest version from server
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

    -- Enable diff view for conflict resolution
    -- Defer conflict count check until diff is initialized
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

    -- Ensure buffer is modifiable
    if not vim.api.nvim_get_option_value("modifiable", { buf = bufnr }) then
      vim.api.nvim_set_option_value("modifiable", true, { buf = bufnr })
    end

    -- Enable diff overlay
    diff.enable(bufnr)
    diff.map_keys(bufnr)

    -- Set server version as reference (with slight delay for stability)
    vim.defer_fn(function()
      local ok, err = pcall(diff.set_ref_text, bufnr, server_lines)
      if not ok then
        vim.notify("Failed to enable diff overlay: " .. tostring(err), vim.log.levels.ERROR)
        return
      end
    end, 50)

    -- Create one-shot autocmd to cleanup diff and retry save
    local group = vim.api.nvim_create_augroup("neoweaver_conflict_" .. bufnr, { clear = true })
    vim.api.nvim_create_autocmd("BufWriteCmd", {
      group = group,
      buffer = bufnr,
      callback = function()
        -- Check for remaining conflicts
        local remaining = diff.get_conflict_count(bufnr)
        if remaining > 0 then
          vim.notify(
            string.format("⚠️  Saving with %d unresolved conflict%s", remaining, remaining > 1 and "s" or ""),
            vim.log.levels.WARN
          )
        else
          vim.notify("✓ All conflicts resolved - saving", vim.log.levels.INFO)
        end

        -- Disable diff overlay
        diff.disable(bufnr)
        vim.api.nvim_del_augroup_by_id(group)

        -- Update etag to latest and retry save
        vim.b[bufnr].note_etag = latest_note.etag
        M.save_note(bufnr, note_id)
      end,
      once = true,
    })
  end)
end

--- Delete a note by ID
---@param note_id integer The note ID to delete
function M.delete_note(note_id)
  if not note_id then
    vim.notify("Invalid note ID", vim.log.levels.ERROR)
    return
  end

  -- Ask for confirmation
  vim.ui.input({
    prompt = string.format("Delete note %d? (y/N): ", note_id),
  }, function(input)
    if not input or (input:lower() ~= "y" and input:lower() ~= "yes") then
      vim.notify("Delete cancelled", vim.log.levels.INFO)
      return
    end

    -- Call delete API
    ---@type mind.v3.DeleteNoteRequest
    local req = { id = note_id }

    api.notes.delete(req, function(res)
      if res.error then
        vim.notify("Failed to delete note: " .. res.error.message, vim.log.levels.ERROR)
        return
      end

      -- Close buffer if it's open
      local bufnr = buffer_manager.get("note", note_id)
      if bufnr and vim.api.nvim_buf_is_valid(bufnr) then
        vim.api.nvim_buf_delete(bufnr, { force = true })
      end

      vim.notify("Note deleted successfully", vim.log.levels.INFO)
    end)
  end)
end

--- Save note buffer content to server
--- Called by buffer_manager when buffer is saved (:w)
--- Always updates existing note (server-first approach ensures ID exists)
---@param bufnr integer Buffer number
---@param id integer Note ID
function M.save_note(bufnr, id)
  -- Extract buffer content
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local body = table.concat(lines, "\n")

  -- Get stored note data
  local title = vim.b[bufnr].note_title or "Untitled"
  local etag = vim.b[bufnr].note_etag
  local collection_id = vim.b[bufnr].note_collection_id or 1
  local note_type_id = vim.b[bufnr].note_type_id
  local metadata = vim.b[bufnr].note_metadata or {}

  -- Build update request
  ---@type mind.v3.ReplaceNoteRequest
  local req = {
    id = id,
    title = title,
    body = body,
    collectionId = collection_id,
  }

  -- Add optional fields only if they have values
  if note_type_id then
    req.noteTypeId = note_type_id
  end

  if metadata and next(metadata) ~= nil then
    req.metadata = metadata
  end

  -- Call API with etag for optimistic locking
  api.notes.update(req, etag, function(res)
    if res.error then
      if res.error.code == ConnectCode.FAILED_PRECONDITION then
        -- Handle conflict with diff view
        handle_conflict(bufnr, id)
        return
      end

      -- Other errors
      vim.notify("Save failed: " .. res.error.message, vim.log.levels.ERROR)
      return
    end

    -- Update etag and mark buffer as unmodified
    ---@type mind.v3.Note
    local updated_note = res.data
    vim.b[bufnr].note_etag = updated_note.etag
    vim.api.nvim_set_option_value("modified", false, { buf = bufnr })
    vim.notify("Note saved successfully", vim.log.levels.INFO)
  end)
end

function M.setup(opts)
  opts = opts or {}
  allow_multiple_empty_notes = opts.allow_multiple_empty_notes == true

  -- Register note type handlers with buffer manager
  register_handlers()
end

return M
