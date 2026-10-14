# Ticket card — BC-3

```
TICKET=BC-3
TASK_UUID=task_01M4VXK25Y5H07PWF6BX0PMFG1
TITLE=Domain: relevance scoring, deterministic ordering, pagination with ranks (pure Go)
WORKTREE=/Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking
BRANCH=feat/BC-3-domain-ranking
REVIEW_BASE=feat/BC-2-backend-scaffold
ANCHOR=feat/BC-2-backend-scaffold (BC-2 at review 01:47Z, head 5da9b80; not yet landed)
PORTS=none (pure library)
AREA=area:11
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-3 notes (binding)

1. **Press-ahead anchor.** If `ANCHOR` is `feat/BC-2-backend-scaffold`, your branch was cut from BC-2's in-review
   branch so you inherit its shared domain types. Review your own diff against `REVIEW_BASE`. When BC-2 lands
   (watch `git log master`), `git merge master` (never rebase) and switch `REVIEW_BASE` to `master` for any later
   review. If BC-2 gains commits before it lands, `git merge feat/BC-2-backend-scaffold`.
2. **BC-4 runs in parallel on the same anchor** and owns `internal/store` and the `cmd/api/main.go` store wiring.
   You add files under `internal/domain` only (`scoring.go`, `rank.go`, `paginate.go`, tests). Do not modify the
   BC-2 types; if a type change is truly needed, mail the Orchestrator first (BC-4 depends on the same names).
3. Stdlib only; no `time.Now()` anywhere in `internal/domain` (the clock is injected); `slices.SortStableFunc`;
   exported constants `BudgetWeight`, `RecencyWeight`, `RecencyHalfLifeDays`. Keep full precision — rounding is
   BC-5's job. The package doc comment at the top of `scoring.go` states the formula in words (BC-15 quotes it).
4. Name every test case after its §8 domain item so the reviewer and the Result Validator can find each one
   (validation-plan rows 6, 29, 30).
5. Gate: `make lint-api test-api` (full `make check` if BC-7 has landed on master by hand-off).
6. Validation: N/A (pure library). The validator records a one-line justification citing
   `go test -race -v ./internal/domain/...` output with the case names.
