# Testing Neoweaver V3 - Minimal List Notes

## Setup

1. **Make sure server is running** with v3 API:
   ```bash
   task mind:serve
   # Should be running on localhost:9420
   ```

2. **Load plugin in Neovim**:
   ```lua
   -- In your init.lua or test file
   require('neoweaver').setup()
   ```

## Test: List Notes

### Command
```vim
:NotesList
```

### Keymap
```vim
<leader>nl
```

### Expected Behavior

1. Plugin loads with notification: "Neoweaver v3 loaded!"
2. Notes module loads: "Neoweaver (v3) notes module loaded"
3. Command `:NotesList` triggers API call
4. `vim.ui.select` shows list of notes with format: `[<id>] <title>`
5. Selecting a note shows notification with note details

### Debugging

If something fails:

1. **Check API connection**:
   ```vim
   :lua require('neoweaver.api').config
   ```
   Should show `server_url.dev = "http://localhost:9420"`

2. **Toggle debug mode**:
   ```vim
   :MwToggleDebug
   ```
   Then run `:NotesList` again to see debug output

3. **Check types are loaded**:
   ```vim
   :lua print(vim.inspect(require('neoweaver.types')))
   ```

## What Works

- ✅ Type generation (proto → TS → Lua)
- ✅ API layer (Connect RPC)
- ✅ List notes with proper v3 request/response types
- ✅ Basic error handling
- ✅ vim.ui.select display

## What Doesn't Work Yet

- ❌ Opening/editing notes
- ❌ Creating notes
- ❌ Saving notes
- ❌ Deleting notes
- ❌ Quicknotes
- ❌ Conflict resolution
- ❌ Metadata extraction

## Next Steps

After confirming list notes works:
1. Add "view note" - open in buffer (read-only)
2. Add "edit note" - enable saving
3. Add "create note" - new note from scratch
4. Then build from there incrementally

## File Structure

```
clients/neoweaver/
├── lua/neoweaver/
│   ├── api.lua          # ✅ Connect RPC client
│   ├── types.lua        # ✅ Generated types
│   ├── init.lua         # ✅ Entry point
│   ├── notes.lua        # ✅ Minimal list implementation
│   └── notes/
│       └── buffer.lua   # ✅ Buffer utilities (for later)
```

## Common Issues

### "module 'neoweaver' not found"
Make sure the plugin path is in your runtimepath:
```lua
vim.opt.runtimepath:append('~/path/to/mindweaver/clients/neoweaver')
```

### "Connection refused"
Server not running or wrong URL. Check with:
```bash
curl http://localhost:9420/mind.v3.NotesService/ListNotes \
  -H "Content-Type: application/json" \
  -d '{"pageSize": 10}'
```

### "No notes found"
Database is empty. Add some test notes first.
