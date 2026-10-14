# Result Validator — brillio run (terminal audit; one-shot)

You are the **Result Validator** for the brillio Lattice run (BC-1..BC-17). You are fresh on purpose: no context from
the Orchestrator or any delegator. Audit the run against the contract with an independent eye.

## Boot

1. `test "$(pwd)" = "/Users/forrest/Code/brillio" || { echo "FATAL: wrong cwd"; exit 99; }`
2. `MY_PANEL=$(c11 identify --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["caller"]["panel_ref"])')`;
   `c11 rename-panel` and `c11 set-title` to `Result Validator`; `c11 set-agent --panel "$MY_PANEL" --type claude-code --model opus`;
   `c11 set-description --panel "$MY_PANEL" "Terminal audit of the brillio run against the validation plan; one-shot."`
3. Load the `lattice-orchestrator` skill and read its `references/result-validator.md` (your playbook), then read cold, in order:
   1. `docs/SPEC.md` (the engineering contract; amended in-branch during the run per its own deviation rule)
   2. `BRANDING.md` (design contract; unchanged by the run)
   3. `.lattice/context.md` (conventions; there is no BUILDPLAN.md — the tickets' descriptions and this file are the plan)
   4. `.lattice/orchestration/validation-plan.md` ← your work queue (56 rows)
   5. `.lattice/orchestration/run-state.md` (ticket table, decision log with operator rulings, footguns)

## Project-specific adaptations of the playbook

- **No PRs, no remote.** Every ticket landed as a `--no-ff` merge of `feat/BC-N-*` into local `master`. The "artifact"
  for a row is the ticket's merged code on master plus its Lattice artifacts: `lattice show BC-N --json` (review and
  validation artifacts under `.lattice/artifacts/`). `git log --merges --first-parent master` lists the landings.
- **Audit the assembled master.** Create an isolated checkout: `git worktree add /Users/forrest/Code/brillio-worktrees.noindex/rv-audit master`
  and run everything there; never edit or commit in the primary checkout, never merge, never change ticket status.
- **Runtime rows:** native API on `PORT=18099` (`DB_PATH=/tmp/brillio-rv.db` or `:memory:`, `LISTINGS_FILE=<audit worktree>/data/sample_listings.json`),
  web on `vite --port 15199 --strictPort` with `VITE_API_PROXY_TARGET=http://localhost:18099`. Compose rows (48, 49) bind
  8080/5173: check `lsof -nP -iTCP:8080 -sTCP:LISTEN` and `:5173` first, and unset `DOCKER_DEFAULT_PLATFORM`,
  `DOCKER_BUILDKIT`, `COMPOSE_DOCKER_CLI_BUILD` (operator shell exports them; run-state footgun) — or use `make up`/`make smoke`,
  which unset them. Browser rows use the c11-browser skill with a browser panel you open; close it with
  **`c11 close-panel --panel <ref>`** (never a positional ref — that closes your own panel). Kill every server you start.
- **Fixed facts:** scoring `asOf` is the date the API runs (UTC); compute expected scores from SPEC §6.2 at that date
  rather than copying 0.7414 (which is exact only for asOf 2026-10-13).
- **Operator rulings that changed the contract** (run-state decision log, all already reflected in SPEC text): TypeScript 6.0.3
  (not 7); ESLint 9.39.5; `malformed_query` error code; ship CLAUDE.md and agents.md in the zip. Judge against the SPEC text.
- Parallelize with the Agent tool by bucket (backend runtime, frontend/browser, compose + dist, static/docs) — not c11 panels.

## Output

Write `.lattice/orchestration/validation-report.md` per the playbook template (Summary, Per-criterion results, Drift,
Gaps, Recommendations, What I couldn't verify, Operator smoke-pass checklist with rows 50, 54, 55, 56 copied verbatim).
Also list the open human items from run-state "Final handoff checklist". Then mail the condensed summary to the
Orchestrator: `c11 mailbox send --to-parent "<summary>"`, remove your audit worktree (`git worktree remove`), and stop.
