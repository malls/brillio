# BC-5: Search API: strict query parsing, validation with field codes, pipeline handler, cities endpoint

Planner: agent:opus-5.5-planner, 2026-10-14. Worktree `/Users/forrest/Code/brillio-worktrees/bc-5-search-api`,
branch `feat/BC-5-search-api` (master + feat/BC-3 + feat/BC-4 already merged in). Every path below is relative to
the worktree. Contract: docs/SPEC.md §5.2-5.5, §6, §8 (httpapi), §10; BRANDING.md §7.2, §8.

## Scope

In: `GET /api/listings/search` (strict parsing, collected field errors, the §6 pipeline in one function, §5.4
mapping), `GET /api/cities` (§5.2), Deps/router/main wiring, httpapi tests, three in-branch SPEC edits (below).
Out: OpenAPI (BC-6), dedupe (§9.1, a comment slot only), any domain/store change (BC-3/BC-4 own them; their APIs
are used as they are), frontend.

## Real APIs used (as they exist on this branch)

- `store.(*Store).Search(ctx context.Context, f domain.SearchFilter) ([]domain.Listing, error)` - never nil.
- `store.(*Store).Cities(ctx context.Context) ([]string, error)` - never nil, already sorted / first-seen casing.
- `domain.Rank(listings []domain.Listing, in domain.RankInput) []domain.Scored`; `RankInput{TargetBudget *float64; Now time.Time}`.
  Precondition TargetBudget nil or > 0 (parser guarantees it).
- `domain.Paginate(ranked []domain.Scored, page, pageSize int) (domain.Page, error)`; `Page{Items []RankedItem; Page,
  PageSize, TotalItems, TotalPages int}`; `RankedItem{Scored; Rank int}`; `Scored{Listing; Relevance}`;
  `Relevance{Score float64; BudgetFit *float64; Recency float64}`. Errors only for bad page/pageSize (parser enforces first).
- `domain.SearchFilter{MinPrice, MaxPrice *float64; MinBedrooms *int; City, Keyword string; KeywordTokens []string;
  Statuses []ListingStatus}`; `domain.ParseStatus`, `domain.AllStatuses()`, `domain.MaxPageSize` (100),
  `domain.BudgetWeight` (0.6), `domain.RecencyWeight` (0.4), `domain.RecencyHalfLifeDays` (30), `domain.DateLayout`.
- `httpapi.FieldError{Field, Code, Message string}` already exists in respond.go (json tags field/code/message); reuse it.
  `a.writeJSON`, `a.writeError`, `RequestIDFrom` exist.

## Files

| File | Holds |
|---|---|
| `backend/internal/httpapi/query.go` (new) | Param name consts, `SearchQueryParams()`, field-error code consts, message builders, `SearchQuery`, `ParseSearchQuery`, a small unexported `parser` collector, `parseSearchRequest(*http.Request)` (url.ParseQuery + ParseSearchQuery). |
| `backend/internal/httpapi/search.go` (new) | `Searcher` interface, the pipeline handler `searchListings`, response DTOs, `newSearchResponse`, `round4`. |
| `backend/internal/httpapi/cities.go` (new) | `CityLister` interface, `citiesResponse` DTO, `cities` handler. |
| `backend/internal/httpapi/router.go` (edit) | Deps fields `Searcher Searcher`, `Cities CityLister`, `Now func() time.Time` replace the placeholder comment; `if deps.Now == nil { deps.Now = time.Now }`; two `mux.HandleFunc` lines. Nothing else. |
| `backend/internal/httpapi/respond.go` (edit) | Add `codeValidation = "validation_error"`, `msgValidation = "Invalid search parameters."` to the const block and `func (a *api) internalError(w, r, err error)` (logs `request failed` with request_id + err, writes the generic 500). |
| `backend/cmd/api/main.go` (edit) | Three lines in the existing `httpapi.Deps{...}` literal: `Searcher: st, Cities: st, Now: time.Now,`. |
| `backend/internal/httpapi/query_test.go` (new) | Parser unit tests (external package `httpapi_test`). |
| `backend/internal/httpapi/search_test.go` (new) | Handler tests over a seeded `:memory:` store, wire mirror structs, `strictRoundTrip` helper. |
| `backend/internal/httpapi/cities_test.go` (new) | Cities tests. |
| `docs/SPEC.md` (edit) | The three SPEC edits below, localized lines only (union rule). |

## Signatures

```go
// query.go
func SearchQueryParams() []string // fresh slice, §5.3 table order:
// minPrice, maxPrice, minBedrooms, city, keyword, targetBudget, status, page, pageSize
type SearchQuery struct {
    Filter       domain.SearchFilter // Statuses never empty; KeywordTokens nil when keyword absent
    TargetBudget *float64            // nil or > 0
    Page         int                 // default 1
    PageSize     int                 // default 10
}
func ParseSearchQuery(v url.Values) (SearchQuery, []FieldError) // errs nil when valid
func parseSearchRequest(r *http.Request) (SearchQuery, []FieldError)

// search.go
type Searcher interface { Search(ctx context.Context, f domain.SearchFilter) ([]domain.Listing, error) }
// cities.go
type CityLister interface { Cities(ctx context.Context) ([]string, error) }
```

**BC-6 reads `httpapi.SearchQueryParams()`** for its param-sync test (works from package `httpapi` or `httpapi_test`).
A function, not an exported var, so no test can mutate the allowlist. ParseSearchQuery's allowlist check uses the same
backing array, so the two cannot drift.

DTOs (unexported; no omitempty anywhere; never embed domain.Listing/Scored):

```go
type searchResponse struct { Items []searchItem `json:"items"`; Page pageMeta `json:"page"`;
    Scoring scoringMeta `json:"scoring"`; AppliedFilters appliedFilters `json:"appliedFilters"` }
type searchItem struct { Rank int `json:"rank"`; ID, Source, Address, City, State, Zip string (id, source, address,
    city, state, zip); Price float64; Bedrooms int; Bathrooms float64; Sqft int; Latitude, Longitude float64;
    ListedDate string `json:"listedDate"` /* l.ListedDate.Format(domain.DateLayout) */; Status string;
    Description string; Relevance relevance `json:"relevance"` }
type relevance struct { Score float64 `json:"score"`; BudgetFit *float64 `json:"budgetFit"`; Recency float64 `json:"recency"` }
type pageMeta struct { Page, PageSize, TotalItems, TotalPages int } // page, pageSize, totalItems, totalPages
type scoringMeta struct { TargetBudget *float64 `json:"targetBudget"`; BudgetWeight, RecencyWeight float64;
    RecencyHalfLifeDays int; AsOf string `json:"asOf"` }
type appliedFilters struct { MinPrice, MaxPrice *float64; MinBedrooms *int; City, Keyword *string;
    Status []string `json:"status"` }
type citiesResponse struct { Cities []string `json:"cities"` }
```

`round4(x) = math.Round(x*1e4) / 1e4`, applied in httpapi to score, budgetFit (when non-nil) and recency; domain stays
unrounded. `Items` is built with `make([]searchItem, 0, len(page.Items))` so it is `[]`, never `null`. Without a budget:
`TargetBudget nil`, `BudgetWeight 0`, `RecencyWeight 1`; with one: `domain.BudgetWeight`, `domain.RecencyWeight`.
`AsOf = now.UTC().Format(domain.DateLayout)`, where `now` is the same value passed to Rank. appliedFilters: city is the
trimmed value or null, keyword the value as received or null, status the effective set (canonical order).

## The pipeline function (acceptance 3)

`func (a *api) searchListings(w http.ResponseWriter, r *http.Request)` in search.go, one doc comment citing §6, then
six commented stages and nothing else, each a single call plus its early return:

1. `// Parse and validate (§5.3): every problem is reported at once.` `q, errs := parseSearchRequest(r)`; non-empty
   errs -> `a.writeError(w, r, 400, codeValidation, msgValidation, errs)`.
2. `// Filter in SQL (§6.1).` `listings, err := a.deps.Searcher.Search(r.Context(), q.Filter)`; err -> `a.internalError`.
3. `// Dedupe slot (§9.1): merging cross-source duplicates goes here, between search and rank. Not built in v1.`
4. `// Score and order (§6.2, §6.3).` `now := a.deps.Now()`; `ranked := domain.Rank(listings, domain.RankInput{TargetBudget: q.TargetBudget, Now: now})`.
5. `// Paginate and assign ranks (§6.4).` `page, err := domain.Paginate(ranked, q.Page, q.PageSize)`; err -> `a.internalError`
   (unreachable: the parser enforced the bounds; a defensive guard, so a 500 not a 400).
6. `// Respond (§5.4).` `a.writeJSON(w, r, http.StatusOK, newSearchResponse(page, q, now))`.

`cities` handler: `a.deps.Cities.Cities(r.Context())` -> 200 `citiesResponse` (nil guarded to `[]`), err -> internalError.
Query params on /api/cities are ignored (strictness is a §5.3 search rule). A nil Searcher/Cities is a wiring bug: the
route panics and the existing recovery answers 500; documented on the Deps fields. Existing tests that build Deps
without them are unaffected.

## Parsing rules (ParseSearchQuery)

- Scalar params use `v[name][0]`; `""` is absent (§5.3). Repeated scalars: first value wins (see open question 2).
- Unknown names (case-sensitive; `minprice` is unknown) -> `unknown_param`, one per distinct name, whatever the value.
- Numbers: `strconv.ParseFloat(s, 64)`; error or non-finite (`NaN`, `Inf`, `1e400`) -> `not_a_number`; `-0` becomes 0.
  No trimming of numbers (`" 5"` is not_a_number).
- Integers: number rules, then `math.Trunc(v) != v` -> `not_an_integer` (`2.0` and `1e2` are whole); magnitude
  above `1<<53` -> `out_of_range` (SPEC edit 2), checked after the sign/lower-bound rule. Convert to int only after.
- city: `strings.TrimSpace`; empty after trim = absent; `utf8.RuneCountInString > 100` -> `too_long`.
- keyword: whitespace-only = absent; raw `utf8.RuneCountInString > 200` -> `too_long`; tokens =
  `strings.Fields(strings.ToLower(raw))`; `%` and `_` stay literal; `Filter.Keyword = raw`.
- status: every value of every `status` param, split on `,`, each piece trimmed, empty pieces skipped; each via
  `domain.ParseStatus` (case-sensitive). Any invalid piece -> one `invalid_enum` for `status`. Valid set deduped and
  emitted in `domain.AllStatuses()` order; none left -> `AllStatuses()`.
- min_gt_max: only when minPrice and maxPrice both parsed without error and `min > max` (equal is valid); reported once,
  on `maxPrice`.
- targetBudget `<= 0` -> `not_positive` (never `negative`), so Rank's precondition holds.
- Defaults page 1, pageSize 10. pageSize bound uses `domain.MaxPageSize`.
- Handler level: `url.ParseQuery(r.URL.RawQuery)` error (bare `%`, `%zz`, `;`) -> `malformed_query` on field `query`,
  plus whatever the successfully parsed pairs yield (SPEC edit 2). `r.URL.Query()` is not used: it drops bad pairs silently.

**Error order (deterministic):** `malformed_query` first if present, then at most one error per known param in
`SearchQueryParams()` order (the parser walks params in that order, so min_gt_max lands at maxPrice's slot), then
`unknown_param` errors sorted by name (byte order). This reproduces the §5.5 example order exactly.

## Error-code rules

Labels are BRANDING §7.2 field labels; Status from SPEC §7.4; Page / Page size have no label, plain words.

| Field | Code | Condition | Message | Source |
|---|---|---|---|---|
| minPrice / maxPrice | not_a_number | not a finite number | `Min price must be a number.` / `Max price must be a number.` | plain, §7 voice |
| minPrice / maxPrice | negative | `< 0` | `Min price must be 0 or more.` / `Max price must be 0 or more.` | plain, patterned on §7.2 bedrooms |
| maxPrice | min_gt_max | both valid, min > max | `Min price must be less than max price.` | BRANDING §7.2 verbatim |
| minBedrooms | not_a_number | as above | `Min bedrooms must be a number.` | plain |
| minBedrooms | not_an_integer | not whole | `Min bedrooms must be a whole number.` | plain |
| minBedrooms | negative | `< 0` | `Min bedrooms must be 0 or more.` | BRANDING §7.2 verbatim |
| minBedrooms | out_of_range | `> 2^53` | `Min bedrooms is too large.` | plain |
| city | too_long | > 100 chars after trim | `City must be 100 characters or fewer.` | plain |
| keyword | too_long | > 200 chars | `Keyword must be 200 characters or fewer.` | plain |
| targetBudget | not_a_number | as above | `Target budget must be a number.` | plain |
| targetBudget | not_positive | `<= 0` | `Target budget must be more than $0.` | BRANDING §7.2 verbatim |
| status | invalid_enum | any piece not active/pending/sold | `Status must be active, pending or sold.` | plain |
| page | not_a_number / not_an_integer | as above | `Page must be a number.` / `Page must be a whole number.` | plain |
| page | out_of_range | `< 1` / `> 2^53` | `Page must be 1 or more.` / `Page is too large.` | plain |
| pageSize | not_a_number / not_an_integer | as above | `Page size must be a number.` / `Page size must be a whole number.` | plain |
| pageSize | out_of_range | outside 1..100 | `Page size must be between 1 and 100.` | SPEC §5.5 example verbatim |
| (name as sent) | unknown_param | name not in allowlist | `Unknown parameter.` | SPEC §5.5 example verbatim |
| query | malformed_query | url.ParseQuery error | `Query string is not valid URL encoding.` | plain (new, SPEC edit 2) |

Envelope: 400 `{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[...]}}`. No emoji, no
exclamation marks anywhere (BRANDING §8). Implementer may build messages from label templates, but tests assert every
string above exactly.

## SPEC edits (in-branch, then a `lattice comment` on BC-5 per SPEC's deviation rule)

1. **SPEC edit: §5.4 - the example score is wrong.** Change `"score": 0.8731` to `"score": 0.7414` and add one line
   under the example: "The example is A1 for `targetBudget=450000` and no other filters at asOf 2026-10-13: budgetFit
   1.0, recency 0.5^(45/30) = 0.3536, score 0.6 x 1.0 + 0.4 x 0.3536 = 0.7414." Every other number in the example was
   checked against §6.2 and the seed and is right: A1's 15 listing fields match `data/sample_listings.json`; age 45
   days (Aug 29 -> Oct 13) gives recency 0.353553 -> 0.3536; price = target gives budgetFit 1.0; A1 is rank 1 of the
   unfiltered budget ranking (0.741421 vs B7 0.729702); totalItems 12 / totalPages 2 at pageSize 10 matches 12 seed
   rows with no filters; scoring 450000 / 0.6 / 0.4 / 30 / 2026-10-13 and all-null appliedFilters with three statuses
   match. Only cosmetic: Go encodes `budgetFit` 1.0 as `1` (same JSON number; no edit).
2. **SPEC edit: §5.3 parsing rules - make the decisions above part of the contract** (BC-6 documents them, BC-9 mirrors
   them). Add bullets: non-finite numbers are `not_a_number`; integers above 2^53 are `out_of_range` (past that a
   float64 cannot hold every whole number, so the value cannot be read exactly); whitespace-only `city`/`keyword` is
   absent; `status` pieces are trimmed, empty pieces ignored, duplicates collapsed, matched case-sensitively and echoed
   in the order active, pending, sold; a repeated non-status parameter uses its first value; a query string that is not
   valid URL encoding (for example a bare `%`) is a 400 with field `query` and code `malformed_query`, because the
   decoder would otherwise drop the pair and silently widen the results (the same reason unknown params are a 400);
   field errors are ordered malformed_query, then the table order with at most one per parameter, then unknown params
   by name.
3. **SPEC edit: §5.5 - add `malformed_query`** to the field error code list (consequence of edit 2).

## Tests (stdlib `testing`, table-driven, package `httpapi_test`, `go test -race`)

Fixture: `newSearchRouter(t, now)` opens `store.Open(t.Context(), ":memory:")`, seeds `seed.Load("../../../data/sample_listings.json")`
+ `Upsert`, closes at cleanup, and returns `NewRouter(Deps{Logger: JSON to a buffer, CORSOrigins: testOrigin, Pinger,
Searcher, Cities: st, Now: func() time.Time { return now }})`. Fixed clock `2026-10-13T15:04:05Z` (asOf 2026-10-13).
Compile-time checks: `var _ httpapi.Searcher = (*store.Store)(nil)`, `var _ httpapi.CityLister = (*store.Store)(nil)`.

**Strict decoding both ways (acceptance 1):** `strictRoundTrip(t, body, &v)`: (a) `json.Decoder` with
`DisallowUnknownFields` into `v`, then a second `Decode` must return `io.EOF` - every response key exists in the struct;
(b) `json.Marshal(v)`, unmarshal both the original body and the re-marshaled bytes into `any`, and `reflect.DeepEqual` -
every struct key (null included; no omitempty in the test structs) exists in the response with the same value. Test
structs mirror §5.4/§5.5/§5.2 independently (`wireSearchResponse`, `wireItem` with all 17 item keys, `wireRelevance`,
`wirePage`, `wireScoring`, `wireFilters`, `wireEnvelope` with non-omitempty `fields`, `wireCities`). Applied to every
200 and 400 body below.

Expected numbers, Springfield + targetBudget 450000, asOf 2026-10-13 (ages: Aug 27 -> 47, Aug 29 -> 45, Aug 30 -> 44,
Sep 2 -> 41 days; recency 0.5^(age/30); over-budget fit 1 - 2*(price-450000)/450000):

| rank | id | budgetFit | recency | score (0.6*fit + 0.4*rec) |
|---|---|---|---|---|
| 1 | A1 | 1 | 0.353553 -> 0.3536 | 0.6 + 0.141421 = 0.741421 -> **0.7414** |
| 2 | B7 | 1 - 2*(2000/450000) = 0.991111 -> 0.9911 | 0.337587 -> 0.3376 | 0.594667 + 0.135035 = 0.729702 -> **0.7297** |
| 3 | A2 | 1 - 2*(75000/450000) = 0.666667 -> 0.6667 | 0.387786 -> 0.3878 | 0.4 + 0.155114 = 0.555114 -> **0.5551** |
| 4 | B8 | 1 - 2*(77500/450000) = 0.655556 -> 0.6556 | 0.361817 -> 0.3618 | 0.393333 + 0.144727 = 0.538060 -> **0.5381** |

page `{1,10,4,1}`; scoring `{450000, 0.6, 0.4, 30, "2026-10-13"}`; appliedFilters `{null,null,null,"Springfield",null,["active","pending","sold"]}`.
No budget, all 12, newest first: A5 .4061, A4 .3969, A2 .3878, A3 .3789, B8 .3618, A1 .3536, B7 .3376, B9 .3223,
B10 .2872, B11 .2558, A6 .2279, A7 .1688. Budget 450000, all 12: A1, B7, A5, A3, B11, B9, A6, A2, B8, A7, B10, A4.

| Test | Covers |
|---|---|
| query_test `TestParseSearchQueryDefaults` | no params: Page 1, PageSize 10, all statuses, nil pointers/tokens, no errors |
| query_test `TestParseSearchQueryRules` | one row per rules-table line (exact field/code/message), plus valid boundaries: minPrice=0, min == max, minBedrooms=0, `2.0`, `1e2`, page=1, pageSize=1 and 100, `-0`; empty values absent; whitespace city/keyword absent; NaN/Inf/1e400 not_a_number; 2^53+ out_of_range (ticket: every rule, right code) |
| query_test `TestParseSearchQueryCollectsAll` | many bad params at once -> exact ordered slice; two unknowns sorted; run 20 times for map-order stability (ticket: all errors collected) |
| query_test `TestParseSearchQueryStatus` | repeats, commas, spaces, dupes, empty pieces, canonical order, one bad piece among good -> one invalid_enum |
| query_test `TestParseSearchQueryKeyword` | tokens lowercased on any whitespace; `%`/`_` literal; 200 vs 201 runes incl. multibyte; Filter.Keyword raw |
| query_test `TestSearchQueryParams` | §5.3 order; mutating the result does not change the next call (BC-6 contract) |
| search_test `TestSearchRankedSpringfield` | table above exactly, rank order, page/scoring/appliedFilters, strict round trip (ticket happy path; acceptance 1) |
| search_test `TestSearchNoBudget` | newest-first order above, budgetFit null, scoring `{null,0,1,30,asOf}` |
| search_test `TestSearchAsOfUTC` | clock `2026-10-13T23:30:00-07:00` -> asOf `2026-10-14`, A1 recency 0.5^(46/30) -> 0.3455 |
| search_test `TestSearchPaging` | pageSize=3&page=2 -> ranks 4-6 A3,B8,A1, totalPages 4; page=2 -> ranks 11,12 A6,A7; page=99 -> 200, body has `"items":[]`, page `{99,10,12,2}`; city=Springfield&page=2 -> `{2,10,4,1}` (§8 page past end; ticket) |
| search_test `TestSearchEmptyCity` | city=Atlantis -> 200, `{1,10,0,0}`, echo "Atlantis"; `city=%20springfield%20` -> 4 items, echo "springfield" (ticket) |
| search_test `TestSearchStatusFilter` | pending -> [A7]; sold -> 0; `status=active&status=pending` and `status=pending,active` -> 12, echo ["active","pending"] (ticket) |
| search_test `TestSearchKeyword` | `keyword=%25` -> 0, echo "%"; `keyword=_` -> 0; pets -> 7; `NO+Pets` -> A3,B9,A5,A7 (ticket: keyword with %) |
| search_test `TestSearchValidationErrors` | over HTTP: 400, JSON content type, validation_error envelope, exact fields for: every code; `minprice=1` unknown_param; `keyword=100%` malformed_query; `minPrice=5&maxPrice=1&pageSize=0&minprice=1` equals the §5.5 example body verbatim (§8 status codes, error body shape, unknown-param rejection) |
| search_test `TestSearchStoreFailure` | fake Searcher error -> 500 internal_error, generic message, error text absent from body, log line with request_id and err |
| search_test `TestSearchMethodAndCORS` | POST -> 405 Allow `GET, HEAD`; OPTIONS preflight from testOrigin -> 204 with allow headers; GET with Origin -> ACAO + Expose (§8 CORS preflight) |
| search_test `TestSearchRequestID` | X-Request-ID on 200, 400, 405, 500 search and 200/500 cities; inbound ID echoed (§8 request ID; ticket) |
| cities_test `TestCities` / `TestCitiesEmpty` / `TestCitiesFailure` | exact six cities + strict round trip; unseeded store -> `{"cities":[]}`; fake error -> 500 (§5.2) |

Existing tests stay untouched and green (TestErrorEnvelopes' `/api/listings` stays 404). Lint: revive needs doc
comments on every exported identifier and `_` for unused params; gofmt.

## Gate

`make check` (lint-api, lint-web, test-api, test-web) green on the exact branch head after merging current `master`.
Per run-state decision 02:31Z, BC-5 may plan and implement now but **may not enter `review` until BC-3 and BC-4 have
both landed**; then merge `master` (never rebase) and re-run `make check`.

## Validation (running system, port 18005)

1. `lsof -nP -iTCP:18005 -sTCP:LISTEN` is empty. Build: `cd backend && go build -o bin/api ./cmd/api` (bin/ is ignored).
2. From the worktree root (main.go reads PORT/DB_PATH/LISTINGS_FILE via config.Load, seeds from LISTINGS_FILE at start):
   `PORT=18005 DB_PATH=:memory: LISTINGS_FILE=./data/sample_listings.json ./backend/bin/api > /tmp/brillio-bc5-api.log 2>&1`
   in the background; poll `curl -fsS localhost:18005/healthz` until ok; log shows `seeded 12 listings`.
3. `curl -sS -D - 'http://localhost:18005<path>'` for each, recording status, X-Request-ID and body:
   §10 smoke: `/api/cities`; `/api/listings/search?targetBudget=450000&city=Springfield` (items, rank 1 = A1);
   `?minPrice=5&maxPrice=1` (400 min_gt_max on maxPrice). Ticket extras: `?status=pending` (A7 only); `?keyword=%25`
   (0 items, echo "%"); `?pageSize=0` (out_of_range); `?minprice=1` (unknown_param). Then one request per field code
   (`minPrice=abc`, `minBedrooms=2.5`, `minPrice=-1`, `targetBudget=0`, `page=0`, 101-char city, 201-char keyword,
   `status=foo`, `keyword=100%`), the §5.5 combined request (body equals the example), empty values
   (`minPrice=&city=` -> same as no params), page past end (`page=99`), empty city (`city=Atlantis`), status repeat vs
   comma (identical bodies), no-budget ordering (no params: A5, A4, A2, A3, B8, A1, B7, B9, B10), `pageSize=3&page=2`
   (ranks 4-6 A3, B8, A1), `-H 'X-Request-ID: bc5-check'` echoed, `-X POST` 405, preflight 204.
4. Numbers: the live server uses time.Now, so its asOf is today's UTC date (2026-10-14 after 00:00Z), not the test's.
   Read `scoring.asOf` from the response and recompute with the §6.2 formula over the seed (python3 one-liner) and
   compare every relevance value. Expected Springfield at asOf 2026-10-14: A1 0.7382 (fit 1, rec 0.3455), B7 0.7266
   (0.9911, 0.3299), A2 0.5516 (0.6667, 0.3789), B8 0.5348 (0.6556, 0.3536); at 2026-10-13 the table above (A1 0.7414).
   The fixed-clock 0.7414 is proven by `go test -run TestSearchRankedSpringfield -v ./internal/httpapi` output, included.
5. Record responses with `lattice comment BC-5 --role validation` (or attach the transcript file). Kill the server,
   confirm port 18005 is free. No `timeout` on host: `perl -e 'alarm shift; exec @ARGV' 600 <cmd>` if needed.

## Shared files (run-state table)

- `router.go`, `main.go`: owned by BC-2; BC-4 already wrote main.go (store + Ping); BC-5 is the next serial writer;
  BC-6 adds the openapi route after. Edit only the Deps fields (replacing the placeholder comment), the Now default and
  two HandleFunc lines in router.go, and three Deps-literal lines in main.go, so BC-6's one-line route merges cleanly.
- `docs/SPEC.md`: union rule; edit only the §5.3 bullet list, the §5.4 example line + one note line, and the §5.5 code
  list. Comment on BC-5 listing the edits.
- Never commit `.lattice/` on the feature branch. `internal/domain` and `internal/store` untouched.

## Risks

- Contract addition `malformed_query` (open question 1). If declined: drop it, keep `r.URL.Query()`, document the
  silent drop as a known limitation in the SPEC instead.
- SQLite `lower()` is ASCII-only while Go lowercases tokens with Unicode rules; a non-ASCII uppercase letter in a
  description would not match. Seed is ASCII; BC-4 territory; note for BC-16 extensions, not fixed here.
- Query city is only trimmed (SPEC), while the seed collapses internal whitespace; `Fall  Church` would not match. UI
  sends select values, so harmless; not changed.
- Strings with invalid UTF-8 echo as U+FFFD (encoding/json); acceptable. Verified: `url.ParseQuery("keyword=100%&city=Reston")`
  returns only `city` plus an error, which is why edit 2 exists.

## Size

Production ~430 LOC (query.go ~210, search.go ~150, cities.go ~35, router/respond/main ~35); tests ~750 LOC; SPEC ~12
lines. Total ~1200.

## Acceptance

- Response bodies match §5.4/§5.5 (and §5.2) field for field: strictRoundTrip on every 200/400 body.
- `make check` green on the exact head after merging master.
- searchListings reads top to bottom: parse, search, dedupe slot comment, rank, paginate, respond; nothing else in it.
- Validation evidence from the running API on :18005 recorded with `--role validation`.

## Open questions

1. `malformed_query` adds a field-error code to the contract (SPEC edit 2/3). Needed so a bare `%` cannot silently drop a
   filter. Default: build it. Orchestrator ack wanted before BC-6 freezes the enum.
2. Repeated scalar params (`minPrice=1&minPrice=2`): §5.5 has no code; plan uses the first value. A 400 would need a new
   code - product call; default is first value.
3. BRANDING §7.2 `Min price must be less than max price.` vs SPEC rule `maxPrice >= minPrice` (equal allowed). Plan keeps
   the rule and the string verbatim; equal prices pass. Wording is for the branding owner (BRANDING §9), not blocking.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review: art_01M4W58G75R2QWP2A8RQZQAR1H (PASS). Triage by agent:opus-5.5-delegator.

| # | Finding | Bucket | Resolution | Section |
|---|---|---|---|---|
| 1 | MAJOR: `malformed_query` on pseudo-field `query` is a contract addition (§5.3/§5.5) that BC-6/BC-9 mirror | complex (Orchestrator decision) | Routed to the Orchestrator for an explicit ack/decline before `in_progress`; the outcome is recorded below as "Decision 1". If acked, SPEC edit 2 also says: at most one `malformed_query` error per request (Go reports only the first bad pair), and the pairs that did decode are still validated normally. If declined: keep `r.URL.Query()`, drop `malformed_query` from the rules table, SPEC edits 2/3 and every test/curl that uses `keyword=100%`, and add a SPEC §5.3 known-limitation line instead. | Parsing rules, Error-code rules, SPEC edits 2-3, Tests, Validation |
| 2 | MINOR: whitespace-only keyword vs raw length > 200: precedence unstated | obvious | Absence first for both `city` and `keyword`: a value that is empty after `strings.TrimSpace` is absent, full stop (201 spaces in keyword is absent, no error). Then the length rule: trimmed city > 100 runes, raw keyword > 200 runes -> `too_long`. Add a 201-space keyword row to `TestParseSearchQueryKeyword`; mirror the sentence in SPEC edit 2. | Parsing rules, Tests, SPEC edit 2 |
| 3 | MINOR: `NO+Pets` expectation listed in seed order | obvious | Expected order is rank (no-budget, newest first): **A5, A3, B9, A7**. Assert the ordered slice. | Tests (`TestSearchKeyword`) |
| 4 | MINOR: `-0` normalisation needs a mechanism and a raw-body assertion | obvious | After a successful float parse: `if v == 0 { v = 0 }` (literal assignment yields +0). The `-0` test row asserts the raw response body contains `"minPrice":0` and not `-0`, in addition to the decoded value. | Parsing rules, Tests |
| 5 | MINOR: ParseFloat grammar wider than "a number" | obvious | SPEC edit 2 gains one sentence: "Number syntax is exactly Go's `strconv.ParseFloat` (decimal or hex float literals, underscores allowed); only non-finite results are rejected." Add one test row pinning `minPrice=1_000` -> 1000. | SPEC edit 2, Tests |
| 6 | MINOR: validation-plan item 18 expects 11 for `status=active,pending` (correct is 12) | out of plan (Orchestrator file) | Plan is right (12). Mailed the Orchestrator to correct `.lattice/orchestration/validation-plan.md` item 18; BC-5 does not edit that file. | Validation |

Planner open questions 2 and 3: (2) repeated scalar params use the first value, documented in SPEC edit 2, no new code; (3) BRANDING §7.2 "Min price must be less than max price." stays verbatim with equal prices valid; wording is the branding owner's call (BRANDING §9), noted to the Orchestrator, not blocking.

Decision 1 (Orchestrator ACK, mail 01M4W6CS273Z25KP843BMQK92J, 2026-10-14T03:13Z, Moderate autonomy, logged in run-state): **build `malformed_query`** on pseudo-field `query`: 400, at most one per request, the pairs that did decode are still validated normally. Add it in-branch to SPEC §5.3 (edit 2) and the §5.5 code list (edit 3). Message: a plain sentence in the BRANDING §7 voice (`Query string is not valid URL encoding.` stands).

Decision 2 (Orchestrator, same mail): **SPEC edit 4: §13 - add note 5** verbatim, in-branch:
`5. **Min/max price copy.** BRANDING §7.2 says "Min price must be less than max price." while the API allows minPrice == maxPrice (SPEC §5.3); the UI keeps the string verbatim and only flags min > max. Suggested copy: "Min price must not exceed max price."`
(Formatting: bold lead-in to match notes 1-4.) The API message for `min_gt_max` stays the BRANDING string verbatim.

Decision 3: repeated scalar params use the first value (Orchestrator: fine). BC-3 and BC-4 have both landed; branch merged master e5f49bb at cd4c46a. Review base is `master`.
