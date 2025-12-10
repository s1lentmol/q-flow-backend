-- +goose Up
-- +goose StatementBegin
ALTER TABLE queue_participants ADD COLUMN IF NOT EXISTS full_name TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE queue_participants DROP COLUMN IF EXISTS full_name;
-- +goose StatementEnd
