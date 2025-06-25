-- +goose Up
-- +goose StatementBegin
-- Drop existing users table (empty, so safe)
DROP TABLE IF EXISTS users;

-- Create actors table
CREATE TABLE actors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('user', 'agent', 'service', 'system')),
    name TEXT NOT NULL,
    display_name TEXT,
    avatar_url TEXT,
    metadata TEXT, -- JSON for type-specific data
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for actors
CREATE INDEX idx_actors_type ON actors(type);
CREATE INDEX idx_actors_active ON actors(is_active);

-- Create conversations table
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_id INTEGER NOT NULL REFERENCES actors(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    metadata TEXT, -- JSON for conversation-specific data
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for conversations
CREATE INDEX idx_conversations_actor_id ON conversations(actor_id);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);

-- Alter prompts table to use actor_id instead of user_id
ALTER TABLE prompts ADD COLUMN actor_id INTEGER REFERENCES actors(id) ON DELETE CASCADE;
CREATE INDEX idx_prompts_actor_id ON prompts(actor_id);

-- Create chats table (now linked to conversations)
CREATE TABLE chats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,
    conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    actor_id INTEGER NOT NULL REFERENCES actors(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    provider_id INTEGER REFERENCES providers(id) ON DELETE SET NULL,
    model_name TEXT,
    system_prompt_id INTEGER REFERENCES prompts(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for chats
CREATE INDEX idx_chats_uuid ON chats(uuid);
CREATE INDEX idx_chats_conversation_id ON chats(conversation_id);
CREATE INDEX idx_chats_actor_id ON chats(actor_id);
CREATE INDEX idx_chats_created_at ON chats(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_chats_created_at;
DROP INDEX IF EXISTS idx_chats_actor_id;
DROP INDEX IF EXISTS idx_chats_conversation_id;
DROP INDEX IF EXISTS idx_chats_uuid;
DROP TABLE IF EXISTS chats;
DROP INDEX IF EXISTS idx_prompts_actor_id;
ALTER TABLE prompts DROP COLUMN actor_id;
DROP INDEX IF EXISTS idx_conversations_created_at;
DROP INDEX IF EXISTS idx_conversations_actor_id;
DROP TABLE IF EXISTS conversations;
DROP INDEX IF EXISTS idx_actors_active;
DROP INDEX IF EXISTS idx_actors_type;
DROP TABLE IF EXISTS actors;
-- Recreate original users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);
-- +goose StatementEnd
