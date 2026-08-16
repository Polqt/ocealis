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

Implemented the scheduled gyre Drift seam with tests that lock position updates,
Journey events, and the rule that Bottles stay still during Mystery Delay. The
Ocean map now polls every minute and cleans up its timer on exit. Demo with
`go test ./internal/service -run TestDriftTick` and `pnpm test`, or leave a
visible Ocean map open through a Drift tick to see Cork positions refresh.
