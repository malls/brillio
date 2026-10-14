Code review BC-13 cycle 3 (final) — reviewer: fable sub-agent (decision 04:02Z), actor agent:opus-5.5-reviewer

Verdict: PASS
Reviewed HEAD: 21cd239cc4ead693bc80d0411eae00a0473f4422
Gate: make check exit 0 (golangci-lint 0 issues; go test 5 pkgs ok, uncached -race re-run 65 subtests PASS / 0 FAIL; vitest 13 files / 228 tests; eslint --max-warnings 0 + typecheck + format:check clean). sh -n OK; dash -n OK; shellcheck 0.11.0 -s sh clean; env -u DOCKER_DEFAULT_PLATFORM docker compose config -q exit 0 (Compose v5.5.1 / Docker 29.8.1); make -n up down smoke prints `unset DOCKER_DEFAULT_PLATFORM; docker compose up --build` / `docker compose down -v` / `sh scripts/smoke.sh`; scripts/smoke.sh 100755, backend/.dockerignore 100644; status --porcelain (excludesFile=/dev/null) empty before and after. Live make smoke (ports free, no brillio project beforehand): fresh brillio_gomod/brillio_gobuild volumes, api Healthy then web Healthy, PASS a1-a8, `smoke: 8 passed, 0 failed`, exit 0 in 24 s; afterwards brillio-labelled containers/volumes/networks empty, both ports free, tree clean. No branch-only commit touches .lattice/; merge-base diff contains only the 7 BC-13 files.

Cycle-2 finding: fixed. scripts/smoke.sh:52 `set +e` after `status=$?`; :54 `trap '' INT TERM HUP PIPE`; :68 `trap 'exit 141' PIPE` set after preflight (preflight :29-45 runs with no trap). Stub harness: at 95b937f `smoke.sh | head -c 1` and `2>&1 | head -c 1` exit 141 with down never called; at HEAD both exit 141 and `compose down -v` runs. SIGINT/SIGTERM/SIGHUP during compose up (and SIGINT mid-a2) exit 130/143/129 with teardown.

Findings:
none

Acceptance:
- Clean-checkout docker compose up -> healthy api + web on :5173 proxying to api: met (live run from fresh volumes; a6 title, a7 /healthz and a8 /api/cities through the Vite proxy PASS; judged with DOCKER_DEFAULT_PLATFORM unset per the recorded deviation).
- make smoke passes / exit 0 and make down leaves nothing: met (8/8, exit 0; down -v left zero brillio containers/volumes/networks, ports free).
- Go and TSX edits reload without restarting compose: supported by config (.air.toml poll=true/500 ms, entrypoint ./tmp/api, keys verified against air v1.67.4; CHOKIDAR_USEPOLLING=true); live proof is the validation stage's job.
- Validation evidence --role validation: pending validation stage.

Summary: Full branch diff at 21cd239 matches the plan's authoritative blocks and binding decisions. With BC-5 merged every smoke assertion is live and matches the handlers (min_gt_max in the validation_error envelope; cities {"cities":[...]}; rank first field so a4 extraction is field-order-safe). Cycle-2 teardown bug fixed on both the SIGPIPE and EPIPE/errexit paths. Non-blocking notes: SPEC §3 does not list backend/.dockerignore (optional doc touch-up); master has since gained BC-10 (no overlap).