Lattice-Reviewed-Commit: cdec405302066c3d1c4014fdaea583def4b43da2
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking
Lattice-Reviewed-Base: master (83952b0a8cb700f40fcc2dc50f785784215cd561)
Lattice-Reviewed-Head: feat/BC-3-domain-ranking (cdec405302066c3d1c4014fdaea583def4b43da2)

# Code Review: BC-3 — domain relevance scoring, deterministic ranking, pagination with global ranks

Reviewed commit: `cdec405302066c3d1c4014fdaea583def4b43da2` (`feat/BC-3-domain-ranking`) against base `master` @ `83952b0`.
Reviewer: agent:fable-5.1-reviewer (independent, cold read of plan + diff + spec).

## 1. Verdict

**PASS**

No implementation-level or plan-level rework needed. One pre-existing, out-of-scope documentation inconsistency in `docs/SPEC.md` is noted below for routing; it is not a defect in this diff.

## 2. Summary

Reviewed the six new files in `backend/internal/domain` (`scoring.go`, `rank.go`, `paginate.go` and their tests) against SPEC §5.3, §6.2, §6.3, §6.4, §8 and the BC-3 plan including the plan-review cycle 1 resolutions. The implementation is a faithful, minimal transcription of the spec: every formula, tiebreak key, pagination boundary and error contract matches, and the test suite covers every §8 domain item by name plus the subtler float-exactness and calendar-day semantics the plan called out. Quality is high: small pure functions, doc comments in the voice of the existing BC-2 files, no dependencies beyond the standard library, and no `time.Now()` anywhere in the package.

### Verification performed

All run from the worktree at the reviewed commit (clean tree, HEAD = cdec405, only the six intended files differ from `master`, no `.lattice/` changes, `go.mod` and BC-2 files untouched).

| Check | Result |
|---|---|
| `go test -race -v ./internal/domain/...` | PASS, 66 subtests, 0 FAIL / 0 SKIP |
| `go test -race ./...` (whole backend) | 138 passed in 4 packages |
| `golangci-lint run ./...` (v2.14.0, standard + errcheck/govet/staticcheck/revive, gofmt) | No issues found |
| `go vet ./...` | clean |
| `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain` | prints only `brillio/backend/internal/domain` |
| `grep -rn 'time\.Now' internal/domain` | no matches |
| Every §8 domain case and every plan-table subtest present by name | confirmed from the verbose PASS lines (`TestBudgetFit/*`, `TestAgeDays/*`, `TestRecency/*`, `TestScore/*`, `TestRank/*`, `TestPaginate/*`, `TestPaginateInvalid/*`) |
| Heap's-algorithm `permutations` helper in `rank_test.go` | independently re-run on `[1..6]`: 720 visits, 720 distinct orderings, so `input_order_independent` really covers all permutations |
| Plan-review resolution #1/#5 side effects | comments recording the §5.3 ownership split and the `MaxPageSize` / `ErrPageOutOfRange` / `ErrPageSizeOutOfRange` names are present on BC-5 |

### Correctness walk-through (spec vs code)

- **`BudgetFit`** (`scoring.go:49-54`): `price <= target → price/target`; else `max(0, 1-2*((price-target)/target))`. Matches §6.2 exactly, including the 50 %-over zero and the clamp. Precondition (`target > 0`, `price >= 0`) is documented rather than guarded, as the plan decided; float division by zero does not panic in Go, so an out-of-contract call degrades to Inf/NaN rather than crashing.
- **`AgeDays`** (`scoring.go:62-69`): both instants reduced to UTC calendar dates, difference in Unix seconds divided by 86400, clamped at 0. This is the correct "whole calendar days" semantics and avoids the two traps the plan identified (`Sub()/24h` truncation and `Duration` saturation for a zero `ListedDate`). The midnight-UTC expectation on `listedDate` is stated in the doc comment (the last commit on the branch).
- **`Recency`** (`scoring.go:74-76`): `math.Pow(0.5, float64(age)/RecencyHalfLifeDays)`. Because `RecencyHalfLifeDays` is untyped, this is float division; the `age_15_days` / `age_45_days_spec_example` tests would fail on integer division.
- **`Score`** (`scoring.go:83-88`): nil fit → recency unchanged; else `0.6*fit + 0.4*recency`. Matches §6.2.
- **`Rank`** (`rank.go:56-69`): new slice (non-nil for nil/empty input), fresh `fit` variable per listing so pointers are never shared, `slices.SortStableFunc` with `byRelevance`.
- **`byRelevance`** (`rank.go:74-87`): `cmp.Or(score desc, listedDate desc, price asc, source asc, id asc)` on the unrounded score. Matches §6.3 key for key. Since `(Source, ID)` is the identity, the order is total; the 720-permutation test proves input-order independence empirically.
- **`Paginate`** (`paginate.go:56-86`): collects both range errors with `errors.Join` around distinct sentinels; `totalPages = ceil(total/pageSize)` via `/` and `%`; the `page > totalPages` guard runs before any multiplication, so `page = math.MaxInt` cannot overflow and `start < total` is guaranteed on the fill path. Past-the-end and zero-item pages return a non-nil empty `Items` with the requested `Page` echoed, as §5.3 requires for UI recovery. The zero `Page` on error is consistent with the doc comment's "when Paginate returns no error" qualifier.

### Float exactness in tests (spot-checked by hand)

The `==` assertions are all on values that are exact or correctly-rounded-to-the-same-double in IEEE 754: `405000/450000` rounds to the literal `0.9`; `1 - 2*(45000/450000)` lands on the exact midpoint between two doubles and round-half-even selects the literal `0.8`; `472500` gives `1 - 0.1d`, which rounds to `0.9d`, so `tied_scores_with_budget` is a genuine exact tie; `math.Pow(0.5, k)` is exact for integer `k`. Weighted sums and fractional powers use `near` with `eps = 1e-12`, which is the right call given the compiler may fuse `0.6*fit + 0.4*rec` into an FMA on arm64. The tie fixture's "both sides evaluate the same expression on identical inputs" reasoning is sound and FMA-proof.

## 3. Issues

No issues found in the diff.

One pre-existing documentation inconsistency, outside BC-3's scope, is recorded here so the orchestrator can route it explicitly rather than let it drop:

**[minor] docs/SPEC.md:195 — §5.4 example `relevance.score` does not follow the §6.2 formula**
The success-response example shows `{ "score": 0.8731, "budgetFit": 1.0, "recency": 0.3536 }`. By §6.2 that listing scores `0.6*1.0 + 0.4*0.3536 = 0.7414`, which is the value the BC-3 tests (and the plan) correctly use for listing X / A1. The code is right; the example number is stale. Plan-review resolution #5 forbids SPEC edits on this branch, and the plan already notes the corrected value "per the run-state log", so this is not a BC-3 defect.
**Fix:** Out of scope here. Suggested routing: either "ignore with justification" (already acknowledged in the plan) or a small docs task to change `0.8731` to `0.7414` in SPEC §5.4 before BC-15 writes the README, so BC-15 does not quote the wrong number.

## 4. Positive Observations

- **The formula lives in one readable place.** `scoring.go` opens with a plain-English statement of the scoring rules that BC-15 can quote verbatim, and each of the four functions is a two-to-five-line transcription of its spec block. This is exactly what the ticket asked for ("small, readable, exhaustively tested") and what an interviewer probing "why this formula?" needs.
- **Ordering is isolated behind one comparator.** `byRelevance` is the single definition of §6.3, with one comment per key, which is the §9.4 "sort options" seam the spec designs for.
- **Calendar-day age is done correctly, not approximately.** Reducing both instants to UTC dates and dividing Unix seconds avoids the `Sub()/24h` off-by-one across midnight and the `time.Duration` saturation for a zero date; the `calendar_days_not_24h_periods`, `now_time_of_day_ignored` and `now_non_utc_location` cases pin those semantics down.
- **Overflow and allocation discipline in `Paginate`.** Checking `page > totalPages` before `(page-1)*pageSize`, computing the ceiling with `/` and `%` rather than `total+pageSize-1`, and sizing `Items` exactly are small choices that remove whole classes of edge-case bugs; `page_past_end_max_int` proves the first one.
- **Test design is deliberate about float exactness.** The split between `==` (exactly representable or correctly rounded values) and `near` (fractional powers, weighted sums) is documented inline and matches what IEEE 754 actually guarantees. Integer-literal prices in the asymmetry test, with the `450000*1.1 == 495000.00000000006` reason stated, show the author understood the trap.
- **The tie fixture is well constructed.** Six listings arranged so each adjacent pair is decided by exactly one tiebreak key, with the comment on each row naming what the *next* key alone would have done, plus a separate `tied_scores_with_budget` case where only price can break a genuinely exact tie between different sources. The all-permutations test then turns "deterministic" from a claim into evidence.
- **Contracts are explicit.** `Rank` documents non-mutation, non-nil result, per-item `BudgetFit` pointers and the zero-`Now` hazard; `Paginate` documents its sentinels and its relationship to request parsing as a defensive guard. The pointer-uniqueness check in `with_budget` and the `does_not_mutate_input` test enforce the contracts rather than just stating them.
- **Conventions match BC-2.** Sentinel errors follow the `ErrInvalidStatus` pattern, every exported identifier has a name-leading doc comment, tests live in `domain_test` and reuse `domain.DateLayout`, and the lint traps the plan listed (QF1008 promoted selectors, detached header, builtin shadowing) were all avoided on the first lint run.
