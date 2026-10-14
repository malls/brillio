# Plan Review: BC-12 (SearchPage: compose shell, wire URL state and fetch, all states, pager; integration tests and browser e2e)

Reviewer: fresh-context plan reviewer (Fable 5.1). Reviewed against docs/SPEC.md §7.1–7.6, §8, §13; BRANDING.md §1, §4.2, §5.2, §5.3, §5.7, §5.9, §7.2; the BC-12 card; run-state; and the code actually present in the worktree `feat/BC-12-search-page` @1a3c5bb (hooks, FilterBar + model, StatusLine, LoadingLine, ListingList, Pager, NoResults, ErrorState, Header, shell/global CSS, fixtures, client, keyword formatter, backend request logging, seed data).

## 1. Verdict

**PASS**

Implementation can proceed. The findings below are all minor and belong in the "Obvious" triage bucket: short amendments to the plan text, one small decision to record, and one test-harness trap to call out. None changes the approach.

## 2. Summary

The plan composes the landed BC-8/9/10/11 parts into `src/pages/SearchPage.tsx` without re-implementing anything, and every prop, export and behavior it relies on exists exactly as described (checked: `useSearchParamsState` writers, `useListingSearch` state shape and `retry`, `FilterBar` props incl. `serverErrors` and the unmapped `role="alert"` line, `RESULTS_HEADING_ID` + `tabIndex={-1}` heading, `Pager` tolerating `page > totalPages`, `LoadingLine` transparent-at-rest track matching the demo's hairline + bar, `splitKeyword`, `buildSearchQuery` param order producing the exact URLs the plan asserts, ListingCard emphasis as `<em>`). The test matrix covers every §8 SearchPage bullet and every ticket-listed integration case, and the seven browser validation items have concrete, seed-derived oracles that I re-derived from `data/sample_listings.json` and found correct. The key concern is small: the plan's `shown` derivation brings the stale `previous` list back during the loading that follows an error (Try again), which the plan neither records nor guards.

## 3. Issues

**[minor] Step 0 / Risks — anchor state is stale relative to the board**
Step 0 says to merge `feat/BC-9-filterbar` at 9ce69e6 and to merge BC-10 "if it moved past fa7c08d". In fact: BC-9 @9d4925b (which contains 9ce69e6) is already merged into the worktree (1a3c5bb); BC-10 has **landed on master** (merge 5eaacdc, board sync 76d6a97, run-state 04:09Z), and the worktree lacks BC-10's final head 64bad4d and master's tip. `git merge-tree` shows master merges clean. The remaining review hold is on BC-9 only.
**Recommendation:** Rewrite Step 0 as: (1) `git merge --no-ff master` now (BC-10 landed; brings 64bad4d/76d6a97); (2) re-merge `feat/BC-9-filterbar` only if it gains commits; (3) hold `review` until BC-9 lands, then merge master again. Tell the delegator the new anchor shas so run-state's BC-12 row can be updated.

**[minor] Wiring (`shown`) — the stale list reappears during loading after an error**
`useListingSearch` keeps `lastGood` across errors, so after a 500/network failure the state is `error` (ErrorState on screen), and clicking Try again yields `loading` with `previous` = the last 2xx. The plan renders `previous` during `loading`, so the old list (possibly from a different query, e.g. item 5's `pageSize=3` page) flashes back with `Loading | Page 1 of 4` until the retry settles; if the retry fails again the sequence is ErrorState → old list → ErrorState. A 400 arriving while in `error` does the same via `invalid` + `previous`. BRANDING §5.9 says the previous results stay "unchanged", which from ErrorState means nothing comes back. Locally this is a few ms and invisible, and T4 only asserts the end state, so neither the tests nor validation item 6 would catch it. Related: the ticket text says "keep previous results on screen while loading, on invalid input and on error", which contradicts §5.9/SPEC §7.5 (ErrorState replaces the list) and the ticket's own validation item 6. The plan resolves this the right way (ErrorState) but "Findings for the delegator" says "no contradiction found".
**Recommendation:** Either (a) record the flash as an accepted behavior in Risks/Findings, or (b) add a one-line guard: remember the last settled status (a `useRef` updated when `state.status` is not `loading`) and, when `loading`/`invalid` follows `error`, keep rendering ErrorState (or nothing) instead of `previous`. Either way, add the ticket-vs-§5.9 "on error" wording to Findings so the next agent does not re-derive it.

**[minor] Tests T8 (and T2) — the rejecting empty queue turns a legitimate refetch into ErrorState**
The hook holds the page-1 result after `history.back()` (settled key matches), shows it immediately, and **refetches page 1**. With the plan's harness, an empty queue rejects with "unexpected search <url>", which `toApiError` maps to `{kind:'network'}`, the hook logs it and the state becomes `error`, so "first card still `No. 1`" would fail for the wrong reason. T2's post-Clear default search has the same shape (the plan only asserts the call, so it passes, but the rejection is logged through the mocked console).
**Recommendation:** In T8's steps, queue a pending `replyLater()` for the page-1 refetch (third search call) before `history.back()`, and assert `searchCalls().length === 3` with the third URL carrying no `page`. Add a sentence to the harness notes: "every search the page will issue needs a queued entry, including refetches of held results".

**[minor] Validation item 6 — the URL carries `pageSize=3` if run straight after item 5**
`setSearch` preserves `pageSize`, so picking City Vienna after item 5 pushes `?city=Vienna&pageSize=3`, not `?city=Vienna`. The expected status (`2 results · newest first | Page 1 of 1`) is still right, but a validator comparing URLs will stumble.
**Recommendation:** Start item 6 with "navigate to `http://localhost:15112/` first" (or state the `pageSize=3` URL explicitly).

**[minor] Wiring (Pager) — status line and pager disagree during a page-change load**
With `page={params.page ?? 1}` and StatusLine fed `previous`, the moment after clicking `2` reads `Loading | Page 1 of 2` while `2` is the oxblood current page. The plan's reasoning for URL-driven `page` is sound (immediate feedback, correct rapid Next presses, PagerProps already allows it) and the window is one request long.
**Recommendation:** No code change. Record it in "Findings for the delegator" as a BC-14/branding-owner note, next to the first-load LoadingLine ambiguity already listed.

**[minor] Focus and a11y — nested polite live regions**
`main` becomes `aria-live="polite" aria-busy` while ListingList's `<section>` is also `aria-live="polite"`, and StatusLine is `role="status"`. The plan calls the nesting "redundant but harmless"; some screen readers announce nested polite regions twice, and `aria-busy` on the outer region does not reliably suppress the inner one. SPEC §7.6 asks for exactly one live results region.
**Recommendation:** Keep the plan's choice (it is the only way to announce NoResults/ErrorState without editing BC-10), but record it as an explicit decision and hand BC-14 the question of dropping `aria-live` from ListingList's section once SearchPage owns the region.

## 4. Positive Observations

- **Verified against real code, not memory.** Every claim I checked holds: `setSearch` drops `page` and keeps `pageSize`; `clear` keeps `pageSize`; `retry` is the only refetch of unchanged params; `FilterBar` re-seeds drafts on a URL change and reveals seeded fields (so T3's both-marked behavior is real); `mapServerErrors` routes `query`/`malformed_query` to the alert line (T10); `RESULTS_HEADING_ID` is exported and the heading is focusable and inside the scroller; `LoadingLine` is transparent at rest so hairline + track never reads as 3px; `Pager` returns null at `totalPages <= 0`; `.row` is referenced only by the BC-7 placeholder, so deleting it is safe.
- **Seed-derived oracles are correct.** I re-derived the validation expectations from `data/sample_listings.json`: 12 listings newest-first (55 Elm Ct → 55 Elm Court as No. 10, A6/A7 on page 2); `pets` matches exactly 7 descriptions in the listed order (A1 "Pet friendly" and B11 "Pet restrictions" correctly excluded); Springfield + 450000 gives A1, B7, A2, B8 in that order; `pageSize=3` page 2 opens on `789 Pine Rd` No. 4; Vienna gives 2 results. `buildSearchQuery`'s fixed param order yields exactly `?city=Springfield&targetBudget=450000` and `?page=2&pageSize=3`.
- **The "no request was sent" proof works.** The backend logs `r.URL.Path` (no query), so `grep -c '"path":"/api/listings/search"'` counts every search regardless of params.
- **Honest about test limits.** The plan states plainly that T3 cannot prove `serverErrors` reaches the bar (the client marks the pair first) and designates T10 as the test that does; it notes jsdom cannot test scrolling and sends item 7 to the browser with a precise `eval` oracle.
- **Scope discipline.** No component edits, no new helper module, no scroll-to-top invention, fixtures add-only in a labelled block, `App.tsx` ownership per the run-state table. Decisions that could be re-litigated (Pager hidden on error, first-load LoadingLine) are recorded with their rationale.
- **Executable validation script.** Ports, env, health check, proxy check, evidence paths, per-item expected values, StrictMode duplicate-request warning, and teardown are all spelled out; a validator with no other context could run it.
