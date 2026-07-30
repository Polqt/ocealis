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

2026-07-25: Added `POST /api/v1/bottles/:id/stamp`, durable seal/note Journey
metadata, Turnstile + per-IP mutation limits, and a Stamp form in the Open
panel. Demo: Open a Cork, choose a seal and/or write up to 80 characters, then
Stamp; the new event appears in Journey while the Bottle remains in the Ocean.
Verified with `go test ./...`, `go build ./...`, and `pnpm build`.
