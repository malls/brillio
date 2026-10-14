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
- Worktree root: `/Users/forrest/Code/brillio-worktrees/<slug>` for BC-2..BC-13; **from BC-9 onward `/Users/forrest/Code/brillio-worktrees.noindex/<slug>`** (Spotlight skips `*.noindex` folders); branch `feat/BC-N-<slug>`.
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
| BC-3 | Domain ranking | done 02:45Z (merge dc35d1a, head cdec405) | stage sub-agents | feat/BC-3-domain-ranking | anchor feat/BC-2 @5da9b80 | opus |
| BC-4 | SQLite store | done 03:03Z (merge 1c23253, head 9daad85) | stage sub-agents | feat/BC-4-sqlite-store | anchor feat/BC-2 @5da9b80 | opus |
| BC-5 | Search API | done 03:56Z (merge 52c69dd, head 1871b01) | stage sub-agents | feat/BC-5-search-api | master + feat/BC-3 @ee09340 + feat/BC-4 @318ebc5 (both merged at creation) | opus |
| BC-6 | OpenAPI | done 04:44Z (merge 269f686, head 70518b5) | stage sub-agents | feat/BC-6-openapi | feat/BC-5 @75530c0 + master @0e6730f (review only after BC-5 lands) | opus |
| BC-7 | Frontend scaffold | done 02:27Z (merge aa65e61, head 0d1c2a2) | stage sub-agents | feat/BC-7-frontend-scaffold | master | opus |
| BC-8 | Frontend core | done 03:10Z (merge ccd1669, head 0426cb4) | stage sub-agents | feat/BC-8-frontend-core | anchor feat/BC-7 @1d6008a | opus |
| BC-9 | FilterBar | done 04:31Z (merge 9de440a, head 1b8a2a3) | stage sub-agents | feat/BC-9-filterbar | anchor feat/BC-8 @43e8824 | opus |
| BC-10 | Results surface | done 04:09Z (merge 5eaacdc, head 64bad4d) | stage sub-agents | feat/BC-10-results-surface | master @e5f49bb | opus |
| BC-11 | Pager/Status/Loading | done 03:41Z (merge 0abb105, head 0e70187) | stage sub-agents | feat/BC-11-pager-status-loading | master @e5f49bb | opus |
| BC-12 | SearchPage + e2e | done 05:00Z (merge b514f8e, head bb78854) | stage sub-agents | feat/BC-12-search-page | master @c2e5544 + feat/BC-9 @9d4925b (re-merged 1a3c5bb) + feat/BC-10 @fa7c08d (review only after both land) | opus |
| BC-13 | Compose stack | done 04:28Z (merge 4030996, head 103c57e) | stage sub-agents | feat/BC-13-compose-stack | anchor feat/BC-7 @1d6008a + master @7ab182d | opus |
| BC-14 | Branding pass | relaunched 04:56Z panel:37 (first delegator panel:32 killed by close-panel footgun) | stage sub-agents | feat/BC-14-branding-pass | anchor feat/BC-12 @5ac2397 | opus |
| BC-15 | README | done 05:33Z (merge f41fb8c, head 79408db) | stage sub-agents | feat/BC-15-readme | feat/BC-12 @5ac2397 + feat/BC-6 @70518b5 | opus |
| BC-16 | Extensions doc | done 05:43Z (merge 04214e9, head a0866dc) | stage sub-agents | feat/BC-16-extensions-doc | anchor feat/BC-12 @5ac2397 | opus |
| BC-17 | Release packaging | dispatched 05:18Z panel:38 | stage sub-agents | feat/BC-17-release-dist | master + feat/BC-15 @0b4e230 + feat/BC-16 @f4b6c77 (BC-14 merged when it lands) | opus |

Branch-base rule (press-ahead): all deps landed → base `master`. Exactly one dep unlanded and it is at
`review`/`in_validation` → base that dep's branch (with every landed dep already in master merged in).
Several unlanded deps, all at `review` or later and merging cleanly → master + all their branches, and the
child may not enter `review` until they all land (decision 02:31Z). Otherwise wait. Children record their anchor here and merge `master` (never rebase) once the anchor lands.

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
| Text-button style (BRANDING §5.8 / demo `.link`) | BC-11 recipe (relayed 03:26Z) | BC-9 (Clear), BC-10 (Clear filters, Try again) apply the same recipe in their own CSS modules; BC-14 checks the three match and may consolidate | no shared component mid-wave (would conflict across parallel branches) |
| `docs/SPEC.md` | contract | any ticket via SPEC's own deviation rule (edit in-branch + lattice comment) | union |

## Landing checklist additions

- [done 02:27Z] BC-7 landing: also add the ESLint 9.39.5 pin (jsx-a11y peer) to `.lattice/context.md` Versions, matching BC-7's SPEC §2 edit (MV 02:10Z).

## Final handoff checklist (open human items)

- BC-14 needs_human: 14 branding-owner questions O1-O14 (BRANDING §9 process; non-blocking; full list in BC-14 plan §4 and the needs_human reason). Must appear in the Phase 2 handoff (MV 05:40Z).
- Backlog tickets created during the run, for the operator: BC-18 (store hardening nits), BC-19 (text-button tracking/hover, branding owner), BC-20 (Unicode keyword vs SQLite ASCII lower()), BC-21 (a11y live feedback).

## Gate (make check) rule

`make check` green on the exact branch head (after merging current `master` into it) before any merge.
Exception (ENDED 02:27Z — both scaffolds landed; full `make check` now required for every merge), intake decision 1: until both BC-2 and BC-7 had landed, the half owned by the landing ticket must be
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
- 2026-10-14T05:43Z [autonomy: Moderate] BC-16 landed: orchestrator merged master (BC-15 README only, no overlap) → a0866dc; full make check pass; docs/EXTENSIONS.md 221 lines, no emoji; review art_01M4WECRZDRX1JMAQNPE1F7B5X PASS @ 3fc1aa4 (cycle 2, fable), validation art_01M4WESGCGK6X6QP4Q4DWYFFN5 (symbol greps). README link to docs/EXTENSIONS.md now resolves. Merge 04214e9, tree == branch 1bc9fb7, board sync f03ef56.
- 2026-10-14T05:40Z [operator ruling] Zip contents: SHIP CLAUDE.md and agents.md in the recruiter zip (honest disclosure of the AI workflow); SPEC §12 unchanged (it already does not exclude them). BC-17 needs-human cleared.
- 2026-10-14T05:33Z [autonomy: Moderate] BC-15 landed: merge f41fb8c (tree == branch fd79916), README-only diff (+242/-6, 245 lines), full make check pass on 79408db, no emoji/exclamation marks, review art_01M4WDX8JZJ51XARPZCPHZE0DC PASS @ 79408db (fable), validation art_01M4WE5VJ2SFN6E8PBJ7TG0JEZ (native README commands run; compose blocks cite BC-13 smoke per 04:58Z approval). README links docs/EXTENSIONS.md, live once BC-16 lands (before BC-17 zips). Board sync c8837b5. BC-17 mailed.
- 2026-10-14T05:23Z [autonomy: Moderate] Emoji check scope ruling (BC-14 question): the BRANDING §8 / BC-14 / BC-17 emoji grep covers shipped files only (git ls-files minus .lattice/ and the two PDFs); .lattice holds 724 U+2713 marks in vitest logs inside artifacts, never shipped. Validation-plan row 4 aligned.
- 2026-10-14T05:18Z [autonomy: Moderate] Press-ahead exception for BC-17: BC-14 is only at planned (rule needs review), but BC-17's work (Makefile dist block, scripts/verify-dist.sh) does not touch BC-14's frontend CSS, so BC-17 starts on master + BC-15 + BC-16 (both at review, merge clean). It may not review or record final verify-dist evidence until BC-14/15/16 all land, so the shipped zip is built from final master.
- 2026-10-14T05:00Z [autonomy: Moderate] BC-12 landed: merge b514f8e (tree == branch a6a6ef0), full make check (396 FE tests) + build pass on bb78854, review art_01M4WBW2080NK351Q4Q2JPYFJG PASS @ bb78854 (cycle 2, fable), browser validation art_01M4WC9P46N4142PWSVQ1JMF0J 7/7 PASS with 9 screenshots; worktree clean, no servers left on 18012/15112, browser panel closed. Board sync 0b24829. BC-14/15/16 released to review. BC-15 approved to skip a Compose run (cites BC-13 smoke; BC-17 re-proves from the zip).
- 2026-10-14T04:56Z [autonomy: Moderate] BC-14 delegator (panel:32) killed ~04:45Z by its own sub-agent running positional `c11 close-panel panel:35` (closes $C11_PANEL_ID). Recovery: orphan browser panel:35 closed with --panel, worktree verified clean at 5ac2397, plan still scaffold, no servers listening; replacement delegator launched (panel:37) with a replacement note. Brief Standing rules + footgun row added; BC-12/BC-15/BC-16 warned (BC-12 also nudged directly: its validator holds browser panel:36).
- 2026-10-14T04:44Z [autonomy: Moderate] BC-6 landed: merge 269f686 (tree == branch 69e183e), full make check + uncached httpapi race tests pass on 70518b5, review art_01M4WB799MXQFT7VPRPRF22E92 PASS @ 70518b5 (fable sub-agent; an earlier headless PASS at 5c4b993 superseded), validation art_01M4WBDAG3WZDYG9TAK2SNFFDD 8/8 on a live server + redocly 2.63.0 0 errors. No go.mod change. Board sync f0dd9c3. BC-15 mailed.
- 2026-10-14T04:40Z [autonomy: Moderate] Press-ahead final wave: BC-14 and BC-16 cut from feat/BC-12 @5ac2397 (BC-12 at review); BC-15 cut from feat/BC-12 + feat/BC-6 @70518b5 (both at review or later, merge clean). None may enter review until their anchors land. N=5 (BC-6, BC-12, BC-14, BC-15, BC-16). BC-17 waits for BC-14/15/16.
- 2026-10-14T04:31Z [autonomy: Moderate] BC-9 landed: orchestrator merged master (BC-13) → 1b8a2a3; overlap only docs/SPEC.md (BC-13 §3 one line vs BC-9 §7.2/§7.4), and the one test that reads SPEC (httpapi search_test §5.4 example) passed on the merged head, so gate-only; full make check (382 FE tests) + build pass on 1b8a2a3; review art_01M4WADP7WCHKFV2AKQQBQJFVG PASS @ 13ae060 (fable sub-agent); validation N/A-until-BC-12 art_01M4WAK2F08QGNZX8DJEBEAGSY. Cosmetic test-title finding ignored with justification (comment on BC-9). BC-9 created BC-21 (backlog: a11y live feedback). Merge 9de440a, tree == branch ffdc64f, board sync bf36a3d. BC-12 hold released.
- 2026-10-14T04:28Z [autonomy: Moderate] BC-13 landed: orchestrator merged master (BC-10: frontend components/format/test only, no overlap with BC-13 files) → 103c57e; full make check (269 tests) + sh -n + compose config pass on 103c57e; review art_01M4WA5W41DQ0SWCBEHKVY83QT PASS @ 21cd239 (cycle 3, fable sub-agent, no findings) accepted gate-only per refined rule; validation art_01M4WAG41TBY19R1TDE7SXYN1S + real make smoke 8/8 art_01M4WAG45G8GAP1ATBBVZHAVYC + HMR and search screenshots. Merge 4030996, tree == branch 17d8448, board sync d5e55d8. No brillio compose containers or volumes left behind.
- 2026-10-14T04:15Z [autonomy: Moderate] Approved a third BC-13 reviewer safety prompt (rm -rf of $H/stub-* under the reviewer scratchpad, set -u, docker stubbed). Watching BC-13 for review divergence: its final review is exercising smoke.sh signal handling in depth; any new Major there goes to the operator per the 2-cycle cap.
- 2026-10-14T04:09Z [autonomy: Moderate] BC-10 landed: orchestrator merged master (BC-5: backend/ + docs only, no overlap with BC-10 files or shared helpers) → 64bad4d; full make check (269 tests) + build pass on 64bad4d; review art_01M4W8WFSPMXX3TVGXFFKTQHWJ PASS @ fa7c08d accepted gate-only per refined rule 03:55Z; validation art_01M4W9CDTYQ3YAFHA6QT9VP6D4 (visual check 1100/375 + vitest); worktree verified clean of preview scratch files (also with global excludes off). Merge 5eaacdc, tree == branch 1c348b2, board sync 76d6a97. BC-12 mailed.
- 2026-10-14T04:02Z [autonomy: Moderate] Review routing changed again (MV 04:00Z: 5 watchdog deaths in ~15 headless runs, ~50 min lost, size not predictive): delegators stop calling headless `lattice plan-review`/`code-review` entirely. Every review is a fresh-context Agent-tool sub-agent with `model: "fable"` (keeps the Fable-reviews-Opus cross-model property the headless runs gave), same artifact shape and roles. Supersedes the 03:22Z skip rule.
- 2026-10-14T03:58Z [autonomy: Moderate] BC-5 landed: merge 52c69dd (tree == branch e4eb385), full make check + uncached httpapi race tests pass on 1871b01, review art_01M4W88NWW6DEGVRZ9Q9TH0GJT PASS @ 1871b01 (cycle 2; cycle-1 nits fixed), validation art_01M4W8H4WY1MP061Q8GXZ63JG0 42/42 PASS on a live server. SPEC edits landed: §5.3 parsing rules, §5.4 example 0.7414, §5.5 malformed_query, §13 note 5. BC-5 created BC-20 (backlog: Unicode keyword vs SQLite ASCII lower()). Board sync c2e5544. BC-6 and BC-13 released from their BC-5 holds. BC-12 launched on master + BC-9 + BC-10 (both in review, merge clean).
- 2026-10-14T03:55Z [autonomy: Moderate] Landing re-review rule refined (MV 03:49Z): "same top-level package/dir" means overlap with the branch's own files or folders (for the frontend: the same component folder) or with a shared helper the branch consumes (src/api, src/hooks, src/format, src/test/fixtures, styles, Makefile). A moved master that touches neither gets merge + exact-head gate only; any overlap gets a fresh exact-head review. Applies to BC-8 (logged 03:12Z), BC-9, BC-10 and later landings.
- 2026-10-14T03:54Z [autonomy: Moderate] Approved a second BC-13 reviewer safety prompt (shell trap test, no rm present; false positive).
- 2026-10-14T03:42Z [autonomy: Moderate] BC-11 landed: merge 0abb105 (tree == branch 9e65b43), full make check (228 FE tests) + build pass on 0e70187, review art_01M4W7R1433RHFSG03A1HCD8BT PASS @ 0e70187, validation N/A art_01M4W7T6W1CN66SDFA83DGMNDY. Board sync 0e6730f. BC-11 created BC-19 (backlog, low: text-button tracking/hover question for the branding owner). BC-6 launched in the freed slot, anchored on feat/BC-5 + master, review gated on BC-5 landing.
- 2026-10-14T03:37Z [autonomy: Moderate] Approved a BC-13 reviewer sub-agent safety prompt (harness `rm -rf` of a mktemp dir, docker stubbed) via send-key after reading the command.
- 2026-10-14T03:22Z [autonomy: Moderate] Review routing (MV 03:20Z: headless reviews died at the 600 s watchdog 3 times, ~10 min of critical path each): the 600 s watchdog stays; delegators skip the headless call and go straight to the fresh-context fallback reviewer when a plan exceeds 250 lines, a diff exceeds 1500 changed lines (excluding package-lock.json), or the 1-minute load average is above 30. Brief updated; applies from each delegator's next phase boundary.
- 2026-10-14T03:12Z [autonomy: Moderate] BC-8 landed: orchestrator merged master (BC-3/BC-4) into the branch → 0426cb4 (clean; overlap only docs/SPEC.md, different section); full make check (197 FE tests) + npm run build pass on 0426cb4; review art_01M4W5AWFDXPPGNHW06YWPYD3F PASS @ 43e8824 accepted gate-only per landing re-review rule (moved commits touch backend/ + SPEC §4.2, none of BC-8's files). Merge ccd1669, tree == branch 06f5c75, board sync e5f49bb. BC-10/BC-11 then based on master and launched (load back to ~9).
- 2026-10-14T03:04Z [autonomy: Moderate] BC-4 landed: merge 1c23253 (tree == branch c04b652), full make check + uncached store/httpapi race tests pass on 9daad85, only direct Go dep modernc.org/sqlite v1.60.1, review art_01M4W554XM9SHY2F8MP25KD1YD PASS @ 9daad85, validation addendum art_01M4W4XNCKYAQH20402VRMBQK6 @ 9daad85. Board sync 2745b8c. BC-4 created follow-up BC-18 (backlog, low, store hardening nits) — out of contract scope; left in backlog for the operator. Panel close timed out (c11 stall); BC-10 launch deferred while load ~90 and c11 unresponsive.
- 2026-10-14T03:00Z [autonomy: Moderate] Approved BC-5 contract addition: field-error code `malformed_query` on pseudo-field `query` (400, at most one per request) because Go's url.ParseQuery silently drops a malformed pair (bare `%`), which would widen results — the same failure §5.3's unknown_param rule prevents. BC-5 edits SPEC §5.3/§5.5 in-branch; BC-6 must include it in the OpenAPI code enum; BC-9 shows it via the unmapped-field line. Also: validation-plan row 18 corrected to 12 items (BC-5 report); BC-5 adds SPEC §13 note 5 (min/max copy vs equal-allowed) for the branding owner.
- 2026-10-14T02:45Z [autonomy: Moderate] BC-3 landed from verified board state while mail recv was wedged (LAND-READY confirmed afterwards in the c11 event log): merge dc35d1a, tree == branch 763965a, full make check + uncached domain race tests pass on cdec405, review art_01M4W4GP7NV84NM69TY4JRJ6HZ PASS @ cdec405, validation art_01M4W48RXGH4QGBPP4YB7F742W @ cdec405. Board sync c91fa7f. BC-5 mailed. c11 socket back 02:43Z but panel.get_metadata still times out.
- 2026-10-14T02:41Z [autonomy: Moderate] c11 socket down (connection refused; app pid 90026 alive). Orchestrator continues via lattice/git; no new launches or panel closes until c11 answers; operator notified (OS notification + Orchestrator reply).
- 2026-10-14T02:31Z [autonomy: Moderate] Press-ahead extension: BC-5 has two unlanded deps (BC-3 in_validation, BC-4 review), both past implementation and merging cleanly (`git merge-tree` rc 0). Branch-base rule extended: when every unlanded dep is at review or later and they merge cleanly, base on master + all their branches. BC-5 may plan/implement now but may not enter review until both have landed (card note 1).
- 2026-10-14T02:27Z [autonomy: Moderate] BC-7 landed: merge aa65e61 (tree == branch 63f0b0e), full make check + npm run build pass on 0d1c2a2 (orchestrator re-run), review art_01M4W3KK82ANSRJRSAXKGYGDWG PASS @ 0d1c2a2, validation art_01M4W345316DKA55J6NMBF5X4J (browser + screenshot) and addendum art_01M4W39TDPTX0KHPDEPP84NS3K. context.md ESLint 9.39.5 pin added (landing checklist). Board sync 83952b0. Half-gate exception ended. BC-8/BC-13 mailed to merge master.
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
| A delegator looks idle for 20+ min while its sub-agent "runs" | A sub-agent Bash call hit a Claude Code safety prompt despite bypass mode ("This shell -c script runs rm and could not be checked — Do you want to proceed?"); the dialog sits in the delegator panel (BC-13 03:16–03:37Z) | Each tick: `c11 input-state` on every delegator panel; a dialog state → read the screen, verify the command is safe (scratch paths only), approve with `send-key enter`, else answer No and mail the delegator |
| A delegator panel and process vanish mid-phase; its browser panel survives as an orphan "child of" the dead panel | A sub-agent ran `c11 close-panel panel:35` (positional). close-panel ignores positional refs and closes `$C11_PANEL_ID`, inherited from the delegator → closed the delegator itself (BC-14 ~04:45Z) | Always `c11 close-panel --panel <ref>`; brief Standing rules updated; all active delegators warned. Recovery: close the orphan browser panel with `--panel`, kill leftover servers, relaunch the delegator on the same worktree with a replacement note |
| An idle delegator never wakes for mail sent just after its turn ended | Doorbell outcome `buffered` (recipient busy at send time) and the later idle ring did not fire (BC-12 04:28–04:33Z) | Each tick, `c11 mailbox trace` every release/answer sent last tick; `buffered` + idle panel → nudge with `c11 send … && c11 send-key … enter` |
| A delegator sits idle waiting for an answer that was already sent | During a c11 stall the envelope is copied to the inbox but the stdin doorbell times out (`mailbox trace`: handler outcome timeout), so the agent never wakes (BC-5 waited 03:00–03:13Z) | After every stall, `c11 mailbox trace <id>` each answer sent during it; resend and nudge with `c11 send … && c11 send-key … enter` |
| Plain `docker compose up` fails on the operator's laptop (image platform arm64 vs requested amd64) | Operator `~/.zshrc` exports `DOCKER_DEFAULT_PLATFORM=linux/amd64`, `DOCKER_BUILDKIT=0`, `COMPOSE_DOCKER_CLI_BUILD=0` (BC-13 03:12Z) | `make up` / `scripts/smoke.sh` unset them; compose.yaml stays arch-neutral. BC-15 README makes `make up` the primary run command and notes the variable; BC-17 verify-dist runs with them unset. Raise with the operator: the live demo on this laptop must use `make up` or unset the variable |
| `agent.launch` "times out" but the agent starts anyway a few seconds later | c11 stall: the socket call times out while the app still executes the request (BC-9 02:53Z: `rc=1 timeout`, delegator bound its surface at 02:56Z) | After a launch timeout, never relaunch blind: check `lattice show <T>` events and `lsof -a -d cwd` on claude processes for the worktree before retrying |
| Load average spikes to 50–100 and c11 stalls each time a worktree is created or a build writes many files | Spotlight `mdworker_shared` ×16 indexing new worktrees, `node_modules`, build output (02:55Z: 93% CPU) | New worktrees go under `~/Code/brillio-worktrees.noindex/` (Spotlight skips `.noindex` folders); existing active worktrees stay put. Operator can also exclude `~/Code/brillio-worktrees` in Spotlight privacy |
| `c11 mailbox recv --ack` and `get-metadata` time out for every panel while `ping`, `mailbox send`, `close-panel` work | `panel.get_metadata` wedged in c11 under host overload (02:43Z) | Read incoming mail from `~/Library/Application Support/c11/events/events-com.stage11.c11-<c11 pid>.ndjson` (`mailbox.accepted` events whose `to` is the Orchestrator UUID); note handled IDs, the same envelopes arrive again when recv recovers. Land from verified lattice + git state (completion comment + fresh PASS + validation naming HEAD) without waiting for the mail |
| c11 socket refuses connections (errno 61) while the c11 app process is alive; all c11 commands fail | Host overload 02:38Z: load average 90–100 (Spotlight mdworker ×15 indexing the new worktrees + macOS daemons), swap 7.2/8 GB | Keep working through git + lattice (landing does not need c11); defer panel closes and launches; OS notification to operator; do not quit c11 while agents run (kills every agent PTY). Suggest excluding ~/Code/brillio-worktrees from Spotlight |
| `c11 launch-agent` exits 1 with `timeout: method=agent.launch`; `c11 tree` times out or prints "No windows" | Host load average 65+ from macOS daemons (AppleSpell, opendirectoryd, mobileassetd, modelcatalogd; 02:32Z), not from the build | Check `c11 rpc panel.list` (still answers) before retrying a launch — the timed-out launch created no panel; retry once load eases. Never double-launch |
| `c11 tree` prints "No windows" and panel-scoped calls (`mailbox recv`, `get-metadata`) time out after 10 s while `window.list`/`ping` work | Transient c11 server stall (02:04Z, cleared within ~3 min) | Do git/lattice work meanwhile; retry c11 calls next step; never treat an empty tree as "no agents" |
| Headless `lattice plan-review` takes 7–10 min and can die at the 600 s watchdog (`review-status`: dead holder pid, never recorded a result) | `claude -p` reviewing a large plan plus SPEC/BRANDING context runs past `review_timeout_seconds` (BC-7 01:44Z; BC-1/BC-2 finished in ~7 min) | Brief fallback works as written (fresh-context plan-review sub-agent). Keep plans proportional; do not raise the timeout mid-run |
| Delegator panel title becomes `BC-N [status]: <full ticket title>` | `lattice status` on a claimed (surface-bound) ticket renames the panel, as do plan-review/code-review (MV 01:24Z) | Brief now re-applies `rename-panel` + `set-title` after every status transition and review call |
| Delegate area vanishes after closing a finished delegator panel | Closing the last panel in an area collapses the area (01:20Z: areas 8 and 10 lost) | Every delegate area keeps a never-closed placeholder shell panel; check `c11 tree` before `close-panel` |
| `c11 launch-agent ... --json 2>&1 \| jq` fails to parse; zsh `set -- $t` loop launches nothing | launch-agent prints `warning[linked_worktree_cwd]` on stderr for worktree cwds; zsh does not word-split unquoted params | Send stderr to a file, parse stdout only; launch each delegator with an explicit command, never a zsh word-split loop; never test launch-agent with a dummy prompt (it launches a real agent) |
| A file the ticket needs is missing from `git status`/commits with no error | Host `core.excludesFile=~/.gitignore` applies to this repo and ignores `*.out *.log *.zip *.env *.d *.bak *.orig *.old *.tmp *.pid *.[7g]z *.[jrt]ar notes.md notes.txt node_modules` (reported by BC-1 delegator 01:11Z) | Test repo rules with `git -c core.excludesFile=/dev/null check-ignore -v <path>`; `git add -f` a file that must be tracked; `make dist` must not rely on host ignore rules (BC-17) |
