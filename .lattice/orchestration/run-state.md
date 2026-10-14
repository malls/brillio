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
- delegate_view_area_1: area:11 C4DECFF8-1150-4E6E-9C17-277B9378CA7B (right column, under the board; placeholder panel:15 17F67F57-BB05-4DCD-9BA6-8D1C4BDEB7B1)
- delegate_view_area_2: area:9 29806806-32FF-427B-A78B-FA87D1CB29FD (left column, under Main View; placeholder panel:9 DD0783AA-5E18-4BB4-9697-4D7AB4F45C1F)
- (areas 8 and 10 collapsed at 01:20Z when their last panel closed; two delegate areas suffice for N=5 with panels as tabs; soft cap 15 panels/area)
- Master Validator: panel:13 AFD95ABD-DC58-4303-BEBE-8591321113A7 in area:6
- lattice_dashboard_port: 55558 (log /tmp/lattice-dashboard-55558.log)
- Orchestrator mailbox address: brillio-orchestrator. Orchestrator → delegator mail: `c11 mailbox send --to id:<panel UUID from agents.md> "<text>"` (delegator mailbox.address values are not reliable).

## Port and resource allocation (parallel validation)

Native validation runs use per-ticket ports so parallel delegators never collide:
API `PORT=18000+N`, web `vite --port 15100+N --strictPort` with `VITE_API_PROXY_TARGET=http://localhost:18000+N`,
`DB_PATH=/tmp/brillio-bcN.db` (or `:memory:`). Only BC-13 and BC-17 bind 8080/5173 (Compose) and must check
`lsof -nP -iTCP:8080 -sTCP:LISTEN` / `:5173` first. Kill every process you started before reporting.

## Tickets in scope

| Ticket | Title (short) | Status | Mode | Branch | Base | Model |
|---|---|---|---|---|---|---|
| BC-1 | Repo bootstrap | done (f4f8a0b) | fast (stage sub-agents) | master (direct) | — | opus |
| BC-2 | Backend scaffold | done 02:09Z (merge b273a8a, head a9b7f73) | stage sub-agents | feat/BC-2-backend-scaffold | master | opus |
| BC-3 | Domain ranking | dispatched 01:49Z panel:18 | stage sub-agents | feat/BC-3-domain-ranking | anchor feat/BC-2 @5da9b80 | opus |
| BC-4 | SQLite store | dispatched 01:49Z panel:19 | stage sub-agents | feat/BC-4-sqlite-store | anchor feat/BC-2 @5da9b80 | opus |
| BC-5 | Search API | backlog | stage sub-agents | feat/BC-5-search-api | BC-3+BC-4 | opus |
| BC-6 | OpenAPI | backlog | stage sub-agents | feat/BC-6-openapi | BC-5 | opus |
| BC-7 | Frontend scaffold | review ~02:03Z (head 1d6008a) panel:17 | stage sub-agents | feat/BC-7-frontend-scaffold | master | opus |
| BC-8 | Frontend core | dispatched 02:06Z panel:20 | stage sub-agents | feat/BC-8-frontend-core | anchor feat/BC-7 @1d6008a | opus |
| BC-9 | FilterBar | backlog | stage sub-agents | feat/BC-9-filterbar | BC-8 | opus |
| BC-10 | Results surface | backlog | stage sub-agents | feat/BC-10-results-surface | BC-8 | opus |
| BC-11 | Pager/Status/Loading | backlog | stage sub-agents | feat/BC-11-pager-status-loading | BC-8 | opus |
| BC-12 | SearchPage + e2e | backlog | stage sub-agents | feat/BC-12-search-page | BC-5,9,10,11 | opus |
| BC-13 | Compose stack | dispatched 02:12Z panel:21 (validation hold for BC-5) | stage sub-agents | feat/BC-13-compose-stack | anchor feat/BC-7 @1d6008a + master @7ab182d | opus |
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
| `internal/domain` types | BC-2 (frozen at 5da9b80) | BC-3 (logic files only) | BC-3 adds files; type changes go through SPEC edit. `Listing` is the in-memory model: never `json.Marshal`/`Unmarshal` `Listing` (or `Scored`, which embeds it) on the wire; `ListedDate` is `time.Time` — format/parse with `domain.DateLayout` in boundary DTOs (BC-2 note 01:47Z) |
| `frontend/src/format/*`, `src/test/fixtures.ts` | BC-8 | BC-9/10/11 may only add | union; never fork a parallel helper |
| `frontend/src/styles/tokens.css`, `global.css` | BC-7 | BC-14 fixes only | — |
| `frontend/src/App.tsx` | BC-7 (placeholder) | BC-12 (renders SearchPage) | — |
| `README.md` | BC-1 (stub) | BC-15 (real) | — |
| `docs/SPEC.md` | contract | any ticket via SPEC's own deviation rule (edit in-branch + lattice comment) | union |

## Landing checklist additions

- BC-7 landing: also add the ESLint 9.39.5 pin (jsx-a11y peer) to `.lattice/context.md` Versions, matching BC-7's SPEC §2 edit (MV 02:10Z).

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
- 2026-10-14T01:18Z [autonomy: Moderate] BC-1 landed (direct on master f4f8a0b; review PASS-WITH-NITS art_01M4VZQ546WYVMW5DWJH6HMBPN, two .gitignore nits forwarded to BC-2 card item 7; validation N/A art_01M4VZR7YZVPYGH5N4MW71Q9MD). Board sync 2afc005.
- 2026-10-14T02:12Z [autonomy: Moderate] BC-13 now depends_on BC-5 (make smoke asserts search/cities behavior). Press-ahead exception: BC-13 builds on feat/BC-7 + master now, validates compose/health/proxy/hot-reload, then holds at in_validation until BC-5 lands; only then full make smoke + LAND-READY.
- 2026-10-14T02:09Z [autonomy: Moderate] BC-2 landed: merge b273a8a (--no-ff, tree == branch tree 3cd22df), half-gate make lint-api test-api pass (uncached go test -race re-run by orchestrator), review art_01M4W2E06P2SMDVN8PGN8XWYVK PASS @ a9b7f73 with 2 minors skipped-with-reason by delegator, validation art_01M4W2M170ZQMTYVSWZDQZGD5A. Board sync 7ab182d. BC-3/BC-4 mailed to merge master.
- 2026-10-14T02:06Z [autonomy: Moderate] Press-ahead: BC-8 cut from feat/BC-7-frontend-scaffold @1d6008a (BC-7 at review).
- 2026-10-14T01:57Z [autonomy: Moderate] Accepted BC-7 deviation: ESLint held at 9.39.5 (not 10.12.0) because allowlisted eslint-plugin-jsx-a11y 6.10.2 (latest) peers `eslint ^3..^9` (verified with npm view). Same principle as the operator's TS 6.0.3 ruling: compatible versions, no overrides. BC-7 records it as an in-branch SPEC §2 edit.
- 2026-10-14T01:55Z [autonomy: Moderate] Accepted (MV 01:53Z): headless `lattice plan-review`/`code-review` runs `claude -p` on its default model, which self-identifies as **Fable 5.1**, not Opus 5.5. Kept as-is: reviewers then differ from the Opus implementers, which is closer to the skill's cross-model review intent than an all-Opus run. Own-reviewer fallback sub-agents remain Opus.
- 2026-10-14T01:49Z [autonomy: Moderate] Press-ahead: BC-3 and BC-4 cut from feat/BC-2-backend-scaffold @5da9b80 (BC-2 at review). They merge master once BC-2 lands.
- 2026-10-14T01:00Z [operator ruling] **TypeScript pinned to 6.0.3**, not 7. TS 7.0.2 exports no classic compiler
  API (`exports["."]` is only `lib/version.cjs`) and typescript-eslint 8.71.2 peers `typescript >=4.8.4 <6.1.0`.
  Contract amended at ruling time (before the bootstrap commit): docs/SPEC.md §2 table + note, .lattice/context.md
  Versions line, BC-7 title, BC-7 comment.

## Run-time footguns

| Symptom | Cause | Mitigation |
|---|---|---|
| `c11 tree` prints "No windows" and panel-scoped calls (`mailbox recv`, `get-metadata`) time out after 10 s while `window.list`/`ping` work | Transient c11 server stall (02:04Z, cleared within ~3 min) | Do git/lattice work meanwhile; retry c11 calls next step; never treat an empty tree as "no agents" |
| Headless `lattice plan-review` takes 7–10 min and can die at the 600 s watchdog (`review-status`: dead holder pid, never recorded a result) | `claude -p` reviewing a large plan plus SPEC/BRANDING context runs past `review_timeout_seconds` (BC-7 01:44Z; BC-1/BC-2 finished in ~7 min) | Brief fallback works as written (fresh-context plan-review sub-agent). Keep plans proportional; do not raise the timeout mid-run |
| Delegator panel title becomes `BC-N [status]: <full ticket title>` | `lattice status` on a claimed (surface-bound) ticket renames the panel, as do plan-review/code-review (MV 01:24Z) | Brief now re-applies `rename-panel` + `set-title` after every status transition and review call |
| Delegate area vanishes after closing a finished delegator panel | Closing the last panel in an area collapses the area (01:20Z: areas 8 and 10 lost) | Every delegate area keeps a never-closed placeholder shell panel; check `c11 tree` before `close-panel` |
| `c11 launch-agent ... --json 2>&1 \| jq` fails to parse; zsh `set -- $t` loop launches nothing | launch-agent prints `warning[linked_worktree_cwd]` on stderr for worktree cwds; zsh does not word-split unquoted params | Send stderr to a file, parse stdout only; launch each delegator with an explicit command, never a zsh word-split loop; never test launch-agent with a dummy prompt (it launches a real agent) |
| A file the ticket needs is missing from `git status`/commits with no error | Host `core.excludesFile=~/.gitignore` applies to this repo and ignores `*.out *.log *.zip *.env *.d *.bak *.orig *.old *.tmp *.pid *.[7g]z *.[jrt]ar notes.md notes.txt node_modules` (reported by BC-1 delegator 01:11Z) | Test repo rules with `git -c core.excludesFile=/dev/null check-ignore -v <path>`; `git add -f` a file that must be tracked; `make dist` must not rely on host ignore rules (BC-17) |
