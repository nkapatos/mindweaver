# Mindweaver

[![CI](https://github.com/nkapatos/mindweaver/actions/workflows/ci.yml/badge.svg)](https://github.com/nkapatos/mindweaver/actions/workflows/ci.yml)
[![Release](https://github.com/nkapatos/mindweaver/actions/workflows/release-mindweaver.yml/badge.svg)](https://github.com/nkapatos/mindweaver/actions/workflows/release-mindweaver.yml)
[![Latest Release](https://img.shields.io/github/v/release/nkapatos/mindweaver?include_prereleases)](https://github.com/nkapatos/mindweaver/releases)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

> *Where your thoughts find structure, and AI brings clarity.*

**Mindweaver** is a personal knowledge management system that helps you **do something** with your knowledge—not just collect it. Stop endlessly organizing your vault and start having conversations with your notes, finding unexpected connections, and turning scattered thoughts into actionable insights.

## Why Mindweaver?

You collect notes, articles, and ideas. You organize folders and tags. But when you need that insight, it's buried somewhere you can't remember.

Mindweaver changes this:

- **Talk to your knowledge** - Ask questions in natural language, get answers from across all your notes
- **Discover connections** - AI finds relationships between ideas you never explicitly linked
- **Stop organizing, start creating** - Let AI handle structure while you focus on thinking
- **Privacy-first** - Your notes live in local databases you control
- **Works offline** - Full-text search, collections, and AI assistance without internet

## Core Features

- **Intelligent search** - Full-text search with AI-powered semantic understanding
- **Wikilinks & backlinks** - Obsidian-style linking with automatic resolution
- **Collections & tags** - Flexible organization without rigid hierarchies
- **Conversation with notes** - Ask questions, get answers from your knowledge base
- **Background analysis** - AI indexes and understands your notes as you write
- **Templates** - Consistent note structure without manual formatting
- **Local LLMs** - Privacy-first AI using your choice of local models

## Architecture

Mindweaver consists of two independent services:

**Mind Service** - Your personal knowledge layer: markdown notes, wikilinks, collections, tags, and fast SQLite-based full-text search.

**Brain Service** - Your intelligent companion: autonomous context retrieval, tiered reasoning (fast SQL queries → small model routing → full LLM analysis), and conversational memory.

The services communicate seamlessly—Brain queries Mind when it needs knowledge, Mind notifies Brain when notes change. All local, all under your control.

## Current Status

Mindweaver is in active development. Here's what exists today and what's coming:

### Available Now

- **Mind Service** - Full PKM backend with notes, collections, tags, templates, wikilinks, and FTS5 search
- **Brain Service** - SQL schemas and store layer (AI integration in progress)
- **Connect RPC API** - Type-safe gRPC/HTTP API for building clients

### Coming Next

- Neovim plugin for terminal-native note-taking
- Desktop application (macOS, Windows, Linux)
- Web interface (self-hosted)
- Import/export tools for existing markdown vaults

## Getting Started

**Prerequisites:**

- **[mise](https://mise.jdx.dev/)** - A polyglot tool version manager that replaces tools like asdf, nvm, pyenv, and rbenv. Manages the toolchain versions specified in `.mise.toml`. → [Installation & docs](https://mise.jdx.dev/getting-started.html)

- **[Task](https://taskfile.dev/)** - A fast, cross-platform task runner and build tool. A modern alternative to Make with support for dependencies, variables, and platform-specific commands. → [Installation & docs](https://taskfile.dev/installation/)

- **[Buf](https://buf.build/)** - Your one-stop shop for local Protobuf development. Handles compilation, linting, breaking change detection, and code generation for protocol buffers. → [Installation & docs](https://buf.build/docs/installation)

- **OpenAI-compatible API** (optional) - Required for Brain service AI features. Any OpenAI API-compatible server (local or remote) works. → Examples: [Ollama](https://ollama.com/), [LM Studio](https://lmstudio.ai/), [vLLM](https://docs.vllm.ai/), or OpenAI directly

**Documentation:**
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) - Developer guide and setup
- [docs/WORKFLOW.md](docs/WORKFLOW.md) - Contribution workflow and PR guidelines
- Component-specific docs in each package directory

## Author

Nick Kapatos (@nkapatos)
