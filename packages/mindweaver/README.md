# Mindweaver Server

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)

The mindweaver server provides the Mind and Brain services for personal knowledge management with AI capabilities.

## Prerequisites

- **Go 1.25.5** - See root `.mise.toml` or install manually
- **[sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)** - SQL code generator
- **[goose](https://github.com/pressly/goose)** - Database migration tool
- **[air](https://github.com/air-verse/air)** (optional) - Hot reload for development

Task definitions include precondition checks with installation instructions.

## Quick Start

```bash
# Show available tasks
task --list

# Build (auto-generates protos)
task mw:build

# Run in dev mode with hot reload
task mw:dev
```

## Development

### Code Generation

Generated code is not committed. Run generation tasks when:
- Protocol buffer definitions change (`../../proto/mind/v3/*.proto`)
- SQL queries change (`store/*/sql/*.sql`)

Use `task --list` to see available generation tasks for each service.

### Database Migrations

Each service (Mind and Brain) has its own SQLite database and migration tasks:

```bash
# Migrate both databases
task mw:db:migrations:up

# Reset migrations
task mw:db:migrations:reset

# Full reset (migrations + regenerate store code)
task mw:db:reset
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/mind/notes
go test ./internal/brain/store
```

## Configuration

The server uses environment variables with sensible defaults. See `.env.example` for all available options.

```bash
# Copy example configuration
cp .env.example .env.local

# Edit with your preferences
# .env.local is gitignored
```

Key configuration:
- `MODE` - Runtime mode: `combined` (default), `mind`, or `brain`
- `PORT` - Server port (default: 9421)
- `MIND_DB_PATH` - Mind database path
- `BRAIN_DB_PATH` - Brain database path
- `LLM_ENDPOINT` - OpenAI-compatible API endpoint for Brain service

## Architecture

The server is a single Go binary containing:
- **Mind Service** - Markdown notes, wikilinks, collections, tags, FTS5 search
- **Brain Service** - AI assistant with context retrieval and conversational memory
- **Connect RPC API** - gRPC/HTTP API for clients

See [docs/architecture.md](docs/architecture.md) for detailed architecture patterns.

## See Also

- [Root README](../../README.md) - Project overview
- [docs/WORKFLOW.md](../../docs/WORKFLOW.md) - Contribution guidelines
- [docs/api.md](docs/api.md) - API design patterns
