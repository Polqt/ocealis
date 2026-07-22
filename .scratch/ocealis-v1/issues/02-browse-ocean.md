# Browse Ocean



Status: done



## What to build



Visitor lands on a 2D MapLibre Ocean map: land clicks do nothing; pan/zoom seas; far zoom shows heat density; near zoom shows Corks. Seed Bottles exist and are visible (no Mystery Delay) so the sea is never empty. Map query API returns heat vs corks by zoom/viewport and excludes not-yet-visible Bottles.



Demo: open app → map → see seeds as heat/corks; click land = noop; click cork deferred to Open issue if needed (listing corks enough here).



## Acceptance criteria



- [x] MapLibre map is the main browse surface after any placeholder home

- [x] Land interaction is inert; Ocean is interactive for browse

- [x] Far zoom returns/shows heat (or equivalent density), not thousands of markers

- [x] Near zoom shows Corks in viewport (with a sane cap if needed)

- [x] Seed Bottles are present and immediately visible

- [x] Bottles inside Mystery Delay never appear in map query results

- [x] Tests cover zoom policy + invisible exclusion + seed visibility



## Blocked by



- `.scratch/ocealis-v1/issues/00-prefactor-align-domain-kill-jwt.md`



## User stories



2–7



## Comments



Published from `/to-issues`.



Done (2026-07-22):



- `GET /api/v1/discovery/map?min_lat&max_lat&min_lng&max_lng&zoom` → `mode: heat|corks`

- Pure `internal/discovery.QueryOcean` + in-memory `Seeds()` (style=9, negative ids); Mystery Delay filtered

- Zoom &lt; 5 = heat cells; ≥5 = Corks (cap 200)

- Client: MapLibre home `/`; Cast moved to `/cast`

- Demo: run API + `pnpm dev` in client → zoom Pacific → Seed Corks; zoom out → heat

- Skipped: DB-persisted seeds / Sink immortality (issue 07); cork Open click (issue 03)

