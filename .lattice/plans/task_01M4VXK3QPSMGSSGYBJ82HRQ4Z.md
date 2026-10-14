# BC-15 plan: README (run, architecture, scoring rationale, trade-offs)

Planner: agent:opus-5.5-planner. Facts come from `feat/BC-15-readme` @261760f, which merges master bf36a3d, feat/BC-12 @5ac2397 and feat/BC-6 @70518b5. Responses were observed on a native API run with PORT=18015 and DB_PATH=:memory:.

## Scope and files

- Only `README.md` changes. It replaces the 8-line BC-1 stub; per the shared-files table, BC-1 owns the stub and BC-15 is the real writer.
- No SPEC edit is needed. Leave out `make dist`, which is still a BC-17 stub.
- Length: 200-240 lines, hard cap 250.
- Voice: BRANDING §7. No emoji, no exclamation marks, and no markdown images or HTML comments (both contain an exclamation mark).

## Facts the README asserts (all verified)

- **Constants**: `backend/internal/domain/scoring.go` holds three untyped consts: `BudgetWeight = 0.6`, `RecencyWeight = 0.4`, `RecencyHalfLifeDays = 30`. They are not config, and no other scoring constant exists.
- **Budget fit**: at or under the target, fit is `price/target`. Over the target, `BudgetFit` computes `max(0, 1-2*((price-target)/target))`. The factor 2 is a literal, not a constant.
- **Recency**: `0.5^(AgeDays/30)`. AgeDays counts whole UTC calendar days; a future date counts as 0.
- **Score**: `0.6*fit + 0.4*recency`, or recency alone without a budget.
- **Quote**: scoring.go lines 1-13, verbatim, in a fenced block that keeps the `//` prefixes. A blank line separates them from `package domain`, so call it "the comment at the top of scoring.go".
- **Order**: `byRelevance` in rank.go sorts by unrounded score desc, then listedDate desc, price asc, source asc, id asc. The order is total. Without a budget, results run newest first, and the cheaper listing comes first within a day.
- **Paging**: `domain.MaxPageSize = 100`. The defaults (page 1, pageSize 10) and the length limits (city 100, keyword 200 characters) are unexported in `httpapi/query.go`. `rank` runs across the whole result set, so page 2 starts at 11.
- **Rounding**: the API rounds score, budgetFit and recency to 4 decimals; the UI shows 2; sorting uses unrounded values.
- **Scoring echo**: every response carries a `scoring` block with `targetBudget`, `budgetWeight`, `recencyWeight`, `recencyHalfLifeDays` and `asOf`. Without a budget the first three read null, 0 and 1.
- **Worked values** (SPEC §6.2, consistent with the code):
  - Fit at a 450,000 target: 450,000 gives 1.0, 405,000 gives 0.9, 495,000 gives 0.8, 562,500 gives 0.5, and 675,000 or more gives 0.
  - Recency at 0, 30, 60 and 90 days: 1, 0.5, 0.25 and 0.125.
- **Example** (SPEC §5.4): A1 at $450,000, listed 2026-08-29, asOf 2026-10-13. Age 45 days, recency 0.3536, score 0.6 + 0.4 x 0.3536 = 0.7414.
  - The README must name that date. On 2026-10-14 the same request gives 0.7382.
  - The tests pin the clock at 2026-10-13 and assert the §5.4 block (`specExample` in `httpapi/search_test.go`).
- **Rationale** (SPEC §6.2): budget is the stated intent; recency breaks near-ties; "a month old counts half"; over budget hurts twice as much as under.
- **Endpoints** (`httpapi/router.go` registers every route):
  - `/healthz` returns 200 `{"status":"ok"}`, or 503 with `degraded`.
  - `/api/cities` returns the six cities, Chantilly through Vienna.
  - `/api/listings/search` accepts nine params: minPrice, maxPrice, minBedrooms, city, keyword, targetBudget, status (repeatable or comma-separated), page, pageSize.
  - `/api/openapi.yaml` serves OpenAPI 3.1.1.
  - Every response sets `X-Request-ID`.
- **Error envelope**: `{"error":{"code","message","fields":[{"field","code","message"}]}}`, with `fields` on a 400 only.
  - Top-level codes: `validation_error` 400, `not_found` 404, `method_not_allowed` 405, `internal_error` 500.
  - Ten field codes: not_a_number, not_an_integer, negative, not_positive, out_of_range, too_long, invalid_enum, min_gt_max, unknown_param, and malformed_query (reported on the field `query`).
- **Observed**:
  - The SPEC §5.5 example request returns that exact body.
  - `?city=%zz` returns 400 `malformed_query`.
  - `?page=99` returns 200 with empty items.
  - `?city=Nowhere` returns 200 with `totalItems: 0`.
- **Make targets**:
  - `check` is the default goal; `lint` runs `lint-api` and `lint-web`; `test` runs `test-api` and `test-web`.
  - `lint-api` runs golangci-lint; `test-api` runs `go test -race ./...`.
  - `lint-web` runs eslint, `tsc -b --noEmit` and prettier; `test-web` runs `vitest run`. Web targets run `npm ci` first when needed.
  - `dev-api` builds `backend/bin/api` and runs it from the repo root. `dev-web` runs `npm run dev`.
  - `up` runs `unset DOCKER_DEFAULT_PLATFORM; docker compose up --build` in the foreground.
  - `down` runs `docker compose down -v`. `smoke` runs `sh scripts/smoke.sh`.
- **Compose** (project `brillio`):
  - `api` runs golang:1.27-alpine with air on :8080.
  - `web` runs node:24-alpine with Vite HMR on :5173, starts once `api` is healthy, and proxies `/api` and `/healthz`.
  - Natively, the proxy target defaults to `http://localhost:8080` (`VITE_API_PROXY_TARGET`).
- **Platform footgun**: `make up` and smoke.sh unset only `DOCKER_DEFAULT_PLATFORM`, not the BuildKit variables.
  - The README says plain `docker compose up` works. A shell that exports `DOCKER_DEFAULT_PLATFORM` should use `make up` or `env -u DOCKER_DEFAULT_PLATFORM docker compose up` (compose.yaml's own wording).
- **smoke.sh**:
  - Refuses to start when :8080 or :5173 is busy or a brillio stack exists.
  - Brings the stack up with `--wait` (300 s timeout), then runs 8 checks: a1-a5 against the API, a6-a8 through the proxy.
  - Always tears down with `down -v`.
- **Versions**:
  - Backend: Go 1.27; modernc.org/sqlite v1.60.1, the only direct Go dependency.
  - Node 24, at least 24.15 because jsdom 30.1.2 requires it; tested on 24.21.0.
  - Frontend packages: React 19.3.0, Vite 8.3.4, TypeScript 6.0.3, Vitest 5.0.4, ESLint 9.39.5, Prettier 3.9.10.
  - Tooling: golangci-lint v2 (tested 2.14.0); Docker 29.8.1 with Compose v5.5.1; GNU Make 3.81 works.
- **Pins** (SPEC §2): one line each, with the reason.
  - TypeScript 6.0.3: TS 7 has no classic compiler API, and typescript-eslint needs a version below 6.1.
  - ESLint 9.39.5: jsx-a11y 6.10.2 peers ESLint up to 9.
- **Config** (SPEC §11): `PORT` 8080, `DB_PATH` ./data/listings.db (`:memory:` allowed), `LISTINGS_FILE`, `LOG_LEVEL` info, `CORS_ORIGINS`, `SHUTDOWN_TIMEOUT` 10s.
- **No dark mode**: decided in BRANDING §2.5 ("None, by decision") and §8. The CSS declares `color-scheme: light`.
- **Duplicates**: four cross-feed pairs stay as separate rows (A1/B7, A2/B8, A3/B9, A5/B11), and each card shows its source.
  - `search.go` marks the dedupe slot between `Search` and `Rank`.
  - Springfield at 450000 ranks A1, then B7.
- **Backend tests**: stdlib, table-driven, run with `-race`. They cover config, domain, httpapi, store and store/seed, including:
  - SPEC example asserts and OpenAPI sync tests.
  - The `%`/`_` literal-match tests and the idempotent-seed test.
- **Frontend tests**: Vitest, RTL and jsdom with fetch stubbed, across 23 test files.
- **No automated e2e**: BC-12's browser e2e was a manual validation, so the README must not claim an e2e suite. No hard-coded test counts.

## README outline (one section per Scope bullet)

1. `# Brillio Listings`: what it is, in two sentences.
2. `## Run it`:
   - Tools and versions.
   - `make up`, then open http://localhost:5173. The note on `docker compose up` and the platform variable.
   - `make dev-api` and `make dev-web`, plus the config line.
   - `make check`, the per-half targets and `make smoke`.
   - The two pin lines.
3. `## Project layout`: the SPEC §3 tree, matched to `git ls-files`.
   - Add `openapi.go`, `store/seed/`, `store/migrations/` and `format/keyword.ts`.
   - Keep `docs/EXTENSIONS.md`.
4. `## API`:
   - A table of the four endpoints.
   - One curl each for healthz, cities and search (`targetBudget=450000&city=Springfield`).
   - A pointer to openapi.yaml.
   - The envelope in five lines, then the codes. Show only stable outputs.
5. `## Scoring`:
   - The formula in words and symbols, the worked values and the dated A1 example.
   - The three reasons, tie-breaking, the no-budget order, rounding, and the constants' location.
   - The quote.
6. `## Handling invalid input`, each rule with its reason:
   - `unknown_param` 400: a typo must not widen results.
   - `malformed_query` 400: the decoder would drop the pair.
   - All errors are collected; `pageSize=0` is `out_of_range`; min > max is `min_gt_max` on maxPrice (equal allowed).
   - Page past the end and a city with no matches are 200s (`totalItems: 0`); an empty `city=` is absent.
7. `## Trade-offs`:
   - SQL filters with Go ranking. Scaling path: score in SQL or a precomputed column, keyset paging, FTS5.
   - SQLite.
   - Visible duplicates.
   - Stdlib only.
   - Strict input.
   - No dark mode.
8. `## Extending it`: three sentences, a link to `docs/EXTENSIONS.md`, and SPEC §9.
9. `## Testing`:
   - What each half covers, mapped to the handout's edge cases.
   - How to run each: `make test`, `make test-api`, `make test-web`.

## Acceptance checks

- **A1**: every fenced command in the README runs once (see Validation), and the outputs are cited in the completion comment.
- **A2**:
  - `sed -n 1,13p` of scoring.go diffs empty against the quoted block.
  - 0.6, 0.4, 30 and 100 match the code. 0.7414 is recomputed.
  - All 10 field codes appear, each as defined in `query.go`.
- **A3**:
  - `grep -c '!' README.md` prints 0.
  - An emoji-range `perl -CSD` grep prints nothing.
- **Card notes**:
  - `make up` is the primary command.
  - Both pin lines are present.
  - Trade-offs names `unknown_param`, `malformed_query` and the duplicates.
  - `wc -l` is 250 or less.
  - Every path in the tree exists, apart from the EXTENSIONS.md forward link.

## Validation (record N/A with `--role validation`, plus these runs)

Check ports 18015, 15115, 8080 and 5173 with `lsof -nP -iTCP:<port> -sTCP:LISTEN` first. Write logs to the scratchpad.

1. Run `make check` on the final head.
2. Run `PORT=18015 DB_PATH=:memory: make dev-api &`.
   - Run each README curl with 8080 changed to 18015.
   - Kill the listener PID (from `lsof -t`) and its make parent. Never `pkill -f backend/bin/api`: other worktrees share that command line.
3. Run `cd frontend && VITE_API_PROXY_TARGET=http://localhost:18015 npm run dev -- --port 15115 --strictPort`, which is the dev-web recipe on the ticket's port.
   - curl `/` for the page title, then `/api/cities` through the proxy.
   - Kill the process.
4. Compose, only if 8080 and 5173 are free and no brillio stack exists (`docker compose ps -a -q` prints nothing):
   - Run `make up &`, poll `:5173/api/cities` for up to 300 s, then `make down`.
   - Run `env -u DOCKER_DEFAULT_PLATFORM docker compose up -d --wait`, curl, then `make down`.
   - Run `make smoke` and expect 8 passed, 0 failed.
   - If the ports are busy, skip this step and record why.
5. Run the acceptance checks. Confirm that no listener and no brillio container remains.

## Risks and size

- **Anchors**: BC-6 and BC-12 are not merged yet. Re-verify the tree, versions and Makefile after merging master. BC-14 may touch `frontend/src/`.
- **Forward link**: `docs/EXTENSIONS.md` (BC-16) does not exist yet. BC-17's fresh-directory check covers it.
- **Dates and Docker**: live scores drift daily, so pin the example's date. The Docker footgun is described above.
- **Size**: about 220 README lines, one commit.

## SPEC vs code

No contradictions in the scoring numbers. Minor drift, which the README handles by describing the code:

- §3 omits `openapi.go`, `store/seed/` and `format/keyword.ts`.
- §10 shows a shorter Vite `--host` flag set than compose.yaml uses.
- §12 and the ticket say `docker compose up`, which fails on a host that exports `DOCKER_DEFAULT_PLATFORM` and runs the classic builder.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review: art_01M4WC4P3HKCDXPGM3HC7Y76XW (fable sub-agent, FAIL, amendment-level). Branch head is now c4d50a9 (master merged after BC-6 landed at 269f686); `git diff --stat 261760f c4d50a9 -- . ':!.lattice'` is empty, so every fact above still holds.

- **M1 (quote range)** — The comment at the top of `backend/internal/domain/scoring.go` is **lines 1-14** (line 14 is `// listed date counts as age 0.`), not 1-13. Quote all 14 lines verbatim. A2 becomes count-independent: `awk 'NF==0{exit} {print}' backend/internal/domain/scoring.go | diff - <quoted block extracted from README>` prints nothing. Sections: Facts > Quote, Acceptance > A2.
- **M2 (Compose outside scope)** — Validation step 4 is **dropped by default**. Run-state scopes Compose on 8080/5173 to BC-13/BC-17, and `make down` is `docker compose down -v` on the fixed project `brillio`. For the Docker blocks (`make up`, `docker compose up`, `make smoke`, `make down`) cite BC-13's landed evidence (real `make smoke` 8/8, art_01M4WAG45G8GAP1ATBBVZHAVYC, run-state 04:28Z) and note that BC-17 re-proves them from the zip. The delegator has asked the Orchestrator; if approval arrives it is recorded on the ticket and the validator runs `make smoke` once only (preflight: 8080/5173 free and `docker compose -p brillio ps -a -q` empty). Section: Validation.
- **Platform wording (card note 2)** — `make up` and smoke.sh unset only `DOCKER_DEFAULT_PLATFORM`. The README says: `make up` is the primary command; plain `docker compose up` also works, but a shell that exports `DOCKER_DEFAULT_PLATFORM` (it bites with BuildKit off, `DOCKER_BUILDKIT=0`) must unset it first, e.g. `env -u DOCKER_DEFAULT_PLATFORM docker compose up`. Do not claim `make up` touches the BuildKit variables. Section: Run it.
- **m1 (architecture)** — Add a short "How a search flows" part (6-8 lines) under Project layout: parse and validate, collecting every error -> SQL filter (`store.Search`) -> dedupe slot (marked, not built) -> `domain.Rank` -> `domain.Paginate` (ranks over the whole set) -> JSON DTOs rounded to 4 decimals. Mirror the stage comments in `backend/internal/httpapi/search.go`, which the implementer reads first. Plus one frontend line: the URL query string is the state, a hook fetches, components render. Section: outline 3.
- **m2 (counts)** — No test counts and no test-file counts anywhere in the README. Describe coverage by area (backend: config, domain, httpapi, store; frontend: components, formatters, hooks, SearchPage). Section: Facts > Frontend tests, outline 9.
- **m3 (make down)** — Run it gains one line for `make down` (removes containers, network and volumes) and one note that the first `make up` is slow (module download and `npm ci`; health checks allow a 180 s start period). Verify the 180 s value in compose.yaml before writing it. Section: outline 2.
- **n2** — The §5.5 example request is `minPrice=5&maxPrice=1&pageSize=0&minprice=1` (search_test.go:725). Use it for the README's error example and the validation curl.
- **n3** — Voice: BRANDING §7 also bans "please", apologies and cheer. A3 adds `grep -niE '\bplease\b|sorry|oops|awesome|great news' README.md` printing nothing.
- **n4** — Validation step 3 (`npm run dev`) depends on `frontend/node_modules`, which step 1's `make check` installs via `npm ci`. Keep that order.
- **n5** — The Project layout tree omits the two handout PDFs and `.lattice/`, which SPEC §12 excludes from the zip; it lists nothing that does not ship. Other top-level files (`CLAUDE.md`, `agents.md`) may be left out for brevity. SPEC §12 also says a stranger must be able to follow the README to run `docker compose up` and `make check`, so the plain-compose line stays in Run it.
- **n1** — Anchor updated above (c4d50a9).
