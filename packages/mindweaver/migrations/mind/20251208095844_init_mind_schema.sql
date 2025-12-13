-- +goose Up
-- +goose StatementBegin

-- ========================================
-- Mind Service Schema - Consolidated Init
-- ========================================
-- This migration consolidates all mind service schema requirements
-- from incremental development into a clean, production-ready schema.

-- ========================================
-- 1. Note Types: Semantic Identity & Classification
-- ========================================
CREATE TABLE note_types (
id INTEGER PRIMARY KEY AUTOINCREMENT,
-- 'book', 'person', 'project', 'default', 'quicknote'
type TEXT NOT NULL UNIQUE,
-- 'Book', 'Person', 'Project', 'Default', 'Quick Note'
name TEXT NOT NULL,
description TEXT,
icon TEXT,
color TEXT,
is_system BOOLEAN DEFAULT 0 NOT NULL,  -- Protects system types from deletion
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ;

-- ========================================
-- 2. Collections: Hierarchical Organization (Multi-root Forest)
-- ========================================
CREATE TABLE collections (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
parent_id INTEGER NULL,          -- NULL = top-level collection
path TEXT NOT NULL UNIQUE,  -- Computed path for lookups (e.g., 'work/projects')
description TEXT,
position INTEGER DEFAULT 0,     -- For manual ordering within same parent
-- Protects system collections from deletion
is_system BOOLEAN DEFAULT 0 NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

FOREIGN KEY (parent_id) REFERENCES collections (id) ON DELETE CASCADE,

-- Ensure no duplicate names at same level
UNIQUE (parent_id, name)
) ;

-- ========================================
-- 3. Notes: Universal Content Unit
-- ========================================
CREATE TABLE notes (
id INTEGER PRIMARY KEY AUTOINCREMENT,
uuid TEXT NOT NULL UNIQUE,  -- UUIDv7 for sync and time-ordering
title TEXT NOT NULL,
-- Nullable: allows title-only notes (quick capture)
body TEXT,
description TEXT,
frontmatter TEXT,                   -- Parsed YAML frontmatter stored as STRING
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
version INTEGER DEFAULT 1 NOT NULL,  -- For optimistic locking / version history
note_type_id INTEGER,
collection_id INTEGER NOT NULL DEFAULT 1,  -- Every note belongs to a collection
is_template BOOLEAN DEFAULT 0,     -- Marks notes used as templates

FOREIGN KEY (note_type_id) REFERENCES note_types (id) ON DELETE SET NULL,
FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE SET DEFAULT,

-- Title must be unique within a collection
UNIQUE (collection_id, title)
) ;

-- ========================================
-- 4. Templates: Pointers to Starter Notes
-- ========================================
CREATE TABLE templates (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,  -- 'Book Review', 'Daily Log'
description TEXT,
starter_note_id INTEGER NOT NULL UNIQUE,  -- Points to a note with is_template=1
note_type_id INTEGER,                  -- Optional type association
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

FOREIGN KEY (starter_note_id) REFERENCES notes (id) ON DELETE RESTRICT,
FOREIGN KEY (note_type_id) REFERENCES note_types (id) ON DELETE SET NULL
) ;

-- ========================================
-- 5. Note Meta: EAV for Arbitrary Key-Value Pairs
-- ========================================
CREATE TABLE note_meta (
id INTEGER PRIMARY KEY AUTOINCREMENT,
note_id INTEGER NOT NULL,
key TEXT NOT NULL,  -- 'author', 'isbn', 'status', 'rating'
value TEXT,           -- Stringly-typed for flexibility
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

FOREIGN KEY (note_id) REFERENCES notes (id) ON DELETE CASCADE,

-- One value per key per note
UNIQUE (note_id, key)
) ;

-- ========================================
-- 6. Tags: Cross-cutting Filtering
-- ========================================
CREATE TABLE tags (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ;

CREATE TABLE note_tags (
note_id INTEGER NOT NULL,
tag_id INTEGER NOT NULL,

PRIMARY KEY (note_id, tag_id),
FOREIGN KEY (note_id) REFERENCES notes (id) ON DELETE CASCADE,
FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
) ;

-- ========================================
-- 7. Links: WikiLinks with Async Resolution
-- ========================================
CREATE TABLE notes_links (
id INTEGER PRIMARY KEY AUTOINCREMENT,
src_id INTEGER NOT NULL,
dest_id INTEGER,                -- NULL = unresolved link
dest_title TEXT,                   -- Target title for resolution
-- Custom display text from [[target|display]]
display_text TEXT,
is_embed BOOLEAN DEFAULT 0,     -- 0=[[link]], 1=![[embed]]
resolved INTEGER DEFAULT 0,      -- 0=pending, 1=resolved, -1=broken
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

FOREIGN KEY (src_id) REFERENCES notes (id) ON DELETE CASCADE,
FOREIGN KEY (dest_id) REFERENCES notes (id) ON DELETE SET NULL,

-- Avoid duplicate links (when resolved)
UNIQUE (src_id, dest_id, display_text, is_embed)
) ;

-- ========================================
-- 8. Full-Text Search (SQLite FTS5)
-- ========================================
CREATE VIRTUAL TABLE notes_fts USING fts5 (
title,
body,
content = 'notes',
content_rowid = 'id'
) ;

-- Triggers to keep FTS in sync with notes
CREATE TRIGGER notes_fts_insert AFTER INSERT ON notes
BEGIN
INSERT INTO notes_fts (rowid, title, body)
VALUES (new.id, new.title, COALESCE (new.body, '')) ;
END ;

CREATE TRIGGER notes_fts_update AFTER UPDATE ON notes
BEGIN
INSERT INTO notes_fts (notes_fts, rowid, title, body)
VALUES ('delete', old.id, old.title, COALESCE (old.body, '')) ;
INSERT INTO notes_fts (rowid, title, body)
VALUES (new.id, new.title, COALESCE (new.body, '')) ;
END ;

CREATE TRIGGER notes_fts_delete AFTER DELETE ON notes
BEGIN
INSERT INTO notes_fts (notes_fts, rowid, title, body)
VALUES ('delete', old.id, old.title, COALESCE (old.body, '')) ;
END ;

-- ========================================
-- 9. Indexes: Performance Optimization
-- ========================================

-- Note Types
CREATE INDEX idx_note_types_type ON note_types (type) ;

-- Collections
CREATE INDEX idx_collections_parent ON collections (parent_id) ;
CREATE INDEX idx_collections_path ON collections (path) ;

-- Notes
CREATE INDEX idx_notes_uuid ON notes (uuid) ;
CREATE INDEX idx_notes_note_type_id ON notes (note_type_id) ;
CREATE INDEX idx_notes_collection_id ON notes (collection_id) ;
CREATE INDEX idx_notes_is_template ON notes (is_template) ;
CREATE INDEX idx_notes_created_at ON notes (created_at) ;

-- Templates
CREATE INDEX idx_templates_note_type_id ON templates (note_type_id) ;

-- Note Meta
CREATE INDEX idx_note_meta_note_id ON note_meta (note_id) ;
CREATE INDEX idx_note_meta_key ON note_meta (key) ;
CREATE INDEX idx_note_meta_key_value ON note_meta (key, value) ;
CREATE INDEX idx_note_meta_note_id_key ON note_meta (note_id, key) ;

-- Tags
CREATE INDEX idx_note_tags_note_id ON note_tags (note_id) ;
CREATE INDEX idx_note_tags_tag_id ON note_tags (tag_id) ;

-- Links
CREATE INDEX idx_notes_links_src ON notes_links (src_id) ;
CREATE INDEX idx_notes_links_dest ON notes_links (dest_id) ;
CREATE INDEX idx_notes_links_unresolved ON notes_links (resolved,
dest_title) WHERE resolved = 0 ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop in reverse order to respect foreign key constraints
DROP INDEX IF EXISTS idx_notes_links_unresolved ;
DROP INDEX IF EXISTS idx_notes_links_dest ;
DROP INDEX IF EXISTS idx_notes_links_src ;
DROP INDEX IF EXISTS idx_note_tags_tag_id ;
DROP INDEX IF EXISTS idx_note_tags_note_id ;
DROP INDEX IF EXISTS idx_note_meta_note_id_key ;
DROP INDEX IF EXISTS idx_note_meta_key_value ;
DROP INDEX IF EXISTS idx_note_meta_key ;
DROP INDEX IF EXISTS idx_note_meta_note_id ;
DROP INDEX IF EXISTS idx_templates_note_type_id ;
DROP INDEX IF EXISTS idx_notes_created_at ;
DROP INDEX IF EXISTS idx_notes_is_template ;
DROP INDEX IF EXISTS idx_notes_collection_id ;
DROP INDEX IF EXISTS idx_notes_note_type_id ;
DROP INDEX IF EXISTS idx_notes_uuid ;
DROP INDEX IF EXISTS idx_collections_path ;
DROP INDEX IF EXISTS idx_collections_parent ;
DROP INDEX IF EXISTS idx_note_types_type ;

DROP TRIGGER IF EXISTS notes_fts_delete ;
DROP TRIGGER IF EXISTS notes_fts_update ;
DROP TRIGGER IF EXISTS notes_fts_insert ;
DROP TABLE IF EXISTS notes_fts ;

DROP TABLE IF EXISTS notes_links ;
DROP TABLE IF EXISTS note_tags ;
DROP TABLE IF EXISTS tags ;
DROP TABLE IF EXISTS note_meta ;
DROP TABLE IF EXISTS templates ;
DROP TABLE IF EXISTS notes ;
DROP TABLE IF EXISTS collections ;
DROP TABLE IF EXISTS note_types ;

-- +goose StatementEnd

