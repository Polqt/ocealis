# Stamp

Status: done

## What to build

Visitor Stamps an opened Bottle: seal icon and/or short note (≤80). One Journey event. Abuse edge (Turnstile + IP rate limit) on Stamp. Bottle stays in Ocean.

## Acceptance criteria

- [x] Stamp accepts seal and/or note ≤80; rejects over-limit
- [x] Journey gains a stamp event; Bottle remains discoverable
- [x] Stamp protected by Turnstile + IP rate limit
- [x] UI: Stamp from Open view
- [x] Tests: journey append + abuse rejection + limit

## Blocked by

- `.scratch/ocealis-v1/issues/03-open-journey.md`

## User stories

17, 24–25

## Comments

Published from `/to-issues` against `docs/prd-v1.md`.

Done (2026-07-29): `POST /api/v1/bottles/:id/stamp` now verifies Turnstile and
the strict IP limit, sanitizes a seal and/or ≤80-character note, and appends one
durable Stamp event without moving or claiming the Bottle. Demo: Open a Cork,
choose a seal or write a note, then select “Add Stamp” to see it in the Journey.
