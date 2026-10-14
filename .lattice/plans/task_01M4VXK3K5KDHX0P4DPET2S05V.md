# BC-14 plan: branding conformance pass (BRANDING.md + docs/branding-demo.html)

Planner: agent:opus-5.5-planner, cold read of HEAD bb78854 (= feat/BC-12; BC-12 has since landed on master as b514f8e, same tree). Complexity: low-medium.

## 1. Scope and approach
Static audit of every BRANDING.md section and every ticket checklist item against the frontend source (CSS values, markup, strings), with the demo's CSS as the numeric reference. Six small fixes (D1-D6) in frontend CSS/markup/a11y attributes plus one SPEC sentence. Everything that would change BRANDING.md, or where BRANDING and the demo disagree, goes to the owner (O1-O14) and is not edited here. A validator then proves the result in a real browser at 1280 and 375 wide, with a keyboard-only pass. Never edit BRANDING.md or docs/branding-demo.html.

## 2. Audit table
Paths are relative to `frontend/src/` unless they start with `docs/`, `backend/` or `index.html`. "runtime" means the validator confirms it in the browser (section 8).

| # | Item | Finding from the code | Expected |
|---|---|---|---|
| A0 | §0 What was decided | Light only, Bodoni + Figtree, oxblood, list cards, city monogram, rank + decimal, hairlines, the name "Brillio Listings" (index.html:13, Header.tsx:13), and the pinned shell. Each pick is realised; the rows below give the evidence. | pass |
| A1 | §1 Principles | Serif only on the address, price, state headings and monogram (ListingCard.module.css:29,89; NoResults/ErrorState `.heading`; Monogram.module.css:10). Everything else inherits Figtree (styles/global.css:29). | pass |
| A2 | §2.1 Tokens | styles/tokens.css:8-13 matches §2.1 byte for byte, including `color-scheme`. No other hex values and no inline `style=` anywhere in src (grep). | pass |
| A3 | §2.2 Five oxblood jobs | (1) Search fill, FilterBar.module.css:149. (2) Clear, Clear filters, Try again, pager buttons: FilterBar:185, NoResults:36, ErrorState:28, Pager:45. (3) Focus underline and rings: FilterBar:75,165,200, Pager:56, NoResults:51, ErrorState:43. (4) Invalid underline and message: FilterBar:75,120. (5) Current page: Pager:71-75. Oxblood also appears on the loading segment (LoadingLine.module.css:26). §5.9 and §6 require it there, but §2.2 does not list it (O6). The comment at LoadingLine.module.css:20 wrongly calls it a §2.2 job (D5). | pass + D5 (O6), runtime |
| A4 | §2.3 Semantic color | No green, amber or red. The line for unmapped 400 errors is in ink (FilterBar.module.css:211-217). | pass |
| A5 | §2.4 Contrast | Only pairs from the table are used. Placeholders are ink-2 (FilterBar:67-70). The monogram stays 26px at every width (Monogram:11, no breakpoint rule). The demo drops it to 20px at ≤760 (demo:102), against §2.4 (O8). | pass |
| A6 | §2.5 Dark mode | `color-scheme: light` (tokens.css:12, global.css:21-23). No `prefers-color-scheme` rule exists. | pass |
| A7 | §3.1 Faces | The fonts URL is identical to §3.1 (index.html:9-12). Font stacks at tokens.css:17-18. `font-optical-sizing: auto` (global.css:32). Weights used: Bodoni 400, 500 and 400 italic; Figtree 400, 500, 600. The default bold of h1/h2/h3 is overridden everywhere it shows. | pass |
| A8 | §3.2 Scale | Every role matches tokens.css:21-47 and the CSS that uses it. Price 28/31/500/-0.01em (ListingCard:87-95). Address 22/26 with balance (:26-34). State heading 22/26, 400 italic. Monogram 26/1. Body 15/22 (FilterBar `.control` :59-60, NoResults `.body`). Description 14/20 (:59-68). Meta 13/18 (:37-45). Status, pager, score 13/18. Button 13/1/500/+0.02em (FilterBar:151-154). Wordmark and label (Header.module.css; FilterBar:32-42). Rank 11/1.3/500/+0.04em (ScorePill:18-28). Text buttons carry no tracking (BC-19, O10). Non-current page numbers are 500 (O11). | pass (O10, O11) |
| A9 | §3.3 Rules | Bodoni's smallest use is 22px. tabular-nums on meta, score, pager, status and inputs (ListingCard:44, ScorePill:12, Pager:16, StatusLine:14, FilterBar:62); prices stay proportional. Uppercase only on labels, the wordmark and the scope and status labels. Italic only on the two state headings (the match `<em>` is normal, ListingCard:72). Rank and price tracking break "tracking only where uppercase is" (O10). | pass (O10) |
| A10 | §4.1 Grid | Gutter 24, and 16 at ≤760 (tokens.css:53,72-76). Content width 1040, centred through `--inline-pad` (:60). Card padding 20, column gap 20, monogram column 72 (:61-63). Label gap 4 (:52). | pass |
| A11 | §4.2 App shell | Rows `auto auto auto minmax(0,1fr) auto` (styles/shell.module.css:12). html, body and #root at 100% with overflow hidden (global.css:10-19). Results scroll with `overflow-y: auto` and `overscroll-behavior: contain` (shell:17-22). Header 56px and pager 48px including their hairlines. There is no keyboard path to scroll the results (D1). Short viewports: O14. | pass + D1, runtime |
| A12 | §4.3 Radius, elevation | Radius 0 on controls (global.css:44; FilterBar:55,148,183). The only 999px radius is ScorePill:25. No shadow, blur, gradient or filter (grep). Phones flash a translucent, rounded tap highlight by default (D6). | pass + D6, runtime |
| A13 | §4.4 Breakpoints | ≤760: the bar has two columns, the actions get their own row and Search is flex:1 (FilterBar:220-236, as demo:98-100). The card has two columns with the price block under the meta line, left-aligned, price before rank (ListingCard:102-122). >760: the seventh field pushes the actions to row two at 1040 (FilterBar:15, the §5.2 wrap rule), not §4.4's one row (O1). Search + Clear (about 135px) may overflow their 131px column by a few px: measure only. | pass (O1), runtime |
| A14 | §4.5 Alignment | Everything is left-aligned except the price block (ListingCard:77-84) and the centred page numbers (Pager:19-23). Labels sit above values. | pass |
| A15 | §5.1 Header | 56px with a hairline. Wordmark h1 plus the scope label (Header.tsx:12-15). | pass |
| A16 | §5.2 Filter bar | Fields, placeholders and the format on blur (FilterBar/filterBarModel.ts:21-29,147-155). Focus is a 2px underline with no ring (FilterBar.module.css:73-77). Invalid: both fields marked, message under the second (filterBarModel.ts:319-351). Disabled is ink-3 on a `line` underline (:79-86). The ▾ glyph is ink-2 at right (:99-109). Enter submits. Search is disabled while loading or invalid (FilterBar.tsx:112,287). The Status field is O1. | pass (O1) |
| A17 | §5.3 Status line | format/listing.ts:73-95 and StatusLine.tsx:12-44 give results / newest first / Loading / 0 results / empty on error, and `Page N of M`, empty with no pages. | pass |
| A18 | §5.4 Listing card | Order and hairlines (ListingCard.tsx:25-50; ListingList.module.css:34-36). No hover rule. Address verbatim. Meta order (format/listing.ts:45-54). Emphasis by weight only. At 375 a meta item splits across lines (`980 / sqft`, D3). | pass + D3 |
| A19 | §5.5 Monogram | 72×54 stone block, Bodoni 26 in ink-2, centred, aria-hidden (Monogram.*; format/listing.ts:67-70). | pass |
| A20 | §5.6 Score | Pill: 1px ink border, 999px radius, 11/500, padding 2px 8px. Score to 2 decimals, 13px ink-2 tabular, right-aligned under the price (ScorePill.*). The rank comes from the API (O3). | pass (O2, O3) |
| A21 | §5.7 Pager | 48px with a hairline. `‹ Previous` and `Next ›` text buttons with aria-hidden glyphs. Page numbers centred with 16px gaps. The current page is oxblood, 500, underlined through `[aria-current]`. Disabled ends use the `disabled` attribute and ink-3 (Pager.tsx:26-65; Pager.module.css). | pass (O11) |
| A22 | §5.8 Buttons | The primary button matches demo `.btn` (FilterBar:144-173). The four text-button implementations compute the same values (section 6). NoResults and ErrorState write the recipe as a `font` shorthand (D4). | pass + D4 |
| A23 | §5.9 States | Loading: `Loading` status, LoadingLine at 40% width, 1.2s linear; Search disabled; the old list stays (pages/SearchPage.tsx:21-33,102,107). No results: centred, 96px from the top, verbatim copy. Invalid: strings verbatim (filterBarModel.ts:56-61; backend/internal/httpapi/query.go:93-94). Error: heading and Try again, status empty, console only (SearchPage.tsx:87-88,109). Loading line on the first load: O7. Nested live regions: D2. | pass + D2 (O7) |
| A24 | §5.10 Icons | The only glyphs are ‹ and › (Pager.tsx:35,63) and ▾ (FilterBar.module.css:100-101), all set as text. The favicon is empty (index.html:6). | pass |
| A25 | §6 Motion | The only `animation` in src is the loading line; there is no `transition` anywhere (grep). Under reduced motion the line is static and full width (LoadingLine.module.css:41-46). | pass, runtime |
| A26 | §7 Voice | Visible copy is sentence case, with no exclamation marks, "please" or apologies. ARIA names (Filters, Results, Pagination, Page N, Loading) are plain nouns. | pass |
| A27 | §7.1 Formats | format/money.ts, listing.ts and date.ts cover every row, each with unit tests. 2.0 baths shows as `2 ba`. | pass |
| A28 | §7.2 Strings | Labels, placeholders and buttons (filterBarModel.ts:21-41; NoResults.tsx:13-16; ErrorState.tsx:13-15; Pager.tsx:35,63), status strings and the three invalid strings are all verbatim. Strings beyond §7.2: O13. | pass (O13) |
| A29 | §8 Don'ts | Each bullet maps to a row above. No spinners, skeletons, toasts, modals, zebra striping or relative dates. | pass |
| A30 | §9 Change process | This ticket edits neither BRANDING.md nor the demo. The O items go to the owner through §9. | n/a |
| A31 | §10 Open decisions | The build matches each stated default: pinned shell, Figtree wordmark, both sources shown and labelled, inert cards, newest first with no budget. | pass (O4) |
| C1 | Oxblood only in the five jobs | Same as A3. | pass + D5 (O6), runtime |
| C2 | No radius except the rank pill | Same as A12. | pass + D6, runtime |
| C3 | No shadows, gradients, translucency, icons, images | No `<img>`, `<svg>` or `<canvas>`, and no background-image (grep). A12, A24. | pass + D6, runtime |
| C4 | Bodoni never below 18px | Smallest is 22px (A8). | pass, runtime |
| C5 | Tabular figures where digits sit in columns | A9. | pass |
| C6 | Uppercase only in labels and the wordmark | A9. | pass, runtime |
| C7 | Every §7.2 string verbatim | A28. | pass |
| C8 | Every §7.1 format | A27. | pass |
| C9 | The document never scrolls | A11. | runtime |
| C10 | Pinned header, filter bar, status line, pager | A11. | runtime |
| C11 | Loading line is the only animation; static under reduced motion | A25. | runtime |
| C12 | Disabled looks use ink-3 and stone, not opacity | FilterBar:79-86,169-173,205-208; Pager:61-64. The only `opacity` is the placeholder reset to 1 (FilterBar:69). | pass |
| C13 | Keyboard-only pass | Search and pagination work by keyboard, but the only scroller cannot be reached (D1). | D1 + runtime |
| C14 | Emoji grep | Run on tracked files minus the two PDFs: no hits outside `.lattice/`, and 724 × U+2713 (the vitest pass mark) in 9 tracked `.lattice/` logs (S8). | pass for shipped files (S8) |

## 3. Deviations to fix
Implementer step 0: `git merge master`, never rebase (it brings only the BC-12 landing and the board sync; the frontend tree is unchanged).

- **D1. The results region is reachable and scrollable from the keyboard** (SPEC §7.6 "keyboard-only use must work end to end"; BRANDING §4.2 makes it the only scroller). Files: pages/SearchPage.tsx:110, styles/shell.module.css, pages/SearchPage.test.tsx, docs/SPEC.md §7.6. Now: `<main>` cannot take focus and the cards hold nothing focusable, so in WebKit (no auto-focusable scrollers) Tab goes from Clear straight to the pager, and page 1 cannot be scrolled past the cards on screen. Change: add `tabIndex={0}` to `<main>` with `// eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- the results scroller must be keyboard-scrollable (SPEC section 7.6)` on the line above the attribute (attributes one per line). In shell.module.css add `.results:focus-visible { outline: 2px solid var(--oxblood); outline-offset: -2px; }`; the outline is inset because the region is full-bleed (this is the job-3 ring, O12). Tests: T0 asserts `tabindex="0"`; a new case focuses Clear, presses `user.tab()` and expects the results region to have focus. SPEC edit: §7.6 gains "The results region is focusable (tabindex 0) so the keyboard can scroll it." Record it with a `lattice comment` per SPEC's deviation rule.
- **D2. One live region for the results** (BC-12 handover b). ListingList.tsx:21 drops `aria-live="polite"`; the comment at :16-18 says SearchPage's `<main>` is the live region (SPEC §7.6), so the section is not one too (nested regions can announce twice). In ListingList.test.tsx:25-31, rename the test and assert `not.toHaveAttribute('aria-live')`. SearchPage T0 keeps asserting `polite` on `<main>`.
- **D3. Meta items never split** (BC-10 handover 2; at 375 the card shows `980 / sqft`; BRANDING says nothing; typesetting only). In ListingCard.tsx:30-38, add `const NBSP = ' '`, `const ITEM_SEPARATOR = META_SEPARATOR.replace(' ', NBSP)` (that is `NBSP·space`, so a dot stays with the item before it) and `keepWhole = (t: string) => t.replaceAll(' ', NBSP)`. Render `formatListingMeta(item).map(keepWhole).join(ITEM_SEPARATOR)`, and use ITEM_SEPARATOR after the status label. The shared formatters and the status line do not change. Tests: getByText already treats NBSP as whitespace, so most matchers stand. ListingCard.test.tsx:45-46 and :54-55 compare raw textContent: compare `textContent?.replaceAll(' ', ' ')`. Add "keeps each meta item whole": A1's meta textContent equals the six items with NBSPs, joined by `' · '`.
- **D4. One textual form for the text-button recipe** (note 2; BC-10 handover 3). In NoResults.module.css:35 and ErrorState.module.css:27, replace `font: 500 13px/1 var(--font-ui);` with `font-size: var(--text-button); line-height: var(--leading-button); font-weight: 500;`; the family already arrives through global.css:41. Add `cursor: default` to `.action:disabled`, as FilterBar and Pager have. No computed change. Guard: the validator's computed-style comparison (V5).
- **D5. The loading-line comment** (LoadingLine.module.css:20): "Section 2.2: the loading segment is one of oxblood's jobs" becomes "Sections 5.9 and 6 make the loading segment oxblood; section 2.2's five jobs do not list it (owner question, BC-14)". Comment only.
- **D6. No translucent tap flash** (§4.3 "no translucency", §8). Add `-webkit-tap-highlight-color: transparent;` to the `html` block at global.css:21-23, with a one-line comment. Mobile Safari and Chrome on Android otherwise flash a translucent, rounded grey box on every tapped button. Guard: the computed value (V4j).

## 4. Needs branding owner (the delegator raises them in one `lattice needs-human`; none blocks D1-D6)
- **O1** §5.2 field table, §4.4 ">760" row, §7.2 (SPEC §13.1): add the Status select row (Any / Active / Pending / Sold). With seven fields the actions wrap to a second row at 1040px, so §4.4's "one row" no longer holds.
- **O2** §5.6 (SPEC §13.2): `relevanceScore` should read `relevance.score`.
- **O3** §5.6 (SPEC §13.3): the rank is the API's `rank`, not computed by the UI. Wording only.
- **O4** §5.3 and the last bullet of §10 (SPEC §13.4): with no budget the order is newest first by construction. Close the open decision.
- **O5** §7.2 "Invalid, price" (SPEC §13.5): the API allows min == max, and the build flags only min > max with the verbatim string. Suggested wording: "Min price must not exceed max price."
- **O6** §2.2 vs §5.9 and §6: the loading segment is oxblood but is not one of the "exactly five jobs". Add a sixth job or recolour it.
- **O7** §5.9 Loading, last sentence (BC-12 handover c): does "First load with no prior results: the status line alone" mean there is no loading line on the first load? The build shows the line on every load (SPEC §7.5). If the owner says no line, it is one line at SearchPage.tsx:107 (`active={loading && shown !== undefined}`), plus SearchPage T0/T1 and a §7.5 SPEC edit.
- **O8** §2.4, §3.2, §4.1, §4.4 and §5.5 vs the demo's ≤760 CSS (demo:101-106; BC-10 handover 5). The demo shrinks the monogram to 56×42 at 20px (below the 24px floor), cuts the card gap to 14px, sets the price at 24px and the address at 20px, and puts the price block after the description. The build follows BRANDING's text. Fix the demo or BRANDING via §9.
- **O9** The demo file (§9; BC-10 handover 1): docs/branding-demo.html has no `<!doctype html>` and no `<meta charset>`. Opened from disk it renders in quirks mode with a sniffed encoding, so `·`, `‹`, `▾` and `−` are at risk.
- **O10** §3.2 Button, §3.3 and §2.1 vs §5.8 and demo `.link`: text-button tracking and hover colour, already BC-19. Also §3.3 "tracking only where uppercase is" vs the +0.04em rank pill and the -0.01em price. Fold this into BC-19.
- **O11** §3.2 (pager at 400) and §5.7 ("current … weight 500") vs §5.8 and demo `.link` (every page number at 500): which weight do non-current page numbers take? Also, a hovered page number (3px underline) looks the same as the current page.
- **O12** §2.2 job 3: D1 makes the results region focusable and gives it the job-3 ring (2px oxblood, inset). Confirm this, or name another focus look.
- **O13** §7.2 and §5.9 Invalid: copy the build shows that BRANDING lacks. The API's other field messages ("… must be a number.", "… must be 0 or more.", "Min bedrooms must be a whole number.", "Status must be active, pending or sold.", "… must be N characters or fewer."). Also the ink notice line under the bar for errors no field can show ("Page must be 1 or more.", "Page size must be between 1 and 100.").
- **O14** §4.2 and §4.4: everything except the results is pinned at every height. At 375 wide the chrome is about 504px (header 56 + seven-field bar about 366 + line 2 + status 32 + pager 48), which leaves room for about one card on a 375×667 phone. In phone landscape (667×375) the pager is clipped and cannot be reached, because the document never scrolls. The validator measures this. Question: should the bar unpin at ≤760?

## 5. Skipped / deferred (no new tickets)
- **S1** BC-10 handover 4, the `height: 54px` literal (Monogram.module.css:8): the value is correct, nothing changes visually, and it would mean churn in a landed module.
- **S2** BC-12 handover a, the status line showing `Page 1 of N` while the pager marks the clicked page during a load: the status line describes the list on screen (§5.9 "stay on screen, unchanged"; the demo's loading box shows `Page 1 of 2`) and the pager shows the URL's page. The gap lasts under 100 ms locally.
- **S3** BC-10 handover 3, FilterBar part (`--tracking-button` and the `:active` colour on Clear): already removed in 13ae060 (BC-9). The remaining textual difference is D4.
- **S4** At 375 the one-line ellipsis can cut a keyword match: that is §5.4's one-line rule.
- **S5** Browser chrome outside BRANDING (text-selection highlight, scrollbars, the native select popup that §5.2 allows, autofill tint).
- **S6** The visually hidden "Results" heading takes focus after a page change with no visible ring (ListingList.module.css:23-25): it is not interactive, and this is SPEC §7.4's design.
- **S7** Screen-reader feedback for an invalid submit is already BC-21.
- **S8** Emoji grep: the only hits are in `.lattice/` (lattice artifact payloads and orchestration evidence, which hold vitest and build logs). These records are immutable, are not product, and are left out of the dist (SPEC §12), so they are not edited. The C14 row records "no hits in shipped files" with the exact command. Validators should summarise vitest output instead of pasting its ✓ lines into new evidence.

## 6. Text-button recipe decision (ticket-card note 2)
All four implementations compute the same values: 13px, line-height 1, weight 500, letter-spacing normal, oxblood, no background, border or padding, nowrap, pointer cursor; hover adds an underline at 3px offset; focus-visible is a 2px oxblood outline at 3px offset; disabled is ink-3. They match demo `.link`. The implementations are FilterBar `.clear` (tokens; fixed in 13ae060), Pager `.button` (tokens), and NoResults/ErrorState `.action` (a `font` shorthand, which also resets `font-variant-*`; harmless, since these buttons have no digits). The decision is not to consolidate into a shared class or component. A global `.textButton` would be a new pattern beside the CSS modules, touch four TSX files, four modules and global.css, and BC-19 may still change the recipe. D4 aligns the textual form instead, so the four blocks read the same and BC-19's eventual change is mechanical. BC-19 is the right place to consolidate if the recipe changes.

## 7. Shared-file handling
- tokens.css: no change (every value checked). global.css: D6 only, a fix as the run-state row allows.
- Component modules touched, all minimal and local: ListingList (D2), ListingCard (D3), NoResults and ErrorState (D4), all from BC-10; LoadingLine (D5, BC-11). SearchPage and shell.module.css (D1) belong to BC-12, which has now landed, so no live branch conflicts. Nothing in src/format, src/hooks or src/api changes, and no new modules or helpers are added.
- Suggested commits: (a) D1 + D2 + the SPEC sentence; (b) D3; (c) D4 + D5 + D6. Each ends with the attribution trailer.

## 8. Validation (validator sub-agent, actor agent:opus-5.5-validator)
- **V0 Servers**, from the worktree root WT, both started with run_in_background:
  - `cd $WT/backend && go build -o bin/api ./cmd/api`
  - `cd $WT && PORT=18014 DB_PATH=/tmp/brillio-bc14.db CORS_ORIGINS=http://localhost:15114 ./backend/bin/api`, logging to the scratchpad.
  - `cd $WT/frontend && VITE_API_PROXY_TARGET=http://localhost:18014 npx vite --port 15114 --strictPort`
  - Check: `curl -s localhost:18014/healthz` returns ok, and `curl -s localhost:15114/api/cities` lists 6 cities.
- **V1 c11 panel** (operator view, as the card specifies): `c11 new-panel --type browser --area area:9` on http://localhost:15114/. WKWebView cannot emulate a width or send trusted keys (see the BC-10 and BC-12 evidence), so this panel is a visual cross-check only. Close it with `c11 close-panel --panel <ref>`, never with a positional ref.
- **V2 Evidence browser**: headless Chrome 155 driven over CDP by a Node 24 script in the scratchpad (built-in WebSocket and fetch, no installs). Start it with `"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new --remote-debugging-port=19214 --user-data-dir=<scratch>/chrome-bc14 about:blank`. Use `Emulation.setDeviceMetricsOverride` at 1280×800 and 375×812 (DPR 2), `Page.captureScreenshot`, `Runtime.evaluate` (awaitPromise), `Input.dispatchKeyEvent` (Tab 9, Enter 13, PageDown 34, ArrowDown 40, Shift modifier 8), `Input.insertText` for typing, `Fetch.enable` to hold a search request while checking the loading state, and `Emulation.setEmulatedMedia` for prefers-reduced-motion. Wait for `document.fonts.ready` and the expected status text before every shot.
- **V3 Screenshots** go to /Users/forrest/Code/brillio/.lattice/orchestration/evidence/ as `BC-14-<state>-<1280|375>.png`, eight in all:
  - default: `/`, status `12 results · newest first`.
  - noresults: `/?city=Atlantis`, showing `No results.` and `0 results`.
  - invalid: start from a settled `/`. Type Min price 500000, Tab, Max price 400000, Tab. Both fields are marked, the single message `Min price must be less than max price.` sits under Max price, the list stays, and no new search reaches the API (check the API log).
  - error: stop the API, then search with a new query (Keyword `pets`, Enter). Expect `Couldn't load results.` and an empty status line.
  - Optional: `BC-14-default-375x667.png` for O14.
- **V4 Runtime oracles** (JS via Runtime.evaluate at both widths; paste the outputs into the evidence):
  - (a) `scrollingElement.scrollHeight <= innerHeight` and `scrollY == 0`, before and after scrolling the results to the bottom.
  - (b) The header, form, status and nav rects do not change when the results scroll.
  - (c) Oxblood census: elements whose own computed color, background, border (width > 0), outline (style not none) or text-decoration is rgb(122,31,43) or rgb(94,22,32), skipping a `color` inherited unchanged from the parent. Expect only the five jobs, plus the segment while loading.
  - (d) Any border-radius other than 0px appears only on the rank pills.
  - (e) No box-shadow, background-image, filter or backdrop-filter, no opacity below 1, and no img, svg or canvas.
  - (f) No Bodoni below 18px, and `document.fonts.check` is true for both faces.
  - (g) Uppercase only on labels, the wordmark, the scope and the status labels.
  - (h) `document.getAnimations()` is empty when idle and holds only the segment's `slide` while a held search loads. Under reduced motion it is empty and the segment is as wide as the track.
  - (i) Row 44: every input and select has exactly one label. `<main>` has aria-live polite, and aria-busy is true while the search is held and false after. A progressbar named Loading exists while it is held. The current page has `aria-current="page"`. The disabled end is `disabled`.
  - (j) The `html` element's computed `-webkit-tap-highlight-color` is `rgba(0, 0, 0, 0)` (D6).
  - (k) No meta item breaks inside itself at 375 (screenshot).
- **V5 Text-button parity**: compare computed font-size, line-height, weight, letter-spacing, color, padding, text-decoration and cursor across Clear, Clear filters, Try again, Previous, Next and a page number. All must be identical.
- **V6 Keyboard-only pass** at 1280, trusted keys only, transcript in BC-14-keyboard.md. Log one line per key: key, then activeElement (role and name), URL, status halves, main.scrollTop.
  - (1) From `/`, Tab repeatedly. Expect Min price, Max price, Min bedrooms, City, Keyword, Target budget, Status, Search, Clear, the results region (with a visible ring), Page 1, Page 2, Next; Previous is skipped.
  - (2) Shift+Tab to Target budget, type 450000, Enter. Expect `?targetBudget=450000` and `12 results · ranked for $450,000`.
  - (3) Tab to the results and press PageDown. main.scrollTop rises and scrollY stays 0.
  - (4) Tab to Page 2, Enter. Expect `page=2`, `Page 2 of 2`, focus on #results-heading and the first rank `No. 11`. Tab now reaches Previous while Next is skipped; Enter on Previous returns to page 1.
  - (5) Stop the API. Move to Keyword, type `pets`, Enter. Expect the error state. Restart the API, Tab to Try again, Enter. Expect `7 results · newest first`.
- **V7 Evidence and checklist**: write BC-14-validation.md (HEAD sha, commands, oracle outputs, keyboard transcript, file list) and attach each file with `LATTICE_ROOT=/Users/forrest/Code/brillio lattice attach BC-14 <file> --role validation --actor agent:opus-5.5-validator`. The validator also posts the acceptance checklist comment (`lattice comment BC-14 --file … --role validation`). It has one line per A and C row: pass, fixed (Dn) or owner (On), plus the runtime result.
- **V8 Teardown**: kill the API, vite and Chrome; `rm -f /tmp/brillio-bc14.db*`; close the c11 panel with `--panel`; confirm `git status --porcelain` is clean.

## 9. Gate
On the final HEAD: `cd $WT && make check` (lint and tests for both halves) and `cd $WT/frontend && npm run build`, both green, with the sha named in the evidence.

## 10. Risks and size
- jsx-a11y `no-noninteractive-tabindex` fails D1 without the disable comment, and prettier may move the comment. Run `npm run lint` and `npm run format:check`.
- D3 adds NBSPs to the DOM text. Any raw `textContent` equality on the meta line must normalise them, and copying the meta line yields NBSPs (acceptable).
- D1 adds one Tab stop and an oxblood ring. If the owner answers O12 differently, only the CSS changes.
- The tooling is the main risk. If headless Chrome over CDP fails, fall back to BC-10's iframe-host method in c11 for screenshots, and flag the keyboard pass as not executable (`needs-human`) rather than faking it.
- The literal reading of the emoji acceptance fails because of `.lattice/` (S8). The delegator states the scoping in the checklist.
- Size: about 50-60 changed lines in frontend (D1 ~12 with tests, D2 ~6, D3 ~22 with tests, D4 ~8, D5 ~2, D6 ~3) plus one SPEC sentence.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)
Review: art_01M4WDQP79JFKTS1PVSDWKGBMY (fable sub-agent, PASS-WITH-NITS @ 0b24829). Every finding was obvious; none deferred, none needs a human beyond the O list.

| # | Concern | Resolution | Section |
|---|---|---|---|
| R1 | Two later handover comments untriaged (BC-12 validation item 7; BC-12 implementation FilterBar comment) | **S9**: cards cut off flush under the status line when the results scroll. Same construction as the demo (demo:42 `.status` no border, :43 `.results` padding-top 6px, :45 first card no hairline) and BRANDING §4.2; a top fade would be a gradient (§4.3, §8). No change; whether the status row should carry a hairline is appended to **O14**. **D7**: FilterBar.module.css:2-3 comment ("a direct child of the shell") is rewritten to say the bar sits in SearchPage's row-2 wrapper with the LoadingLine. Comment only; lands in the D4-D6 commit. | §3, §4, §5 |
| R2 | D1's oxblood ring on `<main>` is a placement beyond §2.2 job 3 ("buttons and links") but A3/C1 read pass | A3 and C1 now read **"pass + D1 ring (O12), D5 (O6)"**; the checklist comment says it in words. D1 proceeds (tabIndex is engineering, SPEC §7.6). The SPEC §7.6 sentence states only the engineering fact (focusable, keyboard-scrollable), never the ring's look. O12 also notes D1 adds a Tab stop in the no-results and error states, where nothing scrolls. The O list is raised as `needs-human` before implementation starts. | §2, §3 D1, §4 O12 |
| R3 | V6(5) expected string wrong (targetBudget persists from step 2) | V6(5) becomes: press Clear first (keyboard), then Keyword `pets`, Enter with the API stopped → error state; restart the API, Try again → `7 results · newest first`. | §8 V6 |
| R4 | Headless Chrome 130+ makes scrollers Tab-reachable without tabIndex, so V6(1) cannot attribute the stop to D1 | At the results stop V6(1) also logs `activeElement.tagName`, `getAttribute('tabindex') === '0'` and computed `outline-color === rgb(122, 31, 43)`. BC-14-keyboard.md states that the WebKit path is proven by the new SearchPage jsdom unit test, because the WKWebView panel cannot send trusted keys. | §8 V6 |
| R5 | D2 misdescribes ListingList.tsx:16-18 (the comment says the section itself is the polite live region) | D2 rewrites that comment to: "The live region is SearchPage's `<main>` (SPEC section 7.6); this section is labelled by its heading only." Rename the test at ListingList.test.tsx:24 to match. | §3 D2 |
| R6 | C14 `grep -P` does not exist on macOS grep | Executable command: `git ls-files -z \| grep -zv '\.pdf$' \| xargs -0 perl -CSD -ne 'print "$ARGV:$.\n" if /[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]/'`, with `.lattice/` reported separately. Checklist wording: "no hits outside `.lattice/`; inside it, U+2713 in vitest/validation logs plus binary noise in PNGs (immutable records, not shipped, SPEC §12)". The validator writes no new ✓ lines into evidence (summarise vitest output). | §2 C14, §5 S8 |
| R7 | Nit: D3 invisible NBSP literal | Write `const NBSP = ' '`. | §3 D3 |
| R8 | Nit: V5 page number | Compare against a non-current page number; `font-variant-numeric` is deliberately not compared (page numbers inherit tabular-nums from `.pager`). | §8 V5 |
| R9 | Nit: V0 paths | `DB_PATH` lives in the validator's session scratchpad, not `/tmp`; `/healthz` returns `{"status":"ok"}`. V8 removes the scratch DB. | §8 V0, V8 |
| R10 | Nit: D6 vs S5 | Keep D6; record it in the checklist under §4.3 and §8 as "fixed (D6)", no owner question. | §2 A12, C3 |
