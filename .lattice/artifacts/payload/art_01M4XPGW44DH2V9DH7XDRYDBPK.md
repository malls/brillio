Verdict: PASS-WITH-NITS
Reviewed HEAD: 400017d4498a11dc499be908c61ad55f8227e5ef

Reviewer: agent:fable-reviewer (fresh context; did not write the code).
Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-25-quiet-pagination, branch feat/BC-25-quiet-pagination.
Diff reviewed: git diff master...HEAD (12 files, +500/-59), commits f243484 and 400017d.
Plan: .lattice/plans/task_01M4XMEXWQ2358J32KSKGZRYBV.md, with section 13 as the authoritative resolutions block.

No critical, major or minor findings. Four nits, none of which needs rework before validation.

## Findings

### Critical
None.

### Major
None.

### Minor
None.

### Nit

N1. frontend/src/pages/SearchPage.test.tsx: no test for rapid pager clicks (plan section 4, row "Rapid clicks": page 2, then page 3 before 2 lands; status shows `Page 3 of M`, the page-2 request is aborted, quiet throughout). The derivation makes this follow directly from T6 plus the hook's existing abort-on-change test, and the plan's section 6 does not require it. Recommendation: optional; add it if a later ticket touches the pager.

N2. .lattice/plans/task_01M4XMEXWQ2358J32KSKGZRYBV.md section 11, AC 7: section 13 (M1) says "Extend AC 7 with this case" (a page click while a filter search loads keeps the full treatment), but AC 7's text was not amended. The behaviour is implemented and tested (T18, and the hook test "compares a page change with the response on screen, not with the params before"). Recommendation: orchestrator appends one clause to AC 7 in the plan file; nothing on the branch changes.

N3. frontend/src/hooks/useListingSearch.ts:11: the `loading` variant's type does not encode the invariant "requestedPage only with previous"; it is carried by a comment and enforced by the hook (requestedPage is set only when lastGood is non-null, which is exactly when previous is set). StatusLine.tsx:28-31 checks `previous` before `requestedPage`, so a violation could not crash it. This is the shape the plan chose (section 3). Recommendation: none; recorded so the next reader knows it is deliberate.

N4. docs/branding-demo.html:168-174: in the Page change tile the page-1 card's meta line is clipped by the pager at desktop width (the tile is 280px tall and the pager takes 48px of it). This matches the Invalid input tile, whose card is clipped by the tile edge, and the app shell's pinned pager over a scrolling list, so it reads as intended. Recommendation: none required; the owner sees it at the section 9 step 3 review after the republish.

## Correctness of page-change detection (checklist item 1)

Rule as implemented (useListingSearch.ts:36-43, 54-57): a load is a page change when `toSearchKey(attempt, params)`, the attempt plus the query with `page` deleted, equals the key stored on `lastGood`, the 2xx response whose list is on screen. `requestedPage` is `params.page ?? 1`. It is computed in `deriveState` on every render; no new `useState`.

Checked against the hook's contract and the cases in the brief:

- Abort on change, aborted requests never update state: unchanged. Both callbacks still return on `controller.signal.aborted` before any `setFetchState` (lines 88, 97). Effect deps gain `searchKey`, which is a function of `attempt` and the non-page params, both of which already change `requestKey`, so no request is added (the existing "does not refetch for equal params in a new object" test still passes).
- The key stored on `lastGood` is the key of the request that produced it: the success callback closes over the same render's `searchKey` as its `requestKey` (line 93). The error path keeps `prev.lastGood` (line 107), so the key on screen survives a failed page change.
- Page click during a filter search (M1): the request's filters differ from the list on screen, so no `requestedPage`; the search treatment continues. Tested in the hook (line 359) and T18. Comparing against the previous URL params would have got this wrong; the implementation compares against `lastGood`, as planned.
- Filter change during a page load: the new search key differs, full treatment, the page request is aborted, a late page response is read and ignored. T16.
- Rapid clicks: both requests share the search key, so quiet throughout; `requestedPage` follows the URL. Not directly tested (N1).
- Error hold: SearchPage.tsx:74-76 gates `paging` on `!holdError`, and the StatusLine receives `withoutPrevious(state)`, which rebuilds `{ status: 'loading' }` and so drops `requestedPage` too. The two agree on every held-error load. T15 (Try again) and T19 (back/forward while held).
- Retry: `attempt` is in both keys, so a retry never matches the list on screen. Hook test at line 383; T15.
- popstate between pages of the same filters: quiet, pager current at once, no focus move. T7. popstate to different filters: the key differs, search treatment (unchanged path).
- Page past the end: `Page 3 of 2` from `formatPageOf(requestedPage, previous.page.totalPages)` while loading, then the existing past-end rendering. StatusLine test at line 83.
- pageSize change: part of the compared key, so a search. Hook test at line 337.
- Status-line consistency: during a quiet load the left half is `formatResultSummary(previous)`, whose count and order (`formatOrdering(data.scoring.targetBudget)`) come from a response with the same filters, including `targetBudget`, as the request in flight; the right half's `totalPages` comes from the same response. I found no case where the count or order text disagrees with the request in flight.
- Filter searches going quiet: only if `buildSearchQuery` serialised two different filter sets identically. It is a canonical serialiser of the typed params and `parseSearchParams` drops empties, so the only "quiet" filter-bar actions are the option-A cases (Search, Enter or Clear that change only the page), which write the URL Page 1 writes. Decided by the orchestrator; T17 pins it.
- TypeScript: `paging = busy && !holdError && state.requestedPage !== undefined` narrows `state` through the const alias `busy`; `tsc -b --noEmit` passes.

## Tests (checklist item 2)

Non-vacuous by inspection: every quiet-load assertion (T6, T7, T16, T17) and every full-treatment assertion (T14, T15, T16, T18, T19) sits between the deferred reply and its `resolve`, so a wrong `loading` flips a progressbar query or a Search enabled/disabled query and fails the test. T6 and T7 are updated in place, not deleted. The new helpers `expectPageLoading` and `expectSearchLoading` each assert three independent facts (progressbar, Search enabled state, `aria-busy`).

Coverage against plan sections 11 and 13:
- AC 1-5: T6. AC 6 (with the m3 focus carve-out): T7. AC 7 including the M1 extension: T14, T18, T0/T1/T12 unchanged, hook tests at lines 337, 359, 383. AC 8: T15. AC 9: T16. AC 10: existing hook tests unchanged and green; `requestedPage` derived; abort and held-key tests unchanged. AC 11: T6/T7 updated; T14-T17 present; T18 (M1) and T19 (error-hold gate, implementer deviation 1) added. StatusLine: four new cases at lines 68-93.
- Option A (section 10, orchestrator decision): T17.
- m1 (await the City options before `selectOptions`): T14, T16 and T18 call `await screen.findByRole('option', { name: 'Vienna' })` first.
- Implementer deviation 3 (`toStrictEqual` on the positive page-change cases): fine, and it also proves no stray key is set.

## Contract text (checklist item 3)

BRANDING.md: status line bumped to v1.1 draft, 2026-10-14 with the BC-25 note; 5.2 "Search is disabled while a search loads; a page change leaves it enabled (5.9)"; 5.3 left and right halves describe the search and page-change cases; 5.9 Loading row now names what counts as a search, and the new Page change row carries the m2 option-A clause and the after-error clause. Section 9 step 2 (republish) is deferred to the orchestrator by decision; step 4 (version bump) is done.
docs/SPEC.md: 7.2 prose and the type line match useListingSearch.ts:11 exactly (including the trailing comment); 7.4 rewrapped; 7.5 has the two `loading` bullets, with the error-hold condition stated; 7.6 "a page change included"; 8 names the two SearchPage cases.
README.md:247: "quiet paging with continuous ranks".
Demo: Page change tile after Loading, Search enabled, no `.bar`, status `12 results · ranked for $450,000` / `Page 2 of 2`, one No. 1 card, pager with 2 current (markup idiom matches the app shell's pager at line 150). Heading per m4, note and layout comment updated, header label v1.1 (deviation 2, consistent with the version bump). One new rule, `.state .pager`, uses no colour. Rendered headless at 1400px: as described. Rendered at 390px: the right edge clips, and master's copy clips identically at that width, so it is pre-existing and not caused by this branch.
Voice: the diff's additions contain no emoji and no exclamation marks (scan of all `+` lines, 0 hits). Commit messages carry no emoji and end with the Co-Authored-By trailer.

## Style, idiom, scope (checklist item 4)

`toSearchKey` mirrors `filterKey` in filterBarModel.ts (spread then `delete`, safe under exactOptionalPropertyTypes). The conditional spread for `requestedPage` matches the existing `previous` spread. Header comments in the hook, StatusLine, Pager (n3) and FilterBar are accurate. The 12 changed files are the ones the plan names; LoadingLine, useSearchParamsState and CSS are untouched. Deviations 1-3 in the implementer's comment are reasonable and recorded.

## Gate (checklist item 5)

`make check` from the worktree root: exit 0.
- golangci-lint: 0 issues. go test -race ./...: ok.
- eslint --max-warnings 0, tsc -b --noEmit, prettier --check: clean.
- vitest: 23 files passed, 412 tests passed (was 398; 14 new).
`cd frontend && npm run build`: built in 109ms. Working tree clean afterwards.

## Acceptance criteria checked

1 pass (T6). 2 pass (T6, StatusLine tests). 3 pass (T6, T7). 4 pass (T6). 5 pass (T6). 6 pass (T7, focus carved out per m3). 7 pass including M1 (T14, T18, hook tests). 8 pass (T15). 9 pass (T16). 10 pass (hook unchanged in contract; existing hook tests green). 11 pass. 12 pass (text verified against the code). 13 pass (tile, note, tokens only, no emoji). 14 pass (make check exit 0, build OK).

Recommendation: advance to in_validation. Validation should follow plan section 12 (delay `/api/listings/search` in the page, check the quiet swap mid-flight, Back, a filter search, and the `window.__marker` survival).
