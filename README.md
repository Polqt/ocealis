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
# Create DB, then apply schema (or use migrations)
createdb ocealis
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/ocealis?sslmode=disable

# Apply migrations with golang-migrate, goose, or plain psql:
psql "$DATABASE_URL" -f server/db/schema.sql

# Or:
make migrate
```

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
# http://localhost:3000 — set VITE_API_URL if the API is not on localhost:8080
```

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
