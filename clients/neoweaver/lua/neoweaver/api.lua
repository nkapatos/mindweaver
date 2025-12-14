---
--- API module for Neoweaver plugin (v3)
--- Provides centralized HTTP request handling with Connect RPC transport
--- All types are generated from proto/mind/v3/*.proto
---
--- Connect RPC Transport:
--- - All RPC methods use POST regardless of operation type (Connect RPC protocol)
--- - Method parameter kept for developer clarity (GET/POST/PUT/DELETE semantic meaning)
--- - Actual HTTP method is always POST under the hood
---
--- Architecture Decision:
--- - NO client-side validation of request parameters (id, body, etc.)
--- - Backend is the single source of truth for all validation logic
--- - Client only handles network/transport errors (JSON decode, HTTP status)
--- - This ensures consistency across all clients and simplifies maintenance
---
--- Recent changes (2025-12-14):
--- - Migrated from REST API v1 to Connect RPC v3 transport
--- - All HTTP methods now use POST (Connect RPC requirement)
--- - Updated resource paths from /api/* to /mind.v3.*Service/*
--- - Request payloads moved from query params to POST body
--- - Updated all type annotations to use generated v3 types
---
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
	server_url = {
		dev = "http://localhost:9421",
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
		vim.notify("Neoweaver API Setup (v3) Completed", vim.log.levels.INFO)
	end
end

---Centralized API request handler for Connect RPC
---Connect RPC returns the proto message directly (no {"data": ...} wrapper for success)
---Errors are returned as {"code": "...", "message": "...", "details": [...]}
---@param method string Semantic HTTP method ("GET", "POST", "PUT", "DELETE") - all become POST
---@param endpoint string Connect RPC endpoint (e.g., "/mind.v3.NotesService/GetNote")
---@param opts table Request options (body, headers)
---@param cb fun(res: ApiResponse) Callback function
local function request(method, endpoint, opts, cb)
	local base_url = M.config.use_prod and M.config.server_url.prod or M.config.server_url.dev
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
