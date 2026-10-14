# Run state — brillio coding challenge (BC-1 .. BC-17)

Started 2026-10-14T00:45Z by agent:opus-5.5-orchestrator (resume mode: board pre-minted by
agent:fable-planner on behalf of human:Forrest; tickets re-bound, none re-minted).

## Configuration

- Autonomy: **Moderate** (default; operator did not override)
- Max concurrent delegators N: **5** (peak graph width is 5: BC-5 ‖ BC-9 ‖ BC-10 ‖ BC-11 ‖ BC-13)
- Worker routing (operator-confirmed 2026-10-14): **Opus 5.5 for every delegator and stage sub-agent**.
  Codex and Grok binaries are not installed, so the cross-family reviewer rule cannot be honored;
  reviews are fresh-context Claude (`claude -p` via `lattice plan-review` / `code-review --mode single`).
- Stage mode (operator-confirmed): **one delegator panel per ticket; planner, implementer, reviewer
  fallback and validator run as fresh-context in-process sub-agents (Agent tool)** — satisfies the
  project CLAUDE.md "Sub-Agent Execution Model". Plan and code reviews are explicit headless
  `lattice plan-review` / `lattice code-review` calls under a 600 s watchdog.
- Landing policy (operator-confirmed): **Orchestrator lands serially into local `master`** with
  `git merge --no-ff`, after exact-head gate. Delegators stop at `in_validation` with validation
  evidence recorded and send `LAND-READY`.
- Git remote: **none** (local only, by decision). Integration branch: `master`. No push steps anywhere.
- Status vocabulary: classic + stage11 preset. Terminal pre-merge status for delegators: **`in_validation`**
  (with `--role validation` evidence attached). Orchestrator owns `pr_open` (= "validated branch merged
  into master", per .lattice/context.md) and `done` (requires `--role review` artifact).
- Ticket fidelity: verbose (pre-minted; each cites SPEC/BRANDING sections — SPEC section numbers serve as criterion IDs).
- plan_review_mode: single · review_mode: single · auto-fire on transition: ON in config, but delegators pass
  `--no-auto-review` and run reviews explicitly (base must be `master`, head the feature branch, worktree explicit).
- Master Validator: ON · Result Validator: ON · auto-close finished panels: ON
- c11 workspace: workspace:3 B95A86DC-BFF6-4582-91A2-0749194ECB26
- Runtime/browser approval scope: local only — native API/web on per-ticket ports, c11 browser panels,
  Docker Compose on 8080/5173 for BC-13/BC-17 only.
- Delivery receipts: `.lattice/orchestration/delivery-receipts.jsonl`
- Worktree root: `/Users/forrest/Code/brillio-worktrees/<slug>`; branch `feat/BC-N-<slug>`.
- Host toolchain (verified 2026-10-14): go1.27.1, node v24.21.0 / npm 11.19.0, Docker 29.8.1,
  Compose v5.5.1, golangci-lint 2.14.0 (installed this run via brew, operator-approved), jq, sqlite3.
  Missing: `timeout`/`gtimeout` (use `perl -e 'alarm shift; exec @ARGV' 600 <cmd>`), `air` (container only).

## Workspace areas (c11 refs)

- workspace: workspace:3 B95A86DC-BFF6-4582-91A2-0749194ECB26
- main_view_area: area:6 D887B46C-F2FA-4097-ABB0-B453319B8C6B (Orchestrator panel:6 73809EE8-EE77-4F1E-8EE2-8E8B05308939)
- control_surface: area:7 A5B37262-2484-45A4-A2BE-0987250E3E95 (Lattice Board panel:11 A74A7F97-53D0-4D25-BC11-B7F4B37881B1)
- delegate_view_area_1: area:8 1F87B38B-FFDB-4C0E-A2EF-E5287C79925A (495 px tall)
- delegate_view_area_2: area:10 6D7EE8CD-366D-4B7C-9E3F-95726089895C (247 px tall — use last)
- delegate_view_area_3: area:9 29806806-32FF-427B-A78B-FA87D1CB29FD (495 px tall)
- lattice_dashboard_port: 55558 (log /tmp/lattice-dashboard-55558.log)
- Orchestrator mailbox address: brillio-orchestrator

## Port and resource allocation (parallel validation)

Native validation runs use per-ticket ports so parallel delegators never collide:
API `PORT=18000+N`, web `vite --port 15100+N --strictPort` with `VITE_API_PROXY_TARGET=http://localhost:18000+N`,
`DB_PATH=/tmp/brillio-bcN.db` (or `:memory:`). Only BC-13 and BC-17 bind 8080/5173 (Compose) and must check
`lsof -nP -iTCP:8080 -sTCP:LISTEN` / `:5173` first. Kill every process you started before reporting.

## Tickets in scope

| Ticket | Title (short) | Status | Mode | Branch | Base | Model |
|---|---|---|---|---|---|---|
| BC-1 | Repo bootstrap | backlog | fast (stage sub-agents) | master (direct) | — | opus |
| BC-2 | Backend scaffold | backlog | stage sub-agents | feat/BC-2-backend-scaffold | master | opus |
| BC-3 | Domain ranking | backlog | stage sub-agents | feat/BC-3-domain-ranking | BC-2 | opus |
| BC-4 | SQLite store | backlog | stage sub-agents | feat/BC-4-sqlite-store | BC-2 | opus |
| BC-5 | Search API | backlog | stage sub-agents | feat/BC-5-search-api | BC-3+BC-4 | opus |
| BC-6 | OpenAPI | backlog | stage sub-agents | feat/BC-6-openapi | BC-5 | opus |
| BC-7 | Frontend scaffold | backlog | stage sub-agents | feat/BC-7-frontend-scaffold | master | opus |
| BC-8 | Frontend core | backlog | stage sub-agents | feat/BC-8-frontend-core | BC-7 | opus |
| BC-9 | FilterBar | backlog | stage sub-agents | feat/BC-9-filterbar | BC-8 | opus |
| BC-10 | Results surface | backlog | stage sub-agents | feat/BC-10-results-surface | BC-8 | opus |
| BC-11 | Pager/Status/Loading | backlog | stage sub-agents | feat/BC-11-pager-status-loading | BC-8 | opus |
| BC-12 | SearchPage + e2e | backlog | stage sub-agents | feat/BC-12-search-page | BC-5,9,10,11 | opus |
| BC-13 | Compose stack | backlog | stage sub-agents | feat/BC-13-compose-stack | BC-2+BC-7 | opus |
| BC-14 | Branding pass | backlog | stage sub-agents | feat/BC-14-branding-pass | BC-12 | opus |
| BC-15 | README | backlog | stage sub-agents | feat/BC-15-readme | BC-6,12,13 | opus |
| BC-16 | Extensions doc | backlog | stage sub-agents | feat/BC-16-extensions-doc | BC-12 | opus |
| BC-17 | Release packaging | backlog | stage sub-agents | feat/BC-17-release-dist | BC-14,15,16 | opus |

Branch-base rule (press-ahead): all deps landed → base `master`. Exactly one dep unlanded and it is at
`review`/`in_validation` → base that dep's branch (with every landed dep already in master merged in).
Otherwise wait. Children record their anchor here and merge `master` (never rebase) once the anchor lands.

## Shared files — canonical owner and landing order

| File / object | Owner (creates) | Later writers (in landing order) | Conflict rule |
|---|---|---|---|
| `Makefile` | BC-1 (per-half targets: test-api/test-web/lint-api/lint-web composed by test/lint/check) | BC-2 (api halves, dev-api), BC-7 (web halves, dev-web), BC-13 (up/down/smoke), BC-17 (dist) | union, ticket order |
| `backend/cmd/api/main.go`, `httpapi` Deps/router | BC-2 | BC-4 (store + Ping), BC-5 (search, cities), BC-6 (openapi route) | serial by dependency |
| `internal/domain` types | BC-2 | BC-3 (logic files only) | BC-3 adds files; type changes go through SPEC edit |
| `frontend/src/format/*`, `src/test/fixtures.ts` | BC-8 | BC-9/10/11 may only add | union; never fork a parallel helper |
| `frontend/src/styles/tokens.css`, `global.css` | BC-7 | BC-14 fixes only | — |
| `frontend/src/App.tsx` | BC-7 (placeholder) | BC-12 (renders SearchPage) | — |
| `README.md` | BC-1 (stub) | BC-15 (real) | — |
| `docs/SPEC.md` | contract | any ticket via SPEC's own deviation rule (edit in-branch + lattice comment) | union |

## Gate (make check) rule

`make check` green on the exact branch head (after merging current `master` into it) before any merge.
Exception, intake decision 1: until both BC-2 and BC-7 have landed, the half owned by the landing ticket must be
green and the other half may still be the BC-1 "not yet implemented" stub. From the second of BC-2/BC-7 onward,
full `make check` is required with no exceptions.

## Decision log (append-only)

- 2026-10-14T00:47Z [autonomy: Moderate] Gate exception for BC-2/BC-7 (see Gate rule) — the BC-1 stub makes
  "make check green before any merge" unsatisfiable for the first scaffold to land. BC-1 adds per-half Makefile
  targets so BC-2 and BC-7 edit disjoint lines.
- 2026-10-14T00:47Z [autonomy: Moderate] SPEC §5.4 example `relevance.score` 0.8731 is arithmetically
  inconsistent with its own fields (0.6×1.0 + 0.4×0.3536 = 0.7414 at asOf 2026-10-13). BC-5 corrects the
  example in-branch per SPEC's deviation rule; BC-6's OpenAPI example uses the corrected value.
- 2026-10-14T00:47Z [autonomy: Moderate] No EVALUATION.md/BUILDPLAN.md exist; ticket acceptance + validation
  blocks and context.md lanes are the evaluation contract. Validation plan derived from them (no new criteria).
- 2026-10-14T00:47Z [autonomy: Moderate] `.lattice/` is tracked (per .lattice/.gitignore design). Feature branches
  never commit `.lattice/` (pre-commit guard in every boot prompt). Orchestrator commits board snapshots on
  master after each landing (`chore(lattice): sync board after BC-N`). BC-1's "git status clean" acceptance
  means clean outside `.lattice/` (the live board is written continuously).
- 2026-10-14T00:47Z [autonomy: Moderate] Reviews: delegators pass `--no-auto-review` and run
  `lattice code-review --mode single --base master --head <branch> --worktree <wt>` explicitly under a 600 s
  watchdog; auto-fire base resolution on this no-remote, master-not-main repo is unverified.
- 2026-10-14T00:47Z [autonomy: Moderate] BC-1 runs in the primary checkout on master (no commits exist to branch
  from). Its review is a fresh-context review sub-agent (the initial-commit diff includes PDFs and the board and
  exceeds review_max_diff_chars).
- 2026-10-14T00:47Z [autonomy: Moderate] Landing re-review rule: at landing, if master moved since the branch's
  review and the moved commits touch the same top-level package/dir as the branch, request a fresh exact-head
  code-review; otherwise gate-only, logged here.
- 2026-10-14T00:49Z [operator] golangci-lint 2.14.0 installed via brew on the host.
- 2026-10-14T01:00Z [operator ruling] **TypeScript pinned to 6.0.3**, not 7. TS 7.0.2 exports no classic compiler
  API (`exports["."]` is only `lib/version.cjs`) and typescript-eslint 8.71.2 peers `typescript >=4.8.4 <6.1.0`.
  Contract amended at ruling time (before the bootstrap commit): docs/SPEC.md §2 table + note, .lattice/context.md
  Versions line, BC-7 title, BC-7 comment.

## Run-time footguns

| Symptom | Cause | Mitigation |
|---|---|---|
| A file the ticket needs is missing from `git status`/commits with no error | Host `core.excludesFile=~/.gitignore` applies to this repo and ignores `*.out *.log *.zip *.env *.d *.bak *.orig *.old *.tmp *.pid *.[7g]z *.[jrt]ar notes.md notes.txt node_modules` (reported by BC-1 delegator 01:11Z) | Test repo rules with `git -c core.excludesFile=/dev/null check-ignore -v <path>`; `git add -f` a file that must be tracked; `make dist` must not rely on host ignore rules (BC-17) |
