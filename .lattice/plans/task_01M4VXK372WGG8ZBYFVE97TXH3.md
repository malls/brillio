# BC-11: Pager, StatusLine and LoadingLine components

Planner: agent:opus-5.5-planner · Worktree `/Users/forrest/Code/brillio-worktrees.noindex/bc-11-pager-status-loading`
(branch `feat/BC-11-pager-status-loading`, base master @e5f49bb). Contract: SPEC §7.2, §7.4–§7.6, §8;
BRANDING §2.1, §2.2 (jobs 2, 3, 5), §3.2/§3.3, §4.2, §5.3, §5.7, §5.8, §5.9, §5.10, §6, §7.1/§7.2, §8;
reference markup/CSS `docs/branding-demo.html` (`.status`, `.pager`, `.bar`, the four state panels).

## Scope

Three presentational components, props exactly as SPEC §7.4, each a folder with `X.tsx`,
`X.module.css`, `X.test.tsx` (Header's pattern: named export, exported `XProps` interface, CSS module,
tokens via `var(--…)`, `@media (max-width: 760px)` literal if ever needed — none is needed here because
`--inline-pad`/`--gutter` already switch at 760px in tokens.css).

Out of scope: placing them in the shell, focus after a page change, Pager visibility per state, aria-busy
(all BC-12); `App.tsx`, `tokens.css`, `global.css`, `shell.module.css`, `vite.config.ts`, other tickets'
component folders. No SPEC edits.

## Files (worktree-relative, all new)

- `frontend/src/components/Pager/Pager.tsx`, `Pager.module.css`, `Pager.test.tsx`
- `frontend/src/components/StatusLine/StatusLine.tsx`, `StatusLine.module.css`, `StatusLine.test.tsx`
- `frontend/src/components/LoadingLine/LoadingLine.tsx`, `LoadingLine.module.css`, `LoadingLine.test.tsx`

Imports: `type SearchState` from `src/hooks/useListingSearch` (it lives there, not in `src/api`; type-only
import); `formatResultSummary`, `formatPageOf` from `src/format/listing`; `searchResponse`,
`emptySearchResponse`, `listingA7` from `src/test/fixtures`. **No changes to BC-8 files**: every fixture
variant is a local spread in the test file (the pattern `format/listing.test.ts` already uses for its
no-budget case), with a comment naming the real search it stands for.

## StatusLine — `StatusLine({ state }: StatusLineProps)`

Rule: **the left half says what the list on screen is (or `Loading`); the right half says where that list
sits in the pages.** A pure module-local `describe(state): [left, right]` switch, using only BC-8 formatters
plus the literal `Loading` (§7.2). `summarize(data) = [formatResultSummary(data), formatPageOf(data.page.page, data.page.totalPages)]`.

| State | Left | Right | Why |
|---|---|---|---|
| idle | `` | `` | Nothing searched, nothing on screen. |
| loading, no previous | `Loading` | `` | §5.9 "first load: the status line alone"; no list yet. |
| loading, previous | `Loading` | `formatPageOf(previous)` e.g. `Page 1 of 2` | §5.9 previous list stays on screen; demo Loading panel shows exactly `Loading` / `Page 1 of 2`. |
| success (budget) | `12 results · ranked for $450,000` | `Page 1 of 2` | §5.3 / §7.2 verbatim. |
| success (no budget) | `12 results · newest first` | `Page 1 of 2` | §5.3. |
| success, 1 item | `1 result · ranked for $450,000` | `Page 1 of 1` | §7.1 singular. |
| success, page past end (3 of 2) | `12 results · ranked for $450,000` | `Page 3 of 2` | SPEC §5.3 returns true metadata; Pager offers recovery. Reachable only by a hand-edited URL. |
| empty | `0 results` | `` | §5.3; `formatPageOf(1, 0)` is ``. |
| invalid, previous | `summarize(previous)` | `summarize(previous)` | SPEC §7.5 "previous list stays; StatusLine unchanged" made concrete: it keeps describing the list still shown. Demo Invalid panel shows `12 results · ranked for $450,000` / `Page 1 of 2`. If previous was an empty response this reads `0 results` / ``. |
| invalid, no previous | `` | `` | e.g. a shared link `?minPrice=-1` on first load: no list on screen, nothing is loading, so idle. |
| error (with or without previous) | `` | `` | §5.3 "On error: empty", §5.9 and SPEC §7.5 "StatusLine empty"; ErrorState replaces the list, so no page is on screen (demo Error panel: both empty). |

Markup: `<div role="status" className={styles.status}><span className={styles.half}>{left}</span><span className={styles.half}>{right}</span></div>`.
`role="status"` keeps the BC-7 placeholder row's role (App.tsx row 3) so BC-12 swaps it in place; it also
announces the count politely. Both spans always render (stable structure, right half stays right).

CSS (`StatusLine.module.css`), from demo `.status`: `display: flex; flex-wrap: wrap; justify-content: space-between;
gap: 0 16px; padding: 14px var(--inline-pad) 0; font-size: var(--text-meta); line-height: var(--leading-meta);
color: var(--ink-2); font-variant-numeric: tabular-nums;` and `.half { min-height: var(--leading-meta); }` so
the row stays 32px (14 + 18) when both halves are empty (error/idle) and the results region does not jump.
No oxblood. Row gap 0 (not the demo's 16) so a wrapped status reads as two Meta lines on narrow phones.

## Pager — `Pager({ page, totalPages, onChange }: PagerProps)`

- `totalPages <= 0` → `return null` (hidden entirely; SPEC §7.5 empty, §5.7).
- Root `<nav aria-label="Pagination" className={styles.pager}>` (same role/name as the BC-7 placeholder row 5).
- Children in DOM order: Previous button, `<div className={styles.pages}>` with one `<button>` per page
  `1..totalPages` (every page, no windowing), Next button. All `type="button"`.
- Previous: `disabled={page <= 1}`; click → `onChange(Math.min(page - 1, totalPages))`.
  Next: `disabled={page >= totalPages}`; click → `onChange(page + 1)`.
  Page `n`: `aria-label={`Page ${n}`}`, `aria-current={n === page ? 'page' : undefined}`; click → `onChange(n)`
  unless `n === page` (current page stays focusable for screen-reader discovery but is a no-op).
- **Page past the end** (`page > totalPages`, SPEC §5.3 200 with empty items): every page renders, none is
  current, Next is disabled, Previous is enabled and goes to the **last page** (`min(page - 1, totalPages)`),
  so recovery is one click. `page < 1` cannot come from the URL parser; Previous is disabled anyway.
- Glyphs are text (§5.10): `<span aria-hidden="true">‹</span> Previous` and `Next <span aria-hidden="true">›</span>`.
  Visible `textContent` is exactly `‹ Previous` / `Next ›` (U+2039/U+203A, one ASCII space); accessible names
  are `Previous` / `Next` (verbs, §5.8), page buttons `Page 1`… (mirrors §7.1 `Page N of M`; visible text
  stays the bare digit, WCAG label-in-name holds). BC-12's tests must query these names.
- Disabled ends use the `disabled` attribute, so they leave the tab order (SPEC §7.6). The Pager never
  moves focus; BC-12 moves it to the results heading after `onChange` (SPEC §7.4), which also covers focus
  sitting on Next when it becomes disabled on the last page.

CSS (`Pager.module.css`), from demo `.pager` + BRANDING §5.7/§5.8:
- `.pager`: `display: grid; grid-template-columns: 1fr auto 1fr; align-items: center; column-gap: 16px;
  height: var(--pager-height)` (48px incl. the hairline, border-box, like Header's 56px)`; padding-inline: var(--inline-pad);
  border-top: 1px solid var(--line); font-size: var(--text-meta); line-height: var(--leading-meta);
  font-variant-numeric: tabular-nums;`
- `.pages`: `display: flex; justify-content: center; gap: 16px;` · `.previous { justify-self: start }` · `.next { justify-self: end }`
- `.button` (text button §5.8; Button row §3.2): `padding: 0; border: 0; background: none; white-space: nowrap;
  font-size: var(--text-button); line-height: var(--leading-button); font-weight: 500;
  letter-spacing: var(--tracking-button); color: var(--oxblood); cursor: pointer;`
  `.button:hover:not(:disabled, [aria-current='page']) { color: var(--oxblood-deep); text-decoration: underline; text-underline-offset: 3px; }`
  (§2.1 oxblood-deep is the hover state of anything oxblood; §5.8 underline at 3px.)
  `.button:focus-visible { outline: 2px solid var(--oxblood); outline-offset: 3px; }` (§2.2 job 3, §5.8).
  `.button:disabled { color: var(--ink-3); cursor: default; }` (no opacity, §8).
- Current page styled by the attribute, so semantics and look cannot drift:
  `.button[aria-current='page'] { color: var(--oxblood); font-weight: 500; text-decoration: underline 1px; text-underline-offset: 3px; cursor: default; }`
  (§5.7 "oxblood, weight 500, 1px oxblood underline"; text-decoration instead of the demo's border-bottom so
  the current button's box height matches its siblings.)
- Oxblood here = jobs 2 (text buttons, page numbers), 3 (focus ring), 5 (current page) only. No hex values.

## LoadingLine — `LoadingLine({ active }: LoadingLineProps)`

- Active: `<div className={`${styles.track} ${styles.active}`} role="progressbar" aria-label="Loading"><div className={styles.segment} /></div>`.
  Indeterminate progressbar: no `aria-valuenow` (correct per ARIA; progressbar has no required props, jsx-a11y
  clean — verified by probe). Name `Loading` reuses §7.2 verbatim; no new copy.
- Inactive: `<div className={styles.track} aria-hidden="true" />` — same 2px box, **no role, no name, no segment,
  no color**. The role exists only while something is loading, so AT never meets an idle progressbar.
- Interpretation of "inactive renders the empty track": the 2px box stays (layout never shifts) but is
  transparent. Reason: placed under the FilterBar's own 1px `line` hairline, a resting `line`-colored track
  would read as a permanent 3px rule, against §1/§4.3 hairlines; the demo draws `.bar` only in the Loading
  panel and the resting app frame shows just the 1px hairline. If the owner wants the grey track at rest,
  it is a one-line move of `background: var(--line)` from `.active` to `.track` (flagged below).

CSS (`LoadingLine.module.css`), from demo `.bar` and §5.9/§6:
```css
.track { position: relative; height: 2px; overflow: hidden; }          /* full-bleed: no inline padding */
.active { background: var(--line); }
.segment { position: absolute; inset-block: 0; inset-inline-start: 0; width: 40%;
  background: var(--oxblood); animation: slide 1.2s linear infinite; }
@keyframes slide { from { transform: translateX(-100%); } to { transform: translateX(250%); } }
@media (prefers-reduced-motion: reduce) { .segment { width: 100%; animation: none; } }
```
(translateX % is of the 40% segment: −100% = off left, 250% = off right. CSS modules scope `slide`.)
Placement note for BC-12 (shell.module.css: "the 2px loading line goes inside row 2 or 3"): render it always
mounted, directly under the FilterBar, e.g. first child of a row-3 wrapper holding `<LoadingLine/><StatusLine/>`.

**Reduced-motion testing under jsdom.** jsdom evaluates no media queries and Vitest does not process CSS
(CSS modules become a proxy returning stable `_name_<hash>` class names). A JS `matchMedia` branch would add
runtime code only to be testable and would still not prove the CSS. So the CSS media query is the
implementation and the test asserts the shipped stylesheet source contains it. Verified by a throwaway probe
(removed; worktree clean): in a test file
```ts
interface NodeFs { readFileSync(path: string, encoding: 'utf8'): string }   // app tsconfig loads no Node types
const fs = await vi.importActual<NodeFs>('node:fs');                        // call as fs.readFileSync (unbound-method lint)
const css = fs.readFileSync(decodeURIComponent(new URL('./LoadingLine.module.css', import.meta.url).pathname), 'utf8');
```
passes vitest, `eslint --max-warnings 0` and `tsc -b`. Gotchas found: `import 'node:fs'` fails tsc (no
`@types/node` in tsconfig.app); passing the `URL` object fails at runtime (jsdom's `URL` is not Node's);
`?raw`/`?inline` CSS imports return the proxy/empty string under Vitest. Keep the helper local to
`LoadingLine.test.tsx` (only consumer).

## Tests (Vitest + RTL + user-event; `describe`/`it` from 'vitest', no globals)

Ticket "Tests:" bullet → cases. Strings asserted with `toBe` on exact `textContent` (U+00B7 middle dot).

StatusLine.test.tsx — helper `halves()` returns `[left, right]` textContent of `getByRole('status').children`:
- S1 idle → `['', '']`.
- S2 loading without previous → `['Loading', '']`.
- S3 loading with `previous: searchResponse` → `['Loading', 'Page 1 of 2']`.
- S4 success `searchResponse` → `['12 results · ranked for $450,000', 'Page 1 of 2']`.
- S5 success, no budget (spread: `targetBudget: null, budgetWeight: 0, recencyWeight: 1`; the real
  unfiltered search) → `['12 results · newest first', 'Page 1 of 2']`.
- S6 success singular (spread: `items: [{ ...listingA7, rank: 1 }]`, `totalItems: 1, totalPages: 1`,
  `city: 'Chantilly'`; the real search `targetBudget=450000&city=Chantilly`) → `['1 result · ranked for $450,000', 'Page 1 of 1']`.
- S7 success page past end (`page: 3`) → right `Page 3 of 2`.
- S8 empty `emptySearchResponse` → `['0 results', '']`.
- S9 invalid with previous `searchResponse` → same as S4; invalid with previous `emptySearchResponse` → `['0 results', '']`.
- S10 invalid without previous → `['', '']` (error object from `validationErrorBody` shape: `{ kind: 'validation', … }`).
- S11 error with previous and without (`{ kind: 'network', message }`) → `['', '']` both.
- S12 role: renders one `status` region with exactly two halves.

Pager.test.tsx (`onChange = vi.fn()`; `userEvent.setup()`):
- P1 renders `navigation` "Pagination" with buttons in order Previous, `Page 1`…`Page 3`, Next; textContent
  exactly `‹ Previous`, `1`, `2`, `3`, `Next ›` (glyphs are text: no svg/img inside).
- P2 `aria-current="page"` only on the current page.
- P3 bounds: page 1 of 3 → Previous disabled, Next enabled; 3 of 3 → Next disabled; 1 of 1 → both disabled.
- P4 disabled ends not focusable: page 1 of 2, `user.tab()` ×3 visits `Page 1`, `Page 2`, Next; Previous never
  focused. Page 2 of 2: tabbing visits Previous, `Page 1`, `Page 2`, never Next.
- P5 callbacks: Next → `onChange(3)` from page 2 of 3; Previous → `onChange(1)`; `Page 3` → `onChange(3)`;
  keyboard: focus Next + `{Enter}` → called (keyboard-only, SPEC §7.6).
- P6 no-ops: clicking the current page and clicking a disabled end never call `onChange`.
- P7 page past end (3 of 2): no `aria-current`, Next disabled, Previous → `onChange(2)`.
- P8 hidden at zero pages: `container` is empty and there is no `navigation` role.

LoadingLine.test.tsx:
- L1 active → `progressbar` named `Loading`; root has `styles.track` + `styles.active`; one `styles.segment` child.
- L2 inactive → no progressbar (`queryByRole('progressbar', { hidden: true })` null); root still renders with
  `styles.track`, `aria-hidden="true"`, no `styles.active`, no segment (same box ⇒ no layout shift).
- L3 motion (CSS source): `.track` height `2px`; `.active` background `var(--line)`; `.segment` `width: 40%`,
  `background: var(--oxblood)`, `animation: slide 1.2s linear infinite`; keyframes `translateX(-100%)` → `translateX(250%)`.
- L4 reduced-motion branch (CSS source): an `@media (prefers-reduced-motion: reduce)` block whose `.segment`
  rule sets `width: 100%` and `animation: none`.
- L5 oxblood only in `.segment` (exactly one `var(--oxblood)` in the file) and no hex literal.

Acceptance mapping:
- "Strings match §5.3/§7.2 verbatim" → S1–S11, P1, L1 (exact `toBe`).
- "Only oxblood uses are pager text buttons, current page, loading segment" → L5 + review grep:
  `grep -n oxblood frontend/src/components/{Pager,StatusLine,LoadingLine}/*.module.css` shows only Pager
  `.button` color/hover/focus/current rules and LoadingLine `.segment`; StatusLine none;
  `grep -nE '#[0-9a-fA-F]{3,8}'` on the three CSS files is empty.
- "Tests and lint green; no emoji" → gate below + `perl -CSD -ne 'print "$ARGV:$.\n" if /[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}\x{FE0F}]/'`
  on the nine files prints nothing (macOS grep has no `-P`; the only non-ASCII characters are `·`, `‹`, `›`).
- SPEC §8 "each component: renders from props, interactions, roles/labels, copy verbatim" → all of the above.

## Gate and validation

- Gate: from the worktree root `make check` (lint-api, lint-web = eslint + `tsc -b` + prettier --check,
  test-api, test-web) and `cd frontend && npm run build`, on the exact branch head. Run `npx prettier --write`
  on the nine files before committing. `node_modules` is already installed in this worktree (planner ran `npm ci`).
- Validation: **N/A until BC-12 wires these into SearchPage** (no running surface renders them yet; PORTS=none).
  Validator records `lattice comment BC-11 --role validation` with a one-line justification and cites the
  output of `cd frontend && npx vitest run src/components/Pager src/components/StatusLine src/components/LoadingLine`
  (test names + pass count, ~25 tests) at the branch head SHA.

## Shared files (run-state table)

None touched. `src/format/*` and `src/test/fixtures.ts` are consumed only; no new formatter is needed
(`formatResultSummary` and `formatPageOf` already cover §5.3; `Loading` is a §7.2 literal). No `App.tsx`,
styles, tokens, Makefile or SPEC changes. Commits never include `.lattice/`.

## Risks and hand-offs

- **Resting track** is transparent (see LoadingLine); owner may prefer a visible grey track — one-line swap.
- Parallel text buttons (BC-9 Clear, BC-10 Try again) may choose hover color/tracking differently;
  this plan follows §2.1 (oxblood-deep hover) and §3.2 (+0.02em on buttons). BC-14 reconciles if they differ.
- BC-12: two polite live regions (StatusLine `role="status"` and the results region) may both announce;
  move focus after `onChange`; keep LoadingLine mounted; query page buttons by `Page N`.
- `Page 3 of 2` on a hand-edited past-end URL is accurate but odd; Pager's Previous recovers in one click.
- CSS-source tests are coupled to the stylesheet text; regexes allow whitespace so prettier changes do not break them.

## Size

~560 LOC: Pager ~55 tsx / ~50 css / ~150 test; StatusLine ~45 / ~20 / ~120; LoadingLine ~25 / ~25 / ~70.
Complexity: low–medium. One commit per component or one commit total, implementer's choice.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review: art_01M4W70T5RKM712EBW3C1DRCGG (PASS, four minors). All triaged obvious; none needs a human.

1. **Text-button recipe (Pager CSS `.button`).** Concern: plan used +0.02em tracking and an `oxblood-deep`
   hover color; BRANDING §5.8's Text button row and the demo `.link` (the reference implementation) have
   neither. Resolution: follow §5.8 + demo exactly. `.button`: Figtree 13/1 (`--text-button`/`--leading-button`),
   weight 500, `color: var(--oxblood)`, **no `letter-spacing`**, no background/border/padding, `white-space: nowrap`.
   Hover (`:hover:not(:disabled, [aria-current='page'])`): **underline only**, `text-underline-offset: 3px`,
   **no color change**. Focus-visible: 2px `var(--oxblood)` outline, offset 3px. Disabled: `var(--ink-3)`,
   `cursor: default`, no underline on hover. Current page unchanged (oxblood, 500, 1px underline, offset 3px).
   `oxblood-deep` is not used anywhere in BC-11. (Sections: Pager CSS, Risks bullet 2.) The delegator relays
   this recipe to BC-9/BC-10 via the Orchestrator.
2. **Resting LoadingLine track.** Keep transparent at rest (2px box kept; `background: var(--line)` only on
   `.active`), per the demo. Add a CSS comment above `.track` stating this and the one-line flip. The delegator
   records a board comment now; the completion comment repeats it. (Section: LoadingLine.)
3. **S10/S11 error objects.** Build them explicitly, type-only imports from `src/api/types`:
   `const invalid: ValidationError = { kind: 'validation', message: validationErrorBody.error.message, fields: validationErrorBody.error.fields ?? [] };`
   and `const failure: ApiError = { kind: 'network', message: 'Request failed.' };`. No `Response`/`parseErrorResponse`.
   (Section: Tests, S10/S11.)
4. **CSS-source tests (L3–L5) hardening.** (a) Assert the read stylesheet is non-empty before any regex.
   (b) Anchor L4 to the `.segment` rule inside the `@media (prefers-reduced-motion: reduce)` block (one regex
   spanning the block with `[\s\S]`/`[^}]` so a stray `animation: none` elsewhere cannot satisfy it); it must
   check both `width: 100%` and `animation: none` inside that rule. (c) A short comment in the test file:
   these cases verify the shipped stylesheet source; the visual behavior is proven in BC-12's browser validation.
   (Section: Tests, LoadingLine.)
