# Stamp

Status: ready-for-agent

## What to build

Visitor Stamps an opened Bottle: seal icon and/or short note (≤80). One Journey event. Abuse edge (Turnstile + IP rate limit) on Stamp. Bottle stays in Ocean.

## Acceptance criteria

- [ ] Stamp accepts seal and/or note ≤80; rejects over-limit
- [ ] Journey gains a stamp event; Bottle remains discoverable
- [ ] Stamp protected by Turnstile + IP rate limit
- [ ] UI: Stamp from Open view
- [ ] Tests: journey append + abuse rejection + limit

## Blocked by

- `.scratch/ocealis-v1/issues/03-open-journey.md`

## User stories

17, 24–25

## Comments

Published from `/to-issues` against `docs/prd-v1.md`.
