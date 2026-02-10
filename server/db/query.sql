-- name: CreateUser :one
INSERT INTO users (nickname, avatar_url) VALUES ($1, $2) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateBottle :one
INSERT INTO bottles (sender_id, message_text, bottle_style, start_lat, start_lng, scheduled_release)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetBottle :one
SELECT * FROM bottles WHERE id = $1;

-- name: CreateBottleEvent :one
INSERT INTO bottle_events (bottle_id, event_type, lat, lng)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetBottleEvents :many
SELECT * FROM bottle_events WHERE bottle_id = $1 ORDER BY created_at DESC;