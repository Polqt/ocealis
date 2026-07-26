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

2026-07-26: Added `POST /api/v1/bottles/:id/stamp` with sanitized seal/note
validation, Turnstile, strict IP rate limiting, and durable Journey payloads.
Stamp does not move or claim the Bottle. Demo: Open a Cork, choose a seal and/or
write an ≤80-character note, then Add Stamp and see it appear in the Journey.
Verify with `go test ./...` in `server/` and `pnpm build` in `client/`.
