# Commands

Neoweaver provides the following commands for note management:

| Command                         | Description                                    |
| ------------------------------- | ---------------------------------------------- |
| `:NeoweaverNotesList`           | Fetch and display the first page of notes     |
| `:NeoweaverNotesOpen`           | Open a note by ID (prompts for ID)            |
| `:NeoweaverNotesNew`            | Create a new untitled note                     |
| `:NeoweaverNotesQuick`          | Open an ephemeral quicknote capture window     |
| `:NeoweaverNotesQuickList`      | List saved quicknotes (see issue #45)          |
| `:NeoweaverNotesQuickAmend`     | Reopen last quicknote (see issue #45)          |
| `:NeoweaverNotesTitle`          | Edit the title of the currently active note   |
| `:NeoweaverNotesDelete`         | Delete a note by ID (prompts for ID)          |
| `:NeoweaverServerUse`           | Switch to a configured backend server          |
| `:NeoweaverToggleDebug`         | Toggle API debug notifications on/off          |

## Usage Examples

### Creating Notes

```vim
" Create a new untitled note
:NeoweaverNotesNew

" Capture a quicknote
:NeoweaverNotesQuick
```

### Managing Notes

```vim
" List all notes
:NeoweaverNotesList

" Open a specific note
:NeoweaverNotesOpen
" Then enter the note ID when prompted

" Edit the title of the current note
:NeoweaverNotesTitle

" Delete a note
:NeoweaverNotesDelete
```

### Server Management

```vim
" Switch to a different configured server
:NeoweaverServerUse cloud

" Toggle debug logging
:NeoweaverToggleDebug
```
