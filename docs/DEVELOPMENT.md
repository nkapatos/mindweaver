# Development Guide

Guide for developers who want to build, run, or contribute to Mindweaver.

## Prerequisites

- **Go 1.24+** (for project-local tool installation)
- **Task** - [Installation guide](https://taskfile.dev/installation/)
- **Ollama** (optional, for AI features) - [ollama.com](https://ollama.com)

## Quick Start

### Using Task (Recommended)

```bash
# Clone the repository
git clone https://github.com/nkapatos/mindweaver.git
cd mindweaver

# Install development tools (air for hot reload, sqlc for code generation)
task dev:init

# Start development server with hot reload
task dev

# Application runs on http://localhost:9421
```

### Manual Build & Run

```bash
# Build the binary
go build -o mindweaver ./cmd/mindweaver

# Run with defaults (combined mode, port 9421)
./mindweaver

# Run with specific mode
./mindweaver --mode=combined  # Both services (default)
./mindweaver --mode=mind      # Notes/PKM only
./mindweaver --mode=brain     # AI assistant only
```

## Runtime Configuration

### Command-Line Flags

```bash
# Set runtime mode
./mindweaver --mode=combined   # Default
./mindweaver --mode=mind
./mindweaver --mode=brain
```

### Environment Variables

All configuration uses environment variables with sensible defaults:

```bash
# Override mode via environment
MODE=mind ./mindweaver

# Custom ports
PORT=8080 ./mindweaver --mode=combined
MIND_PORT=8080 BRAIN_PORT=8081 ./mindweaver --mode=mind

# Custom database paths
MIND_DB_PATH=/data/notes.db BRAIN_DB_PATH=/data/ai.db ./mindweaver

# LLM configuration
LLM_ENDPOINT=http://localhost:11434 ./mindweaver
LLM_SMALL_MODEL=phi3-mini LLM_BIG_MODEL=phi4 ./mindweaver

# Logging
LOG_LEVEL=DEBUG LOG_FORMAT=json ./mindweaver
```

### Using .env.local

For persistent local configuration:

```bash
# Copy example file
cp .env.example .env.local

# Edit with your preferences
# .env.local is gitignored and won't be committed
```

See `.env.example` for all available options.

## Development Workflow

### Hot Reload Development

```bash
# Start with hot reload (watches for file changes)
task dev

# Or manually with air
go tool air
```

### Running Tests

```bash
# Run all tests
task test

# Run specific package tests
go test ./internal/mind/notes
go test ./pkg/config
```

### Code Generation

```bash
# Generate SQL query code (after modifying .sql files)
task generate

# Or manually
go tool sqlc generate
```

## Running in Different Modes

### Combined Mode (Default)

Both Mind and Brain services in one process:

```bash
./mindweaver --mode=combined
# Runs on http://localhost:9421
# Brain uses in-process adapter (no network calls)
```

### Standalone Mode

Run services separately (useful for distributed deployments):

```bash
# Terminal 1: Start Mind service
./mindweaver --mode=mind
# Runs on http://localhost:9421

# Terminal 2: Start Brain service
MIND_SERVICE_URL=http://localhost:9421 ./mindweaver --mode=brain
# Runs on http://localhost:9422
# Brain uses HTTP adapter to talk to Mind
```

### Docker

```bash
# Build image
docker build -t mindweaver:dev -f docker/Dockerfile .

# Run combined mode
docker run -p 9421:9421 \
  -v $(pwd)/data:/data \
  -e MIND_DB_PATH=/data/mind.db \
  -e BRAIN_DB_PATH=/data/brain.db \
  mindweaver:dev

# Run with environment file
docker run -p 9421:9421 \
  --env-file .env.local \
  -v $(pwd)/data:/data \
  mindweaver:dev --mode=combined
```

## Project Structure

```
mindweaver/
├── cmd/mindweaver/    # Main entry point
├── internal/          # Private packages
│   ├── mind/         # Notes/PKM service
│   ├── brain/        # AI assistant service
│   └── imex/         # Import/export functionality
├── pkg/              # Public packages
│   ├── config/       # Configuration management
│   ├── logging/      # Structured logging
│   └── ...
├── sql/              # Database schemas
│   ├── mind/         # Notes database
│   └── brain/        # AI database
├── docs/             # Documentation
└── tasks/            # Task runner configs
```

## Configuration Deep Dive

For detailed information about how configuration affects deployment, see [pkg/config/README.md](../pkg/config/README.md).

Key points:
- **No config file required** - sensible defaults work out of the box
- **Single binary** - mode selected at runtime with `--mode` flag
- **Two databases** - architecturally isolated (mind.db + brain.db)
- **Environment variables** - override any default setting

## Troubleshooting

### Port already in use

```bash
# Change the port
PORT=8080 ./mindweaver

# Or for standalone mode
MIND_PORT=8080 ./mindweaver --mode=mind
```

### Database locked

```bash
# Make sure no other instance is running
pkill mindweaver

# Or use different database paths
MIND_DB_PATH=./dev-mind.db BRAIN_DB_PATH=./dev-brain.db ./mindweaver
```

### LLM connection failed

```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama if needed
ollama serve

# Or point to different endpoint
LLM_ENDPOINT=http://ollama.example.com:11434 ./mindweaver
```

## Contributing

### Before Submitting PR

```bash
# Run tests
task test

# Run linting
task lint

# Format code
task fmt
```

### Code Style

- Follow standard Go conventions
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

### Testing

- Write tests for new features
- Ensure existing tests pass
- Test both combined and standalone modes if applicable

## See Also

- [Configuration System](../pkg/config/README.md) - Detailed configuration docs
- [Main README](../README.md) - Project overview
- `.env.example` - All available configuration options
