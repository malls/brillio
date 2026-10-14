# BC-16: Extension playbook (docs/EXTENSIONS.md)

Docs-only spike, complexity low. Writes one new file, `docs/EXTENSIONS.md`. No code, and no SPEC, BRANDING or README edits. Base: feat/BC-12 @5ac2397.
**BC-6 decision:** BC-6 landed on master during planning (merge 269f686), so the doc names its real file and tests (`backend/openapi.yaml`,
`TestOpenAPISearchParamsMatchHandler` and so on). They are not on this branch until the delegator merges master, which is required anyway after BC-12
lands and before review. Until then the implementer checks those lines with `git grep -nE '<pat>' master -- <path>`. The implementer runs no git
operations beyond writing the file.

## Outline of docs/EXTENSIONS.md (target 240-285 lines, hard cap 299)
1. `# Extension playbook`: 4-6 lines. Every backticked name exists, grepped at the commit the intro names, except names written as "new `X`" (proposals).
2. `## Where things go`: about 12 lines. The seams, plus two shared checklists that entries point to instead of repeating them.
   **Param checklist:** const and entry in `searchParams` (SPEC §5.3 table order), a parser method called from `ParseSearchQuery`, a `SearchFilter` field,
   a `buildSearch` clause, the `appliedFilters` echo, a `backend/openapi.yaml` parameter, a SPEC §5.3 row; FE `SearchParams`, `buildSearchQuery`
   (same order), `parseSearchParams` (it drops unknown names, so a param it does not read is lost on the next push).
   **Response-field checklist:** DTO in search.go, the SPEC §5.4 example (asserted by `TestSearchRankedAll` and `TestOpenAPIExamplesMatchSPEC`), the
   search_test.go wire types (strict round trip), the openapi.yaml schema (`ListingResult`/`ScoringMeta`/`AppliedFilters`), types.ts, fixtures.ts.
   A new field error code also goes in the SPEC §5.5 list and the `FieldError` enum (`TestOpenAPIErrorCodeEnums`).
3. `## Extension points`: `### 1. Cross-source dedupe` to `### 8. Scale`, numbered as SPEC §9. Each is under 25 lines, with the labels Requirement / Touch / API / UI / Tests / Estimate.
4. `## Rehearsal`: `### Cross-source dedupe`, `### Geo radius`, `### Sort options`. Numbered steps to say aloud (what, then why), about 20-28 lines each, ending in one question to ask the interviewer.
5. `## What I would say`: `### Scoring weights`, `### SQL versus Go ranking`. 6-9 lines each, first person.

## Symbol table (every line run in the worktree 2026-10-14; each returned its count, BC-6 lines on master)
The doc names only these existing symbols. If it needs another existing name, add a grep line here first.
```
grep -nE 'func \(a \*api\) searchListings\(|Dedupe slot \(section 9\.1\)|^type searchItem struct|^func newSearchItem\(|^type appliedFilters struct|^func newAppliedFilters\(|^type scoringMeta struct|^func newScoringMeta\(|^type Searcher interface' backend/internal/httpapi/search.go  # 9
grep -nE '^var searchParams = |^func SearchQueryParams\(|^func ParseSearchQuery\(|^func \(p \*parser\) (statuses|minBedrooms|nonNegative|pageSize|unknownParams)\(|(codeInvalidEnum|codeOutOfRange|codeNotPositive) +=' backend/internal/httpapi/query.go  # 11
grep -nE 'param[A-Za-z]+ += "' backend/internal/httpapi/query.go  # 9 (SPEC §5.3 param names)
grep -nE '^func Rank\(|^type RankInput struct|^func byRelevance\(|SortStableFunc\(out, byRelevance\)' backend/internal/domain/rank.go  # 4
grep -nE '^func (BudgetFit|Recency|Score)\(|(BudgetWeight|RecencyWeight|RecencyHalfLifeDays) += ' backend/internal/domain/scoring.go  # 6
grep -nE '^func Paginate\(|^const MaxPageSize = 100' backend/internal/domain/paginate.go  # 2
grep -nE '^type SearchFilter struct' backend/internal/domain/filter.go  # 1
grep -nE '^type Listing struct|^func (ParseStatus|AllStatuses)\(' backend/internal/domain/listing.go  # 3
grep -nE '^func \(s \*Store\) Search\(|^func (buildSearch|scanListing)\(|(whereMinBedrooms|whereStatusIn|orderBySourceID) += ' backend/internal/store/listings.go  # 6
grep -nE '^CREATE INDEX' backend/internal/store/migrations/0001_listings.sql  # 4 (next file is new 0002_*.sql)
grep -nE '^func Load\(|^type RecordError struct|DisallowUnknownFields' backend/internal/store/seed/seed.go  # 3
grep -nE '^\s+ListingsFile string|^func (Load|parseOrigins)\(' backend/internal/config/config.go  # 3
grep -nE 'seed\.Load\(cfg\.ListingsFile\)' backend/cmd/api/main.go; grep -nE 'LISTINGS_FILE:' compose.yaml  # 1 + 1
grep -nE '^func (TestParseSearchQueryRules|TestParseSearchQueryStatus|TestSearchQueryParams)\(' backend/internal/httpapi/query_test.go  # 3
grep -nE '^func (TestSearchRankedAll|TestSearchStatusFilter|TestSearchPaging|assertItemsMatchSeed)\(|^type wireItem struct' backend/internal/httpapi/search_test.go  # 5
grep -nE '^func (TestRank|permutations)\(' backend/internal/domain/rank_test.go; grep -nE '^func TestScore\(' backend/internal/domain/scoring_test.go  # 2 + 1
grep -nE '^func TestBuildSearch\(' backend/internal/store/store_internal_test.go; grep -nE '^func TestSearchFilters\(' backend/internal/store/store_test.go  # 1 + 1
grep -nE '^func TestLoadSample\(' backend/internal/store/seed/seed_test.go; grep -nE '^func TestLoad\(' backend/internal/config/config_test.go  # 1 + 1
grep -nE '^ +- name: (minPrice|pageSize)$|^    (ListingResult|ScoringMeta|AppliedFilters|FieldError):$' backend/openapi.yaml  # 6 [BC-6]
grep -nE '^func (TestOpenAPISearchParamsMatchHandler|TestOpenAPIExamplesMatchSPEC|TestOpenAPIErrorCodeEnums)\(' backend/internal/httpapi/openapi_test.go  # 3 [BC-6]
grep -nE "^export const (FIELDS|COPY) = |^type FieldKind = |name: 'status', label: 'Status'|^export function (draftsFrom|validate|toSearchParams|filterKey)\(" frontend/src/components/FilterBar/filterBarModel.ts  # 8
grep -nE "^export function FilterBar\(|kind === 'status'" frontend/src/components/FilterBar/FilterBar.tsx  # 2
grep -nE '^export (function StatusLine\(|interface StatusLineProps)|^function describeState\(|role="status"' frontend/src/components/StatusLine/StatusLine.tsx  # 4
grep -nE '^export function (formatListingMeta|formatSource|formatOrdering|formatResultSummary)\(' frontend/src/format/listing.ts  # 4
grep -nE '^export function ListingCard\(' frontend/src/components/ListingCard/ListingCard.tsx  # 1
grep -nE '^export function SearchPage\(|<StatusLine |<FilterBar' frontend/src/pages/SearchPage.tsx  # 3 [BC-12]
grep -nE '^export function (useSearchParamsState|parseSearchParams)\(|const (setSearch|clear) = useCallback' frontend/src/hooks/useSearchParamsState.ts  # 4
grep -nE '^export interface (SearchParams|ListingResult|ScoringMeta) ' frontend/src/api/types.ts; grep -nE '^export function buildSearchQuery\(' frontend/src/api/client.ts  # 3 + 1
grep -nE '^export function Pager\(|Array\.from\(\{ length: totalPages \}' frontend/src/components/Pager/Pager.tsx  # 2
grep -nE '^export const (listingA1|listingB7|searchResponse): ' frontend/src/test/fixtures.ts  # 3
ls frontend/src/{components/FilterBar/{filterBarModel.test.ts,FilterBar.test.tsx},components/StatusLine/StatusLine.test.tsx,components/ListingCard/ListingCard.test.tsx,format/listing.test.ts,hooks/useSearchParamsState.test.tsx,api/client.test.ts,pages/SearchPage.test.tsx,components/Pager/Pager.test.tsx}  # 9 files; SearchPage.test.tsx [BC-12]
```

## Facts per §9 item ("new" marks a proposal; the implementer writes the prose)
1. **Dedupe** (45 min). Touch: new `Dedupe` (new file domain/dedupe.go), called at the `Dedupe slot (section 9.1)` comment in `searchListings`. `Listing`
   gains new `Sources []string`; that is a domain type change, so that branch edits SPEC §4, and the field is never marshaled. `searchItem`/`newSearchItem`
   add `sources` (always at least one entry). API: response checklist, no param; totalItems on the sample goes from 12 to 8. UI: `formatListingMeta` appends `2 sources` after
   the source item (BRANDING §10, §5.4). Tests: new dedupe_test.go (the 4 pairs, order independence via `permutations`), a search_test.go case,
   `assertItemsMatchSeed` drops `sources` the way it drops rank and relevance, listing.test.ts, ListingCard.test.tsx.
2. **Status** (built, 0 min). The worked example of the whole path: `statuses()` to `ParseStatus`/`AllStatuses` to `SearchFilter.Statuses` to `whereStatusIn` to the `appliedFilters` echo to FIELDS `status` and the
   `kind === 'status'` branch. Tests: `TestParseSearchQueryStatus`, `TestSearchStatusFilter`. Multi-select (30 min): the API already takes repeats; a checkbox group is a new control (BRANDING §9 change).
3. **Geo radius** (50 min API, 30 UI) and 4. **Sort** (45 min): the facts are in the Rehearsal section below.
5. **More filters** (30 min first, 10 each after). `minBathrooms` uses `nonNegative` (halves allowed); `minSqft` follows the `minBedrooms` pattern; `zip` is trimmed, exact match, `too_long`
   past 10 characters. Param checklist; clauses go beside `whereMinBedrooms`. UI: new FIELDS entries after Status, from the eighth field on (the grid already
   wraps); rules go in `validate`/`draftsFrom`/`toSearchParams`; new labels get a SPEC §13 note. Tests: those checklists plus filterBarModel.test.ts and FilterBar.test.tsx.
6. **New feed** (25 min). `ListingsFile` holds one path and main.go calls `seed.Load(cfg.ListingsFile)` once. Make LISTINGS_FILE comma-separated (the `parseOrigins` pattern),
   loop over the files, and check (source, id) across files: `Load` only checks within one file, and Upsert would let the later file win silently. A feed with another shape
   needs its own record type (`DisallowUnknownFields`). Upsert never deletes. API/UI: none (`formatSource` gives `MLS C`). Tests: `TestLoad` rows; a cross-file test by `TestLoadSample`.
7. **Page size** (25 min). The API is done (`pageSize()`, `MaxPageSize`). UI: a select on the status line's right side (BRANDING §5.7; SPEC §9.7's citation is correct), 10/25/50,
   label "Page size", outside `role="status"`. New `setPageSize` writes pageSize and drops page. That branch edits SPEC §7.1 ("never written by the UI") and §5.3. Tests: hook, StatusLine, SearchPage.
8. **Scale** (talk). Score in SQL (What I would say). Keyset pagination still needs a count for totalPages and rank. FTS5 turns `instr` substring matching into token matching
   (SPEC §5.3/§6.1 change; check the driver has FTS5). A new 0002 migration adds a lat/lng index. Pager windowing: `Pager` renders every page, and a gap marker needs a glyph
   BRANDING §5.10 forbids (SPEC §7.4 points here for this).

## Rehearsal mini-plans (facts to narrate)
- **Dedupe:** v1 keeps duplicates visible on purpose (SPEC §4.2). Stage: after the SQL filter, before `Rank`, so scores, ranks and totalItems only see survivors. Key: normalized
  address (lowercase, drop punctuation, St/Street, Ave/Avenue, Ct/Court, Rd/Road, Apt/Unit) plus coordinates rounded to 4 decimals (about 11 m). Not zip: A2/B8 are 22150 vs 22151. Not
  coordinates alone: units in one building share a point. Winner: the newest listedDate (A1, A2, A3, A5), ties by source then id; the whole record wins. Order at targetBudget=450000:
  A1, A5, A3, A6, A2, A7, B10, A4. Caveat: maxPrice=451000 filters out B7 (452000) first, so A1 shows one source; at scale, a dedupe key column set at seed time. Ask: a `dedupe=false` toggle?
- **Geo radius:** `lat` [-90, 90], `lng` [-180, 180], `radiusKm` (0, 100], all three or none (new code `missing_param` on the missing one). SQL: two BETWEEN clauses in `buildSearch` from a
  new `BoundingBox` (dLat = km/111.32, dLng = km/(111.32 cos lat)): a superset of the circle, cheap, uses an index. Go: new `HaversineKm` (R 6371.0088) in a new `WithinRadius` stage in
  `searchListings`, after search and before dedupe; the `Searcher` comment then says the store returns the box. Sample: 2 km around A1 (38.7893, -77.1873) gives A1, B7, A2, B8
  (A2 is 1.16 km away; A3/B9 are 12.1 km). Optional `distanceKm` becomes a meta item after the location (BRANDING §5.4), and `sort=distance` follows. The antimeridian and poles are not handled; say so.
  Tests: new geo_test.go, `TestParseSearchQueryRules`, `TestBuildSearch`, `TestSearchFilters`, a search_test.go case. Ask: where does the point come from?
- **Sort:** `sort=relevance|price|newest`, default relevance, a bad value is `invalid_enum`. New `SortOrder`/`ParseSort` go beside `ParseStatus`; `RankInput` gains `Sort`. The one change:
  the comparator in `SortStableFunc(out, byRelevance)`. Price is price ascending, then `byRelevance`; newest is listedDate descending, then `byRelevance`. The order stays total because `byRelevance` ends at
  (source, id). `rank` still means position (BRANDING §5.6); under price order the scores are not descending, and the UI shows that as it is. Echo as `scoring.sort`; `formatOrdering` gains
  `lowest price first`. UI: a select on the right side by §5.7's page-size precedent, outside `role="status"`; `setSearch`/`clear` keep sort the way they keep pageSize; `filterKey` drops it. Ask: high-to-low price too?

## What I would say (cite SPEC §6.2, §6.3, the §6 trade-off)
- Weights: 0.6 budget fit + 0.4 recency. Over budget costs twice as much (10% under is 0.9, 10% over is 0.8, zero at 50% over). Recency halves every 30 days. With no budget the score is recency alone, so newest first.
  The weights are constants in `domain` (not config), and the `scoring` block echoes them. A1 = 0.6 x 1.0 + 0.4 x 0.3536 = 0.7414. To tune: move them to `RankInput`, add `TestScore` rows, use click data.
- SQL versus Go: filter in SQL on indexed columns; score in pure Go with an injected clock, so tests and ties are exact (unrounded, §6.3). Cost: every filtered row loads per request, fine at this scale.
  Path: a score expression in SQL with bound target and now, the same tiebreak, LIMIT, then keyset. A precomputed column does not fit: recency moves daily and budget fit depends on the request.

## Acceptance tests (one per ticket bullet) and validation
1. Symbols exist: each grep line returns its count (BC-6 lines through `git grep ... master` until the merge); every token from `grep -oE '`[^`]+`' docs/EXTENSIONS.md | sort -u`
   is in the table, preceded by "new", or a SPEC param/field/code or a copy string.
2. Voice: read against BRANDING §7; `grep -nE '!|[Pp]lease|[Oo]ops|[Ss]orry|[Ss]imply|[Jj]ust ' docs/EXTENSIONS.md` returns no prose hits.
3. No emoji: `perl -CSD -ne 'print "$.: $_" if /[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{2B00}-\x{2BFF}\x{FE0F}\x{200D}\x{20E3}]/' docs/EXTENSIONS.md` prints nothing.
4. Length: `test $(wc -l < docs/EXTENSIONS.md) -lt 300`. Scope: `grep -cE '^### [1-8]\. '` is 8, and every count from
   `awk '/^#{1,3} /{if(h)print n, h; h=""} /^### [1-8]\. /{h=$0; n=0} h{n++} END{if(h)print n, h}' docs/EXTENSIONS.md` is under 25; Rehearsal has 3 subsections, What I would say has 2.
Validation: N/A (documentation; no ports). With `--role validation` the validator records the grep block output, the token reconciliation, `wc -l`, the empty emoji check and the awk counts.

## Shared files, risks, size
- Only docs/EXTENSIONS.md, which no other ticket owns. Leave README.md (BC-15 links it), SPEC.md and BRANDING.md alone.
- Risks: symbol drift if BC-12 changes before landing or BC-14 renames something, so re-run the block after the master merge. The doc can go stale later, so the intro names the commit it was checked at.
- Size: about 260 lines of doc, about 90 min including verification.

## Findings (no SPEC edit needed by this ticket)
- The ticket says geo has "two new params"; SPEC §9.3 names three (`lat`, `lng`, `radiusKm`). The plan follows SPEC.
- BRANDING §5.2's "a new filter joins as a seventh field" is stale: Status is the seventh (SPEC §7.4, §13 note 1), so new filters start at eight.
- SPEC §9.1's "address + zip" key misses A2/B8, and its list of expansions lacks Ct/Court (A5/B11). The doc says so; §9 is design notes, not contract.
- BRANDING has no slot for sort: the doc uses the §5.7 precedent and says the branding owner needs a note when it is built. StatusLine is one `role="status"` live region, so any select must sit outside it.
