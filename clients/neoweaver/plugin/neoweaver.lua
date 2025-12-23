-- neoweaver.nvim plugin entry point
-- This file is automatically sourced by Neovim on startup
-- Commands are defined here following Neovim plugin best practices

if vim.g.loaded_neoweaver then
  return
end
vim.g.loaded_neoweaver = true

-- Create commands
-- Note: The plugin must be setup via require('neoweaver').setup() before these work

-- Notes commands
vim.api.nvim_create_user_command("NeoweaverNotesList", function()
  require("neoweaver._internal.notes").list_notes()
end, { desc = "List notes" })

vim.api.nvim_create_user_command("NeoweaverNotesOpen", function(opts)
  local id = tonumber(opts.args)
  if id then
    require("neoweaver._internal.notes").open_note(id)
  else
    vim.notify("Usage: :NeoweaverNotesOpen <note_id>", vim.log.levels.WARN)
  end
end, { nargs = 1, desc = "Open note by ID" })

vim.api.nvim_create_user_command("NeoweaverNotesNew", function()
  require("neoweaver._internal.notes").create_note()
end, { desc = "Create new note" })

vim.api.nvim_create_user_command("NeoweaverNotesNewWithTitle", function()
  require("neoweaver._internal.notes").create_note_with_title()
end, { desc = "Create note with title prompt" })

vim.api.nvim_create_user_command("NeoweaverNotesTitle", function()
  require("neoweaver._internal.notes").edit_title()
end, { desc = "Edit current note title" })

vim.api.nvim_create_user_command("NeoweaverNotesDelete", function(opts)
  local id = tonumber(opts.args)
  if id then
    require("neoweaver._internal.notes").delete_note(id)
  else
    vim.notify("Usage: :NeoweaverNotesDelete <note_id>", vim.log.levels.WARN)
  end
end, { nargs = 1, desc = "Delete note by ID" })

vim.api.nvim_create_user_command("NeoweaverNotesMeta", function(opts)
  local id = opts.args ~= "" and tonumber(opts.args) or nil
  require("neoweaver._internal.notes").edit_metadata(id)
end, { nargs = "?", desc = "Edit note metadata - See issue #44" })

vim.api.nvim_create_user_command("NeoweaverNotesQuick", function()
  require("neoweaver._internal.quicknote").open()
end, { desc = "Capture a quicknote" })

vim.api.nvim_create_user_command("NeoweaverNotesQuickList", function()
  require("neoweaver._internal.quicknote").list()
end, { desc = "List quicknotes - See issue #45" })

vim.api.nvim_create_user_command("NeoweaverNotesQuickAmend", function()
  require("neoweaver._internal.quicknote").amend()
end, { desc = "Amend quicknote - See issue #45" })

-- API/Server commands
vim.api.nvim_create_user_command("NeoweaverServerUse", function(opts)
  require("neoweaver._internal.api").set_current_server(opts.args)
end, {
  nargs = 1,
  complete = function(ArgLead)
    local api = require("neoweaver._internal.api")
    local matches = {}
    for _, name in ipairs(api.list_server_names()) do
      if name:find("^" .. vim.pesc(ArgLead)) then
        table.insert(matches, name)
      end
    end
    return matches
  end,
  desc = "Select Neoweaver server by name",
})

vim.api.nvim_create_user_command("NeoweaverToggleDebug", function()
  require("neoweaver._internal.api").toggle_debug()
end, { desc = "Toggle debug logging" })

-- Explorer commands
vim.api.nvim_create_user_command("NeoweaverExplorer", function(opts)
  local explorer = require("neoweaver._internal.explorer")
  local action = opts.args ~= "" and opts.args or "toggle"

  if action == "open" then
    explorer.open()
  elseif action == "close" then
    explorer.close()
  elseif action == "toggle" then
    explorer.toggle()
  elseif action == "focus" then
    explorer.focus()
  else
    vim.notify("Invalid action: " .. action .. ". Use: open, close, toggle, focus", vim.log.levels.WARN)
  end
end, {
  nargs = "?",
  complete = function()
    return { "open", "close", "toggle", "focus" }
  end,
  desc = "Neoweaver collections explorer",
})

-- Collections commands
vim.api.nvim_create_user_command("NeoweaverCollectionCreate", function(opts)
  local collections = require("neoweaver._internal.collections")
  
  -- Parse args: name [parent_id]
  local args = vim.split(opts.args, "%s+")
  local name = args[1]
  local parent_id = args[2] and tonumber(args[2]) or nil
  
  if not name or name == "" then
    vim.notify("Usage: :NeoweaverCollectionCreate <name> [parent_id]", vim.log.levels.WARN)
    return
  end
  
  collections.create_collection(name, parent_id, function(collection, err)
    if err then
      vim.notify("Failed to create collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
      return
    end
    vim.notify("Created collection: " .. collection.displayName, vim.log.levels.INFO)
  end)
end, { nargs = "+", desc = "Create new collection" })

vim.api.nvim_create_user_command("NeoweaverCollectionRename", function(opts)
  local collections = require("neoweaver._internal.collections")
  
  -- Parse args: id new_name
  local args = vim.split(opts.args, "%s+", { trimempty = true })
  local id = tonumber(args[1])
  local new_name = table.concat(vim.list_slice(args, 2), " ")
  
  if not id or not new_name or new_name == "" then
    vim.notify("Usage: :NeoweaverCollectionRename <id> <new_name>", vim.log.levels.WARN)
    return
  end
  
  collections.rename_collection(id, new_name, function(collection, err)
    if err then
      vim.notify("Failed to rename collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
      return
    end
    vim.notify("Renamed collection to: " .. collection.displayName, vim.log.levels.INFO)
  end)
end, { nargs = "+", desc = "Rename collection" })

vim.api.nvim_create_user_command("NeoweaverCollectionDelete", function(opts)
  local collections = require("neoweaver._internal.collections")
  local id = tonumber(opts.args)
  
  if not id then
    vim.notify("Usage: :NeoweaverCollectionDelete <id>", vim.log.levels.WARN)
    return
  end
  
  -- Confirm deletion
  vim.ui.input({ prompt = "Delete collection " .. id .. "? (y/N): " }, function(confirm)
    if confirm ~= "y" and confirm ~= "Y" then
      return
    end
    
    collections.delete_collection(id, function(success, err)
      if err then
        vim.notify("Failed to delete collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
        return
      end
      vim.notify("Deleted collection", vim.log.levels.INFO)
    end)
  end)
end, { nargs = 1, desc = "Delete collection" })

vim.api.nvim_create_user_command("NeoweaverCollectionUpdate", function(opts)
  local collections = require("neoweaver._internal.collections")
  
  -- Parse args: id [displayName=value] [parentId=value] [description=value]
  local args = vim.split(opts.args, "%s+", { trimempty = true })
  local id = tonumber(args[1])
  
  if not id then
    vim.notify("Usage: :NeoweaverCollectionUpdate <id> [displayName=value] [parentId=value]", vim.log.levels.WARN)
    return
  end
  
  local update_opts = {}
  for i = 2, #args do
    local key, value = args[i]:match("^([^=]+)=(.+)$")
    if key == "displayName" then
      update_opts.displayName = value
    elseif key == "parentId" then
      update_opts.parentId = tonumber(value)
    elseif key == "description" then
      update_opts.description = value
    end
  end
  
  if vim.tbl_isempty(update_opts) then
    vim.notify("No update options provided", vim.log.levels.WARN)
    return
  end
  
  collections.update_collection(id, update_opts, function(collection, err)
    if err then
      vim.notify("Failed to update collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
      return
    end
    vim.notify("Updated collection: " .. collection.displayName, vim.log.levels.INFO)
  end)
end, { nargs = "+", desc = "Update collection" })
