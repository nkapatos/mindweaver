# Installation

## Prerequisites

- **Neovim 0.11+** - Required for plugin functionality
- **[plenary.nvim](https://github.com/nvim-lua/plenary.nvim)** - Required dependency for HTTP requests
- **[nui.nvim](https://github.com/MunifTanjim/nui.nvim)** - Required dependency for UI components
- **MindWeaver server** - Running and accessible

## Using lazy.nvim (Neovim 0.11+)

```lua
return {
  {
    "nkapatos/neoweaver.nvim",
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
      api = {
        servers = {
          local = { url = "http://localhost:9421", default = true },
        },
      },
    },
  },
}
```

## Using vim.pack() (Neovim 0.12+)

Neovim 0.12 introduces native package management with `vim.pack()`.

```lua
-- In your init.lua
vim.pack({
  name = "neoweaver",
  url = "https://github.com/nkapatos/neoweaver.nvim",
  dependencies = {
    { url = "https://github.com/nvim-lua/plenary.nvim" },
    { url = "https://github.com/MunifTanjim/nui.nvim" },
  },
})

require('neoweaver').setup({
  api = {
    servers = {
      local = { url = "http://localhost:9421", default = true },
    },
  },
})
```

## Basic Setup

After installation, configure the plugin with at least one server:

```lua
require('neoweaver').setup({
  api = {
    servers = {
      local = { url = "http://localhost:9421", default = true }
    }
  }
})
```

For more configuration options, see `:help neoweaver-configuration`.
