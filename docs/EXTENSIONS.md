# Extension playbook

Where each likely live-session requirement goes, what it touches and what it costs. Entries follow SPEC §9's numbering.
Every code name in backticks was checked against commit 0b24829 on 2026-10-14. A name written as "new `X`" does not exist
yet: it is the proposal. SPEC param names, response fields, error codes and UI strings are in backticks too.
Estimates are build time including tests. Paths are from the repo root.

## Where things go

- **Backend.** `searchListings` in `backend/internal/httpapi/search.go` owns the stage order: parse, `Store.Search`, `Rank`,
  `Paginate`, respond. A new stage is one call between two others. `backend/internal/domain` is pure Go with the clock
  passed in. `buildSearch` in `backend/internal/store/listings.go` builds the search SQL.
- **Frontend.** The URL is the state (`useSearchParamsState`). `FIELDS` in
  `frontend/src/components/FilterBar/filterBarModel.ts` drives the filter bar. `formatListingMeta` builds the card's meta
  line. `StatusLine` sits above the list and `Pager` below it.

**Param checklist.** In `backend/internal/httpapi/query.go`: a param const and its place in `searchParams`, in SPEC §5.3
table order, which `TestSearchQueryParams` pins. A parser method called from `ParseSearchQuery` in that order, so errors
come out in it. A `SearchFilter` field. A `buildSearch` clause. The `appliedFilters` echo, which is a response field, so the
next checklist applies too. A `backend/openapi.yaml` parameter in the same order (`TestOpenAPISearchParamsMatchHandler`).
A SPEC §5.3 row. Frontend: a `SearchParams` field in `frontend/src/api/types.ts`, written by `buildSearchQuery` in the same
order and read by `parseSearchParams`. It drops names it does not read, so a param it misses is lost on the next push.

**Response-field checklist.** The DTO in `search.go`: `searchItem` and `newSearchItem`, `scoringMeta` and `newScoringMeta`,
or `appliedFilters` and `newAppliedFilters`. The SPEC §5.4 example, which `TestSearchRankedAll` compares to the live
response and `TestOpenAPIExamplesMatchSPEC` compares to the OpenAPI example. The wire types in `search_test.go`, such as
`wireItem`, which round-trip strictly. The `backend/openapi.yaml` schema (`ListingResult`, `ScoringMeta` or
`AppliedFilters`) and its examples. Frontend: `frontend/src/api/types.ts`, then `frontend/src/test/fixtures.ts`. The
fixtures are typed, so a new required field on `ListingResult` or `ScoringMeta` fails the typecheck until each has it.
A new field error code also goes in the SPEC §5.5 list, a code const in `query.go` and the `FieldError` enum in
`backend/openapi.yaml`; `TestOpenAPIErrorCodeEnums` lists it.

## Extension points

### 1. Cross-source dedupe

- **Requirement.** Show each property once, not once per feed. v1 keeps the 4 pairs visible on purpose (SPEC §4.2).
- **Touch.** New `Dedupe` in new `backend/internal/domain/dedupe.go`, called at the `Dedupe slot (section 9.1)` comment
  in `searchListings`. `Listing` gains new `Sources []string`. That is a domain type change, so the branch edits SPEC §4.
  `Listing` is never marshaled: `searchItem` and `newSearchItem` add a new `sources` field with at least one entry.
- **API.** The response-field checklist; no param. On the sample, `totalItems` goes from 12 to 8, so the SPEC §5.4
  example reads `totalItems` 8 and `totalPages` 1.
- **UI.** `formatListingMeta` appends `2 sources` after the source item (BRANDING §5.4, §10).
- **Tests.** New `dedupe_test.go` in the same folder: the 4 pairs, and order independence through `permutations`.
  A `search_test.go` case. `assertItemsMatchSeed` drops `sources` the way it drops `rank` and `relevance`. The totals in
  `TestSearchPaging` and `TestSearchStatusFilter` change with the count. `listing.test.ts` and `ListingCard.test.tsx`.
- **Estimate.** 45 min. The Rehearsal section has the key, the winner and the order.

### 2. Status filter

- **Requirement.** Filter by active, pending or sold. Built in v1 (SPEC §5.3, §7.4).
- **Touch.** Nothing. It is the worked example of the whole param path: `statuses()` reads the values with `ParseStatus`
  and `AllStatuses` and fills `SearchFilter.Statuses`. `whereStatusIn` applies it in SQL, and the `appliedFilters` echo
  returns it in enum order. On the frontend, `FIELDS` has a `status` entry that `FilterBar` renders in its
  `kind === 'status'` branch.
- **API.** Done. `status` repeats or comma-separates, and any bad piece gives one `invalid_enum`.
- **UI.** Done. One select: Any, Active, Pending, Sold.
- **Tests.** `TestParseSearchQueryStatus`, `TestSearchStatusFilter`.
- **Estimate.** 0 min. Multi-select is 30 min: the API already takes several values, but a checkbox group is a new
  control, so BRANDING changes first (BRANDING §9).

### 3. Geo radius

- **Requirement.** Listings within a radius of a point. Three new params: `lat`, `lng` and `radiusKm` (SPEC §9.3).
- **Touch.** New `BoundingBox` and new `HaversineKm` in new `backend/internal/domain/geo.go`. Two clauses in
  `buildSearch` beside `whereMinBedrooms`. A new `WithinRadius` stage in `searchListings`, after search and before dedupe.
- **API.** The param checklist, three times. One new error code, `missing_param`, so the error-code steps apply.
- **UI.** Three new `FIELDS` entries after Status, with a new `FieldKind` value for signed decimals. Or one point from
  wherever the interviewer says. A new `distanceKm` per item joins the meta line after the location (BRANDING §5.4).
- **Tests.** New `geo_test.go`, `TestParseSearchQueryRules`, `TestBuildSearch`, `TestSearchFilters`, a `search_test.go`
  case.
- **Estimate.** 50 min for the API, 30 min for the UI. The Rehearsal section has the rules and the sample numbers.

### 4. Sort options

- **Requirement.** Order by price or date as well as relevance (SPEC §9.4).
- **Touch.** New `SortOrder` and new `ParseSort` beside `ParseStatus`. `RankInput` gains new `Sort`. The one real change
  is the comparator in `SortStableFunc(out, byRelevance)`.
- **API.** New `sort=relevance|price|newest`: the param checklist, plus the response-field checklist for the new
  `scoring.sort` echo.
- **UI.** A select on the status line's right side, by BRANDING §5.7's page-size precedent. `formatOrdering` gains a new
  string, `lowest price first`.
- **Tests.** `TestRank` rows for each order, `TestParseSearchQueryRules`, a `search_test.go` case, `listing.test.ts`,
  `useSearchParamsState.test.tsx`, `StatusLine.test.tsx`.
- **Estimate.** 45 min. The Rehearsal section has the details.

### 5. More filters

- **Requirement.** New `minBathrooms`, `minSqft` and `zip` filters (SPEC §9.5).
- **Touch.** The param checklist for each. `minBathrooms` reads with `nonNegative`, so halves are allowed. `minSqft`
  follows `minBedrooms()`: a whole number, 0 or more. `zip` is trimmed, matched exactly, and `too_long` past 10 characters,
  so ZIP+4 fits. Each clause goes beside `whereMinBedrooms`.
- **API.** Three params, three echoes, three `backend/openapi.yaml` parameters. No new error codes.
- **UI.** New `FIELDS` entries after Status. Status is the seventh field, so these are the eighth on; BRANDING §5.2 still
  says "seventh". The grid already wraps. Rules go in `validate`, values in `draftsFrom` and `toSearchParams`. New labels
  get a SPEC §13 note.
- **Tests.** Both checklists' tests, plus `filterBarModel.test.ts` and `FilterBar.test.tsx`.
- **Estimate.** 30 min for the first, 10 min for each after.

### 6. A new feed

- **Requirement.** Load a third feed beside MLS A and MLS B (SPEC §9.6).
- **Touch.** `ListingsFile` holds one path, and `backend/cmd/api/main.go` calls `seed.Load(cfg.ListingsFile)` once. Make
  `LISTINGS_FILE` comma-separated (the `parseOrigins` pattern) and loop over the files. Check (source, id) across files:
  `Load` checks only within one file, and `Upsert` would let the later file win without a word. A feed with another shape
  needs its own record type, because `DisallowUnknownFields` rejects extra keys. `Upsert` never deletes, so a listing
  dropped from a feed stays in a file-backed database.
- **API.** None. Identity is already (source, id).
- **UI.** None. `formatSource` turns a new `MLS_C` source into `MLS C`.
- **Tests.** `TestLoad` rows for the list. A cross-file duplicate test beside `TestLoadSample`.
- **Estimate.** 25 min. Add the file to the `LISTINGS_FILE` line in `compose.yaml` too.

### 7. Page size control

- **Requirement.** Let the user choose how many results a page shows (SPEC §9.7).
- **Touch.** Frontend only. The API is done: `pageSize()` takes 1 to 100 (`MaxPageSize`).
- **API.** None.
- **UI.** A select on the status line's right side (BRANDING §5.7): 10, 25 or 50, labelled "Page size". A URL value
  outside the list shows as an extra option, as City does. It sits outside the `role="status"` live region, which should
  hold only the text it announces. New `setPageSize` writes `pageSize` and drops `page`; `setSearch` and `clear` already
  keep `pageSize`. The branch edits SPEC §7.1 ("never written by the UI") and the `pageSize` row in SPEC §5.3.
- **Tests.** `useSearchParamsState.test.tsx`, `StatusLine.test.tsx`, `SearchPage.test.tsx`.
- **Estimate.** 25 min.

### 8. Scale

- **Requirement.** Hold up at many more listings (SPEC §9.8). A talk item, not a build.
- **Touch.** Score in SQL (What I would say, below). Keyset pagination needs a cursor on the full sort key, and still a
  count for `totalPages` and `rank`. FTS5 turns substring matching into token matching, which changes SPEC §5.3 and §6.1;
  the driver, modernc.org/sqlite v1.60.1, is compiled with FTS5. A new `0002_*.sql` migration adds an index on
  (latitude, longitude) next to the 4 in `0001_listings.sql`; its leading column makes the geo box a range scan.
- **API.** `keyword` matches whole tokens under FTS5. Everything else keeps its shape.
- **UI.** Pager windowing. `Pager` renders every page number. A gap marker needs a glyph that BRANDING §5.10 forbids, so
  it starts as a BRANDING change. SPEC §7.4 points here for this.
- **Tests.** `TestBuildSearch` for each SQL change, `TestSearchFilters` for keyword, `Pager.test.tsx` for windowing.
- **Estimate.** Talk. Each piece is its own ticket.

## Rehearsal

### Cross-source dedupe

1. v1 keeps duplicates visible on purpose. The brief says ids are unique per source and nothing more, so the card shows
   the source (SPEC §4.2).
2. The stage goes after the SQL filter and before `Rank`. Scores, ranks and `totalItems` then count only survivors.
3. The key is the normalized address plus the coordinates rounded to 4 decimals, about 11 m. Normalizing lowercases,
   drops punctuation, and maps St to Street, Ave to Avenue, Ct to Court, Rd to Road and Apt to Unit.
4. Not zip. A2 and B8 are the same house with zips 22150 and 22151. SPEC §9.1's address-plus-zip key misses them, and its
   list lacks Ct/Court, which A5 and B11 need.
5. Not coordinates alone. Units in one building share a point.
6. The record with the newest `listedDate` wins: A1, A2, A3 and A5. Ties go to source, then id. The whole record wins,
   not a field-by-field merge, so a card never mixes two feeds' prices.
7. A new `sources` field lists every feed in the group, so A1's meta line ends `MLS A · 2 sources`.
8. At `targetBudget=450000` and `asOf` 2026-10-13 (the SPEC §5.4 date), the 8 survivors run A1, A5, A3, A6, A2, A7, B10, A4.
   Recency moves with the date, so the order can change day to day.
9. One caveat: the stage sees only what the filter kept. `maxPrice=451000` drops B7 ($452,000) first, so A1 shows one
   source.
10. At scale, a dedupe key column set at seed time replaces the Go pass.

Ask: should a new `dedupe=false` toggle bring the duplicates back?

### Geo radius

1. Three new params: `lat` in [-90, 90], `lng` in [-180, 180], `radiusKm` in (0, 100]. All three or none.
2. A missing one gets the new `missing_param` code on its own name, so the user sees which one. A value out of range
   reuses `out_of_range`, and a radius of 0 or less reuses `not_positive`.
3. SQL first. New `BoundingBox` turns the point and radius into two BETWEEN clauses in `buildSearch`: dLat = km / 111.32,
   and dLng = km / (111.32 × cos lat). The box is a superset of the circle and cheap. The new 0002 migration adds an
   index on (latitude, longitude); its leading column makes the box a range scan.
4. Go second. New `HaversineKm`, with an Earth radius of 6371.0088 km, runs in a new `WithinRadius` stage in
   `searchListings`, after search and before dedupe. It trims the box's corners to the circle.
5. The `Searcher` comment then says the store returns the box, not the final set.
6. On the sample, 2 km around A1 (38.7893, -77.1873) returns A1, B7, A2 and B8. A2 is 1.16 km away. A3 and B9, the next
   nearest, are 12.1 km away.
7. A new `distanceKm` per item can join the meta line after the location (BRANDING §5.4). A new `sort=distance` follows
   from the Sort options work.
8. The limit, said out loud: the box does not handle the antimeridian or the poles. Northern Virginia is far from both.
9. Tests: new `geo_test.go`, `TestParseSearchQueryRules`, `TestBuildSearch`, `TestSearchFilters` and a `search_test.go`
   case.

Ask: where does the point come from: a typed address, the city's center, or the browser's location?

### Sort options

1. New `sort=relevance|price|newest`, with `relevance` the default. A bad value is `invalid_enum`, as for `status`.
2. New `SortOrder` and new `ParseSort` go beside `ParseStatus` in `backend/internal/domain/listing.go`. `RankInput`
   gains new `Sort`.
3. The one real change is the comparator in `SortStableFunc(out, byRelevance)`. The `byRelevance` comment already says a
   new sort option sits beside it.
4. Price is price ascending, then `byRelevance`. Newest is `listedDate` descending, then `byRelevance`. The order stays
   total because `byRelevance` ends at (source, id), so pages never shift.
5. `rank` still means position (BRANDING §5.6). Under price order the scores do not descend, and the UI shows that as it
   is. On the 12 listings, without dedupe, at `targetBudget=450000` and `asOf` 2026-10-13, A3 (0.68) is No. 1 and
   A1 (0.74) is No. 4.
6. The response echoes the order as new `scoring.sort`. `formatOrdering` gains the new string `lowest price first`;
   newest reuses `newest first`.
7. The control is a select on the status line's right side, by BRANDING §5.7's page-size precedent, outside
   `role="status"`. BRANDING has no slot for sort, so the branding owner gets a note.
8. `setSearch` and `clear` keep the new `sort` param the way they keep `pageSize`. A new `setSort` writes it and drops
   `page`. `filterKey` must drop it beside `page` and `pageSize`, so a sort change keeps the drafts.

Ask: is high-to-low price wanted too?

## What I would say

### Scoring weights

I score 0.6 budget fit plus 0.4 recency (SPEC §6.2). Budget leads because it is what the user asked for. Going over
budget costs twice as much as coming in under: budget fit is 0.9 at 10% under, 0.8 at 10% over and 0 at 50% over. Recency
halves every 30 days, so a month-old listing counts half. With no budget the score is recency alone, which is newest first.
The weights are constants, `BudgetWeight` and `RecencyWeight`, not config, and the `scoring` block echoes them.
At `asOf` 2026-10-13, A1 scores 0.6 × 1.0 + 0.4 × 0.3536 = 0.7414; recency moves that day to day.
To tune them I would move them into `RankInput`, add `TestScore` rows, and fit them to click data.

### SQL versus Go ranking

I filter in SQL, where the indexes are, and score in pure Go with the clock passed in (SPEC §6). Tests pin exact
scores, and ties compare unrounded (SPEC §6.3), so the order is exact. The cost is that every filtered row loads into Go
on every request, which is fine at this scale. The path past it: precompute the inputs (a dedupe key; price already has
an index), compute the score in SQL with the request's target and today's date, keep the same tiebreak, LIMIT the page,
then move to keyset pagination. I would not store the score itself: recency changes daily and budget fit depends on the
request.
