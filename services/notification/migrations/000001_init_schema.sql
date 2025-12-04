-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_contacts (
    user_id BIGINT PRIMARY KEY,
    telegram_username TEXT,
    chat_id TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_contacts;
-- +goose StatementEnd
