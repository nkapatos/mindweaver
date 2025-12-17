---
--- API module for MW plugin
--- Provides centralized HTTP request handling with dev/prod server toggling
--- and independent debug logging control.
---
--- Architecture Decision:
--- - NO client-side validation of request parameters (id, body, etc.)
--- - Backend is the single source of truth for all validation logic
--- - Client only handles network/transport errors (JSON decode, HTTP status)
--- - This ensures consistency across all clients and simplifies maintenance
--- - See: Phase 1 architecture decision (2025-11-24)
---
--- Recent changes (2025-11-24):
--- - Fixed typo: debug_infor -> debug_info
--- - Separated debug logging from prod mode (now independent)
--- - Added :MwToggleDebug command for runtime debug control
--- - Improved logging to use vim.notify with proper log levels
--- - Removed client-side validation (architectural decision)
--- - Added lua_ls type annotations for better IDE support
---
--- Next steps:
--- - Consider adding request/response caching for offline support
--- - Add retry logic for failed requests
--- - Implement request timeout configuration
---
local M = {}
local curl = require("plenary.curl")

---@class ApiError
---@field code number Error code (HTTP status or custom)
---@field message string Human-readable error message
---@field status string Error status string (e.g., "INVALID_ARGUMENT", "NOT_FOUND")

---@class ApiResponse
---@field status number HTTP status code
---@field data? table Response data (present on success)
---@field error? ApiError Error object (present on failure)

M.config = {
  server_url = {
    dev = "http://localhost:9420",
    prod = "http://192.168.64.1:9999",
  },
  use_prod = false,
  debug_info = true, -- Can be toggled independently with :MwToggleDebug
}

function M.setup(opts)
  opts = opts or {}
  M.config.server_url.dev = opts.server_url_dev or M.config.server_url.dev
  M.config.server_url.prod = opts.server_url_prod or M.config.server_url.prod
  M.config.use_prod = opts.use_prod or M.config.use_prod
  M.config.debug_info = opts.debug_info or M.config.debug_info
  if M.config.debug_info then
    vim.notify("MW API Setup Completed", vim.log.levels.INFO)
  end
end

---Centralized API request handler
---Handles HTTP requests with automatic error handling and response parsing
---@param method string HTTP method ("GET", "POST", "PUT", "DELETE")
---@param endpoint string API endpoint path (e.g., "/api/notes/123")
---@param opts table Request options (query, body, headers)
---@param cb fun(res: ApiResponse) Callback function following Google API guidelines
local function request(method, endpoint, opts, cb)
  local base_url = M.config.use_prod and M.config.server_url.prod or M.config.server_url.dev
  local url = base_url .. endpoint
  -- TODO: set some default options here
  opts = opts or {}

  if M.config.debug_info then
    vim.notify("API Request: " .. method:upper() .. " " .. url, vim.log.levels.DEBUG)
  end

  opts.callback = function(res)
    vim.schedule(function()
      -- Try to decode the response body. Following google api guidelines, the response body will either
      -- have the field data or error. In case of both present error wins
      local ok, res_body = pcall(vim.json.decode, res.body)

      -- JSON Decoding has failed, return cb with the response body and error
      if not ok then
        cb({
          status = res.status,
          error = { code = res.status, message = "JSON Decode error: " .. tostring(res_body) },
        })
        return
      end

      if res.status >= 200 and res.status < 300 then
        -- Success: Expected response {"data": {}}
        if res_body.data then
          cb({
            status = res.status,
            data = res_body.data,
          })
        else
          cb({
            status = res.status,
            error = {
              code = res.status,
              message = "Server response was ok but response body is missing field 'data'",
            },
          })
        end
      else
        -- Error: Expected response {"error": {code: int, message: string, status: string}}
        local err = res_body.error or { code = res.status, message = "unknown", status = "UNKNOWN" }
        cb({
          status = res.status,
          error = err,
        })
      end
    end)
  end

  local curl_fn = ({
    GET = curl.get,
    POST = curl.post,
    PUT = curl.put,
    DELETE = curl.delete,
  })[method:upper()]

  if M.config.debug_info then
    vim.notify("Request opts: " .. vim.inspect(opts), vim.log.levels.DEBUG)
  end

  if curl_fn then
    curl_fn(url, opts)
  else
    cb({
      error = { code = 0, message = "Unsupported method: " .. method, status = "INVALID METHOD" },
    })
  end
end

---@class ResourceMethods
---@field list fun(query: table?, cb: fun(res: ApiResponse)) List resources with optional query parameters
---@field get fun(id: string, cb: fun(res: ApiResponse)) Get a single resource by ID
---@field create fun(body: table, cb: fun(res: ApiResponse)) Create a new resource
---@field update fun(id: string, body: table, etag: string?, cb: fun(res: ApiResponse)) Update a resource
---@field delete fun(id: string, cb: fun(res: ApiResponse)) Delete a resource by ID

---Create CRUD resource handlers for a given resource type
---All validation is handled by the backend - no client-side validation
---@param resource string Resource name (e.g., "notes", "tags")
---@return ResourceMethods Table with CRUD methods
local function create_resource(resource)
  local base_path = "/api/" .. resource
  return {
    -- GET /api/resource (list with optional query params)
    list = function(query, cb)
      request("GET", base_path, { query = query }, cb)
    end,
    -- GET /api/resource/:id
    -- Backend handles validation of ID parameter
    get = function(id, cb)
      request("GET", base_path .. "/" .. id, {}, cb)
    end,
    -- POST /api/resource (create with body)
    -- Backend handles validation of request body
    create = function(body, cb)
      request(
        "POST",
        base_path,
        { body = vim.json.encode(body), headers = { ["Content-Type"] = "application/json" } },
        cb
      )
    end,
    -- PUT /api/resource/:id (update with body)
    -- Backend handles validation of ID and body parameters
    update = function(id, body, etag, cb)
      request("PUT", base_path .. "/" .. id, {
        body = vim.json.encode(body),
        headers = {
          ["Content-Type"] = "application/json",
          ["If-Match"] = etag or "*",
        },
      }, cb)
    end,
    -- DELETE /api/resource/:id
    -- Backend handles validation of ID parameter
    delete = function(id, cb)
      request("DELETE", base_path .. "/" .. id, {}, cb)
    end,
  }
end

-- Expose resources (add more as needed)
---@type ResourceMethods
M.notes = create_resource("mind/notes")

-- Helper function to list notes by collection ID
-- Uses the nested endpoint: GET /api/mind/collections/:id/notes
---@param collection_id number|string The collection ID
---@param query table? Optional query parameters
---@param cb fun(res: ApiResponse) Callback function
M.list_notes_by_collection = function(collection_id, query, cb)
  request("GET", "/api/mind/collections/" .. tostring(collection_id) .. "/notes", { query = query }, cb)
end

-- Toggle command for dev/prod server selection
-- Note: Debug logging is independent - use :MwToggleDebug to control it
vim.api.nvim_create_user_command("MwToggleProd", function()
  M.config.use_prod = not M.config.use_prod
  local url = M.config.use_prod and M.config.server_url.prod or M.config.server_url.dev
  vim.notify("Server: " .. (M.config.use_prod and "PROD" or "DEV") .. " (" .. url .. ")", vim.log.levels.INFO)
end, { desc = "Toggle between dev and prod server" })

-- Toggle command for debug logging (independent of server mode)
-- Useful for debugging production issues or cleaning up dev logs
vim.api.nvim_create_user_command("MwToggleDebug", function()
  M.config.debug_info = not M.config.debug_info
  vim.notify("Debug logging: " .. (M.config.debug_info and "ON" or "OFF"), vim.log.levels.INFO)
end, { desc = "Toggle debug logging" })

return M
