# Ocealis Backend Architecture

## Product Direction

Ocealis is an anonymous shared living ocean:

- bottles are long-lived world objects, not just posts
- drift, discovery, and re-release are domain events
- the map, journey, and ambient WebSocket stream are first-class product features
- the unique part comes from persistence, motion, and surprise — not stack novelty
- no email, no OAuth; anonymous JWT is abuse control, not a user product

That means the backend should optimize for domain clarity, event history, and reliable realtime delivery.

## Architecture Decision

Build this as a modular monolith first, not independent microservices.

Why:

- current size does not justify distributed complexity
- core features share one database and one dominant domain
- you need fast iteration on drift rules, discovery rules, and realtime behavior
- clean package boundaries inside one Go module give maintainability without ops tax

Split into separate services later only when one of these becomes true:

- drift simulation needs independent scaling
- realtime fan-out becomes its own operational concern
- discovery/search needs specialized storage or geospatial indexing
- auth/users become a shared platform for other products

## Framework Choice

Stay on Fiber for this project.

- already in use
- works well for HTTP and WebSocket-heavy apps
- switching to Gin would not improve architecture quality

Do not use Hono for the Go backend. If you ever want stdlib-aligned HTTP later, prefer `net/http` or `chi` over a framework switch for its own sake.

## Core Backend Packages

- HTTP: `github.com/gofiber/fiber/v3`
- DB pool: `github.com/jackc/pgx/v5/pgxpool`
- Queries: `sqlc`
- Validation: `github.com/go-playground/validator/v10`
- Auth: `github.com/golang-jwt/jwt/v5`
- Logging: `go.uber.org/zap`
- Config: env vars (internal config package when it hurts enough)
- Jobs: in-process scheduler; extract a worker only if load demands it
- Realtime: in-memory WebSocket hub; Redis pub/sub only when horizontally scaling
- Migrations: `server/db/schema.sql` + `server/db/migrations`
- Testing: standard `testing`, table-driven tests

Defer Prometheus, OpenTelemetry, and service extraction until the client feels alive and something actually hurts.

## Current Layout

```text
server/
  main.go
  api/handler|middleware
  internal/domain|repository|service
  db/migrations|ocealis|schema.sql|query.sql
  ws/
  util/
```

Organize by domain module later if the flat handler/service/repository split becomes painful. Do not reshuffle packages preemptively.

## Domain Boundaries

1. `user` — anonymous identity, token issuance
2. `bottle` — create, read, release, state transitions
3. `discovery` — nearby search, cursor pagination
4. `drift` — simulation rules, scheduled movement/release
5. `event` — journey timeline
6. `realtime` — WebSocket subscriptions and fan-out

## API Design

- `POST /api/v1/users`
- `GET /api/v1/users/profile`
- `GET /api/v1/ocean/bottles`
- `POST /api/v1/bottles`
- `GET /api/v1/bottles/:id`
- `GET /api/v1/bottles/:id/journey`
- `POST /api/v1/bottles/:id/discover`
- `POST /api/v1/bottles/:id/release`
- `GET /api/v1/discovery`
- `GET /ws`

## Clean Code Rules

- handlers translate HTTP to service calls
- services own business rules and transactions
- repositories only access persistence
- blocking ops take `context.Context`
- typed domain errors mapped once at the API layer
- do not leak sqlc models outside repositories
- constructor injection; no package globals except process-wide infra

## Scalability Path

### Now: Strong Modular Monolith

One API service, PostgreSQL, in-process scheduler, in-memory hub, sqlc, migrations.

### Later: Production Hardening (when deploying for real users)

Migration automation in CI, metrics/tracing, request IDs, integration tests, idempotency on mutations.

### Only If Needed: Selective Extraction

`drift-worker`, `realtime-gateway`, or PostGIS discovery — never by default.

## Immediate Next Steps (MVP-critical)

1. Keep schema + migrations aligned with sqlc
2. Topic-correct WebSocket fan-out (`ocean:all`, `region:*`, `bottle:N`)
3. Pagination limits honored in SQL
4. Ocean client: ambient map, cast, discover, journey, re-release
5. Light tests on drift and discover/release rules

Do not start observability stacks or package reshuffles before the ocean feels alive.

## Recommendation Summary

- stay on Fiber
- do not build microservices yet
- keep pgx + sqlc + zap + validator + JWT
- ship the feeling (map + motion) before polish infra
