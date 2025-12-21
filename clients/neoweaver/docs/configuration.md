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

## Complete Setup Example

```lua
require('neoweaver').setup({
  allow_multiple_empty_notes = true,
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
  },
})
```
