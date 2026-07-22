# Ship infra (Fly + Postgres + Turnstile)

Status: ready-for-agent

## What to build

Production-shaped wiring: API (+ drift scheduler) on Fly.io, managed Postgres (Neon or Fly Postgres), real Cloudflare Turnstile keys for Cast/Stamp/Re-release. Client deploy path documented (Fly or Pages).

## Acceptance criteria

- [ ] API deploys to Fly and serves health + core routes against managed Postgres
- [ ] Migrations/schema apply cleanly in that environment
- [ ] Turnstile verified against Cloudflare in deployed env (not dev bypass only)
- [ ] Short runbook: env vars, deploy, seed note
- [ ] No microservices/Redis introduced

## Blocked by

- `.scratch/ocealis-v1/issues/01-cast-e2e.md`
- `.scratch/ocealis-v1/issues/03-open-journey.md`
- `.scratch/ocealis-v1/issues/04-stamp.md`
- `.scratch/ocealis-v1/issues/05-re-release.md`

## User stories

26

## Comments

Drift/Sink/Splash nice-to-have before ship but not hard-blocked here. Published from `/to-issues`.
