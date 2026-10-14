Lattice-Reviewed-Commit: 8dc03cbf2b6f65a1bdd9f94e9345784eec85a20d
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking
Lattice-Reviewed-Base: master (7ab182d6817d3418d45ebe6dee860540e49f7730)
Lattice-Reviewed-Head: feat/BC-3-domain-ranking (8dc03cbf2b6f65a1bdd9f94e9345784eec85a20d)

Lattice-Reviewed-Commit: 8dc03cbf2b6f65a1bdd9f94e9345784eec85a20d
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking
Lattice-Reviewed-Base: master (7ab182d6817d3418d45ebe6dee860540e49f7730)
Lattice-Reviewed-Head: feat/BC-3-domain-ranking (8dc03cbf2b6f65a1bdd9f94e9345784eec85a20d)

# Code Review: BC-3 — domain relevance scoring, deterministic ordering, pagination with ranks

Reviewer: agent:fable-5.1-reviewer (independent, cold read of plan + diff; did not write the code)

## 1. Verdict

**PASS**

## 2. Summary

Reviewed the six new files in `backend/internal/domain` (`scoring.go`, `rank.go`, `paginate.go` and their tests; 873 added lines, no existing file touched) against SPEC §6.2–6.4, §5.3 (page/pageSize), §8 and the BC-3 plan including the Plan-Review Cycle 1 resolutions. The implementation matches the spec formulas line for line, every §8 domain case is present by name, and all four acceptance criteria hold when re-run from a clean checkout of the reviewed commit. The only finding is a one-line doc nuance on `Page.Items`; nothing blocks advancing to `in_validation`.

## 3. Issues

**[MINOR] backend/internal/domain/paginate.go:32 — `Page.Items` doc says "never nil" but the error path returns a zero `Page` with nil `Items`**
The field doc reads "It is never nil: a page past the end, or a page of an empty result set, has no items." On the invalid-input path `Paginate` returns `Page{}` (nil `Items`) alongside the error, which `TestPaginateInvalid` asserts with `reflect.DeepEqual`. The `Paginate` doc does say "returns the zero Page and an error", so the behavior is documented, but the two sentences read as contradictory to someone who only reads the type. No code change needed.
**Fix:** Qualify the field doc, e.g. "It is never nil when Paginate returns no error: ...". Optional; safe to apply inline or skip with a comment.

**[NOTE, non-blocking] backend/internal/domain/scoring.go:49 — `BudgetFit(0, 0)` yields NaN**
Probed: `BudgetFit(0, 0)` = NaN, `BudgetFit(100, 0)` = 0. Both are outside the documented precondition (`target > 0`), which the plan explicitly chose to leave unguarded and untested ("no panic, no invented behavior"), and BC-5 rejects `targetBudget <= 0` at the boundary. A NaN score would still sort deterministically (`cmp.Compare` orders NaN last via the `b, a` argument order). Recording this so the orchestrator can route it explicitly; recommended routing is "ignore with justification" per the plan.

No other issues found.

## 4. Positive Observations

- **Formulas are transcribed exactly from §6.2.** `BudgetFit` uses the `price <= target` branch so price == target lands on the under branch and returns exactly 1.0; the over branch is `max(0, 1-2*pctOver)` with the builtin `max`. `Recency` divides `float64(age)` by the untyped `RecencyHalfLifeDays`, and the plan's integer-division trap is covered by `age_15_days` / `age_45_days_spec_example`.
- **`AgeDays` is the careful version.** Both instants are reduced to UTC midnight before differencing, and the difference is taken in Unix seconds rather than `to.Sub(from)`. Probed: `AgeDays(time.Time{}, fixedNow)` = 739901 (no saturation), `Recency` of that = 0. The three non-obvious cases (12 h across midnight, time-of-day of `now`, non-UTC `now`) each have a named subtest.
- **Ordering lives in exactly one place.** `byRelevance` is a single `cmp.Or` chain with one comment per key, matching the §9.4 "new sort option sits beside it" seam. The comparator is total because `(Source, ID)` is the identity, which is why `input_order_independent` can assert all 720 permutations produce the same order.
- **The tie fixture is well constructed.** Each adjacent pair is decided by exactly one key, and the comment on each row states what the *next* key alone would have done. I hand-checked all five pairs; each claim is correct, including the honest note on row 5 that listedDate also agrees there and that `tied_scores_with_budget` is the case only the score decides.
- **Float exactness is decided per assertion, not globally.** `==` is used only where IEEE-754 guarantees it (BudgetFit worked values, whole-half-life recency, `Score(nil, r) == r`, the [0,1] bound); `near` with eps 1e-12 is used for fractional exponents and weighted sums, with a comment explaining the FMA concern. Probed on this arm64 machine: `405000/450000 == 0.9` true, `BudgetFit(472500, 450000) == 0.9` true, `math.Pow(0.5, 1.5)` = 0.35355339059327373 (matches the test literal), `Score(&1, 1) == 1` true.
- **Pagination overflow is handled by ordering, not by arithmetic tricks.** The `page > totalPages` check runs before `(page-1)*pageSize`, and `page_past_end_max_int` exercises it. `totalPages` uses `/` and `%` rather than `total+pageSize-1`.
- **Error design follows the existing `ErrInvalidStatus` pattern.** Two sentinels, wrapped with `%w`, joined with `errors.Join` so both page and pageSize problems surface together (§5.3 "collected and returned together"). `MaxPageSize` is exported so BC-5 shares the bound.
- **Per-item `BudgetFit` pointer is fresh.** `with_budget` asserts no two items share a pointer, catching the classic loop-variable aliasing bug.
- **Non-mutation and non-nil contracts are both tested**, not just documented (`does_not_mutate_input` via `slices.Equal` against a clone; `no_matches` covers nil and empty input with and without a budget).
- **Doc comments are complete and lint-clean under the strict config** (revive `exported`, staticcheck QF1008 promoted selectors, detached file header that does not trip `package-comments`). The scoring.go header states the formula in plain words as BC-15 requires.

## 5. Verification performed

All run from the worktree at the reviewed commit (`git diff --quiet 8dc03cb -- backend/internal/domain` confirmed the tree matches HEAD).

| Check | Command | Result |
|---|---|---|
| Tests, race detector | `cd backend && go test -race -v ./internal/domain/...` | PASS, 66 subtests (full name list below) |
| Repo-root gate | `make lint-api test-api` | golangci-lint 2.14.0: 0 issues; `go test -race ./...` ok for config, domain, httpapi |
| Stdlib-only | `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain` | prints only `brillio/backend/internal/domain` |
| No wall clock | `grep -rn 'time\.Now' backend/internal/domain` | no matches |
| go vet / gofmt | `go vet ./...`, `gofmt -l internal/domain` | clean |

§8 domain cases present by name (every one passed):
`TestBudgetFit/{budget_equal_to_price, worked_405000_10pct_under, worked_495000_10pct_over, worked_562500_25pct_over, worked_675000_50pct_over, beyond_50pct_over_clamps_to_0, price_0, asymmetry_at_matching_percentages}`;
`TestAgeDays/{age_0_days, age_30_days, age_60_days, age_90_days, future_listed_date, calendar_days_not_24h_periods, now_time_of_day_ignored, now_non_utc_location}`;
`TestRecency/{age_0_days, age_30_days, age_60_days, age_90_days, future_listed_date, age_15_days, age_45_days_spec_example}`;
`TestScore/{no_budget, with_budget_worked_values, score_in_0_1}`;
`TestRank/{no_matches, no_budget, with_budget, tied_scores_tiebreak_order, tied_scores_with_budget, input_order_independent, does_not_mutate_input}`;
`TestPaginate/{page_1, last_partial_page, last_full_page, page_past_end, page_past_end_max_int, page_size_1, page_size_100, page_size_100_last_page, zero_items, rank_continuity_across_pages}`;
`TestPaginateInvalid/{invalid_page_0, invalid_page_negative, invalid_page_size_0, invalid_page_size_101, invalid_page_size_negative, invalid_page_and_page_size}`.
The §5.3 status enum rule remains covered by BC-2's `TestParseStatus` (unchanged), per Plan-Review resolution 5.

A temporary probe test was added to confirm the exactness and precondition claims above, then deleted; the worktree is clean.

## Recommendation

Advance to `in_validation`. Validation is N/A for this pure library per the ticket; the test output above is the citable evidence. Optionally apply the one-line `Page.Items` doc tweak inline before the PR.
