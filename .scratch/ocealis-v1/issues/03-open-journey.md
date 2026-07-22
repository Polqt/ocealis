# Open + Journey read

Status: done

## What to build

Visitor clicks a Cork, Opens the Bottle, reads Message + Nickname, and can view Journey history. Opening does not claim or remove the Bottle; it stays drifting and readable by others.

## Acceptance criteria

- [x] Open returns Message, Nickname, and enough metadata to show the cork story
- [x] Journey lists cast/stamp/re-release/drift events in order (empty stamps OK)
- [x] Open does not change Bottle ownership or remove it from the Ocean
- [x] UI: select Cork → read Message; Journey visible
- [x] Tests: open is read-only w.r.t. claim/remove; journey ordering

## Blocked by

- `.scratch/ocealis-v1/issues/02-browse-ocean.md`

## User stories

15, 16, 20

## Comments

Published from `/to-issues` against `docs/prd-v1.md`.

2026-07-22: Open stays `GET /api/v1/bottles/:id` (no claim). Journey `GET .../journey` now returns events oldest-first (service sort + SQL ASC). Client: cork click → Open panel with Message/Nickname + Journey list. Demo: zoom to corks → click cork → read message; `go test ./internal/service/ -run 'TestOpen|TestJourney'`.
