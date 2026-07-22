# Ocealis QA

Review the latest Ocealis slice (diff vs main or uncommitted). Find real defects only.

## Steps

1. Read `CONTEXT.md`, `docs/prd-v1.md`, and the issue marked `Status: done` most recently under `.scratch/ocealis-v1/issues/` (or the path the human names).
2. Inspect the diff for that slice. Flag only issues that affect correctness, security, abuse controls, glossary violations, or PRD out-of-scope creep.
3. Output findings first: `[P0|P1|P2|P3] title — path:line` + one short paragraph each.
4. If none: `No findings.`
5. Do not implement fixes unless the human says so. Do not commit/push unless asked.
6. STOP.
