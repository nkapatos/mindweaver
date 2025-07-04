# AGENTS.md - Agentic Coding Guidelines for Mindweaver

## Build, Lint, Test Commands
- Build all: `task build`
- Build Go only: `task build:go`
- Build templates only: `task build:templ`
- Build frontend (TS/CSS): `task build:esbuild`
- Live reload all: `task live`
- Live reload templates only: `task live:templ`
- Live reload frontend only: `task live:esbuild`
- Format templates: `task format:templ`
- Run all tests: `go test ./...`
- Run single test: `go test -run ^TestName$ ./path/to/package`

## Code Style Guidelines
- Follow Go idiomatic style and standards
- Use structured logging with slog
- Use sqlc for type-safe DB queries
- Organize imports: stdlib, external, internal
- Use clear, descriptive names for variables, functions, types
- Error handling: check and return errors promptly
- Use UUID v7 for message IDs
- Frontend uses Templ templates with Tailwind CSS and DaisyUI
- Frontend structure: elements (smallest), components (composed), layouts, pages
- Avoid frontend frameworks; prefer Alpine.js and HTMX for interactivity

## Project Structure
- `cmd/mindweaver`: app entry point
- `internal/`: core app code (adapters, handlers, router, services, store, templates)
- `assets/`: frontend assets (TS, CSS)
- `migrations/`: DB migrations
- `sql/`: SQLC schema and queries
- `test/`: tests

## Cursor Rules Highlights
- Providers combine LLM service, model, and config
- Frontend templ files follow https://templ.guide/llms.md best practices
- Use DaisyUI styles with Tailwind CSS
- Frontend structure: elements, components, layouts, pages
- Avoid frontend frameworks; use Alpine.js and HTMX

## Copilot Instructions
- No specific Copilot instructions found

---
This file guides agentic coding agents for consistent development in Mindweaver.
