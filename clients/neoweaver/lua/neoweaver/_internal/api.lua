local M = {}
local curl = require("plenary.curl")

---@class ConnectError Connect RPC error format
---@field code string Error code (e.g., "invalid_argument", "not_found")
---@field message string Human-readable error message
---@field details? table[] Additional error details

---@class ApiResponse
---@field status number HTTP status code
---@field data? table Response data (present on success) - direct proto message
---@field error? ConnectError Error object (present on failure)

M.config = {
  servers = {},
  current_server = nil,
  debug_info = true, -- Can be toggled independently with :MwToggleDebug
}

local function normalize_servers(servers)
  if type(servers) ~= "table" or vim.tbl_isempty(servers) then
    error("neoweaver.api.setup: 'servers' table is required")
  end

  local normalized = {}
  local default_name = nil

  for name, value in pairs(servers) do
    if type(name) ~= "string" or name == "" then
      error("neoweaver.api.setup: server names must be non-empty strings")
    end

    local url
    local is_default = false

    if type(value) == "string" then
      url = value
    elseif type(value) == "table" then
      url = value.url
      is_default = value.default == true
    else
      error(string.format("neoweaver.api.setup: invalid configuration for server '%s'", name))
    end

    if type(url) ~= "string" or url == "" then
      error(string.format("neoweaver.api.setup: server '%s' must provide a non-empty url", name))
    end

    normalized[name] = { url = url }

    if is_default and not default_name then
      default_name = name
    end
  end

  return normalized, default_name
end

function M.setup(opts)
  opts = opts or {}

  if opts.debug_info ~= nil then
    M.config.debug_info = opts.debug_info
  end

  if opts.servers then
    local normalized, default_name = normalize_servers(opts.servers)
    M.config.servers = normalized
    M.config.current_server = default_name
  else
    M.config.servers = {}
    M.config.current_server = nil
  end

  if vim.tbl_isempty(M.config.servers) then
    error("neoweaver.api.setup: at least one server must be configured")
  end

  if M.config.debug_info then
    local msg = "Neoweaver API configured"
    if M.config.current_server then
      msg = msg .. " (default: " .. M.config.current_server .. ")"
    else
      msg = msg .. " (no default server)"
    end
    vim.notify(msg, vim.log.levels.INFO)
  end
end

---Centralized API request handler for Connect RPC
---Connect RPC returns the proto message directly (no {"data": ...} wrapper for success)
---Errors are returned as {"code": "...", "message": "...", "details": [...]}
---@param method string Semantic HTTP method ("GET", "POST", "PUT", "DELETE") - all become POST
---@param endpoint string Connect RPC endpoint (e.g., "/mind.v3.NotesService/GetNote")
---@param opts table Request options (body, headers)
---@param cb fun(res: ApiResponse) Callback function
local function get_current_server_url()
  local name = M.config.current_server
  if not name or name == "" then
    error("neoweaver.api: no server selected. Use :MwServerUse <name> or set a default server.")
  end

  local entry = M.config.servers[name]
  if not entry then
    error(string.format("neoweaver.api: server '%s' is not configured", name))
  end

  return entry.url
end

local function request(method, endpoint, opts, cb)
  local base_url = get_current_server_url()
  local url = base_url .. endpoint
  opts = opts or {}

  if M.config.debug_info then
    vim.notify("API Request: " .. method:upper() .. " " .. url, vim.log.levels.DEBUG)
  end

  opts.callback = function(res)
    vim.schedule(function()
      -- Try to decode the response body
      local ok, res_body = pcall(vim.json.decode, res.body)

      -- JSON Decoding has failed
      if not ok then
        cb({
          status = res.status,
          error = { code = "parse_error", message = "JSON Decode error: " .. tostring(res_body) },
        })
        return
      end

      if res.status >= 200 and res.status < 300 then
        -- Success: Connect RPC returns proto message directly (not wrapped in {"data": ...})
        cb({
          status = res.status,
          data = res_body,
        })
      else
        -- Error: Connect RPC format {"code": "...", "message": "...", "details": [...]}
        local err = res_body or { code = "unknown", message = "unknown error" }
        cb({
          status = res.status,
          error = err,
        })
      end
    end)
  end

  -- Connect RPC always uses POST regardless of semantic method
  -- We keep the method parameter for developer clarity (GET/PUT/DELETE semantics)
  -- but all requests are actually POST under the hood
  local curl_fn = curl.post

  if M.config.debug_info then
    vim.notify("Request opts: " .. vim.inspect(opts), vim.log.levels.DEBUG)
  end

  curl_fn(url, opts)
end

---@class NotesMethods Notes service methods
---@field list fun(req: mind.v3.ListNotesRequest, cb: fun(res: ApiResponse)) List notes
---@field get fun(req: mind.v3.GetNoteRequest, cb: fun(res: ApiResponse)) Get a note
---@field create fun(req: mind.v3.CreateNoteRequest, cb: fun(res: ApiResponse)) Create a note
---@field new fun(req: mind.v3.NewNoteRequest, cb: fun(res: ApiResponse)) Create a new note with auto-generated title
---@field find fun(req: mind.v3.FindNotesRequest, cb: fun(res: ApiResponse)) Find notes by title and filters
---@field update fun(req: mind.v3.ReplaceNoteRequest, etag: string?, cb: fun(res: ApiResponse)) Update a note
---@field delete fun(req: mind.v3.DeleteNoteRequest, cb: fun(res: ApiResponse)) Delete a note

-- Notes Service
---@type NotesMethods
M.notes = {
  -- POST /mind.v3.NotesService/ListNotes
  -- Request: mind.v3.ListNotesRequest
  -- Response: mind.v3.ListNotesResponse
  list = function(req, cb)
    request("GET", "/mind.v3.NotesService/ListNotes", {
      body = vim.json.encode(req or {}),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.NotesService/GetNote
  -- Request: mind.v3.GetNoteRequest
  -- Response: mind.v3.Note
  get = function(req, cb)
    request("GET", "/mind.v3.NotesService/GetNote", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.NotesService/CreateNote
  -- Request: mind.v3.CreateNoteRequest
  -- Response: mind.v3.Note
  create = function(req, cb)
    request("POST", "/mind.v3.NotesService/CreateNote", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.NotesService/NewNote
  -- Request: mind.v3.NewNoteRequest
  -- Response: mind.v3.Note
  new = function(req, cb)
    request("POST", "/mind.v3.NotesService/NewNote", {
      body = vim.json.encode(req or {}),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.NotesService/FindNotes
  -- Request: mind.v3.FindNotesRequest
  -- Response: mind.v3.FindNotesResponse
  find = function(req, cb)
    request("POST", "/mind.v3.NotesService/FindNotes", {
      body = vim.json.encode(req or {}),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.NotesService/ReplaceNote
  -- Request: mind.v3.ReplaceNoteRequest
  -- Response: mind.v3.Note
  -- Requires If-Match header with etag for optimistic locking
  update = function(req, etag, cb)
    request("PUT", "/mind.v3.NotesService/ReplaceNote", {
      body = vim.json.encode(req),
      headers = {
        ["Content-Type"] = "application/json",
        ["If-Match"] = etag or "*",
      },
    }, cb)
  end,

  -- POST /mind.v3.NotesService/DeleteNote
  -- Request: mind.v3.DeleteNoteRequest
  -- Response: google.protobuf.Empty
  delete = function(req, cb)
    request("DELETE", "/mind.v3.NotesService/DeleteNote", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,
}

-- Helper function to list notes by collection ID
-- Uses collectionId field in ListNotesRequest
---@param collection_id number The collection ID
---@param query? table Optional additional query parameters (pageSize, pageToken, etc.)
---@param cb fun(res: ApiResponse) Callback function
M.list_notes_by_collection = function(collection_id, query, cb)
  local req = vim.tbl_extend("force", query or {}, {
    collectionId = collection_id,
  })
  M.notes.list(req, cb)
end

---@class CollectionsMethods Collections service methods
---@field list fun(req: mind.v3.ListCollectionsRequest, cb: fun(res: ApiResponse)) List collections
---@field get fun(req: mind.v3.GetCollectionRequest, cb: fun(res: ApiResponse)) Get a collection
---@field create fun(req: mind.v3.CreateCollectionRequest, cb: fun(res: ApiResponse)) Create a collection
---@field update fun(req: mind.v3.UpdateCollectionRequest, cb: fun(res: ApiResponse)) Update a collection
---@field delete fun(req: mind.v3.DeleteCollectionRequest, cb: fun(res: ApiResponse)) Delete a collection

-- Collections Service
---@type CollectionsMethods
M.collections = {
  -- POST /mind.v3.CollectionsService/ListCollections
  -- Request: mind.v3.ListCollectionsRequest
  -- Response: mind.v3.ListCollectionsResponse (contains collections array and totalSize)
  list = function(req, cb)
    request("GET", "/mind.v3.CollectionsService/ListCollections", {
      body = vim.json.encode(req or {}),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.CollectionsService/GetCollection
  -- Request: mind.v3.GetCollectionRequest
  -- Response: mind.v3.Collection
  get = function(req, cb)
    request("GET", "/mind.v3.CollectionsService/GetCollection", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.CollectionsService/CreateCollection
  -- Request: mind.v3.CreateCollectionRequest
  -- Response: mind.v3.Collection
  create = function(req, cb)
    request("POST", "/mind.v3.CollectionsService/CreateCollection", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.CollectionsService/UpdateCollection
  -- Request: mind.v3.UpdateCollectionRequest
  -- Response: mind.v3.Collection
  update = function(req, cb)
    request("PUT", "/mind.v3.CollectionsService/UpdateCollection", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,

  -- POST /mind.v3.CollectionsService/DeleteCollection
  -- Request: mind.v3.DeleteCollectionRequest
  -- Response: google.protobuf.Empty
  delete = function(req, cb)
    request("DELETE", "/mind.v3.CollectionsService/DeleteCollection", {
      body = vim.json.encode(req),
      headers = { ["Content-Type"] = "application/json" },
    }, cb)
  end,
}

function M.set_current_server(name)
  if not name or name == "" then
    error("MwServerUse: server name is required")
  end

  if not M.config.servers[name] then
    error(string.format("MwServerUse: server '%s' is not configured", name))
  end

  M.config.current_server = name

  if M.config.debug_info then
    vim.notify("Neoweaver server set to '" .. name .. "'", vim.log.levels.INFO)
  end
end

function M.list_server_names()
  local names = {}
  for name, _ in pairs(M.config.servers) do
    if type(name) == "string" then
      table.insert(names, name)
    end
  end
  table.sort(names)
  return names
end

function M.toggle_debug()
  M.config.debug_info = not M.config.debug_info
  vim.notify("Debug logging: " .. (M.config.debug_info and "ON" or "OFF"), vim.log.levels.INFO)
end

return M
