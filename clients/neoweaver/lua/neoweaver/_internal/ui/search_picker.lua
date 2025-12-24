---
--- search_picker.lua - Interactive search picker with input field and results
---
--- Features:
--- - Input field with debounced search
--- - Dynamic results display using nui.Menu
--- - Pagination support (fetch more on scroll)
--- - Keyboard navigation
---
local Layout = require("nui.layout")
local Input = require("nui.input")
local Menu = require("nui.menu")

local M = {}

--- State for debouncing
local debounce_timer = nil

--- Create a search picker with input field and results list
---@param opts table Configuration options
---   - prompt: string - Input field prompt (default: "Search:")
---   - min_query_length: number - Minimum characters before search (default: 3)
---   - debounce_ms: number - Debounce delay in milliseconds (default: 300)
---   - search_fn: function(query, page_token, callback) - Search function, callback(items, error, has_more, next_token)
---   - format_item: function(item, idx) -> string - Format item for display
---   - on_select: function(item, idx) - Callback when item is selected
---   - on_close: function() - Callback when picker is closed
---   - empty_message: string - Message when no results (default: "No results found")
---   - size: table - { width = number, height = number }
---   - position: string|table - "50%" or { row = "50%", col = "50%" }
function M.show(opts)
  opts = opts or {}

  -- Configuration
  local prompt = opts.prompt or "Search:"
  local min_query_length = opts.min_query_length or 3
  local debounce_ms = opts.debounce_ms or 300
  local empty_message = opts.empty_message or "No results found"
  local size = opts.size or { width = 80, height = 25 }
  local position = opts.position or "50%"

  -- Validate required callbacks
  if not opts.search_fn then
    error("search_picker.show requires opts.search_fn")
  end
  if not opts.format_item then
    error("search_picker.show requires opts.format_item")
  end
  if not opts.on_select then
    error("search_picker.show requires opts.on_select")
  end

  -- State
  local current_results = {}
  local current_query = ""
  local is_loading = false
  local has_more_pages = false
  local next_page_token = nil
  local is_mounted = true -- Track if picker is still mounted

  -- Forward declarations
  local input_popup
  local results_menu
  local layout
  local perform_search
  local update_menu_items
  local close_picker

  -- Helper to safely close the picker
  close_picker = function()
    is_mounted = false
    if layout and layout._.mounted then
      layout:unmount()
    end
  end

  -- Create input field
  input_popup = Input({
    position = { row = 1, col = 1 }, -- Will be positioned by layout
    size = { width = size.width - 4, height = 1 },
    border = {
      style = "single",
      text = {
        top = " " .. prompt .. " ",
        top_align = "left",
      },
      padding = {
        bottom = 1, -- Add 1 line of padding below input for visual spacing
      },
    },
    win_options = {
      winblend = 0,
      winhighlight = "Normal:Normal,FloatBorder:FloatBorder",
    },
  }, {
    prompt = "> ",
    default_value = "",
    on_close = function()
      -- Input closed, close entire picker
      close_picker()
      if opts.on_close then
        opts.on_close()
      end
    end,
    on_submit = function()
      -- Enter in input field: select first result if available
      if #current_results > 0 then
        local first_result = current_results[1]
        opts.on_select(first_result, 1)
        close_picker()
      end
    end,
    on_change = function(value)
      current_query = value

      -- Cancel previous debounce timer
      if debounce_timer then
        vim.fn.timer_stop(debounce_timer)
        debounce_timer = nil
      end

      -- Check minimum length
      if #value < min_query_length then
        current_results = {}
        has_more_pages = false
        next_page_token = nil
        update_menu_items()
        return
      end

      -- Debounce search
      debounce_timer = vim.fn.timer_start(debounce_ms, function()
        debounce_timer = nil
        perform_search(value, nil) -- nil = first page
      end)
    end,
  })

  -- Create results menu with initial placeholder
  results_menu = Menu({
    position = { row = 4, col = 1 }, -- Will be positioned by layout
    size = { width = size.width - 4, height = size.height - 7 },
    border = {
      style = "single",
      text = {
        top = " Results ",
        top_align = "center",
      },
      padding = {
        top = 1,    -- Add 1 line of padding above results
        bottom = 1, -- Add 1 line of padding below results
        left = 2,   -- Add 2 columns of padding on left
        right = 2,  -- Add 2 columns of padding on right
      },
    },
    win_options = {
      winblend = 0,
      winhighlight = "Normal:Normal,FloatBorder:FloatBorder,CursorLine:Visual",
    },
  }, {
    lines = {
      Menu.separator("Type at least " .. min_query_length .. " characters to search", { text_align = "center" }),
    },
    max_width = size.width - 4,
    keymap = {
      focus_next = { "j", "<Down>", "<Tab>" },
      focus_prev = { "k", "<Up>", "<S-Tab>" },
      close = { "<Esc>", "q" },
      submit = { "<CR>" },
    },
    on_submit = function(item)
      -- Item contains the original data via item.item
      if item.item then
        opts.on_select(item.item, item.index)
        close_picker()
      end
    end,
    on_close = function()
      close_picker()
      if opts.on_close then
        opts.on_close()
      end
    end,
  })

  -- Create layout (proper sizing without manual spacers)
  layout = Layout(
    {
      relative = "editor",
      position = position,
      size = size,
    },
    Layout.Box({
      Layout.Box(input_popup, { size = 3 }), -- Input box: 1 line + borders + bottom padding
      Layout.Box(results_menu, { grow = 1 }), -- Results menu: takes remaining space with internal padding
    }, { dir = "col" })
  )

  --- Update menu items based on current state
  update_menu_items = function()
    if not is_mounted or not results_menu or not results_menu.tree then
      return
    end

    local menu_items = {}

    if is_loading then
      table.insert(menu_items, Menu.separator("🔍 Searching...", { text_align = "center" }))
    elseif #current_results == 0 then
      if #current_query < min_query_length then
        table.insert(
          menu_items,
          Menu.separator("Type at least " .. min_query_length .. " characters to search", { text_align = "center" })
        )
      else
        table.insert(menu_items, Menu.separator(empty_message, { text_align = "center" }))
      end
    else
      -- Add actual results
      for idx, result in ipairs(current_results) do
        local text = opts.format_item(result, idx)
        table.insert(menu_items, Menu.item(text, { item = result, index = idx }))
      end

      -- Add "Load more" indicator if there are more pages
      if has_more_pages then
        table.insert(menu_items, Menu.separator(""))
        table.insert(menu_items, Menu.separator("Press <C-n> to load more...", { text_align = "center" }))
      end
    end

    -- Update tree and render - schedule to avoid textlock during on_change callback
    results_menu.tree:set_nodes(menu_items)
    vim.schedule(function()
      if is_mounted and results_menu and results_menu.tree then
        results_menu.tree:render()
      end
    end)
  end

  --- Perform search via callback
  ---@param query string
  ---@param page_token string|nil
  perform_search = function(query, page_token)
    is_loading = true
    update_menu_items() -- Show loading state

    opts.search_fn(query, page_token, function(items, error, more_pages_available, next_token)
      is_loading = false

      if error then
        current_results = {}
        has_more_pages = false
        next_page_token = nil
        update_menu_items()
        vim.notify("Search failed: " .. vim.inspect(error), vim.log.levels.ERROR)
        return
      end

      -- Append or replace results
      if page_token then
        -- Appending next page
        vim.list_extend(current_results, items or {})
      else
        -- First page
        current_results = items or {}
      end

      has_more_pages = more_pages_available or false
      next_page_token = next_token

      update_menu_items()
    end)
  end

  --- Load next page
  local function load_next_page()
    if has_more_pages and next_page_token and not is_loading then
      perform_search(current_query, next_page_token)
    end
  end

  --- Switch focus to input field
  local function focus_input()
    if input_popup and input_popup.winid and vim.api.nvim_win_is_valid(input_popup.winid) then
      vim.api.nvim_set_current_win(input_popup.winid)
      vim.cmd("startinsert")
    end
  end

  -- Additional keymaps for results menu
  results_menu:map("n", "<C-n>", function()
    load_next_page()
  end, { noremap = true })

  results_menu:map("n", "i", function()
    focus_input()
  end, { noremap = true })

  -- Allow switching focus between input and results with Tab
  input_popup:map("i", "<Tab>", function()
    if results_menu and results_menu.winid and vim.api.nvim_win_is_valid(results_menu.winid) then
      vim.api.nvim_set_current_win(results_menu.winid)
    end
  end, { noremap = true })

  input_popup:map("i", "<C-n>", function()
    -- Navigate down in results from input
    if results_menu and results_menu.menu_props then
      results_menu.menu_props.on_focus_next()
    end
  end, { noremap = true })

  input_popup:map("i", "<C-p>", function()
    -- Navigate up in results from input
    if results_menu and results_menu.menu_props then
      results_menu.menu_props.on_focus_prev()
    end
  end, { noremap = true })

  input_popup:map("i", "<Esc>", function()
    close_picker()
    if opts.on_close then
      opts.on_close()
    end
  end, { noremap = true })

  -- Normal mode keymaps for input: j/k navigate results menu
  input_popup:map("n", "j", function()
    -- Navigate down in results from input normal mode
    if results_menu and results_menu.menu_props then
      results_menu.menu_props.on_focus_next()
    end
  end, { noremap = true })

  input_popup:map("n", "k", function()
    -- Navigate up in results from input normal mode
    if results_menu and results_menu.menu_props then
      results_menu.menu_props.on_focus_prev()
    end
  end, { noremap = true })

  input_popup:map("n", "i", function()
    -- Return to insert mode at current cursor position
    vim.cmd("startinsert")
  end, { noremap = true })

  input_popup:map("n", "a", function()
    -- Append mode: move cursor right then enter insert
    local cursor = vim.api.nvim_win_get_cursor(0)
    local line_length = vim.api.nvim_buf_line_count(input_popup.bufnr)
    if line_length > 0 then
      local line = vim.api.nvim_buf_get_lines(input_popup.bufnr, 0, 1, false)[1] or ""
      -- Only move right if not at end of line
      if cursor[2] < #line then
        vim.api.nvim_win_set_cursor(0, { cursor[1], cursor[2] + 1 })
      end
    end
    vim.cmd("startinsert")
  end, { noremap = true })

  input_popup:map("n", "A", function()
    -- Append at end of line
    vim.cmd("startinsert!")
  end, { noremap = true })

  input_popup:map("n", "<CR>", function()
    -- Enter in normal mode: select first result if available
    if #current_results > 0 then
      local first_result = current_results[1]
      opts.on_select(first_result, 1)
      close_picker()
    end
  end, { noremap = true })

  input_popup:map("n", "<Esc>", function()
    close_picker()
    if opts.on_close then
      opts.on_close()
    end
  end, { noremap = true })

  -- Mount layout
  layout:mount()

  -- Initial menu render
  update_menu_items()

  -- Focus input field and enter insert mode
  vim.schedule(function()
    focus_input()
  end)

  return layout
end

return M
