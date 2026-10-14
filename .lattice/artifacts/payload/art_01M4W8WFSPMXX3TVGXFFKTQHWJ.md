Lattice-Reviewed-Commit: fa7c08dfe01e3d52922a57765c5b2a269ca34e80
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-10-results-surface
Lattice-Reviewed-Base: master (0e6730fcfbd67c41afecd8c35bc346f22da4c6d9)
Lattice-Reviewed-Head: feat/BC-10-results-surface (fa7c08dfe01e3d52922a57765c5b2a269ca34e80)

# Code Review: BC-10 — results surface (ListingCard, Monogram, ScorePill, ListingList, NoResults, ErrorState)

Reviewed commit: `fa7c08d` on `feat/BC-10-results-surface` (base `master` @ `0e6730f`).
Reviewer: fresh-context Claude Fable 5.1 (agent:fable-reviewer). Read cold against the plan (including the Cycle 1 resolutions R1–R11), SPEC §7.4–7.6, BRANDING §2.2/§3.2/§3.3/§4.1/§4.4/§5.4–5.9/§7–8, and `docs/branding-demo.html`.

## 1. Verdict

**PASS**

No implementation- or plan-level rework is needed. The findings below are minor and informational; the only action item is a missing Lattice completion comment, which the orchestrator can record before `in_validation`.

## 2. Summary

The diff adds six component folders (`.tsx` + `.module.css` + `.test.tsx`), `src/format/keyword.ts` with its test, and one appended fixture block, 1,026 lines total and nothing outside `frontend/`. Everything in the plan's §3 tables, §5 test list, and all eleven Cycle 1 resolutions is present and correct; the full gate is green (`make check`, `npm run build`, 41/41 targeted tests, 269/269 overall), the §7.2 greps are clean, and I independently recomputed the 12-listing ranking from `data/sample_listings.json` per SPEC §6.2–6.3 and it reproduces the B7 fixture (rank 2, 0.7297 / 0.9911 / 0.3376) and A1/A3/A7 exactly. The one gap is procedural: the plan's required completion comment (grep output, demo parity table, BC-12 hookup notes) was never recorded, so I include the parity table and gate evidence in this review.

## 3. Issues

**[MINOR] Lattice task BC-10 (no file) — Required completion comment was not recorded**
Plan §7.2 says to paste the gate greps into the completion comment, §7.3 requires the demo-selector parity table there, D5 puts the BC-12 hookup instructions there, and R4 says to record the hidden-heading interpretation there. The task moved `in_progress -> review` at 03:48:52Z with only the three pre-implementation triage comments; I checked both the worktree's and the main checkout's `.lattice/events/task_01M4VXK32W6DAAFD03TCHJ523E.jsonl`. BC-12's planner will not find `RESULTS_HEADING_ID`, `splitKeyword`, the aria-busy placement, or the duplicate-landmark warning without it.
**Fix:** Before `in_validation`, `lattice comment BC-10 "..."` with: (a) the grep output (reproduced in the appendix below, all clean); (b) the parity table (appendix); (c) BC-12 hookup: call `document.getElementById(RESULTS_HEADING_ID)?.focus()` after pushing the page, use `splitKeyword(urlState.keyword)` for the `keywordTokens` prop, put `aria-busy` on the container, and keep exactly one live region / one "Results" landmark (see next finding); (d) the R4 interpretation (visually hidden focusable `h2`) and the D1 ≤760px open question.

**[MINOR] frontend/src/components/ListingList/ListingList.tsx:21 — Nested polite live regions and duplicate "Results" landmark once composed**
`ListingList` renders `<section aria-labelledby="results-heading" aria-live="polite">`; BC-7's `App.tsx:12` still has `<main aria-label="Results" aria-live="polite">`. When BC-12 composes them, there are two "Results" landmarks and a polite live region nested in another. The plan's "innermost wins" (D6) is not guaranteed across assistive technology; NVDA and JAWS can announce the swapped list twice. This is exactly what the ticket asked for on `ListingList` and the plan defers it to BC-12, so nothing changes here — but it must be in the handoff.
**Fix:** No change in BC-10. BC-12 should keep one live region in the results area: either drop `aria-live` from `<main>` (and rename or unlabel it so there is one "Results" landmark), or keep the live region on the container that also hosts `NoResults`/`ErrorState` and drop it from `ListingList`. Record the choice in BC-12's plan.

**[MINOR] frontend/src/components/NoResults/NoResults.module.css:35 and frontend/src/components/ErrorState/ErrorState.module.css:27 — Text-button recipe differs in form from the merged Pager's**
`.action` uses `font: 500 13px/1 var(--font-ui)` with literal px, as R1 mandated ("EXACTLY"). The already-merged Pager (BC-11) expresses the same recipe as `font-size: var(--text-button); line-height: var(--leading-button); font-weight: 500` and inherits the family via `global.css`'s `button { font: inherit }`. Computed values are identical (13px/1, Figtree, 500), so there is no visual effect, but the repo now carries two textual variants of "the shared recipe" and the plan's own §3 rule says values come from tokens where a token exists (`--text-button`, `--leading-button` both exist). Not a BC-10 defect; R1 is authoritative.
**Fix:** None here. Note for BC-14: pick one form (tokens, matching Pager) or hoist a shared text-button class; also resolve the BRANDING §3.2/§3.3 "+0.02em on buttons" vs §5.8 text-button ambiguity in BRANDING itself (the Orchestrator's reading, primary only, is the one implemented in all three tickets).

**[MINOR] frontend/src/components/ListingList/ListingList.test.tsx:7 — Test name overstates what is asserted**
`renders one card per item in rank order inside an ordered list` asserts that input order is preserved; `ListingList` does not sort, and the fixture happens to be in rank order. The behaviour is right (the API guarantees order), the name just implies a sort.
**Fix:** Rename to `renders one card per item in the given order inside an ordered list`, or add a comment that order is the API's. Nit; fine to skip.

**[MINOR] frontend/src/components/Monogram/Monogram.module.css:8 — `height: 54px` literal next to a tokenised width**
No height token exists, so the plan permits the literal. `calc(var(--monogram-width) * 3 / 4)` would tie the §5.5 4:3 ratio to the width token so a future width change cannot drift the ratio.
**Fix:** Optional; skip or fold into BC-14.

**[INFO] frontend/src/components/ListingCard/ListingCard.tsx:47 — `<em>` for keyword hits**
`<em>` carries stress-emphasis semantics and some screen readers voice it; `<mark>` is the element for search-hit highlighting. The plan chose `<em>` to match the demo's `.desc em` and avoid undoing UA `mark` styling, and the weight-600/no-colour rule is honoured either way. Mention only; no change requested.

**[INFO] frontend/src/format/keyword.ts:8 — `splitKeyword` matches the documented contract, not live Go code**
SPEC §5.3 ("whitespace-split into tokens") and `backend/internal/domain/filter.go:19` ("split on whitespace and lowercased") are what the client mirrors. No Go code currently populates `KeywordTokens` (only the store consumes it), so there is no implementation to diff against yet. Lowercasing is irrelevant to emphasis, which is case-insensitive by regex flag. Keyword is capped at 200 chars server-side, so token count is bounded. Nothing to do.

## 4. Positive Observations

- **`emphasisSegments` is small, correct, and property-tested.** A boolean mask over code units with `lastIndex = index + 1` restart gives overlapping, adjacent, multi-token, order-independent, case-preserving matches in ~20 lines; metacharacters are escaped. The tests assert the join-invariant (segments concatenate back to the input) on real fixture descriptions, including the `top-fl`/`floor` overlap, `aaa`/`aa`, every metacharacter, and the `P.ts` non-wildcard case.
- **The card is one flat grid with template areas.** The ≤760px reflow is grid-areas only, DOM order equals reading order at both widths, and `keeps the price block between the meta line and the description` is an honest jsdom stand-in for the media query via `compareDocumentPosition`.
- **Keys are `${source}:${id}`** with a test that spies `console.error` on a same-id/different-source pair, so a regression to `key={item.id}` fails loudly.
- **Callbacks are invoked with no arguments** (`onClick={() => onRetry()}`), asserted with `toHaveBeenCalledWith()`, so a `MouseEvent` can never leak into `onClear`/`onRetry`.
- **Fake timers are scoped to `Date` only and restored in `afterEach`;** suites that do not pin the clock (ListingList) use regexes like `/ · MLS A$/` that survive the year suffix.
- **Fixture B7 is real and the numbers are right.** Independent recomputation from `sample_listings.json` with target 450,000 as of 2026-10-13 reproduces B7's rank 2 and score/budgetFit/recency to four decimals, and A1/A3/A7 unchanged.
- **Every Cycle 1 resolution landed:** R1 recipe, R3 `font-family: var(--font-display)` on `.monogram`, R7 copied (not imported) CSS in ErrorState, R8 ErrorState keyboard and `type=button` tests, R9 no SPEC attribution on 0.8731, R10 `splitKeyword`, R11 comment on the redundant `text-transform`.
- **BRANDING §2.2 and §8 hold:** oxblood appears only in the two `.action` rules (text + focus ring), no hex, no `:hover` on the card, no emoji, no `<img>`/`<svg>`/`<i>`/`<a>`/`<button>` in the card, uppercase only in the Label style, italic only on the two state headings, tabular figures on meta and score.
- **Comments cite the governing section** on every rule and component, matching Header/Pager/StatusLine conventions; named exports, `XProps` interfaces, property-signature callbacks.

---

## Appendix A — Gate evidence (re-run by the reviewer, 2026-10-13 23:50 local)

| Step | Result |
|---|---|
| `cd frontend && npm ci` | exit 0 |
| `make check` (eslint `--max-warnings 0`, `tsc -b --noEmit`, `prettier --check`, `go test -race ./...`, `vitest run`) | exit 0 — 20 files / 269 tests passed |
| `cd frontend && npm run build` | exit 0 — 20 modules, `dist/assets/index-*.css` 2.56 kB |
| `npx vitest run --reporter=verbose` on the six component folders + `keyword.test.ts` | 7 files / 41 tests passed |

§7.2 greps (run in `frontend/`, `C=src/components`):

| Command | Output |
|---|---|
| `grep -rn oxblood $C/{ListingCard,Monogram,ScorePill,ListingList}` | nothing |
| `grep -n -B3 oxblood $C/{NoResults,ErrorState}/*.module.css` | only `.action` (lines 36/51 NoResults, 28/43 ErrorState) and `.action:focus-visible` |
| `grep -rnE '#[0-9A-Fa-f]{3,8}\b'` over the six `*.css` | nothing |
| `grep -n ':hover' $C/ListingCard/*.css` | nothing |
| `perl -CSD -ne 'print if /\p{Extended_Pictographic}/'` over all new files | nothing |

## Appendix B — Demo parity (docs/branding-demo.html lines 44–55 → module rules), desktop

| Demo selector | Module rule | Match | Note |
|---|---|---|---|
| `.card{grid; 72px minmax(0,1fr) auto; gap:20px; align-items:start; padding:20px 0; border-top:1px line}` | `ListingCard .card`: grid, `var(--monogram-width) minmax(0,1fr) auto`, `column-gap: var(--card-column-gap)`, `row-gap: 0`, `align-items: start`, `padding-block: var(--card-padding-block)`; hairline on `ListingList .item + .item` | ✓ | demo's row-gap is moot with one row |
| `.card:first-child{border-top:0}` | `.item + .item` | ✓ | |
| `.mono{72×54; stone; flex-centre; display face; 26px; ink-2; lh 1}` | `Monogram .monogram` | ✓ | |
| `.addr{display; 22px; lh 1.2; 400; ls 0; text-wrap:balance}` | `.address` 22/26px, 400, balance | ✓ | 26 vs 26.4px — D1 |
| `.meta{13px; ink-2; mt 4px; tabular}` | `.meta` 13/18, ink-2, `margin: 4px 0 0`, tabular | ✓ | 18 vs inherited 18.85px — D1 |
| `.meta .st{ls .08em; uppercase; 11px; 500}` | `.statusLabel` 11/14, 500, `--tracking-label`, uppercase | ✓ | |
| `.desc{14px; mt 6px; nowrap; hidden; ellipsis}` | `.description` 14/20, `margin: 6px 0 0`, nowrap/hidden/ellipsis | ✓ | |
| `.desc em{font-style:normal; 600}` | `.match` | ✓ | |
| `.pricecol{text-align:right}` + `.score{flex; flex-end; centre; gap 8; mt 8}` | `.priceBlock` flex column, `align-items: flex-end`, `gap: 8px`, right; `ScorePill .score` inline-flex, centre, gap 8 | ✓ | 8px top gap moves from `.score` margin to `.priceBlock` gap |
| `.price{display; 28px; lh 1.1; 500; ls -0.01em; nowrap}` | `.price` 28/31, 500, `--tracking-price`, nowrap | ✓ | 31 vs 30.8px — D1 |
| `.score{13px; ink-2; tabular}` | `.score` 13/18, ink-2, tabular | ✓ | |
| `.rank{11px; 500; ls .04em; ink; 1px ink border; 999px; 2px 8px; lh 1.3; nowrap}` | `.rank` | ✓ | identical |

A1 text equals demo line 137 except the score (demo 0.94 illustrative; real 0.74), as the plan states. At ≤760px the implementation follows BRANDING over the demo (72×54 monogram, 22/28px type, 20px gap, price block under the meta line) — plan D1, open question for the branding owner, unchanged by this review.
