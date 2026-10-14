# BC-3: Domain: relevance scoring, deterministic ordering, pagination with ranks (pure Go)

Implement the ranking core as pure Go with no I/O: budget fit, recency, score, deterministic ordering, pagination with global ranks. Spec: docs/SPEC.md §6.2, §6.3, §6.4, §8 (domain list). Clock is injected; never call time.Now() in this package.

Why: this is the logic the interviewer will probe hardest ("why this formula?"). It must be small, readable and exhaustively tested, and it must not depend on the store (BC-4) so both can land in parallel.

Scope (package internal/domain, building on the types from BC-2):
- Exported constants `BudgetWeight = 0.6`, `RecencyWeight = 0.4`, `RecencyHalfLifeDays = 30`.
- `BudgetFit(price, target float64) float64` asymmetric per §6.2 (under: price/target; over: max(0, 1 - 2*pctOver)).
- `AgeDays(listedDate, now time.Time) int` whole calendar days in UTC, clamped at 0 for future dates.
- `Recency(listedDate, now time.Time) float64` = 0.5^(ageDays/30).
- `Score(budgetFit *float64, recency float64) float64`: with budget 0.6*fit + 0.4*recency; without budget, recency alone.
- `Rank(listings []Listing, in RankInput) []Scored` where `RankInput{TargetBudget *float64; Now time.Time}` and `Scored{Listing; Relevance{Score float64; BudgetFit *float64; Recency float64}}`. Sort per §6.3 on unrounded score desc, then listedDate desc, price asc, source asc, id asc. Use slices.SortStableFunc.
- `Paginate(ranked []Scored, page, pageSize int) (Page, error)` returning `Page{Items []RankedItem; Page, PageSize, TotalItems, TotalPages int}` where RankedItem adds `Rank int` = 1-based index in the whole set. Invalid page (<1) or pageSize (outside 1..100) returns a typed error the handler can map to §5.5 codes; page past end returns an empty Items with correct metadata.
- Full precision kept here; rounding to four decimals is the handler's job (BC-5).
- Tests, table-driven, per §8: worked values from §6.2 (450000/405000/495000/562500/675000; ages 0/30/60/90), asymmetry at equal percentages, future dates, no budget, tied scores with the exact tiebreak order asserted on a crafted fixture, no matches, every pagination boundary including rank continuity across pages (page 2 of pageSize 10 starts at 11), pageSize 1 and 100, zero items giving totalPages 0.

Acceptance:
- `go test -race ./internal/domain/...` green with every §8 domain case present by name.
- Package imports nothing outside the standard library.
- golangci-lint clean.
- A short doc comment at the top of scoring.go states the formula in words; BC-15 will quote it.

Validation: N/A (pure library, no running surface). Record a one-line justification with --role validation and cite the test output.
