-- +goose Up
-- +goose StatementBegin

-- ========================================
-- Brain Service Schema - Consolidated Init
-- ========================================
-- This migration consolidates all brain service schema requirements
-- from incremental development into a clean, production-ready schema.

-- ========================================
-- 1. Prompts: Reusable system prompts
-- ========================================
-- System prompts that define assistant behavior
-- Can be assigned to multiple assistants
CREATE TABLE prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,                          -- "Code Review Expert", "Research Assistant"
    content TEXT NOT NULL,                        -- The actual system prompt text
    category TEXT,                                -- "code", "writing", "research", "chat"
    is_system BOOLEAN DEFAULT 0,                  -- 1 = built-in, 0 = user-created
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- 2. Assistants: Named AI assistants
-- ========================================
-- User-created assistants with specific configurations
-- Examples: "My Code Buddy", "Research Assistant", "Writing Helper"
-- Each combines: provider + credentials + model config + system prompt
CREATE TABLE assistants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,                    -- "My Code Buddy"
    description TEXT,                             -- Optional description
    
    -- Provider Configuration
    provider_type TEXT NOT NULL,                  -- "openai", "anthropic", "ollama", "github_copilot"
    api_key TEXT,                                 -- API credentials (nullable for local models)
    base_url TEXT NOT NULL,                       -- API endpoint URL
    organization TEXT,                            -- Optional (e.g., OpenAI org ID)
    
    -- LLM Configuration (JSON for flexibility)
    llm_config TEXT NOT NULL,                     -- {"model": "gpt-4", "temperature": 0.7, "max_tokens": 2000}
    
    -- System Prompt
    system_prompt_id INTEGER,                     -- Link to prompts table
    
    -- Status
    is_active BOOLEAN DEFAULT 1,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (system_prompt_id) REFERENCES prompts(id) ON DELETE SET NULL
);

-- ========================================
-- 3. Conversations: Chat sessions
-- ========================================
-- Multi-turn conversations with assistants
-- Can be user-assistant chat, or assistant doing research/tasks
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    
    -- Which assistant is used
    assistant_id INTEGER NOT NULL,
    
    -- Conversation type and context
    conversation_type TEXT DEFAULT 'user_chat',   -- 'user_chat', 'assistant_research', 'assistant_task'
    
    -- Link to user's notes (cross-DB reference, no FK enforcement)
    linked_note_id INTEGER,                       -- References notes.id in notes.db
    
    -- Summary and metadata
    summary TEXT,                                 -- Auto-generated conversation summary
    metadata TEXT,                                -- JSON: {"tags": [...], "context": "..."}
    
    -- Status and activity
    is_active BOOLEAN DEFAULT 1,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (assistant_id) REFERENCES assistants(id) ON DELETE RESTRICT
);

-- ========================================
-- 4. Messages: Conversation content
-- ========================================
-- Individual messages within conversations
-- Uses UUID v7 for chronological ordering
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    uuid TEXT UNIQUE NOT NULL,                    -- UUID v7 for natural ordering
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata TEXT,                                -- JSON: {"tokens": 150, "model": "gpt-4", "finish_reason": "stop"}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

-- ========================================
-- 5. Assistant Notes: Assistant's own notes
-- ========================================
-- Assistant's internal knowledge base (like user's notes)
-- Assistant can create notes, reminders, observations, research
CREATE TABLE assistant_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,                    -- For sync/ordering
    title TEXT NOT NULL,
    body TEXT NOT NULL,                           -- Markdown content
    description TEXT,                             -- Optional summary
    note_type TEXT,                               -- 'reminder', 'observation', 'task', 'knowledge', 'research'
    
    -- Relations
    related_conversation_id INTEGER,              -- Link to conversation that created this
    related_note_id INTEGER,                      -- Link to user's note (in notes.db)
    created_by_assistant_id INTEGER,              -- Which assistant created this
    
    -- Task/Reminder fields
    priority INTEGER DEFAULT 0,                   -- 0-5 priority level
    due_date TIMESTAMP,                           -- For time-based reminders
    is_completed BOOLEAN DEFAULT 0,               -- For tasks
    
    -- Status
    is_active BOOLEAN DEFAULT 1,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (related_conversation_id) REFERENCES conversations(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by_assistant_id) REFERENCES assistants(id) ON DELETE SET NULL
);

-- ========================================
-- 6. Assistant Note Meta: EAV pattern
-- ========================================
-- Flexible key-value pairs for assistant notes (like note_meta)
CREATE TABLE assistant_note_meta (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assistant_note_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assistant_note_id) REFERENCES assistant_notes(id) ON DELETE CASCADE,
    UNIQUE (assistant_note_id, key)
);

-- ========================================
-- 7. Assistant Note Links: Enable Brain PKM
-- ========================================
-- Allow Brain to link its own notes (like Mind's notes_links)
-- Use Cases:
--   - Link observations over time ("continues learning X")
--   - Link hypotheses to supporting evidence
--   - Link improved prompts to original versions
--   - Build knowledge graphs of concepts
CREATE TABLE assistant_note_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Link endpoints
    src_note_id INTEGER NOT NULL,              -- Source assistant_note
    dest_note_id INTEGER NOT NULL,             -- Destination assistant_note
    
    -- Link semantics
    link_type TEXT NOT NULL DEFAULT 'related', -- 'related', 'supports', 'contradicts', 'improves', 'replaces', 'references'
    context TEXT,                              -- Why this link exists (optional explanation)
    
    -- Attribution
    created_by_assistant_id INTEGER,           -- Which assistant created this link
    
    -- Metadata
    strength REAL DEFAULT 1.0,                 -- Link confidence/weight (0.0-1.0)
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (src_note_id) REFERENCES assistant_notes(id) ON DELETE CASCADE,
    FOREIGN KEY (dest_note_id) REFERENCES assistant_notes(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by_assistant_id) REFERENCES assistants(id) ON DELETE SET NULL,
    
    -- Prevent duplicate links of same type between same notes
    UNIQUE(src_note_id, dest_note_id, link_type),
    
    -- Prevent self-links
    CHECK (src_note_id != dest_note_id)
);

-- ========================================
-- 8. Tags: Categorization system
-- ========================================
-- Tags that assistant can use to organize notes
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assistant_note_tags (
    assistant_note_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (assistant_note_id, tag_id),
    FOREIGN KEY (assistant_note_id) REFERENCES assistant_notes(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- ========================================
-- 9. Brain Tools: Tool registry
-- ========================================
-- Registry of tools available to the Brain for query execution
CREATE TABLE brain_tools (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    tool_type TEXT NOT NULL CHECK(tool_type IN ('mind_query', 'brain_query', 'composite')),
    endpoint TEXT,
    parameters TEXT, -- JSON
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- 10. Full-Text Search (SQLite FTS5)
-- ========================================
-- Enable full-text search on assistant notes (like notes_fts)
CREATE VIRTUAL TABLE assistant_notes_fts USING fts5 (
    title,
    body,
    content = 'assistant_notes',
    content_rowid = 'id'
);

-- Triggers to keep FTS in sync with assistant_notes
CREATE TRIGGER assistant_notes_ai
    AFTER INSERT ON assistant_notes
BEGIN
    INSERT INTO assistant_notes_fts (rowid, title, body) 
    VALUES (new.id, new.title, new.body);
END;

CREATE TRIGGER assistant_notes_au
    AFTER UPDATE ON assistant_notes
BEGIN
    INSERT INTO assistant_notes_fts (assistant_notes_fts, rowid, title, body) 
    VALUES ('delete', old.id, old.title, old.body);
    INSERT INTO assistant_notes_fts (rowid, title, body) 
    VALUES (new.id, new.title, new.body);
END;

CREATE TRIGGER assistant_notes_ad
    AFTER DELETE ON assistant_notes
BEGIN
    INSERT INTO assistant_notes_fts (assistant_notes_fts, rowid, title, body) 
    VALUES ('delete', old.id, old.title, old.body);
END;

-- ========================================
-- 11. Indexes: Performance Optimization
-- ========================================

-- Prompts indexes
CREATE INDEX idx_prompts_category ON prompts(category);
CREATE INDEX idx_prompts_is_system ON prompts(is_system);

-- Assistants indexes
CREATE INDEX idx_assistants_provider_type ON assistants(provider_type);
CREATE INDEX idx_assistants_is_active ON assistants(is_active);
CREATE INDEX idx_assistants_system_prompt_id ON assistants(system_prompt_id);

-- Conversations indexes
CREATE INDEX idx_conversations_assistant_id ON conversations(assistant_id);
CREATE INDEX idx_conversations_type ON conversations(conversation_type);
CREATE INDEX idx_conversations_linked_note_id ON conversations(linked_note_id);
CREATE INDEX idx_conversations_is_active ON conversations(is_active);
CREATE INDEX idx_conversations_last_activity ON conversations(last_activity DESC);
CREATE INDEX idx_conversations_created_at ON conversations(created_at DESC);

-- Messages indexes
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_uuid ON messages(uuid);
CREATE INDEX idx_messages_role ON messages(role);
-- Composite index for efficient conversation history retrieval
CREATE INDEX idx_messages_conversation_ordered ON messages(conversation_id, uuid);

-- Assistant Notes indexes (mirror notes table indexes)
CREATE INDEX idx_assistant_notes_uuid ON assistant_notes(uuid);
CREATE INDEX idx_assistant_notes_note_type ON assistant_notes(note_type);
CREATE INDEX idx_assistant_notes_related_conversation_id ON assistant_notes(related_conversation_id);
CREATE INDEX idx_assistant_notes_related_note_id ON assistant_notes(related_note_id);
CREATE INDEX idx_assistant_notes_created_by_assistant_id ON assistant_notes(created_by_assistant_id);
CREATE INDEX idx_assistant_notes_priority ON assistant_notes(priority);
CREATE INDEX idx_assistant_notes_due_date ON assistant_notes(due_date);
CREATE INDEX idx_assistant_notes_is_completed ON assistant_notes(is_completed);
CREATE INDEX idx_assistant_notes_is_active ON assistant_notes(is_active);
CREATE INDEX idx_assistant_notes_created_at ON assistant_notes(created_at DESC);

-- Assistant Note Meta indexes
CREATE INDEX idx_assistant_note_meta_assistant_note_id ON assistant_note_meta(assistant_note_id);
CREATE INDEX idx_assistant_note_meta_key ON assistant_note_meta(key);

-- Assistant Note Links indexes
CREATE INDEX idx_assistant_note_links_src ON assistant_note_links(src_note_id);
CREATE INDEX idx_assistant_note_links_dest ON assistant_note_links(dest_note_id);
CREATE INDEX idx_assistant_note_links_type ON assistant_note_links(link_type);
CREATE INDEX idx_assistant_note_links_creator ON assistant_note_links(created_by_assistant_id);
CREATE INDEX idx_assistant_note_links_bidirectional ON assistant_note_links(src_note_id, dest_note_id);

-- Tags indexes
CREATE INDEX idx_assistant_note_tags_assistant_note_id ON assistant_note_tags(assistant_note_id);
CREATE INDEX idx_assistant_note_tags_tag_id ON assistant_note_tags(tag_id);

-- Brain Tools indexes
CREATE INDEX idx_brain_tools_name ON brain_tools(name);
CREATE INDEX idx_brain_tools_type ON brain_tools(tool_type);
CREATE INDEX idx_brain_tools_uuid ON brain_tools(uuid);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop in reverse order to respect foreign key constraints
DROP INDEX IF EXISTS idx_brain_tools_uuid;
DROP INDEX IF EXISTS idx_brain_tools_type;
DROP INDEX IF EXISTS idx_brain_tools_name;
DROP INDEX IF EXISTS idx_assistant_note_tags_tag_id;
DROP INDEX IF EXISTS idx_assistant_note_tags_assistant_note_id;
DROP INDEX IF EXISTS idx_assistant_note_links_bidirectional;
DROP INDEX IF EXISTS idx_assistant_note_links_creator;
DROP INDEX IF EXISTS idx_assistant_note_links_type;
DROP INDEX IF EXISTS idx_assistant_note_links_dest;
DROP INDEX IF EXISTS idx_assistant_note_links_src;
DROP INDEX IF EXISTS idx_assistant_note_meta_key;
DROP INDEX IF EXISTS idx_assistant_note_meta_assistant_note_id;
DROP INDEX IF EXISTS idx_assistant_notes_created_at;
DROP INDEX IF EXISTS idx_assistant_notes_is_active;
DROP INDEX IF EXISTS idx_assistant_notes_is_completed;
DROP INDEX IF EXISTS idx_assistant_notes_due_date;
DROP INDEX IF EXISTS idx_assistant_notes_priority;
DROP INDEX IF EXISTS idx_assistant_notes_created_by_assistant_id;
DROP INDEX IF EXISTS idx_assistant_notes_related_note_id;
DROP INDEX IF EXISTS idx_assistant_notes_related_conversation_id;
DROP INDEX IF EXISTS idx_assistant_notes_note_type;
DROP INDEX IF EXISTS idx_assistant_notes_uuid;
DROP INDEX IF EXISTS idx_messages_conversation_ordered;
DROP INDEX IF EXISTS idx_messages_role;
DROP INDEX IF EXISTS idx_messages_uuid;
DROP INDEX IF EXISTS idx_messages_conversation_id;
DROP INDEX IF EXISTS idx_conversations_created_at;
DROP INDEX IF EXISTS idx_conversations_last_activity;
DROP INDEX IF EXISTS idx_conversations_is_active;
DROP INDEX IF EXISTS idx_conversations_linked_note_id;
DROP INDEX IF EXISTS idx_conversations_type;
DROP INDEX IF EXISTS idx_conversations_assistant_id;
DROP INDEX IF EXISTS idx_assistants_system_prompt_id;
DROP INDEX IF EXISTS idx_assistants_is_active;
DROP INDEX IF EXISTS idx_assistants_provider_type;
DROP INDEX IF EXISTS idx_prompts_is_system;
DROP INDEX IF EXISTS idx_prompts_category;

DROP TRIGGER IF EXISTS assistant_notes_ad;
DROP TRIGGER IF EXISTS assistant_notes_au;
DROP TRIGGER IF EXISTS assistant_notes_ai;
DROP TABLE IF EXISTS assistant_notes_fts;

DROP TABLE IF EXISTS brain_tools;
DROP TABLE IF EXISTS assistant_note_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS assistant_note_links;
DROP TABLE IF EXISTS assistant_note_meta;
DROP TABLE IF EXISTS assistant_notes;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS assistants;
DROP TABLE IF EXISTS prompts;

-- +goose StatementEnd
