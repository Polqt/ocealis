# Drift on map

Status: done

## What to build

Visible Bottles Drift over time via gyre model (existing simplified currents OK). Map reflects new positions on refresh/poll (WebSocket optional — poll is enough for v1).

## Acceptance criteria

- [x] Scheduled Drift tick updates Bottle positions for drifting visible Bottles
- [x] Journey can record drift-related progress as needed for “traveled” feel (not necessarily every tick)
- [x] Map query/UI shows updated positions after poll/refresh
- [x] Tests: tick moves bottles; invisible (Mystery Delay) bottles still progress or rules documented consistently

## Blocked by

- `.scratch/ocealis-v1/issues/02-browse-ocean.md`

## User stories

21

## Comments

Live ocean current APIs out of scope. Published from `/to-issues`.

Done (2026-08-07): the scheduled tick moves only visible, released Bottles,
persists each new position, and appends matching Drift Journey progress.
Bottles in Mystery Delay remain fixed at their Shoreline drop point until
visible. The Ocean polls map discovery every 60 seconds. Demo with
`go test ./...` in `server/`, then run the API and client and watch Cork
positions refresh after a tick.
