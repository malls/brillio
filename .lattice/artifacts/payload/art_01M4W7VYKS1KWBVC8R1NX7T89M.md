# Code Review: BC-5 (Search API)

Verdict: PASS-WITH-NITS
Reviewed HEAD: 75530c05f2b3d551e3b482ec7cddc22ef8d657fa
Base: master e5f49bb
Reviewer: agent:opus-5.5-reviewer (fresh-context own-reviewer fallback; diff exceeds 1500 changed lines, run-state decision 03:22Z)

## Gate

- `make check` from the worktree root: **exit 0**. Tail: `golangci-lint run` -> `0 issues.`; eslint, `tsc -b --noEmit`, prettier clean; `go test -race ./...` ok for config, domain, httpapi, store, store/seed; vitest `Test Files 10 passed (10)`, `Tests 197 passed (197)`.
- `cd backend && go test -race -count=1 ./...`: exit 0, all 5 test packages ok. Counted via `-json`: 65 top-level tests and 367 subtests passed, 0 failed, 0 skipped. httpapi alone: 32 top-level, 193 subtests.
- `go test -race -count=5 -shuffle=on ./internal/httpapi/`: ok (no order dependence, no flakiness). `go vet ./...` clean, `gofmt -l .` empty.
- Worktree left clean (`git status --porcelain` empty). All probes ran in a scratch copy (`git archive HEAD`), not in the worktree.
- Intermediate commit 3954ab0, linted in a scratch copy: `golangci-lint` reports 1 issue, `unused: const fieldQuery = "query"` (query.go). HEAD is clean because 75530c0 adds `parseSearchRequest`, which uses it. Recorded here. Per the brief, it is not a reason to fail, and history does not need rewriting.

## Acceptance

1. **Response bodies match §5.4 and §5.5 (and §5.2) field for field: met.**
   - `strictRoundTrip` (search_test.go:130-154) checks both directions:
     - (a) It decodes with `DisallowUnknownFields` and requires `io.EOF` on a second `Decode`. Every response key must exist in the hand-written wire struct, with no trailing data.
     - (b) It re-marshals the struct (no `omitempty` anywhere) and `reflect.DeepEqual`s it against the body decoded into `any`. Every struct key must therefore exist in the body with the same value. A missing key fails because it re-encodes as a zero value or null. A `null` in a non-pointer field fails because it re-encodes as `0` or `""`. A key that only matched case-insensitively also fails.
   - It runs on every 200 (`search()`), every 400 (`searchInvalid()`) and every cities 200.
   - The SPEC examples are also checked independently of the wire structs:
     - `TestSearchRankedAll/SPEC section 5.4 example` compares the whole `targetBudget=450000` response, cut to its first item, with the §5.4 JSON as `map[string]any`. That covers items, page `{1,10,12,2}`, scoring and the all-null appliedFilters.
     - `TestSearchValidationErrors/SPEC section 5.5 example` byte-compares the 400 body with the compacted §5.5 JSON.
   - `assertItemsMatchSeed` proves the 15 listing fields of every returned item equal the raw seed record.
   - Cities: the exact six-name list plus a strict round trip. `{"cities":[]}` for an unseeded store and for a nil slice.
2. **`make check` green: met.** See Gate.
3. **Pipeline reads parse, search, dedupe slot, rank, paginate, respond: met.**
   - search.go:22-54 is six commented stages, each one call plus its early return, and nothing else:
     - parse (L23-28)
     - `Searcher.Search` (L30-35)
     - the dedupe-slot comment citing §9.1 (L37-38)
     - `now := a.deps.Now()` and `domain.Rank` (L40-42)
     - `domain.Paginate`, whose defensive error goes to the 500 path (L44-50)
     - `writeJSON(newSearchResponse(page, q, now))` (L52-53)
   - The same `now` feeds both Rank and `scoring.asOf`.
4. (Plan, not ticket) Validation evidence from the running API is for the validation stage. It is not reviewed here.

Happy-path numbers, recomputed by hand from SPEC §6.2 and data/sample_listings.json (and cross-checked with an independent python script over the seed):
- A1: price 450000 = target, so budgetFit 1. Listed Aug 29, so age 45 (2 + 30 + 13). Recency 0.5^1.5 = 0.353553. Score 0.6 + 0.141421 = 0.741421 -> **0.7414**.
- B8: price 527500, 77500 over, so pctOver 0.172222 and fit 1 - 0.344444 = 0.655556 -> 0.6556. Listed Aug 30, so age 44. Recency 0.5^(44/30) = e^-1.016616 = 0.361817 -> 0.3618. Score 0.393333 + 0.144727 = 0.538060 -> **0.5381**.
- B7 0.7297 (fit 0.9911, rec 0.3376) and A2 0.5551 (fit 0.6667, rec 0.3878) also match.
- The script also reproduces the 12-row budget order and the 12-row no-budget order the tests assert. It also gives `pets` -> A5, A4, A3, B7, B9, A6, A7 and `NO Pets` -> A5, A3, B9, A7.

## Findings

### Critical
None.

### Major
None.

### Minor
None.

### Nit

1. **query.go:178-186 and docs/SPEC.md:180: Go's 10000-parameter cap surfaces as `malformed_query` with a message that does not fit, and nothing else is validated.**
   - `url.ParseQuery` (Go 1.27, `urlmaxqueryparams`, default 10000) fails with "number of URL query parameters exceeded limit" before it decodes anything.
   - Probe: 10001 copies of `minPrice=1&`, then `minprice=1`, gives 400 with the single field `{query, malformed_query, "Query string is not valid URL encoding."}`.
   - So the query is valid encoding, and the SPEC's "the pairs that did decode are still validated normally" is vacuous here: no `unknown_param` appears for `minprice`.
   - It fails safe (400, nothing widened) and only pathological clients hit it.
   - Recommendation: one SPEC clause, e.g. "(also returned when the query has more than 10000 parameters)". No code change needed.
2. **docs/SPEC.md:176: "Lengths count characters" is ambiguous for mirrors.**
   - The code counts Unicode code points (`utf8.RuneCountInString`).
   - BC-9's UI will most likely count UTF-16 code units (`maxlength`, `.length`).
   - Those agree except for astral-plane characters, where the client is stricter, so a surprise 400 is impossible.
   - Recommendation: say "Unicode code points (characters)" so BC-6 and BC-9 can mirror it exactly.
3. **docs/SPEC.md:177 and query.go:201-207: the first-value rule combined with empty-as-absent is pinned by a test but not spelled out.**
   - `minPrice=&minPrice=5` is absent. The test row "empty first value is absent" covers it.
   - This is the literal composition of two SPEC rules, and it matches `URLSearchParams.get()` in the browser.
   - Recommendation: add the example to the bullet so mirrors do not read it as "first non-empty". Optional.
4. **search_test.go:312-333: `specExample` reads the first `json` fence after the `### 5.4` and `### 5.5` headings, and nothing in docs/SPEC.md says a Go test depends on those blocks.**
   - Judged acceptable (see Positive observations): drift fails loudly in `make check` with a clear message ("SPEC has no heading ..."). It is the exact check that would have caught the old 0.8731 example.
   - The §5.5 check is byte-for-byte after `json.Compact`, so reordering keys in that example, while semantically equal, would also fail it.
   - Recommendation (optional): an HTML comment beside the two examples, e.g. `<!-- asserted by backend/internal/httpapi/search_test.go -->`, so a doc editor knows.
5. **Out of scope, pre-existing (store/listings.go:35 vs query.go:288): keyword matching is case-insensitive only for ASCII.**
   - Tokens are lowercased with Go's Unicode `strings.ToLower`, but the SQL side uses SQLite's ASCII-only `lower(description)`.
   - Scenario: `keyword=École` becomes the token `école`, which never matches a description containing "École".
   - The seed is all ASCII, so nothing is observable today. The plan's Risks section already names it as BC-4/BC-16 territory.
   - Recommendation: open a backlog task (none exists on the board). Do not fix it in BC-5.

## Implementer deviations

1. **Integers of 2^53 or more are `out_of_range` (plan said > 2^53): accepted.**
   - This is the correct boundary. "9007199254740993" parses to exactly 2^53, so 2^53 is the first parsed value that may not be what the client sent. Accepting it would break the "read exactly" guarantee the rule exists for.
   - SPEC §5.3, the `maxExactInt` comment, the ParseSearchQuery doc and the tests (2^53 and 2^53+1 rejected, 2^53-1 accepted for minBedrooms and page) all agree.
   - Side note, not a finding: from 2^52 a fractional input such as `minBedrooms=4503599627370496.5` rounds to a whole float and passes as an integer. This follows from SPEC's ParseFloat-then-check-whole definition, which already treats `2.0000000000000001` as whole, and it has no practical effect.
2. **Comments say "section 5.3" rather than "§5.3": accepted.** This matches the pre-existing convention in respond.go and router.go ("docs/SPEC.md section 5.5") and keeps Go sources ASCII.
3. **SPEC edits committed first (c3e99a2): accepted.**
   - A docs-only commit makes the contract change reviewable on its own, and the code follows in the next two commits.
   - The brief window where SPEC names `malformed_query` before the code exists is internal to the branch and harmless.
   - The SPEC diff is the four intended hunks: the §5.3 bullets, the §5.4 score plus arithmetic line, the §5.5 code list and §13 note 5. There is no unrelated churn.
   - §13 note 5 is character-identical to the requested text.
   - The §5.4 line's arithmetic is correct: 0.6 x 1.0 + 0.4 x 0.3536 = 0.74144 -> 0.7414.
4. **Extra §5.3 wording: accepted.** Each addition describes the code exactly and is pinned by a test:
   - "whatever its value (`minprice=` included)": rows `minprice=` and `foo=1&foo=2&bar=`.
   - Lengths in characters: `RuneCountInString` with two-byte `é` cases at 100/101 and 200/201.
   - `;` as a `malformed_query` example: `url.ParseQuery` rejects semicolon separators. It is tested via httptest, and I confirmed it through a real TCP `httptest.NewServer`, which returns 400 `malformed_query` with an X-Request-Id.

Implementer-flagged items:
- `specExample` coupling: acceptable. See Nit 4 and Positive observations.
- 3954ab0 lint finding: confirmed (`unused` fieldQuery). Noted under Gate, not failing.
- `minPrice=&minPrice=5` as absent: acceptable. See Nit 3.

## Positive observations

- **Rules match §5.3/§5.5 throughout.** I checked each one against the code and its test row:
  - Empty value is absent. A whitespace-only city or keyword is absent before the length rule.
  - Numbers are untrimmed and must be finite, so NaN, Inf, infinity and 1e400 are `not_a_number`. Underflow such as 1e-400 reads as 0, consistent with "only non-finite rejected".
  - `-0` is normalised to +0, with a raw-body assertion that no `-0` is echoed.
  - Checks run whole before sign before magnitude, and there is no int conversion before the bounds check, so no overflow (`page=1e300` and `pageSize=1e300` are `out_of_range`).
  - `targetBudget <= 0` is `not_positive`, never `negative`, so Rank's precondition holds. Tiny (5e-324) and maximal (1.797e308) budgets produce finite scores.
  - `min_gt_max` appears only when both bounds are valid, once, on maxPrice, with min == max allowed.
  - Status accepts repeats and commas, trims pieces, skips empty ones, collapses duplicates, matches case-sensitively and echoes in canonical order. Any number of bad pieces gives one `invalid_enum`. `slices.DeleteFunc` runs on a fresh `AllStatuses()` clone, so the package array is never mutated.
  - Unknown names are case-sensitive and sorted by byte order.
  - Errors come out in a deterministic order and are all collected. The map-order stability test runs 20 iterations.
- **Messages use BRANDING §7.2 verbatim** for min_gt_max, the bedrooms negative message and not_positive. The others are plain sentences with no exclamation marks, and the test suite asserts every string exactly.
- **`malformed_query` is implemented as approved.** It uses `url.ParseQuery(r.URL.RawQuery)` rather than `r.URL.Query()`, returns at most one error, places it first, and still validates the decoded pairs. It is tested for a bare `%`, `%zz` in a name, `;`, two bad pairs, and mixed with other errors.
- **Binding upstream rules hold:**
  - Own DTOs: domain.Listing and domain.Scored are never marshaled, and `listedDate` goes through `domain.DateLayout`.
  - Rounding is `round4` in httpapi only, with the domain left unrounded for ordering.
  - Items, appliedFilters.status and cities encode as `[]`, never `null`.
  - `budgetFit` is null without a budget, and weights are 0/1 then.
  - `asOf = now.UTC()`, tested with a -07:00 clock crossing midnight, which gives 2026-10-14 and A1 recency 0.3455.
  - The 500 path logs the error with request_id, and the body is the generic envelope. Tests assert the "secret detail" text is absent.
  - X-Request-ID is asserted on search 200/400/405/500 and cities 200/500, including echo of an inbound ID.
- **Shared files are minimal:**
  - router.go: three Deps fields replace the placeholder comment, plus the `Now` default and one `HandleFunc` line per route.
  - respond.go: two consts and `internalError`.
  - main.go: three lines in the Deps literal.
  - No go.mod change, no `.lattice/` files, no emoji. The only non-ASCII added is `×`/`§` in SPEC and `é` in test data.
- **`SearchQueryParams()` hands out a fresh copy** of the same array the parser checks, so BC-6's param-sync test cannot drift from or mutate the allowlist. A test proves both properties.
- **Test quality is high:**
  - Fixed clock everywhere and no sleeps.
  - The `:memory:` store is closed in Cleanup.
  - Fakes count calls, and the "400 never reaches the store" check uses that count.
  - Expected numbers are written out literally rather than recomputed by the code under test.
- **The SPEC-example tests are a real asset.** They turn the documented contract into an executable assertion, and the branch's own §5.4 fix (0.8731 -> 0.7414) shows that kind of drift happens.
- **Readability:** comment density and voice match the surrounding package, every exported identifier has a doc comment, there is no dead code at HEAD, and the parser's small collector type keeps each rule to a few lines without over-abstraction.