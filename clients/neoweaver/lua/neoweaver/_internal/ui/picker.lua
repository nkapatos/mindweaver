---
--- picker.lua - Generic entity picker using nui.menu
---
--- Decoupled from domain logic - can be used for notes, collections, tasks, etc.
--- Provides a unified selection interface with keyboard navigation
---
local Menu = require("nui.menu")

local M = {}

--- Create a picker for selecting from a list of items
---@param items table[] Array of items to display
---@param opts table Configuration options
---   - format_item: function(item, idx) -> string (required) - formats each item for display
---   - prompt: string (optional) - prompt text, default: "Select:"
---   - on_submit: function(item, idx) (required) - callback when item is selected
---   - on_close: function() (optional) - callback when picker is closed without selection
---   - size: table (optional) - { width = number, height = number }
---   - position: string|table (optional) - "50%" or { row = "50%", col = "50%" }
---   - border: table (optional) - nui border config
function M.pick(items, opts)
  opts = opts or {}

  if not items or #items == 0 then
    vim.notify("No items to display", vim.log.levels.WARN)
    return
  end

  if not opts.format_item then
    error("picker.pick requires opts.format_item function")
  end

  if not opts.on_submit then
    error("picker.pick requires opts.on_submit callback")
  end

  -- Build menu lines
  local menu_items = {}
  for idx, item in ipairs(items) do
    local text = opts.format_item(item, idx)
    table.insert(menu_items, Menu.item(text, { item = item, index = idx }))
  end

  -- Default configuration
  local size = opts.size or { width = 60, height = 20 }
  local position = opts.position or "50%"
  local border_config = opts.border or {
    style = "rounded",
    text = {
      top = opts.prompt and (" " .. opts.prompt .. " ") or " Select ",
      top_align = "center",
    },
  }

  local menu = Menu({
    relative = "editor",
    position = position,
    size = size,
    border = border_config,
    win_options = {
      winblend = 0,
      winhighlight = "Normal:Normal,FloatBorder:FloatBorder",
    },
  }, {
    lines = menu_items,
    max_width = size.width,
    keymap = {
      focus_next = { "j", "<Down>", "<Tab>" },
      focus_prev = { "k", "<Up>", "<S-Tab>" },
      close = { "q", "<Esc>", "<C-c>" },
      submit = { "<CR>", "<Space>" },
    },
    on_close = function()
      if opts.on_close then
        opts.on_close()
      end
    end,
    on_submit = function(selected)
      -- Call the user's callback with the original item and index
      opts.on_submit(selected.item, selected.index)
    end,
  })

  menu:mount()

  -- Auto-focus first item
  vim.schedule(function()
    if menu.tree then
      menu.tree:render()
    end
  end)

  return menu
end

return M
