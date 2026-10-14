# Plan Review: BC-3 — domain relevance scoring, deterministic ordering, pagination with ranks

## 1. Verdict

**PASS** — Plan is complete, feasible, and aligned. Implementation can proceed. The four minor items below are plan-text amendments and one delegator decision; none requires a return to `in_planning`.

## 2. Summary

Reviewed the BC-3 plan against the ticket, SPEC §5.3–§5.5, §6.2–§6.4, §8, §9.4, the BC-2 domain package on the `feat/BC-3-domain-ranking` worktree (cut at 5da9b80), `backend/.golangci.yml`, the Makefile, and the BC-5 ticket. The plan is unusually precise: every exported identifier matches the ticket's signatures, every §8 domain case is mapped to a named subtest, and the plan's numerical and lint claims were checked empirically on this machine (go1.27.1 darwin/arm64, golangci-lint 2.14.0) and all hold. The only open question is a cross-ticket one the plan already surfaces: whether to amend SPEC §8's "each validation rule in §5.3" wording, which the delegator must decide before the Result Validator checks §8 by name.

## 3. Issues

**[minor] Files and API → paginate.go — Overstated claim about BC-5's behavior**
The plan says "BC-5 checks each with `errors.Is` and maps ErrPageOutOfRange to field `page` ... (as BC-5's ticket specifies)". BC-5's ticket specifies the `out_of_range` code for `page`/`pageSize` but assigns the bounds check to `httpapi.ParseSearchQuery`; it says nothing about consuming domain errors. The sentinel design is still right (it keeps `Paginate` self-defending and gives the handler a precise mapping if it wants one), but a BC-5 planner reading this plan could believe a contract exists that BC-5 never agreed to.
**Recommendation:** Reword to "BC-5 *can* distinguish the two with `errors.Is`; its parser enforces the bounds first, so these errors are a defensive guard." Record a one-line `lattice comment` on BC-5 naming `MaxPageSize`, `ErrPageOutOfRange` and `ErrPageSizeOutOfRange` so the seam is discoverable rather than assumed.

**[minor] Files and API → rank.go — `RankInput.Now` zero value is silently wrong, undocumented**
Verified: with `Now == time.Time{}`, `AgeDays` returns 0 for every listing (the future clamp fires), so every recency is 1.0 and ordering collapses to the tiebreak keys. This is a caller bug, not a domain bug, and BC-5 injects the clock, but the `Rank` doc contract lists three guarantees and omits this precondition.
**Recommendation:** Add one sentence to the `Rank`/`RankInput.Now` doc comment: "Now must be the request clock; the zero time makes every listing look future-dated (age 0)." No guard, no test.

**[minor] Tests → table — Fractional-exponent literals must stay under `near`**
`math.Pow(0.5, 1.5)` evaluates to `0.35355339059327373` here; the plan writes `≈ 0.3535533905932738` and the derived score `0.7414213562373095` (which does match exactly). The plan already assigns these to `near`, so nothing fails as written, but the mismatched literal is a trap if an implementer tidies a `near` into `==`.
**Recommendation:** Keep the `age_15_days`, `age_45_days_spec_example`, `with_budget_worked_values` and the Rank `with_budget` score assertions on `near`; do not promote any Score sum or fractional recency to `==`. Also worth stating in the plan's Exactness section that arm64 may fuse `0.6*fit + 0.4*rec` into an FMA, which is the concrete reason Score sums are `near` and never `==`.

**[minor] Header / Shared-file handling — "toolchain go1.27.1" is not in go.mod**
`backend/go.mod` is `go 1.27` with no `toolchain` directive; go1.27.1 is the installed toolchain. Harmless, but the plan also says `go.mod` is untouched, and an implementer could read the header as license to add a `toolchain` line.
**Recommendation:** Change the header to "go 1.27 (installed toolchain go1.27.1)".

**[minor, delegator decision] SPEC edit (proposed) — §8 "each validation rule in §5.3" needs an explicit call**
The plan correctly notes that §8's domain list claims "each validation rule in §5.3" while BC-5's ticket owns all of those rules except `page`/`pageSize` bounds and the status enum. The plan proposes a SPEC wording change and a fallback (lattice comments on BC-3 and BC-5). This is a "complex" triage item: it changes what the Result Validator will check by name, so it must not be silently dropped.
**Recommendation:** Delegator picks one before `in_progress`: (a) adopt the SPEC edit in-branch under the SPEC's deviation rule, consistent with how run-state already handles the §5.4 example correction, or (b) record the split as comments on both tickets. Either way, note the choice in a `lattice comment` on BC-3.

## 4. Positive Observations

- **Signatures match the ticket exactly.** Constants, `BudgetFit`, `AgeDays`, `Recency`, `Score`, `Rank`/`RankInput`/`Scored`/`Relevance`, `Paginate`/`Page`/`RankedItem`, and `slices.SortStableFunc` are all as specified; nothing is added beyond `MaxPageSize`, the two sentinels and the unexported comparator, each with a stated reason.
- **Every §8 domain case and every ticket bullet is mapped to a named subtest**, including the ones the ticket calls out by value (450000/405000/495000/562500/675000; ages 0/30/60/90; pageSize 1 and 100; rank 11 on page 2; totalPages 0). The tie fixture is well crafted: each adjacent pair is decided by exactly one key and the table says what the next key alone would have done, which makes a wrong comparator fail loudly.
- **The numerical claims are correct, not hand-waved.** Checked on go1.27.1 darwin/arm64: all BudgetFit worked values are exact, including `1-2*(45000/450000) == 0.8` via round-half-even and the `405000` vs `472500` budget tie at exactly `0.9`; `0.6+0.4 == 1.0`; `math.Pow(0.5, k)` exact at whole half-lives; the [0,1] bound holds exactly on the test grid; `450000*1.1 != 495000` as the plan warns; `AgeDays` yields 0/30/60/90/0/45 for the listed dates, 1 for the 12 h-across-midnight case, 31 for the EDT case, and does not overflow for a zero `ListedDate`.
- **The lint analysis is accurate.** A throwaway skeleton of the planned API (detached file header, const block, embedded `Scored`, `Page.Page`, `RankedItem.Rank` beside `func Rank`, `errors.Join`, `cmp.Or` comparator) lints at 0 issues under `backend/.golangci.yml` with golangci-lint 2.14.0, vets clean, and `go list -deps` shows only stdlib. QF1008 does fire on `a.Relevance.Score` under the default config, so the promoted-selector rule is a real trap, correctly flagged.
- **Boundary and overflow reasoning is sound.** Validation precedes the division, `page > totalPages` precedes the multiplication, and the ceiling uses `/` and `%`, so `page = math.MaxInt` is safe and tested.
- **Architecture respects the seams.** No store dependency (BC-3 and BC-4 land in parallel), ordering lives in one comparator (§9.4 sort extension point), rounding and JSON stay in BC-5, and the plan follows BC-2's conventions: external `domain_test` package, sentinel-error pattern of `ErrInvalidStatus`, doc-comment style, no marshaling of `Listing`-embedding types.
- **Cross-ticket awareness.** The plan uses the corrected §5.4 example (0.7414, per run-state) rather than the SPEC's inconsistent 0.8731, confines itself to new files, and spells out the merge discipline against BC-2/BC-4.
- **Baseline confirmed:** `make lint-api test-api` is green at 5da9b80 on the BC-3 worktree, as the plan states.
