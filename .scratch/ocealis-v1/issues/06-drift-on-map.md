# Drift on map

Status: ready-for-agent

## What to build

Visible Bottles Drift over time via gyre model (existing simplified currents OK). Map reflects new positions on refresh/poll (WebSocket optional — poll is enough for v1).

## Acceptance criteria

- [ ] Scheduled Drift tick updates Bottle positions for drifting visible Bottles
- [ ] Journey can record drift-related progress as needed for “traveled” feel (not necessarily every tick)
- [ ] Map query/UI shows updated positions after poll/refresh
- [ ] Tests: tick moves bottles; invisible (Mystery Delay) bottles still progress or rules documented consistently

## Blocked by

- `.scratch/ocealis-v1/issues/02-browse-ocean.md`

## User stories

21

## Comments

Live ocean current APIs out of scope. Published from `/to-issues`.
