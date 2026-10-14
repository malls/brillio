Lattice-Reviewed-Commit: ee09340b9f844c98cf3c092a41bd0cebe7777036
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking
Lattice-Reviewed-Base: master (83952b0a8cb700f40fcc2dc50f785784215cd561)
Lattice-Reviewed-Head: feat/BC-3-domain-ranking (ee09340b9f844c98cf3c092a41bd0cebe7777036)

# Code Review: BC-3 — domain relevance scoring, deterministic ranking, pagination with global ranks

Reviewed: `feat/BC-3-domain-ranking` @ ee09340 against `master` @ 83952b0, in worktree
`/Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking`. Six new files under
`backend/internal/domain` (874 insertions); no other file differs from master. Branch commits
contain no `.lattice/` state.

## 1. Verdict

**PASS**

## 2. Summary

I reviewed the scoring, ranking and pagination core against SPEC §6.2–6.4, §8 and the plan
(including the Cycle 1 resolutions), re-ran the gate, and probed the float and edge-case claims
independently. The implementation is small, matches the spec formulas line for line, and every
test named in the plan's §8 table exists and passes. The only findings are two documentation
nits, one of which is a known SPEC example inconsistency that is explicitly out of scope for this
branch. Nothing needs rework.

## 3. Issues

**[MINOR] backend/internal/domain/scoring.go:58 — `AgeDays` doc overstates that "location ... does not matter"**
The comment says the time of day and the location of either value do not matter. That is true
for `now`, but a `listedDate` constructed as a local midnight east of UTC lands on the previous
UTC calendar date and gains a day. Verified: `AgeDays(2026-09-13 00:00 JST, fixedNow)` returns
31 where the UTC-midnight value returns 30. The behavior is correct per §6.2 ("calendar days ...
UTC") and unreachable in practice because BC-2's `Listing.ListedDate` contract is "midnight UTC
via DateLayout", so this is wording only.
**Fix:** Tighten the sentence to something like: "Each instant is taken on the UTC calendar, so
the time of day of either value does not matter. `listedDate` is expected to be the midnight-UTC
value that parsing with DateLayout yields." No test change needed.

**[MINOR, out of scope for BC-3] docs/SPEC.md:195 — §5.4 example `score: 0.8731` contradicts the §6.2 formula**
With `budgetFit 1.0` and `recency 0.3536` the formula gives 0.6 + 0.4×0.3536 = 0.7414, not 0.8731.
The implementation and the test (`rank_test.go:156`, listing X in `TestRank/with_budget`)
correctly use 0.7414…; the plan's Cycle 1 resolution forbids editing SPEC in this branch, so the
discrepancy is now recorded in a test comment only.
**Fix:** Do not change this branch. Create a small Lattice task (or a note on BC-5/BC-15, which
will quote this example) to correct the §5.4 example to `0.7414`, so the README and the
OpenAPI example do not propagate the wrong number.

Optional nit, no action required: `TestRank/input_order_independent` (`rank_test.go:239`) asserts
720 visits but not 720 *distinct* orderings. I verified empirically that the Heap's-algorithm
helper does produce 720 distinct permutations, so the test is sound as written; a `map[string]bool`
of seen keys would make that property self-evident if anyone later touches the helper.

## 4. Positive Observations

**Gate and acceptance criteria, all verified by me on this worktree:**

| Check | Result |
|---|---|
| `go test -race -count=1 ./internal/domain/...` | ok, 66 tests |
| `golangci-lint run` (v2.14.0, standard + errcheck/govet/staticcheck/revive, gofmt) | 0 issues |
| `go vet`, `gofmt -l` | clean |
| Non-stdlib deps of `./internal/domain` | none (prints only the package itself) |
| `grep -rn 'time\.Now' internal/domain` | no matches |
| Statement coverage of the new code | 100% on every function |
| `GOARCH=amd64 go test` (no FMA contraction) | 66 passed |
| `go test -count=5` | 330 passed, no flakes |

**Every §8 domain item and every plan-table subtest is present by name.** I diffed the verbose
run against the plan's table: `TestBudgetFit` (7 worked rows + asymmetry), `TestAgeDays` (8
including `calendar_days_not_24h_periods`, `now_time_of_day_ignored`, `now_non_utc_location`),
`TestRecency` (7 including the fractional-exponent `age_15_days` / `age_45_days_spec_example`),
`TestScore` (3), `TestRank` (7), `TestPaginate` (10) and `TestPaginateInvalid` (6). The §5.3
split (page/pageSize bounds here, the rest in BC-5, status enum via BC-2's `TestParseStatus`)
follows the plan's Cycle 1 resolution #5.

**Formula fidelity.** `BudgetFit`, `AgeDays`, `Recency` and `Score` are each a few lines and read
exactly like §6.2. The header comment in `scoring.go` states the formula in words as the ticket
requires for BC-15, and the detached-header trick keeps revive's package-comment rule quiet while
`listing.go` keeps the package doc. The constants are untyped, so
`float64(age)/RecencyHalfLifeDays` is float division; the age-15/age-45 tests would catch the
integer-division regression.

**`AgeDays` is done the right way.** Reducing both instants to UTC midnights and dividing Unix
seconds avoids the two classic bugs the plan called out (duration truncation across midnight,
and `Sub` saturation for a zero date). The `now_non_utc_location` test (22:00 EDT = next UTC day)
is a particularly good guard for the request clock.

**Ordering is a single, readable comparator.** `byRelevance` is `cmp.Or` over the five §6.3 keys
with one comment per key, and it is the only place ordering lives (the §9.4 seam). The tie
fixture is well constructed: each adjacent pair is decided by exactly one key, with the comment
honestly noting that without a budget listedDate and score agree, and `with_budget` is where
score alone overrides listedDate (X at 08-29 beats W at 09-13, Z newest lands last). The
720-permutation test plus `does_not_mutate_input` nail determinism and the "never reorders the
caller's slice" contract.

**Float discipline is explicit and correct.** Exact `==` is used only where IEEE-754 makes it
exact (BudgetFit worked values, whole-half-life recency, no-budget `Score == Recency`); `near`
with 1e-12 is used for fractional powers and weighted sums. The amd64 run confirms the
assertions are not FMA-dependent. Integer price literals instead of `450000*1.1` avoid the
`495000.00000000006` trap.

**Pagination is overflow-safe by construction.** `page > totalPages` is checked before
`(page-1)*pageSize`, and `page_past_end_max_int` proves it. `totalPages` uses `/` and `%` rather
than `total+pageSize-1`. Validation collects both errors with `errors.Join` and distinct sentinels
so BC-5 can map each to its own field with `errors.Is`; the doc comment correctly scopes the
"Items never nil" guarantee to the success path (the error path returns the zero `Page`, which
`TestPaginateInvalid` asserts with `reflect.DeepEqual`). Rank continuity is asserted for sizes
1, 7, 10 and 100, with the ticket's "page 2 of 10 starts at 11" called out explicitly.

**Hygiene.** Four focused commits matching the plan's suggested split plus a doc-tightening
commit; the `master` merge is a true merge (no rebase) and introduces nothing outside the six
new files. Doc comments on every exported identifier follow the `listing.go` style, promoted
selectors are used throughout (no QF1008 hits), and the `RankInput.Now` zero-value note from
Cycle 1 resolution #2 is present on both the field and `Rank`.

**Recommended next step for the orchestrator:** advance to `in_validation`, record the N/A
validation justification citing the PASS lines above, and open the PR. Apply the `AgeDays`
wording tweak inline if desired; it does not warrant a rework cycle.
