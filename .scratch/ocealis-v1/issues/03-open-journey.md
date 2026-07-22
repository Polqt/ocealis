# Open + Journey read

Status: ready-for-agent

## What to build

Visitor clicks a Cork, Opens the Bottle, reads Message + Nickname, and can view Journey history. Opening does not claim or remove the Bottle; it stays drifting and readable by others.

## Acceptance criteria

- [ ] Open returns Message, Nickname, and enough metadata to show the cork story
- [ ] Journey lists cast/stamp/re-release/drift events in order (empty stamps OK)
- [ ] Open does not change Bottle ownership or remove it from the Ocean
- [ ] UI: select Cork → read Message; Journey visible
- [ ] Tests: open is read-only w.r.t. claim/remove; journey ordering

## Blocked by

- `.scratch/ocealis-v1/issues/02-browse-ocean.md`

## User stories

15, 16, 20

## Comments

Published from `/to-issues` against `docs/prd-v1.md`.
