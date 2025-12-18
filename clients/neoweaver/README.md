# Neoweaver

Neovim client for MindWeaver. Provides note management commands inside Neovim, communicating with the MindWeaver server over the Connect RPC API.

## Prerequisites

- **Neovim 0.11+** - Required for plugin functionality
- **[plenary.nvim](https://github.com/nvim-lua/plenary.nvim)** - Required dependency for HTTP requests
- **[nui.nvim](https://github.com/MunifTanjim/nui.nvim)** - Required dependency for UI components

## Quick Start

Install with your preferred package manager. Example using **lazy.nvim**:

```lua
return {
  {
    dir = "path/to/neoweaver",
    cmd = {
      "NeoweaverNotesList",
      "NeoweaverNotesOpen",
      "NeoweaverNotesNew",
      "NeoweaverNotesNewWithTitle",
      "NeoweaverNotesTitle",
      "NeoweaverServerUse",
      "NeoweaverToggleDebug",
    },
    dependencies = {
      "nvim-lua/plenary.nvim",
      "MunifTanjim/nui.nvim",
    },
    opts = {
      allow_multiple_empty_notes = false,
      api = {
        servers = {
          local = { url = "http://localhost:9421", default = true },
        },
        debug_info = true,
      },
      keymaps = {
        enabled = true,
      },
    },
  },
}
```

## Development

### Code Generation

Generated Lua type annotations are not committed. Run generation when protocol buffer definitions change:

```bash
# Show available tasks
task --list

# Generate Lua types from protobuf (full pipeline)
task neoweaver:types:generate

# Clean generated files
task neoweaver:types:clean
```

Generation pipeline:
1. Generate TypeScript from protobuf
2. Convert TypeScript to Lua type annotations
3. Clean up temporary files

### Testing

See [TESTING.md](TESTING.md) for test suite documentation.

## Configuration

### `allow_multiple_empty_notes`

- **Type:** boolean (default: `false`)
- When `true`, new note buffers are marked as modified immediately, allowing multiple untitled notes without Neovim blocking on `:w`.

### `api.servers`

- **Type:** table (required)
- Map of server names to configuration. Each entry must provide a `url`. Set `default = true` on one entry to select it automatically.

Example:

```lua
api = {
  servers = {
    local = { url = "http://localhost:9421", default = true },
    cloud = "https://api.example.com",
  },
}
```

### `api.debug_info`

- **Type:** boolean (default: `true`)
- Toggles API logging. Can be toggled at runtime with `:NeoweaverToggleDebug`.

### `keymaps`

- **Type:** table (default: disabled)
- Enable built-in defaults by passing `keymaps = { enabled = true }`.
- Override any mapping by providing the `notes` or `quicknotes` tables.

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

## Commands

| Command                    | Description                            |
| ------------------------- | -------------------------------------- |
| `:NeoweaverNotesList`     | Fetch the first page of notes and pick |
| `:NeoweaverNotesOpen`     | Open a note by ID                      |
| `:NeoweaverNotesNew`      | Create a server-backed untitled note   |
| `:NeoweaverNotesNewWithTitle` | Prompt for a title before creating |
| `:NeoweaverNotesTitle`    | Edit the active note title             |
| `:NeoweaverNotesDelete`   | Delete a note by ID                    |
| `:NeoweaverServerUse`     | Switch to a configured backend server  |
| `:NeoweaverToggleDebug`   | Toggle API debug notifications         |

## Keymaps

Default mappings (can be remapped individually):

| Mapping        | Action                          |
| -------------- | -------------------------------- |
| `<leader>nl`   | `NeoweaverNotesList`            |
| `<leader>no`   | Prompt for note ID (open/edit)  |
| `<leader>nn`   | `NeoweaverNotesNew`             |
| `<leader>nN`   | `NeoweaverNotesNewWithTitle`    |
| `<leader>nt`   | `NeoweaverNotesTitle`           |
| `<leader>nd`   | `NeoweaverNotesDelete`          |

## Architecture

```
lua/neoweaver/
├── api.lua           - HTTP client for Connect RPC API
├── notes.lua         - Note operations and commands
├── buffer/
│   └── manager.lua   - Buffer lifecycle management
└── types.lua         - Generated Lua type annotations
```

The plugin expects a running MindWeaver server exposing the v3 RPC API.

## See Also

- [Root README](../../README.md) - Project overview
- [docs/WORKFLOW.md](../../docs/WORKFLOW.md) - Contribution guidelines
- [docs/guidelines.md](docs/guidelines.md) - Development guidelines
- [rules/conventions.md](rules/conventions.md) - Code conventions
