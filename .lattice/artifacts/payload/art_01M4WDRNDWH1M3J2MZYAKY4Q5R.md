Verdict: PASS-WITH-NITS
Reviewed HEAD: f4b6c77d01ae1939a4c61e89b76b659195d93138 (one file, docs/EXTENSIONS.md, 220 lines; base master 0b24829)
Reviewer: fable sub-agent (decision 04:02Z)

Findings:
- [Minor] docs/EXTENSIONS.md:192 — "A3 (0.68) is No. 1 and A1 (0.74) is No. 4" holds only on the 12 undeduped listings (price asc: A3 399000, B9 399500, A6 415000, A1 450000); on the 8 dedupe survivors A1 is No. 3. — Recommendation: add "on the 12 listings" (or "without dedupe").
- [Nit] docs/EXTENSIONS.md:130 vs 167 — §8 says 0002 "adds a latitude and longitude index"; the geo rehearsal says "A latitude index in the new 0002 migration makes it a range scan". — Recommendation: one phrasing, e.g. "an index on (latitude, longitude); its leading column makes the box a range scan".
- [Nit] docs/EXTENSIONS.md:198 — "`filterKey` drops it" reads as existing behavior; filterKey deletes only page and pageSize today (filterBarModel.ts:138-143). — Recommendation: "`filterKey` drops it beside `page` and `pageSize`".
- [Nit] docs/EXTENSIONS.md:153,192,210 — `asOf` unbackticked; it is a §5.4 response field. — Recommendation: backtick or leave.
- [Nit] process — plan symbol table lacks grep lines for `Upsert` and `SearchFilter.Statuses`. — Recommendation: append `grep -cE '^func \(s \*Store\) Upsert\(' backend/internal/store/listings.go  # 1` and `grep -cE '^[[:space:]]+Statuses \[\]ListingStatus' backend/internal/domain/filter.go  # 1`.

Symbol check: 152 unique backticked tokens; all accounted for (existing in code on HEAD, "new" at first use, or SPEC contract name / error code / UI string). All plan grep lines returned their stated counts on HEAD. No unaccounted name.
Seams: all five card-note seams correct (searchListings + Dedupe slot comment search.go:37; byRelevance in slices.SortStableFunc rank.go; buildSearch listings.go; searchParams query.go pinned by TestSearchQueryParams and TestOpenAPISearchParamsMatchHandler; FIELDS with status seventh).
Facts checked: A1 0.7414 at asOf 2026-10-13; dedupe winners A1/A2/A3/A5; zips; Ct/Court; survivor order A1, A5, A3, A6, A2, A7, B10, A4; totalItems 12 to 8; maxPrice=451000 drops B7; haversine 0.000/1.160/12.104/14.158 km; price-sort ranks (12-listing basis, see Minor); 0.0001 deg ~ 11 m; 0001 has 4 indexes, no lat/lng; modernc.org/sqlite v1.60.1 SQLITE_ENABLE_FTS5=1; TestSearchPaging/TestSearchStatusFilter totals; assertItemsMatchSeed; strictRoundTrip; OpenAPI tests; parseSearchParams/setSearch/clear/buildSearchQuery/filterKey behavior; FIELDS; BRANDING §5.2/5.4/5.6/5.7/5.10/§9/§10; StatusLine single role=status; Pager renders all pages; parser rules; seed/config/Upsert behavior; labelPageSize; typed fixtures; intro commit; length/structure; emoji and voice checks; scope (only docs/EXTENSIONS.md). All verified except the Minor.
Deviations: all six implementer deviations accepted (true against code).
Recommended path: fix inline (192, 130/167, 198), then advance. No implementation-level or plan-level rework.