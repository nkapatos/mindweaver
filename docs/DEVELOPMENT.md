# Development Guide

Guide for developers who want to build, run, or contribute to Mindweaver.

## Prerequisites

- **[mise](https://mise.jdx.dev/)** - Manages toolchain versions (see `.mise.toml` for configured versions)
- **[Task](https://taskfile.dev/installation/)** - Task runner for build automation
- **[Buf](https://buf.build/docs/installation/)** - Protocol buffer tooling
- **OpenAI-compatible API** (optional) - Required for Brain service AI features. Any OpenAI API-compatible server works.

Component-specific tools (sqlc, goose, air) are documented in each component's README.

## Quick Start

### Using Task (Recommended)

The mindweaver server and clients each have their own Taskfiles. Navigate to the component directory and use `task --list` to see available tasks.

```bash
# Clone and navigate to server
git clone https://github.com/nkapatos/mindweaver.git
cd mindweaver/packages/mindweaver

# Show available tasks
task --list

# Application runs on http://localhost:9421
```

### Manual Build & Run

```bash
# From packages/mindweaver directory
cd packages/mindweaver

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

### Task Runner

Each component has its own Taskfile with build, test, and development tasks. Use `task --list` from the component directory to see available tasks.

### Running Tests

```bash
# From packages/mindweaver
go test ./...

# Run specific package tests
go test ./internal/mind/notes
go test ../pkg/config
```

### Code Generation

Code is generated from various sources (protocol buffers, SQL schemas). Use `task --list` in the component directory to see available generation tasks.

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

The monorepo contains a Go workspace in `packages/` (server and shared libraries), client implementations in `clients/`, and protocol definitions in `proto/`. Each component has its own build configuration and task definitions.

## API Design Guidelines

Mindweaver follows [Google's API Improvement Proposals (AIP)](https://google.aip.dev/) for REST API design. AIP provides comprehensive, battle-tested guidance for building consistent, scalable APIs.

### Core Principles

We adhere to these key AIPs:

- **[AIP-121: Resource-oriented design](https://google.aip.dev/121)** - APIs are organized around resources (notes, collections, assistants)
- **[AIP-122: Resource names](https://google.aip.dev/122)** - Consistent resource naming (`notes/123`, `collections/456`)
- **[AIP-131-135: Standard methods](https://google.aip.dev/131)** - GET, LIST, CREATE, UPDATE, DELETE
- **[AIP-158: Pagination](https://google.aip.dev/158)** - Cursor-based pagination with `page_token`
- **[AIP-160: Filtering](https://google.aip.dev/160)** - Structured filtering syntax
- **[AIP-193: Errors](https://google.aip.dev/193)** - Standardized error responses

### Response Wrappers

For improved developer experience, we wrap responses in a consistent envelope structure:

```json
{
  "data": { /* resource or list */ },
  "error": { /* error details if failed */ }
}
```

**Single resource** (GET, CREATE, UPDATE):
```json
{
  "data": {
    "id": 123,
    "title": "My Note",
    "content": "...",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Collection** (LIST):
```json
{
  "data": {
    "kind": "note#list",
    "items": [...],
    "next_page_token": "abc123"
  }
}
```

**DELETE** (AIP-135):
```
HTTP/1.1 204 No Content
(empty body)
```

**Error**:
```json
{
  "error": {
    "code": 400,
    "message": "Invalid note_id parameter",
    "status": "INVALID_ARGUMENT",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
        "reason": "INVALID_PARAMETER_FORMAT",
        "domain": "mind.mindweaver.com",
        "metadata": {
          "parameter": "note_id",
          "error": "strconv.ParseInt: parsing \"invalid\": invalid syntax"
        }
      }
    ]
  }
}
```

This wrapper provides:
- **Consistency** - All responses have the same top-level structure
- **Type safety** - Generic `Response[T]` types in Go
- **Error handling** - Unified error format across all endpoints
- **Partial success** - Can return both `data` and `error` for batch operations

### Standard Methods (AIP-131 to AIP-135)

Following AIP standard methods:

| Method | HTTP | Returns | Status Code |
|--------|------|---------|-------------|
| **List** (AIP-132) | GET /notes | `Response[ListResult[Note]]` | 200 |
| **Get** (AIP-131) | GET /notes/123 | `Response[Note]` | 200 |
| **Create** (AIP-133) | POST /notes | `Response[Note]` (full created resource) | 201 |
| **Update** (AIP-134) | PUT /notes/123 | `Response[Note]` (full updated resource) | 200 |
| **Delete** (AIP-135) | DELETE /notes/123 | Empty body | 204 |

### ETags and Versioning (AIP-154)

ETags are used for optimistic concurrency control to prevent conflicting updates:

**Flow:**
1. Client GETs resource: `GET /notes/123`
2. Server responds with ETag in **header**: `ETag: "v1-abc123"`
3. Client updates with condition: `PUT /notes/123` with `If-Match: "v1-abc123"`
4. Server validates:
   - If ETag matches → update succeeds, returns new ETag
   - If ETag mismatched → `412 Precondition Failed` (conflict detected)

**Important**: ETags are exchanged via HTTP headers, not in response bodies.

### Field Naming

We use **snake_case** for all JSON fields (following AIP-140):

```go
type Note struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    CreatedAt string `json:"created_at"`  // snake_case
    UpdatedAt string `json:"updated_at"`  // snake_case
}
```

### Implementation

Response wrappers are defined in `pkg/types/types.go`:

```go
// Single resource
Response[Note]

// Collection
Response[ListResult[Note]]

// Operation result (minimal response)
Response[OperationResult]
```

See `pkg/types/types.go` for complete type definitions and usage examples.

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

Run tests and linting from the component directory. Use `task --list` to see available tasks for testing, linting, and formatting.

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

- [Configuration System](../packages/pkg/config/README.md) - Detailed configuration docs
- [Main README](../README.md) - Project overview
