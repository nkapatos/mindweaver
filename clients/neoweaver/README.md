# Neoweaver

Neovim client for MindWeaver. Provides note management commands inside Neovim, communicating with the MindWeaver server over the Connect RPC API.

## Prerequisites

- **Neovim 0.11+** - Required for plugin functionality
- **[plenary.nvim](https://github.com/nvim-lua/plenary.nvim)** - Required dependency for HTTP requests

## Quick Start

Install with your preferred package manager. Example using **lazy.nvim**:

```lua
return {
  {
    dir = "path/to/neoweaver",
    cmd = { "NotesList", "NotesOpen", "NotesNew", "MwServerUse", "MwToggleDebug" },
    dependencies = { "nvim-lua/plenary.nvim" },
    opts = {
      allow_multiple_empty_notes = false,
      api = {
        servers = {
          local = { url = "http://localhost:9421", default = true },
        },
        debug_info = true,
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
- Toggles API logging. Can be toggled at runtime with `:MwToggleDebug`.

## Commands

| Command          | Description                                     |
| ---------------- | ----------------------------------------------- |
| `:NotesList`     | Fetch the first page of notes and open a picker |
| `:NotesOpen`     | Open a note by ID                               |
| `:NotesNew`      | Create a new note on the server and open buffer |
| `:MwServerUse`   | Switch to a configured backend server           |
| `:MwToggleDebug` | Toggle API debug notifications                  |

## Keymaps

Default mappings (can be remapped):

| Mapping      | Action             |
| ------------ | ------------------ |
| `<leader>nl` | `NotesList`        |
| `<leader>no` | Prompt for note ID |
| `<leader>nn` | `NotesNew`         |

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
