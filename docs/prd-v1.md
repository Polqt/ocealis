# Ocealis PRD — v1

Status: ready for issue breakdown (`/to-issues`) and agent implement loops.  
Glossary: use terms from [`CONTEXT.md`](../CONTEXT.md) exactly.

## Problem Statement

People want a playful, low-pressure way to leave a short message in the world and maybe find someone else's — without accounts, chat threads, or social graphs. Existing "message in a bottle" toys are either private 1:1 links or empty gimmicks. Ocealis should feel like a real ocean you can browse: corks drift, journeys accumulate, and casting means letting go.

## Solution

An anonymous open-ocean map. Visitors cast a Bottle from their nearest Shoreline, then cannot track it. Other Visitors browse Oceans on a 2D map (heat far away, Corks up close), Open messages, Stamp journeys, and Re-release bottles so they continue elsewhere. Seed Bottles keep the sea alive. No authentication.

## User Stories

1. As a Visitor, I want to see a rotating globe splash with a bottle count, so that the world feels alive before I touch the map.
2. As a Visitor, I want to land on a 2D world map after the splash, so that I can explore oceans.
3. As a Visitor, I want land clicks to do nothing, so that bottles only exist in the Ocean.
4. As a Visitor, I want to click Ocean and pan/zoom seas, so that I can choose where to look.
5. As a Visitor, I want a heat view at world zoom, so that dense cork fields do not melt my browser.
6. As a Visitor, I want Corks to appear when I zoom into a sea, so that I can pick a bottle to open.
7. As a Visitor, I want Seed Bottles with creator/motivational messages, so that the first visit is never an empty sea.
8. As a Visitor, I want to write a Message (≤500 chars) and a required Nickname (≤24 chars), so that I can Cast without creating an account.
9. As a Visitor, I want a single Cast/Release action after writing, so that I do not pick countries or drop pins.
10. As a Visitor, I want my device location used when allowed, so that Cast feels tied to where I am.
11. As a Visitor on land, I want my Cast snapped to the nearest Shoreline just offshore, so that bottles only enter from the coast.
12. As a Visitor who denied location, I want a minimal fallback that still reaches an Ocean basin, so that Cast is not impossible on desktop (implementation may use basin pick — keep UI minimal).
13. As a caster, I want no “your bottles” list and no “your bottle is here” UI, so that letting go stays mysterious.
14. As a caster, I want my Bottle invisible for 15–30 minutes after Cast, so that I cannot immediately spot it on the nearby coast.
15. As a Visitor, I want to Open a Cork and read its Message and Nickname, so that I can receive what someone left.
16. As a Visitor, I want Opening to never claim or remove the Bottle, so that many people can find the same cork while it drifts.
17. As a Visitor, I want to Stamp a Bottle with a seal icon and/or a short note (≤80 chars), so that I leave a passport mark on its Journey.
18. As a Visitor, I want to Re-release a Bottle after finding it, so that it continues from my Shoreline into the unknown again.
19. As a Visitor, I want Re-release to require my Nickname, keep the original Message, append Journey events, apply Mystery Delay again, so that history and mystery both survive.
20. As a Visitor, I want to see a Bottle's Journey (casts, stamps, re-releases, drift highlights), so that the cork feels traveled.
21. As a Visitor, I want Bottles to Drift over time with plausible ocean motion, so that the map feels like a simulation, not static pins.
22. As a Visitor, I want Bottles to Sink after about 2–3 years, so that the ocean does not grow forever without bound.
23. As a Visitor, I want Seed Bottles to remain available long-term, so that creator voice does not vanish with Sink rules.
24. As a Visitor, I want Cast, Stamp, and Re-release protected by IP rate limits and Turnstile, so that bots cannot flood the ocean.
25. As a Visitor, I want clear length limits enforced, so that corks stay readable and payloads stay small.
26. As an operator, I want the API on Fly and Postgres managed, so that cron Drift and real-time needs have a normal server home.
27. As a developer, I want product language in `CONTEXT.md` respected in code and issues, so that agents do not reinvent synonyms.
28. As a developer, I want no auth/JWT product paths in v1, so that the open-world rule stays honest.

## Implementation Decisions

### Product rules

- Anonymous open world: no accounts, no sessions-as-identity, no “my bottles.”
- Nickname is required metadata on Cast and Re-release, not authentication.
- Map is 2D MapLibre for the main experience; splash globe may use light WebGL/Three.js, then hand off to the map.
- Land is inert; Ocean is the only interactive water surface for browse and drop.
- Cast uses geolocation when possible; inland positions snap to nearest Shoreline, drop just offshore, then Drift starts.
- Mystery Delay: random or sampled within 15–30 minutes before a new or re-released Bottle becomes a visible Cork.
- Open does not claim; Bottle stays drifting and readable by others until Sink or until Re-release relocates it.
- Stamp is one action: seal icon and/or note ≤80 chars → one Journey event.
- Re-release relocates to finder's Shoreline, new Mystery Delay, original Message immutable, Journey append-only.
- Sink after ~2–3 years for visitor Bottles; Seed Bottles immortal (or exempt from Sink).
- Abuse: IP rate limit + Cloudflare Turnstile on Cast, Stamp, Re-release; HTML sanitize messages/notes.
- Limits: Message ≤500, Nickname ≤24, Stamp note ≤80.

### Architecture

- Keep a **Go modular monolith** (one deployable) with domain modules: bottle, drift, discovery/map queries, journey/events, geo/coastline, realtime (optional for drift ticks), abuse edge.
- Do **not** extract microservices, Redis, or live current APIs in v1.
- Coastline polygons for land/ocean tests and nearest-shore snap; keep simplified gyre Drift; live currents are a later adapter behind the same Drift seam.
- Deploy API (+ scheduler) on **Fly.io**; Postgres on **Neon** or Fly Postgres; client may sit on Fly or Cloudflare Pages later.
- Existing JWT/user modules in the repo are **not** product for v1 — remove or quarantine during alignment work; do not build UI around them.
- Prefer domain vocabulary from `CONTEXT.md` in APIs and events (`cast`, `stamp`, `re_released`, `drift`, `sink`, statuses aligned to Bottle life).

### Map / client

- Far zoom: heat (or equivalent density) of Bottles.
- Near zoom: individual Corks in viewport (cap/cluster as needed for performance).
- Solid client continues from current scaffold; MapLibre for map; unused heavy 3D main-map paths stay out.

### Seams to preserve (test at these)

1. **Bottle life seam** — Cast → Mystery Delay → visible → Open → Stamp → Re-release → Sink. Highest product seam; services/handlers should express this language.
2. **Geo seam** — point on land vs ocean; nearest Shoreline snap; offshore drop. Pure geo module preferred.
3. **Map query seam** — viewport + zoom → heat vs cork list. Do not leak SQL into the client.
4. **Abuse seam** — Turnstile verify + rate limit before mutating domain.

## Testing Decisions

- Good tests assert **external behavior** through public module interfaces (HTTP API and pure geo helpers), not private cron internals or UI pixel details.
- Prefer vertical slices: one behavior per test cycle (TDD): e.g. “inland cast snaps offshore and is invisible until visible_at.”
- Test Bottle life transitions and Journey append behavior without claiming.
- Test geo: land rejected for interaction; nearest shore snap; ocean accepted.
- Test map query: zoom policy returns heat vs corks; invisible Bottles excluded.
- Test abuse edge rejects missing/invalid Turnstile and rate-limited IPs.
- Prior art: little/none in repo today — establish table-driven Go tests at service/API seams first; add client tests only for non-trivial map/query mappers if needed.

## Out of Scope

- Accounts, JWT product auth, OAuth, email magic links
- “My bottles,” sender tracking, notifications that a bottle was opened
- Live ocean current providers (Phase 2+)
- 3D globe as the main browse UI
- Directed/private bottles to a specific person
- Claiming/collecting bottles as inventory
- Chat, follows, likes-as-social-network, moderation dashboard
- Agents SDK, MCP servers, multi-agent orchestration inside the product
- Microservices, Redis pub/sub, PostGIS-required v1 (simple polygons/geo first)
- Short TTL expiry (days/weeks); Sink is multi-year only

## Further Notes

- Mood reference: https://messenger.abeto.co/ — intimate motion and atmosphere; Ocealis differs by being anonymous open-ocean browse, not 1:1 messenger.
- Root package name still `loveadrift` in places — rename opportunistically; product name is Ocealis.
- Backend architecture doc remains tactical; if it conflicts with this PRD, **this PRD + `CONTEXT.md` win**.
- Next process step: split this PRD into vertical-slice issues via `/to-issues`, then run the [agent engineering loop](./engineering-loop.md) one issue per fresh session.
