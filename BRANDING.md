# Brillio Listings — Branding

| | |
|---|---|
| **Status** | v1.1 draft, 2026-10-14. v1 composed from the branding interview picks; v1.1 makes a page change quiet (§5.9, BC-25). Awaiting owner approval of the demo. |
| **Demo artifact** | https://claude.ai/artifact/HdSoc6KB2Wzwvmjt73AGTn — republish to this URL on every change |
| **Demo source** | `docs/branding-demo.html` — the reference implementation of every token below |
| **Interview record** | https://claude.ai/artifact/87Spdddf136Uv72dn8QWUf — picks of 2026-10-13 |
| **Owner** | Forrest Almasi |

This file is the single source of truth for how Brillio Listings looks and speaks. Every rule an implementer needs is here. If a question is not answered here, the answer is "the plainer option"; then add the rule.

## 0. What was decided

The interview offered rendered options on the brief's own listing data. The picks, and what each one commits us to:

| Question | Pick | Commitment |
|---|---|---|
| Direction | Gallery | Bone paper, hairlines, air. Luxury is editorial restraint. |
| Ground | Light only | One palette, done completely. No dark mode. |
| Type | Bodoni Moda + Figtree | A Didone for what is being sold, a humanist sans for the controls. |
| Accent | Oxblood | One color, `#7A1F2B`, and it means attention. |
| Results | List cards | Two to three lines per listing with an image slot. |
| Stand-in | City monogram | The slot holds a tinted block with the city's initials. No photos exist in the feed. |
| Score | Rank + decimal | "No. 1" pill with the raw score beside it. |
| Surfaces | Hairline | 1px lines, square corners, no fills, no shadows. |
| Voice | Plain product | The fewest words that still say what to do. |
| Name | Brillio Listings | The company prefix; reads as the company's own tool. |
| Never | Emoji | Anywhere in the UI. |
| Reference | App-like scroll feel | Pinned header and pager, only the list scrolls. See §4.2 and §10. |

Constraints inherited from the brief: every result shows address, price, bedrooms and relevance score; filters are min price, max price, min bedrooms, city, keyword and target budget; pagination is visible and operable; loading, no-results, error and invalid-input states exist. The live session adds an unseen requirement, so §5 names where a new field, filter or column goes without a redesign.

## 1. Principles

- **Paper, ink, one color.** The interface is bone paper and near-black ink. Oxblood exists for one reason: to mark what needs you now or takes you somewhere. It never decorates.
- **The property is set in Bodoni; the tool is set in Figtree.** Addresses, prices and state headings are the serif. Labels, inputs, buttons, meta and body are the sans. The split is the whole visual idea.
- **App, not document.** Header, filter bar and pager are pinned. The results list is the only thing that scrolls, with native momentum.
- **Nothing decorative.** No shadows, gradients, icons, illustrations, badges, emoji or fills where a hairline will do. Structure comes from spacing, type size and 1px rules.

## 2. Color

### 2.1 Tokens

| Token | Hex | Use |
|---|---|---|
| `paper` | `#F6F5F1` | Page and app ground. The only ground. |
| `stone` | `#E8E7E1` | Monogram block fill; disabled button fill. The only tinted surface. |
| `line` | `#D9D7D0` | Hairlines: between cards, under the header, above the pager. |
| `ink` | `#161616` | Addresses, prices, body text, rank pill, wordmark. |
| `ink-2` | `#6E6C66` | Meta lines, labels, monogram letters, placeholder text, input underline at rest. |
| `ink-3` | `#A3A19A` | Disabled text only. Never for anything that must be read. |
| `oxblood` | `#7A1F2B` | Attention. §2.2 lists every place it may appear. |
| `oxblood-deep` | `#5E1620` | Hover and pressed state of anything oxblood. |

```css
:root {
  --paper: #F6F5F1; --stone: #E8E7E1; --line: #D9D7D0;
  --ink: #161616; --ink-2: #6E6C66; --ink-3: #A3A19A;
  --oxblood: #7A1F2B; --oxblood-deep: #5E1620;
  color-scheme: light;
}
```

### 2.2 Where oxblood appears

Exactly five jobs. If oxblood is somewhere else, remove it.

1. The primary button fill (one per screen: Search).
2. Text buttons and links (Clear, Try again, Next, page numbers).
3. The focused text field's underline (2px) and the focus ring on buttons and links.
4. An invalid field's underline (2px) and its message.
5. The current page number in the pager.

Oxblood never colors a price, an address, a rank, a score, a monogram, a status label or a hairline.

### 2.3 Semantic color

None. There is no green for success, amber for warning or red for error. Success is results appearing. Invalid input is job 4 above: the field itself is marked, in the one color that already means attention. Errors are stated in ink with an oxblood action (§5.9). A second red beside oxblood would be two reds.

### 2.4 Contrast, computed

WCAG ratios on the surfaces each pair actually meets (relative luminance, sRGB). Text needs 4.5:1, or 3:1 at 24px regular and above; component boundaries need 3:1.

| Pair | Ratio | Requirement | Result |
|---|---|---|---|
| `ink` on `paper` | 16.59:1 | 4.5 | pass |
| `ink-2` on `paper` | 4.81:1 | 4.5 | pass |
| `oxblood` on `paper` (links, messages, current page) | 9.35:1 | 4.5 | pass |
| `oxblood-deep` on `paper` | 11.94:1 | 4.5 | pass |
| `paper` on `oxblood` (button label) | 9.35:1 | 4.5 | pass |
| `ink` on `stone` | 14.61:1 | 4.5 | pass |
| `ink-2` on `stone` (monogram, 26px) | 4.24:1 | 3.0 (large text) | pass; never set the monogram below 24px |
| `ink-2` underline vs `paper` (input boundary) | 4.81:1 | 3.0 | pass |
| `oxblood` focus ring vs `paper` | 9.35:1 | 3.0 | pass |
| `ink-3` on `paper` (disabled) | 2.37:1 | none (inactive) | informational |
| `line` vs `paper` (decorative hairline) | 1.32:1 | none | informational |

### 2.5 Dark mode

None, by decision. The app declares `color-scheme: light` and ignores the system setting. A dark palette would be a second design, not an inversion, and the owner chose to spend the effort on one ground.

## 3. Typography

### 3.1 Faces

| Role | Face | Source | Fallback stack |
|---|---|---|---|
| Display | Bodoni Moda (variable: `ital`, `opsz` 6–96, `wght` 400–900) | Google Fonts | `"Bodoni 72", Didot, Georgia, serif` |
| UI | Figtree (variable: `wght` 300–900) | Google Fonts | `system-ui, -apple-system, "Segoe UI", sans-serif` |

```html
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Bodoni+Moda:ital,opsz,wght@0,6..96,400..700;1,6..96,400..700&family=Figtree:wght@400..600&display=swap">
```

Set `font-optical-sizing: auto` so Bodoni uses its text optical size at 22px and its display size at 28px and above. Two weights of Bodoni (400, 500) plus 400 italic; three of Figtree (400, 500, 600). Nothing heavier.

### 3.2 Scale

| Role | Face | Size / line | Weight | Tracking | Notes |
|---|---|---|---|---|---|
| Price | Bodoni Moda | 28 / 31 | 500 | −0.01em | Right-aligned in the card. |
| Address | Bodoni Moda | 22 / 26 | 400 | 0 | `text-wrap: balance`. |
| State heading | Bodoni Moda | 22 / 26 | 400 italic | 0 | "No results." and "Couldn't load results." only. |
| Monogram | Bodoni Moda | 26 / 1 | 400 | 0 | `ink-2` on `stone`. Never below 24px (§2.4). |
| Page heading (docs, README site) | Bodoni Moda | 24 / 29 | 400 | 0 | Not used inside the app. |
| Body | Figtree | 15 / 22 | 400 | 0 | Input values, state body copy. |
| Description | Figtree | 14 / 20 | 400 | 0 | One line, ellipsis. Keyword match at 600. |
| Meta | Figtree | 13 / 18 | 400 | 0 | `ink-2`, tabular figures. |
| Status line, pager, score | Figtree | 13 / 18 | 400 | 0 | Tabular figures. |
| Button | Figtree | 13 / 1 | 500 | +0.02em | |
| Wordmark | Figtree | 13 / 16 | 600 | +0.16em | Uppercase. |
| Label | Figtree | 11 / 14 | 500 | +0.08em | Uppercase, `ink-2`. Field labels, status labels, section eyebrows. |
| Rank pill | Figtree | 11 / 1.3 | 500 | +0.04em | |

### 3.3 Rules

- Bodoni never sets below 18px. Its hairlines break; anything smaller is Figtree.
- Tabular figures (`font-variant-numeric: tabular-nums`) wherever digits sit in a column: meta, score, pager, status line, inputs. Prices in Bodoni are proportional and right-aligned, which is enough.
- Uppercase only in labels and the wordmark. Tracking only where uppercase is, plus +0.02em on buttons.
- Italic only for the two state headings.
- No text below 4.5:1 except disabled text (`ink-3`), which carries nothing the user needs.

## 4. Layout and spacing

### 4.1 Grid

- 8px grid; 4px for the gap between a label and its value.
- Gutter 24px, 16px under 760px.
- Content width inside the app: 1040px max, centered. The shell itself fills the viewport.
- Card padding 20px top and bottom; column gap 20px; monogram column 72px.
- Section gap in documents 56px.

### 4.2 App shell

```
┌ header 56px ───────────────────────────────┐  pinned
├ filter bar (auto) ─────────────────────────┤  pinned
├ status line (auto) ────────────────────────┤  pinned
│ results (1fr, overflow-y: auto)            │  the only scroller
├ pager 48px ────────────────────────────────┤  pinned
└────────────────────────────────────────────┘
```

`html, body { height: 100%; overflow: hidden }`. The shell is a grid with rows `auto auto auto minmax(0, 1fr) auto`. The results region has `overflow-y: auto; overscroll-behavior: contain`. The browser never scrolls the document, so the filter bar and the pager are always in reach. This is what "app-like scroll feel" means here (§10 flags the reading).

### 4.3 Radius and elevation

- Corner radius 0 on every element except the rank pill, which is `999px`.
- No shadows, ever. No blur, no translucency, no gradients.

### 4.4 Breakpoints

| Width | Filter bar | Card |
|---|---|---|
| > 760px | One row: six fields plus Search and Clear | Three columns: monogram, text, price block right-aligned |
| ≤ 760px | Two columns; Search and Clear on their own row, Search full width | Two columns: monogram, text; price block moves under the meta line, left-aligned, price before rank |

### 4.5 Alignment

Everything left-aligned except the price block (right) and the pager's page numbers (centered). Labels sit above their values, never beside.

## 5. Components

Where a new requirement goes is stated for each component, because the live session will add one.

### 5.1 Header

56px, hairline below. Wordmark `BRILLIO LISTINGS` at left (§3.2). At right, an optional label in `ink-2` for scope ("Northern Virginia"). Nothing else. A new global control (a saved-searches link, a sign-in) goes at the right as a text button.

### 5.2 Filter bar

Six fields in a grid, then the actions. Hairline below the bar.

| Field | Control | Value format |
|---|---|---|
| Min price | number input | `$400,000` on blur; digits while typing; placeholder "Any" |
| Max price | number input | same |
| Min bedrooms | number input, step 1 | `2`; placeholder "Any" |
| City | select | city name; "Any" first |
| Keyword | text input | placeholder "Search descriptions" |
| Target budget | number input | `$450,000`; placeholder "Any" |

Field anatomy: label (§3.2 Label) above, value (Body 15) below, 1px `ink-2` underline. States:

- **Focus:** underline becomes 2px `oxblood`. No ring, no glow.
- **Invalid:** underline 2px `oxblood` and a message beneath the field in `oxblood`, Meta 13. When two fields conflict (min above max), both fields are marked and the message sits under the second.
- **Disabled:** value in `ink-3`, underline `line`.

The select shows a `▾` glyph in `ink-2` at the right of the value; otherwise it matches a text field. Native pickers are fine.

Actions: the Search button (§5.8) and a "Clear" text button. Search submits; Enter in any field submits. Search is disabled while a search loads; a page change leaves it enabled (§5.9).

A new filter joins the grid as a seventh field with the same anatomy. The bar wraps to a second row at 1040px before any field shrinks below 130px.

### 5.3 Status line

Meta 13 in `ink-2`, left and right halves.

- Left: `12 results · ranked for $450,000`. Without a target budget: `12 results · newest first`. While a search loads: `Loading`; while a page loads, the count and order stay. No results: `0 results`. On error: empty.
- Right: `Page 1 of 2`, the page on screen; while a page loads, the page asked for. Empty when there are no pages.

### 5.4 Listing card

Three columns: monogram (§5.5), text, price block. Hairline above every card but the first. No hover state; cards are not links in v1 (§10).

Text column, top to bottom:

1. **Address** (Bodoni 22): the `address` field verbatim. The feed does not normalize addresses and the card does not either.
2. **Meta line** (Figtree 13, `ink-2`), items separated by ` · `: status label if not active (`PENDING`, `SOLD`, Label style) · `City, ST ZIP` · `2 bd` · `1.5 ba` · `980 sqft` · `Listed Aug 29` · `MLS A`.
3. **Description** (Figtree 14): one line, ellipsis. When a keyword filter is active, each match is set at weight 600. Emphasis is weight, not color.

Price block, right-aligned: **price** (Bodoni 28) over **rank pill + score** (§5.6).

A new per-listing field joins the meta line in the position that reads naturally (a lot size after sqft, an HOA fee after price as a second Meta line under it). A new prominent number (price per sqft, days on market) becomes a second line in the price block, Meta 13, `ink-2`.

### 5.5 Monogram

A 72×54 block (4:3) in `stone` holding the city's first two letters, first uppercase, second lowercase (`Sp`, `Vi`, `Fa`, `Re`, `Ma`, `Ch`), Bodoni 26 in `ink-2`, centered. It is a texture that gives the list rhythm, not an identifier: two cities sharing initials is acceptable. If photos ever arrive, they take this slot at the same size and the monogram becomes the fallback.

### 5.6 Score

`No. 1` in a pill (1px `ink` border, radius 999, Figtree 11/500, padding 2px 8px) followed by the score to two decimals (Figtree 13, `ink-2`, tabular), right-aligned under the price.

- Rank is the position in the whole ranked result set, so page 2 begins at `No. 11`.
- Ties show identical decimals and sequential ranks. The engine breaks ties deterministically; the UI never hides that two scores are equal.
- The score is the API's `relevanceScore` in `[0, 1]`, rounded for display only.

### 5.7 Pager

48px, hairline above, pinned. `‹ Previous` at left and `Next ›` at right as text buttons; page numbers centered with 16px gaps. The current page is `oxblood`, weight 500, with a 1px `oxblood` underline. A disabled Previous or Next is `ink-3` and not focusable. Page size is 10 and is not a control in v1; a page-size control, if asked for, joins the right side of the status line as a select.

### 5.8 Buttons

- **Primary** (one per screen): `oxblood` fill, `paper` text, 40px tall, padding 0 20px, Figtree 13/500 +0.02em, radius 0. Hover `oxblood-deep`. Focus-visible: 2px `oxblood` outline, offset 2px. Disabled: `stone` fill, `ink-3` text.
- **Text button** (everything else): Figtree 13/500 in `oxblood`, no background, underline on hover with 3px offset. Focus-visible as above with 3px offset.
- Buttons are verbs: Search, Clear, Try again, Next, Previous. Never "OK", never "Submit".

### 5.9 States

| State | Where | What |
|---|---|---|
| Loading | Status line + a 2px line under the filter bar | A search: the first load, new filters (Search, Enter, Clear, back or forward), or Try again. Status reads `Loading`. The previous results stay on screen, unchanged. A 2px `oxblood` segment (40% wide) slides left to right along a `line`-colored track, 1.2s linear, looping. Search is disabled. First load with no prior results: the status line alone. |
| Page change | Results region | Only the page changes: Previous, Next, a page number, or back or forward between pages of the same filters. A Search, Enter or Clear that changes only the page is a page change too. Nothing outside the results moves: no loading line, Search stays enabled, the status line keeps its count and order and shows the page asked for at once, and the pager marks that page current. The previous page stays on screen, unchanged, until the next one arrives and replaces it in place. After an error there is no list to keep, so the next load is a search. |
| No results | Results region | Centered, 96px from the top: `No results.` (State heading) · `Adjust your filters.` (Body, `ink-2`) · `Clear filters` text button. Status reads `0 results`. |
| Invalid input | The fields | The API's 400 maps to fields. `Min price must be less than max price.` under Max price with both fields marked. `Min bedrooms must be 0 or more.` under that field. The results region keeps the last good results. Nothing is sent until the user corrects it or the API rejects it; the UI validates the same rules the API does, and the API is the authority. |
| Error | Results region | Centered: `Couldn't load results.` (State heading) · `Try again` text button, which re-runs the last search. Status line empty. No error code in the UI; it goes to the console. |

No spinners, no skeleton rows, no toasts, no modals.

### 5.10 Icons

None. The only glyphs are `‹` `›` on the pager and `▾` on the select, set as text in the UI face. A new need for an icon is a need for a word.

## 6. Motion

The loading line (§5.9) is the only animation. Focus underlines change width instantly; hovers change color instantly; results replace in place with no fade. Under `prefers-reduced-motion: reduce` the loading line is a static full-width `oxblood` line.

## 7. Copy

Voice: plain product. Sentence case. The fewest words that still tell the user what to do. No exclamation marks, no emoji, no "please", no "oops", no apologies, no cheer. Labels are nouns; buttons are verbs. Errors say what is wrong and what to do, in that order.

### 7.1 Formatting

| Thing | Format | Example |
|---|---|---|
| Price | `$` + thousands separators, no cents | `$450,000` |
| Bedrooms | integer + ` bd` | `2 bd` |
| Bathrooms | drop `.0`, keep `.5` + ` ba` | `1 ba`, `1.5 ba` |
| Area | thousands separators + ` sqft` | `1,450 sqft` |
| Listed date | `Listed Mon D`; add `, YYYY` when not the current year | `Listed Aug 29` |
| Source | underscore to space | `MLS A` |
| Status | Label style, only when not active | `PENDING` |
| Score | two decimals | `0.94` |
| Rank | `No. ` + integer | `No. 1` |
| Counts | integer + ` results` (also `1 results` becomes `1 result`) | `12 results` |
| Pages | `Page N of M` | `Page 1 of 2` |

### 7.2 Strings

| Where | Text |
|---|---|
| Field labels | Min price · Max price · Min bedrooms · City · Keyword · Target budget |
| Placeholders | Any · Any · Any · Any · Search descriptions · Any |
| Buttons | Search · Clear · Clear filters · Try again · Previous · Next |
| Status, results | `12 results · ranked for $450,000` / `12 results · newest first` |
| Status, loading | `Loading` |
| Status, none | `0 results` |
| No results | `No results.` / `Adjust your filters.` |
| Error | `Couldn't load results.` |
| Invalid, price | `Min price must be less than max price.` |
| Invalid, bedrooms | `Min bedrooms must be 0 or more.` |
| Invalid, budget | `Target budget must be more than $0.` |

## 8. Don'ts

- No emoji, anywhere, including README and commit messages that ship with the project.
- No color other than oxblood on paper and ink; no second accent; no semantic green, amber or red.
- No oxblood outside the five jobs in §2.2.
- No shadows, gradients, blur, translucency, rounded corners (except the rank pill), fills (except `stone` and the primary button), icons, illustrations, stock imagery or photos that are not from the feed.
- No Bodoni below 18px; no third typeface; no weights beyond §3.1.
- No dark mode, no theme switch.
- No spinners, skeletons, shimmer, toasts, modals, fade-ins or list animations.
- No hover backgrounds on cards; no zebra striping.
- No reduced-opacity "disabled" looks; disabled is `ink-3` and `stone`.
- No placeholder data in demos; use `sample_listings.json`.
- No relative dates ("3 days ago"); dates are absolute.

## 9. Change process

1. Edit this file.
2. Update `docs/branding-demo.html` so it demonstrates the changed rule, and republish it with the Artifact tool to the **same URL** recorded at the top of this file.
3. Ask the owner to review the demo. Approval is given in chat or as an artifact comment.
4. Bump the version line at the top.

Once the app exists, a change here goes through a Lattice task like any other change; until then, this file and the demo are the whole record.

## 10. Open decisions

- **"App-like scroll feel."** Read as: pinned header and pager, results as the only scroller (§4.2). If it meant something else (page transitions, a mobile-first stack), say so and §4.2 changes.
- **Wordmark treatment.** The chosen rendering was tracked uppercase Figtree. A Bodoni wordmark was not shown and might suit the serif-for-the-property idea better. Decide on the demo.
- **Near-duplicates.** The feed has the same property under two sources (123 Main St in MLS A and B). The UI shows both, labelled by source, because the brief says ids are unique per source and nothing more. Collapsing duplicates is a product decision for the engine, not a branding one; if it is made, the card gains a `2 sources` item in the meta line.
- **Cards as links.** There is no detail view in the brief, so cards are inert. If one is added, the address underlines on hover in `ink` and the whole card is the link.
- **Target budget absent.** §5.3 says `newest first`; the engine's behavior without a budget must match that wording or the wording changes.
