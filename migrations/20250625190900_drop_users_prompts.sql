-- +goose Up
-- +goose StatementBegin
-- Drop existing users table (empty, so safe)
DROP TABLE IF EXISTS users;

-- Drop existing prompts table
DROP TABLE IF EXISTS prompts;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate original users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);

-- Create prompts table
CREATE TABLE prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_system INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd
