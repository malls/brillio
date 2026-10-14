# BC-25: Pagination reloads the whole page instead of only the results region

Operator report 2026-10-14 (live demo at http://localhost:5173 via make up): changing pages reloads the whole site; only the results region should update. Expected: header, filter bar and layout stay mounted and visually stable; only the result list (and pager/summary) update. Root cause and fix to be determined in planning.

---

## Plan (agent:opus-5.5-planner, 2026-10-14)

Worktree: `/Users/forrest/Code/brillio-worktrees.noindex/bc-25-quiet-pagination`, branch `feat/BC-25-quiet-pagination`. All paths below are relative to the worktree root. Never write `.lattice/` files inside the worktree; with `LATTICE_ROOT=/Users/forrest/Code/brillio`, lattice commands write to the main checkout.

Complexity: low to medium. Three source files, three test files, two contract docs, the branding demo, and one README line. Frontend only, no API change.

### 1. Root cause (confirmed by reading the code)

The document does not reload. `Pager` renders `<button>`s, and `SearchPage.changePage` calls `useSearchParamsState.setPage`, which uses `history.pushState`. What looks like a reload is the full search loading treatment (BRANDING.md §5.9) firing on a page change:

- `useListingSearch` derives `{ status: 'loading', previous }` for every new request key (`frontend/src/hooks/useListingSearch.ts`, `deriveState`).
- `SearchPage` sets `const loading = state.status === 'loading'` and passes it to `FilterBar` (Search disabled, grey `stone` fill), to `LoadingLine` (2px oxblood segment under the bar), and to `main` (`aria-busy`).
- `StatusLine` reads `Loading` on the left. On the right it keeps `previous.page.page`, the old page. When the page arrives, everything flips back.

Against a local API this flashes the whole screen. The pager already moves `aria-current` at once, because `SearchPage` passes `page={params.page ?? 1}` and `params` updates synchronously in `navigate()`. The existing test T6 already asserts this.

### 2. Operator ruling (binding): quiet swap

A page change touches nothing outside the results. Search stays enabled. No loading line runs. The status line keeps its count and order text, and its right half shows the requested `Page N of M` at once. The pager's current page moves at once. The list swaps in place when the new page arrives, with no fade and no blanking. `main` keeps `aria-busy` while the page loads. Searches from the filter bar (Search, Enter, Clear, Try again), the first load, and back/forward that changes filters keep today's full treatment.

### 3. Design decision: how a page change is recognized

**Rule:** a load is a *page change* when the in-flight request differs from the request behind the list on screen **only by `page`**: same filters, same `pageSize`, same retry attempt. Everything else is a *search*.

**Where the rule lives:** in `useListingSearch`. It is derived, not stored, and it is exposed on the `loading` state as `requestedPage`:

```ts
| { status: 'loading'; previous?: SearchResponse; requestedPage?: number } // requestedPage only with previous
```

`requestedPage` is present only when `previous` is present and the request differs from previous's request only by page. Its value is the page asked for (`params.page ?? 1`).

**Why here and in this shape:**

- The hook already owns request identity (`requestKey = ${attempt}|${query}`). The question is "does this request differ from the response on screen only by page?" That needs the key of the response on screen (`lastGood`), and only the hook has it.
- Compare against the list on screen, not against the previous URL params. Example: a filter search to B is in flight while A's list is still shown, and the user clicks page 2. The URL is now `B&page=2`. Compared with the previous params (`B`) that looks page-only. But the list on screen is A's, so a quiet status line would mix A's count with a page of B. Compared with `lastGood` (A), it is correctly a search.
- Including `attempt` in the compared key means a retry is never a page change, so Try again keeps the full treatment from the hook's own contract.
- Carrying the requested page in the state keeps `StatusLine` a pure function of `state`. Its SPEC §7.4 props (`{ state: SearchState }`) stay unchanged, and the hook's contract holds: abort on change, aborted requests never update state, loading is derived.
- No new React state. `FetchState.lastGood` gains the search key it was fetched with (an annotation on existing stored state). Everything else is computed during render.

**Rejected alternatives:**

- (a) `SearchPage` stores the previous params with the "adjust state when a prop changes" pattern and compares them minus page. This adds stored state and gets the in-flight filter search case above wrong.
- (b) An intent flag set by `changePage` and cleared by Search/Clear. This adds stored state and still needs param comparison for back/forward, because popstate carries no intent.
- (c) Comparing `appliedFilters` from the response. It is `Record<string, unknown>` and is not the serialized query, so comparing it is fragile.

**SearchPage gate:** the error view is held across loads (`errorHeld`). When it is up there is no list, no pager and an empty status line, so the next load is always a search, even back/forward between two pages. `SearchPage` therefore derives:

```ts
const busy = state.status === 'loading';
// A page change is quiet (BRANDING.md section 5.9): only the results region knows it is loading.
const paging = busy && !holdError && state.requestedPage !== undefined; // narrowing needs status === 'loading' in the same expression
const loading = busy && !paging; // the search treatment
```

`withoutPrevious()`, used for the status line while the error view is held, already builds a fresh `{ status: 'loading' }`, so `requestedPage` is dropped there too. `StatusLine` and `SearchPage` therefore agree on when the load is quiet.

### 4. Edge cases (decided)

| Case | Behavior |
|---|---|
| Pager click (Previous, Next, a number) from a settled list | Quiet. Status: `<count · order>` / `Page N of M` (N requested, M from the list on screen). Pager `aria-current` on N. Old list stays. `main aria-busy="true"`. Focus moves to the results heading (unchanged). |
| Rapid clicks (page 2, then page 3 before 2 lands) | Quiet throughout. The page-2 request is aborted and ignored. Status shows `Page 3 of M`. |
| Back/forward between two pages of the same filters | Quiet (popstate sets params, and the hook sees a page-only difference). Pager current moves at once. |
| Back/forward to a key the hook still holds as settled (T8 path) | Unchanged: that result shows at once and refetches with no loading state. |
| Back/forward that changes filters | Search treatment (unchanged). |
| A page change, then a filter search before the page lands | The filter search wins. The page request is aborted. Full treatment: `Loading` / previous list's `Page n of M`, LoadingLine on, Search disabled. Search is enabled during the quiet load, so the user can do this. A late page response is ignored. |
| A filter search in flight, then a pager click | Search treatment continues: the request's filters differ from the list on screen. |
| A page change that fails (network or 5xx) | Today's error treatment: ErrorState replaces the list, status line empty (BRANDING §5.9 Error), pager hidden, error logged once by the hook. The quiet treatment ends the moment the request fails. |
| Try again after that failure | Search treatment (the attempt differs, and the error view is held). This matches the ruling. |
| Back/forward between pages while the error view is held | Search treatment (no list on screen; the SearchPage gate). The error view stays until a 2xx (unchanged SPEC §7.5 rule). |
| A page change that returns 400 | Practically unreachable: `page` has no upper bound in the API, and the UI drops pages below 1. If it happens (for example an overflowing hand-edited page via back/forward), today's invalid treatment applies: the list stays, the status line describes the list on screen, and an unmapped `page` field error shows in the bar's alert line. No special code. |
| A page past the end (only from a hand-edited URL via back/forward; the pager never offers one) | Quiet. Status `Page 3 of 2` while loading, then the 200 with empty items keeps `Page 3 of 2` (existing past-end rendering). The pager marks no page current, and Previous goes to the last page (unchanged). |
| A `pageSize` change with the same filters (URL only) | Search treatment: `pageSize` is part of the compared key. |
| First load (including a first load at `?page=2`) | Search treatment (no previous). |
| A list on screen after a 400 (invalid, list kept), then back/forward to a page of the list's filters | Quiet: the list on screen matches the request apart from page. |
| Search, Enter or Clear that changes **only** the page (Search with unedited filters from page 3, or Clear on `/?page=2` with no filters) | Quiet. This is the one case that departs from the ruling's wording; see §10. The URL written is identical to the one Page 1 would write, so a derived rule cannot tell them apart. Search with unchanged filters on page 1 is already a no-op today (`navigate()` skips an equal query). |

### 5. Code changes

**`frontend/src/hooks/useListingSearch.ts`**

- Add `requestedPage?: number` to the `loading` variant of `SearchState`, with a comment: set only with `previous`, when the request differs from previous's only by page (a page change, SPEC §7.2).
- `FetchState.lastGood` becomes `{ searchKey: string; data: SearchResponse } | null`. Update its comment: the latest 2xx response, and the search key it was fetched with.
- Add a helper that builds the search key. Copy the params, `delete copy.page` (the same pattern as `filterKey` in `filterBarModel.ts`; `exactOptionalPropertyTypes` is on, so do not assign `undefined`), and return `${attempt}|${buildSearchQuery(copy)}`. Compute `searchKey` in render next to `requestKey`.
- In the effect, set `lastGood: { searchKey, data }` on success and keep `prev.lastGood` on error. Add `searchKey` to the effect deps. It is a function of `attempt` and `query`, so it causes no extra fetch.
- In `deriveState`, take the request's `searchKey` and `page` (`params.page ?? 1`). In the not-settled branch, add `requestedPage: page` when `lastGood?.searchKey === searchKey`. Use the existing conditional-spread pattern so no `undefined` property is ever set. `previous` stays `lastGood.data` in every branch. Update the `useMemo` deps.
- Update the hook's header comment (one line) to mention page changes.

**`frontend/src/components/StatusLine/StatusLine.tsx`** (the `loading` branch of `describeState`)

- No previous: `['Loading', '']` (unchanged).
- Previous and `requestedPage`: `[formatResultSummary(previous), formatPageOf(requestedPage, previous.page.totalPages)]`, with a comment citing BRANDING §5.3/§5.9: the count and order stand, and the page is the one asked for.
- Previous without `requestedPage`: `['Loading', formatPageOf(previous.page.page, previous.page.totalPages)]` (unchanged).
- Update the function's header comment.

**`frontend/src/pages/SearchPage.tsx`**

- Move the `loading` derivation below the `holdError` block, then derive `busy`, `paging` and `loading` as in §3.
- `FilterBar loading={loading}`, `LoadingLine active={loading}`, `main aria-busy={busy}`.
- No other change. The Pager already reads `params.page`, and the list already stays (`shownResponse` returns `previous` while loading).

**`frontend/src/components/FilterBar/FilterBar.tsx`**: no behavior change. Optionally, add a one-line comment on the `loading` prop: "A search (not a page change) is loading: Search is disabled."

**Do not change** `LoadingLine`, `Pager`, `useSearchParamsState`, or any CSS.

### 6. Tests

Update existing tests in place. Do not delete them. Run `npm test` before and after.

**`frontend/src/pages/SearchPage.test.tsx`**

- Add two helpers near `statusHalves`:
  - `expectPageLoading()`: no progressbar, Search enabled, `main` `aria-busy="true"`.
  - `expectSearchLoading()`: a progressbar named `Loading`, Search disabled, `main` `aria-busy="true"`.
- **T6 (update)**: rename to "changes page quietly: aborts the old request, keeps the list and count while loading, focuses the heading, continues the ranks". While page 2 is pending:
  - Replace `expect(statusHalves()).toEqual(['Loading', 'Page 1 of 2'])` with `['12 results · ranked for $450,000', 'Page 2 of 2']`.
  - Add `expectPageLoading()`.
  - Keep the abort, URL, ranks-still-page-1, `aria-current` and focus asserts.

  After it resolves: `aria-busy="false"` and Search enabled, plus the existing asserts.
- **T7 (update)**: rename to "restores page 1 quietly on popstate". Replace `['Loading', 'Page 2 of 2']` with `['12 results · ranked for $450,000', 'Page 1 of 2']`. Add `expectPageLoading()`, and assert `pageButton(1)` has `aria-current="page"` before the response lands.
- **T14 (new) Search regression**: "keeps the full loading treatment for a filter search from page 2". Steps: `renderBudgetSettled()`, `reply(searchResponsePage2)`, click Page 2, wait for `Page 2 of 2`. Then `replyLater()`, select City `Vienna`, click Search. Assert:
  - `currentUrl()` is `/?city=Vienna&targetBudget=450000` (check the exact order `buildSearchQuery` writes).
  - Status is `['Loading', 'Page 2 of 2']`.
  - `expectSearchLoading()`.
  - The page-2 list is still on screen.

  Resolve, then assert Search is enabled and there is no progressbar.
- **T15 (new) failed page change**: "shows the error view when a page change fails, and Try again loads as a search". Steps: `renderBudgetSettled()`, `reply(internalErrorBody, { status: 500 })`, click Page 2. Assert the error heading, status `['', '']`, no pager, and that console.error was called once. Then `replyLater()` and click Try again. Assert:
  - Status `['Loading', '']`.
  - `expectSearchLoading()`.
  - The error view is still shown.

  Resolve with `searchResponsePage2`, then assert ranks No. 11 and No. 12 and `Page 2 of 2`.
- **T16 (new) interruption**: "lets a filter search interrupt a page change, with the full treatment". Steps: `renderBudgetSettled()`, `const page2 = replyLater()`, click Page 2, `expectPageLoading()`. Then `const search = replyLater()`, select City `Vienna`, click Search. Search must be clickable, which proves it stayed enabled. Assert:
  - `searchSignal(1)?.aborted === true`.
  - Status `['Loading', 'Page 1 of 2']`.
  - `expectSearchLoading()`.

  Resolve `page2` late (ignored, as in T8), then resolve `search` and assert the new list.
- **T17 (new, pins §10 option A; drop or invert if the operator picks B)**: "treats Search with unchanged filters from page 2 as a page change". Settle on page 2, `replyLater()`, click Search. Assert URL `/?targetBudget=450000`, status `['12 results · ranked for $450,000', 'Page 1 of 2']`, and `expectPageLoading()`.
- T0, T1, T12 and the 400-after-error test stay unchanged and keep proving the search treatment on first load and on Try again.

**`frontend/src/hooks/useListingSearch.test.tsx`** (existing tests stay green unchanged: `toEqual` fails on a stray defined `requestedPage`)

- "marks a page-only change with the requested page and keeps the previous response". Steps: `renderSettled({ targetBudget: 450000 })`, a deferred response, `rerender({ p: { targetBudget: 450000, page: 2 } })`. The state should `toEqual({ status: 'loading', previous: searchResponse, requestedPage: 2 })`. Resolve with `searchResponsePage2`. Then rerender back to `{ targetBudget: 450000 }`. The state should `toEqual({ status: 'loading', previous: searchResponsePage2, requestedPage: 1 })`. Settle it.
- "does not mark a change of filters or pageSize as a page change". Rerender to `{ targetBudget: 450000, city: 'Vienna', page: 2 }`, then to `{ targetBudget: 450000, pageSize: 5 }`. Each time the state should `toStrictEqual({ status: 'loading', previous: searchResponse })`.
- "does not mark a retry as a page change". Steps: settle `{ targetBudget: 450000 }`, rerender page 2 with a network failure, wait for `error`, then `retry()`. The state should `toStrictEqual({ status: 'loading', previous: searchResponse })`.

**`frontend/src/components/StatusLine/StatusLine.test.tsx`** (the existing `Loading` cases stay as the search case)

- "keeps the count and order and shows the requested page during a page change": `{ status: 'loading', previous: searchResponse, requestedPage: 2 }` gives `['12 results · ranked for $450,000', 'Page 2 of 2']`. The `noBudgetResponse` variant gives `['12 results · newest first', 'Page 2 of 2']`.
- "shows a requested page past the end": `requestedPage: 3` gives `Page 3 of 2`.
- "keeps 0 results during a page change over an empty list": `previous: emptySearchResponse, requestedPage: 2` gives `['0 results', '']`.

### 7. Contract amendments (in-branch, per the ruling)

Note on section numbers: the status line is BRANDING **§5.3**, not §5.6 (§5.6 is Score). "Search is disabled while loading" is in **§5.2** (Filter bar actions), not §5.8 (§5.8 has no loading rule). §6 Motion ("results replace in place with no fade") is still accurate and is unchanged.

**BRANDING.md line 5** (version bump, §9 step 4)
- Current: `| **Status** | v1 draft, 2026-10-13. Composed from the branding interview picks; awaiting owner approval of the demo. |`
- Proposed: `| **Status** | v1.1 draft, 2026-10-14. v1 composed from the branding interview picks; v1.1 makes a page change quiet (§5.9, BC-25). Awaiting owner approval of the demo. |`

**BRANDING.md §5.2, line 212**
- Current: `Actions: the Search button (§5.8) and a "Clear" text button. Search submits; Enter in any field submits. Search is disabled while loading.`
- Proposed: `Actions: the Search button (§5.8) and a "Clear" text button. Search submits; Enter in any field submits. Search is disabled while a search loads; a page change leaves it enabled (§5.9).`

**BRANDING.md §5.3, lines 220-221**
- Current: ``- Left: `12 results · ranked for $450,000`. Without a target budget: `12 results · newest first`. While loading: `Loading`. No results: `0 results`. On error: empty.``
- Proposed: ``- Left: `12 results · ranked for $450,000`. Without a target budget: `12 results · newest first`. While a search loads: `Loading`; while a page loads, the count and order stay. No results: `0 results`. On error: empty.``
- Current: ``- Right: `Page 1 of 2`. Empty when there are no pages.``
- Proposed: ``- Right: `Page 1 of 2`, the page on screen; while a page loads, the page asked for. Empty when there are no pages.``

**BRANDING.md §5.9, line 263**: replace the Loading row with two rows.
- Current: ``| Loading | Status line + a 2px line under the filter bar | Status reads `Loading`. The previous results stay on screen, unchanged. A 2px `oxblood` segment (40% wide) slides left to right along a `line`-colored track, 1.2s linear, looping. Search is disabled. First load with no prior results: the status line alone. |``
- Proposed:
  - ``| Loading | Status line + a 2px line under the filter bar | A search: the first load, new filters (Search, Enter, Clear, back or forward), or Try again. Status reads `Loading`. The previous results stay on screen, unchanged. A 2px `oxblood` segment (40% wide) slides left to right along a `line`-colored track, 1.2s linear, looping. Search is disabled. First load with no prior results: the status line alone. |``
  - ``| Page change | Results region | Only the page changes: Previous, Next, a page number, or back or forward between pages of the same filters. Nothing outside the results moves: no loading line, Search stays enabled, the status line keeps its count and order and shows the page asked for at once, and the pager marks that page current. The previous page stays on screen, unchanged, until the next one arrives and replaces it in place. After an error there is no list to keep, so the next load is a search. |``

**docs/SPEC.md §7.2, lines 370-373** (append two sentences)
- Current: ``…Aborted requests never update state. The previous successful response stays available during loading so the list does not blank (BRANDING.md §5.9).``
- Proposed: ``…Aborted requests never update state. The previous successful response stays available during loading so the list does not blank (BRANDING.md §5.9). When the request differs from that response's request only by `page` (same filters and `pageSize`, not a retry), `loading` also carries `requestedPage`, the page asked for: a page change, rendered quietly (§7.5). Like `loading` itself, `requestedPage` is derived, never stored.``

**docs/SPEC.md §7.2, line 378** (type block; it must match the code)
- Current: `  | { status: 'loading'; previous?: SearchResponse }`
- Proposed: `  | { status: 'loading'; previous?: SearchResponse; requestedPage?: number } // requestedPage only with previous`

**docs/SPEC.md §7.4, line 455** (FilterBar paragraph)
- Current: `per §5.2, and disables Search while invalid or loading.`
- Proposed: ``per §5.2, and disables Search while invalid or `loading` (SearchPage passes `false` during a page change, §7.5).`` Rewrap the paragraph to the file's line width.

**docs/SPEC.md §7.5, line 476**: replace the one bullet with two.
- Current: ``- `loading`: StatusLine reads `Loading`, LoadingLine active, Search disabled, previous list stays.``
- Proposed:
  - ``- `loading`, a search (no `requestedPage`, or the error view is held): StatusLine reads `Loading`, LoadingLine active, Search disabled, previous list stays.``
  - ``- `loading`, a page change (`requestedPage` set and a list on screen): nothing outside the results changes. LoadingLine inactive, Search enabled, StatusLine keeps the previous count and order and shows `Page {requestedPage} of {totalPages}`, the Pager marks the requested page current, the previous list stays until the new page replaces it in place, and the results region is `aria-busy`.``

**docs/SPEC.md §7.6, lines 484-485**
- Current: ``Every input has a `<label>`; the results region is `aria-live="polite"` and `aria-busy` while loading;``
- Proposed: ``Every input has a `<label>`; the results region is `aria-live="polite"` and `aria-busy` while loading, a page change included;``

**docs/SPEC.md §8, line 504** (SearchPage test list)
- Current: `…page change issues a new request and aborts the old one, rank numbering continues across pages.`
- Proposed: `…page change issues a new request and aborts the old one with no search loading treatment (§7.5), a filter search keeps it, rank numbering continues across pages.`

**README.md line 247** (test summary, optional one-line accuracy fix)
- Current: `…an error with Try again, and paging with continuous ranks.`
- Proposed: `…an error with Try again, and quiet paging with continuous ranks.`

SPEC §13 (notes for the branding owner) needs no entry: BRANDING.md is amended directly.

### 8. Branding demo (`docs/branding-demo.html`)

The demo's States section has one **Loading** tile. It shows a filter-search load: Search `disabled`, the `.bar` loading line, status `Loading` / `Page 1 of 2`, and the list kept. That depiction stays correct for a search. No tile shows a page change. CLAUDE.md requires regenerating the local styling-examples artifact when BRANDING.md changes, and BRANDING §9 step 2 requires the demo to show the changed rule. So:

- Add a fifth tile after Loading, captioned `Page change`, built only from the existing classes and tokens:
  - Filters with an **enabled** Search.
  - **No** `.bar`.
  - Status `12 results · ranked for $450,000` / `Page 2 of 2`.
  - One page-1 card (`No. 1`).
  - A small pager at the tile foot with `2` as `.cur`: `.state .pager{ margin-top:auto; … }` if needed, within the 280px tile.
  - Keep the `Loading` caption on the existing tile.
- Update the States note paragraph. Proposed: "Loading a search keeps the previous results on screen and shows one moving line. A page change moves nothing but the results: the count stays, the page number moves at once, and the next page replaces the list when it arrives. No results and error are set in the display italic. Invalid input is marked on the fields themselves, in the one color that means attention."
- Update the layout comment on line 4 ("then the four states") to "then the states".
- The demo stays light-only by decision (BRANDING §2.5). Use no new colors and no hex outside `:root`.
- **Republish** (BRANDING §9 step 2): after the branch merges to master, the orchestrator reads, then republishes `docs/branding-demo.html` to the same URL, https://claude.ai/artifact/HdSoc6KB2Wzwvmjt73AGTn, and asks the owner to review it (§9 step 3). The implementer does not publish.

### 9. Gate and commands

Run from the worktree root (`node_modules` is already installed in `frontend/`).

- Fast loop: `make lint-web test-web` (eslint, `tsc -b --noEmit`, prettier check, vitest), plus `cd frontend && npm run build`.
- Before hand-off: the full `make check` (golangci-lint, `go test -race`, and the web checks). It must be green.
- Prettier runs only inside `frontend/` (`npm run format` to fix). BRANDING.md, docs/SPEC.md, README.md and the demo are outside its scope, so hand-wrap them to match their surrounding lines.
- Commits: conventional, no emoji, ending with the Co-Authored-By line from the session attribution. Suggested:
  - `fix(web): quiet page changes, no search loading treatment (BC-25)`
  - `docs(branding,spec): a page change is quiet (BC-25)`
- File links: `lattice file-link BC-25 frontend/src/hooks/useListingSearch.ts frontend/src/pages/SearchPage.tsx --reason "page-change detection (requestedPage) and quiet-swap gating" --actor agent:<id>`.

### 10. Open question for the operator (default chosen, flag if it should change)

**Search, Enter or Clear that changes only the page.** On page 3, pressing Search with unedited filters, or pressing Clear on `/?page=2` with no filters, writes exactly the URL that clicking Page 1 writes. The ruling lists Search, Enter and Clear under the full treatment.

- **Option A (default, this plan):** the treatment follows what changes in the URL, so these degenerate cases are quiet. The result is identical to clicking Page 1, and the list and the page number visibly respond. This needs no stored intent and matches "the URL is the search" (SearchPage header comment, SPEC §7.1). Search with unchanged filters on page 1 is already a silent no-op today.
- **Option B (literal ruling):** these cases show the full treatment. This needs stored intent in SearchPage: a flag set by Search/Clear, read at render, and cleared when the load settles. It also needs a test for it. It adds about 10 lines and one piece of state.

Test T17 pins option A. If the operator picks B, T17 inverts and §3 gains the intent flag.

### 11. Acceptance criteria

1. Clicking a pager control (Previous, Next, or a page number) from a settled list leaves Search enabled and starts no loading line: no `role="progressbar"`, and the track stays empty.
2. During that load the status line's left half keeps the count and order text (for example `12 results · ranked for $450,000`, or `12 results · newest first` with no budget). Its right half shows the **requested** `Page N of M` immediately.
3. During that load the pager marks the requested page `aria-current="page"` immediately.
4. During that load the previous page's cards stay on screen, unchanged. When the new page arrives they are replaced in place, with no blank frame, no fade, and ranks continuing (`No. 11` on page 2).
5. During that load `main` (Results) has `aria-busy="true"`, and `false` once the page lands. Focus moves to the results heading as before.
6. Browser back/forward between two pages of the same filters behaves as in 1-5.
7. A filter search (Search, Enter, or Clear with different filters), the first load, Try again, and back/forward to different filters keep the full treatment: status `Loading`, LoadingLine active, Search disabled, `aria-busy="true"`, previous list kept. This includes a pager click made while a filter search is still loading (the request still differs from the list on screen by more than the page).
8. A page change that fails shows the error view with an empty status line and no pager. Try again from there gets the full treatment.
9. A filter search started while a page change is in flight aborts the page request, gets the full treatment, and a late page response never shows.
10. The `useListingSearch` contract is intact: it aborts on change, aborted requests never update state, a held settled key shows at once, and `requestedPage` is derived (no new `useState`). Existing hook tests pass unchanged.
11. T6 and T7 in `SearchPage.test.tsx` are updated, not deleted. The new SearchPage, hook and StatusLine tests in §6 exist and pass.
12. BRANDING.md (§5.2, §5.3, §5.9, version line) and docs/SPEC.md (§7.2 prose and type, §7.4, §7.5, §7.6, §8) carry the §7 text or equivalent wording. The SPEC type block matches the code.
13. `docs/branding-demo.html` has the `Page change` tile and the updated note, uses existing tokens only, and has no emoji.
14. `make check` is green from the worktree root, and `npm run build` succeeds.

### 12. Validation guidance (for the in_validation agent)

- Do not run `make up` or `make down` from the worktree. The compose project name is fixed (`brillio`), so it would recreate the operator's live stack with the worktree mounts.
- Use the API on `localhost:8080`: the operator's stack if it is up, otherwise `make dev-api` from the worktree. Run the worktree's Vite on another port: `cd frontend && npm run dev -- --port 5174` (it proxies `/api` to `http://localhost:8080` by default).
- A local API answers too fast to see the loading state, so delay searches in the page before clicking:

  ```js
  const real = window.fetch.bind(window);
  window.fetch = (input, init) =>
    String(typeof input === 'string' ? input : input.url).startsWith('/api/listings/search')
      ? new Promise((r) => setTimeout(r, 2000)).then(() => real(input, init))
      : real(input, init);
  ```

  Aborts still work, because `real` receives the aborted signal.
- Then, with screenshots or DOM reads mid-flight and after:
  - Click Page 2 and check criteria 1-5.
  - Press Back and check criterion 6.
  - Change City and press Search, and check criterion 7.
  - Set `window.__marker = 1` first and confirm it survives (no document reload).
- Record the evidence with `lattice comment BC-25 --role validation`.

## 13. Plan-Review Cycle 1 Resolutions (AUTHORITATIVE: overrides earlier text on conflict)

Review: art_01M4XNJV99YCPRR1N3XHGVYJ2D (fable, PASS-WITH-NITS). Triage by agent:opus-5.5-orchestrator.

| Finding | Bucket | Resolution |
|---|---|---|
| M1 discriminating case untested | Obvious | Add a hook test: settle A; rerender `{ city: 'Vienna' }` (deferred, no `requestedPage`); rerender `{ city: 'Vienna', page: 2 }` (deferred) and assert still no `requestedPage`. Add the SearchPage variant too (Pager stays over the previous list during a search, click Page 2, Search stays disabled and LoadingLine stays active). Extend AC 7 with this case. |
| m1 T14/T16 City options not awaited | Obvious | Await `screen.findByRole('option', { name: 'Vienna' })` before `selectOptions`, or use `renderDefaultSettled`. |
| m2 BRANDING 5.9 row omits option A | Obvious | The 5.9 "Page change" row adds one clause: a Search, Enter or Clear that changes only the page is a page change too. |
| m3 AC 6 vs focus on popstate | Obvious | AC 6 carves out focus: back/forward behaves as AC 1-4; the focus move (AC 5) belongs to pager clicks only. |
| m4 demo heading "four states" | Obvious | Update the demo heading to match five tiles (for example "The states the brief names, and a page change"); BRANDING voice, no emoji. |
| n1 section 3 comment contradicts code | Obvious | Fix the comment to match the code (narrowing through the const alias). |
| n3 Pager.tsx:4 comment | Obvious | One-word fix: "The page requested" (or equivalent). |
| n2, n4, n5 | No action | Recorded as notes; keep `attempt` in the key, `toStrictEqual` is safe, n5 is covered by the SPEC wording. |
