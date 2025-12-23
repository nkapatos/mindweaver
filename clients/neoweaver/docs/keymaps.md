# Keymaps

## Default Keymaps

When keymaps are enabled (`keymaps = { enabled = true }`), the following default mappings are available:

| Mapping        | Command                      | Action                          |
| -------------- | ---------------------------- | ------------------------------- |
| `<leader>nl`   | `:NeoweaverNotesList`        | List all notes                  |
| `<leader>no`   | `:NeoweaverNotesOpen`        | Open note (prompts for ID)      |
| `<leader>nn`   | `:NeoweaverNotesNew`         | Create new untitled note        |
| `<leader>nt`   | `:NeoweaverNotesTitle`       | Edit current note title         |
| `<leader>nd`   | `:NeoweaverNotesDelete`      | Delete note (prompts for ID)    |
| `<leader>qn`   | `:NeoweaverNotesQuick`       | Capture quicknote               |
| `<leader>ql`   | `:NeoweaverNotesQuickList`   | List quicknotes (see issue #45) |
| `<leader>qa`   | `:NeoweaverNotesQuickAmend`  | Reopen last quicknote (see #45) |
| `<leader>.n`   | `:NeoweaverNotesQuick`       | Capture quicknote (fast)        |
| `<leader>.l`   | `:NeoweaverNotesQuickList`   | List quicknotes (see issue #45) |
| `<leader>.a`   | `:NeoweaverNotesQuickAmend`  | Reopen last quicknote (see #45) |

## Customizing Keymaps

You can override individual keymaps or disable them entirely:

**Custom mappings:**

```lua
require('neoweaver').setup({
  keymaps = {
    enabled = true,
    notes = {
      list = "<leader>fn",      -- Changed from <leader>nl
      open = "<leader>fo",      -- Changed from <leader>no
      new = "<leader>fc",       -- Changed from <leader>nn
      new_with_title = "<leader>fC",
      title = "<leader>fr",
      delete = "<leader>fx",
    },
  },
})
```

**Disable built-in keymaps:**

```lua
require('neoweaver').setup({
  keymaps = {
    enabled = false,  -- No keymaps created
  },
})
```

Then create your own keymaps:

```lua
vim.keymap.set('n', '<leader>fn', ':NeoweaverNotesList<CR>', { desc = 'List notes' })
vim.keymap.set('n', '<leader>fc', ':NeoweaverNotesNew<CR>', { desc = 'Create note' })
-- etc.
```
