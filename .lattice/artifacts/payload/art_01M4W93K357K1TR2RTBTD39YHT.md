Code review BC-13 cycle 2 (own-reviewer fallback, fresh-context Opus, agent:opus-5.5-reviewer)

Verdict: PASS-WITH-NITS
Reviewed HEAD: 95b937f7cc5438a890652ba48aae306dc449cb0f
Gate: make check exit 0 (golangci-lint 0 issues; eslint, tsc -b, prettier clean; go test -race 6 packages ok, 43 top-level tests; vitest 13 files / 228 tests); sh -n and dash -n ok; shellcheck 0.11.0 clean; docker compose config -q ok with and without DOCKER_DEFAULT_PLATFORM; make -n up down smoke ok; TAB recipes; smoke.sh 100755, backend/.dockerignore 100644; tree clean with excludesFile=/dev/null; no .lattice/ in branch commits. Live make smoke (25 s): a1, a6, a7 PASS; a2-a5, a8 FAIL 404 not_found (expected pre-BC-5); diagnostics printed; teardown ran; exit 1. Afterwards 0 brillio containers/volumes/networks, ports free, backend/tmp removed, tree clean.

Cycle-1 findings:
1. curl proxy: fixed — all three curl sites (smoke.sh:43,79,102) start with -q --noproxy '*'; real curl 8.7.1 with .curlrc/http_proxy/ALL_PROXY to a non-routable host: probe rc 7 with the flags (28 without).
2. HUP: fixed — trap 'exit 129' HUP (:64); cleanup ignores INT TERM HUP (:53); stubbed HUP/INT/TERM exit 129/130/143 with teardown under sh and dash.
3. printf: fixed — bad() (:121) and wait_for (:82) use printf; dash prints escapes literally.
4. .dockerignore: fixed — `*`; classic builder sends 796 B, finds Dockerfile.dev and builds; buildx --check ok.

Findings:
- [Minor] scripts/smoke.sh:50-55 — cleanup runs under set -e; when stdout is dead (terminal closed then SIGHUP, or Ctrl-C on `smoke.sh 2>&1 | tee` killing tee -> EPIPE) the first echo in cleanup fails and errexit ends cleanup before `docker compose down -v`. Reproduced under /bin/sh and dash with a stub docker. Recommendation: `set +e` right after `status=$?` in cleanup and `trap '' INT TERM HUP PIPE` at :53 (verified on a scratch copy: down runs, status 129/130 kept).

Acceptance:
- Clean checkout docker compose up -> healthy api, web :5173 proxying: met (live run; fresh-clone proof is validation stage 1).
- make smoke passes and exits 0: met-pending-BC-5 (3/8 live today, expected 404s; stubbed 8/8 exits 0).
- make down leaves no containers or volumes: met.
- Go and TSX edits reload without restarting compose: met by configuration; live proof is validation stage 1.
- Validation evidence: met-pending-BC-5.

Summary: All four cycle-1 findings fixed and verified; branch still matches the plan, its resolutions and SPEC §3/§10/§11. One new non-blocking Minor (teardown skipped when stdout is dead), a two-token fix. Aside: plan Implementation Notes line "no .dockerignore needed" is stale.