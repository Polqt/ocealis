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

Visible Bottles now persist each scheduled Drift before appending a matching Journey event, while Bottles in Mystery Delay stay fixed until visible. The Ocean polls every minute and renders the latest Cork positions. Demo with `go test ./...` in `server/`, `pnpm test` in `client/`, then leave the Ocean open to see Cork positions refresh after a Drift tick.
