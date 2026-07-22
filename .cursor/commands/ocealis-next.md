# Ocealis next issue

Implement the **next** Ocealis v1 vertical slice. One issue only. Then stop.

## How this works (read once)

This is **not** `/loop` and **not** “finish the whole PRD unattended.”

- **You (human):** run `/ocealis-next` (or say “next”) when ready for another slice.
- **Agent:** finds the next unfinished issue, implements it with TDD, marks it done, stops.
- **Why not full auto-chain:** context melts, wrong merges, skipped review. Real agentic engineering here = **queued one-shots**, not infinite autonomous coding.

Optional later: Cursor Cloud Agent / Automation that runs this same prompt on a schedule — still one issue per run.

## Steps

1. Read `CONTEXT.md`, `docs/prd-v1.md`, `docs/engineering-loop.md`, `docs/agents/issue-tracker.md`.
2. Scan `.scratch/ocealis-v1/issues/` for `*.md`.
3. Pick the **lowest `NN`** where `Status:` is not `done` (treat `ready-for-agent` / missing status as eligible). Skip files whose **Blocked by** targets are not all `Status: done`.
4. If none eligible: say so and stop. List what’s blocked.
5. Announce the chosen path in one line. Then implement **only that issue**.
6. Constraints (always):
   - Styles: caveman prose + ponytail YAGNI
   - TDD: failing test first for new behavior; watch RED then GREEN
   - Glossary from `CONTEXT.md` (Bottle, Cast, Cork, Stamp, Journey, Visitor…)
   - No auth/JWT product paths
   - Respect PRD Out of Scope
   - **Do not commit or push** unless the human explicitly asks in this chat
7. When acceptance criteria met: check them off in the issue file, set `Status: done`, add a short `## Comments` note (what changed, how to demo).
8. **STOP.** Do not start the next issue. Reply with the next eligible issue path (or “v1 queue empty”).

## Handoff line for human (end of reply)

```text
Next: /ocealis-next  →  <path or none>
```
