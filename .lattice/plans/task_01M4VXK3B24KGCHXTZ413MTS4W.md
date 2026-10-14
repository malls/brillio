# BC-12 plan: SearchPage composes the shell, wires URL state and fetch, every state, pager; integration tests; browser e2e

Planner: agent:opus-5.5-planner. Worktree `/Users/forrest/Code/brillio-worktrees.noindex/bc-12-search-page`, branch
`feat/BC-12-search-page` (= master @c2e5544 + feat/BC-9 @1763fb3 + feat/BC-10 @fa7c08d). Binding: SPEC §5.3–5.5,
§7.1–7.6, §8; BRANDING §1, §4.2, §5.3, §5.7, §5.9, §7.2; run-state shared-files table and gate rule.

## Scope

Build `src/pages/SearchPage.tsx` from the landed parts. Compose them; do not re-implement them. App.tsx renders it. Add
the SearchPage integration tests, update App.test, run the gate, then prove the page in a real browser on this card's
ports. Out of scope: any API change, a new helper module, BRANDING edits, scroll-to-top on a new search (BRANDING does
not ask for it).

## Step 0 (before any code)

- `git merge --no-ff feat/BC-9-filterbar`. The anchor moved to **9ce69e6** ("FilterBar review nits"). It merges
  cleanly (`git merge-tree` rc 0). It changes the unmapped line: Clear no longer hides it, and it now follows
  `serverErrors` alone. Also merge `feat/BC-10-results-surface` if it moved past fa7c08d. Merge, never rebase. Tell
  the delegator the new anchor shas so run-state can be updated.
- `cd frontend && npm ci`. The worktree has no node_modules.

## Approach

### Tree and the §4.2 grid (shell.module.css: rows `auto auto auto minmax(0,1fr) auto`)

```
<div className={shell.shell}>                        // height 100%; html/body overflow hidden (global.css)
  <Header scope="Northern Virginia" />               // row 1 (same scope BC-7 used)
  <div>                                              // row 2: unstyled wrapper, so the five-row template stays
    <FilterBar .../>                                 //   applies its own inline pad + hairline
    <LoadingLine active={loading} />                 //   the 2px line under the filter bar (demo: .bar after .filters)
  </div>
  <StatusLine state={state} />                       // row 3
  <main className={shell.results} aria-label="Results" aria-live="polite" aria-busy={loading}>  // row 4, the only scroller
    {results}
  </main>
  {shown && <Pager page={params.page ?? 1} totalPages={shown.page.totalPages} onChange={changePage} />}  // row 5
</div>
```
When the Pager is absent (empty, error, first load), `main` is still row 4 (1fr), and row 5 collapses to 0.

### Wiring

- `const { params, setSearch, setPage, clear } = useSearchParamsState()` drives `const { state, retry } =
  useListingSearch(params)`, and `const cities = useCities()`. `loading = state.status === 'loading'`.
- FilterBar: `value={params}`, `cities`, `loading`, `serverErrors={state.status === 'invalid' ? state.error.fields :
  undefined}`, `onSearch={setSearch}` (setSearch drops page, keeps pageSize), `onClear={clear}`.
- Keyword tokens come from the URL (SPEC §7.4): `splitKeyword(params.keyword)` from `src/format/keyword.ts`. No new
  tokenizer.
- `shown` is the response whose list is on screen: `loading` uses `previous`, `success`/`empty` use `data`, `invalid`
  uses `previous`, and `error` uses undefined. Results: `error` gives `<ErrorState onRetry={retry} />`. No `shown`
  gives nothing (first load: "the status line alone", §5.9). `shown.page.totalItems === 0` gives `<NoResults
  onClear={clear} />`. Otherwise `<ListingList items={shown.items} keywordTokens={tokens} />`. So loading, invalid
  and error-free states keep the previous list unchanged (§5.9). A page past the end (`success` with empty items)
  renders the empty list plus a Pager, whose Previous recovers.
- The Pager's `page` comes from the URL, not from `shown.page.page`. The URL is the state (§7.1), a click shows the
  new current page at once, quick Next presses advance correctly, and PagerProps already says the page "may exceed
  totalPages on a hand-edited URL". `totalPages` comes from `shown`. The Pager is hidden on `error` (StatusLine
  already treats error as "no page on screen") and before any 2xx.
- The default search happens with no extra code: an empty URL gives `params = {}`, and the hook fetches
  `/api/listings/search` on mount. With no targetBudget the API scores by recency alone (§5.4, §6.2), so the order
  is newest first. StatusLine shows `12 results · newest first` / `Page 1 of 2` (formatOrdering(null)).
- Try again: `ErrorState onRetry={retry}`. The hook's retry re-runs the current params; it is the only refetch of
  unchanged params.
- The page never calls `console.error`; the hook already logs http and network failures (§7.2).

### Focus and a11y

- `changePage = (n) => { setPage(n); document.getElementById(RESULTS_HEADING_ID)?.focus(); }`. RESULTS_HEADING_ID is
  exported by ListingList for exactly this. On a Pager click the list is on screen and stays mounted through loading
  (previous) and success, so the same `h2` keeps focus. ListingList.module.css keeps the heading inside the scroller,
  so focusing it also scrolls the results to the top. Focus moves only on Pager clicks, not on popstate.
- §7.6: `main` is the polite live region and is `aria-busy` while loading. That way ErrorState and NoResults, which
  sit outside ListingList, are announced. ListingList's own `section aria-live="polite"` stays as BC-10 built it. The
  nesting is redundant but harmless (both polite).

### Component fixes

None needed. I read every prop and behavior the page uses and found nothing that requires a component edit. If the
implementer hits one, make the smallest fix and name the file and the reason in the commit and a lattice comment.

## Files (worktree-relative)

| File | Change |
|---|---|
| `frontend/src/pages/SearchPage.tsx` | **new** (SPEC §3 path), about 90 LOC |
| `frontend/src/pages/SearchPage.test.tsx` | **new**, about 420 LOC |
| `frontend/src/App.tsx` | replace the BC-7 placeholder body with `<SearchPage />` (run-state: BC-7 to BC-12) |
| `frontend/src/App.test.tsx` | rewrite: stub fetch (cities and default search), assert App renders the page (`Brillio Listings` heading, `search` "Filters", then `12 results · newest first`) |
| `frontend/src/styles/shell.module.css` | delete the placeholder-only `.row` rule and fix the header comment (row 2 = FilterBar + LoadingLine wrapper) |
| `frontend/src/test/fixtures.ts` | **add only**, in a `// --- BC-12 additions ---` block at the end (below) |

No changes to `src/format/*`, hooks, api or components.

### Fixture additions (real seed rows; values computed with SPEC §6.2 as of 2026-10-13, matching the existing fixtures)

- `listingB10` (rank 11: 100 Maple Dr, Reston 20190, 585000, 3 bd, listed 2026-08-20, "Townhome with 2-car garage,
  community pool."; relevance score 0.3549, budgetFit 0.4, recency 0.2872) and `listingA4` (rank 12: 22 Birch Ln,
  610000, listed 2026-09-03; 0.3321 / 0.2889 / 0.3969). Copy the other fields verbatim from `data/sample_listings.json`.
- `searchResponsePage2`: items [B10, A4], page `{2, 10, 12, 2}`, scoring and appliedFilters as `searchResponse`.
- `newestFirstResponse`: page 1 of the real no-budget search, trimmed (like `searchResponse`) to A5 `55 Elm Ct` No. 1,
  A4 No. 2, A2 `456 Oak Ave` No. 3. budgetFit is null and score equals recency (0.4061, 0.3969, 0.3878). Scoring is
  `{targetBudget: null, budgetWeight: 0, recencyWeight: 1, ...}`, page `{1, 10, 12, 2}`.
- `minGtMaxErrorBody` (only the maxPrice `min_gt_max` field) and `malformedQueryErrorBody` (`query` /
  `malformed_query` / `Query string is not valid URL encoding.`, the API's string in httpapi/query.go).

## Tests: `SearchPage.test.tsx` (vitest + RTL + user-event, no fake timers, no StrictMode wrapper)

Harness: `beforeEach` calls `window.history.replaceState(null, '', '/')` (jsdom history persists within a file),
`vi.stubGlobal('fetch', fetchMock)`, and `vi.spyOn(console, 'error').mockImplementation(...)`. `fetchMock` routes on
the URL. `/api/cities` gives `jsonResponse({ cities })`. `/api/listings/search...` takes the next entry from a
per-test queue: `reply(body, init)`, `replyLater()` (which returns a BC-8 `deferred<Response>`), or `failNetwork()`
(reject `new TypeError('Failed to fetch')`). An empty queue rejects with "unexpected search <url>". `searchCalls()`
filters `mock.calls` to search URLs. Abort is asserted through `searchCalls()[i][1]?.signal?.aborted`. `renderAt(url)`
does replaceState and then `render(<SearchPage />)`. Wrap deferred resolves in `act` and await with `waitFor`/`findBy`.

| # | Test | Maps to |
|---|---|---|
| T0 | Shell rows in §4.2 order: banner, wrapper (search form + LoadingLine), status, main, Pagination nav (after success) | ticket scope "§4.2 grid" |
| T1 | `/?targetBudget=450000`, deferred: status `Loading`, progressbar `Loading`, main `aria-busy=true`, Search disabled, no cards. Resolve: cards `No. 1`/`No. 4`/`No. 10`, status `12 results · ranked for $450,000` + `Page 1 of 2`, page 1 `aria-current`, no progressbar | loading → success (ticket, §8) |
| T2 | `/?city=Arlington` gives `emptySearchResponse`: `No results.`, `Adjust your filters.`, status `0 results`, no Pagination nav. Click `Clear filters`: URL `/`, search `/api/listings/search` | loading → empty (ticket, §8) |
| T3 | Default load succeeds. pushState `/?minPrice=500000&maxPrice=400000` and dispatch `popstate` (a back/forward to a hand-edited URL; the bar itself blocks this pair, see validation 3). The request carries both params, and the 400 `minGtMaxErrorBody` arrives. Min and Max price are `aria-invalid`, `Min price must be less than max price.` shows once, under Max price, Min price's `aria-describedby` points at it, the previous cards stay, the StatusLine is unchanged, no ErrorState, no `console.error` | 400 min_gt_max marks both, keeps list (ticket, §8) |
| T4 | Default load succeeds. Pick City `Vienna` and click Search: 500 `internalErrorBody`. `Couldn't load results.` shows, no cards, both status halves empty, no Pagination nav, `console.error` called exactly once (by the hook). Click `Try again`: a new search with the same URL `?city=Vienna`, success, cards back | 500 + Try again re-fetches |
| T5 | Mount `failNetwork()`: ErrorState. `Try again`: second search URL `/api/listings/search`, success | network failure + Try again (§8) |
| T6 | `/?targetBudget=450000` settled. Click `Page 2` (replyLater): URL `?targetBudget=450000&page=2`, the search call carries `page=2`, call 1's signal is aborted, status `Loading`, page-1 cards still shown, Results heading `toHaveFocus()`. Resolve `searchResponsePage2`: first card `No. 11`, `Page 2 of 2`, heading still focused | page change, abort, `No. 11`, focus (ticket, §8 "rank numbering continues") |
| T7 | Continue from T6's end state (its own setup). `window.history.back()` (jsdom queues the traversal and fires popstate; await with waitFor): URL `?targetBudget=450000`, the search has no page, resolve page 1, first card `No. 1`, `Page 1 of 2`, page 1 current | popstate restores page 1 |
| T8 | Page 1 settled. Click `Page 2` (replyLater A), then `history.back()` before A resolves: A's signal is aborted and page 1 is requested. Resolve A late with page 2: ignored, first card still `No. 1` | page change aborts the in-flight request (§8) |
| T9 | Empty URL: the first search URL is exactly `/api/listings/search`, `/api/cities` is fetched, `newestFirstResponse` gives `12 results · newest first`, first card `55 Elm Ct` `No. 1`, City options = Any + the 6 cities | default search fires on mount; cities from useCities |
| T10 | Default load succeeds. A 400 `malformedQueryErrorBody` (trigger with any submitted search) shows `Query string is not valid URL encoding.` in the bar's `role="alert"` line, no field is `aria-invalid`, and the list stays | card note 3: malformed_query via the unmapped line |
| T11 | `/?page=2&pageSize=5` settled. Type Keyword `pets` and press Enter: URL `?keyword=pets&pageSize=5` (page reset, pageSize kept), search with that query, then `<em>` `pets` in A3's description (tokens from the URL). Click `Clear`: URL `?pageSize=5` | onSearch / onClear / keyword wiring |

Notes: T3 is the only way to send min > max (the client rule blocks it). The client also marks the seeded pair, so
T10 is the test that proves `serverErrors` reaches the bar. In the browser the app never produces `malformed_query`,
because buildSearchQuery always emits valid encoding; only a stub can.

## Gate

From the worktree: `make check` (lint-api, lint-web = eslint + `tsc -b --noEmit` + prettier, test-api `go test -race`,
test-web), then `cd frontend && npm run build`. Both must pass on the exact head. Run `npx prettier --write` on the
new files first. Commit with conventional messages (`feat(frontend): SearchPage ... (BC-12)`); never commit `.lattice/`.

## Validation (browser, native, this card's ports; record HEAD sha)

Setup (`WT=/Users/forrest/Code/brillio-worktrees.noindex/bc-12-search-page`). First check that the ports are free:
`lsof -nP -iTCP:18012 -sTCP:LISTEN; lsof -nP -iTCP:15112 -sTCP:LISTEN` should print nothing.
1. `cd $WT/backend && go build -o bin/api ./cmd/api`
2. API, background, from the repo root so the default `./data/sample_listings.json` resolves: `cd $WT &&
   PORT=18012 DB_PATH=/tmp/brillio-bc12.db CORS_ORIGINS=http://localhost:15112 ./backend/bin/api
   >>/tmp/brillio-bc12-api.log 2>&1`. Check: `curl -s localhost:18012/healthz` prints `{"status":"ok"}`. The seed
   upserts, so the db file needs no deletion.
3. Web, background: `cd $WT/frontend && VITE_API_PROXY_TARGET=http://localhost:18012 npx vite --port 15112
   --strictPort >/tmp/brillio-bc12-web.log 2>&1`. Check: `curl -s localhost:15112/api/cities` returns the six cities
   through the proxy.
4. Open a c11 browser panel in **area:11** at `http://localhost:15112/` (c11-browser skill). Take one screenshot per
   item into `/Users/forrest/Code/brillio/.lattice/orchestration/evidence/BC-12-item<N>-<slug>.png`. Dev StrictMode
   sends a duplicate, aborted search on load. That is expected.

Expected results (seed computed from data/sample_listings.json; scores depend on today's date, so never assert
decimals):
1. **Default**: status `12 results · newest first` | `Page 1 of 2`. 10 cards, the first `55 Elm Ct` `No. 1`
   (Vienna), then `22 Birch Ln` No. 2 and `456 Oak Ave` No. 3, ending at `55 Elm Court` No. 10. Pager: Previous
   disabled, `1` current, Next enabled. The City select offers Any, Chantilly, Fairfax, Manassas, Reston,
   Springfield, Vienna.
2. **Budget + city**: type Target budget `450000` (it reads `$450,000` after blur), pick City `Springfield`, click
   Search. URL `?city=Springfield&targetBudget=450000`, status `4 results · ranked for $450,000` | `Page 1 of 1`,
   cards in order `123 Main St, Apt 4B` No. 1 (A1), `123 Main Street, Unit 4B` No. 2, `456 Oak Ave` No. 3,
   `456 Oak Avenue` No. 4.
3. **Min > max**: click Clear. Note `grep -c '"path":"/api/listings/search"' /tmp/brillio-bc12-api.log` (N). Type
   Min price `500000`, Tab, Max price `400000`, Tab. Both underlines turn oxblood (both `aria-invalid="true"`),
   `Min price must be less than max price.` shows under Max price only, and Search is disabled. Press Enter in Max
   price. The URL is unchanged and the grep count is still N: no request was sent.
4. **Keyword**: Clear, Keyword `pets`, Enter. URL `?keyword=pets`, `7 results · newest first` | `Page 1 of 1`. The
   seven descriptions (55 Elm Ct, 22 Birch Ln, 789 Pine Rd, 123 Main Street Unit 4B, 789 Pine Rd, 300 Cedar Blvd,
   42 Willow Way) each set `pets`/`Pets` at weight 600. Check with eval:
   `[...document.querySelectorAll('main em')].map(e=>[e.textContent,getComputedStyle(e).fontWeight])` should give
   7 pairs, all `pets`/`Pets` and `600`.
5. **pageSize=3**: go to `http://localhost:15112/?pageSize=3`. `12 results · newest first` | `Page 1 of 4`, cards
   No. 1–3. Click `2`: URL `?page=2&pageSize=3`, the first card is `789 Pine Rd` `No. 4`, `Page 2 of 4`, and
   `document.activeElement.id === 'results-heading'`. Press browser Back: URL `?pageSize=3`, the first card is
   `55 Elm Ct` `No. 1`, `Page 1 of 4`.
6. **API down**: `kill $(lsof -nP -iTCP:18012 -sTCP:LISTEN -t)`. In the bar pick City `Vienna` and click Search.
   The query must change: re-submitting identical filters is a no-op by design (§7.2), and only Try again refetches
   unchanged params. The Vite proxy answers 500, and the page shows `Couldn't load results.` and `Try again`, with
   both status halves empty and no pager. The console shows `Listing search failed`. Restart the API with the same
   command as setup step 2, wait for healthz, click `Try again`: `2 results · newest first` | `Page 1 of 1`, cards
   `55 Elm Ct` No. 1, `55 Elm Court` No. 2.
7. **Only the results scroll**: at `/`, eval `const r=document.querySelector('main'); r.scrollTop=10000;
   [document.scrollingElement.scrollTop, window.scrollY, r.scrollTop>0, getComputedStyle(document.body).overflow,
   getComputedStyle(r).overflowY, document.querySelector('header').getBoundingClientRect().top,
   Math.round(document.querySelector('nav[aria-label=Pagination]').getBoundingClientRect().bottom)===innerHeight]`.
   The result should be `[0, 0, true, 'hidden', 'auto', 0, true]`. If `r.scrollHeight <= r.clientHeight` (a tall
   panel), use `/?pageSize=12` instead. Also wheel-scroll over the list and take the screenshot mid-scroll with the
   header and pager visible.

Teardown: `kill $(lsof -nP -iTCP:18012 -sTCP:LISTEN -t) $(lsof -nP -iTCP:15112 -sTCP:LISTEN -t)`, then check that
both ports are free and close the browser panel. area:11 keeps its placeholder panel:15, so the area survives.
Evidence: `lattice comment BC-12 --role validation` (or attach a markdown) with PASS/FAIL for each of items 1–7, the
screenshot paths, and the HEAD sha.

## Shared files (run-state table)

`src/test/fixtures.ts`: add-only BC-12 block. `src/format/*`: untouched (reuse `splitKeyword`). `App.tsx`: BC-7 to
BC-12 per the table. `shell.module.css` (BC-7, not in the table): a dead-rule removal and a comment fix only.
tokens.css and global.css are untouched (BC-14 owns fixes). Makefile is untouched.

## Risks

- **Anchors move.** BC-9 has already moved (Step 0). If BC-9 or BC-10 gains review-fix commits again, merge the
  branch again before review. Review may not start until both land (run-state rule). After they land, merge `master`.
- **jsdom.** Layout is not tested (no scrolling), so item 7 is browser-only. `focus()` on a `tabIndex=-1` h2 works
  in jsdom. `history.back()` is async in jsdom, so await it. Deferreds from fixtures ignore `signal`, so assert abort
  through `signal.aborted` and prove "ignored" by resolving late.
- **StrictMode** (main.tsx) doubles the mount effects in dev: two searches and two cities fetches, the first of each
  aborted. Tests render without StrictMode, as the BC-8 hook tests do. The validator should expect the duplicate.
- **Keyword emphasis** follows the URL (the contract), so while a new keyword loads, the previous list already shows
  the new emphasis. This is accepted per §7.4.
- **A page change that errors** unmounts the focused heading, and focus falls to body. This is a rare edge and is
  accepted.
- **Size.** About 90 (page) + 420 (tests) + 110 (fixtures) + 35 (App and App.test) + -5 (css), so roughly 650 LOC
  added. The review diff against master stays under 1500 lines once BC-9 and BC-10 have landed. Until then the diff
  includes their code (well over 1500), which routes review to the fresh-context fallback reviewer.

## Findings for the delegator

- Contract: no contradiction found. One soft ambiguity: BRANDING §5.9 says first load is "the status line alone",
  while SPEC §7.5 says LoadingLine is active in `loading`. The plan reads §5.9 as "no list placeholder" and keeps the
  LoadingLine on during first load as well (SPEC is explicit). BC-14 can revisit.
- Decision recorded: the Pager is hidden on `error`. §7.5 lists only ErrorState and the empty StatusLine, and BC-11's
  StatusLine already treats error as "no page on screen".
- No human decision required.
