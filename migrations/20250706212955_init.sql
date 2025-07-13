-- +goose Up
-- +goose StatementBegin
-- MindWeaver Database Schema - Complete MVP Schema
-- This file contains all tables with proper foreign key constraints and ownership model
-- 
-- Core Concept: MindWeaver is a knowledge management and AI chat application
-- that allows users to configure LLM services, create specialized providers,
-- and have conversations with AI through a flexible actor-based system.

-- LLM Services table (external services like OpenAI, Anthropic, etc.)
-- These represent the actual external LLM providers with their API credentials
-- and configuration. All provider-specific settings are stored in llm_service_configs.
CREATE TABLE llm_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,                    -- e.g., "OpenAI", "Anthropic"
    description TEXT,                             -- Human-readable description
    adapter TEXT NOT NULL,                        -- Adapter name for the service
    api_key TEXT NOT NULL,                        -- API key for the service
    base_url TEXT NOT NULL,                       -- Base URL for API calls
    organization TEXT,                            -- Organization ID (if applicable)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT
);

-- Create indexes for llm_services
CREATE INDEX idx_llm_services_created_by ON llm_services(created_by);
CREATE INDEX idx_llm_services_updated_by ON llm_services(updated_by);

-- LLM Service Configurations table
-- Stores different configurations for each LLM service (model, temperature, etc.)
CREATE TABLE llm_service_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_id INTEGER NOT NULL REFERENCES llm_services(id) ON DELETE CASCADE,
    name TEXT NOT NULL,                           -- e.g., "Creative GPT-4", "Analytical Claude"
    description TEXT,                             -- Optional description of this configuration
    configuration TEXT NOT NULL,                  -- JSON with model, temperature, max_tokens, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT
);

-- Create indexes for llm_service_configs
CREATE INDEX idx_llm_service_configs_service_id ON llm_service_configs(llm_service_id);
CREATE INDEX idx_llm_service_configs_name ON llm_service_configs(name);
CREATE INDEX idx_llm_service_configs_created_by ON llm_service_configs(created_by);
CREATE INDEX idx_llm_service_configs_updated_by ON llm_service_configs(updated_by);

-- Create unique constraint to prevent duplicate config names per service
CREATE UNIQUE INDEX idx_llm_service_configs_unique ON llm_service_configs(llm_service_id, name);

-- Models table for caching model data from LLM services
CREATE TABLE models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_id INTEGER NOT NULL REFERENCES llm_services(id) ON DELETE CASCADE,
    model_id TEXT NOT NULL,                       -- Model identifier from service
    name TEXT NOT NULL,                           -- Human-readable model name
    provider TEXT NOT NULL,                       -- Provider name
    description TEXT,                             -- Model description
    created_at INTEGER,                           -- Model creation timestamp
    owned_by TEXT,                                -- Model owner
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT,
    UNIQUE(llm_service_id, model_id)
);

-- Create indexes for models
CREATE INDEX idx_models_llm_service_id ON models(llm_service_id);
CREATE INDEX idx_models_last_fetched ON models(last_fetched_at);
CREATE INDEX idx_models_created_by ON models(created_by);
CREATE INDEX idx_models_updated_by ON models(updated_by);

-- Actors table (users, agents, services, system)
-- Actors represent all participants in the system - humans, AI agents, services, etc.
-- This unified approach allows for flexible conversation scenarios.
-- Note: created_by and updated_by are nullable to handle self-references
CREATE TABLE actors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('user', 'agent', 'service', 'system')),
    name TEXT NOT NULL,                           -- Unique identifier
    display_name TEXT,                            -- Human-readable name
    avatar_url TEXT,                              -- Avatar image URL
    is_active BOOLEAN DEFAULT true,
    metadata TEXT,                                -- JSON: type-specific data (user preferences, agent capabilities, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,                           -- Foreign key to actors.id (nullable for self-references)
    updated_by INTEGER,                           -- Foreign key to actors.id (nullable for self-references)
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE SET NULL
);

-- Create indexes for actors
CREATE INDEX idx_actors_type ON actors(type);
CREATE INDEX idx_actors_active ON actors(is_active);
CREATE INDEX idx_actors_created_by ON actors(created_by);
CREATE INDEX idx_actors_updated_by ON actors(updated_by);

-- Prompts table
-- Stores both system prompts and user prompts.
-- System prompts (is_system = 1) can be reused across providers.
-- User prompts (is_system = 0) are actor-specific prompt templates.
-- Ownership is determined by created_by field
CREATE TABLE prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,                          -- Prompt title/name
    content TEXT NOT NULL,                        -- Actual prompt content
    is_system INTEGER DEFAULT 0,                  -- 1 for system prompts, 0 for user prompts
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id (ownership)
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT
);

-- Create indexes for prompts
CREATE INDEX idx_prompts_is_system ON prompts(is_system);
CREATE INDEX idx_prompts_created_by ON prompts(created_by);
CREATE INDEX idx_prompts_updated_by ON prompts(updated_by);

-- Providers table (user-defined combinations of service + model + config)
-- Providers wrap LLM service configs with specific system prompts.
CREATE TABLE providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_config_id INTEGER NOT NULL REFERENCES llm_service_configs(id) ON DELETE CASCADE,
    system_prompt_id INTEGER REFERENCES prompts(id) ON DELETE SET NULL,  -- Reusable system prompt
    name TEXT NOT NULL UNIQUE,                    -- e.g., "Creative GPT-4", "Analytical Claude"
    description TEXT,                             -- What this provider is used for
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT
);

-- Create indexes for providers
CREATE INDEX idx_providers_config_id ON providers(llm_service_config_id);
CREATE INDEX idx_providers_prompt_id ON providers(system_prompt_id);
CREATE INDEX idx_providers_created_by ON providers(created_by);
CREATE INDEX idx_providers_updated_by ON providers(updated_by);

-- Conversations table
-- Conversations are containers for related messages between participants.
-- Ownership is determined by created_by field
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,                          -- Conversation title
    description TEXT,                             -- Optional description
    is_active BOOLEAN DEFAULT true,
    metadata TEXT,                                -- JSON: conversation-specific data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id (ownership)
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT
);

-- Create indexes for conversations
CREATE INDEX idx_conversations_created_at ON conversations(created_at);
CREATE INDEX idx_conversations_created_by ON conversations(created_by);
CREATE INDEX idx_conversations_updated_by ON conversations(updated_by);

-- Messages table
-- Messages are the actual content exchanged in conversations.
-- UUID v7 is used for natural ordering (generated in application code).
-- Messages can be from any actor type (user, agent, service, system).
-- Ownership is determined by created_by field (who wrote the message)
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,         -- Database relationship ID
    conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    uuid TEXT UNIQUE NOT NULL,                    -- UUID v7 for natural ordering and external references
    content TEXT NOT NULL,                        -- Message content
    message_type TEXT DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file')),  -- Future extensibility
    metadata TEXT,                                -- JSON: message-specific data (attachments, formatting, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,                  -- Foreign key to actors.id (who wrote the message)
    updated_by INTEGER NOT NULL,                  -- Foreign key to actors.id
    FOREIGN KEY (created_by) REFERENCES actors(id) ON DELETE RESTRICT,
    FOREIGN KEY (updated_by) REFERENCES actors(id) ON DELETE RESTRICT
);

-- Create indexes for messages
CREATE INDEX idx_messages_uuid ON messages(uuid);
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_conversation_ordered ON messages(conversation_id, uuid);  -- For efficient conversation retrieval
CREATE INDEX idx_messages_created_by ON messages(created_by);
CREATE INDEX idx_messages_updated_by ON messages(updated_by);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop tables in reverse order to respect foreign key constraints

-- Drop messages table first (references conversations and actors)
DROP TABLE IF EXISTS messages;

-- Drop conversations table (references actors)
DROP TABLE IF EXISTS conversations;

-- Drop providers table (references llm_service_configs and prompts)
DROP TABLE IF EXISTS providers;

-- Drop prompts table (references actors)
DROP TABLE IF EXISTS prompts;

-- Drop actors table
DROP TABLE IF EXISTS actors;

-- Drop models table (references llm_services)
DROP TABLE IF EXISTS models;

-- Drop llm_service_configs table (references llm_services)
DROP TABLE IF EXISTS llm_service_configs;

-- Drop llm_services table last (no dependencies)
DROP TABLE IF EXISTS llm_services;
-- +goose StatementEnd
