# Ocealis Backend Architecture

## Product Direction

Ocealis should feel less like a CRUD app and more like a living simulation:

- bottles are long-lived world objects, not just posts
- drift, discovery, and re-release are domain events
- the map, journey, and ambient WebSocket stream are first-class product features
- the "unique and cool" part comes from persistence, motion, and surprise, not from novelty in the stack

That means the backend should optimize for domain clarity, event history, and reliable real-time delivery.

## Architecture Decision

Build this as a modular monolith first, not independent microservices.

Why:

- your current size does not justify distributed complexity yet
- the core features share one database and one dominant domain
- you need fast iteration on drift rules, discovery rules, and realtime behavior
- clean package boundaries inside one Go module give you most of the maintainability benefits now

Split into separate services later only when one of these becomes true:

- drift simulation needs independent scaling
- realtime fan-out becomes its own operational concern
- discovery/search needs specialized storage or geospatial indexing
- auth/users become a shared platform for other products

## Framework Choice

Use Fiber for this project and stay on it for now.

Fiber is a reasonable fit here because:

- you already started with it
- it works well for HTTP and WebSocket-heavy applications
- route grouping and middleware are straightforward
- switching to Gin now would not materially improve architecture quality

Do not use Hono for the Go backend. Hono is primarily a JavaScript and TypeScript framework. Your frontend can keep its own ecosystem; the Go backend should stay Go-native.

If you ever want the most standard-library-aligned backend later, prefer `net/http` or `chi` over a framework switch for its own sake.

## Core Backend Packages

Recommended packages and tools:

- HTTP framework: `github.com/gofiber/fiber/v3`
- Database driver and pool: `github.com/jackc/pgx/v5/pgxpool`
- Query generation: `sqlc`
- Validation: `github.com/go-playground/validator/v10`
- Auth: `github.com/golang-jwt/jwt/v5`
- Logging: `go.uber.org/zap`
- Config: internal config package backed by env vars
- Background jobs: start with in-process scheduler; later move to a worker service if load demands it
- Realtime: keep WebSocket hub internal for now; introduce Redis pub/sub only when horizontally scaling
- Testing: standard `testing`, `httptest`, table-driven tests, `testify/require` if you want better assertions
- Observability: Prometheus, OpenTelemetry, and `pprof`
- Concurrency helpers: `golang.org/x/sync/errgroup`
- Migrations: `golang-migrate/migrate` or Atlas

## Recommended Module Shape

Keep one deployable backend, but make the internals explicit:

```text
server/
  cmd/
    api/
      main.go
  internal/
    app/
      bootstrap.go
    config/
      config.go
    platform/
      db/
      logger/
      auth/
      telemetry/
    module/
      bottle/
        handler.go
        service.go
        repository.go
        dto.go
      discovery/
        handler.go
        service.go
      drift/
        service.go
        scheduler.go
      event/
        handler.go
        repository.go
      user/
        handler.go
        service.go
        repository.go
      realtime/
        hub.go
        client.go
        broadcaster.go
    domain/
      bottle.go
      bottle_event.go
      user.go
      errors.go
  db/
    migrations/
    query/
    ocealis/
```

Important rule: organize by domain module, not by technical layer only. Your current `handler/service/repository` split is fine, but once the codebase grows, feature grouping becomes easier to maintain.

## Domain Boundaries

Start with these backend modules:

1. `user`
   Anonymous identity, profile, token issuance.

2. `bottle`
   Create bottle, read bottle, release bottle, bottle state transitions.

3. `discovery`
   Nearby search, cursor pagination, future geospatial optimization.

4. `drift`
   Simulation rules, scheduled movement, scheduled release.

5. `event`
   Journey timeline, event history, replay-friendly reads.

6. `realtime`
   WebSocket subscriptions and outbound event fan-out.

## API Design Direction

Keep external APIs simple and product-driven:

- `POST /api/v1/users`
- `GET /api/v1/users/profile`
- `POST /api/v1/bottles`
- `GET /api/v1/bottles/:id`
- `GET /api/v1/bottles/:id/journey`
- `POST /api/v1/bottles/:id/discover`
- `POST /api/v1/bottles/:id/release`
- `GET /api/v1/discovery`
- `GET /ws`

Near-term additions worth planning:

- `GET /api/v1/feed/world` for a global drifting feed
- `GET /api/v1/bottles/:id/stream` if you later need per-bottle realtime channels over SSE
- `POST /api/v1/bottles/:id/react` if you add lightweight social interactions

## Clean Code Rules

Use these rules consistently:

- handlers only translate HTTP to service calls
- services own business rules and transactions
- repositories only access persistence
- all blocking operations accept `context.Context`
- return typed domain errors and map them once at the API layer
- avoid leaking SQLC models outside repository boundaries
- keep interfaces small and local to consumers
- use constructor injection for dependencies
- do not use package globals except for tightly controlled process-wide infrastructure

## Scalability Path

Phase the backend like this:

### Phase 1: Strong Modular Monolith

Ship one API service with:

- PostgreSQL or Neon
- in-process scheduler
- in-memory WebSocket hub
- sqlc repositories
- clean module boundaries

### Phase 2: Production Hardening

Add:

- migration tooling
- metrics, tracing, structured logs
- request IDs
- graceful shutdown for scheduler and sockets
- integration tests against Postgres
- idempotency protection for mutation endpoints

### Phase 3: Selective Service Extraction

Only extract when needed:

- `drift-worker` if simulation becomes heavy
- `realtime-gateway` if WebSocket fan-out needs separate scaling
- `discovery-service` if geospatial queries need PostGIS or a specialized search store

## Immediate Next Steps

Execute backend work in this order:

1. stabilize bootstrap and dependency wiring
2. add a real config package for env loading and validation
3. define API error model and domain error mapping
4. finish bottle, discovery, event, and auth endpoints
5. add tests for handlers, services, and repositories
6. add observability and migration tooling
7. profile drift and discovery logic before any service split

## Recommendation Summary

- stay on Fiber
- do not switch to Gin right now
- do not build microservices yet
- build a disciplined modular monolith with domain modules
- keep pgx + sqlc + zap + validator + JWT
- add config, migrations, metrics, tracing, tests, and error mapping next
