# Neoweaver

Neoweaver is the MindWeaver Neovim client. It communicates with the MindWeaver backend over the v3 Connect RPC API and exposes note-management commands inside Neovim.

## Installation

The plugin can be installed with any package manager. The example below uses **lazy.nvim** and configures the client to talk to a locally running MindWeaver server.

```lua
return {
  {
    dir = vim.fn.expand("~/Workbench/projects/nvmw/mindweaver/clients/neoweaver"),
    dev = true,
    cmd = { "NotesList", "NotesOpen", "NotesNew", "MwServerUse", "MwToggleDebug" },
    dependencies = { "nvim-lua/plenary.nvim" },
    opts = {
      allow_multiple_empty_notes = false,
      api = {
        servers = {
          work = { url = "http://localhost:9421", default = true },
          cloud = "https://api.example.com",
        },
        debug_info = true,
      },
    },
  },
}
```

## Configuration

Neoweaver accepts the following options via `require("neoweaver").setup()`.

### `allow_multiple_empty_notes`

- **Type:** boolean (default: `false`)
- When `true`, newly created note buffers are marked as modified immediately so that you can create and discard multiple untitled notes in succession without Neovim blocking on `:w`. When `false`, buffers start unmodified to prevent accidental empty notes.

### `api.servers`

- **Type:** table (required)
- A map of server names to configuration tables or URL strings. Each entry must provide a `url`. Set `default = true` on one entry to select it automatically; otherwise run `:MwServerUse <name>` to choose a backend after setup.

Example:

```lua
api = {
  servers = {
    work = { url = "http://localhost:9421", default = true },
    cloud = "https://api.example.com",
  },
}
```

### `api.debug_info`

- **Type:** boolean (default: `true`)
- Toggles informational logging from the API module. You can switch this at runtime with `:MwToggleDebug`.

## Commands

| Command         | Description                                     |
|-----------------|-------------------------------------------------|
| `:NotesList`    | Fetch the first page of notes and open a picker |
| `:NotesOpen`    | Open a note by ID                               |
| `:NotesNew`     | Create a new note on the server and open buffer |
| `:MwServerUse`  | Switch to a configured backend server           |
| `:MwToggleDebug`| Toggle API debug notifications                  |

## Keymaps

During setup Neoweaver registers the following default mappings. Remap or clear them if needed.

| Mapping        | Action            |
|----------------|-------------------|
| `<leader>nl`   | `NotesList`       |
| `<leader>no`   | Prompt for note ID|
| `<leader>nn`   | `NotesNew`        |

## Notes

- The plugin expects a running MindWeaver server that exposes the v3 RPC API.
- Conflict handling and metadata extraction are still being migrated from the legacy `mw` client.
