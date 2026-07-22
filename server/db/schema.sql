-- Ocealis bottles schema (reconstructed for sqlc; apply migrations in order).

CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    nickname   TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bottles (
    id                SERIAL PRIMARY KEY,
    sender_id         INT REFERENCES users(id),
    nickname          TEXT NOT NULL DEFAULT '',
    message_text      TEXT NOT NULL,
    bottle_style      INT DEFAULT 0,
    start_lat         DOUBLE PRECISION,
    start_lng         DOUBLE PRECISION,
    current_lat       DOUBLE PRECISION,
    current_lng       DOUBLE PRECISION,
    hops              INT DEFAULT 0,
    status            TEXT NOT NULL,
    scheduled_release TIMESTAMPTZ,
    is_release        BOOLEAN DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bottle_events (
    id         SERIAL PRIMARY KEY,
    bottle_id  INT REFERENCES bottles(id),
    event_type TEXT NOT NULL,
    lat        DOUBLE PRECISION,
    lng        DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
