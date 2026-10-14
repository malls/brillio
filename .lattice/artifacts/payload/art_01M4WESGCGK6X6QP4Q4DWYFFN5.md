# BC-16 validation: docs/EXTENSIONS.md (extension playbook)

- Validator: agent:opus-5.5-validator
- Branch: feat/BC-16-extensions-doc
- HEAD checked: 3fc1aa41a59100e9ac99f08d0a665392c13da6df
- Master checked: c8837b57a3ae79d6e88d620c55cc4fd307cf7bff
- Merge-base: 0b24829a014593929e29b45a8e77f3a9949116b6 (the commit the doc intro names)
- Date: 2026-10-14 (run at 05:33Z)
- Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-16-extensions-doc
- Plan: .lattice/plans/task_01M4VXK3WPPJA1GVM0NZRW4GGD.md (Symbol table, Plan-Review Cycle 1 Resolutions item 2, Review Cycle 1 Findings)

Validation N/A (documentation): docs/EXTENSIONS.md is not runtime behavior; evidence is that every named symbol exists on HEAD and master, per validation-plan row 52

## Verdict table

| Item (validation-plan row 52 + acceptance) | Result | Evidence |
|---|---|---|
| One entry per SPEC §9 item | PASS | 8 `### N.` headings; titles match SPEC §9 items 1-8 (docs/SPEC.md:510-517) in order |
| Every referenced symbol exists (HEAD) | PASS | 37 symbol-table grep lines, 37 OK, 0 mismatches; `ls` line 9/9 files |
| Every referenced symbol exists (master c8837b5) | PASS | same 37 lines via `git grep ... master`, 37 OK, 0 mismatches; `git ls-tree` 9/9 files |
| master drift since 0b24829 touches no named file | PASS | `git diff --name-only 0b24829 master` = README.md + .lattice/* only; doc names neither |
| Backtick-token reconciliation | PASS | 157 unique tokens: 104 existing (a), 31 "new" proposals (b), 22 SPEC/BRANDING contract or copy (c), 0 unclassified; all 13 checked proposals absent on master |
| Rehearsal covers dedupe, geo radius, sort | PASS | `### Cross-source dedupe` (140), `### Geo radius` (161), `### Sort options` (182) under `## Rehearsal` |
| What I would say has 2 subsections | PASS | `### Scoring weights` (205), `### SQL versus Go ranking` (214) |
| Under 300 lines | PASS | 221 lines |
| Each §9 entry under 25 lines (awk) | PASS | 14, 13, 12, 13, 13, 13, 12, 13 |
| No emoji | PASS | perl scan printed nothing |
| Voice (no `!`, please, oops, sorry, simply, just) | PASS | grep returned 0 hits (exit 1), so no code-span or prose hits to classify |
| Scope | PASS | `git diff --name-only master...HEAD` = docs/EXTENSIONS.md only; `git status --porcelain` empty |

Overall: PASS

## Step 1a. Symbol-table grep block, worktree at HEAD

Lines L01-L35 are the plan's "## Symbol table" block (plan lines 25-54; the five two-grep lines 37, 40, 41, 42 and 52 are split into two Lxx each). L12 uses `^[[:space:]]+ListingsFile string` per Plan-Review Cycle 1 Resolution 2. L36-L37 are the two "Review Cycle 1 Findings" lines (run with `-nE` for hit lines instead of `-cE`; same patterns). LS is plan line 55. Each grep ran as `grep -HnE '<pat>' <path>` (`-H` only adds the file name to each hit).

```
L01 $ grep -nE 'func \(a \*api\) searchListings\(|Dedupe slot \(section 9\.1\)|^type searchItem struct|^func newSearchItem\(|^type appliedFilters struct|^func newAppliedFilters\(|^type scoringMeta struct|^func newScoringMeta\(|^type Searcher interface' backend/internal/httpapi/search.go   # expected 9
backend/internal/httpapi/search.go:15:type Searcher interface {
backend/internal/httpapi/search.go:22:func (a *api) searchListings(w http.ResponseWriter, r *http.Request) {
backend/internal/httpapi/search.go:37:	// Dedupe slot (section 9.1): merging cross-source duplicates goes here,
backend/internal/httpapi/search.go:68:type searchItem struct {
backend/internal/httpapi/search.go:105:type scoringMeta struct {
backend/internal/httpapi/search.go:114:type appliedFilters struct {
backend/internal/httpapi/search.go:141:func newSearchItem(it domain.RankedItem) searchItem {
backend/internal/httpapi/search.go:171:func newScoringMeta(targetBudget *float64, now time.Time) scoringMeta {
backend/internal/httpapi/search.go:188:func newAppliedFilters(f domain.SearchFilter) appliedFilters {
-> count 9 / expected 9 : OK

L02 $ grep -nE '^var searchParams = |^func SearchQueryParams\(|^func ParseSearchQuery\(|^func \(p \*parser\) (statuses|minBedrooms|nonNegative|pageSize|unknownParams)\(|(codeInvalidEnum|codeOutOfRange|codeNotPositive) +=' backend/internal/httpapi/query.go   # expected 11
backend/internal/httpapi/query.go:33:var searchParams = [...]string{
backend/internal/httpapi/query.go:41:func SearchQueryParams() []string {
backend/internal/httpapi/query.go:54:	codeNotPositive    = "not_positive"
backend/internal/httpapi/query.go:55:	codeOutOfRange     = "out_of_range"
backend/internal/httpapi/query.go:57:	codeInvalidEnum    = "invalid_enum"
backend/internal/httpapi/query.go:148:func ParseSearchQuery(v url.Values) (SearchQuery, []FieldError) {
backend/internal/httpapi/query.go:240:func (p *parser) nonNegative(name, label string) *float64 {
backend/internal/httpapi/query.go:250:func (p *parser) minBedrooms() *int {
backend/internal/httpapi/query.go:306:func (p *parser) statuses() []domain.ListingStatus {
backend/internal/httpapi/query.go:352:func (p *parser) pageSize() int {
backend/internal/httpapi/query.go:366:func (p *parser) unknownParams() {
-> count 11 / expected 11 : OK

L03 $ grep -nE 'param[A-Za-z]+ += "' backend/internal/httpapi/query.go   # expected 9
backend/internal/httpapi/query.go:19:	paramMinPrice     = "minPrice"
backend/internal/httpapi/query.go:20:	paramMaxPrice     = "maxPrice"
backend/internal/httpapi/query.go:21:	paramMinBedrooms  = "minBedrooms"
backend/internal/httpapi/query.go:22:	paramCity         = "city"
backend/internal/httpapi/query.go:23:	paramKeyword      = "keyword"
backend/internal/httpapi/query.go:24:	paramTargetBudget = "targetBudget"
backend/internal/httpapi/query.go:25:	paramStatus       = "status"
backend/internal/httpapi/query.go:26:	paramPage         = "page"
backend/internal/httpapi/query.go:27:	paramPageSize     = "pageSize"
-> count 9 / expected 9 : OK

L04 $ grep -nE '^func Rank\(|^type RankInput struct|^func byRelevance\(|SortStableFunc\(out, byRelevance\)' backend/internal/domain/rank.go   # expected 4
backend/internal/domain/rank.go:10:type RankInput struct {
backend/internal/domain/rank.go:56:func Rank(listings []Listing, in RankInput) []Scored {
backend/internal/domain/rank.go:67:	slices.SortStableFunc(out, byRelevance)
backend/internal/domain/rank.go:74:func byRelevance(a, b Scored) int {
-> count 4 / expected 4 : OK

L05 $ grep -nE '^func (BudgetFit|Recency|Score)\(|(BudgetWeight|RecencyWeight|RecencyHalfLifeDays) += ' backend/internal/domain/scoring.go   # expected 6
backend/internal/domain/scoring.go:28:	BudgetWeight = 0.6
backend/internal/domain/scoring.go:32:	RecencyWeight = 0.4
backend/internal/domain/scoring.go:34:	RecencyHalfLifeDays = 30
backend/internal/domain/scoring.go:49:func BudgetFit(price, target float64) float64 {
backend/internal/domain/scoring.go:74:func Recency(listedDate, now time.Time) float64 {
backend/internal/domain/scoring.go:83:func Score(budgetFit *float64, recency float64) float64 {
-> count 6 / expected 6 : OK

L06 $ grep -nE '^func Paginate\(|^const MaxPageSize = 100' backend/internal/domain/paginate.go   # expected 2
backend/internal/domain/paginate.go:10:const MaxPageSize = 100
backend/internal/domain/paginate.go:56:func Paginate(ranked []Scored, page, pageSize int) (Page, error) {
-> count 2 / expected 2 : OK

L07 $ grep -nE '^type SearchFilter struct' backend/internal/domain/filter.go   # expected 1
backend/internal/domain/filter.go:6:type SearchFilter struct {
-> count 1 / expected 1 : OK

L08 $ grep -nE '^type Listing struct|^func (ParseStatus|AllStatuses)\(' backend/internal/domain/listing.go   # expected 3
backend/internal/domain/listing.go:37:func AllStatuses() []ListingStatus {
backend/internal/domain/listing.go:44:func ParseStatus(s string) (ListingStatus, error) {
backend/internal/domain/listing.go:61:type Listing struct {
-> count 3 / expected 3 : OK

L09 $ grep -nE '^func \(s \*Store\) Search\(|^func (buildSearch|scanListing)\(|(whereMinBedrooms|whereStatusIn|orderBySourceID) += ' backend/internal/store/listings.go   # expected 6
backend/internal/store/listings.go:32:	whereMinBedrooms = "bedrooms >= ?"
backend/internal/store/listings.go:36:	whereStatusIn     = "status IN ("
backend/internal/store/listings.go:43:	orderBySourceID = " ORDER BY source, id"
backend/internal/store/listings.go:92:func (s *Store) Search(ctx context.Context, f domain.SearchFilter) ([]domain.Listing, error) {
backend/internal/store/listings.go:116:func buildSearch(f domain.SearchFilter) (string, []any) {
backend/internal/store/listings.go:164:func scanListing(rows *sql.Rows) (domain.Listing, error) {
-> count 6 / expected 6 : OK

L10 $ grep -nE '^CREATE INDEX' backend/internal/store/migrations/0001_listings.sql   # expected 4
backend/internal/store/migrations/0001_listings.sql:28:CREATE INDEX listings_city_nocase ON listings (city COLLATE NOCASE);
backend/internal/store/migrations/0001_listings.sql:29:CREATE INDEX listings_price ON listings (price);
backend/internal/store/migrations/0001_listings.sql:30:CREATE INDEX listings_bedrooms ON listings (bedrooms);
backend/internal/store/migrations/0001_listings.sql:31:CREATE INDEX listings_status ON listings (status);
-> count 4 / expected 4 : OK

L11 $ grep -nE '^func Load\(|^type RecordError struct|DisallowUnknownFields' backend/internal/store/seed/seed.go   # expected 3
backend/internal/store/seed/seed.go:52:type RecordError struct {
backend/internal/store/seed/seed.go:81:func Load(path string) ([]domain.Listing, error) {
backend/internal/store/seed/seed.go:123:	dec.DisallowUnknownFields()
-> count 3 / expected 3 : OK

L12 $ grep -nE '^[[:space:]]+ListingsFile string|^func (Load|parseOrigins)\(' backend/internal/config/config.go   # expected 3
backend/internal/config/config.go:52:	ListingsFile string
backend/internal/config/config.go:70:func Load(getenv func(string) string) (Config, error) {
backend/internal/config/config.go:139:func parseOrigins(raw string) ([]string, []error) {
-> count 3 / expected 3 : OK

L13 $ grep -nE 'seed\.Load\(cfg\.ListingsFile\)' backend/cmd/api/main.go   # expected 1
backend/cmd/api/main.go:56:	listings, err := seed.Load(cfg.ListingsFile)
-> count 1 / expected 1 : OK

L14 $ grep -nE 'LISTINGS_FILE:' compose.yaml   # expected 1
compose.yaml:37:      LISTINGS_FILE: /data/sample_listings.json
-> count 1 / expected 1 : OK

L15 $ grep -nE '^func (TestParseSearchQueryRules|TestParseSearchQueryStatus|TestSearchQueryParams)\(' backend/internal/httpapi/query_test.go   # expected 3
backend/internal/httpapi/query_test.go:83:func TestParseSearchQueryRules(t *testing.T) {
backend/internal/httpapi/query_test.go:340:func TestParseSearchQueryStatus(t *testing.T) {
backend/internal/httpapi/query_test.go:442:func TestSearchQueryParams(t *testing.T) {
-> count 3 / expected 3 : OK

L16 $ grep -nE '^func (TestSearchRankedAll|TestSearchStatusFilter|TestSearchPaging|assertItemsMatchSeed)\(|^type wireItem struct' backend/internal/httpapi/search_test.go   # expected 5
backend/internal/httpapi/search_test.go:56:type wireItem struct {
backend/internal/httpapi/search_test.go:293:func assertItemsMatchSeed(t *testing.T, body []byte) {
backend/internal/httpapi/search_test.go:444:func TestSearchRankedAll(t *testing.T) {
backend/internal/httpapi/search_test.go:541:func TestSearchPaging(t *testing.T) {
backend/internal/httpapi/search_test.go:601:func TestSearchStatusFilter(t *testing.T) {
-> count 5 / expected 5 : OK

L17 $ grep -nE '^func (TestRank|permutations)\(' backend/internal/domain/rank_test.go   # expected 2
backend/internal/domain/rank_test.go:72:func permutations(s []domain.Listing, visit func([]domain.Listing)) {
backend/internal/domain/rank_test.go:93:func TestRank(t *testing.T) {
-> count 2 / expected 2 : OK

L18 $ grep -nE '^func TestScore\(' backend/internal/domain/scoring_test.go   # expected 1
backend/internal/domain/scoring_test.go:169:func TestScore(t *testing.T) {
-> count 1 / expected 1 : OK

L19 $ grep -nE '^func TestBuildSearch\(' backend/internal/store/store_internal_test.go   # expected 1
backend/internal/store/store_internal_test.go:200:func TestBuildSearch(t *testing.T) {
-> count 1 / expected 1 : OK

L20 $ grep -nE '^func TestSearchFilters\(' backend/internal/store/store_test.go   # expected 1
backend/internal/store/store_test.go:170:func TestSearchFilters(t *testing.T) {
-> count 1 / expected 1 : OK

L21 $ grep -nE '^func TestLoadSample\(' backend/internal/store/seed/seed_test.go   # expected 1
backend/internal/store/seed/seed_test.go:53:func TestLoadSample(t *testing.T) {
-> count 1 / expected 1 : OK

L22 $ grep -nE '^func TestLoad\(' backend/internal/config/config_test.go   # expected 1
backend/internal/config/config_test.go:20:func TestLoad(t *testing.T) {
-> count 1 / expected 1 : OK

L23 $ grep -nE '^ +- name: (minPrice|pageSize)$|^    (ListingResult|ScoringMeta|AppliedFilters|FieldError):$' backend/openapi.yaml   # expected 6
backend/openapi.yaml:157:        - name: minPrice
backend/openapi.yaml:229:        - name: pageSize
backend/openapi.yaml:414:    ListingResult:
backend/openapi.yaml:468:    ScoringMeta:
backend/openapi.yaml:490:    AppliedFilters:
backend/openapi.yaml:561:    FieldError:
-> count 6 / expected 6 : OK

L24 $ grep -nE '^func (TestOpenAPISearchParamsMatchHandler|TestOpenAPIExamplesMatchSPEC|TestOpenAPIErrorCodeEnums)\(' backend/internal/httpapi/openapi_test.go   # expected 3
backend/internal/httpapi/openapi_test.go:141:func TestOpenAPISearchParamsMatchHandler(t *testing.T) {
backend/internal/httpapi/openapi_test.go:189:func TestOpenAPIExamplesMatchSPEC(t *testing.T) {
backend/internal/httpapi/openapi_test.go:256:func TestOpenAPIErrorCodeEnums(t *testing.T) {
-> count 3 / expected 3 : OK

L25 $ grep -nE "^export const (FIELDS|COPY) = |^type FieldKind = |name: 'status', label: 'Status'|^export function (draftsFrom|validate|toSearchParams|filterKey)\(" frontend/src/components/FilterBar/filterBarModel.ts   # expected 8
frontend/src/components/FilterBar/filterBarModel.ts:10:type FieldKind = 'money' | 'count' | 'city' | 'text' | 'status';
frontend/src/components/FilterBar/filterBarModel.ts:21:export const FIELDS = [
frontend/src/components/FilterBar/filterBarModel.ts:28:  { name: 'status', label: 'Status', placeholder: 'Any', kind: 'status' },
frontend/src/components/FilterBar/filterBarModel.ts:38:export const COPY = {
frontend/src/components/FilterBar/filterBarModel.ts:118:export function draftsFrom(value: SearchParams): Drafts {
frontend/src/components/FilterBar/filterBarModel.ts:137:export function filterKey(value: SearchParams): string {
frontend/src/components/FilterBar/filterBarModel.ts:221:export function validate(drafts: Drafts): ClientMarks {
frontend/src/components/FilterBar/filterBarModel.ts:254:export function toSearchParams(drafts: Drafts): SearchParams {
-> count 8 / expected 8 : OK

L26 $ grep -nE "^export function FilterBar\(|kind === 'status'" frontend/src/components/FilterBar/FilterBar.tsx   # expected 2
frontend/src/components/FilterBar/FilterBar.tsx:79:export function FilterBar({
frontend/src/components/FilterBar/FilterBar.tsx:239:        } else if (kind === 'status') {
-> count 2 / expected 2 : OK

L27 $ grep -nE '^export (function StatusLine\(|interface StatusLineProps)|^function describeState\(|role="status"' frontend/src/components/StatusLine/StatusLine.tsx   # expected 4
frontend/src/components/StatusLine/StatusLine.tsx:6:export interface StatusLineProps {
frontend/src/components/StatusLine/StatusLine.tsx:21:function describeState(state: SearchState): [string, string] {
frontend/src/components/StatusLine/StatusLine.tsx:48:export function StatusLine({ state }: StatusLineProps) {
frontend/src/components/StatusLine/StatusLine.tsx:51:    <div role="status" className={styles.status}>
-> count 4 / expected 4 : OK

L28 $ grep -nE '^export function (formatListingMeta|formatSource|formatOrdering|formatResultSummary)\(' frontend/src/format/listing.ts   # expected 4
frontend/src/format/listing.ts:28:export function formatSource(source: string): string {
frontend/src/format/listing.ts:45:export function formatListingMeta(item: ListingResult, currentYear?: number): string[] {
frontend/src/format/listing.ts:78:export function formatOrdering(targetBudget: number | null): string {
frontend/src/format/listing.ts:84:export function formatResultSummary(data: SearchResponse): string {
-> count 4 / expected 4 : OK

L29 $ grep -nE '^export function ListingCard\(' frontend/src/components/ListingCard/ListingCard.tsx   # expected 1
frontend/src/components/ListingCard/ListingCard.tsx:22:export function ListingCard({ item, keywordTokens }: ListingCardProps) {
-> count 1 / expected 1 : OK

L30 $ grep -nE '^export function SearchPage\(|<StatusLine |<FilterBar' frontend/src/pages/SearchPage.tsx   # expected 3
frontend/src/pages/SearchPage.tsx:52:export function SearchPage() {
frontend/src/pages/SearchPage.tsx:99:        <FilterBar
frontend/src/pages/SearchPage.tsx:109:      <StatusLine state={holdError ? withoutPrevious(state) : state} />
-> count 3 / expected 3 : OK

L31 $ grep -nE '^export function (useSearchParamsState|parseSearchParams)\(|const (setSearch|clear) = useCallback' frontend/src/hooks/useSearchParamsState.ts   # expected 4
frontend/src/hooks/useSearchParamsState.ts:62:export function parseSearchParams(search: string): SearchParams {
frontend/src/hooks/useSearchParamsState.ts:102:export function useSearchParamsState(): SearchParamsState {
frontend/src/hooks/useSearchParamsState.ts:133:  const setSearch = useCallback(
frontend/src/hooks/useSearchParamsState.ts:162:  const clear = useCallback(() => {
-> count 4 / expected 4 : OK

L32 $ grep -nE '^export interface (SearchParams|ListingResult|ScoringMeta) ' frontend/src/api/types.ts   # expected 3
frontend/src/api/types.ts:6:export interface SearchParams {
frontend/src/api/types.ts:18:export interface ListingResult {
frontend/src/api/types.ts:45:export interface ScoringMeta {
-> count 3 / expected 3 : OK

L33 $ grep -nE '^export function buildSearchQuery\(' frontend/src/api/client.ts   # expected 1
frontend/src/api/client.ts:45:export function buildSearchQuery(params: SearchParams): string {
-> count 1 / expected 1 : OK

L34 $ grep -nE '^export function Pager\(|Array\.from\(\{ length: totalPages \}' frontend/src/components/Pager/Pager.tsx   # expected 2
frontend/src/components/Pager/Pager.tsx:18:export function Pager({ page, totalPages, onChange }: PagerProps) {
frontend/src/components/Pager/Pager.tsx:23:  const pages = Array.from({ length: totalPages }, (_, i) => i + 1);
-> count 2 / expected 2 : OK

L35 $ grep -nE '^export const (listingA1|listingB7|searchResponse): ' frontend/src/test/fixtures.ts   # expected 3
frontend/src/test/fixtures.ts:10:export const listingA1: ListingResult = {
frontend/src/test/fixtures.ts:75:export const searchResponse: SearchResponse = {
frontend/src/test/fixtures.ts:187:export const listingB7: ListingResult = {
-> count 3 / expected 3 : OK

L36 $ grep -nE '^func \(s \*Store\) Upsert\(' backend/internal/store/listings.go   # expected 1 (Review Cycle 1)
backend/internal/store/listings.go:57:func (s *Store) Upsert(ctx context.Context, listings []domain.Listing) (int, error) {
-> count 1 / expected 1 : OK

L37 $ grep -nE '^[[:space:]]+Statuses \[\]ListingStatus' backend/internal/domain/filter.go   # expected 1 (Review Cycle 1)
backend/internal/domain/filter.go:24:	Statuses []ListingStatus
-> count 1 / expected 1 : OK

LINES 37  MISMATCHES 0

LS $ ls frontend/src/{components/FilterBar/{filterBarModel.test.ts,FilterBar.test.tsx},components/StatusLine/StatusLine.test.tsx,components/ListingCard/ListingCard.test.tsx,format/listing.test.ts,hooks/useSearchParamsState.test.tsx,api/client.test.ts,pages/SearchPage.test.tsx,components/Pager/Pager.test.tsx}   # expected 9
frontend/src/api/client.test.ts
frontend/src/components/FilterBar/FilterBar.test.tsx
frontend/src/components/FilterBar/filterBarModel.test.ts
frontend/src/components/ListingCard/ListingCard.test.tsx
frontend/src/components/Pager/Pager.test.tsx
frontend/src/components/StatusLine/StatusLine.test.tsx
frontend/src/format/listing.test.ts
frontend/src/hooks/useSearchParamsState.test.tsx
frontend/src/pages/SearchPage.test.tsx
-> 9 / expected 9 : OK
```

## Step 1b. Same patterns against master tip (c8837b5), counts only

Each line ran as `git grep -nE '<pat>' master -- <path>` with the identical pattern (L12 and L37 use `[[:space:]]`, no `\s` anywhere). The file list ran as `git ls-tree --name-only master -- <the 9 paths>`.

| Line | Path | master count / expected | Status |
|---|---|---|---|
| L01 | backend/internal/httpapi/search.go | 9 / 9 | OK |
| L02 | backend/internal/httpapi/query.go | 11 / 11 | OK |
| L03 | backend/internal/httpapi/query.go | 9 / 9 | OK |
| L04 | backend/internal/domain/rank.go | 4 / 4 | OK |
| L05 | backend/internal/domain/scoring.go | 6 / 6 | OK |
| L06 | backend/internal/domain/paginate.go | 2 / 2 | OK |
| L07 | backend/internal/domain/filter.go | 1 / 1 | OK |
| L08 | backend/internal/domain/listing.go | 3 / 3 | OK |
| L09 | backend/internal/store/listings.go | 6 / 6 | OK |
| L10 | backend/internal/store/migrations/0001_listings.sql | 4 / 4 | OK |
| L11 | backend/internal/store/seed/seed.go | 3 / 3 | OK |
| L12 | backend/internal/config/config.go | 3 / 3 | OK |
| L13 | backend/cmd/api/main.go | 1 / 1 | OK |
| L14 | compose.yaml | 1 / 1 | OK |
| L15 | backend/internal/httpapi/query_test.go | 3 / 3 | OK |
| L16 | backend/internal/httpapi/search_test.go | 5 / 5 | OK |
| L17 | backend/internal/domain/rank_test.go | 2 / 2 | OK |
| L18 | backend/internal/domain/scoring_test.go | 1 / 1 | OK |
| L19 | backend/internal/store/store_internal_test.go | 1 / 1 | OK |
| L20 | backend/internal/store/store_test.go | 1 / 1 | OK |
| L21 | backend/internal/store/seed/seed_test.go | 1 / 1 | OK |
| L22 | backend/internal/config/config_test.go | 1 / 1 | OK |
| L23 | backend/openapi.yaml | 6 / 6 | OK |
| L24 | backend/internal/httpapi/openapi_test.go | 3 / 3 | OK |
| L25 | frontend/src/components/FilterBar/filterBarModel.ts | 8 / 8 | OK |
| L26 | frontend/src/components/FilterBar/FilterBar.tsx | 2 / 2 | OK |
| L27 | frontend/src/components/StatusLine/StatusLine.tsx | 4 / 4 | OK |
| L28 | frontend/src/format/listing.ts | 4 / 4 | OK |
| L29 | frontend/src/components/ListingCard/ListingCard.tsx | 1 / 1 | OK |
| L30 | frontend/src/pages/SearchPage.tsx | 3 / 3 | OK |
| L31 | frontend/src/hooks/useSearchParamsState.ts | 4 / 4 | OK |
| L32 | frontend/src/api/types.ts | 3 / 3 | OK |
| L33 | frontend/src/api/client.ts | 1 / 1 | OK |
| L34 | frontend/src/components/Pager/Pager.tsx | 2 / 2 | OK |
| L35 | frontend/src/test/fixtures.ts | 3 / 3 | OK |
| L36 | backend/internal/store/listings.go | 1 / 1 | OK |
| L37 | backend/internal/domain/filter.go | 1 / 1 | OK |
| LS | 9 frontend test files (git ls-tree) | 9 / 9 | OK |

Result line: `LINES 37  MISMATCHES 0`. No mismatch to report.

Path tokens checked as objects on both trees (`git cat-file -t HEAD:<p>` and `master:<p>`): all `blob` for backend/internal/store/migrations/0001_listings.sql, backend/cmd/api/main.go, backend/internal/domain/listing.go, backend/internal/httpapi/query.go, backend/internal/httpapi/search.go, backend/internal/httpapi/search_test.go, backend/internal/store/listings.go, backend/openapi.yaml, compose.yaml, frontend/src/api/types.ts, frontend/src/components/FilterBar/filterBarModel.ts, frontend/src/test/fixtures.ts; `tree` for backend/internal/domain.

## Step 1c. Master drift since the merge-base

```
$ git diff --name-only 0b24829 master
.lattice/artifacts/meta/art_01M4WCKXJV38BKNP5MAJCS9MBY.json
.lattice/artifacts/meta/art_01M4WD65X8Q5CPW0C2KATK709C.json
.lattice/artifacts/meta/art_01M4WDHT3EZ10MG39D70KKZZFZ.json
.lattice/artifacts/meta/art_01M4WDQP79JFKTS1PVSDWKGBMY.json
.lattice/artifacts/meta/art_01M4WDRNDWH1M3J2MZYAKY4Q5R.json
.lattice/artifacts/meta/art_01M4WDX8JZJ51XARPZCPHZE0DC.json
.lattice/artifacts/meta/art_01M4WE5VJ2SFN6E8PBJ7TG0JEZ.json
.lattice/artifacts/payload/art_01M4WCKXJV38BKNP5MAJCS9MBY.md
.lattice/artifacts/payload/art_01M4WD65X8Q5CPW0C2KATK709C.md
.lattice/artifacts/payload/art_01M4WDHT3EZ10MG39D70KKZZFZ.md
.lattice/artifacts/payload/art_01M4WDQP79JFKTS1PVSDWKGBMY.md
.lattice/artifacts/payload/art_01M4WDRNDWH1M3J2MZYAKY4Q5R.md
.lattice/artifacts/payload/art_01M4WDX8JZJ51XARPZCPHZE0DC.md
.lattice/artifacts/payload/art_01M4WE5VJ2SFN6E8PBJ7TG0JEZ.md
.lattice/events/task_01M4VXK3K5KDHX0P4DPET2S05V.jsonl
.lattice/events/task_01M4VXK3QPSMGSSGYBJ82HRQ4Z.jsonl
.lattice/events/task_01M4VXK3WPPJA1GVM0NZRW4GGD.jsonl
.lattice/events/task_01M4VXK4113A6VE5JQHFW73ERX.jsonl
.lattice/orchestration/agents.md
.lattice/orchestration/delivery-receipts.jsonl
.lattice/orchestration/evidence/BC-15-validation.md
.lattice/orchestration/master-validator-log.md
.lattice/orchestration/prompts/BC-17.md
.lattice/orchestration/run-state.md
.lattice/orchestration/validation-plan.md
.lattice/plans/task_01M4VXK3K5KDHX0P4DPET2S05V.md
.lattice/plans/task_01M4VXK3WPPJA1GVM0NZRW4GGD.md
.lattice/plans/task_01M4VXK4113A6VE5JQHFW73ERX.md
.lattice/tasks/task_01M4VXK3K5KDHX0P4DPET2S05V.json
.lattice/tasks/task_01M4VXK3QPSMGSSGYBJ82HRQ4Z.json
.lattice/tasks/task_01M4VXK3WPPJA1GVM0NZRW4GGD.json
.lattice/tasks/task_01M4VXK4113A6VE5JQHFW73ERX.json
README.md

$ grep -nE 'README|\.lattice' docs/EXTENSIONS.md
(no output)
```

No file the doc names changed between 0b24829 and master. The doc names no README.md or .lattice path.

## Step 2. Backtick-token reconciliation

```
$ grep -oE '`[^`]+`' docs/EXTENSIONS.md | sort -u | wc -l
157
```

Classes: (a) existing name or path, cited by the Step 1a line (Lxx file:line) or a direct hit; (b) a proposal, cited by the doc line where "new"/"New" introduces it; (c) SPEC contract name, param, response field, error code, or BRANDING/UI copy string, cited by the SPEC (docs/SPEC.md) or BRANDING (BRANDING.md) line. Result: 104 (a) + 31 (b) + 22 (c) = 157; 0 unclassified.

### (a) Existing names and paths (104)

| # | Token | Evidence |
|---|---|---|
| 1 | `0001_listings.sql` | L10, backend/internal/store/migrations/0001_listings.sql (blob on HEAD and master) |
| 4 | `AllStatuses` | L08 listing.go:37 |
| 5 | `appliedFilters` | L01 search.go:114; also SPEC:211 response field |
| 6 | `AppliedFilters` | L23 openapi.yaml:490 |
| 8 | `assertItemsMatchSeed` | L16 search_test.go:293 |
| 9 | `backend/cmd/api/main.go` | path blob; L13 |
| 12 | `backend/internal/domain/listing.go` | path blob; L08 |
| 13 | `backend/internal/domain` | tree on HEAD and master |
| 14 | `backend/internal/httpapi/query.go` | path blob; L02 |
| 15 | `backend/internal/httpapi/search.go` | path blob; L01 |
| 16 | `backend/internal/store/listings.go` | path blob; L09 |
| 17 | `backend/openapi.yaml` | path blob; L23 |
| 19 | `BudgetWeight` | L05 scoring.go:28 |
| 20 | `buildSearch` | L09 listings.go:116 |
| 21 | `buildSearchQuery` | L33 client.ts:45 |
| 22 | `byRelevance` | L04 rank.go:74 |
| 23 | `clear` | L31 useSearchParamsState.ts:162 |
| 24 | `compose.yaml` | path blob; L14 |
| 25 | `Dedupe slot (section 9.1)` | L01 search.go:37 |
| 29 | `DisallowUnknownFields` | L11 seed.go:123 |
| 31 | `draftsFrom` | L25 filterBarModel.ts:118 |
| 32 | `FieldError` | L23 openapi.yaml:561; SPEC:410 |
| 33 | `FieldKind` | L25 filterBarModel.ts:10 (doc line 68 proposes a new value of this existing type) |
| 34 | `FIELDS` | L25 filterBarModel.ts:21 |
| 35 | `FilterBar.test.tsx` | LS |
| 36 | `FilterBar` | L26 FilterBar.tsx:79 |
| 37 | `filterBarModel.test.ts` | LS |
| 38 | `filterKey` | L25 filterBarModel.ts:137 |
| 39 | `formatListingMeta` | L28 listing.ts:45 |
| 40 | `formatOrdering` | L28 listing.ts:78 |
| 41 | `formatSource` | L28 listing.ts:28 |
| 42 | `frontend/src/api/types.ts` | path blob; L32 |
| 43 | `frontend/src/components/FilterBar/filterBarModel.ts` | path blob; L25 |
| 44 | `frontend/src/test/fixtures.ts` | path blob; L35 |
| 49 | `kind === 'status'` | L26 FilterBar.tsx:239 |
| 52 | `listing.test.ts` | LS |
| 53 | `Listing` | L08 listing.go:61 |
| 54 | `ListingCard.test.tsx` | LS |
| 55 | `ListingResult` | L23 openapi.yaml:414; L32 types.ts:18 |
| 56 | `LISTINGS_FILE` | L14 compose.yaml:37; SPEC:128 |
| 57 | `ListingsFile` | L12 config.go:52 |
| 59 | `Load` | L11 seed.go:81 (doc line 105 means the seed loader) |
| 61 | `MaxPageSize` | L06 paginate.go:10 |
| 64 | `minBedrooms()` | L02 query.go:250 |
| 70 | `newAppliedFilters` | L01 search.go:188 |
| 72 | `newScoringMeta` | L01 search.go:171 |
| 73 | `newSearchItem` | L01 search.go:141 |
| 74 | `nonNegative` | L02 query.go:240 |
| 78 | `Pager.test.tsx` | LS |
| 79 | `Pager` | L34 Pager.tsx:18 |
| 80 | `pageSize()` | L02 query.go:352 |
| 82 | `Paginate` | L06 paginate.go:56 |
| 83 | `parseOrigins` | L12 config.go:139 |
| 84 | `parseSearchParams` | L31 useSearchParamsState.ts:62 |
| 85 | `ParseSearchQuery` | L02 query.go:148 |
| 87 | `ParseStatus` | L08 listing.go:44 |
| 88 | `permutations` | L17 rank_test.go:72 |
| 89 | `query.go` | L02 (backend/internal/httpapi/query.go) |
| 92 | `Rank` | L04 rank.go:56 |
| 93 | `RankInput` | L04 rank.go:10 |
| 94 | `RecencyWeight` | L05 scoring.go:32 |
| 96 | `role="status"` | L27 StatusLine.tsx:51 |
| 99 | `scoringMeta` | L01 search.go:105 |
| 100 | `ScoringMeta` | L23 openapi.yaml:468; L32 types.ts:45 |
| 101 | `search_test.go` | L16 (backend/internal/httpapi/search_test.go, blob) |
| 102 | `search.go` | L01 (backend/internal/httpapi/search.go) |
| 103 | `Searcher` | L01 search.go:15 |
| 104 | `SearchFilter.Statuses` | L07 filter.go:6 + L37 filter.go:24 |
| 105 | `SearchFilter` | L07 filter.go:6 |
| 106 | `searchItem` | L01 search.go:68 |
| 107 | `searchListings` | L01 search.go:22 |
| 108 | `SearchPage.test.tsx` | LS |
| 109 | `searchParams` | L02 query.go:33 |
| 110 | `SearchParams` | L32 types.ts:6 |
| 111 | `seed.Load(cfg.ListingsFile)` | L13 main.go:56 |
| 113 | `setSearch` | L31 useSearchParamsState.ts:133 |
| 120 | `SortStableFunc(out, byRelevance)` | L04 rank.go:67 |
| 124 | `statuses()` | L02 query.go:306 |
| 125 | `StatusLine.test.tsx` | LS |
| 126 | `StatusLine` | L27 StatusLine.tsx:48 |
| 127 | `Store.Search` | L09 listings.go:92 `func (s *Store) Search(` |
| 129 | `TestBuildSearch` | L19 store_internal_test.go:200 |
| 130 | `TestLoad` | L22 config_test.go:20 |
| 131 | `TestLoadSample` | L21 seed_test.go:53 |
| 132 | `TestOpenAPIErrorCodeEnums` | L24 openapi_test.go:256 |
| 133 | `TestOpenAPIExamplesMatchSPEC` | L24 openapi_test.go:189 |
| 134 | `TestOpenAPISearchParamsMatchHandler` | L24 openapi_test.go:141 |
| 135 | `TestParseSearchQueryRules` | L15 query_test.go:83 |
| 136 | `TestParseSearchQueryStatus` | L15 query_test.go:340 |
| 137 | `TestRank` | L17 rank_test.go:93 |
| 138 | `TestScore` | L18 scoring_test.go:169 |
| 139 | `TestSearchFilters` | L20 store_test.go:170 |
| 140 | `TestSearchPaging` | L16 search_test.go:541 |
| 141 | `TestSearchQueryParams` | L15 query_test.go:442 |
| 142 | `TestSearchRankedAll` | L16 search_test.go:444 |
| 143 | `TestSearchStatusFilter` | L16 search_test.go:601 |
| 145 | `toSearchParams` | L25 filterBarModel.ts:254 |
| 148 | `Upsert` | L36 listings.go:57 |
| 149 | `useSearchParamsState.test.tsx` | LS |
| 150 | `useSearchParamsState` | L31 useSearchParamsState.ts:102 |
| 151 | `validate` | L25 filterBarModel.ts:221 |
| 152 | `whereMinBedrooms` | L09 listings.go:32 |
| 153 | `whereStatusIn` | L09 listings.go:36 |
| 154 | `wireItem` | L16 search_test.go:56 |

### (b) Proposals introduced by "new" (31)

| # | Token | Doc line(s) where "new"/"New" introduces it |
|---|---|---|
| 2 | `0002_*.sql` | 130 "A new `0002_*.sql` migration" |
| 10 | `backend/internal/domain/dedupe.go` | 38 "in new `backend/internal/domain/dedupe.go`" |
| 11 | `backend/internal/domain/geo.go` | 65 "in new `backend/internal/domain/geo.go`" |
| 18 | `BoundingBox` | 65 "New `BoundingBox`"; 166 |
| 26 | `dedupe_test.go` | 44 "New `dedupe_test.go`" |
| 27 | `Dedupe` | 38 "New `Dedupe`" |
| 28 | `dedupe=false` | 159 "a new `dedupe=false` toggle" |
| 30 | `distanceKm` | 69 "A new `distanceKm`"; 174 |
| 45 | `geo_test.go` | 70 "New `geo_test.go`"; 177 |
| 46 | `HaversineKm` | 65 "new `HaversineKm`"; 169 |
| 50 | `lat` | 64 "Three new params: `lat`, `lng` and `radiusKm`"; 163; also SPEC:512 (§9.3) |
| 58 | `lng` | 64; 163; also SPEC:512 |
| 60 | `lowest price first` | 194 "gains the new string `lowest price first`"; 81-82 "gains a new / string," |
| 63 | `minBathrooms` | 89 "New `minBathrooms`, `minSqft` and `zip` filters"; also SPEC:514 (§9.5) |
| 65 | `minSqft` | 89; also SPEC:514 |
| 66 | `missing_param` | 67 "One new error code, `missing_param`"; 164 |
| 69 | `MLS_C` | 109 "a new `MLS_C` source" |
| 86 | `ParseSort` | 77 "new `ParseSort`"; 185 |
| 90 | `radiusKm` | 64; 163; also SPEC:512 |
| 97 | `scoring.sort` | 194 "as new `scoring.sort`"; 79-80 "the new / `scoring.sort` echo" |
| 112 | `setPageSize` | 120 "New `setPageSize`" |
| 114 | `setSort` | 198 "A new `setSort`" |
| 115 | `sort` | 198 "the new `sort` param" |
| 116 | `Sort` | 77 "`RankInput` gains new `Sort`"; 186 |
| 117 | `sort=distance` | 174 "A new `sort=distance`" |
| 118 | `sort=relevance\|price\|newest` | 79 "New `sort=relevance\|price\|newest`"; 184; also SPEC:513 (§9.4) |
| 119 | `SortOrder` | 77 "New `SortOrder`"; 185 |
| 121 | `Sources []string` | 39 "`Listing` gains new `Sources []string`" |
| 122 | `sources` | 40 "add a new `sources` field"; 152; also SPEC:510 (§9.1 `sources: string[]`) |
| 155 | `WithinRadius` | 66 "A new `WithinRadius` stage"; 169 |
| 156 | `X` | 4, the convention's own placeholder: 'A name written as "new `X`" does not exist yet' |

Proposals are absent on master, so none is mislabelled as new (`git grep -nE '<s>' master -- backend frontend/src | wc -l`):
`func Dedupe` 0, `BoundingBox` 0, `HaversineKm` 0, `WithinRadius` 0, `SortOrder` 0, `ParseSort` 0, `setPageSize` 0, `setSort` 0, `missing_param` 0, `distanceKm` 0, `Sources \[\]string` 0, `lowest price first` 0, `dedupe=false` 0. `git cat-file -t master:<p>` reports "does not exist in 'master'" for backend/internal/domain/{dedupe.go,geo.go,dedupe_test.go,geo_test.go}.

### (c) SPEC contract names and BRANDING/UI copy strings (22)

| # | Token | SPEC / BRANDING line |
|---|---|---|
| 3 | `2 sources` | BRANDING:341 (§10) "the card gains a `2 sources` item in the meta line"; SPEC:510 (§9.1) |
| 7 | `asOf` | SPEC:209 response field `"asOf": "2026-10-13"` |
| 47 | `invalid_enum` | SPEC:250-251 (§5.5 field error codes) |
| 48 | `keyword` | SPEC:168 (§5.3 param) |
| 51 | `listedDate` | SPEC:110 field |
| 62 | `maxPrice=451000` | SPEC:165 (§5.3 param `maxPrice`) with a sample value |
| 67 | `MLS A · 2 sources` | BRANDING:230 (§5.4 meta line, items separated by ` · `, ends `MLS A`) + BRANDING:341 (`2 sources` item) |
| 68 | `MLS C` | BRANDING:291 (§7.1 Source: underscore to space, `MLS A`), applied to the new `MLS_C`; listing.ts:27-29 `source.replaceAll('_', ' ')` |
| 71 | `newest first` | BRANDING:220 (§5.3 status line copy); listing.ts:79 |
| 75 | `not_positive` | SPEC:250 |
| 76 | `out_of_range` | SPEC:251 |
| 77 | `page` | SPEC:171 (§5.3 param) |
| 81 | `pageSize` | SPEC:172 (§5.3 param) |
| 91 | `rank` | SPEC:400 `interface ListingResult { rank: number; ...` |
| 95 | `relevance` | SPEC:404 `relevance: { score ... }` (doc line 45); SPEC:513 sort value (doc line 184) |
| 98 | `scoring` | SPEC:205 `"scoring": {` response block |
| 123 | `status` | SPEC:170 (§5.3 param) |
| 128 | `targetBudget=450000` | SPEC:169 (§5.3 param `targetBudget`) with the SPEC §5.4 sample value |
| 144 | `too_long` | SPEC:251 |
| 146 | `totalItems` | SPEC:204 |
| 147 | `totalPages` | SPEC:204 |
| 157 | `zip` | SPEC:104 listing field; listing.go:73 `Zip string json:"zip"`; as a new param, doc line 89 and SPEC:514 |

Unclassified tokens: none.

### Spot checks of claims attached to symbols (not gating)

- rank.go:70-72: "byRelevance ... It is the one place that ordering is defined; a new sort option would sit beside it." Matches doc line 187-188.
- filterBarModel.ts:137-142: `filterKey` deletes `page` and `pageSize` only. Matches doc line 199 ("must drop it beside `page` and `pageSize`").
- useSearchParamsState.ts: `setSearch` copies the current `pageSize`; `clear` navigates to `{ pageSize }` when set. Matches doc lines 120-121.
- SPEC:363-364 (§7.1): "`pageSize` is never written by the UI but is preserved if present." Matches the quote at doc line 121.
- backend/go.mod:5 `require modernc.org/sqlite v1.60.1`. Matches doc line 130.
- listing.ts:78-80 `formatOrdering` returns `'newest first'` without a budget. Matches doc line 195.

## Step 3. Structure

```
$ wc -l < docs/EXTENSIONS.md
     221
$ test $(wc -l < docs/EXTENSIONS.md) -lt 300 && echo under-300
under-300
$ grep -cE '^### [1-8]\. ' docs/EXTENSIONS.md
8
$ awk '/^#{1,3} /{if(h)print n, h; h=""} /^### [1-8]\. /{h=$0; n=0} h{n++} END{if(h)print n, h}' docs/EXTENSIONS.md
14 ### 1. Cross-source dedupe
13 ### 2. Status filter
12 ### 3. Geo radius
13 ### 4. Sort options
13 ### 5. More filters
13 ### 6. A new feed
12 ### 7. Page size control
13 ### 8. Scale
$ grep -nE '^## |^### ' docs/EXTENSIONS.md
8:## Where things go
33:## Extension points
35:### 1. Cross-source dedupe
49:### 2. Status filter
62:### 3. Geo radius
74:### 4. Sort options
87:### 5. More filters
100:### 6. A new feed
113:### 7. Page size control
125:### 8. Scale
138:## Rehearsal
140:### Cross-source dedupe
161:### Geo radius
182:### Sort options
203:## What I would say
205:### Scoring weights
214:### SQL versus Go ranking
```

Same awk over every `### ` subsection (Rehearsal and What I would say rows): Cross-source dedupe 21, Geo radius 21, Sort options 21, Scoring weights 9, SQL versus Go ranking 8.

SPEC §9 items (docs/SPEC.md:510-517): 1 Cross-source dedupe, 2 Status filter, 3 Geo radius, 4 Sort options, 5 More filters, 6 A new feed, 7 Page size control, 8 Scale. The doc's 8 headings match them one to one, in order.

## Step 4. No emoji, voice

```
$ perl -CSD -ne 'print "$.: $_" if /[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{2B00}-\x{2BFF}\x{FE0F}\x{200D}\x{20E3}]/' docs/EXTENSIONS.md
(no output; perl exit 0)
$ grep -nE '!|[Pp]lease|[Oo]ops|[Ss]orry|[Ss]imply|[Jj]ust ' docs/EXTENSIONS.md
(no output; grep exit 1)
```

No voice hits, so nothing to classify as code span or prose.

## Step 5. Scope

```
$ git diff --name-only master...HEAD
docs/EXTENSIONS.md
$ git status --porcelain
(no output)
```

Overall: PASS
