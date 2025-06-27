-- +goose Up
-- +goose StatementBegin
-- Drop existing tables in reverse dependency order
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS provider_settings;
DROP TABLE IF EXISTS providers;

-- Create LLM Services table (external services like OpenAI, Anthropic)
CREATE TABLE llm_services (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    adapter TEXT NOT NULL, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    api_key TEXT NOT NULL,
    base_url TEXT NOT NULL,
    organization TEXT,
    configuration TEXT NOT NULL -- default config plus llm provider specific config
);

-- Create Providers table (user-defined combinations of service + model + config)
CREATE TABLE providers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    llm_service_id INTEGER REFERENCES llm_services(id) ON DELETE CASCADE,
    system_prompt TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_providers_llm_service_id ON providers(llm_service_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_providers_llm_service_id;

-- Drop new tables
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS llm_services;

-- Recreate original tables
CREATE TABLE providers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE provider_settings (
    id INTEGER PRIMARY KEY,
    provider_id INTEGER REFERENCES providers(id) ON DELETE CASCADE,
    setting_key TEXT NOT NULL,
    setting_value TEXT NOT NULL,
    is_secret BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, setting_key)
);

CREATE TABLE models (
    id INTEGER PRIMARY KEY,
    provider_id INTEGER REFERENCES providers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    capabilities TEXT, -- JSON string
    default_params TEXT, -- JSON string
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, name)
);
-- +goose StatementEnd
