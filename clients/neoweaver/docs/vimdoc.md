# Neoweaver

Neovim client for MindWeaver - Manage your notes directly from Neovim.

## Introduction

Neoweaver is a Neovim plugin that provides note management capabilities by communicating with the MindWeaver server over the Connect RPC API. It allows you to create, edit, and organize notes directly within Neovim using markdown buffers.

```{.include}
docs/installation.md
```

```{.include}
docs/configuration.md
```

```{.include}
docs/commands.md
```

```{.include}
docs/keymaps.md
```

# API Reference

For detailed API documentation including function signatures, parameters, and types, see:

`:help neoweaver-api`

The API reference is auto-generated from source code annotations and includes documentation for:

- `require('neoweaver').setup()` - Plugin configuration
- Internal modules and functions
- Type definitions

# Architecture

The plugin is organized into the following structure:

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

# See Also

- Project repository: https://github.com/nkapatos/mindweaver
- MindWeaver documentation: (link to server docs when available)

---

*For development and contribution guidelines, see the project repository.*
