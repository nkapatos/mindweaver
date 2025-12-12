# Changelog

All notable changes to Mindweaver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0] - 2025-12-12

### Mind Service

#### Features
- Initial release of Mind service with complete PKM functionality
- Notes management with markdown support and wikilinks
- Collections for organizing notes hierarchically
- Tags extracted from frontmatter and inline hashtags
- Full-text search with FTS5
- Note templates for consistency
- Note types for categorization
- Links and backlinks tracking
- Note metadata (frontmatter)

#### Architecture
- V3 API following Google AIP standards
- Connect-RPC protocol with REST compatibility
- Service layer with business logic
- SQLite with FTS5 for fast search
- Atomic transactions for data integrity
- Comprehensive error handling

#### Technical Implementation
- Go 1.25.5 with modern tooling
- Task runner for development workflow
- Air for hot reload in development
- SQLC for type-safe database queries
- Protocol Buffers for API definitions
- Bruno API tests for integration testing

### Breaking Changes
- **BREAKING**: Initial release establishes V3 API structure

[Unreleased]: https://github.com/nkapatos/mindweaver/compare/mindweaver/v0.9.0...HEAD
[0.9.0]: https://github.com/nkapatos/mindweaver/releases/tag/mindweaver/v0.9.0
