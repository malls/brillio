Code review BC-13 (own-reviewer fallback, fresh-context Opus, agent:opus-5.5-reviewer)

Verdict: PASS-WITH-NITS
Reviewed HEAD: bfdd2063583f22c7cda58e38e87a3595935233bc
Gate: make check exit 0 (golangci-lint 0 issues; go test -race 6 packages ok, 43 top-level tests; vitest 2 files / 4 tests); sh -n + dash -n ok; shellcheck 0.11.0 clean; docker compose config -q ok (with and without DOCKER_DEFAULT_PLATFORM); make -n up down smoke prints the three recipes; TAB recipes; smoke.sh 100755; tree clean with excludesFile=/dev/null.
Live make smoke (26.5 s, ports free, no brillio project beforehand): a1, a6, a7 PASS; a2-a5, a8 FAIL with 404 not_found (expected pre-BC-5); compose ps + logs dumped on failure; teardown ran; exit 1. Afterwards 0 brillio containers/volumes/networks, ports free, backend/tmp removed, git clean. air v1.67.4 on go1.27.2; modernc sqlite downloaded into the gomod volume and built in ~15 s; api seeded 12 listings into /data/listings.db (WAL on the bind mount works); listening [::]:8080.
Stub harness (fake docker/curl, dash and macOS sh): all-pass exit 0; teardown failure -> exit 1 with FAIL teardown; bring-up failure -> exit 1; SIGINT -> 130 with teardown; SIGTERM -> 143 with teardown; DOCKER_DEFAULT_PLATFORM unset for every docker call; a4 rank extraction works with rank after the nested relevance object.

Findings:
- [Minor] scripts/smoke.sh:41,76,99 — curl honours http_proxy/~/.curlrc; with a proxy and no localhost in no_proxy the port probe gets exit 28 not 7 (false "port in use") and assertions go through the proxy. Recommendation: add -q (first arg) and --noproxy '*' to all three curl calls.
- [Nit] scripts/smoke.sh:51,61-63 — SIGHUP not trapped; under dash an untrapped HUP skips the EXIT trap, leaving the stack running. Recommendation: trap 'exit 129' HUP; trap '' INT TERM HUP in cleanup.
- [Nit] scripts/smoke.sh:79,116 — echo prints the body excerpt; dash echo interprets backslash escapes. Recommendation: printf 'FAIL  %s (HTTP %s: %s)\n'.
- [Nit] backend/ has no .dockerignore — with DOCKER_BUILDKIT=0 the classic builder sends all of backend/ (grows with backend/bin/api). Recommendation: backend/.dockerignore containing *.

Acceptance:
- Clean checkout docker compose up -> healthy api, web on :5173 proxying: met (with the accepted env -u DOCKER_DEFAULT_PLATFORM workaround); fresh-clone proof is validation stage 1.
- make smoke passes and exits 0: met-pending-BC-5 (3/8 today, the 5 failures are the expected 404s; stubbed 8/8 exits 0).
- make down leaves no containers or volumes: met.
- Go and TSX edits reload without restarting compose: met by configuration; live edit proof is validation stage 1 steps 4-5.
- Validation evidence: met-pending-BC-5 (validator's job).

Summary: Follows the plan and its authoritative resolutions; Makefile BC-13 block only; compose matches SPEC §10/§11; SPEC §3 edit accurate. Smoke preflight/trap/exit-status logic held up live and under a stub harness. The DOCKER_DEFAULT_PLATFORM unset is correct and scoped. Four small mechanical hardening fixes; none blocks.