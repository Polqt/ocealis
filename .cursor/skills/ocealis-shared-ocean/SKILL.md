---
name: ocealis-shared-ocean
description: Guides Ocealis product and engineering decisions for the anonymous shared living ocean (Go modular monolith + SolidStart). Use when planning features, choosing stack, reviewing scope, implementing cast/drift/discover/journey, or when the user mentions Ocealis, bottles, overengineering, or the shared-ocean PRD.
disable-model-invocation: true
---

# Ocealis Shared Ocean

## Product (locked)

- Anonymous shared living ocean. Anyone can cast; anyone can read discovered messages.
- No email, no OAuth, no private inbox.
- Hook: journey map + ambient ocean (persistence, motion, surprise).
- Core loop only: cast → drift → discover → read journey → re-release.

## Stack (justify once, then keep)

| Choice | Why |
|--------|-----|
| Go + Fiber | Concurrent drift ticks + in-process WS hub |
| Postgres + sqlc | Durable bottles/events, typed queries |
| Anon JWT | Rate limits / soft identity, not accounts |
| SolidStart | Lean client; do not rewrite |
| Three.js + GSAP | Ocean presence only, not CRUD UI |

Do **not** add Redis, PostGIS, OTel, Prometheus, SSE, microservices, or framework switches unless a real pain appears.

## Anti-overengineering rules

1. Ask **why** before **how** — feature must serve the core loop.
2. One deployable API; in-process cron + memory hub stay.
3. Prefer fixing broken behavior over package reshuffles (`cmd/api`, domain modules).
4. Ship the feeling (map + motion) before polish infra.
5. Public read by design.

## Out of scope

Email/SMTP, OAuth, reactions/follows/DMs, world-feed productization, Redis pub/sub, PostGIS, Workers/Agents SDK migration, full observability stacks.

## Implementation map

- Backend: `server/` — handler → service → repository → sqlc
- Realtime: `server/ws/` — subscribe `ocean:all`, `region:*`, `bottle:N`
- Schema: `server/db/schema.sql` + `server/db/migrations/`
- Client: `client/src/` — ambient ocean, cast, bottle panel + journey chapters
- Docs: `README.md`, `docs/backend-architecture.md`

## When adding code

- Keep modular monolith boundaries.
- Sanitize messages; rate-limit casts/discovers.
- WS topic fan-out must respect subscriptions (no accidental global spam).
- Pagination: SQL `LIMIT` must honor handler `limit` (fetch limit+1 for hasMore).
- Client first viewport: brand-first ocean composition, not a dashboard.

## Done means

A stranger can open the app, see live drift, cast anonymously, discover, read journey chapters, and re-release — without signup or email.
