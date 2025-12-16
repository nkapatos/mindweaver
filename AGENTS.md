# MindWeaver Monorepo

## Structure

- **`packages/`** - Go workspace containing server (`mindweaver/`) and shared packages (`pkg/`)
- **`clients/`** - Client implementations: `neoweaver` (Neovim), web (planned), browser plugins (planned)
- **`proto/`** - Protocol buffer definitions

## Task Management

Components have their own Taskfiles for build, run, and database operations. Use `task --list` from the component directory to see available tasks.

## Language Guidelines

- **Go development:** @docs/golang.md
- **Lua client development:** @docs/lua.md

## General Rules

- **Tooling requirements:** @rules/tooling.md
- **Version management:** @rules/versions.md

## Project-Specific Documentation

### When working in `packages/mindweaver/`:
- Architecture patterns: @packages/mindweaver/docs/architecture.md
- API design: @packages/mindweaver/docs/api.md
- Code style rules: @packages/mindweaver/rules/code-style.md

### When working in `clients/neoweaver/`:
- Development guidelines: @clients/neoweaver/docs/guidelines.md
- Conventions: @clients/neoweaver/rules/conventions.md
