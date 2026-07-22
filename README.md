# Ocealis

An anonymous shared living ocean. Cast a bottle with a message, watch it drift on simulated currents, and let strangers discover it, read it, and re-release it. There is no email and no account signup — soft anonymous identity is only for rate limits and “my cast.”

The product is the journey: persistence, motion, and surprise on a living map.

## Stack

| Layer | Choice | Why |
|-------|--------|-----|
| API | Go + Fiber | Concurrent drift ticks + WebSocket fan-out in one process |
| DB | PostgreSQL + sqlc | Durable bottles/events with typed queries |
| Realtime | In-process WebSocket hub | Live bottle positions for the ocean map |
| Client | SolidStart + Three.js + GSAP | Lean UI; Three.js only for the ocean plane |

## Quick start

### Prerequisites

- Go 1.22+
- Node 22+ and pnpm
- PostgreSQL

### Database

```bash
# Create DB once (skip if you already have ocealis)
createdb ocealis
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/ocealis?sslmode=disable

# Apply schema — safe to re-run (IF NOT EXISTS). Skip if tables already exist.
psql "$DATABASE_URL" -f server/db/schema.sql

# Or from server/:
make migrate
```

If `psql` reports `relation "users" already exists`, your database is already set up — continue to the server step.

**Windows / Git Bash:** `psql` does not read `.env`. Export `DATABASE_URL` in that terminal (user `postgres`, host `127.0.0.1`). If `psql` asks for password for user `poyhi`, `$DATABASE_URL` is empty — your OS username is being used as a fallback.

### Server

```bash
cd server
cp .env.example .env   # set DATABASE_URL and JWT_SECRET
go run .
# listens on :8080
```

### Client

```bash
cd client
pnpm install
pnpm dev
# http://localhost:3000
```

Dev uses a Vite proxy: browser calls `/api` and `/ws` on `:3000`, which forward to the Go API on `127.0.0.1:8080` (no CORS). Only set `VITE_API_URL` if you intentionally bypass the proxy.

## Core loop

1. Open the ocean — bottles drift live over WebSocket.
2. Cast anonymously — short message, bottle style, into the sea.
3. Discover — pick up a bottle, read the message, see its journey chapters.
4. Re-release — send it back into the currents; hops accumulate.

## API (v1)

- `POST /api/v1/users` — anonymous user + JWT
- `GET /api/v1/users/profile`
- `GET /api/v1/ocean/bottles` — active drifting bottles for the map
- `POST /api/v1/bottles` — cast
- `GET /api/v1/bottles/:id` / `.../journey` / `.../events`
- `POST /api/v1/bottles/:id/discover` / `.../release`
- `GET /api/v1/discovery` — nearby search
- `GET /ws` — realtime drift stream (`subscribe` to `ocean:all` or `region:*` / `bottle:N`)

## Docs

- [Backend architecture](docs/backend-architecture.md) — modular monolith, domain boundaries, what not to build yet

## Out of scope (for now)

Email, OAuth, reactions, Redis, PostGIS, microservices, full observability stack.
