# BC-10: Results surface: ListingCard, Monogram, ScorePill, ListingList, NoResults, ErrorState

The results surface: ListingCard with Monogram and ScorePill, ListingList, and the NoResults and ErrorState screens. Spec: docs/SPEC.md §7.4, §7.5, §7.6. Design: BRANDING.md §5.4 (card), §5.5 (monogram), §5.6 (score), §5.9 (no results, error), §3.2 (Price, Address, Meta, Description, State heading), §4.1, §4.4, §7.1, §7.2. Reference: docs/branding-demo.html.

Why: the handout requires address, price, bedrooms and relevance score per result; BRANDING fixes exactly how. Keeping the card a pure renderer over ListingResult lets it be tested against the real sample data.

Scope (props per SPEC §7.4):
- ListingCard: three columns (monogram 72px, text, price block right-aligned); hairline above every card but the first (handled by ListingList). Text column: address in Bodoni 22 with `text-wrap: balance`, verbatim; meta line in Figtree 13 ink-2 with ` · ` separators in this order: status label only when not active, `City, ST ZIP`, bd, ba, sqft, `Listed Mon D`, source; description one line with ellipsis, each case-insensitive keyword match wrapped in weight 600 (never color). Price block: price in Bodoni 28/500 right-aligned over ScorePill. At or below 760px the price block moves under the meta line, left-aligned, price before rank (§4.4). No hover state, no link.
- Monogram: 72×54 stone block, city initials (first upper, second lower) in Bodoni 26 ink-2, centered; never below 24px.
- ScorePill: `No. N` pill (1px ink border, radius 999px, Figtree 11/500 +0.04em, padding 2px 8px) followed by the score to two decimals in Figtree 13 ink-2 tabular.
- ListingList: renders cards separated by hairlines inside a region with a visually clear heading for focus management (BC-12 moves focus here), `aria-live="polite"`.
- NoResults: centered 96px from the top: `No results.` (State heading, italic Bodoni 22), `Adjust your filters.` (Body ink-2), `Clear filters` text button calling onClear.
- ErrorState: centered `Couldn't load results.` (State heading) and a `Try again` text button calling onRetry. No error code shown.
- Tests using fixtures from BC-8 built on the real listings: every meta item renders in order; PENDING shows for A7 and nothing for active; keyword emphasis wraps `Pets`/`pets` case-insensitively and leaves the rest untouched; ScorePill text `No. 11` and `0.87`; Monogram `Sp`, `Vi`, `Fa`; NoResults and ErrorState strings verbatim and callbacks fire; no `<img>`, no icons.

Acceptance:
- A card rendered from A1 matches the demo card pixel-for-pixel in structure and type (visual check against docs/branding-demo.html).
- Tests green; lint green; no emoji; no oxblood anywhere in these components except the two text buttons.

Validation: N/A until BC-12. Record a one-line justification with --role validation and cite the test output.
