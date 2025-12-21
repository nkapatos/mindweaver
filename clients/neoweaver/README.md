# Neoweaver

Neovim client for MindWeaver. Manage your notes directly from Neovim, communicating with the MindWeaver server over the Connect RPC API.

## Quick Start

Install with **lazy.nvim**:

```lua
{
  "nkapatos/neoweaver.nvim",
  dependencies = {
    "nvim-lua/plenary.nvim",
    "MunifTanjim/nui.nvim",
  },
  opts = {
    api = {
      servers = {
        local = { url = "http://localhost:9421", default = true },
      },
    },
  },
}
```

For complete installation instructions including **vim.pack()** (Neovim 0.12+), see [Installation](docs/installation.md) or `:help neoweaver-installation`.

## Documentation

**For complete documentation, see `:help neoweaver` inside Neovim.**

The help system provides comprehensive documentation including:

- `:help neoweaver-installation` - Installation instructions
- `:help neoweaver-configuration` - Configuration options
- `:help neoweaver-commands` - Available commands
- `:help neoweaver-keymaps` - Keymap configuration
- `:help neoweaver-api` - API reference (auto-generated from source)

You can also read the documentation in markdown format:

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Commands](docs/commands.md)
- [Keymaps](docs/keymaps.md)

## Development

### Prerequisites

- Go 1.23+
- Task
- buf (protocol buffers)
- Neovim 0.11+

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

### Documentation Generation

```bash
# Generate API reference from LuaLS annotations
task neoweaver:docs:api

# Generate user guide from markdown
task neoweaver:docs:gen:panvimdoc

# Generate help tags
task neoweaver:docs:tags
```

For development guidelines, see [docs/dev/guidelines.md](docs/dev/guidelines.md).

## Architecture

```
lua/neoweaver/
├── init.lua          - Public API and setup
├── _internal/
│   ├── api.lua       - HTTP client for Connect RPC API
│   ├── notes.lua     - Note operations and commands
│   ├── buffer/
│   │   ├── manager.lua   - Buffer lifecycle management
│   │   └── statusline.lua - Status line integration
│   └── explorer/
│       ├── init.lua      - Explorer entry point
│       ├── tree.lua      - Tree rendering
│       └── window.lua    - Window management
└── types.lua         - Generated Lua type annotations
```

The plugin expects a running MindWeaver server exposing the v3 RPC API.

## Acknowledgements

This plugin builds upon excellent work from the Neovim community:

- **[nui.nvim](https://github.com/MunifTanjim/nui.nvim)** by [@MunifTanjim](https://github.com/MunifTanjim) - Provides the UI primitives (NuiTree, NuiSplit, NuiLine) that power the explorer interface. The clean API design and robust buffer/window management made building our tree explorer straightforward and reliable.

- **[neo-tree.nvim](https://github.com/nvim-neo-tree/neo-tree.nvim)** - An excellent file explorer that served as a reference implementation for managing the NuiSplit and NuiTree lifecycle. Their approach to buffer management, state preservation, and tree rendering patterns provided valuable insights during our architecture design.

Special thanks to both projects for their well-documented code and thoughtful API design, which made it possible to build a robust, maintainable explorer with proper separation of concerns.

## See Also

- [Root README](../../README.md) - Monorepo overview
- [Workflow Documentation](../../docs/workflow.md) - Contribution guidelines
