# MindWeaver Development Guidelines

## Quick Reference

**Build/Run:** `task mw:build` | `task mw:dev` (with hot reload) | `task mind:serve`  
**Test:** `go test ./...` | Single test: `go test -run TestName ./path/to/package`  
**Database:** `task mw:db:reset` (reset all) | `task mind:db:migrations:up` | `task mind:db:store:generate` (regenerate sqlc)  
**Format:** Use `gofmt` (standard Go formatting) | Imports: stdlib → external → internal

## ⚠️ CRITICAL RULE: Always Use Task Commands

**NEVER run build commands directly (e.g., `go build ./cmd/mindweaver`)**  
**ALWAYS use the Task runner for build, run, and database operations**

- ✅ **Correct:** `task mw:build` (builds to `/bin/` directory, properly ignored)
- ❌ **Wrong:** `go build ./cmd/mindweaver` (creates binary in root, not ignored, gets committed)
- **Exception:** Testing with `go test` is allowed and encouraged
- **If a task doesn't exist:** STOP and ask before running manual commands

**Why this matters:**
- Task commands output binaries to `/bin/` which is gitignored
- Manual builds create binaries in root or current directory
- Root-level binaries are NOT ignored and will be committed accidentally
- This rule prevents polluting git history with compiled binaries

## Code Style

**Imports:** Group stdlib, external packages, then internal (`github.com/nkapatos/mindweaver/...`) with blank lines between  
**Naming:** CamelCase for exported, camelCase for unexported | Service structs end with `Service`, handlers with `Handler`, converters use `To*` prefix  
**Types:** Use sqlc-generated `store.*` types | API types in `*_v3.go` files | Prefer explicit types over `interface{}`  
**Errors:** Use `connect.NewError()` with proper codes in v3 APIs | Domain errors via helper functions (`newNotFoundError`, `newAlreadyExistsError`)  
**Logging:** Use `slog` with `middleware.GetRequestID(ctx)` for request traceability | Log at service layer only  
**Comments:** Export comments start with type/func name | Document complex logic inline | Keep `TODO:` comments for WIP features  
**Testing:** Use `testify/require` for assertions | Use `t.Helper()` for test utilities | Setup via `mindtesting.SetupTest(t)`

---

# Architecture Layer Guidelines

This document outlines the architectural principles and guidelines for implementing the different layers of the MindWeaver backend. These guidelines ensure consistency, maintainability, and clean separation of concerns across the codebase.

## Service Layer

The service layer is the core business logic layer of the application. It serves as the single source of truth for all domain operations and orchestrates data access.

### Responsibilities

- **Business Logic**: Contains all business rules, validations, and domain logic
- **Database Access**: The only layer that interacts directly with the database via `store.Querier`
- **Logging**: All logging is performed at this layer using `slog` with `middleware.GetRequestID` for traceability
- **CRUD Operations**: Implements standard Create, Read, Update, Delete operations that match the database schema and sqlc-generated methods
- **Error Handling**: Translates database errors into domain-specific errors (e.g., `ErrNotFound`, `ErrAlreadyExists`)

### Design Principles

- **Transport Agnostic**: No HTTP, API, JSON, or transport-specific logic. The service layer is reusable across any API type (REST, gRPC, GraphQL, CLI, etc.)
- **Single Responsibility**: Each service corresponds to a single domain entity and its operations
- **Minimal Interface**: Only CRUD and essential query methods are exposed. No legacy or convenience methods
- **Dependency Injection**: Services receive their dependencies (store, logger) via constructor injection

### Structure

```go
type ExampleService struct {
    store  store.Querier
    logger *slog.Logger
}

func NewExampleService(store store.Querier, logger *slog.Logger, serviceName string) *ExampleService {
    return &ExampleService{
        store:  store,
        logger: logger.With("service", serviceName),
    }
}
```

### Template Status

Service layer implementations serve as templates for future service implementations. When creating a new service, use existing service files as reference for structure, error handling, and logging patterns.

## Handler Layer

The handler layer is the presentation layer that exposes service functionality through HTTP endpoints. It handles all HTTP-specific concerns and API contract enforcement.

### Responsibilities

- **HTTP Protocol**: Manages all HTTP-specific logic (headers, status codes, content negotiation)
- **Request/Response**: Defines and validates API request and response structures
- **API Formatting**: Formats responses according to API specifications (REST, JSON:API, etc.)
- **Type Conversion**: Converts between API types and service/domain types using converter functions
- **Error Translation**: Translates service errors into appropriate HTTP error responses using `echo.NewHTTPError`

### Design Principles

- **No Business Logic**: Handlers delegate all business logic to the service layer
- **No Database Access**: Handlers never interact with the database directly
- **Thin Layer**: Handlers should be minimal wrappers around service calls
- **Consistent Error Handling**: Use middleware-provided error handling patterns

### Structure

```go
type ExampleHandler struct {
    exampleSvc *ExampleService
}

func NewExampleHandler(service *ExampleService) *ExampleHandler {
    return &ExampleHandler{
        exampleSvc: service,
    }
}
```

## Layer Interaction Flow

```
HTTP Request → Handler → Service → Store (Database)
                  ↓         ↓
              Converters  Logger
```

1. **Handler** receives HTTP request and extracts/validates parameters
2. **Converter** transforms API types to service types
3. **Service** executes business logic and database operations
4. **Service** logs operations with request context
5. **Converter** transforms service results to API types
6. **Handler** formats response and returns HTTP response

## Key Takeaways

- Service layer is transport-agnostic and contains all business logic
- Handler layer is thin and HTTP-specific
- Clear separation enables code reuse across different transport protocols
- Logging and error handling follow consistent patterns
- All layers reference sqlc-generated types and methods for database operations
