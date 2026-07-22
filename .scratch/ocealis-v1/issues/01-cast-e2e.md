# Cast E2E

Status: done

## What to build

A Visitor can Cast a Bottle: required Nickname + Message → Cast action → location (or minimal basin fallback) → nearest Shoreline offshore drop → Mystery Delay (15–30 min invisible) → later appears as anonymous Cork. Abuse: IP rate limit + Turnstile on Cast. Limits enforced (500/24).

End-to-end: API + minimal Cast UI + geo snap behavior + tests at Bottle life and Abuse seams. No “your bottles” UI. No country picker.

## Acceptance criteria

- [x] Cast accepts Nickname (≤24) + Message (≤500); rejects over-limit
- [x] Inland position snaps to nearest Shoreline just offshore; Ocean-only drop
- [x] Denied/missing geo has minimal basin fallback (UI stays simple)
- [x] New Bottle has Mystery Delay: invisible to map queries for 15–30 minutes
- [x] No caster tracking UI (“your bottle”, “my bottles”)
- [x] Cast requires valid Turnstile + respects IP rate limit
- [x] Message HTML sanitized
- [x] Tests cover snap + invisible-until-visible_at + abuse rejection
- [x] Minimal UI: write → Cast/Release button → success “it’s out there” (no map pin of own bottle)

## Blocked by

- `.scratch/ocealis-v1/issues/00-prefactor-align-domain-kill-jwt.md`

## User stories

8–14, 24–25

## Comments

Published from `/to-issues` against `docs/prd-v1.md`.

Implement notes:
- Apply `server/db/migrations/001_bottle_nickname.sql` before Cast against real Postgres.
- `TURNSTILE_SECRET` empty → any non-empty token accepted (local). Set secret in prod.
- Client sends `turnstile_token: "dev"` until Turnstile widget wired (`VITE_API_URL` for API).
