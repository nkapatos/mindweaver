# Mindweaver Server

The mindweaver server provides the Mind and Brain services.

## Prerequisites

- Go 1.25+
- [Task](https://taskfile.dev/installation/)
- [Buf](https://buf.build/docs/installation/)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [goose](https://github.com/pressly/goose)
- [air](https://github.com/air-verse/air) (optional, for hot reload)

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
- Protocol buffer definitions change (`proto/mind/v3/*.proto`)
- SQL queries change (`store/*/sql/*.sql`)

Use `task --list` to see available generation tasks for each service.

### Database Migrations

Each service has its own database and migration tasks. See `task --list` for available migration commands.
