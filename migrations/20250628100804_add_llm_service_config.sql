-- +goose Up
-- +goose StatementBegin

-- Drop the existing llm_services table (we'll recreate it cleanly)
DROP TABLE IF EXISTS llm_services;

-- Recreate llm_services table without configuration column
CREATE TABLE llm_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,                    -- e.g., "OpenAI", "Anthropic"
    description TEXT,                             -- Human-readable description
    adapter TEXT NOT NULL,                        -- Adapter name for the service
    api_key TEXT NOT NULL,                        -- API key for the service
    base_url TEXT NOT NULL,                       -- Base URL for API calls
    organization TEXT,                            -- Organization ID (if applicable)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create llm_service_configs table for service configurations
CREATE TABLE llm_service_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_id INTEGER NOT NULL REFERENCES llm_services(id) ON DELETE CASCADE,
    name TEXT NOT NULL,                    -- e.g., "Creative GPT-4", "Analytical Claude"
    description TEXT,                      -- Optional description of this configuration
    configuration TEXT NOT NULL,           -- JSON with model, temperature, max_tokens, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for llm_service_configs
CREATE INDEX idx_llm_service_configs_service_id ON llm_service_configs(llm_service_id);
CREATE INDEX idx_llm_service_configs_name ON llm_service_configs(name);

-- Create unique constraint to prevent duplicate config names per service
CREATE UNIQUE INDEX idx_llm_service_configs_unique ON llm_service_configs(llm_service_id, name);

-- Update providers table to reference llm_service_configs instead of llm_services
DROP TABLE IF EXISTS providers;
CREATE TABLE providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_config_id INTEGER NOT NULL REFERENCES llm_service_configs(id) ON DELETE CASCADE,
    system_prompt_id INTEGER REFERENCES prompts(id) ON DELETE SET NULL,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for providers
CREATE INDEX idx_providers_config_id ON providers(llm_service_config_id);
CREATE INDEX idx_providers_prompt_id ON providers(system_prompt_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the llm_service_configs table
DROP TABLE IF EXISTS llm_service_configs;

-- Recreate the original llm_services table with configuration column
CREATE TABLE llm_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    adapter TEXT NOT NULL,
    api_key TEXT NOT NULL,
    base_url TEXT NOT NULL,
    organization TEXT,
    configuration TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Recreate the original providers table
CREATE TABLE providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_id INTEGER NOT NULL REFERENCES llm_services(id) ON DELETE CASCADE,
    system_prompt_id INTEGER REFERENCES prompts(id) ON DELETE SET NULL,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd
