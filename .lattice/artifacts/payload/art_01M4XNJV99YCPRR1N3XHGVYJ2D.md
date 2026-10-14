# Plan review: BC-25 quiet pagination (agent:fable-plan-reviewer, 2026-10-14)

Reviewed cold: the plan at `.lattice/plans/task_01M4XMEXWQ2358J32KSKGZRYBV.md`, the ticket's two comments (operator ruling "quiet swap"; orchestrator decision: open question 1 = option A), and the worktree source it changes (`useListingSearch.ts`, `SearchPage.tsx`, `StatusLine.tsx`, `useSearchParamsState.ts`, `FilterBar.tsx`, `Pager.tsx`, `LoadingLine.tsx`, the three test files, fixtures, BRANDING.md, docs/SPEC.md, README.md, docs/branding-demo.html).

## Verdict: PASS-WITH-NITS

The design delivers the ruling, the page-change detection is correct against the hook's real code, the contract quotes are verbatim and complete, and the scope is proportional. One Major finding is a missing test, not a design defect: it is a one-item amendment to plan section 6 and should be applied before `in_progress`. The rest are Minor and Nit.

## Findings

### Critical

None.

### Major

**M1. The design's discriminating case has no test (plan section 3 argument; section 4 row "A filter search in flight, then a pager click").**
Section 3 chooses "compare against the list on screen (`lastGood`), not the previous params" and rejects alternative (a) precisely because (a) gets this case wrong: settled A, filter search B in flight, pager click writes `B&page=2`, which looks page-only against the previous params but must be a search because A's list is on screen. Nothing in section 6 exercises it, and acceptance criteria 7 and 9 do not name it. An implementation that compares against the previous params would pass every listed test.
Recommendation: add a hook test (deterministic, cheap): `renderSettled({ targetBudget: 450000 })`; deferred; `rerender({ p: { city: 'Vienna' } })` and assert `toStrictEqual({ status: 'loading', previous: searchResponse })`; deferred; `rerender({ p: { city: 'Vienna', page: 2 } })` and assert the same shape with no `requestedPage`. Optionally a SearchPage test: `renderDefaultSettled()`, `replyLater()`, select City Vienna and click Search, then `replyLater()` and click `pageButton(2)` (the Pager stays rendered over the previous list while a search loads), assert `expectSearchLoading()` and status `['Loading', 'Page 1 of 2']`. Extend AC 7 with "including a pager click while that search is loading".

### Minor

**m1. T14 and T16 select a City right after `renderBudgetSettled()` (section 6).**
`renderBudgetSettled` waits for the status line only. `renderDefaultSettled` additionally awaits `screen.findByRole('option', { name: 'Vienna' })`, and every existing test that selects a city (T4, T10, T12, the 400-after-error test) uses that helper. `user.selectOptions` throws if the option is absent. Both fetches resolve on immediate promises so it will probably pass, but it is an ordering dependency the suite elsewhere avoids.
Recommendation: in T14 and T16, `await screen.findByRole('option', { name: 'Vienna' })` before `selectOptions`, or add that wait to `renderBudgetSettled`.

**m2. BRANDING section 5.9 "Page change" row does not carry the option-A decision (section 7).**
Option A (Search, Enter or Clear that changes only the page is quiet) is a deliberate departure from the literal ruling, recorded in the ticket. The proposed Loading row says "new filters (Search, Enter, Clear, back or forward)", which is consistent only through the word "new"; the Page change row lists only pager controls and back/forward. The branding owner reads BRANDING, not the ticket. SPEC section 7.2's proposed text already defines the rule by URL difference, so SPEC is fine.
Recommendation: one clause in the Page change row, e.g. "...between pages of the same filters, whichever control wrote the URL: Search or Clear that change only the page count as a page change."

**m3. Acceptance criterion 6 overstates (section 11).**
AC 6 says back/forward "behaves as in 1-5", and AC 5 includes "Focus moves to the results heading as before." That focus move lives in `SearchPage.changePage` (pager click only); `useSearchParamsState`'s popstate handler only sets params, and T7 does not assert focus.
Recommendation: AC 6 -> "behaves as in 1-5, except the focus move, which is pager-click behavior."

**m4. Branding demo heading (section 8).**
`docs/branding-demo.html` line 155 is `<h2>The four states the brief names</h2>`. The plan adds a fifth tile and updates line 4's CSS comment ("then the four states" -> "then the states") but not the h2.
Recommendation: decide explicitly. Either keep the h2 and have the note paragraph say a page change is a treatment, not a brief-named state, or reword the h2 (e.g. "The four states the brief names, and a page change"). Silence leaves the section contradicting its own tile count.

### Nit

**n1. Section 3 snippet comment contradicts its code.** The comment says "narrowing needs status === 'loading' in the same expression" while the expression uses the `busy` alias. TypeScript 6.0.3 (anything >= 4.4) narrows through a `const` alias of a discriminant check, so `busy && !holdError && state.requestedPage !== undefined` compiles as written. Either inline `state.status === 'loading'` or drop the comment.

**n2. `attempt` in `searchKey` is belt-and-braces, not load-bearing, given the SearchPage gate.** `attempt` changes only via `retry`, which only exists in `ErrorState` (rendered only while `holdError`), and `holdError` clears only on a 2xx, which writes `lastGood.searchKey` with the current attempt. So whenever `holdError` is false, `lastGood.searchKey`'s attempt already equals the current attempt. Keep it (it makes the hook's own contract "a retry is never a page change" hold standalone, and the hook test depends on it), but the plan's section 3 bullet reads as if it were the mechanism that keeps Try again full-treatment in the page; the gate is.

**n3. `Pager.tsx` line 4 prop doc goes stale.** "The page on screen, 1-based." During a quiet page change the prop is the requested page (it was already the URL page). The plan says do not change Pager; a one-line comment fix ("The requested page, 1-based: the URL's page") is fine to fold in or skip.

**n4. `toStrictEqual` in the hook tests is safe.** Verified: `fixtures.ts` has no `undefined` values (only `null`), so the JSON round-trip through `jsonResponse` yields structurally identical plain objects. Nothing to change; recorded so the implementer does not second-guess it.

**n5. Section 4 row "list on screen after a 400, then back/forward to a page of the list's filters -> Quiet" also covers Back from the 400 URL to the list's own page 1.** In URL terms that is a filter change (X -> A), but the rule makes it quiet because A's list is on screen. It is the right outcome (nothing flashes; the status line matches the request in flight), and SPEC 7.2's URL-difference wording covers it. BRANDING's "between pages of the same filters" does not describe it; no action unless the owner wants 5.9 exhaustive.

## Verified, no finding

- **Ruling delivered.** `FilterBar loading={false}` during paging makes `canSubmit` true and lets `submit()` past its `if (loading) return` guard, so Search, Enter and Clear all work mid-page-change; `LoadingLine active={false}` renders the aria-hidden track (no progressbar); `main aria-busy={busy}` stays true for both kinds of load; the StatusLine quiet branch keeps `formatResultSummary(previous)` and shows `formatPageOf(requestedPage, previous.page.totalPages)`, where `requestedPage = params.page ?? 1` is the same value the Pager gets as `page`, so the two always agree; the list stays because `shownResponse` returns `previous` while loading. Filter searches keep the full treatment because `requestedPage` is absent when the key differs.
- **Detection against the hook's real code.** `requestKey` and `searchKey` are both functions of `attempt` and `query`, so adding `searchKey` to the effect deps causes no extra fetch. The success callback closes over the effect's `searchKey`; the `controller.signal.aborted` check means a late or aborted response never writes `lastGood`. `lastGood.data` is the only thing `shownResponse` shows when not `holdError`, so "compare against the list on screen" is exact, not approximate. Walked: rapid clicks (both keys match; page 2 aborted), popstate between pages (quiet), T8 held-key path (status is `success`, no loading state at all, unchanged), filter search interrupting a page change (Search is enabled; `0|B` != `0|A` -> full treatment; late page response ignored), page change failing (error held; Try again bumps `attempt`; gate holds), back/forward while the error view is held (hook says page change, SearchPage gate says search, `withoutPrevious` strips `requestedPage` so StatusLine and SearchPage agree), page past the end (`Page 3 of 2`), `pageSize` change (in the key -> search), first load at `?page=2` (no `lastGood` -> search), Search with unchanged filters from page 2 (URL differs by page only -> quiet, matches option A and T17). No case found where the status line's count or order mismatches the request in flight, and none where a filter search is quiet.
- **Contract quotes.** All quoted lines match the worktree verbatim: BRANDING.md 5, 212, 220-221, 263; docs/SPEC.md 370-373, 378, 455, 476, 484-485, 504; README.md 247. The section-number corrections (status line is 5.3 not 5.6; "Search is disabled while loading" is 5.2 not 5.8) are right. Grep for other loading rules: BRANDING 6 (line 276) and the 7.2 strings table stay accurate; SPEC 438 (`LoadingLine` type) and 469 (focus after page change) are unchanged and still true; docs/EXTENSIONS.md has nothing on the loading treatment; source comments at `FilterBar.tsx:174` and `SearchPage.tsx:19` stay accurate. Nothing missed.
- **Tests are non-vacuous.** T6 and T7 invert today's `['Loading', ...]` assertions rather than loosening them; T16 proves Search stayed enabled by clicking it; the hook tests pair the positive case with filter, pageSize and retry negatives; StatusLine tests cover budget, no-budget, past-end and empty variants.
- **Scope.** Proportional: three source files, tests, the contract docs the ruling asks for, the demo tile BRANDING section 9 step 2 requires, one README line. No creep. Validation guidance (don't run `make up` from the worktree; Vite on 5174; delayed fetch; `window.__marker`) is sound.
