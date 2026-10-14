# brillio coding challenge

## Purpose

Brillio Listings: the take-home listing search service (Go REST API + React/TypeScript UI)
for the Brillio software engineer assessment. Handout: `Coding_Test_Question_Paper.pdf`;
addendum: `Instructions.pdf`. The live session adds unseen requirements, so extension
points are first-class (docs/SPEC.md §9, docs/EXTENSIONS.md).

Engineering source of truth: `docs/SPEC.md` (tasks cite it by section number).
Design source of truth: `BRANDING.md` with `docs/branding-demo.html` as the reference rendering.

## Related Instances

None.

## Conventions

- **Spec first.** Every task description cites docs/SPEC.md and BRANDING.md sections. Deviate only by updating SPEC.md in the same branch and commenting on the task.
- **Git is local only.** No remote. `master` is the integration branch. One branch per task named `feat/BC-N-short-slug` (link it with `lattice branch-link`). Merge with `git merge --no-ff` only after validation. Conventional commit messages. No emoji in commits, code, docs or UI (BRANDING.md §8).
- **Lifecycle with no PRs.** Full pipeline is still used: in_planning → planned → in_progress → review → in_validation → pr_open → done. Here `pr_open` means "validated branch merged into local master": record validation evidence (`--role validation`) while in in_validation, transition to pr_open, merge `--no-ff` into master, then move to done once the review artifact exists. Never `--force` past a gate.
- **Validation evidence.** Pure-library, documentation and component-only tasks (BC-3, BC-6, BC-8, BC-9, BC-10, BC-11, BC-15, BC-16) record a one-line N/A justification citing test output. BC-2, BC-4, BC-5, BC-12, BC-13, BC-14 and BC-17 require evidence from a running system (curl output, browser transcript or screenshots).
- **`make check` must be green** (lint + tests, both halves) before any merge to master.
- **Actor IDs.** `agent:<model>-planner`, `agent:<model>-impl`, `agent:<model>-reviewer`, `agent:<model>-validator`. Tasks were created by `agent:fable-planner` on behalf of `human:Forrest` from the 2026-10-13 planning interview.
- **Parallel lanes.** After BC-1: BC-2 and BC-7 in parallel. Backend lane: BC-3 ‖ BC-4 → BC-5 → BC-6. Frontend lane: BC-8 → BC-9 ‖ BC-10 ‖ BC-11 → BC-12. Infra: BC-13 once BC-2 and BC-7 exist. Then BC-14, BC-15, BC-16 → BC-17.
- **Versions.** Go 1.27, modernc.org/sqlite 1.60+, Node 24, React 19, Vite 8, TypeScript 6.0.3 (pinned: typescript-eslint does not support TS 7; operator ruling 2026-10-14, SPEC §2), Vitest 5, Docker Compose v5. Use the newest stable patch at implementation time.
- **BRANDING.md changes** go through its own §9 process (edit, update the demo, republish, owner approval). Engineering deltas the branding owner has not yet absorbed are listed in docs/SPEC.md §13.
