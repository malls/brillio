# BC-8 plan: frontend core (API types and client, URL state and search hooks, BRANDING formatters)

Planner: agent:opus-5.5-planner, 2026-10-14. Worktree `/Users/forrest/Code/brillio-worktrees/bc-8-frontend-core`, branch
`feat/BC-8-frontend-core`, anchored on `feat/BC-7-frontend-scaffold` @1d6008a (BC-7 in review). Paths are relative to
`frontend/` unless they start with `docs/`. Toolchain is BC-7's: TS 6.0.3 (`strict`, `exactOptionalPropertyTypes`,
`noUncheckedIndexedAccess`, `verbatimModuleSyntax`, `erasableSyntaxOnly`); ESLint 9 with `recommendedTypeChecked` and
react-hooks 7 `recommended`, at `--max-warnings 0`; Prettier; Vitest 5 on jsdom with `restoreMocks` and `unstubGlobals`.
Card notes 1–6 are binding and are folded in below.

## Scope

In: `src/api/{types,errors,client}.ts`, `src/hooks/{useSearchParamsState,useListingSearch,useCities}.ts`,
`src/format/{money,date,listing}.ts`, `src/test/fixtures.ts`, co-located tests, and one sentence added to `docs/SPEC.md`
§7.2. Out: all components, pages and CSS (BC-9..12); `App.tsx`, `main.tsx`, `styles/*`, `test/setup.ts`; `package.json`
and the lockfile (no new dependencies, and no msw: `fetch` is stubbed with `vi.fn`); eslint, vite and tsconfig files;
`Makefile`; FilterBar input parsing such as `$400,000` back to a number (BC-9 may add it to `format/money.ts`); and the
mapping from SearchState to StatusLine text (BC-11; format supplies the data-level halves).

## Decisions

1. **One serializer.** `buildSearchQuery(params)` in client.ts is the only code that builds a query string. It produces
   the API URL, the `pushState` URL (§7.1 uses §5.3's param names, so URL and API call match) and the equality key in
   `useListingSearch`. Output is canonical: §5.3 table order (minPrice, maxPrice, minBedrooms, city, keyword,
   targetBudget, status, page, pageSize), encoded with `URLSearchParams`.
2. **Omissions.** The serializer omits `undefined`; `''` (§5.3 treats an empty value as absent); non-finite numbers
   (`String(NaN)` is `"NaN"`, which Go's `ParseFloat` accepts); and an empty `status` array (absent means all three).
   Numbers are written as `String(n)`. `status` is repeated once per value, deduped and in enum order, so
   `['sold','active']` and `['active','sold']` give the same URL and key.
3. **Rejections are `ApiRequestError extends Error`, carrying the ApiError.** typescript-eslint `only-throw-error` and
   `prefer-promise-reject-errors` (both errors in recommendedTypeChecked) forbid throwing or rejecting the plain object.
   It holds `readonly apiError: ApiError` and `readonly requestId: string | null` (from `X-Request-ID`), with the
   original error as `cause`. The §7.3 union is unchanged; SPEC does not say what the rejection value is. Only the hooks
   catch client rejections.
4. **Abort vs network.** The client maps every `fetch` rejection to `{ kind: 'network' }`. The message is
   `Request aborted.` when `err.name === 'AbortError'` (checked by name: jsdom and Node each have their own DOMException,
   so `instanceof` is unreliable), and the TypeError's message otherwise. The hooks never classify errors: each `.then`
   handler returns early when its own `controller.signal.aborted` is set, which also covers an abort mid-body-read.
5. **Loading is derived, not set.** react-hooks 7 makes `set-state-in-effect` an error. A planner probe with BC-7's
   plugin confirmed that `setLoading(true)` in an effect body fails and that this pattern passes. State is
   `{ settled: { key, outcome } | null, lastGood: SearchResponse | null }`, and setState runs only in promise callbacks.
   With `requestKey = \`${attempt}|${query}\``, the status is `loading` while `settled?.key !== requestKey`. No refs:
   the `refs` rule forbids reading them during render.
6. **Params identity.** The effect deps are `[query, requestKey]` (strings), so an equal-but-new params object never
   refetches. The effect calls `searchListingsByQuery(query, signal)` and never closes over `params`, so
   exhaustive-deps needs no disable. `searchListings(params, signal)` is that call wrapped in `buildSearchQuery`.
7. **Retry.** The hook returns `{ state, retry }`. `retry()` increments `attempt`, re-running the current query
   ("Try again", BRANDING §5.9). It is the only way to refetch unchanged params.
8. **`previous`** is the latest 2xx response, success or empty. It is attached to `loading`, `invalid` and `error`, and
   absent before the first 2xx. The status is `empty` iff `page.totalItems === 0`, otherwise `success`, including a
   page past the end (§5.3: a 200 with no items and real metadata; the Pager recovers). The hook fetches on mount, so it
   starts in `loading` and never emits `idle`; `idle` stays in the type for components.
9. **Logging (§7.5).** `useListingSearch` logs `http` and `network` failures once, and never validation errors or
   aborts: `console.error('Listing search failed', { ...apiError, requestId })`. SearchPage must not log them again.
   `useCities` logs `console.error('Failed to load cities', err)`.
10. **URL parsing drops what the UI cannot show.** Dropped: type failures (non-numeric or non-finite numbers,
    non-integer `minBedrooms`, `page` or `pageSize`, unknown `status` values, empty values), plus `page < 1` and a
    `pageSize` outside 1..100, because no field could show the server's error. Range errors on visible fields are kept
    (`minPrice=-5`, min > max, `targetBudget=0`, over-long text): the server's 400 marks those fields, and the API
    stays the authority (BRANDING §5.9). Unknown params are ignored and vanish on the next push. For a repeated
    scalar, the first value wins. `status` may be repeated or comma-separated (§5.3).
11. **Writes.** `setSearch(next)` drops `next.page` and keeps `next.pageSize ?? currentUrl.pageSize`. `setPage(n)` keeps
    every current param and omits `page` when `n === 1` (`?city=X` is the same search as `?city=X&page=1`); a
    non-integer or `n < 1` does nothing. `clear()` writes an empty query but keeps `pageSize` if present (Ambiguity 1).
    Writers read the live `window.location.search`, so they are stable `useCallback(…, [])`. A write equal to the
    current search is skipped (no duplicate history entry). `pushState` fires no popstate, so afterwards the hook sets
    state to `parseSearchParams(newQuery)`. An empty query pushes `location.pathname + location.hash`, because a
    relative `''` URL would keep the old query.
12. **Formatters are pure, `en-US`-pinned and never throw.** One `Intl.NumberFormat('en-US', { minimumFractionDigits:
    0, maximumFractionDigits: 0 })` formats price, sqft and counts. Months come from a literal `Jan`..`Dec` table (some
    locales give `Sept`). Dates are parsed from `YYYY-MM-DD` with regex groups and no `Date` is built:
    `new Date('2026-08-29')` is UTC midnight and shows Aug 28 in US time zones (verified on this host). `currentYear`
    defaults to `new Date().getFullYear()` (local) so components can call it bare; tests always pass it. An unparseable
    date falls back to `Listed <raw>`, so bad data stays visible.
13. **Hook members are property functions** (`setSearch: (next: SearchParams) => void`), not method signatures, so
    destructuring them does not trip `@typescript-eslint/unbound-method`.

## Module contents

- **`api/types.ts`:** the §7.3 block verbatim with every declaration exported, `FieldError`, `ValidationError` and
  `ApiError` included. One addition, used by the parser, the serializer and BC-9's Status select:
  `export const LISTING_STATUSES = ['active', 'pending', 'sold'] as const satisfies readonly ListingStatus[]`.
- **`api/errors.ts`:** re-exports those three types, plus `ApiRequestError`, `parseErrorResponse` and `toApiError`.
  `parseErrorResponse(res)` reads `res.json()` into `unknown` (a parse failure or empty body counts as no body).
  `body.error.code === 'validation_error'` gives `{ kind: 'validation', message, fields }`, where `fields` keeps only
  entries with string `field`, `code` and `message` (`[]` if not an array). Anything else gives `{ kind: 'http', status,
  message }`. `message` is `body.error.message` if non-empty, else `Request failed with status ${status}.` (this covers
  the Vite proxy's HTML 502). `toApiError(err)` gives `err.apiError` for an `ApiRequestError`, else `network`.
- **`api/client.ts`:** a private `request(url, signal)` calls `fetch(url, { signal: signal ?? null, headers: { Accept:
  'application/json' } })`; `RequestInit.signal` is `AbortSignal | null`, so `undefined` fails under
  exactOptionalPropertyTypes. A `fetch` rejection becomes `ApiRequestError(network, null, { cause })`; a non-ok
  response becomes `ApiRequestError(await parseErrorResponse(res), requestId)`; an unparseable 2xx body is `network`
  if aborted, else `http` with `Malformed response body.`. Shape guards: a search body needs an `items` array and
  `page` and `scoring` objects, then is cast; a cities body needs a `cities` array, of which only strings are kept.
  URLs are `/api/listings/search` (with `?query` only when non-empty) and `/api/cities`.
- **`hooks/useSearchParamsState.ts`:** `parseSearchParams` implements decision 10. Initial state is
  `useState(() => parseSearchParams(window.location.search))`, and one effect re-parses on `popstate`. The writers
  (decision 11) go through a private `navigate`, and copies use spread then `delete copy.page`, never `page: undefined`.
- **`hooks/useListingSearch.ts`:** exports `SearchState` (§7.2 verbatim) for BC-11 and BC-12, and implements decisions
  5–9. The state is derived in `useMemo([settled, lastGood, requestKey])`, so its identity is stable across unrelated
  renders. `previous` is added by conditional spread, never as `previous: undefined`. The effect cleanup aborts on
  param change, retry, unmount and StrictMode remount.
- **`hooks/useCities.ts`:** one mount effect with its own AbortController; failure leaves `[]` and logs unless aborted.
  **`format/*`:** see the export list (`listing.ts` imports `formatPrice` and `formatListedDate`).
- **`test/fixtures.ts`:** real rows from `data/sample_listings.json`, with ranks and relevance taken from the planner's
  run of the full 12-listing §6.2–6.3 ranking (`targetBudget=450000`, asOf `2026-10-13`):

  | Export | Row | Rank | score / budgetFit / recency |
  |---|---|---|---|
  | `listingA1` | A1, Springfield | 1 | 0.7414 / 1 / 0.3536 |
  | `listingA3` | A3, Fairfax, `1 ba` | 4 | 0.6836 / 0.8867 / 0.3789 |
  | `listingA7` | A7, Chantilly, pending | 10 | 0.4275 / 0.6 / 0.1688 |

  `searchResponse` is page 1 of that real search (`page {1, 10, 12, 2}`, `scoring {450000, 0.6, 0.4, 30,
  '2026-10-13'}`, `appliedFilters` all null plus all three statuses), with `items` trimmed to the three rows (say so in a
  comment). It yields BRANDING's own `12 results · ranked for $450,000` and `Page 1 of 2`. `emptySearchResponse` is
  `city=Arlington` (`items []`, `page {1, 10, 0, 0}`, no-budget scoring `{null, 0, 1, 30, '2026-10-13'}`). Also:
  `cities` (§5.2's six), `validationErrorBody` (§5.5 verbatim), `internalErrorBody`, `jsonResponse(body, init?: {
  status?: number; requestId?: string }): Response` (JSON content type plus `X-Request-ID`), and `deferred<T>()` →
  `{ promise, resolve: (v: T) => void, reject: (e: Error) => void }` (the ES2023 lib lacks `Promise.withResolvers`).
  Every export is type-annotated, so `tsc -b` fails on any `types.ts` drift. Tests spread fixtures, never mutate them.

## Planned exports (republish this block in the completion comment)

```
src/api/types.ts       ListingStatus, SearchParams, ListingResult, PageMeta, ScoringMeta, SearchResponse,
                       FieldError, ValidationError, ApiError   (SPEC §7.3 verbatim)
                       LISTING_STATUSES: readonly ['active', 'pending', 'sold']
src/api/errors.ts      export type { ApiError, ValidationError, FieldError }
                       class ApiRequestError extends Error { readonly apiError: ApiError; readonly requestId: string | null }
                       parseErrorResponse(res: Response): Promise<ApiError>
                       toApiError(err: unknown): ApiError
src/api/client.ts      buildSearchQuery(params: SearchParams): string            // canonical, no leading '?'
                       searchListings(params: SearchParams, signal?: AbortSignal): Promise<SearchResponse>
                       searchListingsByQuery(query: string, signal?: AbortSignal): Promise<SearchResponse>
                       getCities(signal?: AbortSignal): Promise<string[]>       // reject with ApiRequestError
src/hooks/useSearchParamsState.ts
                       parseSearchParams(search: string): SearchParams
                       interface SearchParamsState { params: SearchParams; setSearch: (next: SearchParams) => void;
                         setPage: (page: number) => void; clear: () => void }
                       useSearchParamsState(): SearchParamsState
src/hooks/useListingSearch.ts
                       type SearchState                                         // SPEC §7.2 verbatim
                       interface ListingSearch { state: SearchState; retry: () => void }
                       useListingSearch(params: SearchParams): ListingSearch
src/hooks/useCities.ts useCities(): string[]                                    // [] until loaded and on failure
src/format/money.ts    formatPrice(amount: number): string                      // '$450,000'
src/format/date.ts     formatListedDate(isoDate: string, currentYear?: number): string // 'Listed Aug 29' | '…, 2025'
src/format/listing.ts  META_SEPARATOR = ' · '
                       formatBedrooms(n: number): string                        // '2 bd'
                       formatBathrooms(n: number): string                       // '1 ba', '1.5 ba'
                       formatSqft(n: number): string                            // '1,450 sqft'
                       formatSource(source: string): string                     // 'MLS A'
                       formatStatusLabel(status: ListingStatus): string | null  // 'PENDING' | 'SOLD' | null (active)
                       formatLocation(l: Pick<ListingResult, 'city' | 'state' | 'zip'>): string // 'Springfield, VA 22150'
                       formatListingMeta(item: ListingResult, currentYear?: number): string[]
                         // [location, bd, ba, sqft, listed, source]; status label excluded (BC-10 styles it as Label)
                       formatScore(score: number): string                       // '0.74' (toFixed(2))
                       formatRank(rank: number): string                         // 'No. 1'
                       formatMonogram(city: string): string                     // 'Sp' (first two letters, Ul)
                       formatResultCount(n: number): string                     // '1 result', '12 results', '0 results'
                       formatOrdering(targetBudget: number | null): string      // 'ranked for $450,000' | 'newest first'
                       formatResultSummary(data: SearchResponse): string        // StatusLine left half for a 2xx:
                         // '12 results · ranked for $450,000' | '12 results · newest first' | '0 results'
                       formatPageOf(page: number, totalPages: number): string   // 'Page 1 of 2'; '' when totalPages 0
src/test/fixtures.ts   listingA1, listingA3, listingA7: ListingResult; searchResponse, emptySearchResponse:
                       SearchResponse; cities: string[]; validationErrorBody; internalErrorBody;
                       jsonResponse(body, init?): Response; deferred<T>(): Deferred<T>
```

Usage notes for the completion comment:
- Components call the date and meta formatters with no year. Their tests pin the clock with
  `vi.useFakeTimers({ toFake: ['Date'], now: new Date(2026, 9, 13) })`.
- BC-11 owns the state-to-halves mapping: `Loading`; on `invalid`, the `previous` response; empty on `error`.
- SearchPage wires `retry` to ErrorState and does not log.

## Files (all new, under `frontend/src/`; tests sit next to their module)

`api/types.ts`, `api/errors.ts` + `.test.ts`, `api/client.ts` + `.test.ts`, `hooks/useSearchParamsState.ts` + `.test.tsx`,
`hooks/useListingSearch.ts` + `.test.tsx`, `hooks/useCities.ts` + `.test.tsx`, `format/money.ts` + `.test.ts`,
`format/date.ts` + `.test.ts`, `format/listing.ts` + `.test.ts`, `test/fixtures.ts`. Plus `docs/SPEC.md` §7.2.

## Tests (mapped to the ticket's scope and acceptance, and to SPEC §8)

`fetch` is `vi.fn<typeof fetch>()` installed with `vi.stubGlobal`; responses come from `jsonResponse`. Node's
`Response` stays global under Vitest's jsdom env.

- **client.test.ts (query building, error parsing):** `buildSearchQuery`: `{}` gives `''`; all nine params give one
  exact string in §5.3 order; `status` repeated, deduped, enum-ordered; `''`, `[]`, `NaN`, `Infinity` omitted; plain
  numbers (`targetBudget=450000.5`); `keyword=pets+allowed`. `searchListings` sends the exact URL and signal and returns
  the fixture; `{}` hits `/api/listings/search` with no `?`. Errors: a 400 with the §5.5 body gives `validation` with
  three fields and the `requestId`; a 500 gives `http` 500 with the body's message; a 404 gives `http` 404; a 502 HTML
  body gives the fallback; a non-JSON or wrong-shape 2xx gives `Malformed response body.`; a `TypeError` gives
  `network` with its message; an `AbortError` gives `network` with `Request aborted.`. `getCities`: success, a 500 and a
  wrong shape.
- **errors.test.ts:** `parseErrorResponse` keeps valid fields, filters malformed entries, gives `[]` for missing fields
  and the fallback for an empty or non-object body. `toApiError` on an `ApiRequestError`, an `Error` and a non-Error.
- **money, date and listing tests (every BRANDING §7.1 row plus the §5.3, §5.6 and §5.7 strings):** `$450,000`, `$0`,
  `$1,234,567`, `399000.4` gives `$399,000`; `2 bd`; `1 ba`, `1.5 ba`, `2` gives `2 ba`; `1,450 sqft`, `980 sqft`;
  `Listed Aug 29` (year 2026), `Listed Aug 29, 2026` (2025 or 2027), `Listed Sep 2`, `Listed Jan 1`, all twelve
  months, the `Listed garbage` fallback, and the default year from a faked clock. The date suite runs under
  `vi.stubEnv('TZ', 'America/Los_Angeles')` (Node re-reads TZ on assignment, verified), with `vi.unstubAllEnvs()` in
  `afterAll`. Also `MLS A`; `PENDING`, `SOLD`, `null`; `0.94`, `0.74`, `1.00`, `0.00`; `No. 1`, `No. 11`; `1 result`,
  `12 results`, `0 results`; `Page 1 of 2`, `''` at 0 pages; `ranked for $450,000`, `newest first`.
  `formatResultSummary` on the fixtures gives `12 results · ranked for $450,000`, `… · newest first` with a null
  budget, and `0 results`. `formatLocation(listingA1)`; `formatListingMeta(listingA1, 2026)` in §5.4 order.
  Monograms: `Sp`, `Vi`, `Fa`; `RESTON` gives `Re`; `'  Manassas '` gives `Ma`; `''` gives `''`.
- **useSearchParamsState.test.tsx** (`beforeEach`: `history.replaceState(null, '', '/')`): every param round-trips,
  and for the full set `parseSearchParams(buildSearchQuery(p))` deep-equals `p`. Drops `minPrice=abc`,
  `minBedrooms=2.5`, `page=0`, `pageSize=101`, `status=bogus`, `city=`; keeps `minPrice=-5`; accepts
  `status=active,sold`. Reads the URL on mount. `setSearch` pushes without `page` (URL updated, `history.length` +1,
  params updated). `pageSize` survives `setSearch`, `setPage` and `clear`. `setPage(2)` changes only `page`;
  `setPage(1)` removes it; `setPage(0)` is a no-op; `clear()` gives `''`. Popstate: `pushState` outside the hook, then
  `act(() => dispatchEvent(new PopStateEvent('popstate')))` (jsdom's `history.back()` is async, so it is not used).
  An identical search pushes nothing; `params` identity is stable across a rerender.
- **useListingSearch.test.tsx:** `loading` (no `previous`) then `success`, with the right URL; `loading` then `empty`;
  a params change gives `loading` with `previous`. A change mid-flight aborts the first signal; its late resolution is
  ignored and the second request's data wins (`deferred`). An equal-but-new object, including a reordered `status`,
  does not refetch. A 400 gives `invalid` with `fields` and `previous`, unlogged. A 500 gives `error` (`http`) with
  `previous` and one `console.error` with the request ID. A `TypeError` gives `error` (`network`). `retry()` refetches
  the same URL (`loading` with `previous`, then `success`). A page past the end gives `success` with zero items.
  Unmount aborts. `reactStrictMode: true`: two fetches, the first aborted, final `success`, the abort unlogged.
- **useCities.test.tsx:** success yields the six cities from one fetch of `/api/cities`; a 500 yields `[]` and one
  `console.error` (spied with `mockImplementation`); under StrictMode the aborted request is not logged.
- **Compile acceptance:** `fixtures.ts` is type-checked by `tsc -b` and used by the client, hook and format tests.

Mechanics: resolve deferreds with a synchronous `act(() => { d.resolve(r); })`, then `await waitFor(…)`, since body
reads take several ticks (an `async` callback with no `await` trips `require-await`). Read call arguments with
`fetchMock.mock.calls[0]?.[1]?.signal`.

## Steps

`api/*` and its tests; `format/*` and its tests; `test/fixtures.ts`; the hooks and their tests; the SPEC §7.2 sentence;
`npm run format` (`lint-web` runs `prettier --check`); the gate; a Conventional Commit such as
`feat(frontend): API client, URL and search hooks, formatters (BC-8)` with the attribution trailer. Never stage `.lattice/`.

## Gate (card note 5)

From the worktree root, run `make lint-web test-web`. Its sentinel runs `npm ci`, because the worktree has no
`node_modules`. Then run `cd frontend && npm run build` (`tsc -b --noEmit && vite build`). Run the full `make check` only
if BC-2 has landed on `master` by hand-off; until then the api half is BC-1's stub.

## Validation (card note 6: N/A until BC-12 wires a page)

With the gate green, run `cd frontend && npx vitest run --reporter=verbose > <scratchpad>/bc8-vitest.txt 2>&1`, then
`lattice comment BC-8 --role validation` with one line: "N/A: no page consumes these modules until BC-12; vitest
verbose: <files>/<tests> green, covering client query building and error parsing, useSearchParamsState history/popstate,
useListingSearch transitions, abort, StrictMode and retry, useCities, and every BRANDING §7.1 formatter row". Then
`lattice attach BC-8 <file> --role validation`.

## Shared-file handling (run-state table)

BC-8 creates and owns every file above. BC-9, BC-10 and BC-11 branch from BC-8 and may only **add** exports to
`src/format/*` and `src/test/fixtures.ts`; changing a listed signature is a SPEC-level change routed through the
orchestrator, never a fork. Untouched: `App.tsx`, `styles/*`, `test/setup.ts`, `Makefile`, `package.json`, the
lockfile, eslint/vite/tsconfig. `docs/SPEC.md` follows SPEC's deviation rule (in-branch edit plus `lattice comment`).

## SPEC edits

**SPEC edit: §7.2, one added sentence (a clarification, not a deviation):** "`useListingSearch` returns
`{ state, retry }`; `retry()` re-runs the current params (Try again); params are compared by their serialized query, so
an equal new object does not refetch; the hook logs http and network failures with the request ID (§7.5), and callers
do not log them again."

Why: SPEC fixes neither the return shape nor who logs. BC-12 needs a retry path, because re-pushing the same URL does
not refetch, and it needs to know not to log twice. The §7.3 types are unchanged; `ApiRequestError` (decision 3) is
how errors are delivered, not a type change.

## Risks

- **Timezone:** dates are parsed by parts and no `Date` is built; tests cover Jan 1 and a pinned US time zone.
- **StrictMode double effects:** each mount owns its controller, cleanup aborts it, and an aborted request never sets
  state. Tested.
- **react-hooks 7:** derived loading avoids `set-state-in-effect` (probed). No refs. `exhaustive-deps` is only a
  warning, but `--max-warnings 0` makes it fatal, so deps are strings and stable callbacks.
- **Typed lint:** `only-throw-error` is met by `ApiRequestError`; `no-unsafe-*` by `const body: unknown = await
  res.json()` plus guards; `unbound-method` by property-function members; `no-floating-promises` by `.then(ok, fail)`
  with both handlers and no `async` effects.
- **exactOptionalPropertyTypes:** never assign `undefined` to an optional property. Use a conditional spread or
  `delete` on a local copy, and pass `signal ?? null`.
- **noUncheckedIndexedAccess** (regex groups, month table, `mock.calls`); **act warnings** (`waitFor` settled state).
- **jsdom history:** the default URL is `http://localhost:3000/`, and `history.length` grows within a file, so tests
  assert deltas and dispatch popstate by hand. **Parallel `format/*` adds by BC-9..11** merge as a union; signatures
  are frozen at hand-off.

## Ambiguities resolved here (a reviewer may flip any; each is a one-line change)

1. `clear()` keeps `pageSize`. §7.1 says both "Clear writes an empty query string" and "pageSize … preserved if present".
2. A page past the end is `success` with `items: []`, despite the §7.2 comment `items.length > 0`.
3. The hook never emits `idle`.
4. Which URL values count as invalid (decision 10).
5. `previous` is the latest 2xx, including an empty one.
6. Logging lives in the hook.
7. Re-submitting identical filters does not refetch; BC-12 may call `retry` for that.
8. URL A, then B, then back to A while B is in flight shows A's settled result during A's refetch
   (stale-while-revalidate, no `loading` flash).

## Size

Production about 620 LOC: types 45, errors 85, client 110, useSearchParamsState 110, useListingSearch 120, useCities 30,
money 12, date 40, listing 90. Fixtures about 170. Tests about 1,000. One SPEC sentence.
