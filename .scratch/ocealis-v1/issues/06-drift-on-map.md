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

2026-08-12: The scheduled tick now atomically persists each visible Bottle's
new gyre position and matching Drift Journey event. Mystery Delay Bottles remain
fixed at their Shoreline until visible, and the Ocean map polls every 60 seconds.
Demo with `go test ./internal/service -run TestDriftTick` and leave the map open
to see Cork positions refresh after a tick.
