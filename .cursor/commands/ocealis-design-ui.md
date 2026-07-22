# Ocealis design UI

Frontend design for Ocealis (Solid + MapLibre + splash). Product mood: open ocean, Abeto-like atmosphere, not dashboard clutter.

## Steps

1. Read `CONTEXT.md`, `docs/prd-v1.md`, and the target issue (or next UI-facing open issue under `.scratch/ocealis-v1/issues/`).
2. Propose or implement UI only for that slice: MapLibre browse, Cast form, Open/Stamp/Re-release, or splash globe — whichever the issue needs.
3. Constraints: land inert; heat far → corks near; no “my bottles”; brand/ocean atmosphere over purple SaaS defaults; mobile + desktop.
4. TDD only where behavior is non-trivial (mappers, zoom policy clients). Prefer matching existing client patterns.
5. Do not expand into backend/infra unless blocked. Do not commit/push unless asked.
6. STOP when the issue’s UI acceptance criteria for this slice are met or the human has a design decision to make.
