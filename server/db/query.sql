-- name: CreateBottle :one
INSERT INTO bottles (sender_id, nickname, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, status, is_release, scheduled_release)
VALUES ($1, $2, $3, $4, $5, $6, $5, $6, $7, $8, $9)
RETURNING id, sender_id, nickname, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at;

-- name: GetBottle :one
SELECT id, sender_id, nickname, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at
FROM bottles WHERE id = $1;

-- name: UpdateBottleStatus :one
UPDATE bottles SET status = $2 WHERE id = $1
RETURNING id, sender_id, nickname, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at;

-- name: UpdateBottlePosition :one
UPDATE bottles
SET current_lat = $2,
    current_lng = $3,
    hops = hops + 1,
    status = $4,
    is_release = CASE WHEN $4 = 'drifting' THEN TRUE ELSE is_release END
WHERE id = $1
RETURNING id, sender_id, nickname, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at;

-- name: ListActiveDriftingBottles :many
SELECT id, sender_id, nickname, message_text, bottle_style, start_lat, start_lng, current_lat, current_lng, hops, status, scheduled_release, is_release, created_at
FROM bottles
WHERE status = 'drifting' AND is_release = TRUE;

-- name: ListScheduledBottles :many
SELECT id, sender_id, nickname, message_text, bottle_style, start_lat, start_lng,
       current_lat, current_lng, hops, status, scheduled_release, is_release, created_at
FROM bottles
WHERE is_release = FALSE
  AND status = 'scheduled'
  AND scheduled_release <= NOW();

-- name: GetNearbyBottles :many
SELECT id, sender_id, nickname, message_text, bottle_style,
       start_lat, start_lng, current_lat, current_lng,
       hops, status, scheduled_release, is_release, created_at
FROM bottles
WHERE status = 'drifting'
  AND is_release = TRUE
  AND current_lat BETWEEN sqlc.arg(lat)::float8 - sqlc.arg(radius_deg)::float8
                      AND sqlc.arg(lat)::float8 + sqlc.arg(radius_deg)::float8
  AND current_lng BETWEEN sqlc.arg(lng)::float8 - sqlc.arg(radius_deg)::float8
                      AND sqlc.arg(lng)::float8 + sqlc.arg(radius_deg)::float8
  AND (sqlc.narg(cursor_id)::int IS NULL OR id < sqlc.narg(cursor_id)::int)
ORDER BY id DESC
LIMIT 5;

-- name: CreateBottleEvent :one
INSERT INTO bottle_events (bottle_id, event_type, lat, lng)
VALUES ($1, $2, $3, $4)
RETURNING id, bottle_id, event_type, lat, lng, created_at;

-- name: GetBottleEvents :many
SELECT id, bottle_id, event_type, lat, lng, created_at
FROM bottle_events WHERE bottle_id = $1 ORDER BY created_at ASC, id ASC;

-- name: GetBottleEventsPaginated :many
SELECT id, bottle_id, event_type, lat, lng, created_at
FROM bottle_events
WHERE bottle_id = $1
  AND (sqlc.narg(cursor_id)::int IS NULL OR id < sqlc.narg(cursor_id)::int)
ORDER BY id DESC
LIMIT 3;

-- name: CreateUser :one
INSERT INTO users (nickname, avatar_url) VALUES ($1, $2)
RETURNING id, nickname, avatar_url, created_at;

-- name: GetUser :one
SELECT id, nickname, avatar_url, created_at FROM users WHERE id = $1;
