-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes_links RENAME TO links;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links RENAME TO notes_links;
-- +goose StatementEnd
