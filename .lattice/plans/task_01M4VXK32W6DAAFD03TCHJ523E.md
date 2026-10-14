# BC-10 plan: results surface (ListingCard, Monogram, ScorePill, ListingList, NoResults, ErrorState)

Planner: agent:opus-5.5-planner, 2026-10-14. Worktree `/Users/forrest/Code/brillio-worktrees.noindex/bc-10-results-surface`, branch `feat/BC-10-results-surface` (base master @ e5f49bb; BC-7 and BC-8 have landed).

The contract is the ticket's Scope, Acceptance and Validation blocks, plus:
- SPEC §7.4-7.6 and §8.
- BRANDING §2.2, §3.2, §4.1, §4.4, §5.4-5.6, §5.8, §5.9 and §6-8.
- The reference demo, `docs/branding-demo.html`, lines 44-55, 77-80, 101-106, 137-146, 170 and 182.

All paths are relative to `frontend/`.

## 1. Scope

**In scope:**
- Six component folders under `src/components/`, each holding `X.tsx`, `X.module.css` and `X.test.tsx`.
- A new `src/format/keyword.ts` with its test.
- One appended block in `src/test/fixtures.ts`.

**Out of scope:**
- `App.tsx`, page wiring, aria-busy, scroll reset and state selection: BC-12.
- FilterBar (BC-9); Pager, StatusLine and LoadingLine (BC-11).
- `styles/*`: BC-7 owns it, and BC-14 only fixes it.
- BRANDING.md and the demo are never edited.
- No SPEC edit is needed. No config, Makefile or package.json changes.

## 2. Decisions

**D1. BRANDING text wins; the demo only fills gaps.**
- **Desktop.** A1 matches the demo except for line heights. BRANDING and tokens.css give px values where the demo uses multipliers, and each difference is under 1px.
- **At ≤760px**, the demo contradicts BRANDING:
  - The monogram is 56×42 at 20px. §2.4/§3.2/§5.5 require 72×54 and never below 24px.
  - The address is 20px and the price 24px. §3.2 gives one size for each.
  - The gap is 14px. §4.1 gives one gap.
  - The price block sits after the description. §4.4 says "under the meta line".
- **We build BRANDING:** the same sizes at every width, and the price block between the meta line and the description.

**D2. The card is one flat grid with template areas**, not the demo's three nested columns.
- The mobile reflow can then put the price block under the meta line without duplicating DOM.
- DOM order is monogram slot, address, meta, price block, description. That gives one reading order at both widths, a mobile visual order that matches the DOM, and a desktop layout identical to the demo.

**D3. Keyword emphasis.** BC-12 derives the `keywordTokens: string[]` prop (SPEC §7.4) from the URL `keyword`. `src/format/keyword.ts` exports two functions.

`keywordTokens(keyword: string | undefined): string[]` mirrors the server (SPEC §5.3, `domain.Filter.KeywordTokens`): split on `/\s+/`, then drop empty strings.

`emphasisSegments(text, tokens: readonly string[]): { text: string; match: boolean }[]`:
1. Keep a boolean mask over the UTF-16 units of `text`.
2. For each trimmed, non-blank token, loop `exec` on `new RegExp(escapeRegExp(token), 'gi')`, marking each match and setting `lastIndex = index + 1` so overlapping matches are found.
   - `escapeRegExp` replaces `/[.*+?^${}()|[\]\\]/g` with `'\\$&'`.
   - `RegExp.escape` is not available in the ES2023 lib.
3. Merge the marked runs into segments. Empty text yields `[]`; no tokens yields one plain segment.

The result:
- Every match is found case-insensitively and keeps its original casing.
- Overlapping and adjacent runs merge.
- Regex metacharacters match literally.
- Token order does not matter.

**Rendering.** Each match becomes `<em className={styles.match}>` with `font-style: normal; font-weight: 600` and no color, like the demo's `.desc em`. Not `<mark>`, whose UA background and color would need undoing. Segments render in `<Fragment key={i}>`.

**D4. Monogram initials come from BC-8's `formatMonogram`, unforked.**
- It trims the city name, then uppercases the first character and lowercases the second.
- A multi-word city uses its first two characters: `Falls Church` → `Fa`.
- A one-letter name gives one letter. An empty name gives an empty stone block.
- The block is decorative and the city already appears in the meta line, so the block gets `aria-hidden="true"`.

**D5. Results heading.** ListingList renders a visually hidden `<h2 id={RESULTS_HEADING_ID} tabIndex={-1}>Results</h2>`, with `export const RESULTS_HEADING_ID = 'results-heading'`. The reasons:
- **Hidden:** BRANDING and the demo have no heading above the cards (demo line 136).
- **"Results":** there is no exact copy for it; "Results" is BRANDING's noun and BC-7's region name.
- **h2:** the wordmark is the h1.
- **An id, not a ref:** SPEC §7.4's props stay unchanged, `aria-labelledby` needs the id anyway, and the page has one list.

**BC-12 hookup** (for the completion comment): in Pager `onChange`, after pushing the page, call `document.getElementById(RESULTS_HEADING_ID)?.focus()`. This works for three reasons:
- The previous list stays mounted while loading (SPEC §7.2).
- Focus survives the success render as long as ListingList keeps its tree position.
- The region is `position: relative` inside the scroller, so `focus()` scrolls back to the heading.

BC-12's test: `expect(screen.getByRole('heading', { name: 'Results' })).toHaveFocus()`.

**D6. Live region.**
- ListingList's root is `<section aria-labelledby={RESULTS_HEADING_ID} aria-live="polite">`, as the ticket requires.
- NoResults and ErrorState rely on BC-12's polite container (SPEC §7.6; BC-7's `<main>`). Nesting is harmless, because the innermost region wins.
- Notes for BC-12: aria-busy goes on its container, and it may rename `<main>` so the page does not have two "Results" landmarks.

**D7. Callback props use property signatures** (`onClear: () => void`), not SPEC's method shorthand.
- The types are compatible, so SPEC is unchanged.
- Destructuring a method signature fails `@typescript-eslint/unbound-method` (verified against this repo).

**D8. Spacing.**
- ListingList owns the demo's 6px results top padding, so NoResults sits exactly 96px from the top and BC-12 adds none.
- The gutter stays on the scroller.
- ErrorState reuses NoResults' centered, 96px layout, which matches the demo's shared `.msgbox`.

## 3. Components

General rules:
- Values come from tokens.css as `var(--…)`.
- Literal px only where no token exists; never hex.
- The display face is `var(--font-display)`; the UI face is inherited.
- Headings and `p` elements get `margin: 0` and an explicit weight.

**ScorePill** `({ rank, score }: { rank: number; score: number })`

Markup: `<span class=score><span class=rank>{formatRank(rank)}</span><span>{formatScore(score)}</span></span>`.

| Rule | Declarations |
|---|---|
| `.score` | `inline-flex; align-items: center; gap: 8px`; Meta size and leading; `ink-2`; `tabular-nums`; `nowrap`; no margin |
| `.rank` | `--text-rank`/`--leading-rank`, 500, `--tracking-rank`; ink text with a `1px solid` ink border; `border-radius: 999px; padding: 2px 8px` |

**Monogram** `({ city }: { city: string })`

Markup: `<div class=monogram aria-hidden="true">{formatMonogram(city)}</div>`.

| Rule | Declarations |
|---|---|
| `.monogram` | `width: var(--monogram-width); height: 54px`; `--stone` fill; flex-centered; `--text-monogram`/`--leading-monogram`, 400, `ink-2`. No media query. |

**ListingCard** `({ item, keywordTokens }: { item: ListingResult; keywordTokens: string[] })`

Markup is an `<article class=card>` holding, in this order:
1. `<div class=monogramSlot><Monogram city={item.city}/></div>`
2. `<h3 class=address>{item.address}</h3>`
3. `<p class=meta>`. If `formatStatusLabel(item.status)` is non-null, it starts with `<span class=statusLabel>` and `META_SEPARATOR`. Then comes `formatListingMeta(item).join(META_SEPARATOR)`.
4. `<div class=priceBlock><p class=price>{formatPrice(item.price)}</p><ScorePill rank={item.rank} score={item.relevance.score}/></div>`
5. `<p class=description>`, holding the D3 segments

No link, no button, no `:hover`.

Styles:

| Rule | Declarations |
|---|---|
| `.card` | `grid`; columns `var(--monogram-width) minmax(0, 1fr) auto`; areas `'monogram address price' 'monogram meta price' 'monogram description price'`; rows `auto auto 1fr` (a tall price block grows only the last row); `column-gap: var(--card-column-gap); row-gap: 0; align-items: start; padding-block: var(--card-padding-block)`. Each child sets its `grid-area`. |
| `.address` | display face; `--text-address`/`--leading-address`; 400; `text-wrap: balance` |
| `.meta` | `margin-top: 4px`; Meta size and leading; `ink-2`; `tabular-nums` |
| `.statusLabel` | Label style: `--text-label`/`--leading-label`, 500, `--tracking-label`, uppercase; inherits `ink-2` |
| `.description` | `margin-top: 6px`; `--text-description`/`--leading-description`; `nowrap`; `overflow: hidden`; `text-overflow: ellipsis` |
| `.priceBlock` | `flex; flex-direction: column; align-items: flex-end; gap: 8px; text-align: right` |
| `.price` | display face; `--text-price`/`--leading-price`; 500; `--tracking-price`; `nowrap` |

At `@media (max-width: 760px)`:
- `.card` uses columns `var(--monogram-width) minmax(0, 1fr)`, areas `'monogram address' 'monogram meta' 'monogram price' 'monogram description'`, and rows `auto auto auto 1fr`.
- `.priceBlock` uses `flex-direction: row; flex-wrap: wrap; align-items: baseline; justify-content: flex-start; column-gap: 14px; margin-top: var(--grid); text-align: left`. The 14px comes from the demo.
- Price stays before rank in DOM order. No sizes change.

**ListingList** `({ items, keywordTokens }: { items: ListingResult[]; keywordTokens: string[] })`

The root is `<section class=region aria-labelledby={RESULTS_HEADING_ID} aria-live="polite">`, holding the D5 heading. When `items.length > 0` it also holds an `<ol class=list>` of `<li class=item>` cards. Each `li` is keyed `` `${item.source}:${item.id}` `` because SPEC §4 forbids keying on `id` alone.

| Rule | Declarations |
|---|---|
| `.region` | `position: relative; padding-top: 6px` |
| `.heading` | visually hidden: `absolute`, 1px × 1px, `margin: -1px`, `padding: 0`, `border: 0`, `overflow: hidden`, `clip-path: inset(50%)`, `nowrap`; `:focus { outline: none }`, since oxblood rings are for buttons and links |
| `.list` | `list-style: none; margin: 0; padding: 0` |
| `.item + .item` | `border-top: 1px solid var(--line)` |

Do not add `role="list"`; jsx-a11y `no-redundant-roles` rejects it.

**NoResults** `({ onClear }: { onClear: () => void })` and **ErrorState** `({ onRetry }: { onRetry: () => void })`

Markup: `<div class=root>`, `<h2 class=heading>`, `<p class=body>` (NoResults only), then `<button type="button" class=action onClick>`.

Copy is verbatim from §7.2:
- NoResults: `No results.` / `Adjust your filters.` / `Clear filters`
- ErrorState: `Couldn't load results.` (ASCII apostrophe) / `Try again`, with no error code and no message prop

| Rule | Declarations |
|---|---|
| `.root` | flex column; `align-items: center; gap: 8px; text-align: center; padding-top: 96px` |
| `.heading` | State heading: display face, `--text-address`/`--leading-address`, 400, italic, `ink` |
| `.body` | `--text-body`/`--leading-body`; `ink-2` |
| `.action` | the §5.8 text button, matching BC-9's `.clear`: `margin-top: 6px`; Button size and leading, 500, `--tracking-button`; `color: var(--oxblood)`; no background, border or padding; `cursor: pointer`; `nowrap`. `:hover` adds an underline at 3px offset. `:focus-visible` adds a `2px solid var(--oxblood)` outline at 3px offset. |

Oxblood appears only in `.action` (§2.2 jobs 2 and 3). No motion (§6).

## 4. Files

Create six folders under `src/components/`, each with `.tsx`, `.module.css` and `.test.tsx`:
- `ScorePill/`
- `Monogram/`
- `ListingCard/`
- `ListingList/` (also exports `RESULTS_HEADING_ID`)
- `NoResults/`
- `ErrorState/`

Also create `src/format/keyword.ts` and `keyword.test.ts`, and append to `src/test/fixtures.ts` (§6).

Follow Header's conventions: named exports, an `XProps` interface, a comment citing BRANDING, and `import styles from './X.module.css'`. Tests import from `vitest` and use RTL and user-event. Any suite that fakes Date calls `vi.useRealTimers()` in `afterEach`.

## 5. Tests (exact names)

Meta assertions fake Date with `vi.useFakeTimers({ toFake: ['Date'], now: new Date(2026, 9, 13) })`.

**keyword.test.ts**
- `keywordTokens splits on whitespace like the server`: `'  pets  allowed '` → `['pets','allowed']`; `''` and `undefined` → `[]`
- `returns one plain segment without tokens`
- `returns no segments for empty text`
- `matches case-insensitively and keeps the original casing`: B7 with `['pets']` → `Pets`; A3 with `['PETS']` → `pets`
- `marks every occurrence`: `'pets, pets'` gives two matches
- `handles several tokens in any order`: B7: `['condo','pets']` equals `['pets','condo']`
- `merges overlapping and adjacent matches into one run`: `['pet','pets']` on `Pets allowed` → `Pets`; `['top-fl','floor']` on A1 → `top-floor`
- `treats regex metacharacters literally`: `['.']` on A1 marks only the two periods; `(`, `$` and `\` do not throw
- `ignores blank tokens`: `['', '  ']` gives one plain segment

**ScorePill.test.tsx**
- `renders No. 11 and the score to two decimals`: 11 and 0.8731 → `No. 11` and `0.87` (page 2's first rank; SPEC §5.4's original example score)
- `renders a fixture's rank and score`: A1 → `No. 1` and `0.74`
- `puts the rank pill before the score`

**Monogram.test.tsx**
- `renders the city initials` (it.each): A1 → `Sp`; `Vienna` from `cities` → `Vi`; A3 → `Fa`
- `hides the decorative block from assistive technology`

**ListingCard.test.tsx**
- `renders the address verbatim as a level 3 heading`
- `renders every meta item in BRANDING order`: A1 is exactly `Springfield, VA 22150 · 2 bd · 1.5 ba · 980 sqft · Listed Aug 29 · MLS A` (demo line 137); B7 ends with `Listed Aug 27 · MLS B`
- `shows PENDING first for A7`: the label's parent reads `PENDING · Chantilly, VA 20151 · 4 bd · 2.5 ba · 1,950 sqft · Listed Jul 28 · MLS A`
- `shows no status label for an active listing`: A1 shows no ACTIVE, PENDING or SOLD, and its meta starts `Springfield`
- `emphasizes Pets and pets case-insensitively and leaves the rest untouched`: B7 with `['pets']` has exactly one `em`, `Pets`, and `textContent` equals the raw description; A3 has `em` `pets`
- `renders no emphasis without keyword tokens`
- `renders the price over the score pill`: `$450,000`, `No. 1`, `0.74`
- `renders the city monogram`: `Sp`
- `keeps the price block between the meta line and the description`: `compareDocumentPosition`; stands in for the media query
- `renders no image, icon, link or button`: `querySelector('img, svg, picture, i, a, button')` is null

**ListingList.test.tsx**
- `renders one card per item in rank order inside an ordered list`: `searchResponse.items` → 3 items: A1, A3, A7
- `labels a polite live region with the Results heading`
- `exposes a level 2 Results heading that script can focus`: has the id and `tabindex="-1"`; `toHaveFocus()` after `.focus()`
- `passes keyword tokens to every card`: `[A3, A7]` with `['pets']` → two `em`
- `renders the heading and no list for zero items`
- `keys cards by source and id`: `[A1, { ...A1, source: 'MLS_B' }]` → 2 cards and no `console.error`

**NoResults.test.tsx**
- `renders the no-results copy verbatim`
- `calls onClear once on click`
- `calls onClear from the keyboard`: SPEC §7.6
- `renders a type=button`

**ErrorState.test.tsx**
- `renders the error copy verbatim`
- `calls onRetry once on click`
- `shows no error code`: `textContent` is exactly `Couldn't load results.Try again`

Together these cover SPEC §8's bar for each component: props, interactions, roles and labels, and verbatim copy.

## 6. Shared files (run-state: BC-9/10/11 may only add to `src/format/*` and `src/test/fixtures.ts`)

`src/format/keyword.ts` and `keyword.test.ts` are new files, so they cannot conflict with BC-9 or BC-11. Report them to the orchestrator: BC-12 should use `keywordTokens`.

`src/test/fixtures.ts`: append a `// --- BC-10 additions ---` block with `listingB7: ListingResult`, the real record from `data/sample_listings.json` (`MLS_B`, `123 Main Street, Unit 4B`, Springfield VA 22150, 452000, 2 bd, 1.5 ba, 980, 38.7893/-77.1873, `2026-08-27`, active, `Top floor condo, walk to shopping. Pets allowed.`) with `rank: 2` and `relevance: { score: 0.7297, budgetFit: 0.9911, recency: 0.3376 }`, recomputed per SPEC §6.2-6.3 (target 450000, asOf 2026-10-13; the same run reproduces A1, A3, A7 exactly). B7 is the only real description with a capital `Pets`. The orchestrator union-merges this block.

No BC-8 formatter is changed. The card uses `formatListingMeta`, `formatStatusLabel`, `META_SEPARATOR`, `formatPrice`, `formatRank`, `formatScore` and `formatMonogram` as they are.

## 7. Gate and verification (the implementer runs these; the reviewer re-runs them)

1. Gate: `npx prettier --write` on the touched files, `make check` at the worktree root, then `cd frontend && npm run build`.
2. Greps. Run these in `frontend/` with `C=src/components` and paste the output into the completion comment. They are greps rather than tests because jsdom computes no CSS, and vitest turns `.css?raw` and `?inline` imports into a proxy or `''` (verified).

   | Command | Expected |
   |---|---|
   | `grep -rn oxblood $C/{ListingCard,Monogram,ScorePill,ListingList}` | nothing |
   | `grep -n -B3 oxblood $C/{NoResults,ErrorState}/*.module.css` | only `.action` selectors |
   | `grep -rnE '#[0-9A-Fa-f]{3,8}\b'` over the six `*.css` files | nothing |
   | `grep -n ':hover' $C/ListingCard/*.css` | nothing |
   | `perl -CSD -ne 'print "$ARGV:$.:$_" if /\p{Extended_Pictographic}/; close ARGV if eof' <new files>` | nothing (verified to pass `· ‹ › ▾`) |

3. Demo parity (the acceptance's "pixel-for-pixel in structure and type", with no browser):
   - In the completion comment, add a table mapping each demo selector on lines 44-55 to its module rule and values. Mark every difference D1.
   - A1's text equals demo line 137 except the score: the demo's 0.94 is illustrative, and the real value is 0.74.
   - Pixels are checked at BC-12's browser validation and in BC-14.

## 8. Validation (N/A until BC-12; no ports, no browser)

Run `cd frontend && npx vitest run --reporter=verbose src/components/{ListingCard,Monogram,ScorePill,ListingList,NoResults,ErrorState} src/format/keyword.test.ts > <scratchpad>/bc10-vitest.txt 2>&1`.

Then the delegator runs:
- `lattice comment BC-10 --role validation "N/A: components are not mounted in a page until BC-12; vitest verbose: <files>/<tests> green covering card meta order, PENDING, keyword emphasis, ScorePill, Monogram, ListingList region/heading, NoResults/ErrorState copy and callbacks"`
- `lattice attach BC-10 <file> --role validation`

## 9. Risks

- **No CSS in jsdom.** Tests assert structure, roles and copy instead. CSS is covered by the §7 greps and the parity table, and rendering by BC-12 and BC-14.
- **Raw CSS imports in vitest.** `.css?raw` and `?inline` come back blank (verified). Write no CSS-content tests, and do not add `/// <reference types="node" />`, because it leaks Node globals into the whole `src` typecheck.
- **Lint rules (verified):** `unbound-method` rejects method-signature props, so use D7. jsx-a11y `no-noninteractive-tabindex` flags only `tabIndex >= 0`, so `-1` passes. Never add `role="list"`.
- **Date-dependent meta.** Fake Date, then `useRealTimers` in `afterEach`.
- **Focus on a hidden heading.** The region is `position: relative`, so the heading stays inside the scroller; it also has `outline: none`.
- **Nested live regions and duplicate "Results" landmarks:** covered by the D6 notes to BC-12.
- **Case folding.** SQLite `lower` is ASCII-only, while JS `i` folds Unicode. Only emphasis is affected, and the data is ASCII.
- **fixtures.ts merge:** change only the labelled block at the end.

## 10. Size

About 1,000 LOC: components ~490, `keyword.ts` ~45, fixture ~25, tests ~450. One commit, `feat(frontend): results surface components (BC-10)`, with nothing under `.lattice/`.

## Open questions

- Non-blocking, for the branding owner (BRANDING §9). At 760px and below, the demo contradicts BRANDING in the ways listed in D1 (monogram size, address/price/gap sizes, price block placement). BC-10 follows the BRANDING text. If the demo is what was intended, the two need reconciling through §9, and BC-14 adjusts the CSS.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review: fresh-context Opus plan-review sub-agent (headless review skipped: plan > 250 lines, run-state 03:22Z). Verdict APPROVE-WITH-CHANGES.

| # | Concern | Resolution | Section |
|---|---|---|---|
| R1 | `.action` copied BC-9's `.clear` (+tracking); the Orchestrator's shared text-button recipe (03:25Z mail, BRANDING §5.8 + demo `.link`) is binding for BC-9/10/11 | `.action` in NoResults and ErrorState is EXACTLY: `font: 500 13px/1 var(--font-ui)`; `color: var(--oxblood)`; `background: none; border: 0; padding: 0`; **no letter-spacing**; `white-space: nowrap; cursor: pointer; text-decoration: none`. `:hover` → `text-decoration: underline; text-underline-offset: 3px` (no colour change). `:focus-visible` → `outline: 2px solid var(--oxblood); outline-offset: 3px`. `:disabled` → `color: var(--ink-3)` (never used here; kept for parity). Keep the `margin-top: 6px` spacing on `.action` (layout, not part of the button recipe). Do not use `--tracking-button`. | §3 NoResults/ErrorState |
| R2 | Ticket acceptance asks for a visual check of A1 against `docs/branding-demo.html`; jsdom cannot do that | Implementation unchanged. At validation the validator mounts an **uncommitted** throwaway preview (A1, A7, B7 with `['pets']`, NoResults, ErrorState) on vite port 15110 (run-state 15100+N), opens a c11 browser panel in area:11, screenshots ~1100px and 375px next to the demo, attaches them, removes the preview and kills the server. Pending Orchestrator confirmation (mailed as a non-blocking QUESTION); if refused, the check is deferred to BC-12/BC-14 and recorded. | §7.3, §8 |
| R3 | `.monogram` row omitted the display face | Add `font-family: var(--font-display)` to `.monogram` (Bodoni 26 per BRANDING §5.5). | §3 Monogram |
| R4 | Ticket says "a visually clear heading"; plan's heading is visually hidden | Keep D5 (BRANDING §5.4 and demo line 136 show no heading; BRANDING §3.2 has no type style for one; a visible heading breaks A1 structural parity). Record this interpretation in `## Open questions` and in the completion comment for BC-12/BC-14/human. | D5, Open questions |
| R5 | "B7 is the only real description with a capital `Pets`" is false (A4: "Pets welcome.") | Reason becomes: B7 is demo card 2 and real rank 2, and carries a capital `Pets`. Values unchanged (rank 2, score 0.7297, budgetFit 0.9911, recency 0.3376). | §6 |
| R6 | No `frontend/node_modules` in this worktree; `npx prettier` would fetch an unpinned Prettier | Gate order: `cd frontend && npm ci` first, then `npm run format` (or the repo's pinned prettier via `npx --no-install`), then `make check`, then `npm run build`. | §7.1 |
| R7 | D8's reason was wrong (demo `.msgbox` centers vertically) and "reuses" could tempt a cross-folder CSS import | Reason: BRANDING §5.9 gives 96px for No results and only "Centered" for Error; ErrorState uses the same 96px top offset for consistency. The rules are **copied** into `ErrorState.module.css`; no component imports another folder's CSS module. | D8 |
| R8 | ErrorState tests lacked keyboard activation and `type=button` | Add `calls onRetry from the keyboard` and `renders a type=button` to ErrorState.test.tsx. | §5 ErrorState |
| R9 | 0.8731 attributed to SPEC §5.4 (being corrected by BC-5 to 0.7414) | Keep 0.8731 as the test input; drop the SPEC attribution in the test name/comment. | §5 ScorePill |
| R10 | Helper `keywordTokens` collides with the `keywordTokens` prop name | Rename the helper to `splitKeyword(keyword: string \| undefined): string[]`; tests renamed accordingly (`splitKeyword splits on whitespace like the server`). The prop stays `keywordTokens` (SPEC §7.4). | D3, §5 keyword.test.ts, §6 |
| R11 | `text-transform: uppercase` on `.statusLabel` is redundant (formatter returns `PENDING`) | Skip: keep the BRANDING Label style declarations intact (harmless, keeps the label style uniform); add a one-line CSS comment that the formatter already returns uppercase. | §3 ListingCard |

Additional note for the completion comment: BRANDING §3.2/§3.3 mention +0.02em on buttons while the Orchestrator's recipe and the demo `.link` use none; the recipe is binding here — flag the discrepancy to the Orchestrator for BC-14.

Open questions (amended): (1) the ≤760px demo-vs-BRANDING discrepancy (D1, unchanged, non-blocking); (2) "visually clear heading" interpreted as a visually hidden, focusable `h2` "Results" (R4, non-blocking).
