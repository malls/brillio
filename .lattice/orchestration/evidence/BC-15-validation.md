# BC-15 validation: README run instructions, architecture, scoring rationale and trade-offs

**N/A (documentation): every native README command block ran as written at 79408db with matching output (make check/lint/test and all per-half targets exit 0, dev-api and dev-web served and proxied, all four curl blocks byte- or semantically identical to the README, A1 scored 0.7382 on 2026-10-14 exactly as README line 150 states); Compose blocks rest on BC-13's make smoke 8/8 (art_01M4WAG45G8GAP1ATBBVZHAVYC), with Makefile/compose.yaml/scripts/Dockerfile.dev/.air.toml/vite.config.ts byte-identical to the smoked sha 21cd239, re-proved by BC-17 from the zip.**

- **Ticket:** BC-15
- **Branch / HEAD:** `feat/BC-15-readme` @ `79408dba42fa254cac14c363cc32a4021988c179` (checked with `git rev-parse HEAD` before and after)
- **Worktree:** `/Users/forrest/Code/brillio-worktrees.noindex/bc-15-readme`
- **Validator:** agent:opus-5.5-validator
- **Date:** 2026-10-14, run 05:25-05:30 UTC (01:25-01:30 EDT local; both the same calendar day)
- **Host toolchain:** Go 1.27.1, Node 24.21.0, npm 11.19.0, golangci-lint 2.14.0, GNU Make 3.81 (all match the README tool table, lines 35-41)
- **Overall verdict: PASS**

## Port and invocation adaptations

- 8080 -> 18015 (`PORT=18015 DB_PATH=:memory:`), 5173 -> 15115, per the delegator brief.
- `make dev-web` cannot take a port: its recipe is `cd frontend && npm run dev` (`make -n dev-web` printed only that line, so the `npm ci` sentinel was already current) and `vite.config.ts` sets no `server.port`. The recipe was therefore run directly as `cd frontend && VITE_API_PROXY_TARGET=http://localhost:18015 npm run dev -- --port 15115 --strictPort`.
- Docker / Compose blocks were not run (orchestrator ruling, mail 01M4WCCZ00BY3BHMX0H1DE7ZMT); see "Compose blocks" below.

## Command blocks in README.md

| README lines | Block | Class | How run | Exit | Observed vs README | Result |
|---|---|---|---|---|---|---|
| 7-9 | `make up` | Docker | Not run; BC-13 evidence | n/a | Makefile `up:` is `unset DOCKER_DEFAULT_PLATFORM; docker compose up --build`, as line 13 says | PASS (cited) |
| 15-17 | `env -u DOCKER_DEFAULT_PLATFORM docker compose up` | Docker | Not run; BC-13 ran this exact prefix (its steps 1-3, 9) | n/a | BC-13: healthy api + web on :5173 proxying to api | PASS (cited) |
| 22 | `make dev-api` | native | `PORT=18015 DB_PATH=:memory: make dev-api` (background) | 0 (clean exit on SIGTERM) | Ran `cd backend && go build -o bin/api ./cmd/api` then `./backend/bin/api`; log `seeded 12 listings`, `listening addr=[::]:18015`, `cors_origins=["http://localhost:5173"]`, `shutdown_timeout=10s`, `log_level=INFO`, `listings_file=./data/sample_listings.json` (matches line 26 defaults); on stop: `shutting down`, `server stopped` | PASS |
| 23 | `make dev-web` | native (adapted) | `cd frontend && VITE_API_PROXY_TARGET=http://localhost:18015 npm run dev -- --port 15115 --strictPort` | 143 (SIGTERM, expected) | `VITE v8.3.4 ready`, `Local: http://localhost:15115/`; `/` 200 with `<title>Brillio Listings</title>`; via :15115 `/healthz` -> `{"status":"ok"}` 200, `/api/cities` -> six cities 200, Springfield search -> A1 0.7382, B7 0.7266, A2 0.5516, B8 0.5348; API log shows the three proxied requests | PASS |
| 29 | `make check` | native | as written | 0 | golangci-lint `0 issues.`; eslint, `tsc -b --noEmit`, prettier `All matched files use Prettier code style!`; `go test -race ./...` 5 packages ok; vitest `Test Files 23 passed (23)`, `Tests 396 passed (396)`; 11.0 s | PASS |
| 30 | `make smoke` | Docker | Not run; BC-13 evidence | n/a | BC-13: `smoke: 8 passed, 0 failed`, exit 0 | PASS (cited) |
| 33 (prose) | plain `make`, `make lint`, `make test`, `make lint-api`, `make lint-web` | native | as written | 0, 0, 0, 0, 0 | Same key lines as `make check`, each limited to its half/kind; plain `make` runs lint then test for both halves | PASS |
| 50-79 | project layout tree | illustrative | every path tested with `[ -e ]` | n/a | All paths exist except `docs/EXTENSIONS.md` (see note 1); component, hook, format and api dirs hold exactly the names listed | PASS (with note 1) |
| 106-107 | `curl -s http://localhost:8080/healthz` | native | port 18015 | 0 | `{"status":"ok"}`, byte-identical to line 107 | PASS |
| 108-109 | `curl -s http://localhost:8080/api/cities` | native | port 18015 | 0 | `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}`, byte-identical to line 109 | PASS |
| 110 | `curl -s '...search?targetBudget=450000&city=Springfield'` | native | port 18015 | 0 | 4 items ranked A1, B7, A2, B8 (line 113); scores 0.7382, 0.7266, 0.5516, 0.5348; `scoring.asOf` `2026-10-14` | PASS |
| 115-120 | response shape | illustrative | checked against the line-110 body | n/a | Top-level keys exactly items, page, scoring, appliedFilters; item has rank, listing fields, relevance {score, budgetFit, recency}; page {page, pageSize, totalItems, totalPages}; scoring {targetBudget, budgetWeight, recencyWeight, recencyHalfLifeDays, asOf}; appliedFilters nulls for absent filters, status `["active","pending","sold"]` | PASS |
| 124-130 | error envelope | illustrative | `curl -s '...search?minPrice=5&maxPrice=1'` | 0 | `{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"maxPrice","code":"min_gt_max","message":"Min price must be less than max price."}]}}`, same content as lines 125-129 | PASS |
| 138-144 | scoring formula | illustrative | constants grep + arithmetic | n/a | See acceptance checks; 0.5^(45/30) = 0.3536, score 0.7414; 0.5^(46/30) = 0.3455, score 0.7382 | PASS |
| 163-178 | scoring.go header quote | acceptance | diff (below) | 0 | identical | PASS |
| 185 | `curl -s '...search?minPrice=5&maxPrice=1&pageSize=0&minprice=1'` | native | port 18015 | 0 (HTTP 400) | Raw body parsed equals the JSON at lines 189-199 (`a == b` True); README JSON re-serialized compactly is byte-identical to the raw body (key order and field order included) | PASS |
| 226 | `make test` | native | as written | 0 | go 5 packages ok; vitest 23 files / 396 tests passed | PASS |
| 227 | `make test-api` | native | as written | 0 | `cd backend && go test -race ./...`, 5 packages ok (cached); an extra uncached `go test -race -count=1 ./...` also passed: config 1.217s, domain 1.231s, httpapi 1.643s, store 1.546s, store/seed 1.176s | PASS |
| 228 | `make test-web` | native | as written | 0 | `cd frontend && npm test` -> vitest v5.0.4, 23 files / 396 tests passed | PASS |

### Dated score example (README line 150)

The run scored against `scoring.asOf = 2026-10-14`. A1 ($450,000, listed 2026-08-29) is 46 days old: budgetFit 1, recency 0.3455, score **0.7382**, exactly the README's "the same request on 2026-10-14 gives 0.7382". The 2026-10-13 figures (recency 0.3536, score 0.7414) are asserted with the clock pinned at 2026-10-13 in `backend/internal/httpapi/search_test.go:407,425,469`, which passed in the uncached run above.

### Prose spot-checks against the running API (extra, not command blocks)

| README claim | Observed |
|---|---|
| line 101: `GET /api/openapi.yaml` serves the document | 200 `application/yaml` |
| line 103: `X-Request-ID` echoes the client's or a new one | sent `bc15-validate` -> `X-Request-Id: bc15-validate`; none sent -> `X-Request-Id: d06242b92cdf6f744e41304cae522342` |
| line 132: 404 / 405 codes | `/api/nope` -> 404 `not_found`; `POST /api/cities` -> 405 `method_not_allowed` |
| line 203: `?city=%zz` -> `malformed_query` on `query` | 400, `{"field":"query","code":"malformed_query",...}` |
| line 205: page past the end is 200 with real counts | `page=99` -> 200, 0 items, `totalItems 12, totalPages 2` |
| line 206: city with no listings is 200, `totalItems: 0`; `city=` is absent | `city=Nowhere` -> totalItems 0; `city=` -> appliedFilters.city null, 12 items |
| lines 157, 161: no budget = recency only, weights 0 and 1, newest first | scoring `{targetBudget: null, budgetWeight: 0, recencyWeight: 1, recencyHalfLifeDays: 30}`; listedDate descending A5 2026-09-04 ... B11 2026-08-15 |
| line 43: pinned versions | go.mod `modernc.org/sqlite v1.60.1`; package.json react 19.3.0, vite 8.3.4, typescript 6.0.3, vitest 5.0.4, eslint 9.39.5, prettier 3.9.10, jsdom 30.1.2; `FROM golang:1.27-alpine`, `AIR_VERSION=v1.67.4`, `image: node:24-alpine` |
| line 11: health checks allow 180 seconds | compose.yaml `start_period: 180s` on api and web |
| line 221: seams | `func BudgetFit` (scoring.go:49), `byRelevance` (rank.go:67,71), dedupe slot comment (search.go:37) |

## Acceptance checks

```
$ diff <(awk 'NF==0{exit} {print}' backend/internal/domain/scoring.go) <(sed -n '164,177p' README.md)
(no output)                                   diff exit: 0

$ grep -c '!' README.md
0

$ perl -CSD -ne 'print if /[\x{1F000}-\x{1FFFF}\x{2600}-\x{27BF}]/' README.md
(no output; 0 lines)

$ wc -l README.md
     245 README.md

$ grep -nE '^\s*(BudgetWeight|RecencyWeight|RecencyHalfLifeDays) =|^const MaxPageSize|return max\(0, 1-2\*' backend/internal/domain/scoring.go backend/internal/domain/paginate.go
backend/internal/domain/scoring.go:28:	BudgetWeight = 0.6
backend/internal/domain/scoring.go:32:	RecencyWeight = 0.4
backend/internal/domain/scoring.go:34:	RecencyHalfLifeDays = 30
backend/internal/domain/scoring.go:53:	return max(0, 1-2*((price-target)/target))
backend/internal/domain/paginate.go:10:const MaxPageSize = 100
```

README line 161 states `BudgetWeight = 0.6`, `RecencyWeight = 0.4`, `RecencyHalfLifeDays = 30`, the over-budget factor 2 as a literal in `BudgetFit`, and `MaxPageSize = 100` in `domain/paginate.go`: all match.

## Compose blocks (README lines 7-9, 15-17, 30): BC-13 evidence

Not run here, per the orchestrator ruling. BC-13 artifact `art_01M4WAG45G8GAP1ATBBVZHAVYC` ("make smoke output 21cd239", role validation, BC-13 status done). From `/Users/forrest/Code/brillio/.lattice/orchestration/evidence/BC-13-smoke.txt`:

```
PASS  a1 api /healthz is ok
PASS  a2 api /api/cities lists Springfield
PASS  a3 api search returns items
PASS  a4 api search ranks the first item 1
PASS  a5 api search rejects minPrice > maxPrice
PASS  a6 web serves the app shell
PASS  a7 web proxies /healthz to api
PASS  a8 web proxies /api/cities to api
smoke: 8 passed, 0 failed
teardown: docker compose down -v
```

`BC-13-validation.md` records "Overall verdict: PASS", Docker 29.8.1 and Compose v5.5.1 (as README line 37 states), raw `env -u DOCKER_DEFAULT_PLATFORM docker compose ...` runs reaching a healthy stack, `make smoke` exit 0, and `make down` leaving no containers, volumes or networks.

The stack BC-13 smoked is the stack the README describes:

```
$ git diff 4030996 HEAD -- Makefile compose.yaml scripts/ backend/Dockerfile.dev frontend/vite.config.ts
(empty)
$ git diff 21cd239 HEAD -- Makefile compose.yaml scripts/ backend/Dockerfile.dev backend/.air.toml frontend/vite.config.ts
(empty)       # 21cd239 is the exact sha BC-13 smoked; 4030996 is its merge on master
```

Makefile `up:` recipe is `unset DOCKER_DEFAULT_PLATFORM; docker compose up --build` and `down:` is `docker compose down -v`, matching README lines 11 and 13. `scripts/smoke.sh` refuses to start when a brillio stack exists or 8080/5173 are in use and tears down via an EXIT trap, matching line 33. BC-17 re-proves these blocks from the zip.

## Notes

1. `docs/EXTENSIONS.md` (layout line 55, link at line 221) does not exist on master or at 79408db. It is a planned forward link: it lives on `feat/BC-16-extensions-doc` (commits f4b6c77, 3fc1aa4; BC-16 in `review`), and the BC-15 plan names it ("Forward link: docs/EXTENSIONS.md (BC-16) does not exist yet. BC-17's fresh-directory check covers it."). Not a README defect; it becomes one only if BC-16 does not land before the zip.
2. `make check` / `make test-api` reported Go packages as `(cached)`; the uncached `-count=1` run above removes that doubt.

## Process cleanup

- Started: api PID 42083 (parent `make dev-api` 42074) on 18015; vite node PID 44527 (parent `npm run dev` 44505) on 15115. Stopped with SIGTERM to the listening PIDs; the make and npm parents exited on their own (api: `server stopped`, make exit 0; npm exit 143). No `pkill`.
- After: `lsof -nP -iTCP:18015 -sTCP:LISTEN` and `lsof -nP -iTCP:15115 -sTCP:LISTEN` both empty (exit 1); all six PIDs in the chains gone.
- Worktree: `git status --porcelain` and `git -c core.excludesFile=/dev/null status --porcelain` both empty; HEAD still `79408dba42fa254cac14c363cc32a4021988c179`.
- Logs: `/private/tmp/claude-503/-Users-forrest-Code-brillio-worktrees-noindex-bc-15-readme/f5f3c1cf-65b0-4ee8-8523-13bc9ba26dc6/scratchpad/validation/` (make-*.log, dev-api.log, dev-web.log, *.out).
