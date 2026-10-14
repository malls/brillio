Verdict: PASS-WITH-NITS
Reviewed plan: task_01M4VXK3WPPJA1GVM0NZRW4GGD.md (118 lines)
Reviewed HEAD: 5ac23977f2523a1fcc47e9797454c6e41f8a063c (feat/BC-16-extensions-doc; an ancestor of master, which took BC-12 at b514f8e during this review)
Reviewer: fable sub-agent (decision 04:02Z)

Findings:
- [Minor] Rehearsal/Dedupe and What I would say: the order A1, A5, A3, A6, A2, A7, B10, A4 and A1 = 0.7414 hold only at asOf 2026-10-13/14; recency moves daily (from 2026-11-28 A6 passes A3). Recommendation: pin both to "asOf 2026-10-13 (the SPEC §5.4 date)" and say the order drifts with the clock.
- [Minor] Symbol table line 36: `^\s+ListingsFile string` returns 3 under grep -E but 2 under git grep -E (POSIX ERE has no \s). Recommendation: `^[[:space:]]+ListingsFile string`.
- [Minor] Header base and BC-6 decision are stale: master (0b24829) holds BC-6 and BC-12. Every grep line returns the same count on master. Recommendation: merge master first, run the block once, name that commit in the intro; drop the git grep master workaround.
- [Minor] Size budget: stated upper bounds sum to 312 against a 299 cap, and the awk count includes heading and trailing blank, so 23 body lines fails "under 25". Recommendation: target 16-20 lines per entry, 20-24 per rehearsal; state that the awk count includes heading and blank line.
- [Nit] SQL versus Go: "a precomputed column does not fit" reads as contradicting SPEC §6's scaling path ("pushing the score into SQL or precomputing it"). Recommendation: "precompute the inputs, compute the score in SQL".
- [Nit] Voice check: `!` and `[Jj]ust ` will hit code samples (`!=`); validator must eyeball. Acceptable as written.

Coverage: every ticket scope/acceptance bullet and all three card notes covered. Seams verified real: searchListings + `// Dedupe slot (section 9.1)` (httpapi/search.go), byRelevance in slices.SortStableFunc (domain/rank.go), buildSearch (store/listings.go), searchParams read by TestOpenAPISearchParamsMatchHandler via SearchQueryParams() (master), FIELDS with status seventh (filterBarModel.ts). Plan findings correct: SPEC §9.3 names three geo params; BRANDING §5.2 seventh-field note stale; SPEC §9.1 address+zip misses A2/B8 (22150 vs 22151) and lacks Ct/Court (A5/B11). No scope creep. Acceptance commands run as written on macOS.

Grep block: all 38 lines + 9-file ls matched in the worktree; on master all match except line 36 (regex dialect).

Facts checked: cross-feed pairs A1/B7, A2/B8, A3/B9, A5/B11 verified; totalItems 12 to 8 verified; winners A1, A2, A3, A5 verified; survivor order verified at asOf 2026-10-13/14 (needs pin); maxPrice=451000 removes B7 verified; A1 coordinates verified; 2 km result A1, B7, A2, B8 (1.160 km), A3/B9 12.104 km verified; zips verified; weights, budget-fit points, half-life verified in scoring.go; A1 = 0.7414 at asOf 2026-10-13 verified; 4-decimal rounding ~11 m verified; parseSearchParams/setSearch/clear/filterKey behavior verified; StatusLine single role=status verified; nonNegative *float64 verified; Load per-file check and Upsert later-wins verified; parseOrigins comma-split verified; assertItemsMatchSeed and DisallowUnknownFields verified; SPEC §5.4 example asserted by both tests verified; Pager renders every page verified; Listing never marshaled verified.