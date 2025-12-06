-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS telegram_link_tokens (
    token TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    telegram_username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_link_tokens_user ON telegram_link_tokens(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS telegram_link_tokens;
-- +goose StatementEnd
