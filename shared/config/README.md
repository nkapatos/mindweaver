# Configuration System

This package manages all configuration for Mindweaver services using [Viper](https://github.com/spf13/viper).

## Configuration Precedence

Configuration is loaded in the following order (highest priority first):

1. **Environment variables** (`MW_*` prefix)
2. **Config file** (`config.yaml`)
3. **Built-in defaults**

## Quick Start

### For Developers

No configuration needed! The defaults work out of the box:

```bash
# Just run it
task mw:dev

# Or customize via .env.local
cp .env.example .env.local
# Edit .env.local as needed
```

### For Docker Deployments

```bash
docker run -d \
  -p 9421:9421 \
  -v mindweaver-data:/data \
  -e MW_DATA_DIR=/data \
  codefupandas/mindweaver:latest
```

Or use docker-compose (see `docker-compose.yml` in repo root).

## Environment Variables Reference

All environment variables use the `MW_` prefix. Nested config uses underscores.

| Variable | Default | Description |
|----------|---------|-------------|
| `MW_DATA_DIR` | `./data` | Root directory for all data |
| `MW_MODE` | - | Override deployment mode (combined/standalone) |
| `MW_PORT` | 9421 | Port override for combined mode |
| `MW_MIND_PORT` | 9421 | Mind service port |
| `MW_MIND_DB_PATH` | `$DATA_DIR/mind.db` | Mind SQLite database |
| `MW_BRAIN_PORT` | 9422 | Brain service port |
| `MW_BRAIN_DB_PATH` | `$DATA_DIR/brain.db` | Brain SQLite database |
| `MW_BRAIN_BADGER_DB_PATH` | `$DATA_DIR/badger/` | BadgerDB for title index |
| `MW_BRAIN_MIND_SERVICE_URL` | `http://localhost:9421` | Mind URL (standalone mode) |
| `MW_BRAIN_LLM_ENDPOINT` | `http://localhost:11434` | Ollama/LLM endpoint |
| `MW_BRAIN_SMALL_MODEL` | `phi3-mini` | Fast model for routing |
| `MW_BRAIN_BIG_MODEL` | `phi4` | Powerful model for reasoning |
| `MW_LOG_LEVEL` | `INFO` | DEBUG, INFO, WARN, ERROR |
| `MW_LOG_FORMAT` | `text` | text or json |
| `MW_SECURITY_ETAG_SALT` | (random) | ETag hashing salt |

## Data Directory Structure

All persistent data lives under `MW_DATA_DIR`:

```
$MW_DATA_DIR/
├── mind.db        # Mind SQLite database
├── brain.db       # Brain SQLite database
├── badger/        # BadgerDB for title index
└── config.yaml    # Optional config file (future)
```

### Path Derivation

If individual DB paths are not set, they are derived from `MW_DATA_DIR`:

- `MW_MIND_DB_PATH` → `$MW_DATA_DIR/mind.db`
- `MW_BRAIN_DB_PATH` → `$MW_DATA_DIR/brain.db`
- `MW_BRAIN_BADGER_DB_PATH` → `$MW_DATA_DIR/badger/`

You can override any path individually while still using the data directory for others.

## Deployment Modes

### Combined Mode (Default)

All services run in a single process on one port:

```bash
./mindweaver --mode=combined
# Or just: ./mindweaver
```

### Standalone Mode

Services run as separate processes:

```bash
# Terminal 1 - Mind service
./mindweaver --mode=mind

# Terminal 2 - Brain service
MW_BRAIN_MIND_SERVICE_URL=http://localhost:9421 ./mindweaver --mode=brain
```

## Config File (Future)

In future releases, Mindweaver will support a `config.yaml` file in the data directory for end-user configuration. The file will be created by a setup wizard on first run.

Config file search paths:
1. `/data/config.yaml` (Docker)
2. `$HOME/.config/mindweaver/config.yaml` (Linux)
3. `$HOME/Library/Application Support/Mindweaver/config.yaml` (macOS)
4. `./config.yaml` (current directory)

## Validation

On startup, Mindweaver validates:

1. Data directory exists or can be created
2. Database parent directories are writable
3. BadgerDB directory exists or can be created

If validation fails, the application exits with a clear error message.

## Security

### ETag Salt

ETags are hashed for security. By default, a random salt is generated per process, which means ETags change on restart.

For production, set a persistent salt:

```bash
# Generate once
MW_SECURITY_ETAG_SALT=$(openssl rand -hex 32)
```

### Database Security

- SQLite databases are stored on disk (not encrypted by default)
- File permissions control access
- For encryption, use encrypted volumes or filesystem-level encryption

## Architecture Notes

### Why Two Services?

- **Mind**: Pure PKM/notes functionality, no AI dependencies
- **Brain**: AI features, requires LLM infrastructure

This separation allows deploying Mind without AI overhead if needed.

### Why Two Databases?

- Architectural isolation - services never cross-reference data
- Communication happens via APIs (MindOperations interface)
- Allows independent scaling, backup strategies, and data retention policies
