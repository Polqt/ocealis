-- +goose up

-- +goose statementbegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    nickname TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd