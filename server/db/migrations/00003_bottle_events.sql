-- +goose up

-- +goose statementbegin
CREATE TABLE bottle_events (
    id SERIAL PRIMARY KEY,
    bottle_id INTEGER REFERENCES bottles(id),
    event_type TEXT NOT NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd