# Listing Search Service — Build Spec

Decided 2026-10-13 in a planning interview (Forrest + Claude), then reconciled with
`BRANDING.md` v1. This document is the single source of truth for engineering; `BRANDING.md`
is the single source of truth for how the UI looks and speaks. Where this file describes UI,
it defers to BRANDING.md and only adds what BRANDING.md leaves to engineering. Every Lattice
task references sections here by number. If an implementer must deviate, update this file in
the same branch and explain why in a `lattice comment` on the task.

Source material: `Coding_Test_Question_Paper.pdf` (the handout), `Instructions.pdf`
(take-home addendum), `data/sample_listings.json` (12 listings, two feeds), `BRANDING.md`.

---

## 1. Goal

A full-stack listing search: a Go REST API over the sample listings and a React/TypeScript
UI that calls it. The user filters listings, supplies a target budget, and sees results
ranked by a budget-and-recency relevance score, paged. Invalid input produces clear errors,
never crashes or silently wrong data.

The interviewer will add unseen requirements live. Clean extension points (§9) matter as
much as the features.

Interview bar: **senior**. That means clean layering, table-driven tests, lint, structured
logging with request IDs, graceful shutdown, CORS, a health endpoint, an OpenAPI document,
and a README that explains the scoring and its trade-offs.

## 2. Stack and versions

Latest stable as of 2026-10-13. Verify with `go version` / `npm view <pkg> version` at
implementation time and use the newest stable patch.

| Layer | Choice |
|---|---|
| Backend | Go 1.27 (`go 1.27` in go.mod). Stdlib only: `net/http` with Go 1.22+ `ServeMux` method/path patterns, `log/slog`, `database/sql`, `embed`. |
| SQLite driver | `modernc.org/sqlite` v1.60+ (pure Go, no cgo, cross-compiles cleanly in Docker). |
| Database | SQLite file at `DB_PATH` (`:memory:` allowed). Schema from embedded SQL migrations. Seeded from JSON on every start (idempotent upsert). |
| Frontend | Vite 8, React 19, TypeScript 6.0.x (strict; see note below), ESLint + typescript-eslint + Prettier, Vitest 5 + React Testing Library + jsdom. |
| FE state | Plain hooks + `fetch`. No state library, no router. Filters live in the URL query string. |
| Styling | CSS custom properties from BRANDING.md §2.1 in `src/styles/tokens.css`, CSS modules per component. Google Fonts link from BRANDING.md §3.1 in `index.html`. No Tailwind, no component library. |
| Runtime | Docker Compose v5 dev stack with hot reload: `api` runs Go under `air`, `web` runs the Vite dev server proxying `/api`. Native `make` targets also work without Docker. |
| Git | Local only, no remote. `master` is the integration branch. Branch per task `feat/BC-N-short-slug`. Merge with `--no-ff` once validated. Conventional commit messages. No emoji in commits, code, docs or UI (BRANDING.md §8). |

TypeScript version (operator ruling, 2026-10-14): TypeScript 7.0 ships only the native compiler and
exports no classic compiler API, and typescript-eslint (8.71, latest) requires `typescript <6.1.0`. The
frontend therefore pins TypeScript 6.0.3, the last 6.x release, so ESLint type-aware linting works with no
workarounds. Moving to TypeScript 7 is a one-line bump once typescript-eslint supports it.

ESLint version (2026-10-14): eslint-plugin-jsx-a11y 6.10.2 (latest) peers `eslint ^3 … ^9`, so the
frontend pins ESLint 9.39.5, the newest 9.x, with no peer overrides (typescript-eslint and react-hooks
already accept ^10). Moving to ESLint 10 is a one-line bump once jsx-a11y supports it. `@types/node`
tracks the Node 24 runtime (24.x), not the newest major.

Rules: no third-party HTTP router, no testify, no ORM. On the frontend, no state library,
no router, no msw, no icon set. Fewer dependencies means less to defend live.

## 3. Repository layout

```
/
  compose.yaml            dev stack (§10)
  Makefile                up, down, dev-api, dev-web, test, lint, check, smoke, dist
  README.md               run instructions, architecture, scoring rationale (BC-15)
  BRANDING.md             design system (owned separately; see its §9 change process)
  CLAUDE.md, agents.md    agent instructions (existing)
  docs/SPEC.md            this file
  docs/EXTENSIONS.md      live-session extension playbook (BC-16)
  docs/branding-demo.html reference rendering of BRANDING.md (owned by branding)
  data/sample_listings.json
  scripts/smoke.sh        make smoke: stack up, API and proxy assertions, teardown (§10)
  backend/
    go.mod, go.sum
    cmd/api/main.go                 wiring only: config, store, server, shutdown
    internal/config/                env parsing (§11)
    internal/domain/                Listing, SearchFilter, scoring, ordering, paging (pure Go)
    internal/store/                 sqlite open, migrations (embedded *.sql), seed, queries
    internal/httpapi/               router, handlers, query parsing/validation, errors, middleware
    openapi.yaml                    served at GET /api/openapi.yaml (embedded)
    Dockerfile.dev, .air.toml, .golangci.yml
  frontend/
    package.json, vite.config.ts, tsconfig.json, eslint.config.js, .prettierrc
    index.html                      fonts link, <title>Brillio Listings</title>
    src/main.tsx, src/App.tsx
    src/api/            client.ts, types.ts, errors.ts
    src/hooks/          useSearchParamsState.ts, useListingSearch.ts, useCities.ts
    src/format/         money.ts, date.ts, listing.ts   (BRANDING.md §7.1 formatters)
    src/components/     Header/, FilterBar/, StatusLine/, LoadingLine/, ListingCard/, Monogram/,
                        ScorePill/, ListingList/, Pager/, NoResults/, ErrorState/
    src/pages/          SearchPage.tsx
    src/styles/         tokens.css, global.css, shell.module.css
    src/test/           setup.ts, fixtures.ts
```

## 4. Data model

Listing fields are exactly the handout's:

| Field | Go type | JSON | Notes |
|---|---|---|---|
| id | string | `id` | Unique per source only |
| source | string | `source` | e.g. `MLS_A`, `MLS_B` |
| address | string | `address` | Free text, not normalized |
| city, state, zip | string | same | Minor inconsistencies across sources |
| price | float64 | `price` | USD list price |
| bedrooms | int | `bedrooms` | |
| bathrooms | float64 | `bathrooms` | |
| sqft | int | `sqft` | |
| latitude, longitude | float64 | same | |
| listedDate | time.Time (date only) | `listedDate` | ISO `YYYY-MM-DD` |
| status | string enum | `status` | `active` \| `pending` \| `sold` |
| description | string | `description` | Free text, keyword-searchable |

**Identity is the pair `(source, id)`.** The handout says `id` is unique per source, not
across sources. Never key anything on `id` alone.

### 4.1 Schema

Table `listings` with one column per field, snake_case, `PRIMARY KEY (source, id)`,
`listed_date TEXT` holding `YYYY-MM-DD`, `status TEXT CHECK (status IN ('active','pending','sold'))`.
Indexes on `city COLLATE NOCASE`, `price`, `bedrooms`, `status`. A `schema_migrations` table
records applied migration filenames. Migrations live in
`backend/internal/store/migrations/NNNN_name.sql`, embedded with `embed.FS`, applied in order
inside a transaction on open.

### 4.2 Seed

On startup the API reads `LISTINGS_FILE`, validates every record, normalizes it, and upserts
with `ON CONFLICT(source, id) DO UPDATE`. Validation: all fields present, `status` in the
enum, `listedDate` parses as `YYYY-MM-DD`, `price >= 0`, `bedrooms >= 0`, `sqft >= 0`,
`id` and `source` non-empty after trimming, and `(source, id)` unique within the file.
Normalization: trim every string, collapse internal whitespace in `city` and `state`,
uppercase `state`. Any invalid record fails the whole seed and the process exits non-zero
with a log line naming the record index and the problem. Log the count seeded at info level.

The sample data contains four near-duplicate pairs across feeds (A1/B7, A2/B8, A3/B9,
A5/B11). **v1 does not dedupe them.** They remain distinct rows and the card's meta line
shows the source (`MLS A`, `MLS B`) so the duplication is visible. See §9.1.

## 5. API contract

Base path `/api`. JSON in and out, UTF-8, `Content-Type: application/json; charset=utf-8`.
Every response carries `X-Request-ID` (generated per request, or echoed if the client sent
one). Field names are camelCase.

### 5.1 `GET /healthz`

`200 {"status":"ok"}`. Also pings the database; returns `503 {"status":"degraded","error":"..."}`
if that fails.

### 5.2 `GET /api/cities`

```json
{"cities": ["Chantilly", "Fairfax", "Manassas", "Reston", "Springfield", "Vienna"]}
```

Distinct case-insensitively, sorted A→Z, display casing is the first-seen value. The UI
uses this to populate the City `<select>` (BRANDING.md §5.2), with "Any" first.

### 5.3 `GET /api/listings/search`

| Param | Type | Rule |
|---|---|---|
| `minPrice` | number | optional, `>= 0` |
| `maxPrice` | number | optional, `>= 0`, and `>= minPrice` when both present |
| `minBedrooms` | integer | optional, `>= 0` |
| `city` | string | optional, trimmed, max 100 chars, case-insensitive exact match |
| `keyword` | string | optional, max 200 chars, whitespace-split into tokens, every token must appear (case-insensitive substring) in `description` |
| `targetBudget` | number | optional, `> 0` |
| `status` | enum, repeatable | optional, each value in `active|pending|sold`; repeat the param or comma-separate; absent means all three |
| `page` | integer | optional, default 1, `>= 1` |
| `pageSize` | integer | optional, default 10, `1..100`. Not a UI control in v1 (BRANDING.md §5.7); honored if present in the URL. |

Parsing rules:

- An empty value (`minPrice=`) is treated as absent.
- Numbers parse with `strconv.ParseFloat`; integers must be whole (`minBedrooms=2.5` is a field error).
- **Unknown query parameters are a 400** with a field error naming the parameter. Rationale: a typo like `minprice=500000` must not silently widen the result set. Deliberate; say it out loud in the interview.
- All validation problems are collected and returned together, not first-error-only.
- `page` past the last page is **not** an error: `200` with empty `items` and correct `page` metadata so the UI can recover.
- A city with no listings is **not** an error: `200` with `totalItems: 0`. The UI renders the no-results state.

### 5.4 Success response

```json
{
  "items": [
    {
      "rank": 1,
      "id": "A1", "source": "MLS_A",
      "address": "123 Main St, Apt 4B", "city": "Springfield", "state": "VA", "zip": "22150",
      "price": 450000, "bedrooms": 2, "bathrooms": 1.5, "sqft": 980,
      "latitude": 38.7893, "longitude": -77.1873,
      "listedDate": "2026-08-29", "status": "active",
      "description": "Bright top-floor condo near shops and transit. Pet friendly.",
      "relevance": { "score": 0.8731, "budgetFit": 1.0, "recency": 0.3536 }
    }
  ],
  "page": { "page": 1, "pageSize": 10, "totalItems": 12, "totalPages": 2 },
  "scoring": {
    "targetBudget": 450000,
    "budgetWeight": 0.6, "recencyWeight": 0.4,
    "recencyHalfLifeDays": 30,
    "asOf": "2026-10-13"
  },
  "appliedFilters": {
    "minPrice": null, "maxPrice": null, "minBedrooms": null,
    "city": null, "keyword": null, "status": ["active", "pending", "sold"]
  }
}
```

- `rank` is the 1-based position in the **whole** ranked result set, so page 2 with pageSize 10 starts at 11 (BRANDING.md §5.6).
- `relevance.score` is in `[0, 1]`, rounded to four decimals. `budgetFit` and `recency` are in `[0, 1]`, four decimals. The UI shows the score to two decimals.
- When `targetBudget` is absent: `scoring.targetBudget` is `null`, `budgetWeight` is `0`, `recencyWeight` is `1`, and `relevance.budgetFit` is `null`.
- `appliedFilters` echoes the effective filters after parsing, with `null` for absent ones, so the client never has to guess what the server understood.

### 5.5 Error response

```json
{
  "error": {
    "code": "validation_error",
    "message": "Invalid search parameters.",
    "fields": [
      { "field": "maxPrice",  "code": "min_gt_max",     "message": "Min price must be less than max price." },
      { "field": "pageSize",  "code": "out_of_range",   "message": "Page size must be between 1 and 100." },
      { "field": "minprice",  "code": "unknown_param",  "message": "Unknown parameter." }
    ]
  }
}
```

| HTTP | `code` | When |
|---|---|---|
| 400 | `validation_error` | any §5.3 rule fails; `fields` is non-empty |
| 404 | `not_found` | unknown route |
| 405 | `method_not_allowed` | wrong method on a known route |
| 500 | `internal_error` | unexpected failure; message is generic, details go to the log with the request ID, never to the client |

Field error `code` values: `not_a_number`, `not_an_integer`, `negative`, `not_positive`,
`out_of_range`, `too_long`, `invalid_enum`, `min_gt_max`, `unknown_param`. The min/max
conflict is reported once, on `maxPrice`, with code `min_gt_max`; the UI marks both price
fields and shows the message under Max price (BRANDING.md §5.2). Messages use the exact
strings from BRANDING.md §7.2 where one exists, otherwise a plain sentence in the same voice
(sentence case, says what is wrong and what to do, no exclamation marks).

### 5.6 CORS and headers

Allow origins from `CORS_ORIGINS` (default `http://localhost:5173`), methods `GET, OPTIONS`,
headers `Content-Type, X-Request-ID`. Expose `X-Request-ID`. Preflight returns 204.

### 5.7 `GET /api/openapi.yaml`

Serves the embedded OpenAPI 3.1 document describing §5.1–5.5 (BC-6).

## 6. Search pipeline

One function in `httpapi` owns the stage order so the interviewer can see it and so a new
stage slots in cleanly:

```
parse+validate query  →  store.Search(ctx, filter)              // SQL WHERE, §6.1
                      →  [dedupe slot, not built, §9.1]
                      →  domain.Rank(listings, RankInput)        // score + order, §6.2–6.3
                      →  domain.Paginate(ranked, page, pageSize) // §6.4, assigns rank
                      →  response mapping (§5.4)
```

Trade-off, stated in the README: every filtered row loads into Go per request. Correct and
simple at this scale; the scaling path is pushing the score into SQL or precomputing it,
plus keyset pagination.

### 6.1 Filtering (in SQL, `internal/store`)

| Filter | SQL |
|---|---|
| minPrice | `price >= ?` |
| maxPrice | `price <= ?` |
| minBedrooms | `bedrooms >= ?` |
| city | `city = ? COLLATE NOCASE` (input trimmed) |
| keyword | per token: `instr(lower(description), ?) > 0` with the lowercased token. `instr`, not `LIKE`, so `%` and `_` in user input stay literal. |
| status | `status IN (?, ?, ...)` |

Results come back ordered by `(source, id)` for determinism before ranking.

### 6.2 Scoring (pure Go, `internal/domain`)

Inputs: the listing, `targetBudget` (optional), and `now` (injected; never call `time.Now()`
inside domain code).

**Budget fit** (asymmetric; going over budget hurts twice as much as coming in under):

```
if price <= target:  pctUnder  = (target - price) / target
                     budgetFit = 1 - pctUnder               // equals price/target
if price >  target:  pctOver   = (price - target) / target
                     budgetFit = max(0, 1 - 2 * pctOver)    // zero at 50% over
```

Worked values for target 450,000: price 450,000 → 1.0; 405,000 (10% under) → 0.9;
495,000 (10% over) → 0.8; 562,500 (25% over) → 0.5; 675,000 and above → 0.

**Recency** (exponential decay, half-life 30 days):

```
ageDays = max(0, calendarDays(listedDate → now, UTC))   // whole days; future dates clamp to 0
recency = 0.5 ^ (ageDays / 30)
```

Worked values: 0 days → 1.0; 30 days → 0.5; 60 days → 0.25; 90 days → 0.125.

**Score** (in `[0, 1]`):

```
with budget:    score = 0.6 * budgetFit + 0.4 * recency
without budget: score = recency
```

Weights and half-life are exported constants in `domain`, not config. Rationale for the
README: budget dominates because it is the user's stated intent; recency breaks the
near-ties that real markets produce; a 30-day half-life makes "a month old counts half"
easy to state. Without a budget the ordering is strictly newest-first (recency is monotonic
in `listedDate`), which is what the status line promises (BRANDING.md §5.3, §10).

### 6.3 Ordering

Sort on the **unrounded** score, descending. Ties (exact equality) break by: `listedDate`
desc, then `price` asc, then `source` asc, then `id` asc. The order is fully deterministic,
so pages never shift between requests and tied listings show identical decimals with
sequential ranks (BRANDING.md §5.6).

### 6.4 Pagination

`totalPages = ceil(totalItems / pageSize)` (0 when `totalItems` is 0). Items for page `p`
are the slice `[(p-1)*pageSize, min(p*pageSize, totalItems))`, each carrying
`rank = index + 1` over the whole ranked set. A page past the end yields an empty slice with
the same metadata.

## 7. Frontend contract

Everything visual is specified in BRANDING.md: tokens (§2), type (§3), layout and the
pinned app shell (§4), each component (§5), motion (§6), copy and formatting (§7), don'ts
(§8). This section adds only the engineering behavior.

### 7.1 URL state

The query string is the state. Param names are identical to §5.3 so a search URL is
shareable and the API call is a direct translation. `useSearchParamsState()` reads
`window.location.search` on mount, writes with `history.pushState`, and listens to
`popstate`. Submitting the filter bar writes the new filters with `page` omitted (= 1).
Changing page writes only `page`. Clear writes an empty query string. Absent params mean
defaults. `pageSize` is never written by the UI but is preserved if present.

### 7.2 Fetch lifecycle

`useListingSearch(params)`: on every params change, abort the previous in-flight request
(`AbortController`), set `loading`, call the client, then set `data` or `error`. Aborted
requests never update state. The previous successful response stays available during
loading so the list does not blank (BRANDING.md §5.9).

```ts
type SearchState =
  | { status: 'idle' }
  | { status: 'loading'; previous?: SearchResponse }
  | { status: 'success'; data: SearchResponse }        // data.items.length > 0
  | { status: 'empty'; data: SearchResponse }          // totalItems === 0
  | { status: 'invalid'; error: ValidationError; previous?: SearchResponse }
  | { status: 'error'; error: ApiError; previous?: SearchResponse };
```

`useCities()` loads once on mount. If it fails, the City select offers only "Any" and the
console gets the error; the search still works.

### 7.3 Types and client

```ts
// src/api/types.ts  — mirrors §5.3–5.5 exactly
type ListingStatus = 'active' | 'pending' | 'sold';
interface SearchParams { minPrice?: number; maxPrice?: number; minBedrooms?: number;
  city?: string; keyword?: string; targetBudget?: number;
  status?: ListingStatus[]; page?: number; pageSize?: number }
interface ListingResult { rank: number; id: string; source: string; address: string;
  city: string; state: string; zip: string; price: number; bedrooms: number;
  bathrooms: number; sqft: number; latitude: number; longitude: number;
  listedDate: string; status: ListingStatus; description: string;
  relevance: { score: number; budgetFit: number | null; recency: number } }
interface PageMeta { page: number; pageSize: number; totalItems: number; totalPages: number }
interface ScoringMeta { targetBudget: number | null; budgetWeight: number; recencyWeight: number;
  recencyHalfLifeDays: number; asOf: string }
interface SearchResponse { items: ListingResult[]; page: PageMeta; scoring: ScoringMeta;
  appliedFilters: Record<string, unknown> }
interface FieldError { field: string; code: string; message: string }
type ValidationError = { kind: 'validation'; message: string; fields: FieldError[] };
type ApiError = ValidationError
  | { kind: 'http'; status: number; message: string }
  | { kind: 'network'; message: string };

// src/api/client.ts
searchListings(params: SearchParams, signal?: AbortSignal): Promise<SearchResponse>
getCities(signal?: AbortSignal): Promise<string[]>
```

The client builds the query string from `SearchParams`, omitting undefined values, sends
`status` as repeated params, parses non-2xx bodies into `ApiError`, and maps aborts and
fetch failures to `{ kind: 'network' }`.

### 7.4 Components (props)

Names follow BRANDING.md §5. Each is a folder with `Component.tsx`, `Component.module.css`,
`Component.test.tsx`.

```ts
Header({ scope?: string })                                   // BRANDING §5.1
FilterBar({ value: SearchParams; cities: string[]; loading: boolean;
            serverErrors?: FieldError[];
            onSearch(next: SearchParams): void; onClear(): void })   // §5.2
StatusLine({ state: SearchState })                           // §5.3
LoadingLine({ active: boolean })                             // §5.9 loading line
ListingList({ items: ListingResult[]; keywordTokens: string[] })
ListingCard({ item: ListingResult; keywordTokens: string[] }) // §5.4
Monogram({ city: string })                                   // §5.5
ScorePill({ rank: number; score: number })                   // §5.6
Pager({ page: number; totalPages: number; onChange(page: number): void }) // §5.7
NoResults({ onClear(): void })                               // §5.9
ErrorState({ onRetry(): void })                              // §5.9
SearchPage()                                                 // composes the shell, §4.2
```

FilterBar behavior: seven fields in BRANDING.md §5.2's anatomy. The seventh is **Status**, a
select with options Any / Active / Pending / Sold (label "Status", placeholder "Any"); it is
the one addition beyond the brief, chosen in the engineering interview, and it follows
§5.2's "a new filter joins the grid as a seventh field" rule. The bar keeps a local draft so
typing never fires requests. It mirrors the server's cheap rules client-side (min ≤ max,
non-negative numbers, budget > 0) using the exact strings in BRANDING.md §7.2, marks fields
per §5.2, and disables Search while invalid or loading. Enter in any field submits. Server
field errors (§5.5) are mapped by `field` + `code` onto the same marks; `min_gt_max` marks
both price fields and shows under Max price.

ListingCard: the keyword tokens come from the URL state; each case-insensitive match inside
the one-line description is wrapped in weight 600 (BRANDING.md §5.4). Formatters in
`src/format/` implement BRANDING.md §7.1 exactly and have their own unit tests.

Pager: renders every page number (the dataset is small; windowing is a later concern noted
in docs/EXTENSIONS.md). Previous/Next disabled at the bounds per §5.7. After a page change
focus moves to the results region heading.

### 7.5 States

Rendered exactly as BRANDING.md §5.9. In engineering terms:

- `loading`: StatusLine reads `Loading`, LoadingLine active, Search disabled, previous list stays.
- `success`: list + StatusLine + Pager.
- `empty`: NoResults in the results region; StatusLine `0 results`; Pager hidden.
- `invalid`: fields marked from the server's `fields`; previous list stays; StatusLine unchanged.
- `error`: ErrorState in the results region; StatusLine empty; the error object is logged to the console with its request ID when present.

### 7.6 Accessibility baseline

Every input has a `<label>`; the results region is `aria-live="polite"` and `aria-busy`
while loading; the pager uses real `<button>`s with `aria-current="page"` on the current
page; disabled Previous/Next are not focusable (BRANDING.md §5.7). The loading line has
`role="progressbar"` with an accessible name. Keyboard-only use must work end to end. Under
`prefers-reduced-motion: reduce` the loading line is static (BRANDING.md §6).

## 8. Testing requirements

**Backend** (`go test -race ./...`, table-driven, stdlib `testing` only):

- domain: no matches; tied scores (identical price and date → tiebreak order verified); each validation rule in §5.3; pagination boundaries (page 1, last partial page, page past end, pageSize 1, pageSize 100, zero items) including `rank` continuity across pages; budget equal to price; the over/under asymmetry at matching percentages; no budget; future `listedDate`; ages 0, 30, 60 days.
- store: open `:memory:`, migrate, seed the real `data/sample_listings.json`, exercise every filter including keyword tokens containing `%` and `_`, status sets, and city case-insensitivity. Seed twice and assert the row count is unchanged (idempotent).
- httpapi: `httptest` against a seeded in-memory store. Status codes, error body shape with `code`s, request ID header, CORS preflight, unknown-param rejection, page past end.

**Frontend** (`vitest`, RTL, jsdom; `fetch` stubbed with `vi.fn`):

- each component: renders from props, user interactions, correct roles/labels, BRANDING copy strings verbatim.
- formatters: every row of BRANDING.md §7.1.
- `useSearchParamsState`: round-trips every param, `popstate` handling, page reset on search, `pageSize` preserved.
- `SearchPage`: loading → success, loading → empty, validation error mapped to the right field (including `min_gt_max` marking both), network error with Try again, page change issues a new request and aborts the old one, rank numbering continues across pages.

**Lint and checks:** `golangci-lint run` (default linters + `errcheck`, `govet`, `staticcheck`, `revive`), `eslint`, `tsc -b --noEmit` (plain `tsc --noEmit` on the solution tsconfig checks no files), `prettier --check`. `make check` runs lint + tests for both halves and must pass before any merge to `master`.

## 9. Extension points (designed for, not built)

These are the likely live-session requirements. Each has a known seam. BC-16 writes them
up in `docs/EXTENSIONS.md` with the concrete files to touch.

1. **Cross-source dedupe.** A stage between `store.Search` and `domain.Rank` (§6). Key candidates: normalized address (lowercase, expand St/Street, Ave/Avenue, Unit/Apt) + zip; or lat/long within ~10 m. Keep the most recent `listedDate`, carry `sources: string[]` on the result; the card gains a `2 sources` meta item (BRANDING.md §10).
2. **Status filter.** Already built (§5.3, §7.4) because it is cheap and obvious.
3. **Geo radius.** `lat`, `lng`, `radiusKm` params → bounding-box pre-filter in SQL, haversine predicate in Go.
4. **Sort options.** `sort=relevance|price|newest`; ordering already lives in one domain function.
5. **More filters.** `minBathrooms`, `minSqft`, `zip`: one new WHERE clause, one validation rule, one seventh/eighth field in the bar.
6. **A new feed.** Seed loader takes a list of files; identity already includes `source`.
7. **Page size control.** API already supports it; BRANDING.md §5.7 says where the select goes.
8. **Scale.** Score in SQL or a precomputed column, keyset pagination, FTS5 for keyword search, pager windowing.

## 10. Docker Compose (dev)

`compose.yaml` services:

- `api`: builds `backend/Dockerfile.dev` (`golang:1.27-alpine` + `air`), bind-mounts `./backend` and `./data`, env from §11 with `DB_PATH=/data/listings.db`, port `8080:8080`, healthcheck `wget -qO- http://localhost:8080/healthz`.
- `web`: `node:24-alpine`, working dir `/app`, bind-mounts `./frontend` with an anonymous volume for `node_modules`, runs `npm ci && npm run dev -- --host`, port `5173:5173`, `VITE_API_PROXY_TARGET=http://api:8080`, `depends_on: api: condition: service_healthy`.

`vite.config.ts` proxies `/api` and `/healthz` to `VITE_API_PROXY_TARGET` (default
`http://localhost:8080` for native dev).

`make smoke` brings the stack up, waits for health, curls `/api/cities` and a search with
`targetBudget=450000&city=Springfield`, asserts non-empty items and `rank: 1` on the first,
asserts a 400 with code `min_gt_max` for `minPrice=5&maxPrice=1`, then tears down.

## 11. Configuration (environment)

| Var | Default | Notes |
|---|---|---|
| `PORT` | `8080` | |
| `DB_PATH` | `./data/listings.db` | `:memory:` allowed |
| `LISTINGS_FILE` | `./data/sample_listings.json` | seed source |
| `LOG_LEVEL` | `info` | `debug|info|warn|error` |
| `CORS_ORIGINS` | `http://localhost:5173` | comma-separated |
| `SHUTDOWN_TIMEOUT` | `10s` | graceful drain |

Logging: `slog` JSON handler to stdout. One request log line per request: method, path,
status, duration, request ID, remote addr. Panics recovered to 500 with the request ID logged.

## 12. Deliverable

A zip (`make dist`) of the repo excluding `.git/`, `node_modules/`, `*.db`, `.lattice/`,
and the two handout PDFs, containing a README that a stranger can follow to run
`docker compose up` and `make check`. Verified by extracting into a fresh directory and
running both. No emoji anywhere in the zip (BRANDING.md §8).

## 13. Notes for the branding owner

Engineering decisions that BRANDING.md v1 does not yet reflect. Each goes through
BRANDING.md §9 when convenient; none blocks implementation.

1. **Status filter** is in v1 as the seventh field (§7.4). BRANDING.md §5.2's field table should gain a "Status · select · Any / Active / Pending / Sold" row.
2. **Score field name.** BRANDING.md §5.6 says `relevanceScore`; the API field is `relevance.score` (§5.4), same `[0, 1]` range.
3. **Rank** is supplied by the API (`rank` per item) rather than computed by the UI.
4. **No budget** ordering is newest-first by construction (§6.2), so "newest first" in §5.3 is accurate.
