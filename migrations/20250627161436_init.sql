-- +goose Up
-- +goose StatementBegin
-- MindWeaver Database Schema
-- This file contains all tables in their final form after all migrations
-- 
-- Core Concept: MindWeaver is a knowledge management and AI chat application
-- that allows users to configure LLM services, create specialized providers,
-- and have conversations with AI through a flexible actor-based system.

-- LLM Services table (external services like OpenAI, Anthropic, etc.)
-- These represent the actual external LLM providers with their API credentials
-- and configuration. All provider-specific settings (temperature, max_tokens, etc.)
-- are stored in the configuration JSON field to avoid over-engineering.
CREATE TABLE llm_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,                    -- e.g., "OpenAI", "Anthropic"
    description TEXT,                             -- Human-readable description
    adapter TEXT NOT NULL,                        -- Adapter name for the service
    api_key TEXT NOT NULL,                        -- API key for the service
    base_url TEXT NOT NULL,                       -- Base URL for API calls
    organization TEXT,                            -- Organization ID (if applicable)
    configuration TEXT NOT NULL,                  -- JSON: model settings, capabilities, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Providers table (user-defined combinations of service + model + config)
-- Providers wrap LLM services with specific configurations and system prompts.
-- Each provider has a 1:1 relationship with an LLM service to avoid configuration
-- coupling issues. Users can clone services to create new providers with different configs.
CREATE TABLE providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_id INTEGER UNIQUE REFERENCES llm_services(id) ON DELETE CASCADE,  -- 1:1 relationship
    system_prompt_id INTEGER REFERENCES prompts(id) ON DELETE SET NULL,  -- Reusable system prompt
    name TEXT NOT NULL UNIQUE,                    -- e.g., "Creative GPT-4", "Analytical Claude"
    description TEXT NOT NULL,                    -- What this provider is used for
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for providers
CREATE INDEX idx_providers_llm_service_id ON providers(llm_service_id);
CREATE INDEX idx_providers_system_prompt_id ON providers(system_prompt_id);

-- Actors table (users, agents, services, system)
-- Actors represent all participants in the system - humans, AI agents, services, etc.
-- This unified approach allows for flexible conversation scenarios.
-- Participants in conversations are inferred from messages.sender_actor_id
CREATE TABLE actors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('user', 'agent', 'service', 'system')),
    name TEXT NOT NULL,                           -- Unique identifier
    display_name TEXT,                            -- Human-readable name
    avatar_url TEXT,                              -- Avatar image URL
    is_active BOOLEAN DEFAULT true,
    metadata TEXT,                                -- JSON: type-specific data (user preferences, agent capabilities, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for actors
CREATE INDEX idx_actors_type ON actors(type);
CREATE INDEX idx_actors_active ON actors(is_active);

-- Conversations table
-- Conversations are containers for related messages between participants.
-- actor_id represents the conversation owner/creator.
-- Participants are inferred from messages.sender_actor_id within the conversation.
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_id INTEGER NOT NULL REFERENCES actors(id) ON DELETE CASCADE,  -- Conversation owner
    title TEXT NOT NULL,                          -- Conversation title
    description TEXT,                             -- Optional description
    is_active BOOLEAN DEFAULT true,
    metadata TEXT,                                -- JSON: conversation-specific data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for conversations
CREATE INDEX idx_conversations_actor_id ON conversations(actor_id);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);

-- Prompts table
-- Stores both system prompts and user prompts.
-- System prompts (is_system = 1) can be reused across providers.
-- User prompts (is_system = 0) are actor-specific prompt templates.
CREATE TABLE prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_id INTEGER,                             -- NULL for system prompts, actor_id for user prompts
    title TEXT NOT NULL,                          -- Prompt title/name
    content TEXT NOT NULL,                        -- Actual prompt content
    is_system INTEGER DEFAULT 0,                  -- 1 for system prompts, 0 for user prompts
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (actor_id) REFERENCES actors(id) ON DELETE CASCADE
);

-- Create indexes for prompts
CREATE INDEX idx_prompts_actor_id ON prompts(actor_id);
CREATE INDEX idx_prompts_is_system ON prompts(is_system);

-- Messages table (replaces the old chats table)
-- Messages are the actual content exchanged in conversations.
-- UUID v7 is used for natural ordering (generated in application code).
-- Messages can be from any actor type (user, agent, service, system).
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,         -- Database relationship ID
    conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_actor_id INTEGER NOT NULL REFERENCES actors(id) ON DELETE CASCADE,  -- Who sent the message
    uuid TEXT UNIQUE NOT NULL,                    -- UUID v7 for natural ordering and external references
    content TEXT NOT NULL,                        -- Message content
    message_type TEXT DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file')),  -- Future extensibility
    metadata TEXT,                                -- JSON: message-specific data (attachments, formatting, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for messages
CREATE INDEX idx_messages_uuid ON messages(uuid);
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_sender_actor_id ON messages(sender_actor_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_conversation_ordered ON messages(conversation_id, uuid);  -- For efficient conversation retrieval 
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop tables in reverse order to respect foreign key constraints

-- Drop messages table first (references conversations and actors)
DROP TABLE IF EXISTS messages;

-- Drop conversations table (references actors)
DROP TABLE IF EXISTS conversations;

-- Drop prompts table (references actors)
DROP TABLE IF EXISTS prompts;

-- Drop providers table (references llm_services and prompts)
DROP TABLE IF EXISTS providers;

-- Drop actors table
DROP TABLE IF EXISTS actors;

-- Drop llm_services table last (no dependencies)
DROP TABLE IF EXISTS llm_services;
-- +goose StatementEnd
