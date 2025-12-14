# Configuration System

This package manages all configuration for Mindweaver services through environment variables.

## How It Works

Configuration loading order:
1. Built-in defaults
2. Environment variables override defaults
3. Command-line `--mode` flag sets deployment mode
4. `MODE` env var can override the flag

## Deployment Modes

Mindweaver supports two deployment architectures:

### Combined Mode (Default)
All services run in a single process on one port. Brain service uses a local adapter to communicate with Mind service directly in-process (zero network overhead).

**Use case**: Local development, simple deployments, desktop applications

```bash
# Runs on PORT (default: MIND_PORT=9421)
./mindweaver --mode=combined
```

### Standalone Mode
Services run as separate processes, each on their own port. Brain service uses HTTP adapter to communicate with Mind service over the network.

**Use case**: Distributed deployments, scaling individual services, microservices architecture

```bash
# Terminal 1 - Mind service on port 9421
./mindweaver --mode=mind

# Terminal 2 - Brain service on port 9422
./mindweaver --mode=brain
```

## Configuration Loading

1. **Command-line flag** sets initial mode (`--mode=combined|mind|brain`)
2. **MODE environment variable** can override the flag
3. **All other config** loaded from environment variables with built-in defaults
4. **No config file required** - works out of the box

## Build & Deployment Impact

### Binary Size
Single binary for all modes (~38MB). Mode selection happens at runtime, not compile time.

### Database Paths
- Defaults to relative paths: `db/mind.db`, `db/brain.db`
- For containers: Mount volumes to `/data` and set `MIND_DB_PATH=/data/mind.db`
- Services maintain **separate databases** - they never share or cross-reference

### Port Assignment
- **Combined mode**: Single port (defaults to 9421, override with `PORT`)
- **Standalone mode**: Each service has its own port (Mind: 9421, Brain: 9422)
- Clients connect to the same routes regardless of mode (transparent)

### Service Communication
- **Combined mode**: Local adapter (in-process, no network calls)
- **Standalone mode**: HTTP adapter (Brain calls Mind via `MIND_SERVICE_URL`)
  - Note: Only Brain needs Mind's URL (Brain queries notes, Mind doesn't query AI)

### LLM Dependencies
- Brain service requires an LLM endpoint (default: Ollama at `localhost:11434`)
- Mind service has no LLM dependencies
- Models specified via `LLM_SMALL_MODEL` / `LLM_BIG_MODEL` (not bundled with binary)

## Security Considerations

### ETag Salt
- Generated randomly per process by default
- **Problem**: ETags change on restart, breaking client caching
- **Solution**: Set `ETAG_SALT` env var for production (persistent across restarts)

```bash
# Generate once, store securely
ETAG_SALT=$(openssl rand -hex 32)
```

### Database Security
- SQLite databases stored on disk (not encrypted by default)
- File permissions control access
- For encryption: Use encrypted volumes or filesystem-level encryption

## CI/CD Integration

### Development
```bash
# Uses defaults, no env vars needed
task dev:init
go tool air
```

### Testing
```bash
# Point to test databases
MIND_DB_PATH=test/mind.db BRAIN_DB_PATH=test/brain.db go test ./...
```

### Production Build
```bash
# Single binary, mode selected at runtime
go build -o mindweaver ./cmd/mindweaver
```

### Container Deployment
```dockerfile
# Environment variables in docker-compose.yml or k8s ConfigMap
ENV MODE=combined
ENV MIND_DB_PATH=/data/mind.db
ENV BRAIN_DB_PATH=/data/brain.db
ENV LLM_ENDPOINT=http://ollama:11434
ENV ETAG_SALT=${ETAG_SALT}
```

## Configuration Reference

See `.env.example` in repository root for complete list of environment variables and their defaults.

## Architecture Notes

### Why Two Services?
- **Mind**: Pure PKM/notes functionality, no AI dependencies
- **Brain**: AI features, requires LLM infrastructure
- Separation allows deploying Mind without AI overhead if needed

### Why Two Databases?
- Architectural isolation - services never cross-reference data
- Communication happens via APIs (MindOperations interface)
- Allows independent scaling, backup strategies, and data retention policies

### Database Isolation (Critical)
Mind and Brain databases are **architecturally isolated**:
- No foreign keys between databases
- No shared tables or views
- Communication only through service APIs
- This isolation is intentional and must be maintained
