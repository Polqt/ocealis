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
    current_lat DOUBLE PRECISION,
    current_lng DOUBLE PRECISION,
    hops INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'drifting',
    scheduled_release TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_release BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX bottles_drifting_idx ON bottles (status, is_release) WHERE status = 'drifting' AND is_release = TRUE;
CREATE INDEX bottles_position_idx ON bottles (current_lat, current_lng) WHERE status = 'drifting' AND is_release = TRUE;

CREATE TABLE bottle_events (
    id SERIAL PRIMARY KEY,
    bottle_id INTEGER REFERENCES bottles(id),
    event_type TEXT NOT NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX bottle_events_bottle_id_idx ON bottle_events (bottle_id, id DESC);
