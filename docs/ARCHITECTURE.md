# Mindweaver Architecture

## System Overview

Mindweaver follows a clean, layered architecture pattern with clear separation of concerns. The application is built around the concept of managing LLM providers and their configurations.

## Architecture Layers

### 1. Presentation Layer
**Location**: `internal/handlers/`

- **Web Handlers** (`web/`): Server-side rendered pages using Templ
  - Home, Prompts, Providers, LLM Services, Settings, Conversations
- **API Handlers** (`api/`): REST API endpoints
  - Actor, Prompt, LLM (basic structure)

### 2. Business Logic Layer
**Location**: `internal/services/`

Services encapsulate business logic and coordinate between different components:

- **ProviderService**: Manages LLM provider CRUD operations and relationships
- **LLMService**: Handles LLM service configurations
- **PromptService**: Manages system prompts and templates
- **ActorService**: Handles user/actor management
- **ConversationService**: Manages conversation flows
- **MessageService**: Handles message storage and retrieval

### 3. Data Access Layer
**Location**: `internal/store/`

- **SQLC Generated Code**: Type-safe database queries and models
- **Database Interface**: Abstracted through the `Querier` interface
- **Migrations**: Version-controlled database schema changes

### 4. External Integration Layer
**Location**: `internal/adapters/`

- **Base Adapter Interface**: Common interface for all LLM providers
- **OpenAI Adapter**: Implementation for OpenAI and compatible services
- **Adapter Factory**: Creates appropriate adapters based on configuration

## Core Components

### Router & Middleware
**Location**: `internal/router/`

- **Echo Framework**: HTTP routing and middleware
- **Route Organization**: Separated by API and web routes
- **Error Handling**: Centralized error handling and 404 responses
- **Middleware Stack**: Logging and recovery middleware

### Configuration
**Location**: `config/`

- **Web Routes**: Route configuration and setup

## Data Flow

### 1. Web Request Flow
```
HTTP Request → Router → Middleware → Web Handler → Service → Store → Database
```

### 2. API Request Flow
```
HTTP Request → Router → Middleware → API Handler → Service → Store → Database
```

### 3. LLM Provider Flow
```
Provider Request → Service → Adapter → External LLM API → Response
```

## Database Design

### Core Entities

1. **llm_services**: External LLM service providers
2. **providers**: User-configured LLM provider instances
3. **prompts**: System prompts and templates
4. **actors**: Users or agents using the system
5. **conversations**: Multi-turn conversation sessions
6. **messages**: Individual messages within conversations

### Relationships
- Providers belong to LLM services
- Providers can have optional system prompts
- Conversations belong to actors
- Messages belong to conversations

## Adapter Pattern

### Base Interface
```go
type LLMProvider interface {
    Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error)
    GetName() string
}
```

### Supported Adapters
- **OpenAI**: Primary adapter for OpenAI and compatible services
- **OpenRouter**: Uses OpenAI adapter (compatible API)
- **Ollama**: Uses OpenAI adapter (compatible API)

## Service Layer Patterns

### Query + Relationship Pattern
Services follow a simple but effective pattern:
- Basic CRUD operations use direct SQL queries
- Relationship methods load related entities through separate queries
- Prioritizes maintainability and flexibility over raw performance

## Current Implementation Status

### Fully Implemented
- **Provider Management**: CRUD operations with relationships
- **LLM Service Management**: Basic service configuration
- **Prompt Management**: System prompt storage and retrieval
- **Web UI**: Basic pages for all major features
- **Database Layer**: Complete with SQLC-generated code

### Partially Implemented
- **API Endpoints**: Basic structure exists but limited functionality
- **Conversation System**: Database schema exists, UI placeholder
- **Message System**: Database schema exists, basic service layer

### Not Yet Implemented
- **Authentication**: No user authentication system
- **LLM Integration**: Adapters exist but not integrated with providers
- **Conversation UI**: Placeholder pages only
- **Settings**: Placeholder page only

## Development Workflow

### Build Process
1. **TypeScript/CSS**: esbuild bundles and watches assets
2. **Templates**: Templ generates Go code from .templ files
3. **Database**: SQLC generates type-safe database code
4. **Go Binary**: Standard Go build process

### Live Development
- **Air**: Hot reload for Go code
- **esbuild**: Watch mode for frontend assets
- **Templ**: Watch mode for template generation
- **Task**: Orchestrates all development processes

## Configuration Management

### Environment Variables
- `APP_PORT`: Application port (default: 8080)
- `PRODUCTION`: Production mode flag
- Database connection string (SQLite file)

### Build Configuration
- `Taskfile.yml`: Build and development tasks
- `sqlc.json`: Database code generation configuration
- `build.js`: Frontend build configuration

## Logging

- **Structured Logging**: slog for consistent log format
- **Log Levels**: Debug, Info, Error levels
- **Context Information**: Request context and error details
