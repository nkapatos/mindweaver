# Mindweaver - Your AI Knowledge Ally

## Overview

Mindweaver is a knowledge-aware AI system designed to be your indispensable ally in learning, creating, and connecting ideas. Unlike traditional AI tools that expect you to be a perfect prompt engineer, Mindweaver recognizes that AI is most powerful when it understands your context, knowledge, and goals.

### The Vision

Most of us don't use AI to its full potential because we expect it to "read our minds" and figure out what we want. Mindweaver takes a different approach - it's designed to help you:

- **Iterate through ideas** and explore concepts more effectively
- **Revisit work and notes** with AI that remembers your context
- **Connect the dots** between scattered information from various sources
- **Build knowledge-aware AI tools** that understand your projects and goals
- **Accelerate learning and creation** by having an AI ally that grows with you

### How It Works

Mindweaver allows you to create projects and add knowledge from various sources (web articles, notes, code, etc.). The AI then becomes aware of your context and can help you:

- **Understand your codebase** and suggest improvements based on your knowledge
- **Iterate through findings** from research and reading
- **Connect new information** to your existing projects and goals
- **Maintain context** across different work sessions and sources

### Example Scenario

Imagine you're working on a Go project (like this one!) and reading about best practices online. With Mindweaver, you could:

1. Send articles you're reading to your project
2. Iterate through findings with AI that understands your codebase
3. Get suggestions that are contextual to your specific project
4. Have AI that remembers your architecture decisions and can explain why certain choices make sense

## Key Concepts

### Provider
A provider is a user-defined entity that combines:
- An LLM service (e.g., OpenAI)
- A specific model from that service
- Configuration settings (temperature, max tokens, etc.)
- Optional system prompts

### LLM Service
External LLM service providers (OpenAI, etc.) that offer models and APIs. The application provides adapters to bridge these external services.

### Adapters
Go packages that bridge external LLM services with the application. Currently supports:
- OpenAI adapter (for OpenAI and OpenAI-compatible services)

### Knowledge Context
The core feature that makes Mindweaver unique - the ability to maintain context about your projects, research, and goals so your AI tools can be truly helpful.

## Tech Stack

### Backend
- **Language**: Go 1.24.2
- **Framework**: Echo v4 (HTTP server)
- **Database**: SQLite with SQLC for type-safe queries
- **Templating**: Templ (Go-based templating)
- **Logging**: Structured logging with slog

### Frontend
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4 + DaisyUI
- **Bundling**: esbuild
- **Icons**: Lucide Static

### Development Tools
- **Task Runner**: Task (build automation)
- **Live Reload**: Air (Go) + esbuild watch (TypeScript)
- **Database**: SQLC for code generation
- **Package Manager**: pnpm

## Quick Start

### Prerequisites
- Go 1.24.2+
- Node.js 18+
- pnpm

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd mindweaver

# Install Go dependencies
go mod download

# Install Node.js dependencies
pnpm install

# Generate SQLC code
sqlc generate

# Run the application
task live
```

The application will be available at:
- Main app: http://localhost:8080
- Development proxy: http://localhost:8081

## Development Commands

### Live Development
```bash
task live          # Start live reload for Go, templates, and assets
task live:templ    # Live reload templates only
task live:esbuild  # Live reload TypeScript/CSS only
```

### Building
```bash
task build         # Build everything for production
task build:go      # Build Go binary only
task build:templ   # Build templates only
task build:esbuild # Build TypeScript/CSS only
```

### Code Quality
```bash
task format:templ  # Format all templ files
```

## Current Features

### Foundation (Implemented)
- **Provider Management**: Create and configure LLM providers for different AI services
- **LLM Service Management**: Manage connections to various LLM services
- **Prompt Management**: Store and manage system prompts for consistent AI behavior
- **Web Interface**: Modern UI built with Templ and Tailwind

### Coming Soon
- **Project Management**: Create projects to organize your work and knowledge
- **Knowledge Ingestion**: Add information from web articles, notes, and other sources
- **Context-Aware AI**: AI that understands your projects and can provide contextual help
- **Knowledge Connections**: Discover relationships between different pieces of information
- **Iterative Workflows**: Work with AI to explore ideas and connect insights

## Project Structure

```
mindweaver/
├── cmd/mindweaver/          # Application entry point
├── internal/                # Private application code
│   ├── adapters/           # LLM service adapters
│   ├── config/             # Configuration
│   ├── handlers/           # HTTP handlers (API + Web)
│   ├── router/             # Routing and middleware
│   ├── services/           # Business logic
│   ├── store/              # Database layer (SQLC)
│   └── templates/          # Templ templates
├── assets/                 # Frontend assets (TS, CSS)
├── migrations/             # Database migrations
├── sql/                    # SQLC schema and queries
├── docs/                   # Documentation
└── test/                   # Tests
```

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture information.

## Contributing

1. Follow Go coding standards
2. Use structured logging with slog
3. Write tests for new functionality
4. Update documentation as needed
5. Use Task commands for development workflows

## License

See [LICENSE](LICENSE) file for details.