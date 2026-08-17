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

Done (2026-08-17): the 15-minute tick now atomically persists each visible
Bottle's new gyre position and matching Drift Journey event. Bottles remain
fixed during Mystery Delay, and the visible Ocean refreshes every 60 seconds
while the page is active. Demo with the focused Drift test, or run the API and
client and observe Cork positions refresh after a tick.
