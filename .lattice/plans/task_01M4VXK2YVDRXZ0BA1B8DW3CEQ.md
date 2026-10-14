# BC-9: FilterBar: seven fields, client-side validation with exact copy, server error mapping, Search and Clear

The FilterBar: seven fields with the BRANDING field anatomy, client-side validation using the exact copy, server error mapping, Search and Clear. Spec: docs/SPEC.md §7.4 (FilterBar), §5.5 (codes). Design: BRANDING.md §5.2 (fields and states), §5.8 (buttons), §5.9 (invalid input row), §7.2 (strings), §4.4 (breakpoints), §3.2 (Label, Body, Button). Reference: docs/branding-demo.html.

Why: this is the "real inputs" deliverable and where invalid input is caught before it is sent. The Status field is the one addition beyond the brief (engineering interview decision, SPEC §13.1) and must follow §5.2's seventh-field rule exactly.

Scope (src/components/FilterBar, props per SPEC §7.4):
- Fields in order: Min price, Max price, Min bedrooms, City (select, "Any" first, options from the cities prop), Keyword, Target budget, Status (select: Any / Active / Pending / Sold). Labels and placeholders verbatim from §7.2.
- Anatomy: label above (Label style), value (Body 15), 1px ink-2 underline; focus 2px oxblood underline and no ring; invalid 2px oxblood underline with the message beneath in oxblood Meta 13; disabled ink-3 with line underline. Select shows the ▾ glyph in ink-2.
- Local draft state seeded from the `value` prop and re-seeded when it changes (URL navigation). Price and budget inputs show digits while typing and `$400,000` on blur (formatters from BC-8); min bedrooms step 1.
- Client-side rules with exact strings: `Min price must be less than max price.` (marks both price fields, message under Max price), `Min bedrooms must be 0 or more.`, `Target budget must be more than $0.`; negative prices use the same voice. Search disabled while invalid or while `loading`.
- Server errors: map `serverErrors` by field and code onto the same marks; `min_gt_max` marks both and shows under Max price; unmapped fields fall through to a single line under the bar.
- Search is the one primary button (§5.8), Clear a text button. Enter in any field submits. `onSearch` receives a SearchParams with numbers parsed, empty values omitted, status omitted when Any.
- Layout per §4.4: one row above 760px (wrapping before any field drops below 130px), two columns at or below 760px with Search full width on its own row with Clear.
- Tests: labels and placeholders verbatim, blur formatting, each invalid rule and its placement, server error mapping including min_gt_max, submit payload shape, Clear, Enter submits, disabled while loading, aria-invalid and aria-describedby wiring.

Acceptance:
- Side by side with docs/branding-demo.html the bar is indistinguishable at both breakpoints (allowing for the added Status field).
- Tests green; no oxblood outside §2.2 jobs 1, 3 and 4; no icons, no emoji.

Validation: N/A until BC-12. Record a one-line justification with --role validation and cite the test output.
