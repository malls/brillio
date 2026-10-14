# Code Review (cycle 2): BC-12 SearchPage
Reviewer: fable sub-agent (decision 04:02Z)
Verdict: PASS
Reviewed HEAD: bb78854286750e47db7b21028b7848fb232e408f
Gate: make check exit 0 (frontend 396 tests / 23 files; Go 5 packages ok; golangci-lint 0 issues; eslint --max-warnings 0, tsc -b --noEmit, prettier clean), npm run build exit 0, flake runs 3/3 (`npx vitest run src/pages`, 14/14 each, exit 0 each)

## Cycle-1 findings
- Minor, T8 late-response assertion gap (`frontend/src/pages/SearchPage.test.tsx:445-456`) — **resolved**. Commit `326f1cd` adds, after `waitFor(late.bodyUsed)`, an `await act(async () => {})` flush and then asserts `cardRanks()[0] === 'No. 1'`, `queryByText('No. 11')` absent, and `statusHalves()` still `['12 results · ranked for $450,000', 'Page 1 of 2']`, all before `refetch.resolve(...)` (the refetch now resolves at line 460). Mutation check: I deleted the data-path `if (controller.signal.aborted) return;` guard at `frontend/src/hooks/useListingSearch.ts:66-68`, ran `npx vitest run src/pages/SearchPage.test.tsx -t "ignores a page response"`, and it failed exactly at the new assertion (`SearchPage.test.tsx:454`, expected `'No. 1'`, received `'No. 11'`, exit 1). Restored with `git checkout --`; `git status --porcelain` empty afterwards. The test now distinguishes a working abort guard from a broken one.
- Nit, T11 first reply page meta vs ranks (`SearchPage.test.tsx:515-526`) — **resolved**. The `page=2&pageSize=5` reply now carries A1 at rank 6 and B7 at rank 7 with `relevance { score: 0.3536, budgetFit: null, recency: 0.3536 }` and `{ 0.3376, null, 0.3376 }`, plus a new assertion `cardRanks()` equals `['No. 6', 'No. 7']`. Spot-checked against `data/sample_listings.json` with SPEC §6.2/§6.3 at asOf 2026-10-13 (recency = 0.5^(ageDays/30), no budget so score = recency, newest first): A5 (39 d) 1, A4 (40 d) 2, A2 (41 d) 3, A3 (42 d) 4, B8 (44 d) 5, A1 (45 d, 0.3536) 6, B7 (47 d, 0.3376) 7, B9 8, B10 9, B11 10, A6 11, A7 12. Page 2 of size 5 is ranks 6-10 = A1, B7, B9, B10, B11 per §6.4, so the trimmed first two rows and their decimals are exactly right. The comment on line 515 states the trim.

## Summary
The two changes since 5ac2397 are the test-only fix commit `326f1cd` (+13/-1 in `SearchPage.test.tsx`, nothing else) and the merge of master `bb78854`. The merge is clean: `git merge-tree --write-tree 326f1cd f0dd9c3` yields tree `a6a6ef0e…`, identical to `bb78854^{tree}`, so it is the automatic merge with no hand resolution; what it brought in is BC-6's backend OpenAPI files (`backend/openapi.yaml`, `backend/openapi.go`, `backend/internal/httpapi/openapi*.go`, `router.go`/`router_test.go`, `yamlsubset_test.go`) and `.lattice/` state; `git diff --stat 5ac2397^2 f0dd9c3 -- frontend/` is empty, so master changed nothing in BC-12's area. `git diff master...HEAD` is still exactly the six frontend files (`SearchPage.tsx` +118, `SearchPage.test.tsx` +627, `fixtures.ts` +148/-0, `App.tsx`, `App.test.tsx`, `shell.module.css`), with no `.lattice/` paths.

I re-read `SearchPage.tsx` and the full test file cold against the hook (`useListingSearch.ts`), `useSearchParamsState.ts`, `Pager`, `StatusLine`, `ListingList`, SPEC §6.2-6.4 / §7.1-7.6 / §8 and BRANDING §5.9 / §7.2 / §8. Nothing cycle 1 missed rises to a defect: the `errorHeld` render-time state adjustment holds the error view through `loading` and `invalid` and releases on `success`/`empty` as the Resolution 2 deviation says; `shownResponse` keeps `previous` on `loading`/`invalid` and drops it on `error`; the Pager is keyed to the URL page and hidden on error and before any 2xx (`Pager` itself returns null at `totalPages <= 0`, so the empty state shows none); `formatPageOf` returns `''` at zero pages, so a loading state whose `previous` is an empty response reads `Loading` / `''` rather than `Page 1 of 0`; `splitKeyword` accepts `undefined`; `changePage` focuses the heading synchronously while the list is still mounted; `main` carries `aria-live="polite"` and `aria-busy`. `shell.module.css` keeps the five-row template `auto auto auto minmax(0,1fr) auto` with `.results` the only scroller; no `shell.row`/`styles.row` reference remains anywhere under `frontend/src`. The diff's only non-ASCII character is the BRANDING §7.2 `·` separator (BRANDING §8: no emoji).

Gate re-run on this exact HEAD: `make check` exit 0, `npm run build` exit 0 (tsc -b --noEmit, vite 52 modules), three consecutive `npx vitest run src/pages` runs 14/14. Working tree clean at the end (`git status --porcelain` empty, HEAD still `bb78854`).

Browser validation (plan items 1-7, including the scroller check that jsdom cannot cover) remains the validation stage's job and is not claimed here.

## Findings
### Critical
None.

### Major
None.

### Minor
None.

### Nit
None.
