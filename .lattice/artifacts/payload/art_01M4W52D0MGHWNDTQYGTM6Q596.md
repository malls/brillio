Plan review BC-13 (fresh-context fallback reviewer, agent:opus-5.5-plan-reviewer)

Verdict: APPROVE-WITH-CHANGES

Findings:
1. [Minor] Teardown failure can exit 0 — smoke.sh step 2 (Trap). cleanup warns on `docker compose down -v` failure but keeps the saved status, so a passing run with a failed teardown exits 0 and leaves brillio-* containers/volumes. Recommendation: if saved status is 0 and down -v fails, exit 1; add `trap '' INT TERM` at the top of cleanup so a second Ctrl-C cannot abort teardown.
2. [Minor] Container logs lost on assertion failure — smoke.sh steps 3-5. Logs print only on bring-up failure; the EXIT trap's down -v destroys them. Recommendation: when M > 0, print `docker compose logs --no-color --tail=60` before exiting; `rm -f "$tmp/body"` before each fetch so a connection failure does not quote a stale body.
3. [Minor] Stage 1 step 1 does not force a cold build — `docker image rm brillio-api` leaves the BuildKit layer cache. Recommendation: `docker compose build --no-cache api` (project-scoped) or drop the wording.
4. [Minor] Browser-search fallback is an unrecorded acceptance deviation — stage 2 step 3. Recommendation: if the JSON-through-proxy fallback is used (BC-12 not on master), record it in the validation evidence and a lattice comment.
5. [Minor] SQLite WAL on the bind mount discovered late — Risks. BC-4 uses `_pragma=journal_mode(WAL)`. Recommendation (optional): one-off alpine sqlite3 WAL probe against ./data in stage 1, then delete the probe files.
6. [Minor] Every make smoke / make up after make down starts cold — consequence of binding decision 4 (down -v wipes gomod/gobuild). Recommendation: keep; note the cost for the orchestrator; follow-up only if it hurts.
7. [Nit] a6 must grep the raw body — compact() turns `Brillio Listings` into `BrillioListings`. State that a2, a6, a8 grep "$tmp/body" directly.
8. [Nit] Plan says BC-7 still at review — it landed (aa65e61) and master is merged in (53f7b2e). Update header and Risks.
9. [Nit] Volume diff can flag other projects' volumes — stage 1 step 6. Treat an extra volume as failure only if labelled com.docker.compose.project=brillio.
10. [Nit] Stage 2 step 2 uses plain `git status` — use `git -c core.excludesFile=/dev/null status --porcelain`.

Verified, no change needed: air v1.67.4 newest; all chosen air keys exist (poll, poll_interval>=500, send_interrupt, delay, log, exclude_regex, entrypoint, misc.clean_on_exit, log.time, screen.clear_on_rebuild); kill_delay "2s" parses; stop_on_error=false restarts the old binary; filenotify poller tracks dir entries; Vite 8.3.4 bundles chokidar 3.6.0 honoring CHOKIDAR_USEPOLLING/INTERVAL; loadEnv copies VITE_* from process env; allowedHosts admits localhost/IP; lockfile has linux-arm64-musl rolldown/lightningcss; Compose v5.5.1 supports --wait/--wait-timeout/--quiet-pull; ports 8080/5173 free; none of the new paths ignored.

Summary: Technically sound; air config keys, Vite/chokidar polling, proxy env, node_modules volume and health gating checked against installed sources. Smoke design is correct POSIX and maps every ticket and SPEC §10 assertion. No blocking finding; fix 1-3 and 7 before in_progress, record 4 and 6 as comments, 5 is an optional early de-risk.