# Mindweaver

> *Where your thoughts find structure, and AI brings clarity.*

**Mindweaver** is a personal knowledge management system that helps you **do something** with your knowledge—not just collect it. Stop endlessly organizing your vault and start having conversations with your notes, finding unexpected connections, and turning scattered thoughts into actionable insights.

## Why Mindweaver?

### For Everyone Who Takes Notes

**The Problem:** You collect notes, articles, ideas, and highlights. You spend hours organizing folders and tags. But when you actually need that insight? It's buried somewhere, and you can't remember where or how it connects to what you're working on now.

**The Mindweaver Difference:**
- **Talk to your knowledge**: Ask questions in natural language, get answers from across all your notes
- **Discover connections**: AI finds relationships between ideas you never explicitly linked
- **Stop organizing, start creating**: Let AI handle structure while you focus on thinking
- **Privacy-first**: Your notes live in local databases you control—no cloud, no subscriptions, no prying eyes
- **Works offline**: Full-text search, collections, and AI assistance without an internet connection

### For Developers (Neovim Plugin)

If you live in your editor and think in code, Mindweaver meets you where you work:

- **Project-aware notes**: Automatically links notes to the codebase you're working on
- **Markdown-native**: Full wikilink support, frontmatter, and seamless integration with your workflow  
- **Never leave your editor**: Native Neovim plugin captures context without breaking flow
- **Code-aware AI**: Ask questions that understand both your notes and your project structure
- **Built for terminal workflows**: Fast, keyboard-driven, zero friction

### For Knowledge Workers

Whether you're a researcher, writer, student, or lifelong learner:

- **Cross-reference automatically**: Wikilinks and backlinks show how ideas connect
- **Collections, not folders**: Group notes by topic without rigid hierarchies
- **Tag intelligently**: Multiple dimensions of organization that actually help you find things
- **Export and integrate**: Import/export tools (imex) work with your existing markdown notes
- **Markdown everywhere**: Plain text files you can read in any editor, forever

## Core Features at a Glance

- **Intelligent search**: Full-text search with AI-powered semantic understanding
- **Wikilinks & backlinks**: Obsidian-style linking with automatic resolution
- **Collections & tags**: Multiple dimensions of organization without rigid hierarchies
- **Conversation with your notes**: Ask questions, get answers from across your knowledge base
- **Project-aware**: Automatically links notes to codebases and work contexts (Neovim)
- **Background analysis**: AI indexes and understands your notes as you write them
- **Import/export**: Move your existing markdown vaults in and out seamlessly
- **Templates**: Consistent note structure without manual formatting
- **Offline-capable**: Full functionality without internet connection
- **Local LLMs**: Privacy-first AI using your choice of local models

## What Makes It Different?

Traditional PKM tools make you choose: simple and limited, or powerful but overwhelming. Mindweaver gives you both:

1. **Action over organization**: AI helps you retrieve and synthesize, not just file away
2. **Proactive AI assistance**: Works in the background, surfaces relevant notes as you work—no manual searching required
3. **Local-first intelligence**: Runs on efficient local LLMs for privacy and speed
4. **Context-aware**: Understands projects, timestamps, and relationships between your ideas
5. **Your data, your rules**: SQLite databases on your machine, not locked in someone's cloud
6. **Fast and lightweight**: Designed for efficiency—blazing full-text search, minimal resource usage
7. **Built for interoperability**: Desktop app (coming), web interface (coming), and Neovim plugin (available now)

## Technical Philosophy

Mindweaver is built on principles that respect your time, privacy, and hardware:

### Privacy-First Architecture
- **Local LLMs by default**: Uses small models for fast routing and larger models for deeper analysis
- **Your data never leaves**: Everything runs on your machine—notes, embeddings, conversations
- **No tracking, no telemetry**: What you write stays with you

### Efficient by Design
- **Small and fast**: Optimized for speed with tiered query execution
- **Smart background processing**: AI analyzes and indexes notes without interrupting your workflow
- **Minimal resource footprint**: Runs comfortably on modest hardware with local models

### Intelligent Without Being Intrusive
- **Proactive context retrieval**: Brain service autonomously searches your notes when relevant
- **Multi-tier reasoning**: Fast queries skip AI entirely, complex questions get full LLM treatment
- **Tool-based architecture**: AI decides when to search, summarize, or synthesize—not you

## How It Works

Mindweaver consists of two independent services that work together:

### Mind Service (PKM Core)
Your personal knowledge layer—markdown notes, wikilinks, collections, and tags. Handles:
- Note creation, editing, and full-text search
- Automatic link resolution (Obsidian-style wikilinks)
- Collections for organization, templates for consistency
- Fast SQLite-based storage with FTS5 search

### Brain Service (AI Assistant)
Your intelligent companion that works proactively in the background:
- **Autonomous context retrieval**: Searches your notes automatically when questions arise
- **Tiered reasoning system**: 
  - Tier 1: Fast SQL queries for simple lookups
  - Tier 2: Small model routing for intent classification
  - Tier 3: Full LLM analysis for complex synthesis
- **Background ingestion**: Analyzes and indexes notes without blocking your work
- **Conversational memory**: Maintains context across interactions

The two services communicate seamlessly—Brain autonomously queries Mind when it needs knowledge, and Mind notifies Brain when notes change. All running locally, all under your control.

## Getting Started

### For End Users

**Desktop Application (Coming Soon)**

Mindweaver will be available as a native desktop application for macOS, Windows, and Linux. You'll be able to:

- Download the installer for your platform
- Install and run without any technical setup
- Start creating notes immediately with a clean, intuitive interface
- Access all features through the desktop app—no terminal required

**Web Interface (Coming Soon)**

A self-hosted web interface will let you access Mindweaver from any browser while keeping your data local.

**Current Status:** Mindweaver is currently in alpha (v1.0.0-alpha) and available for developers via the Neovim plugin. Desktop and web applications are in active development.

**Want to try it now?** If you're comfortable with developer tools, see the "For Developers" section below.

### For Developers

**Full documentation for developers is available in [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).**

Quick start for developers who want to contribute or use the Neovim plugin:

**Prerequisites:**
- Go 1.24+ (for project-local tool installation)
- Task runner - [Installation guide](https://taskfile.dev/installation/)
- Ollama (optional, for local LLM) - [ollama.com](https://ollama.com)

**Setup:**

```bash
# Clone and setup
git clone https://github.com/nkapatos/mindweaver.git
cd mindweaver
task dev:init

# Start development server (Mind + Brain combined)
go tool air
```

**What's available now:**
- Neovim plugin for terminal-native note-taking
- Full API access for building custom clients
- Local-first architecture ready for desktop/web UIs

**Documentation:**
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) - Complete developer guide with examples
- [pkg/config/README.md](pkg/config/README.md) - Configuration and deployment reference
- [clients/nvim/](clients/nvim/) - Neovim plugin installation and usage

## Repository Structure

This repository uses a flat Go layout with isolated clients:

```
mindweaver/
├── cmd/          # Go binaries (mindweaver, lsp, imex)
├── internal/     # Private packages (mind, brain, imex)
├── pkg/          # Public packages (config, transport, logging)
├── sql/          # Database schemas (ISOLATED: mind.db + brain.db)
├── clients/      # Client implementations (nvim, web, desktop, browser)
├── docker/       # Container configurations
├── docs/         # Documentation (DEVELOPMENT.md, etc.)
└── tasks/        # Task runner configs
```

## Author

Nick Kapatos (@nkapatos)
