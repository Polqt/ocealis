# Ocealis

Anonymous message-in-a-bottle on a living ocean map. No accounts. Cast from your nearest shore, lose track of the cork, browse seas, open messages, stamp journeys, re-release bottles into the unknown.

Design mood reference: [messenger.abeto.co](https://messenger.abeto.co/)

## Product (v1)

- **Open world** — no authentication; nickname is a string on the bottle, not an identity system
- **Map** — 2D MapLibre; land clicks do nothing; heat at world zoom → corks when zoomed into a sea
- **Splash** — rotating globe + bottle count, then the map
- **Cast** — nickname + message → geolocation → nearest shoreline offshore → Mystery Delay (15–30 min) → appears as anonymous cork
- **Open / Stamp / Re-release** — read freely; stamp = seal + optional short note; re-release relocates with a new mystery delay
- **Seeds** — creator/motivational bottles so the ocean is never empty
- **Life** — bottles sink after ~2–3 years; seeds stay

Limits: message ≤500, nickname ≤24, stamp note ≤80.

Abuse floor: IP rate limits + Cloudflare Turnstile on cast, stamp, and re-release.

## Stack (v1)

| Piece | Choice |
| --- | --- |
| API | Go modular monolith on [Fly.io](https://fly.io) |
| DB | Postgres (Neon or Fly Postgres) |
| Client | Solid (SolidStart scaffold today); MapLibre for the map; light WebGL/Three only for splash globe |
| Reality | Coastline polygons now; simplified gyre drift now; live currents later |

## Repo layout

- `client/` — Solid frontend
- `server/` — Go API, drift scheduler, WebSockets
- `docs/` — PRD, engineering loop, architecture notes
- `CONTEXT.md` — domain glossary (product language)

## Docs

- [Product PRD (v1)](docs/prd-v1.md)
- [Agent engineering loop](docs/engineering-loop.md)
- [Backend architecture notes](docs/backend-architecture.md) — tactical; product truth lives in the PRD + `CONTEXT.md`

## Status

Planning locked for v1. Backend has early bottle/drift/discovery work that must be aligned to the anonymous open-world model (existing JWT/user paths are not product). Client is still scaffold.
