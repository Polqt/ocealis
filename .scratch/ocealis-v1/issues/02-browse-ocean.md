# Browse Ocean

Status: ready-for-agent

## What to build

Visitor lands on a 2D MapLibre Ocean map: land clicks do nothing; pan/zoom seas; far zoom shows heat density; near zoom shows Corks. Seed Bottles exist and are visible (no Mystery Delay) so the sea is never empty. Map query API returns heat vs corks by zoom/viewport and excludes not-yet-visible Bottles.

Demo: open app → map → see seeds as heat/corks; click land = noop; click cork deferred to Open issue if needed (listing corks enough here).

## Acceptance criteria

- [ ] MapLibre map is the main browse surface after any placeholder home
- [ ] Land interaction is inert; Ocean is interactive for browse
- [ ] Far zoom returns/shows heat (or equivalent density), not thousands of markers
- [ ] Near zoom shows Corks in viewport (with a sane cap if needed)
- [ ] Seed Bottles are present and immediately visible
- [ ] Bottles inside Mystery Delay never appear in map query results
- [ ] Tests cover zoom policy + invisible exclusion + seed visibility

## Blocked by

- `.scratch/ocealis-v1/issues/00-prefactor-align-domain-kill-jwt.md`

## User stories

2–7

## Comments

Cast (#01) optional for demo — Seeds alone suffice. Published from `/to-issues`.
