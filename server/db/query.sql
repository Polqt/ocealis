-- name: CreateUser :one
INSERT INTO users (nickname, avatar_url) VALUES ($1, $2)
RETURNING id, nickname, avatar_url, created_at;

-- name: GetUser :one
SELECT id, nickname, avatar_url, created_at FROM users WHERE id = $1;

-- Canonical column order for bottles: id, sender_id, message_text, bottle_style,
-- start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at

-- name: CreateBottle :one
INSERT INTO bottles (sender_id, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, scheduled_release)
VALUES ($1, $2, $3, $4, $5, $4, $5, $6)
RETURNING id, sender_id, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at;

-- name: GetBottle :one
SELECT id, sender_id, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at
FROM bottles WHERE id = $1;

-- name: UpdateBottleStatus :one
UPDATE bottles SET status = $2 WHERE id = $1
RETURNING id, sender_id, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at;

-- name: UpdateBottlePosition :one
UPDATE bottles
SET current_lat = $2, current_lng = $3, hops = hops + 1, status = $4
WHERE id = $1
RETURNING id, sender_id, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at;

-- name: ListActiveDriftingBottles :many
SELECT id, sender_id, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at
FROM bottles
WHERE status = 'drifting' AND is_release = TRUE;

-- name: CreateBottleEvent :one
INSERT INTO bottle_events (bottle_id, event_type, lat, lng)
VALUES ($1, $2, $3, $4)
RETURNING id, bottle_id, event_type, lat, lng, created_at;

-- name: GetBottleEvents :many
SELECT id, bottle_id, event_type, lat, lng, created_at
FROM bottle_events WHERE bottle_id = $1 ORDER BY created_at DESC;