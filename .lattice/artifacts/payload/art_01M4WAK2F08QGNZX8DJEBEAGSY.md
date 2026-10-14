# BC-9 Validation: FilterBar (seven fields, client rules, server error mapping, Search and Clear)

- Ticket: BC-9 (task_01M4VXK2YVDRXZ0BA1B8DW3CEQ)
- Validator: agent:opus-5.5-validator
- Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-9-filterbar
- Branch: feat/BC-9-filterbar
- HEAD: 13ae060fdab81e47bc29ef2f3c6ba46c82dfc33b (verified with `git rev-parse HEAD`; working tree clean before and after)
- Base: master 76d6a97bf7f18da629e9294b139ae65147212ca3 (= `git merge-base HEAD master`)
- Code review at this head: art_01M4WADP7WCHKFV2AKQQBQJFVG (PASS)
- Date: 2026-10-14T04:25Z (UTC)

## Justification (N/A for end-to-end)

N/A: FilterBar is a component-only ticket with no rendered surface until BC-12 wires SearchPage (`frontend/src/App.tsx` line 10 still renders the BC-7 placeholder `<div role="search" aria-label="Filters" />` and nothing under `frontend/src` outside `components/FilterBar/` imports FilterBar; the production build bundles 20 modules without it), so it is validated by the component test suite under jsdom (113/113 below) plus the full gate (`make check`, `npm run build`), all green on the exact head.

## Commands run

No dev server, browser or port was used (PORTS=none). Logs were kept in the session scratchpad, not in the worktree.

| Command (cwd) | Exit | Result |
|---|---|---|
| `npx vitest run src/components/FilterBar --reporter=verbose` (frontend/) | 0 | Test Files 2 passed (2); Tests 113 passed (113) |
| `make check` (worktree root) | 0 | golangci-lint 0 issues; eslint --max-warnings 0 clean; `tsc -b --noEmit` clean; prettier clean; go test -race ok; vitest 22 files / 382 tests passed |
| `npm run build` (frontend/) | 0 | `tsc -b --noEmit && vite build`: `✓ built in 95ms` |
| `git status --porcelain` (worktree root, after all runs) | 0 | empty (`frontend/dist/` is gitignored build output) |

### Gate summary lines (verbatim)

`make check`:
```
cd backend && golangci-lint run
0 issues.
> eslint . --max-warnings 0
> tsc -b --noEmit
> prettier --check .
Checking formatting...
All matched files use Prettier code style!
cd backend && go test -race ./...
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	(cached)
ok  	brillio/backend/internal/domain	(cached)
ok  	brillio/backend/internal/httpapi	(cached)
ok  	brillio/backend/internal/store	(cached)
ok  	brillio/backend/internal/store/seed	(cached)
cd frontend && npm test
> vitest run
 Test Files  22 passed (22)
      Tests  382 passed (382)
```

`npm run build`:
```
vite v8.3.4 building client environment for production...
✓ 20 modules transformed.
dist/index.html                   0.77 kB │ gzip:  0.41 kB
dist/assets/index-Bi9deQIM.css    2.56 kB │ gzip:  1.06 kB
dist/assets/index-C_W2cHaM.js   220.30 kB │ gzip: 68.89 kB
✓ built in 95ms
```

(The backend Go packages were reported `(cached)`: this branch does not touch `backend/`; the diff vs master is `docs/SPEC.md` plus the five files under `frontend/src/components/FilterBar/`.)

## FilterBar test cases (verbatim, vitest v5.0.4, `--reporter=verbose`)

filterBarModel.test.ts: 51 cases. FilterBar.test.tsx: 62 cases. 113 total, 113 passed, 0 failed, 0 skipped.

```
 RUN  v5.0.4 /Users/forrest/Code/brillio-worktrees.noindex/bc-9-filterbar/frontend

 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "400000" as {"kind":"number","value":400000} 1ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "$400,000" as {"kind":"number","value":400000} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads " 400000 " as {"kind":"number","value":400000} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "5." as {"kind":"number","value":5} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads ".5" as {"kind":"number","value":0.5} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "2.0" as {"kind":"number","value":2} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "-0" as {"kind":"number","value":0} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "-5" as {"kind":"number","value":-5} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "" as {"kind":"empty"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "   " as {"kind":"empty"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "abc" as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "1e5" as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "+5" as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "4 00" as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "$" as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "-" as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "." as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > parseNumber > reads "9999…9999" (400-digit overflow string, elided here) as {"kind":"invalid"} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > finds nothing in empty drafts 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"minPrice":"abc"} gives {"minPrice":"Min price must be a number."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"minPrice":"-5"} gives {"minPrice":"Min price must be 0 or more."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"maxPrice":"abc"} gives {"maxPrice":"Max price must be a number."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"maxPrice":"-1"} gives {"maxPrice":"Max price must be 0 or more."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"minBedrooms":"two"} gives {"minBedrooms":"Min bedrooms must be a number."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"minBedrooms":"-1"} gives {"minBedrooms":"Min bedrooms must be 0 or more."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"minBedrooms":"2.5"} gives {"minBedrooms":"Min bedrooms must be a whole number."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"minBedrooms":"-1.5"} gives {"minBedrooms":"Min bedrooms must be a whole number."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"targetBudget":"lots"} gives {"targetBudget":"Target budget must be a number."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"targetBudget":"0"} gives {"targetBudget":"Target budget must be more than $0."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > {"targetBudget":"-5"} gives {"targetBudget":"Target budget must be more than $0."} 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > accepts the boundaries 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > flags min above max, but not equal prices 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > skips the min/max conflict when either price has its own error 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > validate > has no rules for city, keyword or status 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > toSearchParams > omits every empty field 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > toSearchParams > parses numbers, trims text and wraps the status 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > draftsFrom > writes numbers as plain digits and picks a single status 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > draftsFrom > writes tiny fractions without an exponent, so they seed a valid draft 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > draftsFrom > shows Any for several statuses and empties non-finite numbers 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > filterKey > ignores page, pageSize and status order 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > displayValue and normalizeDraft > formats a blurred whole money draft and leaves everything else as typed 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > displayValue and normalizeDraft > turns a readable numeric draft into plain digits on blur 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > cityChoice > matches a listed city case-insensitively 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > cityChoice > offers an unlisted city as an extra option 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > mapServerErrors > maps fields, keeps the first error per field and the empty message 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > mapServerErrors > reads min_gt_max on maxPrice as the conflict and joins unmapped messages once 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > resolveMarks > shows a client error only once its field is revealed 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > resolveMarks > puts the conflict under Max price and marks Min price, once both are revealed 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > resolveMarks > prefers a revealed client error to a server error 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > resolveMarks > drops the server conflict when either price is dismissed 0ms
 ✓ src/components/FilterBar/filterBarModel.test.ts > resolveMarks > keeps Min price own server message over the conflict mark 0ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > renders seven labelled fields in order with the verbatim labels 77ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > uses the BRANDING placeholders on inputs and Any first in both selects 20ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > lists Any then the cities prop in City 8ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > shows a URL city missing from the list as an extra selected option 9ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > shows a URL city missing from no list yet as an extra selected option 7ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > offers Any, Active, Pending and Sold in Status 7ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar copy and structure > is a search landmark named Filters with Search (submit) and Clear (button) and no emoji 12ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar seeding > seeds every field from value, money formatted, single status selected 7ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar seeding > re-seeds on a different value and keeps typing on an equal object or a page change 52ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar seeding > shows the filters, not old typing, on returning to earlier filters (A, B, A) 41ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar seeding > shows Any for several statuses and omits status on submit 12ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar seeding > marks invalid values seeded from value at once 6ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar money and number display > Min price shows digits while focused and undefined,000 on blur 50ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar money and number display > Max price shows digits while focused and undefined,000 on blur 51ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar money and number display > Target budget shows digits while focused and undefined,000 on blur 50ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar money and number display > accepts a typed $400,000 and submits 400000 58ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar money and number display > never formats Min bedrooms as money 27ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar money and number display > leaves a fraction or an invalid money value as typed on blur 54ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > marks both prices for min above max, with one message inside Max price 60ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > accepts equal prices 70ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > a negative Min price reads "Min price must be 0 or more." under that field 24ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > a negative Max price reads "Max price must be 0 or more." under that field 22ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > a non-numeric Min price reads "Min price must be a number." 25ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > a non-numeric Max price reads "Max price must be a number." 24ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > a non-numeric Min bedrooms reads "Min bedrooms must be a number." 26ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > a non-numeric Target budget reads "Target budget must be a number." 26ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > Min bedrooms -1 reads "Min bedrooms must be 0 or more." 24ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > Min bedrooms 2.5 reads "Min bedrooms must be a whole number." 27ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > Target budget 0 reads "Target budget must be more than $0." 23ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > Target budget -5 reads "Target budget must be more than $0." 23ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > shows the message on blur, not while typing, yet disables Search at once 26ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > reveals every message on Enter in an invalid bar and sends nothing 44ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > reveals the messages when the disabled Search is pressed 19ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > reveals nothing while Clear is pressed, so Clear stays put until the click lands 27ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar client rules > removes the mark and enables Search once the value is corrected 44ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > maps the fixture 400: min_gt_max on both prices, the rest on one line under the bar 8ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks minPrice with the server message 5ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks maxPrice with the server message 5ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks minBedrooms with the server message 5ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks city with the server message 5ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks keyword with the server message 5ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks targetBudget with the server message 7ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks status with the server message 4ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > marks a field for an empty server message without describing it 2ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > clears a server mark on edit; Min price clears min_gt_max; new errors restore marks 29ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar server errors > restores a dismissed server mark after a submit that gets the same errors 27ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > sends parsed numbers, the trimmed city and keyword and the status, nothing else 152ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > sends a chosen city 27ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > omits empty fields, Any status, page and pageSize 12ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in Min price submits exactly once 15ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in Max price submits exactly once 15ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in Min bedrooms submits exactly once 15ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in City submits exactly once 16ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in Keyword submits exactly once 15ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in Target budget submits exactly once 15ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > Enter in Status submits exactly once 16ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > ignores the Enter that ends an IME composition 4ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar submit > while loading disables Search, keeps fields enabled and sends nothing on Enter 37ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar Clear > calls onClear and empties drafts and field marks at once, keeping the unmapped line 46ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar Clear > empties typed drafts when value is already empty 44ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar ids > leaves valid fields without aria-invalid or aria-describedby 3ms
 ✓ src/components/FilterBar/FilterBar.test.tsx > FilterBar ids > gives two instances distinct ids with the same labels 6ms

 Test Files  2 passed (2)
      Tests  113 passed (113)
   Start at  00:24:17
   Duration  2.43s (tests 54%, environment 36%, setup 7%, import 2%, transform 1%)
```

Only edit to the paste: the one 400-digit `9999…` input in the `parseNumber` overflow case is shortened here; the case name and result are otherwise verbatim.

## Ticket Tests bullets mapped to passing cases

| Ticket Tests bullet | Passing case(s) (FilterBar.test.tsx unless marked "model") | Verdict |
|---|---|---|
| Labels and placeholders verbatim | "renders seven labelled fields in order with the verbatim labels" (DOM label text equals `Min price`, `Max price`, `Min bedrooms`, `City`, `Keyword`, `Target budget`, `Status`, in order, and control order matches); "uses the BRANDING placeholders on inputs and Any first in both selects" (`Any` x4, `Search descriptions`, `Any` first option in City and Status); "lists Any then the cities prop in City"; "offers Any, Active, Pending and Sold in Status" (values `''/active/pending/sold`) | PASS |
| Blur formatting | "{Min price, Max price, Target budget} shows digits while focused and $400,000 on blur" x3 (asserts `400000` focused, `$400,000` blurred, `400000` on refocus); "accepts a typed $400,000 and submits 400000"; "never formats Min bedrooms as money" (`2.0` -> `2`); "leaves a fraction or an invalid money value as typed on blur"; "seeds every field from value, money formatted…"; model "displayValue and normalizeDraft" x2 | PASS |
| Each invalid rule and its placement | "marks both prices for min above max, with one message inside Max price" (one message, inside Max price's field, not Min price's; both `aria-invalid`; both described by it; Search disabled); "accepts equal prices"; "a negative {Min,Max} price reads "… must be 0 or more." under that field" x2; "a non-numeric {Min price, Max price, Min bedrooms, Target budget} reads "… must be a number."" x4; "Min bedrooms -1 / 2.5" x2; "Target budget 0 / -5 reads "Target budget must be more than $0."" x2; timing: "shows the message on blur, not while typing, yet disables Search at once", "reveals every message on Enter in an invalid bar and sends nothing", "reveals the messages when the disabled Search is pressed", "removes the mark and enables Search once the value is corrected", "marks invalid values seeded from value at once"; model "validate" x16 (precedence and boundaries) | PASS |
| Server error mapping incl. min_gt_max | "maps the fixture 400: min_gt_max on both prices, the rest on one line under the bar" (Max price carries the message, Min price `aria-invalid` and described by the same element; `pageSize`, unknown param and `query`/`malformed_query` render as one `role="alert"` line that is the form's last child; Search stays enabled); "marks {minPrice, maxPrice, minBedrooms, city, keyword, targetBudget, status} with the server message" x7; "marks a field for an empty server message without describing it"; "clears a server mark on edit; Min price clears min_gt_max; new errors restore marks"; "restores a dismissed server mark after a submit that gets the same errors"; model "mapServerErrors" x2, "resolveMarks" x5 | PASS |
| Submit payload shape | "sends parsed numbers, the trimmed city and keyword and the status, nothing else" (`toStrictEqual({minPrice: 400000, maxPrice: 550000, minBedrooms: 2, city: 'Reston', keyword: 'pets allowed', targetBudget: 450000, status: ['pending']})`); "sends a chosen city"; "omits empty fields, Any status, page and pageSize" (`{page: 3, pageSize: 20}` -> `onSearch({})`); "shows Any for several statuses and omits status on submit"; model "toSearchParams" x2 | PASS |
| Clear | "calls onClear and empties drafts and field marks at once, keeping the unmapped line" (onClear once, onSearch never, all seven fields empty and unmarked, Search enabled); "empties typed drafts when value is already empty"; "reveals nothing while Clear is pressed, so Clear stays put until the click lands" | PASS |
| Enter submits | "Enter in {Min price, Max price, Min bedrooms, City, Keyword, Target budget, Status} submits exactly once" x7 (`toHaveBeenCalledTimes(1)`, payload equals the seeded value); "ignores the Enter that ends an IME composition" | PASS |
| Disabled while loading | "while loading disables Search, keeps fields enabled and sends nothing on Enter" (Search disabled, all seven fields and Clear enabled, typing still edits, Enter and click send nothing) | PASS |
| aria-invalid and aria-describedby wiring | "leaves valid fields without aria-invalid or aria-describedby" (all seven, seeded valid); "gives two instances distinct ids with the same labels"; every `expectMarked` call (asserts `aria-invalid="true"` and that the `aria-describedby` target's text is the message) across the client-rule and server-error cases; Min price in a min/max conflict is described by Max price's message element (client and server cases) | PASS |

9 of 9 bullets PASS.

## jsdom render description (from the passing structural assertions and FilterBar.tsx at this head)

A `<form role="search" aria-label="Filters">` containing, in DOM order, seven `div` field cells, each `<label for=…>` followed by its control and, only when marked with a non-empty message, a `<p id="…-msg">` message: Min price, Max price, Min bedrooms (inputs `type="text"`, `inputMode` decimal/numeric, `autoComplete="off"`, placeholder `Any`), City (`<select>` in a span wrapper; `Any`, the `cities` prop, then an extra option for an unlisted URL city), Keyword (input, placeholder `Search descriptions`), Target budget (input, placeholder `Any`), Status (`<select>`: Any / Active / Pending / Sold). Then an actions `div` with Search (`<button type="submit">`, natively `disabled` while invalid or loading, inside a span carrying the pointer-down reveal) and Clear (`<button type="button">`). Exactly two buttons, no emoji in the text content. When the server sends unmapped field errors, a `<p role="alert">` is the form's last child. Ids come from one `useId()` per instance, so two bars have distinct ids with identical labels. The CSS module contains no hex colours (grep for `#[0-9a-fA-F]{3,8}` finds nothing).

## Not covered here (deferred)

Browser validation of the bar's look (both breakpoints, above 760px and at or below 760px, side by side with `docs/branding-demo.html`, including the ticket acceptance "indistinguishable … allowing for the added Status field" and the oxblood-jobs check in rendered pixels) belongs to BC-12 (which wires FilterBar into SearchPage) and BC-14 (the visual pass). No browser is meaningful before then because nothing renders FilterBar.

## Observation (non-blocking)

The three blur-formatting cases print as "… shows digits while focused and undefined,000 on blur": vitest's `it.each` title formatter treats `$400` in the title string `'%s shows digits while focused and $400,000 on blur'` (FilterBar.test.tsx line 210) as a `$`-variable interpolation. Only the title is affected; the assertions inside check the literal `'$400,000'` and pass. A one-character title fix (`$$400,000` or rewording) could ride along with any later FilterBar touch; not a validation failure.

## Verdict

PASS. Component-only ticket, N/A for end-to-end until BC-12; the FilterBar suite (113/113), `make check` (exit 0, 382/382 frontend tests, lint/typecheck/format/golangci-lint clean) and `npm run build` (exit 0) are green on 13ae060fdab81e47bc29ef2f3c6ba46c82dfc33b, and every ticket Tests bullet maps to passing cases.
