# Brillio Listings

A listing search service: a Go REST API over the 12 sample listings (two MLS feeds, stored in SQLite) and a React/TypeScript UI that filters them, ranks them by budget fit and recency, and pages through the results. The engineering contract is `docs/SPEC.md`; the design system is `BRANDING.md`.

## Run it

```sh
make up
```

Then open http://localhost:5173. The API is on http://localhost:8080. `make up` runs `docker compose up --build` in the foreground, with hot reload for both halves; Ctrl-C stops it, and `make down` removes the containers, network and volumes. The first start is slow (image build, Go module download, `npm ci` in the container); the health checks allow 180 seconds.

Plain `docker compose up` also works. A shell that exports `DOCKER_DEFAULT_PLATFORM` must unset it first: with BuildKit off (`DOCKER_BUILDKIT=0`), a foreign platform there stops the api container from being created. `make up` unsets it itself; by hand:

```sh
env -u DOCKER_DEFAULT_PLATFORM docker compose up
```

On Linux with Docker Engine (no Docker Desktop or rootless uid mapping), the stack leaves a root-owned `frontend/node_modules` (an empty mount point) and `data/listings.db` in the checkout, so `make check`, `make dev-web` and `make dev-api` fail on the host. Run `make check` before `make up` and `node_modules` stays yours; after `make down`, `sudo chown -R "$(id -u):$(id -g)" frontend/node_modules data` hands both back.

Without Docker, in two terminals:

```sh
make dev-api    # builds backend/bin/api and serves http://localhost:8080
make dev-web    # Vite on http://localhost:5173, proxying /api and /healthz to :8080
```

The API reads `PORT` (8080), `DB_PATH` (`./data/listings.db`; `:memory:` works), `LISTINGS_FILE` (`./data/sample_listings.json`), `LOG_LEVEL` (info), `CORS_ORIGINS` (http://localhost:5173) and `SHUTDOWN_TIMEOUT` (10s), and seeds the database from the JSON file on every start. The Vite proxy target is `VITE_API_PROXY_TARGET` (default http://localhost:8080).

```sh
make check      # lint and test both halves; plain `make` runs this
make smoke      # Compose stack up, HTTP checks, stack down
```

`make check` runs `lint-api` (golangci-lint), `lint-web` (eslint, `tsc -b --noEmit`, prettier), `test-api` (`go test -race ./...`) and `test-web` (`vitest run`); each also runs alone, and `make lint` and `make test` run one kind for both halves. The web targets run `npm ci` first when `node_modules` is missing or older than the lockfile. `make smoke` needs ports 8080 and 5173 free and no brillio stack running, and always tears the stack down.

| Tool | Version |
|---|---|
| Docker with Compose | tested on Docker 29.8.1, Compose v5.5.1 |
| Go | 1.27 |
| Node and npm | Node 24 (24.15 or later, which jsdom 30 needs); tested on 24.21.0 |
| golangci-lint | v2; tested on 2.14.0 |
| make, curl | GNU Make 3.81 works; `make smoke` uses curl |

Pinned versions: `modernc.org/sqlite` v1.60.1 is the only direct Go dependency. The frontend uses React 19.3.0, Vite 8.3.4, TypeScript 6.0.3, Vitest 5.0.4, ESLint 9.39.5 and Prettier 3.9.10. The Compose images are `golang:1.27-alpine` (with air v1.67.4) and `node:24-alpine`.

- TypeScript 6.0.3 is pinned on purpose: TypeScript 7 has no classic compiler API, and typescript-eslint requires a TypeScript below 6.1.
- ESLint 9.39.5 is pinned on purpose: eslint-plugin-jsx-a11y 6.10.2 supports ESLint only up to 9.

## Project layout

```
compose.yaml               dev stack: api (Go under air), web (Vite)
Makefile                   check, lint, test, dev-api, dev-web, up, down, smoke, dist
BRANDING.md                design system: color, type, components, copy
docs/SPEC.md               engineering contract
docs/EXTENSIONS.md         extension playbook
docs/branding-demo.html    reference rendering of BRANDING.md
data/sample_listings.json  the 12 listings, feeds MLS_A and MLS_B
scripts/smoke.sh           make smoke
scripts/dist.sh            make dist: the release zip
scripts/verify-dist.sh     checks a release zip from a fresh directory
backend/
  go.mod, go.sum
  openapi.yaml, openapi.go   OpenAPI 3.1 document, embedded in the binary
  cmd/api/main.go            wiring only: config, seed, store, server, shutdown
  internal/config/           environment variables
  internal/domain/           listing, filter, scoring, ranking, paging; pure Go
  internal/store/            SQLite open, migrations/, queries
  internal/store/seed/       JSON loading and validation
  internal/httpapi/          router, handlers, query parsing, errors, middleware
  Dockerfile.dev, .air.toml, .golangci.yml
frontend/
  index.html, package.json, vite.config.ts, tsconfig.json, eslint.config.js, .prettierrc
  src/api/          client, types, errors
  src/hooks/        useSearchParamsState, useListingSearch, useCities
  src/format/       money, date, listing, keyword
  src/components/   Header, FilterBar, StatusLine, LoadingLine, ListingCard, Monogram,
                    ScorePill, ListingList, Pager, NoResults, ErrorState
  src/pages/        SearchPage.tsx
  src/styles/       tokens.css, global.css, shell.module.css
  src/test/         setup.ts, fixtures.ts
```

### How a search flows

`backend/internal/httpapi/search.go` holds the whole pipeline, one call per stage:

1. Parse and validate the query (`query.go`), collecting every problem before answering.
2. Filter in SQL (`store.Search`): one WHERE clause per filter (one per keyword token), rows ordered by `(source, id)`.
3. Dedupe slot: marked between search and rank, not built (see trade-offs).
4. Score and order every filtered listing at full precision (`domain.Rank`).
5. Paginate the whole ranked set (`domain.Paginate`), so ranks continue across pages.
6. Map to JSON, with relevance rounded to four decimals.

The UI keeps the search in the URL query string (`useSearchParamsState`), fetches it with `useListingSearch`, which aborts a superseded request, and renders each state BRANDING.md §5.9 names.

## API

| Endpoint | Returns |
|---|---|
| `GET /healthz` | 200 `{"status":"ok"}`, or 503 `degraded` when the database does not answer |
| `GET /api/cities` | the distinct cities, A to Z |
| `GET /api/listings/search` | one page of ranked listings |
| `GET /api/openapi.yaml` | the OpenAPI 3.1 document for all of the above |

Every response carries `X-Request-ID`: the client's, or a new one.

```sh
curl -s http://localhost:8080/healthz
# {"status":"ok"}
curl -s http://localhost:8080/api/cities
# {"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}
curl -s 'http://localhost:8080/api/listings/search?targetBudget=450000&city=Springfield'
```

Search takes nine parameters: `minPrice`, `maxPrice`, `minBedrooms`, `city`, `keyword`, `targetBudget`, `status` (`active`, `pending` or `sold`; repeat it or comma-separate it), `page` (default 1) and `pageSize` (default 10, at most 100). The Springfield search returns four listings, ranked A1, B7, A2, B8, in a body with four parts:

```
items           rank, the listing's fields, relevance {score, budgetFit, recency}
page            page, pageSize, totalItems, totalPages
scoring         targetBudget, budgetWeight, recencyWeight, recencyHalfLifeDays, asOf
appliedFilters  the filters as the server read them, null when absent; status echoes the effective set
```

API errors (400, 404, 405, 500) share one envelope; `fields` appears on a 400 only:

```
{"error": {
  "code": "validation_error",
  "message": "Invalid search parameters.",
  "fields": [{"field": "maxPrice", "code": "min_gt_max", "message": "Min price must be less than max price."}]
}}
```

The codes are `validation_error` (400), `not_found` (404), `method_not_allowed` (405) and `internal_error` (500; the detail goes to the log with the request ID, never to the client). The field codes are `not_a_number`, `not_an_integer`, `negative`, `not_positive`, `out_of_range`, `too_long`, `invalid_enum`, `min_gt_max`, `unknown_param` and `malformed_query` (reported on the field `query`).

## Scoring

Every listing gets a relevance score from 0 to 1:

```
budgetFit = price / target                              if price <= target
          = max(0, 1 - 2 * (price - target) / target)   if price > target
recency   = 0.5 ^ (ageDays / 30)
score     = 0.6 * budgetFit + 0.4 * recency             with a target budget
          = recency                                     without one
```

`ageDays` is the number of whole UTC calendar days from `listedDate` to the request's date; a future date counts as 0.

At a $450,000 target, $450,000 fits 1.0, $405,000 (10 percent under) 0.9, $495,000 (10 percent over) 0.8, $562,500 (25 percent over) 0.5, and $675,000 or more 0. Recency at 0, 30, 60 and 90 days is 1, 0.5, 0.25 and 0.125.

Example: A1 is $450,000, listed 2026-08-29. As of 2026-10-13 it is 45 days old, so budgetFit is 1.0, recency is `0.5 ^ (45 / 30)` = 0.3536, and the score is `0.6 * 1.0 + 0.4 * 0.3536` = 0.7414. Scores drift down each day: the same request on 2026-10-14 gives 0.7382. Each response names the date it scored against in `scoring.asOf`.

Why these choices:

- **Budget carries 60 percent.** The target budget is the one thing the user said they want. Recency carries 40 percent to break the near-ties a real market produces: A1 and B7 are $2,000 apart.
- **Over budget hurts twice as much as under.** A buyer can take a cheaper home but may not be able to stretch to a more expensive one. Fit reaches 0 at 50 percent over, and under the target only at a price of 0.
- **A 30-day half-life.** "A month old counts half" is easy to state and to check, and the curve is smooth: a listing loses a little each day instead of dropping out at a cut-off.
- **No budget, newest first.** Without a budget the score is recency alone, which falls with age, so results run newest first, as the status line says.

Ties: results sort by the unrounded score, descending, then `listedDate` descending, price ascending, source ascending and id ascending. `(source, id)` is unique, so the order is total and pages do not shift between requests with the same `scoring.asOf`. Without a budget, two listings from the same day show the cheaper one first. The API rounds `score`, `budgetFit` and `recency` to four decimals and the UI shows the score to two. Sorting uses full precision, so two cards can show the same score at sequential ranks.

The constants live in `backend/internal/domain/scoring.go`: `BudgetWeight = 0.6`, `RecencyWeight = 0.4` and `RecencyHalfLifeDays = 30`. They are code, not configuration, and every response echoes them in its `scoring` block (weights 0 and 1 without a budget). The over-budget factor 2 is a literal in `BudgetFit`, and `MaxPageSize = 100` is in `domain/paginate.go`. The comment at the top of `scoring.go`:

```go
// Relevance scoring (docs/SPEC.md section 6.2).
//
// Every listing gets a relevance score from 0 to 1. With a target budget the
// score is 60 percent budget fit plus 40 percent recency; without one it is
// recency alone, so results run newest first.
//
// Budget fit is 1 when the price equals the target. Under the target it is the
// price divided by the target (10 percent under scores 0.9). Over the target it
// falls twice as fast (10 percent over scores 0.8) and reaches 0 at 50 percent
// over.
//
// Recency halves every 30 days: 1 on the listing date, 0.5 at 30 days, 0.25 at
// 60. Age is whole UTC calendar days from the listed date to now; a future
// listed date counts as age 0.
```

## Handling invalid input

Validation runs before the database is touched, and every problem in a request comes back in one 400, in a fixed order, so a client can fix them all at once. This request returns the body below (formatted here):

```sh
curl -s 'http://localhost:8080/api/listings/search?minPrice=5&maxPrice=1&pageSize=0&minprice=1'
```

```json
{
  "error": {
    "code": "validation_error",
    "message": "Invalid search parameters.",
    "fields": [
      { "field": "maxPrice", "code": "min_gt_max", "message": "Min price must be less than max price." },
      { "field": "pageSize", "code": "out_of_range", "message": "Page size must be between 1 and 100." },
      { "field": "minprice", "code": "unknown_param", "message": "Unknown parameter." }
    ]
  }
}
```

- **An unknown parameter is a 400** (`unknown_param`), whatever its value. A typo such as `minprice=500000` would otherwise be ignored and silently widen the results.
- **A query string that is not valid URL encoding is a 400** (`malformed_query` on the field `query`), for example `?city=%zz`. Go's decoder would drop that pair and widen the results the same way.
- **Min above max** is `min_gt_max` on `maxPrice`; equal values are allowed. Numbers must be finite, prices and bedrooms 0 or more, `targetBudget` more than 0, and integers whole.
- **A page past the end is a 200** with empty `items` and the real page counts, so the client can page back.
- **A city with no listings is a 200** with `totalItems: 0`: no matches is an answer, not a mistake, and the UI shows its no-results state. An empty value such as `city=` counts as absent.

The filter bar checks the cheap rules before sending (min not above max, numbers 0 or more, budget more than $0) with the BRANDING.md §7.2 strings, and maps the server's field errors onto the same fields.

## Trade-offs

- **SQL filters, Go ranking.** Every filtered row loads into Go on each request, and scoring is pure, table-tested Go. At 12 rows that is correct and simple. The scaling path: compute the score in SQL or keep a precomputed column, use keyset pagination, and move keyword search to FTS5.
- **SQLite over a server database.** One file (or `:memory:`), a pure-Go driver with no cgo, migrations embedded in the binary and an idempotent seed on every start: nothing to provision. A server database would change `internal/store` and its wiring, not the domain or HTTP layers.
- **Cross-source duplicates stay visible.** Four pairs describe the same homes in both feeds (A1/B7, A2/B8, A3/B9, A5/B11). The handout says ids are unique per source and nothing more, so v1 keeps both rows and each card shows its source; a Springfield search with a $450,000 budget ranks A1, then B7. Merging them is a product decision, and `search.go` marks its slot between search and rank.
- **Strict input.** `unknown_param` and `malformed_query` make the API less forgiving (a client that adds a tracking parameter gets a 400) in exchange for never returning a wider result set than was asked for.
- **Stdlib only.** `net/http` routing, `log/slog`, `database/sql` and `testing`; the SQLite driver is the one direct Go dependency. The frontend has no router, state library, component library or mock server. Fewer dependencies leave less to defend.
- **No dark mode.** Decided in BRANDING.md §2.5: one light palette, done completely. The app declares `color-scheme: light`.

## Extending it

Each likely addition has a seam: a new filter is a rule in `httpapi/query.go`, a WHERE clause in `internal/store` and a field in the filter bar; a sort option sits beside `byRelevance` in `domain/rank.go`; dedupe fills the marked slot in `search.go`. [docs/EXTENSIONS.md](docs/EXTENSIONS.md) names the files to touch for each. `docs/SPEC.md` §9 lists the candidates.

## Testing

```sh
make test       # both halves
make test-api   # go test -race ./...
make test-web   # vitest run
```

Backend: stdlib `testing`, table-driven, run with `-race`.

- `domain`: budget fit at the target and at matching percentages under and over, recency at 0, 30 and 60 days and for a future date, every tie-break key, and pagination boundaries with rank continuity across pages.
- `store`: an in-memory database seeded from the real sample file; every filter, `%` and `_` matched literally, city case-insensitivity, and a second seed that leaves the row count unchanged.
- `httpapi`: `httptest` against a seeded store for status codes, error envelopes and codes, request IDs, CORS preflight, unknown parameters and a page past the end. The SPEC §5.4 and §5.5 examples are asserted against the handler with the clock pinned at 2026-10-13, and the OpenAPI document's parameters and examples are checked against the handler, its error codes against the SPEC.
- `config` and `store/seed`: environment parsing, and seed validation and normalization.

Frontend: Vitest, React Testing Library and jsdom, with `fetch` stubbed.

- Components render from props with BRANDING.md copy verbatim, and correct roles and labels.
- Formatters cover every row of BRANDING.md §7.1.
- Hooks: the URL round trip, `popstate`, and aborting a superseded request.
- `SearchPage`: loading to results, no results, a 400 mapped onto fields (`min_gt_max` marks both prices), an error with Try again, and paging with continuous ranks.

There is no automated browser suite. `make smoke` exercises the running stack over HTTP, against the API directly and through the Vite proxy.
