# Tree Architecture Refactoring Plan

## Current State (Working)
Branch: `feat/notes-list-field-masking`
Commit: `d8c7a48` - Server context in explorer and buffer statuslines

## Goal: Three-Layer Architecture

### Layer 1: Explorer (Container/Orchestrator)
- Owns the buffer (persists for entire session)
- Sets up keymaps **once** on initialization
- Manages "modes" or "views" (collections, tags, mixed, etc.)
- Delegates actions to domain modules based on current mode
- Keymaps are semantic: `a` = "add", `r` = "rename", `d` = "delete", `R` = "refresh"

### Layer 2: Tree (Generic Data Structure & Renderer)
- Pure tree logic - nodes, edges, hierarchy
- No domain knowledge (doesn't know what collections, tags, notes are)
- Provides generic operations: `get_node()`, `expand()`, `collapse()`, `render()`
- Accepts generic node structure: `{ id, type, name, children[], data }`
- Recursive build logic works for **any** hierarchical data

### Layer 3: Domain Modules (collections.lua, tags.lua, etc.)
- Know their API and business rules
- Provide callbacks for actions: `on_create()`, `on_rename()`, `on_delete()`, `on_refresh()`
- Build domain-specific data into generic tree structure
- Handle their own API calls and validation

---

## Phase 1: Explorer Owns Buffer & Keymaps

###  Goals
- Move buffer lifecycle management from tree to explorer
- Move keymap setup from tree to explorer (one-time setup)
- Explorer tracks state (bufnr, keymaps_configured, current_mode)

### Changes

**File: `explorer/init.lua`**
- Add state: `{ bufnr, keymaps_configured, current_mode }`
- Add `setup_keymaps(bufnr)` - sets up once
- Keymaps call `M.handle_create()`, `M.handle_rename()`, `M.handle_delete()`
- These handlers delegate to mode-specific functions
- Track bufnr in explorer state instead of tree

**File: `explorer/tree.lua`**
- Remove `state.bufnr` tracking
- Remove `state.keymaps_configured` tracking
- Remove `setup_keymaps()` function entirely
- `init()`, `refresh()` accept bufnr as parameter (don't store it)
- Tree becomes stateless regarding buffer

### Testing
- [ ] Explorer opens and displays tree
- [ ] Keymaps work (a, r, d, R, Enter, hjkl)
- [ ] Create/rename/delete collections work
- [ ] Refresh works
- [ ] Toggle explorer preserves state

---

## Phase 2: Make Tree Generic

### Goals
- Tree accepts generic node structure
- Remove collection/note specific logic from tree
- Tree only knows about rendering nodes with standard fields

### Standard Node Structure
```lua
---@class GenericTreeNode
---@field id string|number Unique identifier
---@field type string Node type ("server", "collection", "note", "tag")
---@field name string Display name
---@field icon? string Icon override (optional)
---@field highlight? string Highlight group override (optional)
---@field data table Domain-specific data (the original entity)
---@field children? GenericTreeNode[] Child nodes
```

### Changes

**File: `explorer/tree.lua`**
- Rename `build_nodes()` to `build_hierarchy()` - accepts generic nodes
- `build_tree()` accepts pre-built node array (not collections data)
- Remove `count_notes()` helper (domain-specific)
- Remove `load_and_render()` - this becomes explorer's responsibility
- Keep only:
  - `M.build_tree(bufnr, nodes)` - Build NuiTree from generic nodes
  - `M.render()` - Render current tree
  - `M.get_node()` - Get node at cursor
  - Generic node operations (expand, collapse, etc.)

**File: `explorer/init.lua`**
- Add `load_tree_data()` - fetches collections/notes, builds nodes
- Call `tree.build_tree(bufnr, nodes)` with prepared data
- Handle loading state, statusline updates

### Testing
- [ ] Explorer still displays collections tree
- [ ] All functionality works as before
- [ ] Tree module has no domain knowledge

---

## Phase 3: Move Domain Logic to Collections Module

### Goals
- Collections module builds tree-compatible data
- Collections module owns all collection-specific logic
- Actions are cleanly separated

### Changes

**File: `collections.lua`**
- Add `M.build_tree_nodes()` - Builds generic tree nodes from collections
- Move server node creation here
- Move collection/note node building here
- Returns array of `GenericTreeNode`

**File: `explorer/init.lua`**
- Use `collections.build_tree_nodes()` to get data
- Action handlers stay here (already domain-specific)

### Testing
- [ ] Collections tree still works
- [ ] Server nodes appear correctly
- [ ] Create/rename/delete work
- [ ] Refresh works

---

## Phase 4: Add Mode Management

### Goals
- Explorer supports multiple modes (collections, tags, etc.)
- Easy to add new modes
- Mode switching infrastructure

### Changes

**File: `explorer/init.lua`**
- Add mode registry:
  ```lua
  local modes = {
    collections = {
      load_data = collections.build_tree_nodes,
      actions = { create = ..., rename = ..., delete = ... }
    }
  }
  ```
- `M.switch_mode(mode_name)` - switches active mode
- Action handlers check `state.current_mode` and dispatch
- Initial load uses default mode ("collections")

**File: `collections.lua`**
- Export action functions:
  - `M.handle_create(node, bufnr)`
  - `M.handle_rename(node, bufnr)`
  - `M.handle_delete(node, bufnr)`

### Testing
- [ ] Collections mode works as before
- [ ] Infrastructure ready for tags mode (future)

---

## Rollback Strategy

Each phase can be rolled back independently:
- After Phase 1: `git checkout d8c7a48 -- clients/neoweaver/lua/neoweaver/_internal/explorer/`
- After Phase 2: Revert tree.lua changes only
- After Phase 3: Revert collections.lua changes only
- After Phase 4: Revert mode management additions only

## Testing Between Phases

After each phase, run full manual test:
1. Open explorer
2. Create collection
3. Rename collection
4. Delete collection
5. Refresh
6. Toggle explorer
7. Open note

If anything breaks, rollback and fix before proceeding.

---

## Future Enhancements (Post-Refactor)

- Add tags mode
- Add mixed view mode
- Add winbar for mode switching UI
- Smart refresh (preserve expanded state)
- External change detection
- Multi-server view (show all servers)
