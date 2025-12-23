# UI Components

## Picker

Generic entity picker using `nui.menu` - decoupled from domain logic.

### Usage

```lua
local picker = require("neoweaver._internal.ui.picker")

-- Example: Pick from a list of notes
picker.pick(notes, {
  prompt = "Select a note",
  format_item = function(note, idx)
    return string.format("[%d] %s", note.id, note.title)
  end,
  on_submit = function(note, idx)
    -- Handle selection
    open_note(note.id)
  end,
  on_close = function()
    -- Optional: handle picker close without selection
  end,
  -- Optional: override config defaults
  size = { width = 60, height = 20 },
  position = "50%",
  border = { style = "rounded" },
})
```

### Configuration

Users can configure picker defaults in their setup:

```lua
require("neoweaver").setup({
  picker = {
    size = { width = 80, height = 25 },
    position = "50%",
    border = { style = "rounded" },
  }
})
```

### Keymaps

Default navigation:
- `j`, `<Down>`, `<Tab>` - Next item
- `k`, `<Up>`, `<S-Tab>` - Previous item
- `<CR>`, `<Space>` - Select item
- `q`, `<Esc>`, `<C-c>` - Close picker

### Reusability

The picker is completely decoupled from domain logic. Use it for:
- Notes listing (implemented)
- Collections listing (future)
- Tasks listing (future)
- Any entity selection UI

Simply provide:
1. Array of items
2. Format function for display
3. Submit callback for selection
