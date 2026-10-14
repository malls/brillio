# BC-3 validation: HEAD cdec405302066c3d1c4014fdaea583def4b43da2

- Branch: `feat/BC-3-domain-ranking` (worktree `/Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking`)
- Date: 2026-10-13
- Validator: `agent:opus-5.5-validator` (go1.27.1 darwin/arm64, golangci-lint 2.14.0)
- **Verdict: PASS**

Justification: N/A (pure library, no running surface). BC-3 adds only pure functions in `internal/domain`, with no binary, endpoint or UI. Its behavior is proven by `go test -race -count=1 -v ./internal/domain/...` at cdec405: 66 PASS / 0 FAIL / 0 SKIP (10 top-level tests, 56 subtests). Every §8 domain case is present by name and passing (table below). The full gate `GOFLAGS=-count=1 make check` exits 0 (`0 issues.`, eslint/tsc/prettier clean, go `ok` x3, vitest 2 files / 4 tests passed).

Revalidation: earlier validations passed at e9144d6 (art_01M4W3SQP5AMAVCHKSXAM40PPX) and ee09340 (art_01M4W3XEZH9E0BBPZFWTQFRJR8). cdec405 has a single parent, ee09340, and changes only the doc comment on `AgeDays`. `git diff --stat ee09340 HEAD` shows `backend/internal/domain/scoring.go | 7 ++++---` (4 insertions, 3 deletions), and every changed line is a `//` comment line. The comment now says that `now` may be in any location and that `listedDate` is expected to be midnight UTC, as parsing with `DateLayout` produces. No code changed.

## SPEC §8 domain items mapped to subtests (all from the run below)

| §8 item | Subtest(s) | Result |
|---|---|---|
| No matches | TestRank/no_matches, TestPaginate/zero_items | PASS |
| Tied scores, tiebreak order verified | TestRank/tied_scores_tiebreak_order, TestRank/tied_scores_with_budget, TestRank/input_order_independent | PASS |
| §5.3 rules that live in domain | TestPaginateInvalid/{invalid_page_0, invalid_page_negative, invalid_page_size_0, invalid_page_size_101, invalid_page_size_negative, invalid_page_and_page_size}; status enum: TestParseStatus (BC-2). The other §5.3 rules belong to BC-5 (decision recorded on BC-3/BC-5). | PASS |
| Pagination boundaries | TestPaginate/{page_1, last_partial_page, page_past_end, page_size_1, page_size_100, zero_items} | PASS |
| Rank continuity across pages | TestPaginate/rank_continuity_across_pages | PASS |
| Budget equal to price | TestBudgetFit/budget_equal_to_price | PASS |
| Over/under asymmetry at matching percentages | TestBudgetFit/asymmetry_at_matching_percentages | PASS |
| No budget | TestScore/no_budget, TestRank/no_budget | PASS |
| Future listedDate | TestAgeDays/future_listed_date, TestRecency/future_listed_date | PASS |
| Ages 0, 30, 60 days | TestAgeDays/{age_0_days, age_30_days, age_60_days}, TestRecency/{age_0_days, age_30_days, age_60_days} | PASS |
| Ticket worked values 405000/495000/562500/675000, age 90 | TestBudgetFit/{worked_405000_10pct_under, worked_495000_10pct_over, worked_562500_25pct_over, worked_675000_50pct_over}, TestAgeDays/age_90_days, TestRecency/age_90_days | PASS |

I matched each of the 36 required names against `--- PASS: <name> (` in the captured output: 36 of 36 found, none missing.

## Commands and output

### 1. `cd backend && go test -race -count=1 -v ./internal/domain/... 2>&1` (exit 0)

```
--- PASS: TestParseStatus (0.00s)
    --- PASS: TestParseStatus/active (0.00s)
    --- PASS: TestParseStatus/pending (0.00s)
    --- PASS: TestParseStatus/sold (0.00s)
    --- PASS: TestParseStatus/Active (0.00s)
    --- PASS: TestParseStatus/_active (0.00s)
    --- PASS: TestParseStatus/#00 (0.00s)
    --- PASS: TestParseStatus/closed (0.00s)
--- PASS: TestAllStatuses (0.00s)
--- PASS: TestListingJSONKeys (0.00s)
--- PASS: TestPaginate (0.00s)
    --- PASS: TestPaginate/page_1 (0.00s)
    --- PASS: TestPaginate/last_partial_page (0.00s)
    --- PASS: TestPaginate/last_full_page (0.00s)
    --- PASS: TestPaginate/page_past_end (0.00s)
    --- PASS: TestPaginate/page_past_end_max_int (0.00s)
    --- PASS: TestPaginate/page_size_1 (0.00s)
    --- PASS: TestPaginate/page_size_100 (0.00s)
    --- PASS: TestPaginate/page_size_100_last_page (0.00s)
    --- PASS: TestPaginate/zero_items (0.00s)
    --- PASS: TestPaginate/rank_continuity_across_pages (0.00s)
--- PASS: TestPaginateInvalid (0.00s)
    --- PASS: TestPaginateInvalid/invalid_page_0 (0.00s)
    --- PASS: TestPaginateInvalid/invalid_page_negative (0.00s)
    --- PASS: TestPaginateInvalid/invalid_page_size_0 (0.00s)
    --- PASS: TestPaginateInvalid/invalid_page_size_101 (0.00s)
    --- PASS: TestPaginateInvalid/invalid_page_size_negative (0.00s)
    --- PASS: TestPaginateInvalid/invalid_page_and_page_size (0.00s)
--- PASS: TestRank (0.01s)
    --- PASS: TestRank/no_matches (0.00s)
    --- PASS: TestRank/no_budget (0.00s)
    --- PASS: TestRank/with_budget (0.00s)
    --- PASS: TestRank/tied_scores_tiebreak_order (0.00s)
    --- PASS: TestRank/tied_scores_with_budget (0.00s)
    --- PASS: TestRank/input_order_independent (0.01s)
    --- PASS: TestRank/does_not_mutate_input (0.00s)
--- PASS: TestBudgetFit (0.00s)
    --- PASS: TestBudgetFit/budget_equal_to_price (0.00s)
    --- PASS: TestBudgetFit/worked_405000_10pct_under (0.00s)
    --- PASS: TestBudgetFit/worked_495000_10pct_over (0.00s)
    --- PASS: TestBudgetFit/worked_562500_25pct_over (0.00s)
    --- PASS: TestBudgetFit/worked_675000_50pct_over (0.00s)
    --- PASS: TestBudgetFit/beyond_50pct_over_clamps_to_0 (0.00s)
    --- PASS: TestBudgetFit/price_0 (0.00s)
    --- PASS: TestBudgetFit/asymmetry_at_matching_percentages (0.00s)
--- PASS: TestAgeDays (0.00s)
    --- PASS: TestAgeDays/age_0_days (0.00s)
    --- PASS: TestAgeDays/age_30_days (0.00s)
    --- PASS: TestAgeDays/age_60_days (0.00s)
    --- PASS: TestAgeDays/age_90_days (0.00s)
    --- PASS: TestAgeDays/future_listed_date (0.00s)
    --- PASS: TestAgeDays/calendar_days_not_24h_periods (0.00s)
    --- PASS: TestAgeDays/now_time_of_day_ignored (0.00s)
    --- PASS: TestAgeDays/now_non_utc_location (0.00s)
--- PASS: TestRecency (0.00s)
    --- PASS: TestRecency/age_0_days (0.00s)
    --- PASS: TestRecency/age_30_days (0.00s)
    --- PASS: TestRecency/age_60_days (0.00s)
    --- PASS: TestRecency/age_90_days (0.00s)
    --- PASS: TestRecency/future_listed_date (0.00s)
    --- PASS: TestRecency/age_15_days (0.00s)
    --- PASS: TestRecency/age_45_days_spec_example (0.00s)
--- PASS: TestScore (0.00s)
    --- PASS: TestScore/no_budget (0.00s)
    --- PASS: TestScore/with_budget_worked_values (0.00s)
    --- PASS: TestScore/score_in_0_1 (0.00s)
PASS
ok  	brillio/backend/internal/domain	1.201s
```
Tallies: `--- PASS` = 66 (10 top-level, 56 subtests); `--- FAIL` or `FAIL` = 0; `--- SKIP` = 0. The `=== RUN` lines are left out.

### 2. `GOFLAGS=-count=1 make check` from the worktree root (exit 0)
`check: lint test` runs `lint-api lint-web test-api test-web`. Frontend deps were already installed (`frontend/node_modules/.package-lock.json` present), so no `npm ci` ran.
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
ok  	brillio/backend/internal/config	1.115s
ok  	brillio/backend/internal/domain	1.119s
ok  	brillio/backend/internal/httpapi	1.143s
cd frontend && npm test
> vitest run
 RUN  v5.0.4 /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking/frontend
 Test Files  2 passed (2)
      Tests  4 passed (4)
```
eslint and tsc printed no diagnostics. The go packages ran uncached (timings shown, no `(cached)`).

### 3. Stdlib-only check (exit 0): PASS
`cd backend && go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain`
```
brillio/backend/internal/domain
```
The only non-standard package is the domain package itself, so it has no third-party or intra-repo imports.

### 4. No wall-clock reads: PASS
`grep -rn 'time\.Now' backend/internal/domain` printed nothing and exited 1 (no matches).

### 5. Tree and diff scope: PASS
- `git rev-parse HEAD` printed `cdec405302066c3d1c4014fdaea583def4b43da2`, the expected HEAD (single parent ee09340).
- `git status --porcelain` was empty before and after the runs.
- `git diff --name-only master...HEAD` listed exactly the six expected files:
```
backend/internal/domain/paginate.go
backend/internal/domain/paginate_test.go
backend/internal/domain/rank.go
backend/internal/domain/rank_test.go
backend/internal/domain/scoring.go
backend/internal/domain/scoring_test.go
```
- `git diff --stat ee09340 HEAD` printed `backend/internal/domain/scoring.go | 7 ++++---` (1 file, 4 insertions, 3 deletions). All seven changed lines are in the `AgeDays` doc comment.

## Notes
- No processes were started besides go test, golangci-lint and the `make check` recipes (eslint, tsc, prettier, vitest). PORTS=none. No source files were changed and nothing was committed.
- TestParseStatus, TestAllStatuses and TestListingJSONKeys are BC-2 tests already on master; the same run shows they still pass.
