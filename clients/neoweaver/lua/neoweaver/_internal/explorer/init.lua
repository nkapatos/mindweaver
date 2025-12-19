--- Explorer module for neoweaver
--- Orchestrates tree display and domain actions (collections, notes, tags)
---
local M = {}

local window = require("neoweaver._internal.explorer.window")
local tree = require("neoweaver._internal.explorer.tree")
local collections = require("neoweaver._internal.collections")

--- Setup action handlers for collection tree operations
local function setup_collection_actions()
  -- Handle create action
  tree.on_create = function(node, bufnr)
    -- Only allow creating collections under collection nodes
    if node.type ~= "collection" then
      vim.notify("Can only create collections under other collections", vim.log.levels.WARN)
      return
    end
    
    -- Prompt for collection name
    vim.ui.input({ prompt = "New collection name: " }, function(name)
      if not name or name == "" then
        return
      end
      
      collections.create_collection(name, node.collection_id, function(collection, err)
        if err then
          vim.notify("Failed to create collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
          return
        end
        
        vim.notify("Created collection: " .. collection.displayName, vim.log.levels.INFO)
        tree.refresh(bufnr)
      end)
    end)
  end

  -- Handle rename action
  tree.on_rename = function(node, bufnr)
    -- Only allow renaming collections
    if node.type ~= "collection" then
      vim.notify("Can only rename collections", vim.log.levels.WARN)
      return
    end
    
    -- Don't allow renaming system collections
    if node.is_system then
      vim.notify("Cannot rename system collections", vim.log.levels.WARN)
      return
    end
    
    -- Prompt for new name
    vim.ui.input({ prompt = "Rename to: ", default = node.name }, function(new_name)
      if not new_name or new_name == "" or new_name == node.name then
        return
      end
      
      collections.rename_collection(node.collection_id, new_name, function(collection, err)
        if err then
          vim.notify("Failed to rename collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
          return
        end
        
        vim.notify("Renamed collection to: " .. collection.displayName, vim.log.levels.INFO)
        tree.refresh(bufnr)
      end)
    end)
  end

  -- Handle delete action
  tree.on_delete = function(node, bufnr)
    -- Only allow deleting collections
    if node.type ~= "collection" then
      vim.notify("Can only delete collections", vim.log.levels.WARN)
      return
    end
    
    -- Don't allow deleting system collections
    if node.is_system then
      vim.notify("Cannot delete system collections", vim.log.levels.WARN)
      return
    end
    
    -- Confirm deletion
    vim.ui.input({ 
      prompt = "Delete collection '" .. node.name .. "'? (y/N): " 
    }, function(confirm)
      if confirm ~= "y" and confirm ~= "Y" then
        return
      end
      
      collections.delete_collection(node.collection_id, function(success, err)
        if err then
          vim.notify("Failed to delete collection: " .. (err.message or vim.inspect(err)), vim.log.levels.ERROR)
          return
        end
        
        vim.notify("Deleted collection: " .. node.name, vim.log.levels.INFO)
        tree.refresh(bufnr)
      end)
    end)
  end
end

--- Open the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.open(opts)
  local split = window.open(opts)
  
  -- Initialize and render tree if we just opened the window
  if split and split.bufnr then
    -- Setup action handlers for collections (current mode)
    -- In the future, this could be: setup_tags_actions() or setup_mixed_actions()
    setup_collection_actions()
    
    tree.init(split.bufnr)
  end
end

--- Close the explorer sidebar
function M.close()
  window.close()
end

--- Toggle the explorer sidebar
---@param opts? { position?: "left"|"right", size?: number }
function M.toggle(opts)
  if window.is_open() then
    window.close()
  else
    M.open(opts)
  end
end

--- Focus the explorer window (if open)
function M.focus()
  window.focus()
end

--- Check if explorer is currently open
---@return boolean
function M.is_open()
  return window.is_open()
end

return M
