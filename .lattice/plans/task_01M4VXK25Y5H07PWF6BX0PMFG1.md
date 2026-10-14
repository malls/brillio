# BC-3 Plan: domain relevance scoring, deterministic ordering, pagination with ranks

Branch `feat/BC-3-domain-ranking`, cut from `feat/BC-2-backend-scaffold` @ 5da9b80. Go module
`brillio/backend` (go 1.27, toolchain go1.27.1). Paths are worktree-relative. Baseline
`make lint-api test-api` is green at 5da9b80 (golangci-lint 2.14.0: 0 issues).

## Scope

In: the ranking core in `backend/internal/domain`: SPEC §6.2 (budget fit, recency, score), §6.3
(ordering), §6.4 (pagination with global rank), the domain-side `page`/`pageSize` bounds, and
the §8 domain test list.

Out (BC-5, `httpapi.ParseSearchQuery`): every other §5.3 rule (unknown params, number/integer
parsing, minPrice/maxPrice/`min_gt_max`, minBedrooms, city/keyword length, `targetBudget > 0`,
status at the HTTP boundary, defaults page 1 / pageSize 10), rounding to four decimals, all
JSON (DTOs, `[]` vs `null`, the `scoring` block with weights 0/1 when there is no budget, `asOf`).
The status enum is already validated and tested in domain (`ParseStatus`, `TestParseStatus`, BC-2).

Untouched: every BC-2 file and type (`Listing`, `SearchFilter`, `ListingStatus`, `DateLayout`;
`listing.go`, `filter.go`, `listing_test.go`), `Makefile`, `go.mod`, `docs/`. No type change needed.

## Binding rules

- Stdlib only. Code imports `cmp`, `errors`, `fmt`, `math`, `slices`, `time`; tests add `testing`, `reflect`.
- No `time.Now()` anywhere in `internal/domain`, tests included (fixed clocks only).
- Full precision everywhere; no rounding in domain.
- Tests: external package `domain_test` (like `listing_test.go`), table-driven, stdlib `testing`,
  subtests named after §8 items exactly as listed in the table below.

## Files and API

### `backend/internal/domain/scoring.go`

Starts with a file comment that states the formula in words (BC-15 quotes it), separated from
`package domain` by one blank line. It is deliberately not a package doc: `listing.go` already
owns `// Package domain ...`; revive `package-comments` only complains about a detached comment
that begins with "Package domain", and staticcheck ST1000 is off in golangci's defaults.
Wording may be tightened, content may not change:

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

package domain
```

- `const ( BudgetWeight = 0.6; RecencyWeight = 0.4; RecencyHalfLifeDays = 30 )`, one doc comment
  per constant, all untyped. Untyped matters: `float64(age)/RecencyHalfLifeDays` is float division.
  A typed `int` would fail to compile there, and the tempting fix `age/RecencyHalfLifeDays` silently
  becomes integer division (45 days gives 1). The age-15 and age-45 recency tests catch that bug.
  BC-5 can assign them directly to float64/int response fields.
- `func BudgetFit(price, target float64) float64`: `if price <= target { return price / target }`,
  else `return max(0, 1-2*((price-target)/target))` (builtin `max`). Doc: precondition `target > 0`
  and `price >= 0`. The API rejects `targetBudget <= 0` (§5.3, BC-5). Results outside the
  precondition are unspecified: no panic, no invented behavior, no test.
- `func AgeDays(listedDate, now time.Time) int`: take each instant's UTC calendar date, then count days:
  ```go
  l, n := listedDate.UTC(), now.UTC()
  from := time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
  to := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
  return max(0, int((to.Unix()-from.Unix())/secondsPerDay)) // unexported const secondsPerDay = 86400
  ```
  Both values are UTC midnights, so the difference is an exact multiple of 86400: UTC has no DST and
  Go time has no leap seconds. Two shortcuts are wrong. `now.Sub(listedDate)/24h` truncates
  durations, so 12 h across midnight would give 0 days. `to.Sub(from)` saturates at ±292 years for a
  zero `ListedDate`. A `now` in any location or at any time of day is handled by `.UTC()` first.
- `func Recency(listedDate, now time.Time) float64`: `math.Pow(0.5, float64(AgeDays(listedDate, now))/RecencyHalfLifeDays)`.
- `func Score(budgetFit *float64, recency float64) float64`: `nil` returns `recency` unchanged;
  otherwise `BudgetWeight*(*budgetFit) + RecencyWeight*recency`. Doc: with inputs in [0,1] the
  result is in [0,1]. In float64, 0.6+0.4 rounds to exactly 1.0, so the upper bound is exact.

### `backend/internal/domain/rank.go`

```go
type RankInput struct { TargetBudget *float64; Now time.Time } // nil budget = recency only; Now injected
type Relevance struct { Score float64; BudgetFit *float64; Recency float64 } // BudgetFit nil without budget
type Scored struct { Listing; Relevance } // both embedded, per the ticket
func Rank(listings []Listing, in RankInput) []Scored
func byRelevance(a, b Scored) int // unexported comparator; the ONE place ordering lives (§9.4 seam)
```
- Doc comments on every type and every exported field, in the style of `listing.go`. No JSON tags
  on `Relevance`: these are in-memory types and BC-5 maps them to DTOs (BC-2 note: never marshal
  `Listing` or `Scored`).
- `Rank` builds a new slice, `out := make([]Scored, 0, len(listings))`, and fills it per listing:
  `rel := Relevance{Recency: Recency(l.ListedDate, in.Now)}`. If `in.TargetBudget != nil`, it adds
  `fit := BudgetFit(l.Price, *in.TargetBudget); rel.BudgetFit = &fit`, a fresh variable per item,
  never a shared pointer. Then `rel.Score = Score(rel.BudgetFit, rel.Recency)` and the item is
  appended. Last step: `slices.SortStableFunc(out, byRelevance)`.
  Contract (in the doc comment):
  - The caller's slice is never reordered or modified. `Listing` has no reference fields, so the
    copies are independent.
  - The result is never nil (empty, non-nil for nil or empty input). The wire `[]` vs `null` is
    still BC-5's job when it builds the DTO slice.
  - The order is total and deterministic because (Source, ID) is the identity.
- `byRelevance` returns `cmp.Or(cmp.Compare(b.Score, a.Score), b.ListedDate.Compare(a.ListedDate),
  cmp.Compare(a.Price, b.Price), cmp.Compare(a.Source, b.Source), cmp.Compare(a.ID, b.ID))`, one
  comment per key. It compares the unrounded score, so ties mean exact equality (§6.3).

### `backend/internal/domain/paginate.go`

```go
const MaxPageSize = 100 // exported so BC-5's parser uses the same bound
var ErrPageOutOfRange = errors.New("page out of range")         // page < 1
var ErrPageSizeOutOfRange = errors.New("pageSize out of range") // pageSize outside 1..MaxPageSize
type RankedItem struct { Scored; Rank int } // Rank = 1-based position in the WHOLE ranked set
type Page struct { Items []RankedItem; Page, PageSize, TotalItems, TotalPages int }
func Paginate(ranked []Scored, page, pageSize int) (Page, error)
```
- A field named `Page` inside type `Page`, and a field `Rank` next to func `Rank`, are legal Go and
  no enabled linter flags them.
- Validation collects every problem (§5.3 says collect all). Each bad value adds
  `fmt.Errorf("%w: got %d, want >= 1", ErrPageOutOfRange, page)` or
  `fmt.Errorf("%w: got %d, want 1..%d", ErrPageSizeOutOfRange, pageSize, MaxPageSize)`, and the
  function returns `Page{}, errors.Join(errs...)`. BC-5 checks each with `errors.Is` and maps
  ErrPageOutOfRange to field `page` and ErrPageSizeOutOfRange to field `pageSize`, both with code
  `out_of_range` (as BC-5's ticket specifies).
  - Sentinels follow the existing `ErrInvalidStatus` pattern. A struct error type was rejected:
    the Join form already carries both fields.
  - BC-5's parser validates these bounds first, so in practice this is a guard.
- Body:
  ```go
  total := len(ranked)
  totalPages := total / pageSize
  if total%pageSize != 0 {
  	totalPages++
  }
  p := Page{Items: []RankedItem{}, Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
  if page > totalPages {
  	return p, nil // past the end, or zero items: empty, non-nil Items, same metadata
  }
  start := (page - 1) * pageSize
  end := min(start+pageSize, total)
  // then fill Items[k] = RankedItem{Scored: ranked[start+k], Rank: start + k + 1}
  ```
  The `page > totalPages` check runs BEFORE any multiplication. That way `start < total` always
  holds, and `page = math.MaxInt` cannot overflow. The ceiling is computed with `/` and `%`,
  never `total+pageSize-1`.
  The `Page` field echoes the requested page even past the end (§5.3: "correct page metadata so
  the UI can recover").

## Lint notes (backend/.golangci.yml: standard + errcheck, govet, staticcheck, revive; gofmt; no exclusion presets)

- revive `exported`: every exported const/var/type/func has a doc comment that starts with its
  name.
- staticcheck in golangci v2 runs "all" except ST1000/1003/1016/1020/1021/1022 (checked in the
  2.14 binary), so QF1008 is on. Always use promoted selectors, in code and tests: `a.Score`,
  `a.Price`, `it.ID`. Never `a.Relevance.Score` or `a.Listing.Price`.
- ST1005 / revive `error-strings`: error strings are lowercase with no trailing period.
- revive `redefines-builtin-id`: no variables named `min`, `max`, `len`, `cap`, `new`.
- revive `package-comments`: see the header rule above.

## Tests

Shared helpers (in `scoring_test.go`, reused across files):
- `mustDate(s)`: `time.Parse(domain.DateLayout, s)`, panics on error.
- `fixedNow = time.Date(2026, 10, 13, 15, 30, 0, 0, time.UTC)`. The time of day is deliberate.
- `ptr(f) *float64`
- `const eps = 1e-12` and `near(a, b)`, which tests `math.Abs(a-b) <= eps`.

Exactness (checked in IEEE-754 doubles):
- Exact, asserted with `==`: all BudgetFit worked values, and Recency at whole half-lives
  (1, 0.5, 0.25, 0.125).
  - `405000/450000 == 0.9`.
  - `1-2*(45000/450000) == 0.8` (round-half-even).
  - The 25% and 50% cases are exact.
  - `math.Pow(0.5, k)` for integer `k` is exact.
- Asserted with `near`: fractional exponents (ages 15 and 45) and Score sums.
  For example, 0.6*0.9+0.4*1 = 0.9400000000000001.
- The [0,1] bound is asserted exactly.
- Asymmetry prices are integer literals (e.g. 495000), never `450000*1.1`, because
  `450000*1.1 == 495000.00000000006`.

| SPEC §8 domain item / ticket bullet | Test / subtest | File |
|---|---|---|
| budget equal to price | `TestBudgetFit/budget_equal_to_price` (450000 → 1.0) | scoring_test.go |
| worked values 405000/495000/562500/675000 | `TestBudgetFit/worked_405000_10pct_under` (0.9), `/worked_495000_10pct_over` (0.8), `/worked_562500_25pct_over` (0.5), `/worked_675000_50pct_over` (0), `/beyond_50pct_over_clamps_to_0` (900000 → 0), `/price_0` (0) | scoring_test.go |
| over/under asymmetry at matching percentages | `TestBudgetFit/asymmetry_at_matching_percentages`, which loops pairs 5% (427500/472500), 10% (405000/495000), 25% (337500/562500), 40% (270000/630000) and asserts over < under and `near(1-over, 2*(1-under))` | scoring_test.go |
| ages 0, 30, 60 (+90) days | `TestAgeDays/age_0_days` (2026-10-13), `/age_30_days` (2026-09-13), `/age_60_days` (2026-08-14), `/age_90_days` (2026-07-15); `TestRecency/age_0_days` (1), `/age_30_days` (0.5), `/age_60_days` (0.25), `/age_90_days` (0.125) | scoring_test.go |
| future `listedDate` | `TestAgeDays/future_listed_date` (2026-10-20 → 0); `TestRecency/future_listed_date` (1.0); also row 1 of the Rank tie fixture | scoring_test.go, rank_test.go |
| AgeDays semantics (ticket: whole UTC calendar days) | `TestAgeDays/calendar_days_not_24h_periods` (2026-10-12T18:00Z → now 2026-10-13T06:00Z = 1); `/now_time_of_day_ignored` (now 23:59:59.999Z and 00:00Z both give 30 for 2026-09-13); `/now_non_utc_location` (now 2026-10-13T22:00 in `time.FixedZone("EDT", -4*3600)` = Oct 14 UTC → 31 for 2026-09-13) | scoring_test.go |
| fractional exponent (not integer division) | `TestRecency/age_15_days` (≈ √0.5); `/age_45_days_spec_example` (2026-08-29 → ≈ 0.3535533905932738, SPEC §5.4 recency 0.3536) | scoring_test.go |
| no budget | `TestScore/no_budget` (nil returns recency unchanged, `==`); `TestRank/no_budget` (ages 0/30/60 given oldest-first, so the output is newest first, every BudgetFit nil, Score == Recency exactly) | scoring_test.go, rank_test.go |
| score formula with budget | `TestScore/with_budget_worked_values` (`near`): (1,1)→1.0, (0.9,0.5)→0.74, (0.8,0.25)→0.58, (0.5,0.125)→0.35, (0,1)→0.4, (1, 0.3535533905932738)→0.7414213562373095 | scoring_test.go |
| score in [0,1] | `TestScore/score_in_0_1`: fits {nil, 0, 0.5, 0.8, 0.9, 1} × recencies {0, 0.125, 0.25, 0.5, 1}, asserting `0 <= s && s <= 1` exactly | scoring_test.go |
| no matches | `TestRank/no_matches` (nil and empty input, with and without budget, give non-nil len 0); `TestPaginate/zero_items` | rank_test.go, paginate_test.go |
| tied scores (identical price and date → tiebreak order verified) | `TestRank/tied_scores_tiebreak_order` (fixture below; asserts the exact (Source, ID) sequence and that rows 1–5 have Score == 1.0 exactly) | rank_test.go |
| tiebreak with budget | `TestRank/tied_scores_with_budget`: target 450000, same date; {MLS_B,B9,405000} and {MLS_A,A2,472500} both fit exactly 0.9, so the scores are `==`. Price asc puts B9 first even though source would not | rank_test.go |
| ranking with budget | `TestRank/with_budget`: target 450000, now fixedNow. Expected order Y, X, W, Z: Y{495000, 10-13}: fit 0.8, score ≈0.88. X{450000, 08-29}: fit 1.0, rec ≈0.35355, score ≈0.74142 (the §5.4 A1 example, corrected per the run-state log). W{405000, 09-13}: 0.9, 0.5, ≈0.74. Z{675000, 10-13}: 0, 1, ≈0.4. Each BudgetFit is non-nil with its own value | rank_test.go |
| deterministic order | `TestRank/input_order_independent` (all 720 permutations of the tie fixture, via a small Heap's-algorithm helper, give the identical order); `TestRank/does_not_mutate_input` (`slices.Equal` against a clone; `Listing` is comparable) | rank_test.go |
| page 1 | `TestPaginate/page_1` (25 items, size 10 → ranks 1–10, TotalPages 3) | paginate_test.go |
| last partial page | `TestPaginate/last_partial_page` (page 3 → ranks 21–25); `/last_full_page` (20 items, page 2 → 11–20, TotalPages 2) | paginate_test.go |
| page past end | `TestPaginate/page_past_end` (page 4 → Items non-nil and empty, Page 4, PageSize 10, TotalItems 25, TotalPages 3); `/page_past_end_max_int` (page `math.MaxInt`, size 100 → empty, no panic) | paginate_test.go |
| pageSize 1 | `TestPaginate/page_size_1` (25 items, page 25 → one item, rank 25, TotalPages 25) | paginate_test.go |
| pageSize 100 | `TestPaginate/page_size_100` (150 items, page 1 → 1–100, TotalPages 2); `/page_size_100_last_page` (page 2 → 101–150) | paginate_test.go |
| zero items | `TestPaginate/zero_items` (nil input, page 1 → TotalItems 0, TotalPages 0, Items non-nil and empty, Page 1) | paginate_test.go |
| rank continuity across pages | `TestPaginate/rank_continuity_across_pages`: for sizes {1, 7, 10, 100} over 25 items, walking all pages gives ranks 1..25 contiguous with matching IDs. Explicitly, page 2 of size 10 starts at Rank 11 with ID L11 | paginate_test.go |
| §5.3 rules living in domain (page ≥ 1, pageSize 1..100) | `TestPaginateInvalid/invalid_page_0`, `/invalid_page_negative`, `/invalid_page_size_0`, `/invalid_page_size_101`, `/invalid_page_size_negative`, `/invalid_page_and_page_size` (0,0 → `errors.Is` true for BOTH sentinels). Each asserts the unexpected sentinel is absent and the returned Page is the zero value (`reflect.DeepEqual`) | paginate_test.go |
| §5.3 status enum | existing `TestParseStatus` (BC-2), unchanged | listing_test.go |

Every valid TestPaginate row asserts Page, PageSize, TotalItems and TotalPages, that Items is
non-nil, its length, and for each item both `Rank == start+k+1` and `ID == ranked[Rank-1].ID`.
The fixture is `scoredN(n)`, which gives IDs "L01".."Lnn" in ranked order.

Tie fixture: `fixedNow`, no budget, input given in reverse. Expected order, with the key that
decides each adjacent pair:

| # | Source | ID | ListedDate | Price | Score | Decides vs next (and what the next key alone would do) |
|---|---|---|---|---|---|---|
| 1 | MLS_B | B7 | 2026-10-20 (future) | 600000 | 1.0 | listedDate desc (price alone would put #2 first) |
| 2 | MLS_B | B3 | 2026-10-13 | 300000 | 1.0 | price asc (source alone would put #3 first) |
| 3 | MLS_A | Z1 | 2026-10-13 | 450000 | 1.0 | source asc (id alone would put #4 first) |
| 4 | MLS_B | B1 | 2026-10-13 | 450000 | 1.0 | id asc ("B1" < "B2"; same price and date) |
| 5 | MLS_B | B2 | 2026-10-13 | 450000 | 1.0 | score desc (price, source and id would all put #6 first) |
| 6 | MLS_A | A1 | 2026-09-13 | 100000 | 0.5 | — |

## Gate and validation

- Gate, from the repo root: `make lint-api test-api`. These run `cd backend && golangci-lint run`
  and `cd backend && go test -race ./...`; both must be green. Targets already exist (BC-2).
- Also check before handing off:
  - Prints only `brillio/backend/internal/domain`:
    `cd backend && go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain`
  - Prints nothing: `grep -rn 'time\.Now' backend/internal/domain`
- Validation: N/A. This is a pure library with no running surface. Record a one-line
  `--role validation` comment citing the PASS lines of
  `cd backend && go test -race -v ./internal/domain/...` for every subtest name above.

## Shared-file handling (run-state)

- Only new files: `backend/internal/domain/{scoring,rank,paginate}.go` and `{scoring,rank,paginate}_test.go`.
- Confirmed no edits to:
  - `Makefile`: `lint-api`/`test-api` exist.
  - `go.mod`: no dependencies.
  - BC-2 files.
  - `docs/SPEC.md`: unless the delegator adopts the SPEC edit below.
- No `.lattice/` in commits.
- BC-4 (parallel) touches `internal/store`, `go.mod`/`go.sum` and main wiring, so the two are
  disjoint. Merge `master` (never rebase) once BC-2 lands. No overlap is expected.
- Suggested commits: `feat(backend): add domain relevance scoring`, `... deterministic ranking`,
  `... pagination with global ranks`, each with its tests.

## Risks

- Lint traps: QF1008 embedded selectors, the detached-header rule, builtin shadowing. Mitigated by
  the lint notes; run `make lint-api` before review.
- Float equality: exact vs `near` is decided per assertion above. A wrong `==` on a Score sum
  would be a flaky-looking failure.
- Integer division in Recency, or a shared `BudgetFit` pointer: caught by `age_15_days` /
  `age_45_days_spec_example` and by `with_budget`'s per-item values.
- Pagination overflow: guarded by the order (check before multiply) and by `page_past_end_max_int`.

## Size estimate

- Code: about 230 LOC (scoring.go ~85 including the header, rank.go ~70, paginate.go ~75).
- Tests: about 480 LOC (scoring_test.go ~190, rank_test.go ~160, paginate_test.go ~130).

## SPEC edit (proposed; delegator decides, not made by default)

§8 — the domain list says "each validation rule in §5.3", but BC-5's ticket puts those rules
in `httpapi.ParseSearchQuery`.

- Proposed domain wording: "…; `page`/`pageSize` bounds (Paginate) and status values
  (ParseStatus); …".
- Proposed httpapi wording: add "each validation rule in §5.3".

Why: the Result Validator checks §8 by name, and as written BC-3 would look incomplete for
rules that do not belong in domain. Without the edit, BC-3 records this split as a lattice
comment on BC-3 and BC-5.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review art_01M4W2GSCK5ZH6F3NFEK578VKW: PASS, five minor findings. Triage by agent:opus-5.5-delegator.

1. **Overstated BC-5 contract (paginate.go section).** Concern: the plan says BC-5 maps the sentinels
   "as BC-5's ticket specifies"; BC-5's ticket assigns page/pageSize bounds to `httpapi.ParseSearchQuery`
   and never mentions domain errors. Resolution (obvious): read that sentence as "BC-5 *can* distinguish
   the two with `errors.Is`; its parser enforces the bounds first, so these errors are a defensive
   guard." The `Paginate` doc comment says exactly that (no claim about BC-5). The delegator records a
   comment on BC-5 naming `MaxPageSize`, `ErrPageOutOfRange`, `ErrPageSizeOutOfRange`.
2. **`RankInput.Now` zero value undocumented (rank.go section).** Resolution (obvious): the
   `RankInput.Now` field doc (and one sentence in the `Rank` doc) states: "Now must be the request
   clock; the zero time makes every listing look future-dated (age 0)." No guard, no test.
3. **Fractional-exponent literals (Tests section).** `math.Pow(0.5, 1.5)` is `0.35355339059327373` on
   this machine, not `0.3535533905932738`. Resolution (obvious): use `0.35355339059327373` as the
   literal (or `math.Sqrt(0.125)`), and keep `age_15_days`, `age_45_days_spec_example`,
   `with_budget_worked_values` and every Rank `with_budget` score on `near`. Never promote a Score sum
   or fractional recency to `==`. Reason, added to Exactness: the Go spec lets arm64 fuse
   `0.6*fit + 0.4*rec` into an FMA, so sums are compared with `near`. Exact `==` stays only where the
   plan already lists it (BudgetFit worked values, whole-half-life recency, no-budget Score == Recency,
   the tie fixtures where both sides evaluate the same expression on identical inputs, the [0,1] bound).
4. **Header "toolchain go1.27.1".** Resolution (obvious): read the header as "go 1.27 (installed
   toolchain go1.27.1)". `go.mod` stays untouched; do not add a `toolchain` directive.
5. **SPEC §8 "each validation rule in §5.3" (SPEC edit section).** Resolution (delegator decision):
   **no SPEC edit.** BC-5's ticket already requires "Tests per §8 httpapi list ... every rule returns
   400 with the right code", so the rules are owned and tested there. BC-3 covers only the §5.3 rules
   that live in domain (page >= 1, pageSize 1..100 via `Paginate`; status enum via BC-2's
   `TestParseStatus`). The split is recorded as lattice comments on BC-3 and BC-5. Do not edit
   `docs/SPEC.md` in this branch.
