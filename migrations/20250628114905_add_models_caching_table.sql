-- +goose Up
-- +goose StatementBegin
-- Create models table for caching model data from LLM services
CREATE TABLE IF NOT EXISTS models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    llm_service_id INTEGER NOT NULL,
    model_id TEXT NOT NULL,
    name TEXT NOT NULL,
    provider TEXT NOT NULL,
    description TEXT,
    created_at INTEGER,
    owned_by TEXT,
    last_fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (llm_service_id) REFERENCES llm_services(id) ON DELETE CASCADE,
    UNIQUE(llm_service_id, model_id)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_models_llm_service_id ON models(llm_service_id);
CREATE INDEX IF NOT EXISTS idx_models_last_fetched ON models(last_fetched_at); 
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes first
DROP INDEX IF EXISTS idx_models_last_fetched;
DROP INDEX IF EXISTS idx_models_llm_service_id;

-- Drop the models table
DROP TABLE IF EXISTS models;
-- +goose StatementEnd
