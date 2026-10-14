# Code Review: BC-12 SearchPage
Reviewer: fable sub-agent (decision 04:02Z)
Verdict: PASS-WITH-NITS
Reviewed HEAD: 5ac23977f2523a1fcc47e9797454c6e41f8a063c
Gate: make check exit 0 (frontend 396 tests / 23 files; Go 5 packages ok with -race; golangci-lint 0 issues; eslint, tsc -b, prettier clean), npm run build exit 0, flake runs 3/3 (`npx vitest run src/pages`, 14/14 each)

## Summary
The page is a thin composition over the landed hooks and components, exactly as the plan asked: no helper was forked, no component was edited, fixtures.ts is add-only, App.tsx hands BC-7's placeholder to `<SearchPage />`, and the only CSS change removes the dead `.row` rule (no remaining references). Every URL writer, the abort path, the previous-list behaviour on loading/invalid, ErrorState on error, the default search, heading focus, keyword tokens from the URL, `serverErrors` only from `invalid`, and Try again → `retry` are wired correctly and each is pinned by an integration test that asserts observable DOM, not mocks. The Resolution 2 guard is implemented with the documented set-state-during-render pattern (lint-clean, no ref mutation in render) and the declared deviation (hold the error view through `invalid` too, until a 2xx settles) is strictly closer to SPEC §7.5 / BRANDING §5.9 than the resolution's "last settled" wording; T12 and T12b cover both halves. I recomputed every BC-12 fixture value from `data/sample_listings.json` with SPEC §6.2 at asOf 2026-10-13: B10/A4 are ranks 11/12 with the stated score/fit/recency, the no-budget order is A5, A4, A2 with the stated recencies, and `keyword=pets` returns 7 rows starting A5, A4, A3. The one substantive finding is a test whose headline claim (a late response is ignored) is not actually distinguished by its assertions; it does not block.

## Findings
### Critical
None.

### Major
None.

### Minor
- `frontend/src/pages/SearchPage.test.tsx:446-465` (T8, "ignores a page response that arrives after the user went back") — after `page2.resolve(late)` the test only waits for `late.bodyUsed` and then immediately resolves the page-1 refetch. The final assertions (cards A1/B7, no `No. 11`, `Page 1 of 2`) would also pass if the hook's `controller.signal.aborted` guard were broken, because a leaked page-2 settle would be overwritten by the refetch's settle before anything is asserted. The plan's T8 row calls for "resolve A late with page 2: ignored, first card still `No. 1`". Recommendation: after `await waitFor(() => expect(late.bodyUsed).toBe(true))`, flush the promise chain (`await act(async () => {})`) and assert `cardRanks()[0] === 'No. 1'`, `statusHalves()` still `['12 results · ranked for $450,000', 'Page 1 of 2']` and `queryByText('No. 11')` absent, before `refetch.resolve(...)`. T6's `searchSignal(0)?.aborted === true` and the BC-8 hook tests cover the mechanism, so this is a test-strength gap, not a product bug.

### Nit
- `frontend/src/pages/SearchPage.test.tsx:509-512` (T11 first reply) — the response claims `page: { page: 2, pageSize: 5, totalItems: 12, totalPages: 3 }` but carries items ranked 1–3; per SPEC §6.4 page 2 of size 5 starts at rank 6. Nothing asserts those ranks, so it is harmless, but the file otherwise insists on real responses. Recommendation: re-rank the three items 6–8 (or add a one-line comment that only the page meta is real here).

Accepted without finding (already decided and handed to BC-14 by lattice comment): the nested polite live regions (`main` plus ListingList's `section`) and the two landmarks both named "Results"; the Pager marking the clicked page while the StatusLine still reads the previous page for the duration of one request.

## Acceptance / SPEC §8 coverage
- §8 loading → success renders cards, status line and pager: T1 (plus T0 for the §4.2 row order, `aria-live="polite"`, `aria-busy`).
- §8 loading → empty renders NoResults and `0 results`, Pager hidden: T2 (also proves `Clear filters` → `/` and the follow-up default search).
- §8 validation error mapped to the right field, `min_gt_max` marking both price fields, message once under Max price, previous list and status unchanged, no console.error: T3 (via popstate to a hand-edited URL, the only route past the client rule).
- Card note: `malformed_query` reaches the bar's unmapped `role="alert"` line with no field marked, list kept: T10.
- §8 network error with Try again re-fetching: T5; 500 with Try again (same URL, logged once by the hook with request ID, nothing in the UI): T4.
- §8 page change issues a new request with `page=2`, aborts the previous one (`signal.aborted`), keeps the page-1 list while loading, Pager follows the URL at once: T6. Late response after Back: T8 (see Minor — the abort is proven, the "ignored" half is not distinguished).
- §8 rank numbering continues across pages (`No. 11`, `No. 12`): T6, with fixtures verified against SPEC §6.2.
- Ticket: popstate restores page 1 (hook refetch queued, status `Loading` / `Page 2 of 2` during it): T7.
- Ticket: default search fires on mount with the exact URL `/api/listings/search`, cities loaded, `12 results · newest first`: T9; App-level smoke in App.test.tsx.
- Ticket / §7.1: search resets page and keeps `pageSize`, Clear keeps `pageSize`, keyword emphasis follows the URL (`<em>` per card in its own casing): T11; pager writes only `page`: T6.
- Ticket / §7.4: focus moves to the results heading after a page change and survives the load: T6.
- Resolution 2 guard: error view held through Try again's loading (no cards, `Loading` alone, no pager, progressbar active) and through a following 400 (empty status, alert line): T12, T12b.
- §7.6: `main` is the polite region and `aria-busy` while loading; progressbar named `Loading`; pager `aria-current="page"`: T0, T1, T6, T12.
- §4.2 / BRANDING §1: results region the only scroller — `shell.module.css` rows `auto auto auto minmax(0,1fr) auto`, `.results { min-height: 0; overflow-y: auto; overscroll-behavior: contain }`, html/body `overflow: hidden` in global.css. Not testable in jsdom; browser validation item 7 remains for the validation stage.
- Acceptance "`make check` green for both halves": verified on 5ac2397 (exit 0), plus `npm run build` exit 0 and 3/3 flake runs.
- Acceptance "full flow works in a real browser": out of scope for this review; owned by the validation stage (items 1–7 in the plan).
- BRANDING §7.2 copy: the page introduces no copy of its own; all strings come from the already-reviewed components. §8: no emoji in the diff (only the `·` separator from §7.2).
- Run-state shared-file rules: fixtures.ts add-only (+148/-0 in a labelled block), src/format/* untouched (`splitKeyword` reused), App.tsx BC-7 → BC-12, no `.lattice/` in the branch diff.
