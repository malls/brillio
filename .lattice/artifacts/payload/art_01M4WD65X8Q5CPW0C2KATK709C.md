Verdict: PASS-WITH-NITS
Reviewed HEAD: c374ce86ddd2d640dbef2c779541ee577e5c1575
Reviewer: fable sub-agent (decision 04:02Z)

## Findings
### Critical
none
### Major
none
### Minor
none
### Nit
- [N1] README.md:86 — "one WHERE clause per filter" is loose for `keyword`: the store emits one `instr(lower(description), ?) > 0` clause per token (backend/internal/store/listings.go:35, SPEC §6.1), so a two-word keyword is two clauses. — Recommendation: "one WHERE clause per filter (one per keyword token)".
- [N2] README.md:119 — "appliedFilters: the filters as the server read them, null when absent" is true for five of six keys; `status` is never null, it echoes the effective set ["active","pending","sold"] when absent (backend/internal/httpapi/search.go:120, 186-204; observed live). Mirrors SPEC §5.4 wording; precision nit. — Recommendation: "null when absent; `status` echoes the effective set".
- [N3] README.md:39 — "Node 24, 24.15 or later (jsdom 30 requires it)" reads as a hard floor at Node 24; jsdom 30.1.2 engines are `^22.22.2 || ^24.15.0 || >=26.0.0` and Vitest 5.0.4 accepts `^22.12.0`. Node 24 is the chosen line (node:24-alpine, SPEC §2). — Recommendation: optional, "Node 24 (24.15 or later, which jsdom 30 needs)" or leave as is.

## Checks run
- HEAD c374ce8; diff master...HEAD = README.md only (+242/-6).
- Quote: awk leading comment of scoring.go vs README.md:164-177 identical, 14 lines.
- grep -c '!' = 0; emoji perl grep empty; wc -l = 245; voice grep empty; sentence-case headings.
- Constants/formulas vs code: 0.6/0.4/30 (scoring.go:28-34), literal factor 2 (scoring.go:53), MaxPageSize 100 (paginate.go:10), defaults 1/10 (query.go:67-68), tie order (rank.go:74-87), round4 (search.go:208), UI toFixed(2). Worked values recomputed, 0.7414 @2026-10-13 and 0.7382 @2026-10-14.
- Error codes: 4 envelope (respond.go:14-17) + 10 field codes (query.go:51-60), malformed_query on `query`.
- Endpoints (router.go:62-65), healthz 503 degraded, cities sorted, openapi 3.1.1, X-Request-ID.
- Makefile/compose/smoke.sh by reading: targets, up unsets DOCKER_DEFAULT_PLATFORM only, down -v, start_period 180s, smoke preflight/8 checks/down -v trap. Platform wording matches compose.yaml:16-21.
- Versions vs go.mod/package.json/lockfile and local toolchain: all match.
- Config defaults vs config.go:28-33 match. Layout tree vs git ls-files: all exist except intended forward link docs/EXTENSIONS.md (BC-16).
- Duplicate pairs vs data/sample_listings.json: A1/B7, A2/B8, A3/B9, A5/B11 confirmed.
- Native runs on 18015 (:memory:): healthz, cities, Springfield budget search, §5.5 error example (body diffs clean), %zz, page=99, Nowhere, city=, no-budget order, 404, 405, X-Request-Id echo — all as README states. Processes killed by PID.
- make check rc=0 (5 Go packages ok, Vitest 23 files / 396 tests).
- Docker blocks not executed per scope; judged by reading.

## Summary
Every factual claim checks out against the merged code, the live API, the lockfile and the toolchain. All acceptance checks pass. Three wording-precision nits; none needs a rework cycle.