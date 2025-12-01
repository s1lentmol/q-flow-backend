-- +goose Up
-- +goose StatementBegin
CREATE TYPE queue_mode AS ENUM ('live', 'managed', 'random', 'slots');
CREATE TYPE queue_status AS ENUM ('active', 'archived');

CREATE TABLE IF NOT EXISTS queues (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    mode queue_mode NOT NULL,
    status queue_status NOT NULL DEFAULT 'active',
    group_code TEXT NOT NULL,
    owner_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS queue_participants (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    position INT NOT NULL,
    slot_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(queue_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_queues_group ON queues(group_code);
CREATE INDEX IF NOT EXISTS idx_participants_queue_position ON queue_participants(queue_id, position);
CREATE INDEX IF NOT EXISTS idx_participants_queue_slot_time ON queue_participants(queue_id, slot_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS queue_participants;
DROP TABLE IF EXISTS queues;
DROP TYPE IF EXISTS queue_mode;
DROP TYPE IF EXISTS queue_status;
-- +goose StatementEnd
