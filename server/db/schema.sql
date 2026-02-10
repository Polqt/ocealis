CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    nickname TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

CREATE TABLE bottle_events (
    id SERIAL PRIMARY KEY,
    bottle_id INTEGER REFERENCES bottles(id),
    event_type TEXT NOT NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);