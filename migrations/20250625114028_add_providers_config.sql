-- +goose Up
-- +goose StatementBegin
-- Provider configurations
CREATE TABLE providers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Provider-specific settings (key-value pairs)
CREATE TABLE provider_settings (
    id INTEGER PRIMARY KEY,
    provider_id INTEGER REFERENCES providers(id) ON DELETE CASCADE,
    setting_key TEXT NOT NULL,
    setting_value TEXT NOT NULL,
    is_secret BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, setting_key)
);

-- Model configurations
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

-- Insert default OpenAI provider
/* INSERT INTO providers (name, type) VALUES ('openai', 'llm');

-- Insert OpenAI settings (you'll need to set these via environment or config)
INSERT INTO provider_settings (provider_id, setting_key, setting_value, is_secret) VALUES 
    (1, 'api_key', '', true),
    (1, 'base_url', 'https://api.openai.com/v1', false),
    (1, 'organization', '', false);

-- Insert default OpenAI models
INSERT INTO models (provider_id, name, capabilities, default_params) VALUES 
    (1, 'gpt-4', '{"supports_streaming": true, "max_tokens": 8192}', '{"temperature": 0.7, "max_tokens": 1000}'),
    (1, 'gpt-3.5-turbo', '{"supports_streaming": true, "max_tokens": 4096}', '{"temperature": 0.7, "max_tokens": 1000}'); */
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE models;
DROP TABLE provider_settings;
DROP TABLE providers;

-- +goose StatementEnd