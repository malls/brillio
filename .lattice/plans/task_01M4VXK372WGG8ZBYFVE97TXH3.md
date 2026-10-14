# BC-11: Pager, StatusLine and LoadingLine components

Pager, StatusLine and LoadingLine: the three pinned chrome pieces around the results. Spec: docs/SPEC.md §7.4, §7.5, §7.6. Design: BRANDING.md §5.3 (status line), §5.7 (pager), §5.9 (loading row), §6 (motion), §2.2 jobs 2 and 5, §7.2 strings. Reference: docs/branding-demo.html.

Why: pagination must be visible and operable (handout), the status line is where "ranked for $450,000" tells the user what the ordering means, and the loading line is the only animation in the product.

Scope (props per SPEC §7.4):
- Pager: 48px, hairline above. `‹ Previous` left and `Next ›` right as text buttons; every page number centered with 16px gaps (the dataset is small; windowing is noted in docs/EXTENSIONS.md). Current page in oxblood weight 500 with a 1px oxblood underline and `aria-current="page"`. Disabled Previous/Next in ink-3 using the `disabled` attribute so they are not focusable. Glyphs are text, not icons. Hidden entirely when totalPages is 0.
- StatusLine: Meta 13 ink-2, two halves. Left derived from SearchState with the BC-8 formatters: `12 results · ranked for $450,000`, `12 results · newest first` when scoring.targetBudget is null, `1 result` singular, `Loading` while loading, `0 results` when empty, empty on error. Right: `Page 1 of 2`, empty when there are no pages.
- LoadingLine: 2px track in line color under the filter bar; when active a 40% oxblood segment slides left to right, 1.2s linear, looping; `role="progressbar"` with an accessible name; under prefers-reduced-motion a static full-width oxblood line (§6). Inactive renders the empty track so layout never shifts.
- Tests: every StatusLine string for every state (idle, loading with and without previous, success with and without budget, empty, invalid, error); singular/plural; Pager bounds and callbacks, aria-current, disabled ends not focusable, hidden at zero pages; LoadingLine active/inactive and the reduced-motion branch.

Acceptance:
- Strings match BRANDING §5.3 and §7.2 verbatim (tests assert exact text).
- Only oxblood uses are the pager text buttons, the current page and the loading segment.
- Tests and lint green; no emoji.

Validation: N/A until BC-12. Record a one-line justification with --role validation and cite the test output.
