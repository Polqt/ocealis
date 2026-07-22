# Prefactor: align Bottle domain, kill JWT product path

Status: done

## What to build

Make the codebase honest about the anonymous open world: Bottle / Journey language matches `CONTEXT.md`, and JWT/user auth is no longer a product path (remove or quarantine so no route requires a token for core bottle flows).

Demo: health + any remaining bottle read paths work without “login”; domain types/events use Cast/Stamp/Re-release/Sink vocabulary (or clear aliases mapped to it). Tests lock “no auth required for anonymous Visitor flows” where routes already exist.

## Acceptance criteria

- [x] Product HTTP paths for bottles do not require JWT
- [x] User/JWT create-login flow is removed or clearly non-product (not linked from client, not required by Cast/Open/Stamp/Re-release)
- [x] Domain statuses/events align with PRD life cycle (including room for `visible_at` / Mystery Delay and Sink — stubs OK if unused yet)
- [x] Glossary terms used in public API JSON field names or documented mapping in code comments only where rename is deferred
- [x] At least one test fails if auth middleware is re-applied to anonymous bottle reads

## Blocked by

None - can start immediately

## User stories

27, 28

## Comments

Published from `/to-issues` against `docs/prd-v1.md`.
