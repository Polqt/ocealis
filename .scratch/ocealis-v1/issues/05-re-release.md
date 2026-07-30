# Re-release

Status: done

## What to build

Finder Re-releases a Bottle: required Nickname, original Message immutable, optional prior Stamp stays on Journey, Bottle relocates to finder’s Shoreline (geo snap), new Mystery Delay, then appears elsewhere as anonymous Cork. Abuse edge on Re-release.

## Acceptance criteria

- [x] Re-release requires Nickname; keeps original Message unchanged
- [x] Journey appends re-release (and prior stamps remain)
- [x] Drop uses finder Shoreline snap; Mystery Delay 15–30 min before visible again
- [x] Old map position no longer shows the Bottle once re-released (relocated)
- [x] Turnstile + IP rate limit enforced
- [x] UI: Re-release from Open/Stamp flow
- [x] Tests: immutability, delay, relocation, abuse

## Blocked by

- `.scratch/ocealis-v1/issues/04-stamp.md`

## User stories

18, 19, 11, 14

## Comments

Re-release now validates the finder’s Nickname and abuse checks, snaps to Shoreline, preserves the Message and prior Journey, and hides the relocated Bottle for a fresh Mystery Delay. Demo: Open a Cork, enter a Nickname, choose Re-release, then see the Cork leave the Ocean until its scheduled visibility flip.
