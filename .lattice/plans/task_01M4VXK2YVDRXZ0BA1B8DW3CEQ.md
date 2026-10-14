# BC-9 plan: FilterBar (seven fields, client rules with exact copy, server error mapping, Search and Clear)

Planner: agent:opus-5.5-planner. Worktree `/Users/forrest/Code/brillio-worktrees.noindex/bc-9-filterbar`, branch
`feat/BC-9-filterbar`, anchor `feat/BC-8-frontend-core` @43e8824 (BC-8 in review; merge `master` once it lands).
Contract: SPEC §7.4 (FilterBar), §7.1, §5.3, §5.5, §7.6, §8, §13.1; BRANDING §2.2, §3.2, §4.4, §5.2, §5.8, §5.9,
§6, §7.2, §8; look from `docs/branding-demo.html` lines 27-41 and 94-100 (`.filters`, `.field`, `.val`, `.msg`,
`.actions`, `.btn`, `.link`).

## Scope

In: `frontend/src/components/FilterBar/` only, plus one SPEC §7.4 sentence (below). Props exactly SPEC §7.4:
`FilterBar({ value: SearchParams; cities: string[]; loading: boolean; serverErrors?: FieldError[]; onSearch(next: SearchParams): void; onClear(): void })`.
Out: `App.tsx` and page wiring (BC-12), shell styles and tokens (BC-7/BC-14), LoadingLine (BC-11), `src/format/*`
and `src/test/fixtures.ts` (nothing missing; see Shared files), any new dependency.

## Approach

### Copy (verbatim; one `as const` table in the model file, the only place these strings live)
- Labels (BRANDING §7.2 + SPEC §7.4), in DOM order: `Min price` · `Max price` · `Min bedrooms` · `City` · `Keyword` · `Target budget` · `Status`.
- Placeholders: `Any` · `Any` · `Any` · `Any` · `Search descriptions` · `Any` · `Any`. Inputs use the `placeholder`
  attribute; the two selects use a first `<option value="">Any</option>`.
- Status options: `Any` / `Active` / `Pending` / `Sold`, values `''`/`active`/`pending`/`sold`, built from BC-8's
  `LISTING_STATUSES` with a `Record<ListingStatus, string>` label map (a new status fails `tsc`).
- Buttons: `Search` (type submit), `Clear` (type button). Form accessible name `Filters` (the name BC-7's shell
  placeholder already gives this region); DOM label text is sentence case and CSS uppercases it, so
  `getByLabelText('Min price')` works for BC-12.
- Messages. BRANDING §7.2 verbatim: `Min price must be less than max price.`, `Min bedrooms must be 0 or more.`,
  `Target budget must be more than $0.` BRANDING has no string for the rest, so the client uses the exact
  sentences BC-5 pinned for the server (BC-5 plan rules table; same SPEC §5.5 voice), so a value reads the same
  whether the client or the API catches it: `Min price must be 0 or more.`, `Max price must be 0 or more.`,
  `Min price must be a number.`, `Max price must be a number.`, `Min bedrooms must be a number.`,
  `Min bedrooms must be a whole number.`, `Target budget must be a number.`

### State model
- `FieldName = 'minPrice' | 'maxPrice' | 'minBedrooms' | 'city' | 'keyword' | 'targetBudget' | 'status'`;
  `Drafts = Record<FieldName, string>`. Typing only changes drafts; nothing fires a request.
- Seeding: `draftsFrom(value)`: numbers `String(n)`, text as is, `status` = the single status when
  `value.status` has exactly one entry, else `''` (Any). Seed key `filterKey(value)` = BC-8's `buildSearchQuery`
  of the seven filter fields only (page and pageSize excluded), so an equal new object, a page change or a
  pageSize change never wipes typing, and `status` order is canonical.
- Re-seed with keyed local state, no effect: `useState({ key, drafts, revealed })`; each render uses
  `local.key === key ? local : seeded(value)`, and every update starts from that current snapshot and stores the
  current key. (This avoids setState in render and in effects; react-hooks 7 `set-state-in-render` and
  `set-state-in-effect` stay quiet.) A re-seed resets `revealed` to the fields with a non-empty seed, so invalid
  values arriving by URL are marked at once.
- `focused: FieldName | null` (set on focus, cleared on blur). Server-error dismissals live in
  `{ key: serverKey, fields: Set<FieldName> }` with `serverKey = JSON.stringify(serverErrors ?? [])`; a different
  key means nothing dismissed (content key, so a parent that rebuilds the array each render cannot resurrect a
  dismissed mark).
- Ids: one `useId()` per instance; `${base}-${field}` for controls, `${base}-${field}-msg` for messages. Unique per
  instance, label text stable.

### Parsing and display (private model helpers)
- `parseNumber(text)`: trim, delete every `$` and `,`, then accept `/^-?(\d+\.?\d*|\.\d+)$/` -> finite number;
  `''` after trim -> empty; anything else (letters, `1e5`, `+5`, inner spaces) -> not a number. So `$400,000`
  round-trips, `5.` and `.5` parse, `-5` parses (and fails the sign rule with the clearer message), `-0` -> 0.
- Money fields (Min price, Max price, Target budget): focused shows the raw draft (digits while typing); blurred
  shows `formatPrice(n)` from `src/format/money.ts` when the draft parses to a whole number `>= 0`, else the raw
  draft (a fraction is never rounded into a different-looking filter; an invalid draft stays visible as typed).
- On blur every numeric field (including Min bedrooms) normalises a parseable draft to `String(n)` so a refocus
  shows plain digits (`$400,000` -> `400000`, `2.0` -> `2`). Min bedrooms is never `$`-formatted (`2`).
- Controls are `type="text"` with `inputMode="decimal"` (money) / `inputMode="numeric"` (bedrooms),
  `autoComplete="off"`. Not `type="number"`: it cannot display `$400,000`, renders a spinner glyph (BRANDING
  §5.10 allows only `‹ › ▾`), and hides bad text from script (`value === ''`). "Step 1" is enforced as the
  whole-number rule.

### Client rules (pure `validate(drafts)`; first failing rule per field wins)
| Field | Order of rules -> message |
|---|---|
| Min price / Max price | not a number -> `<Label> must be a number.`; `< 0` -> `<Label> must be 0 or more.` |
| cross | both prices parse with no own error and min `>` max (equal is valid, SPEC §5.3) -> `Min price must be less than max price.` on Max price, both fields marked |
| Min bedrooms | not a number; `< 0` -> `Min bedrooms must be 0 or more.`; not whole -> `Min bedrooms must be a whole number.` (BC-5's order) |
| Target budget | not a number; `<= 0` (negatives too) -> `Target budget must be more than $0.` |
| City, Keyword, Status | none client-side (length and enum are the server's; mapped below) |

- Timing (chosen): rules run on every change and drive `canSubmit` at once; a field's message and mark show
  once the field is revealed: on blur, on any Enter attempt (reveals all), or on seed. The cross error shows when
  both price fields are revealed. Rationale: live messages flash `must be a number` on a bare `-` and the
  min-above-max message on every keystroke of a max price; reveal-on-blur keeps the bar quiet while typing, and
  Enter on an invalid bar explains itself instead of doing nothing.
- `canSubmit = !loading && no client errors` (revealed or not). Search is `disabled={!canSubmit}`. Server errors
  do not disable Search: the API is the authority and may be asked again.

### Server error mapping (`mapServerErrors(serverErrors)`)
- `field` in FieldName -> that field's server mark (first entry per field wins), displaying the server `message`
  verbatim (BC-5 sends the same strings; an empty message still marks the field, with no description).
- `maxPrice` + `min_gt_max` -> message under Max price and Min price marked too, exactly like the client cross rule.
- Any other `field` (`pageSize`, `page`, `unknown_param` names such as `minprice`) -> unmapped; their messages,
  deduplicated, joined with one space, render as one `<p role="alert">` that is the bar's last grid row (full
  width, Meta 13 in `ink`: it is not a field's message, so oxblood job 4 does not apply; §2.3 "errors are stated
  in ink"). Rendered only when non-empty.
- Display per field: revealed client error, else undismissed server error. `aria-invalid="true"` when the field
  shows a message or is Min price in a min/max conflict; `aria-describedby` = its own message id, or for Min price
  in a conflict, Max price's message id. Neither attribute is present on a valid field.
- Clearing server marks: editing field X dismisses X's server mark; editing Min price also dismisses a
  `min_gt_max` on Max price. A new `serverErrors` content (e.g. undefined while loading, then a new 400) resets
  dismissals. Clear dismisses all.

### Submit, Enter, Clear
- One `submit()` used by the form's `onSubmit` (Search click) and by an `onKeyDown` on every input and select
  (Enter without modifiers and not `isComposing`: `preventDefault()` then `submit()`). Handling Enter on the
  fields, not via implicit submission, is needed because browsers skip implicit submission when the submit
  button is disabled, and selects never submit implicitly; "Enter in any field submits" covers both selects.
  `preventDefault` also stops the native implicit submit, so one Enter = one call.
- `submit()`: loading -> nothing. Client errors -> reveal every field, no call. Else `onSearch(toSearchParams(drafts))`.
- Payload: only fields with a non-empty draft, keys omitted otherwise (`exactOptionalPropertyTypes`): numbers
  parsed (bedrooms an integer), `city`/`keyword` trimmed, `status: [draft]` unless Any. Never `page` or
  `pageSize`: BC-8's `setSearch` drops `page` and re-reads `pageSize` from the live URL (SPEC §7.1), so carrying
  them would double-handle.
- `value.status` with several entries (hand-made URL; the UI never writes one) shows Any and is omitted on the
  next submit. Same principle as BC-8's `parseSearchParams`, which drops what no field can show.
- City: the select's value is the `cities` entry matching the draft case-insensitively, else the draft itself,
  appended as an extra option after the list when non-empty. That covers a URL city before `useCities` resolves
  (it returns `[]` until loaded), after it fails, and a city with no listings (`city=Arlington`, the empty-state
  fixture). Submit sends what the select shows.
- Clear: calls `onClear()` and resets drafts to empty, `revealed`, and dismisses all server marks immediately.
  Needed because `clear()` is a no-op navigation when the URL is already empty, so no re-seed would arrive. Clear
  stays enabled while loading.

### Accessibility (SPEC §7.6)
`<form role="search" aria-label="Filters" onSubmit>`; each control has a real `<label htmlFor>`; messages are
`<p id>` referenced by `aria-describedby`; `aria-invalid` as above; Search uses native `disabled`; Clear is a
`<button type="button">`. The `▾` is CSS generated content on a wrapper span, `content: "▾" / ""` so it is not
announced. Fields stay enabled while loading (typing never sends).

### CSS (`FilterBar.module.css`, tokens only via `var(--…)`, no hex, no transitions)
- `.bar` = the form, a direct child of `.shell` (shell.module.css contract): `padding-inline: var(--inline-pad)`,
  `padding-block: 20px 18px`, `border-bottom: 1px solid var(--line)`, `display: grid`,
  `grid-template-columns: repeat(auto-fit, minmax(130px, 1fr))`, `gap: 20px`, `align-items: start`.
- Layout reconciliation: §4.4's "one row" was written for six fields; §5.2's seventh-field rule says the bar
  wraps at 1040px before a field drops below 130px. Eight cells (seven fields + actions) need 1180px, more than
  the 1040px content width, so above 760px the bar is one row when it fits and otherwise wraps: at full width row
  one is the demo's six fields in demo order, row two is Status then Search and Clear. No breakpoint is invented.
- `.field`: flex column, `gap: var(--gap-label)`, `min-width: 0`. `.label`: §3.2 Label (`--text-label`,
  `--leading-label`, 500, `--tracking-label`, uppercase, `ink-2`).
- `.control` (inputs and selects): `appearance: none`, transparent, `border: 0`, `border-bottom: 1px solid
  var(--ink-2)`, `padding: 6px 0`, Body 15/22, `ink`, tabular figures, `width: 100%`, ellipsis, `outline: none`
  (the 2px underline is the focus indicator, 9.35:1). Box height 35px in every state.
  `:focus` and `[aria-invalid='true']`: `border-bottom: 2px solid var(--oxblood)`, `padding-bottom: 5px`.
  `:disabled`: `color: var(--ink-3)`, underline `var(--line)` (anatomy completeness; nothing disables a field
  in v1). `::placeholder`: `ink-2`, `opacity: 1`. Select "Any" renders in `ink` as in the demo (`.val.sel`).
- `.select` wrapper: `position: relative`; `::after` `▾`, `ink-2`, 12px, right 0, on the 22px value line,
  `pointer-events: none`; select gets `padding-right: 16px`.
- `.message`: `margin: 6px 0 0`, Meta 13/18, `oxblood` (demo `.msg`).
- `.actions`: flex, `gap: 18px`, `align-items: flex-end`, `padding-bottom: 1px`, fixed height = label line +
  gap + control box (`calc(var(--leading-label) + var(--gap-label) + 35px)`). Identical to the demo's
  `align-items: end` at rest, but a message under a field no longer pushes that field's underline out of line.
- `.search` (§5.8 primary): oxblood fill, paper text, 40px, `padding: 0 20px`, Button 13/1 500 +0.02em,
  `white-space: nowrap`; `:hover:not(:disabled)` `oxblood-deep`; `:focus-visible` 2px oxblood outline offset 2px;
  `:disabled` `stone` fill, `ink-3` text, `cursor: default`.
- `.clear` (§5.8 text button): Button type, `oxblood`, no background or border, padding 0; hover underline offset
  3px; `:focus-visible` outline offset 3px. Letter spacing +0.02em per §3.2/§3.3 (the demo's `.link` omits it;
  0.26px).
- `@media (max-width: 760px)`: `.bar` `grid-template-columns: 1fr 1fr; padding-block: 16px`; `.actions`
  `grid-column: 1 / -1; height: auto; justify-content: space-between`; `.search` `flex: 1` (demo lines 98-100).
  Gutter 16px comes from tokens.css.
- Motion (§6): none; underline width and hover colour change instantly, so no reduced-motion block is needed.
- Oxblood jobs used (§2.2): 1 Search fill (`oxblood-deep` hover per §2.1), 2 Clear, 3 focus underline and button
  focus rings, 4 invalid underline and field message. Nothing else.

## Files (worktree-relative)
- `frontend/src/components/FilterBar/FilterBar.tsx`: component and `FilterBarProps`.
- `frontend/src/components/FilterBar/FilterBar.module.css`.
- `frontend/src/components/FilterBar/filterBarModel.ts`: private pure helpers (copy table, `FieldName`, `Drafts`,
  `draftsFrom`, `filterKey`, `parseNumber`, `displayValue`, `normalizeDraft`, `validate`, `toSearchParams`,
  `mapServerErrors`). Justified: the rules are the bulk of the logic and unit-test faster than through the DOM;
  the component stays readable. Not exported outside the folder by convention.
- `frontend/src/components/FilterBar/FilterBar.test.tsx`, `frontend/src/components/FilterBar/filterBarModel.test.ts`.
- `docs/SPEC.md` §7.4 (and the matching clause of §7.2), one in-branch edit (below).

## Tests (RTL + `@testing-library/user-event` 14.6.8, already a devDependency; fixtures `cities`, `validationErrorBody`)
`FilterBar.test.tsx`, one `userEvent.setup()` per test:
| Test | Covers |
|---|---|
| renders seven labelled fields in order with the verbatim labels | labels verbatim; §7.4 seventh field |
| uses the BRANDING placeholders on inputs and Any first in both selects | placeholders verbatim |
| City lists Any then the cities prop; a URL city missing from the list is an extra selected option (also with `cities=[]`) | §5.2 City, SPEC §7.2 |
| Status offers Any, Active, Pending, Sold | §7.4, §13.1 |
| is a search landmark named Filters with Search (submit) and Clear (button) and no emoji | roles; no emoji |
| seeds every field from value, money formatted, single status selected | seeding |
| re-seeds on a different value; keeps typing on an equal object or a page/pageSize-only change | re-seed rule |
| several statuses in value show Any and are omitted on submit | status decision |
| it.each(Min price, Max price, Target budget): digits while focused, `$400,000` on blur, digits again on refocus | blur formatting |
| accepts a typed `$400,000` and submits 400000; Min bedrooms is never `$`-formatted | parse round-trip |
| min above max marks both, one message inside Max price's field, both described by it; equal is valid | min_gt_max placement |
| it.each negative Min/Max price -> `… must be 0 or more.` under that field | negative prices |
| it.each non-numeric money and bedrooms -> `… must be a number.` | not a number |
| Min bedrooms: `-1` -> `Min bedrooms must be 0 or more.`; `2.5` -> `Min bedrooms must be a whole number.` | bedrooms rules |
| it.each(0, -5) Target budget -> `Target budget must be more than $0.` | budget rule |
| message appears on blur, not while typing, yet Search is disabled at once | timing |
| Enter on an invalid bar reveals every message and does not call onSearch | invalid Enter |
| correcting the value removes the mark and re-enables Search | recovery |
| invalid values seeded from value are marked immediately | URL seed |
| fixture 400: min_gt_max marks both price fields, unmapped `pageSize`/`minprice` render one alert line | server mapping |
| it.each server errors on minPrice, minBedrooms, city, keyword, targetBudget, status mark that field with the server message | every mapped field |
| editing a field clears its server mark (Min price clears min_gt_max); new serverErrors restore marks; server marks leave Search enabled | server clear rules |
| Search sends parsed numbers, trimmed city and keyword, `status: ['pending']`, nothing else | payload shape |
| omits empty fields, Any status, page and pageSize (`onSearch({})` from `{page: 3, pageSize: 20}`) | payload omissions |
| it.each(seven labels) Enter submits exactly once | Enter submits |
| while loading Search is disabled, fields stay enabled, Enter sends nothing | disabled while loading |
| Clear calls onClear and empties drafts and marks at once (also when value is already empty) | Clear |
| valid fields carry no aria-invalid or aria-describedby; two instances get distinct ids with the same labels | aria wiring, ids |

`filterBarModel.test.ts`: `parseNumber` table (`400000`, `$400,000`, ` 400000 `, `5.`, `.5`, `-0`, `-5`, `''`, `abc`,
`1e5`, `+5`, `4 00`); `validate` precedence (number, sign, whole; cross skipped when a side has its own error);
`toSearchParams` key omission; `filterKey` ignores page, pageSize and status order; `mapServerErrors` first entry
per field, empty message, unmapped dedupe.

Gate: `make check` (eslint, `tsc -b --noEmit`, prettier, both test halves) and `cd frontend && npm run build`.

## Validation
N/A until BC-12 wires the page (component-only ticket, `.lattice/context.md`). The validator records
`--role validation`: one-line justification plus the output of
`cd frontend && npx vitest run src/components/FilterBar --reporter=verbose` (every case above listed, all passing)
and the `make check` / `npm run build` result on the exact head. Optionally a jsdom render description. No browser,
no ports.

## Shared-file handling (run-state table)
- `src/format/*`: consumed (`formatPrice`), not changed. The input parser is private: it parses user input, not a
  §7.1 format, and BC-8 has no parser to fork (its URL reader is a different grammar that accepts exponents).
- `src/test/fixtures.ts`: consumed (`cities`, `validationErrorBody`), not changed. Per-field FieldErrors are inline.
- `src/api/types.ts`, `client.ts` (`buildSearchQuery`), `LISTING_STATUSES`: consumed, not changed.
- Not touched: `App.tsx`, `tokens.css`, `global.css`, `shell.module.css`, `Makefile`, `package.json`, lockfile.
- `docs/SPEC.md` (union, contract deviation rule; delegator adds the lattice comment):
  SPEC edit: §7.4, FilterBar behavior paragraph, add: "A `value.city` that `cities` lacks (not loaded yet, failed,
  or no listings) is shown as an extra option so the select always shows the applied filter. A `value.status`
  with one status selects it; none or several show Any, and submitting then sends no status." And in §7.2 change
  "offers only "Any"" to "offers only "Any" (plus the URL's city, §7.4)". Why: §7.2 says only Any on failure,
  which would show Any while a city filter is applied.
- `.lattice/` is never committed from the branch.

## Risks / open points
1. Ticket acceptance says "no oxblood outside §2.2 jobs 1, 3 and 4", but §2.2 job 2 names Clear explicitly and §5.8
   and the demo set Clear in oxblood. BRANDING is design truth: jobs 1-4 are used. Not a design change; the
   delegator should note it in a lattice comment so the reviewer does not flag Clear.
2. At the 1040px content width the bar is two rows (six fields; Status + actions), per §5.2 rather than §4.4's
   six-field "one row". If the owner wants another arrangement, that is a BRANDING §9 change -> needs-human; no
   BRANDING edit here.
3. Multi-status URLs show Any and widen on the next submit. Alternative if a human prefers: a dynamic extra option
   ("Active, Pending"). Not blocking.
4. Client messages beyond §7.2 copy BC-5's pinned server sentences; if BC-5's review changes them, both sides must
   change together (BC-12's integration test would catch the mismatch).
5. BC-8 is unlanded; if its review renames `formatPrice`, `buildSearchQuery` or `LISTING_STATUSES`, follow on the
   merge of `master`.
6. Actions content is about 135px; at the exact wrap width its 130px cell overflows a few px into the gutter.
   Invisible; accepted.

No needs-human question blocks implementation.

## Size estimate
FilterBar.tsx ~230 LOC, filterBarModel.ts ~170, FilterBar.module.css ~150, FilterBar.test.tsx ~480,
filterBarModel.test.ts ~140, SPEC +3 lines. About 1,170 LOC, two thirds of it tests.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review: art_01M4W6JVK8PG4DFPD0558XZM7W (PASS, 2 major, 6 minor). Triage by agent:opus-5.5-delegator.

0. **Anchor (minor 8, obvious).** BC-8 landed (`ccd1669`); the branch was fast-forwarded to `master` @e5f49bb
   before implementation. REVIEW_BASE is now `master`. Bring later master moves in with `git merge master` only.
1. **Wide layout (major 1, obvious).** Keep the CSS (`repeat(auto-fit, minmax(130px, 1fr))`, gap 20px, actions
   auto-placed as the last grid item, no `grid-column` on `.actions` above 760px). Replace the "Layout
   reconciliation" paragraph with the true outcome: at the 1040px content width, row one is all seven fields
   (~131px each, the demo's field width) and row two is the actions cell alone, left-aligned under Min price
   (BRANDING §4.5 "everything left-aligned"; DOM order = visual order = tab order). Between ~880px and 1029px
   content: six fields, then Status + actions. Narrower: more rows, never a field under 130px. At <= 760px: two
   columns, actions full width per the demo. Reviewer acceptance at 1040px: "seven fields on row one, Search and
   Clear on row two at the left".
2. **Server-mark dismissals (major 2, obvious).** `submit()` also clears the dismissal set (the API is being asked
   again), in addition to the content key. New test: dismiss a server mark by editing, submit, receive identical
   `serverErrors` content -> the mark is back.
3. **`exactOptionalPropertyTypes` (minor 3, obvious).** Declare `serverErrors?: FieldError[] | undefined` so BC-12
   can pass `undefined` directly. Same meaning as SPEC §7.4.
4. **Disabled Search with unrevealed errors (minor 4, obvious).** Add a `pointerdown` handler on the `.actions`
   wrapper that reveals every field when `!canSubmit && !loading`. Test it with `fireEvent.pointerDown` on the
   wrapper if user-event refuses to point at a disabled button. Best effort across browsers (Enter and blur
   remain the primary reveal paths).
5. **`type="text"` numeric inputs (minor 5, obvious).** Recorded deviation from BRANDING §5.2 "number input":
   add to the SPEC §7.4 edit: "Numeric fields are text inputs with a numeric `inputMode`; `type=number` cannot
   show the §7.1 price format. Step 1 is enforced as the whole-number rule." Delegator records a lattice comment.
6. **Oxblood jobs (minor 6, obvious).** Acceptance reads "no oxblood outside BRANDING §2.2 jobs 1-4" (Clear is
   job 2 by name). The ticket line "jobs 1, 3 and 4" is a slip; delegator records a lattice comment.
7. **Multi-status URL (minor 7, evolutionary -> skipped).** Keep as planned: several statuses show Any and are
   omitted on the next submit (documented in the SPEC §7.4 edit). The UI never writes such a URL.
8. **Implementer notes (minor 8, obvious).** The form's `onSubmit` calls `preventDefault()`. `parseNumber` is
   deliberately stricter than the server (`1e5`, `+5` rejected client-side); say so in a one-line code comment.
9. **Orchestrator note 03:08Z.** BC-5 adds field-error code `malformed_query` on the pseudo-field `query`. It is not
   a FieldName, so it falls through to the unmapped line under the bar (no field mark). Add it to the unmapped
   server-mapping test (`{ field: 'query', code: 'malformed_query', message: <any> }` renders in the alert line).
   BC-8's exports and usage notes are in its completion comment (`lattice show BC-8`).
