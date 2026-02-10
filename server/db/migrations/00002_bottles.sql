-- +goose up

-- +goose statementbegin
CREATE TABLE bottles (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER REFERENCES users(id),
    message_text TEXT NOT NULL,
    bottle_style INTEGER,
    start_lat DOUBLE PRECISION,
    start_lng DOUBLE PRECISION,
    hops INTEGER DEFAULT 0,
    scheduled_release TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_release BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd