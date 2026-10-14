# BC-10 validation: results surface components

- Validator: agent:opus-5.5-validator (fresh context)
- Worktree: `/Users/forrest/Code/brillio-worktrees.noindex/bc-10-results-surface`, branch `feat/BC-10-results-surface`
- HEAD: `fa7c08dfe01e3d52922a57765c5b2a269ca34e80` (checked before starting and again after cleanup)
- Date: 2026-10-14 (run 03:58Z to 04:05Z; the vitest header shows local time 23:58 on 10-13)
- Scope: the Orchestrator-approved items (03:39Z): 1 tests, 2 visual check against `docs/branding-demo.html` (plan R2), 3 cleanup

## Overall: PASS

| Item | Result |
|---|---|
| 1. Test evidence (vitest verbose) | PASS: 7 files, 41 tests |
| 2a. Visual check, desktop (1100px) | PASS: A1 matches the demo card in structure and type; differences are only the known accepted ones |
| 2b. Visual check, 375px | PASS: BRANDING §4.4 reflow; differences from the demo are exactly the accepted D1 set |
| 3. Cleanup | PASS: server killed, port free, my panel closed, preview files deleted, `git status --porcelain` empty |

No implementation-level or plan-level findings. Two low observations follow at the end; neither is a BC-10 defect.

## 1. Tests

Command, run from `frontend/`:

```
npx vitest run --reporter=verbose src/components/{ListingCard,Monogram,ScorePill,ListingList,NoResults,ErrorState} src/format/keyword.test.ts
```

Result: exit 0, `Test Files 7 passed (7)`, `Tests 41 passed (41)`. The full output is in `BC-10-vitest.txt`.

Which tests cover each ticket requirement:

- **Meta order:** `ListingCard > renders every meta item in BRANDING order`.
- **PENDING for A7, none for active listings:** `ListingCard > shows PENDING first for A7` and `ListingCard > shows no status label for an active listing`.
- **`Pets`/`pets` emphasis:**
  - `ListingCard > emphasizes Pets and pets case-insensitively and leaves the rest untouched`
  - `ListingCard > renders no emphasis without keyword tokens`
  - `emphasisSegments > matches case-insensitively and keeps the original casing`, plus 7 more keyword.ts cases (`splitKeyword ...`, every occurrence, token order, merging overlapping and adjacent runs, regex metacharacters, blank tokens, empty text).
- **ScorePill `No. 11` and `0.87`:**
  - `ScorePill > renders No. 11 and the score to two decimals`
  - `renders a fixture's rank and score`
  - `puts the rank pill before the score`
- **Monogram `Sp`/`Vi`/`Fa`:**
  - `Monogram > renders the city initials: Springfield as Sp`, `...: Vienna as Vi` and `...: Fairfax as Fa`
  - `hides the decorative block from assistive technology`
- **ListingList region and focusable heading:**
  - `labels a polite live region with the Results heading`
  - `exposes a level 2 Results heading that script can focus`
  - `renders one card per item in rank order inside an ordered list`
  - `passes keyword tokens to every card`
  - `renders the heading and no list for zero items`
  - `keys cards by source and id`
- **NoResults copy and callbacks:**
  - `renders the no-results copy verbatim`
  - `calls onClear once on click`
  - `calls onClear from the keyboard`
  - `renders a type=button`
- **ErrorState copy and callbacks:**
  - `renders the error copy verbatim`
  - `calls onRetry once on click`
  - `calls onRetry from the keyboard`
  - `renders a type=button`
  - `shows no error code`
- **No img or icons:** `ListingCard > renders no image, icon, link or button`.
- **Price block and order:**
  - `ListingCard > renders the price over the score pill`
  - `keeps the price block between the meta line and the description`
  - `renders the address verbatim as a level 3 heading`
  - `renders the city monogram`

## 2. Visual check against the demo

### Method

The preview was throwaway and uncommitted. It was deleted afterwards (see §3).

**Files:**

- `frontend/preview-bc10.html`: the same `<head>` as `index.html`, including the same Google Fonts link for Bodoni Moda and Figtree.
- `frontend/src/preview-bc10.tsx`:
  - It imports `styles/tokens.css` and `styles/global.css` the way `main.tsx` does.
  - It renders inside a `div` with the shell's own `shell.results` class (`styles/shell.module.css`), so the gutter is `--inline-pad`, the same as the app's results row.
  - That div holds `<ListingList items={[listingA1, listingB7, listingA7]} keywordTokens={['pets']} />`, then `NoResults` and `ErrorState`, whose callbacks call `console.log`. Each state sits in a 280px box with a dashed top rule; the rule is preview chrome only.
  - `?view=list|states` filters what is shown.
- `frontend/preview-bc10-frame.html`: a host page. It puts each `src` in an `<iframe>` of exactly `w` CSS px, so media queries follow the iframe width. When the row is wider than the panel, the host scales it down uniformly with `transform: scale()`. Layout inside the iframe is unaffected.
- `frontend/preview-bc10-demo.html`: a byte-identical copy of `docs/branding-demo.html` (`cmp` clean). It let the demo load in same-origin iframes at 1100px and 375px. vite's fs allow-list blocks `docs/`, and an http page cannot iframe a `file://` page.

**Width method: iframes.** WKWebView viewport emulation is not supported, and the panel viewport was 763x430 CSS px at DPR 2.

- 1100px iframes were scaled 0.692 to fit the panel.
- 375px iframes are unscaled, ours on the left and the demo on the right.

**Server and browser:**

- Server: `cd frontend && npx vite --port 15110 --strictPort` in the background. npx was pid 35035; the listener was node pid 35079.
- `/preview-bc10.html`, the frame page, the demo copy and the entry all returned 200.
- Browser: `c11 new-panel --type browser --area area:11` created `panel:30`, renamed "BC-10 Preview".
- Before every screenshot, the iframe documents reported `fonts.status === "loaded"`, with Bodoni Moda normal and italic 400-700 and Figtree 400-600 loaded.

**Direct load of the original demo:** `file:///.../docs/branding-demo.html` was also loaded at the panel's own 763px width (`BC-10-demo-fileurl-763.png`).
- It renders the same desktop card structure, with `grid-template-columns: 72px 431px 122px`.
- It shows `Â·` mojibake in place of `·`, because the demo has no `<meta charset>` and WebKit falls back to Latin-1 for `file://`. The vite-served copy is decoded as UTF-8.
- This comes from the demo file, not from BC-10. See observation O1.

### Screenshots (all 1526x860 px, in `.lattice/orchestration/evidence/`)

| File | What it shows |
|---|---|
| `BC-10-ours-1100-overview.png` | Ours at 1100px: A1, B7, A7 (`['pets']`), then NoResults |
| `BC-10-ours-1100-cards.png` | Ours at 1100px: the three cards |
| `BC-10-demo-1100-cards.png` | Demo at 1100px: results region, cards 1-4 |
| `BC-10-ours-1100-states.png` | Ours at 1100px: NoResults and ErrorState |
| `BC-10-demo-1100-states.png` | Demo at 1100px: the states grid (No results, Error) |
| `BC-10-375-ours-vs-demo-A1.png` | 375px: ours (left) and demo (right), A1 and B7 |
| `BC-10-375-ours-vs-demo-A7.png` | 375px: ours and demo, the A7 PENDING card |
| `BC-10-375-ours-vs-demo-noresults.png` | 375px: NoResults, ours and demo |
| `BC-10-375-ours-vs-demo-error.png` | 375px: ErrorState, ours and demo |
| `BC-10-demo-fileurl-763.png` | The original demo via `file://` at 763px (charset artifact, O1) |

### 2a. Desktop (1100px): PASS

What the screenshots show:

- **Layout:** three columns.
  - A 72x54 stone monogram with Bodoni initials `Sp`/`Sp`/`Ch` in ink-2.
  - A Bodoni address.
  - A Figtree meta line in ink-2 with ` · ` separators, in BRANDING order: `Springfield, VA 22150 · 2 bd · 1.5 ba · 980 sqft · Listed Aug 29 · MLS A`.
  - One description line.
  - The price, right-aligned, over the `No. N` pill and the score.
- **Keyword hits:** `Pets` (B7) and `pets` (A7) are weight 600 in ink (`rgb(22, 22, 22)`), the same color as the description. A1's "Pet friendly." does not match `pets` and is correctly left plain.
- **Status label:** A7 shows `PENDING` first as a Label, 11px/500, +0.88px (0.08em), uppercase, ink-2.
- **Hairlines:** one between cards, none above the first. Computed: `li[0] border-top: none`; `li[1]` and `li[2]`: `1px solid rgb(217, 215, 208)` (`--line`). `ol` has `list-style: none`.
- **Images and icons:** `img, svg, picture, i, a` count is 0. The only buttons are the two text buttons.
- **Oxblood:** the only elements with computed `color` or `background-color` `rgb(122, 31, 43)` are the two `.action` buttons.

Computed styles for the A1 card, ours vs the demo, both in 1100px iframes:

| Element | Ours | Demo | Same? |
|---|---|---|---|
| card grid | `72px 806px 122px`, gap 20, padding 20/20 | `72px 708px 122px`, gap 20, padding 20/20 | yes. The middle column differs only because the demo's card sits inside its page frame (card 942px wide vs ours 1040px with the app's `--inline-pad` of 30px) |
| monogram | 72x54, Bodoni Moda 26/26, 400, `rgb(110,108,102)` on `rgb(232,231,225)` | identical | yes |
| address | Bodoni Moda 22px, 400, `text-wrap: balance`, ink, lh 26px | same, lh 26.4px | yes; lh is D1 (<1px) |
| meta | Figtree 13px, 400, ink-2, tabular-nums, margin-top 4, lh 18px | same, lh 18.85px | yes; lh is D1 |
| description | Figtree 14px, 400, ink, nowrap, ellipsis, overflow hidden, margin-top 6, lh 20px | same, lh 20.3px | yes; lh is D1 |
| em | Figtree 14px, 600, normal style, ink | demo `.desc em` 600, normal | yes |
| price | Bodoni Moda 28px, 500, -0.28px (-0.01em), right, 122px wide, lh 31px | same, lh 30.8px | yes; lh is D1 |
| rank pill | Figtree 11px, 500, 0.44px (0.04em), lh 14.3px, 1px solid ink, radius 999px, padding 2px 8px, 47.08x20.3 | identical, 47.08x20.3 | yes |
| score | Figtree 13px, 400, ink-2, tabular-nums, gap 8 | same (lh 18.85) | yes |
| right edges | price and pill flush with the card's right edge (0px) | the same, visually | yes |
| card height | 114px | 115.53px | sum of the D1 line-height deltas |

- **A1 text:** our `innerText` is `Sp | 123 Main St, Apt 4B | Springfield, VA 22150 · 2 bd · 1.5 ba · 980 sqft · Listed Aug 29 · MLS A | $450,000 | No. 1 | 0.74 | Bright top-floor condo near shops and transit. Pet friendly.` It equals the demo's card text except for two things:
  - The score: 0.74 real vs 0.94 illustrative, which is accepted.
  - DOM order: the price block comes before the description, by design (plan D2). The desktop rendering is unchanged.
- **NoResults and ErrorState:** centered, with `padding-top: 96px` and `text-align: center`.
  - Headings: `h2`, Bodoni Moda 22/26, 400, italic, ink.
  - Body: Figtree 15/22, ink-2.
  - Buttons: `type="button"`, Figtree 13px/1, 500, no letter-spacing, oxblood `rgb(122,31,43)`, no background or border.
  - CSSOM confirms the `:hover` rule (underline, 3px offset) and the `:focus-visible` rule (`2px solid var(--oxblood)`, offset 3px) for both buttons. c11's synthetic hover and Tab do not trigger native `:hover` or `:focus-visible` in WKWebView, so these states were confirmed in CSS rather than in pixels.
  - Clicking `Clear filters` and `Try again` (refs e2 and e3) logged `onClear` and `onRetry` in the page console.
  - Visually they match the demo's `.msgbox`: italic display heading, ink-2 body and oxblood text link.
- **Results heading:** `section[aria-labelledby=results-heading][aria-live=polite]` holds `h2#results-heading` "Results" with `tabindex=-1`. It is 1x1 with `clip-path: inset(50%)`, so it is visually hidden. `.focus()` made it `document.activeElement`, with `outline: none`. No heading is visible above the cards, the same as the demo.

### 2b. 375px: PASS

Measured on all three cards:

- Grid columns are `72px 251px`.
- The monogram stays 72x54 with 26px initials. The address stays 22px and the price 28px.
- The price block is `flex-direction: row`, `justify-content: flex-start`, `text-align: left`, `margin-top: 8px`, `column-gap: 14px`.
- The price's left edge equals the address's left edge, so the block is left-aligned in the text column.
- The price sits below the meta line, and the description sits below the price block.
- The price comes before the pill on the same row, 14px apart.
- Hairlines are still between cards only.
- NoResults and ErrorState are centered at 96px with the same type at both widths.

**Known accepted differences from the demo at ≤760px (plan D1; BRANDING wins), all as expected:**

| | Demo | Ours |
|---|---|---|
| monogram | 56x42 at 20px | 72x54 at 26px |
| address | 20px | 22px |
| price | 24px | 28px |
| column gap | 14px | 20px |
| price block | after the description | between the meta line and the description |

Also accepted:

- The demo's A1 score is the illustrative 0.94; the real value is 0.74.
- Line heights are px tokens instead of the demo's multipliers, each under 1px apart.
- The middle column is wider at 1100px because the demo frames its app inside a page; this is not a card difference.

No other differences were observed.

## 3. Cleanup: PASS

- `frontend/dist` already existed before validation, so I left it in place. It is gitignored, and I ran no build, so nothing wrote to it.
- `c11 close-panel --panel panel:30` returned OK. `c11 tree` for area:11 now lists only panel:15, panel:29 and panel:26. No other panel was touched.
- I killed vite: the node listener 35079 and npx 35035. `lsof -nP -iTCP:15110 -sTCP:LISTEN` is empty (exit 1), and `ps` shows neither pid.
- `rm` removed all four preview files: `frontend/preview-bc10.html`, `frontend/preview-bc10-frame.html`, `frontend/preview-bc10-demo.html` and `frontend/src/preview-bc10.tsx`.
- `git status --porcelain` is empty (0 lines). Ignored entries are only the existing `frontend/dist/` and `frontend/node_modules/`.
- HEAD is still `fa7c08dfe01e3d52922a57765c5b2a269ca34e80`.

## Observations (low; not BC-10 defects, no rework)

- **O1 (low, demo file, out of scope):** `docs/branding-demo.html` has no `<!doctype>` or `<meta charset="utf-8">`. Opened as `file://` in c11/WebKit, every `·` and `›` renders as mojibake (`Â·`); see `BC-10-demo-fileurl-763.png`. Served over HTTP it is fine. BC-10 must not edit the demo, so this is a candidate for the branding owner or BC-14 if anyone opens the demo from disk.
- **O2 (low, polish, shared with the demo):** at 375px the meta line wraps inside an item, for example `980 / sqft` and `2.5 / ba`. The demo wraps the same way (`MLS / A`), and BRANDING does not specify it. A possible BC-14 polish is a no-break space inside each meta item. Separately, at 375px the one-line ellipsis can cut a keyword hit (B7 `Pet…`); that follows BRANDING's one-line description rule and the demo.
