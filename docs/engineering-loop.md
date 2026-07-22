# Ocealis — Agent engineering loop

How to build v1 without melting context windows or overbuilding. Product truth: [`prd-v1.md`](./prd-v1.md) + [`CONTEXT.md`](../CONTEXT.md).

## Principles

- **One issue per session.** Fresh agent context each vertical slice.
- **TDD vertical slices.** One failing behavior → minimal code → green → next. No horizontal “all tests then all code.”
- **Glossary is law.** Say Bottle, Cast, Stamp, Cork, Journey — not post/user/like.
- **YAGNI.** If it is in PRD Out of Scope, agents must not “helpfully” add it.
- **No commit/push unless the human explicitly asks.**

## Driver commands (slash)

| Command | Role |
| --- | --- |
| `/ocealis-next` | **Implement** next eligible issue (one shot) |
| `/ocealis-plan` | **Plan** / grill — no code |
| `/ocealis-qa` | **QA / review** last slice — findings only |
| `/ocealis-design-ui` | **Frontend design** for a UI slice |

Backend + infra stay inside `/ocealis-next` when the issue is server/Fly (issues 00–09 already cover that). Do not invent parallel “do all layers” mega-agent.

Cloud Automation (optional): schedule or webhook runs the **same** `/ocealis-next` prompt on `Polqt/ocealis` — still one issue per run.

## Loop (repeat per issue)

```text
1. PICK     Grab one ready issue (from /to-issues breakdown)
2. CONTEXT  Read CONTEXT.md + prd-v1.md + the single issue
3. SEAM     Name the seam under test (Bottle life / Geo / Map query / Abuse)
4. RED      Write one failing test for one acceptance criterion
5. GREEN    Minimal production code to pass
6. REFACTOR Keep tests green; deepen module if interface stays small
7. DEMO     Show the slice works (API call, map action, or test output)
8. STOP     Do not start the next issue in the same degraded window
```

## Agent roles (same human, different sessions)

Use these as **session modes**, not microservices.

| Mode | Job | Skills / commands |
| --- | --- | --- |
| **Planner** | Grill, PRD, issue split | `/grill-with-docs`, `/to-prd`, `/to-issues` |
| **Implementer** | One issue, TDD | `/implement` or agent with `tdd` + `ponytail` |
| **Reviewer** | Diff only; find real defects | `/review-agent` or cavecrew-reviewer |
| **Architect** | Occasional deepening | `/improve-codebase-architecture` when idle |
| **Debugger** | Failing slice | `/diagnosing-bugs` / systematic debugging |

Do **not** run Planner + Implementer + Reviewer in one mega-prompt. Hand off with a short issue id + PRD link.

## Published v1 issues (local)

Tracker: `.scratch/ocealis-v1/issues/` — see `docs/agents/issue-tracker.md`.

| # | File | Start when |
| --- | --- | --- |
| 00 | `00-prefactor-align-domain-kill-jwt.md` | now |
| 01 | `01-cast-e2e.md` | after 00 |
| 02 | `02-browse-ocean.md` | after 00 (parallel with 01) |
| 03 | `03-open-journey.md` | after 02 |
| 04 | `04-stamp.md` | after 03 |
| 05 | `05-re-release.md` | after 04 |
| 06 | `06-drift-on-map.md` | after 02 |
| 07 | `07-sink-seeds.md` | after 02 |
| 08 | `08-splash-globe.md` | after 02 |
| 09 | `09-ship-infra.md` | after 01 + 03 + 04 + 05 |

One fresh chat per file. Do not chain issues in one window.

## Handoff

**Preferred:** new chat → type `/ocealis-next`

**Manual override** (specific issue):

```text
Implement Ocealis issue: .scratch/ocealis-v1/issues/02-browse-ocean.md
Read: CONTEXT.md, docs/prd-v1.md, docs/engineering-loop.md, that issue file
Constraints: caveman + ponytail + TDD; no auth; glossary terms; do not commit/push unless I say so
```

## Done for a slice

- [ ] Acceptance criteria checked
- [ ] Tests watched fail then pass for new behavior
- [ ] No Out-of-Scope features added
- [ ] Human reviewed or `/review-agent` on the diff
- [ ] Commit only if human asked
