Lattice-Reviewed-Commit: e9144d642dd266e86b05f11dada700dd4cfd0435
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking
Lattice-Reviewed-Base: master (7ab182d6817d3418d45ebe6dee860540e49f7730)
Lattice-Reviewed-Head: feat/BC-3-domain-ranking (e9144d642dd266e86b05f11dada700dd4cfd0435)

# Code Review: BC-3 — domain relevance scoring, deterministic ranking, pagination with global ranks

Reviewed commit: `e9144d642dd266e86b05f11dada700dd4cfd0435` on `feat/BC-3-domain-ranking`
Base: `master` @ `7ab182d`
Reviewer: agent:claude-fable-5.1-reviewer (cold read; did not write this code)

## 1. Verdict

**PASS**

## 2. Summary

Reviewed the six new files under `backend/internal/domain` (`scoring.go`, `rank.go`, `paginate.go` and their tests) against the BC-3 ticket, the plan (including the five authoritative Plan-Review Cycle 1 resolutions), and SPEC §5.3, §5.4, §6.2–§6.4 and §8. The implementation is a faithful, small and well-documented transcription of the spec formulas; every acceptance criterion holds when re-run independently, and the test suite covers every §8 domain item by name plus the float-exactness and ordering edge cases the plan called out. I found no correctness, security or quality defects that require rework; the only items below are non-blocking notes, one of which points at a pre-existing inconsistency in `docs/SPEC.md` that is out of scope for this branch.

## Verification performed

All run from the worktree at HEAD `e9144d6` (clean working tree, HEAD matches `Lattice-Reviewed-Commit`).

| Check | Result |
|---|---|
| `make lint-api test-api` (repo root) | PASS — golangci-lint 2.14.0: `0 issues.`; `go test -race ./...` ok for config, domain, httpapi |
| `go test -race -count=1 -v ./internal/domain/...` | PASS — 66 tests (10 top-level, 56 subtests); every subtest name in the plan's §8 table is present and passes |
| `gofmt -l ./internal/domain` | clean |
| `go vet ./internal/domain/...` | clean |
| `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain` | prints only `brillio/backend/internal/domain` (stdlib-only) |
| `grep -rn 'time\.Now' backend/internal/domain` | no matches |
| `git diff --name-status master..HEAD` | only `backend/internal/domain/{scoring,rank,paginate}{,_test}.go`; `docs/`, `Makefile`, `go.mod`, BC-2 files and `.lattice/` untouched |

Acceptance criteria from the ticket:

- [x] `go test -race ./internal/domain/...` green with every §8 domain case present by name
- [x] Package imports nothing outside the standard library (`cmp`, `errors`, `fmt`, `math`, `slices`, `time`)
- [x] golangci-lint clean
- [x] Short doc comment at the top of `scoring.go` states the formula in words (lines 1–14), detached from `package domain` by one blank line as the plan requires

Plan-Review Cycle 1 resolutions, all honored:

1. `Paginate` doc frames the sentinel errors as a defensive guard behind request parsing (`paginate.go:49–53`); no claim about BC-5's mapping.
2. `RankInput.Now` zero value documented on the field (`rank.go:14–16`) and in the `Rank` doc (`rank.go:53–54`).
3. Fractional literal is `0.35355339059327373`; every Score sum and fractional recency uses `near`, and `==` is confined to the cases the plan enumerates.
4. `go.mod` untouched, no `toolchain` directive.
5. No edit to `docs/SPEC.md`.

## 3. Issues

No blocking issues found. The following are non-blocking notes for the orchestrator to route (none require changes to this branch).

**[minor / out of scope] docs/SPEC.md:190 (§5.4 example) — example `score: 0.8731` contradicts the §6.2 formula**
The §5.4 response example shows `budgetFit: 1.0`, `recency: 0.3536`, `score: 0.8731`. By §6.2, `0.6 * 1.0 + 0.4 * 0.3536 = 0.7414`. The implementation and its tests (`TestRecency/age_45_days_spec_example`, `TestScore/with_budget_worked_values`, `TestRank/with_budget` row `X`) use the correct `0.7414213562373095`, consistent with the plan's note "corrected per the run-state log". This is a spec-document defect, not a BC-3 defect.
**Fix:** Route as "create new task" (or fold into BC-15 docs): change the §5.4 example score to `0.7414` so BC-5's handler tests and the README do not inherit the wrong number. Do not edit the spec in this branch (plan resolution #5).

**[informational] backend/internal/domain/scoring.go:47–52 — `BudgetFit` with `target <= 0` is unspecified by design**
`BudgetFit(price, 0)` yields `NaN` or `+Inf`, and `Rank` with `TargetBudget` pointing at `0` would propagate that into scores (`cmp.Compare` orders NaN below every number, so sorting stays total but meaningless). The plan explicitly chose "precondition, no panic, no invented behavior, no test" because §5.3 rejects `targetBudget <= 0` at the HTTP boundary; the precondition is documented on `BudgetFit` and on `RankInput.TargetBudget`.
**Fix:** None for this branch. Make sure the planned lattice comment on BC-5 names this precondition alongside `MaxPageSize`, `ErrPageOutOfRange` and `ErrPageSizeOutOfRange`.

**[informational] backend/internal/domain/scoring.go:59–66 — `AgeDays` assumes `ListedDate` is a UTC calendar date**
`AgeDays` takes each instant's UTC calendar date, exactly as §6.2 specifies. If the BC-4 loader ever produced a `ListedDate` at local midnight in a UTC+N zone, `.UTC()` would shift it to the previous day. `Listing.ListedDate`'s doc already mandates `time.Parse(DateLayout, s)` (midnight UTC), so the contract is in place.
**Fix:** None. Worth a one-line reminder in BC-4's review that the store must return midnight-UTC dates.

**[nit] backend/internal/domain/rank_test.go:48–66 — `tieFixture` builds a slice named `want` and returns it reversed**
Readable thanks to the comment, but the name `want` for what becomes the *input* costs a second read.
**Fix:** Optional; `ordered` or `inRankOrder` would be clearer. Not worth a rework cycle.

## 4. Positive Observations

**Spec fidelity.** `BudgetFit`, `AgeDays`, `Recency`, `Score`, `byRelevance` and `Paginate` map one-to-one onto §6.2, §6.3 and §6.4, with the §6.2 worked values (450000/405000/495000/562500/675000; ages 0/30/60/90) asserted exactly. The ordering comparator (`rank.go:74–87`) is the single seam §9.4 asks for, with one comment per key.

**Numerics handled with care, not hand-waving.** The plan's exact-vs-`near` split is applied precisely: `==` only where IEEE-754 guarantees it (correctly rounded division, `math.Pow(0.5, k)` for integer `k`, same-expression-same-inputs ties), `near` for fractional exponents and weighted sums that arm64 may fuse into an FMA. The `tied_scores_with_budget` test is a nice touch: it proves `405000/450000` and `1 - 2*(22500/450000)` land on the identical double, so the price tiebreak is genuinely exercised on an exact score tie. Asymmetry prices are integer literals, sidestepping `450000*1.1 == 495000.00000000006`.

**Calendar-day arithmetic done right.** `AgeDays` normalizes to UTC midnights and subtracts Unix seconds, avoiding both the `Sub/24h` truncation bug and `Duration` saturation for a zero `ListedDate`. The three dedicated tests (`calendar_days_not_24h_periods`, `now_time_of_day_ignored`, `now_non_utc_location`) pin each of those behaviors.

**Overflow-safe pagination.** `page > totalPages` is checked before `(page-1)*pageSize` is ever computed (`paginate.go:74–78`), and `page_past_end_max_int` proves it. `totalPages` uses `/` and `%` rather than `total+pageSize-1`. Validation collects both errors with `errors.Join` so BC-5 can map each with `errors.Is`, mirroring the existing `ErrInvalidStatus` pattern.

**Determinism proven, not asserted.** `input_order_independent` runs all 720 permutations of the tie fixture through a compact iterative Heap's algorithm and asserts the count; the fixture is constructed so each adjacent pair is decided by exactly one tiebreak key, with the key that *would* have reversed it noted in a comment. `does_not_mutate_input` and `no_matches` (nil and empty, with and without budget) close the remaining contract clauses of `Rank`.

**Pointer hygiene.** `Rank` allocates a fresh `BudgetFit` per listing, and `TestRank/with_budget` detects any shared pointer via a `map[*float64]string`. Easy to get wrong, cheap to test, and tested.

**Documentation that will survive a cold read.** Every exported identifier carries a doc comment in the `listing.go` voice; the `scoring.go` header reads as prose BC-15 can quote verbatim; the in-memory-only nature of `Relevance`/`Scored` is stated where a future agent would look. The last commit (`e9144d6`) tightening the `Page.Items` non-nil guarantee to the success path shows attention to doc precision.

**Clean handoff.** Three logically separate feature commits plus a `master` merge (not a rebase), no shared-file edits, no `.lattice/` state in the tree, and the lint traps the plan warned about (QF1008 promoted selectors, builtin shadowing, detached header) are all avoided.
