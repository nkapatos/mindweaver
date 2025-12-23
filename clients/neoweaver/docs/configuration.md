# Configuration

## Configuration Options

### allow_multiple_empty_notes

- **Type:** boolean
- **Default:** `false`

When `true`, new note buffers are marked as modified immediately, allowing multiple untitled notes without Neovim blocking on `:w`.

```lua
require('neoweaver').setup({
  allow_multiple_empty_notes = true,
})
```

### api.servers

- **Type:** table (required)
- **Description:** Map of server names to configuration

Each server entry must provide a `url`. Set `default = true` on one entry to select it automatically.

**Single server example:**

```lua
api = {
  servers = {
    local = { url = "http://localhost:9421", default = true },
  },
}
```

**Multiple servers example:**

```lua
api = {
  servers = {
    local = { url = "http://localhost:9421", default = true },
    cloud = { url = "https://api.mindweaver.example.com" },
    staging = { url = "https://staging.mindweaver.example.com" },
  },
}
```

Switch between servers using `:NeoweaverServerUse <server_name>`.

### api.debug_info

- **Type:** boolean
- **Default:** `true`

Toggles API request/response logging. Can be toggled at runtime with `:NeoweaverToggleDebug`.

```lua
api = {
  debug_info = false,  -- Disable debug logging
}
```

### keymaps

- **Type:** table
- **Default:** disabled

Enable built-in default keymaps by passing `keymaps = { enabled = true }`. You can override any mapping by providing the `notes` table.

**Enable defaults:**

```lua
keymaps = {
  enabled = true,
}
```

**Customize keymaps:**

```lua
keymaps = {
  enabled = true,
  notes = {
    list = "<leader>nl",
    open = "<leader>no",
    new = "<leader>nn",
    new_with_title = "<leader>nN",
    title = "<leader>nt",
    delete = "<leader>nd",
  },
}
```

### quicknotes

- **Type:** table
- **Default:** see below

Quicknotes configure the floating capture window. The default configuration matches the built-in capture flow:

```lua
quicknotes = {
  title_template = "%Y%m%d%H%M",
  collection_id = 2,
  note_type_id = 2,
  popup = {
    relative = "editor",
    position = "50%",
    size = {
      width = "40%",
      height = "20%",
    },
    border = {
      style = "rounded",
      text = {
        top = "Quick Note",
        top_align = "center",
      },
    },
  },
}
```

Notes:

- `title_template` is passed to `strftime()` when a quicknote saves. Use any valid strftime pattern.
- `collection_id` and `note_type_id` are currently hard-coded defaults while user preferences sync is designed.
- Metadata enrichment is not yet configurable (see issue #47).

## Complete Setup Example

```lua
require('neoweaver').setup({
  allow_multiple_empty_notes = true,
  quicknotes = {
    popup = {
      position = {
        row = 4,
        col = 10,
      },
      size = {
        width = 80,
        height = 20,
      },
      border = {
        text = {
          top = "Scratchpad",
        },
      },
    },
  },

}
```

Notes:

- `title_template` is passed to `strftime()` when a quicknote saves. Use any valid strftime pattern.
- `collection_id` and `note_type_id` are currently hard-coded defaults while user preferences sync is designed.
- Metadata enrichment is not yet configurable (see issue #47).

## Complete Setup Example

```lua
require('neoweaver').setup({
  allow_multiple_empty_notes = true,
  quicknotes = {
    window = {
      width = 80,
      height = 20,
      row = 2,
      col = 10,
      title = "Scratchpad",
    },
  },
  api = {
    servers = {
      local = { url = "http://localhost:9421", default = true },
      cloud = { url = "https://api.mindweaver.example.com" },
    },
    debug_info = true,
  },
  keymaps = {
    enabled = true,
    notes = {
      list = "<leader>nl",
      open = "<leader>no",
      new = "<leader>nn",
      new_with_title = "<leader>nN",
      title = "<leader>nt",
      delete = "<leader>nd",
    },
    quicknotes = {
      new = "<leader>qn",
      new_fast = "<leader>.n",
    },
  },
})
```
