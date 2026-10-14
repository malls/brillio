# BC-12: SearchPage: compose shell, wire URL state and fetch, all states, pager; integration tests and browser e2e

Compose the whole page: the pinned shell, URL state, fetching, every state, the pager, keyword emphasis, focus management, and the integration tests. Then prove it against the real API. Spec: docs/SPEC.md §7.1–7.6, §8 (SearchPage list). Design: BRANDING.md §4.2 (shell), §5.9 (states), §1 ("App, not document").

Why: this is the "wired together end to end" deliverable. Everything before it is a part; this is the product. Its validation is the real end-to-end check for the frontend.

Scope:
- src/pages/SearchPage.tsx composes Header, FilterBar, LoadingLine, StatusLine, the results region (ListingList, NoResults or ErrorState), and Pager in the §4.2 grid from BC-7; the results region is the only scroller.
- Wiring: useSearchParamsState feeds useListingSearch; FilterBar `value` comes from the URL, `onSearch` pushes new params (page reset), `onClear` clears the URL, Pager `onChange` pushes page; keyword tokens for emphasis come from the URL state; `serverErrors` come from the 'invalid' state; Try again re-runs the last request; cities from useCities.
- Initial load with an empty URL performs a default search (all listings, newest first). After a page change move focus to the results heading. Keep previous results on screen while loading, on invalid input and on error, per §5.9.
- Remove the BC-7 placeholders; App.tsx renders SearchPage.
- Integration tests with stubbed fetch and the BC-8 fixtures: loading then success renders cards and status; loading then empty renders NoResults and `0 results`; a 400 with min_gt_max marks both price fields and keeps the previous list; a 500 and a network failure render ErrorState and Try again re-fetches; page change issues a new request with page=2, aborts the previous one, and the first card reads `No. 11`; popstate restores page 1; the default search fires on mount.

Acceptance:
- All §8 SearchPage tests green; `make check` green for both halves.
- With api and web running the full flow works in a real browser.

Validation (required, this is the frontend's e2e gate): run the API (native or compose) and the web app; in a browser via the c11-browser skill: (1) default load shows 12 results, newest first, `12 results · newest first`; (2) Target budget 450000 and City Springfield gives 4 results ranked with A1 as No. 1 and `ranked for $450,000`; (3) Min price 500000 and Max price 400000 marks both fields with the message under Max price and sends no request; (4) Keyword `pets` emphasizes matches in descriptions; (5) with pageSize=3 in the URL, page 2 shows No. 4 and the back button returns to page 1; (6) stop the API, search, see `Couldn't load results.`, start the API, Try again recovers; (7) the document never scrolls, only the results region. Attach screenshots or the browser transcript with --role validation.
