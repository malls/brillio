# BC-8 validation evidence

**N/A: BC-8 is library code (API client, hooks, formatters) that no page consumes until BC-12; evidence is the green gate and the verbose vitest run below.**

- Ticket: BC-8, "Frontend core: API client and types, URL state and search hooks, BRANDING formatters"
- Validator: agent:opus-5.5-validator
- Date: 2026-10-13 22:56 EDT
- Worktree: `/Users/forrest/Code/brillio-worktrees/bc-8-frontend-core`, branch `feat/BC-8-frontend-core`
- HEAD: `43e88241c99be1cef02d5cf44659fc3f17abdff3` (verified with `git rev-parse HEAD` before and after the run)
- Toolchain: node v24.21.0, go 1.27.1, vitest 5.0.4, vite 8.3.4
- Ports: none. No server or browser was started; the modules are tested under jsdom with `fetch` stubbed.

## Commands run

| # | Command (from the worktree root) | Exit |
|---|---|---|
| 1 | `git rev-parse HEAD` | 0 (`43e8824…`, matches) |
| 2 | `make check` (golangci-lint, eslint `--max-warnings 0`, `tsc -b --noEmit`, prettier `--check`, `go test -race ./...`, `vitest run`) | **0** |
| 3 | `cd frontend && npm run build` (`tsc -b --noEmit && vite build`) | **0** |
| 4 | `cd frontend && npx vitest run --reporter=verbose > …/BC-8-vitest-verbose.txt 2>&1` | **0** |
| 5 | `git status --porcelain` (after all runs) | 0, empty output (`frontend/dist/` is ignored by `.gitignore:25`) |

The Go packages reported `(cached)`. That is expected: `git diff --stat master...HEAD` touches only `docs/SPEC.md` and
`frontend/src/**` (19 files), so the backend is unchanged from `master`.

## Gate output tail (`make check`)

```
cd backend && golangci-lint run
0 issues.
cd frontend && npm run lint && npm run typecheck && npm run format:check
> eslint . --max-warnings 0
> tsc -b --noEmit
> prettier --check .
Checking formatting...
All matched files use Prettier code style!
cd backend && go test -race ./...
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	(cached)
ok  	brillio/backend/internal/domain	(cached)
ok  	brillio/backend/internal/httpapi	(cached)
cd frontend && npm test
> vitest run
 RUN  v5.0.4 /Users/forrest/Code/brillio-worktrees/bc-8-frontend-core/frontend
 Test Files  10 passed (10)
      Tests  197 passed (197)
```

## Build output tail (`npm run build`)

```
> tsc -b --noEmit && vite build
vite v8.3.4 building client environment for production...
✓ 20 modules transformed.
dist/index.html                   0.77 kB │ gzip:  0.41 kB
dist/assets/index-Bi9deQIM.css    2.56 kB │ gzip:  1.06 kB
dist/assets/index-C_W2cHaM.js   220.30 kB │ gzip: 68.89 kB
✓ built in 119ms
```

## Per-file test counts (verbose run, all passing, 0 failed, 0 skipped)

| File | Owner | Tests |
|---|---|---|
| `src/api/client.test.ts` | BC-8 | 30 |
| `src/api/errors.test.ts` | BC-8 | 13 |
| `src/hooks/useSearchParamsState.test.tsx` | BC-8 | 62 |
| `src/hooks/useListingSearch.test.tsx` | BC-8 | 14 |
| `src/hooks/useCities.test.tsx` | BC-8 | 3 |
| `src/format/money.test.ts` | BC-8 | 10 |
| `src/format/date.test.ts` | BC-8 | 26 |
| `src/format/listing.test.ts` | BC-8 | 35 |
| `src/components/Header/Header.test.tsx` | BC-7 | 3 |
| `src/App.test.tsx` | BC-7 | 1 |
| **Total** | 10 files | **197** (BC-8: 8 files, 193 tests) |

## Coverage cited from the verbose output

- **Client query building:** `buildSearchQuery` gives `''` for no params; writes every param in SPEC §5.3 order; repeats
  `status` deduplicated in enum order; omits empty / whitespace-only city, empty keyword, empty status, NaN, Infinity,
  -Infinity; plain number strings; trims and encodes city and keyword.
- **Client error parsing:** `rejects a 400 with a validation error, its fields and the request ID`; `rejects a 500 with an
  http error carrying the body message`; 404 http; dev-proxy HTML 502 fallback; malformed 2xx (non-JSON, no items, no
  page, JSON array); `rejects a fetch TypeError with a network error and its message`; `rejects an abort with a network
  error`; `treats a body read cut short by an abort as an abort`; `getCities` success, non-string drop, 500, wrong shape.
- **errors:** `parseErrorResponse` (all fields, malformed entries dropped, no fields, other codes as http, empty / HTML /
  array / no-error / empty-message fallbacks); `ApiRequestError` carries ApiError, request ID and cause; `toApiError` on
  ApiRequestError, Error and non-Error.
- **useSearchParamsState:** each of the nine params read individually, `round-trips the full set through
  buildSearchQuery`; 19 dropped invalid values; 7 kept range errors; status comma/repeat forms; trimming; first value of
  a repeated scalar; unknown params ignored; `follows popstate`, `stops following popstate after unmount`; `setSearch
  pushes the new filters without page` (page reset on search); `keeps pageSize through setSearch, setPage and clear`;
  `setPage` changes only page, `setPage(1)` removes it, no-ops for 0 / -1 / 1.5 / NaN; `clear writes an empty query`;
  identical search pushes nothing; params and writers stable across a rerender.
- **useListingSearch:** `starts loading with no previous, then succeeds`; `starts loading, then reports empty`; `keeps the
  previous response while new params load`; `aborts a request when params change and ignores its late response`; `does
  not refetch for equal params in a new object`; `reports a 400 as invalid with its fields and the previous response,
  unlogged`; `reports a 500 as an http error … and logs it once`; `reports a fetch failure as a network error and logs
  it`; `retry refetches the same query, keeping the previous response while loading`; `keeps retry stable across
  renders`; `reports a page past the end as success with no items`; `returning to a held query shows its result at once
  and refetches without loading` (held result, plan-review R2); `aborts the request on unmount without updating or
  logging`; `under StrictMode aborts the first request unlogged and settles on the second`.
- **useCities:** loads the cities once from `/api/cities`; stays empty and logs once on failure; StrictMode does not log
  the aborted first request.
- **BRANDING §7.1 formatters (all 11 rows):** Price (`$450,000`, `$0`, `$980`, `$1,234,567`, `399000.4` → `$399,000`,
  `399999.5` → `$400,000`); Bedrooms (`2 bd`); Bathrooms (`1 ba`, `1.5 ba`, `2 ba`, `2.5 ba`); Area (`1,450 sqft`,
  `980 sqft`); Listed date with injected year (current year omits it, 2025 and 2027 add it, all twelve months, leading
  zero dropped, Jan 1 stays on its day, raw fallback for 7 bad inputs, default year from the local clock, run under a
  timezone where `Date` parsing would shift the day); Source (underscore to space); Status (label only when not active);
  Score (`0.94`, `0.7414` → `0.74`, `1.00`, `0.00`); Rank (`No. ` + integer); Counts (`1 result`, `12 results`,
  `0 results`, `1,200 results`); Pages (`Page N of M`, empty at 0 pages).
- **Card and status line:** location `City, ST ZIP`; meta items in BRANDING §5.4 order without the status label;
  monograms (`Sp`, `Vi`, `Fa`, `RESTON` → `Re`, `'  Manassas '` → `Ma`, `chantilly` → `Ch`, `''` → `''`); `names the
  ordering`; result summary for ranked, no-budget and empty responses; page position of the fixtures.

## Validation items

| Item | Result |
|---|---|
| (a) One-line N/A justification recorded (no page consumes BC-8 until BC-12) | **PASS** (top of this file) |
| (b) Vitest output cites the client, hook and formatter cases | **PASS** (8 BC-8 files, 193 tests, every case group above present by name) |
| (c) Gate green on HEAD `43e8824` (`make check` 0, `npm run build` 0, verbose vitest 0) | **PASS** |

**Overall: PASS.**

## Raw output

Full verbose vitest output: `/Users/forrest/Code/brillio/.lattice/orchestration/evidence/BC-8-vitest-verbose.txt`
(attached alongside this file as "Vitest verbose 43e8824").
