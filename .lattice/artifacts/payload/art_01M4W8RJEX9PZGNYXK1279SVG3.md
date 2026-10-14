# Code review — BC-9 FilterBar
Verdict: PASS-WITH-NITS
Reviewed HEAD: f79fa6fc16c125a727a6d93dba9850604a50be8f
Base: master @ e5f49bb9d424d29e304faf33dca345917cdc657e (merge base; master has since moved to 0e6730f with BC-11, and `git merge-tree` reports a clean merge)
Gate: `make check` exit 0 (golangci-lint 0 issues; eslint, `tsc -b --noEmit` and prettier clean; `go test -race` ok; vitest 12 files, 307/307 passed). `npx vitest run src/components/FilterBar --reporter=verbose` exit 0 (2 files, 110/110 passed). `npm run build` exit 0.

Reviewer: agent:opus-5.5-reviewer (own-reviewer fallback; headless lattice code-review hit the 600 s watchdog with no result).

## Summary
The change matches the plan and its authoritative Cycle 1 resolutions. All copy sits in one table and matches BRANDING §7.2 and BC-5's server sentences word for word. The client rules, the server mapping (`min_gt_max`, unmapped `query`/`malformed_query`), the payload shape, Enter handling, Clear and the aria wiring are all correct and well tested. Scope is clean: only `frontend/src/components/FilterBar/` and the planned SPEC §7.2/§7.4 sentences changed, with no `.lattice/` files. The setState-during-render re-seed was checked under StrictMode with a temporary probe test (since deleted; tree confirmed clean): no loop, no console errors, A→B→A and identical repeat serverErrors handled, `react-hooks/set-state-in-render` lint passes. Remaining findings are small; the most notable is that pressing Clear can reveal a message and move the button before the click lands (one-line guard).

## Findings
### Critical
- None
### Major
- None
### Minor
- `frontend/src/components/FilterBar/FilterBar.tsx:184-189, 285` — The pointerdown reveal is attached to the whole `.actions` wrapper, so it also fires when the user presses Clear. If the focused field holds an unrevealed invalid draft, pressing Clear shows its message between press and release, making row one taller by one message line (~24px); at 1040px and at ≤760px the actions are below, so Clear moves down before release and the click can be lost. On touch, pointerdown precedes the compatibility mousedown and blur. (Message appearance verified in jsdom; the layout shift is inferred from the CSS.) — Recommendation: limit the reveal to Search (return early when `(event.target as Element).closest('button[type="button"]')` matches, or move `onPointerDown` to a wrapper around Search only). Optionally `onMouseDown={(e) => { e.preventDefault(); }}` on Clear so the blur-triggered reveal on desktop cannot move it either.
### Nit
- `frontend/src/components/FilterBar/FilterBar.module.css:123-127` — Comment says "a message under a field never moves the buttons"; true only when the actions share a row with that field. — Recommendation: reword to "a message in the same row".
- `frontend/src/components/FilterBar/FilterBar.tsx:65-74` — `isPlainEnter` relies only on `nativeEvent.isComposing`; in WebKit the Enter that commits an IME composition arrives after `compositionend` with `isComposing === false` and `keyCode === 229`, so a Safari CJK user confirming a conversion in Keyword triggers a search. — Recommendation: also bail when `event.nativeEvent.keyCode === 229`.
- `frontend/src/components/FilterBar/filterBarModel.ts:101-105` — `numberText` keeps exponent notation for non-integers (`String(1e-7)` is `'1e-7'`), which `parseNumber` rejects; `?minPrice=0.0000001` (accepted by BC-8's parser and the API) seeds a draft marked "Min price must be a number." and disables Search. Contrived. — Recommendation: write non-integers with `n.toLocaleString('en-US', { useGrouping: false, maximumFractionDigits: 20 })`.
- `frontend/src/components/FilterBar/FilterBar.tsx:119, 166` — Clear's `all: true` also hides the unmapped line. With `?pageSize=500`, Clear keeps pageSize (SPEC §7.1) so the URL and the 400 stay the same, but the only explanation disappears until the next Search. — Recommendation: Clear dismisses field marks only and leaves `notice` alone, or keep and note the decision.
- `frontend/src/components/FilterBar/FilterBar.tsx:155-157, 276-280` — Messages revealed by an invalid Enter or a server 400 are linked via `aria-describedby` but not in a live region; a screen-reader user pressing Enter on an invalid bar gets no feedback. SPEC §7.6 does not require it. — Recommendation: follow-up task (e.g. focus the first invalid field on an invalid submit). Not for this ticket.

## Deviations assessed
- Guarded setState-during-render re-seed (`FilterBar.tsx:96-104`): accepted. React's documented pattern for adjusting state on prop change; both writes run only on key change and the next render matches the stored key, so no loop. StrictMode + Profiler A→B→A with an equal-object rerender stayed under 10 commits with no `console.error`. Pure derivation would bring back stale typing on A→B→A and stale dismissals on X→undefined→X; both have regression tests (`FilterBar.test.tsx:179-187, 428-451`). Lint passes with `react-hooks/set-state-in-render` at error. Handlers start from the current key (`updateLocal`).
- Min bedrooms wholeness before sign: accepted. BC-5's `query.go` calls `p.whole()` before the `< 0` check, so `-1.5` reads the same on both sides; `-1` still uses the BRANDING string (`filterBarModel.test.ts:66-69`).
- Actions `align-items: center` with computed `padding-top`: accepted. Demo `.actions` is `align-items:center`; 14 + 4 + 35 − 40 − 1 = 12px puts Search's bottom on the field underline and keeps buttons still when a same-row field shows a message (wording nit above).
- SPEC wording cites "BRANDING.md §7.1": accepted; corrects the plan (SPEC §7.1 is URL state).
- Small additions accepted: pressed state `oxblood-deep` (BRANDING §2.1 "Hover and pressed state of anything oxblood"); disabled select glyph `ink-3`; label ellipsis (longest label fits 130px); plain-digit seeding of huge integers via BigInt; trimmed, case-insensitive city match (tested).
- Delegator decisions (oxblood jobs 1-4; `type="text"` + `inputMode`): accepted. Clear is job 2 by name; SPEC §7.4 clause added in-branch.

## Acceptance check
- Seven fields in order with verbatim labels: met (`FilterBar.test.tsx:81-89`).
- Placeholders verbatim, Any first in both selects: met (`:91-104`).
- City options incl. URL city as extra selected option (also `cities=[]`): met (`:106-120`). Status options Any/Active/Pending/Sold: met (`:122-132`).
- Search landmark "Filters", Search submit, Clear button, no emoji; only glyph is CSS `▾` with empty alt: met (`:134-142`, `FilterBar.module.css:100-101`).
- Seeding, re-seed rules, A→B→A: met (`:146-187`). Multiple statuses show Any and are omitted: met (`:189-196`). Invalid seeds marked immediately: met (`:198-206`).
- Blur formatting for all three money fields; typed `$400,000` submits 400000; bedrooms never money; fractions/invalid left as typed: met (`:210-220`).
- min>max marks both, one message under Max price, both described by it; equal valid: met (`:254-283`).
- Negative prices, non-numbers, bedrooms `-1`/`2.5`, budget `0`/`-5` exact strings: met (`:285-330`).
- Reveal timing, invalid Enter reveals all and sends nothing, pointerdown reveal, recovery: met (`:332-378`); pointerdown scope see Minor.
- Server mapping incl. `min_gt_max`, unmapped `pageSize`/`minprice`/`query`/`malformed_query` on one `role="alert"` line; Search stays enabled: met (`:387-402`). Every mapped field: met (`:404-419`). Empty message: met. Dismissals on edit, reset on new content, restored after submit: met (`:428-465`).
- Payload shape and omissions (`{}` from `{page:3,pageSize:20}`): met (`:469-506`, `filterBarModel.test.ts:107-134`).
- Enter submits exactly once from all seven fields incl. selects (`toStrictEqual`); removing `preventDefault` would double-submit, so the test catches it: met (`:508-515`).
- Loading: Search disabled, fields and Clear enabled, Enter/click send nothing; typing survives loading: met (`:517-529`).
- Clear empties drafts and marks at once, also when value already empty: met (`:533-568`).
- No aria-invalid/describedby on valid fields; unique ids across two instances: met (`:572-590`).
- Layout >760px: `repeat(auto-fit, minmax(130px, 1fr))`, gap 20; 1040px → seven fields row one, actions row two left; no field under 130px: met by reading the CSS (not rendered; validation N/A until BC-12).
- Layout ≤760px: two columns, actions full width, Search `flex: 1`, space-between, padding-block 16px: met (`FilterBar.module.css:210-225`, demo lines 98-100).
- Field anatomy and states per BRANDING §3.2/§5.2: met.
- Tokens only via `var(--…)`, no hex, no transitions: met.
- Oxblood only in §2.2 jobs 1-4, `oxblood-deep` only hover/pressed; unmapped line in `ink` per §2.3: met.
- Consumed helpers used, not forked; private parser intentionally stricter than the API with a comment: met.
- `serverErrors?: FieldError[] | undefined`; form `onSubmit` preventDefault: met.
- SPEC edits (§7.2 clause, §7.4 sentences): met.
- Scope: FilterBar folder + SPEC only; no `.lattice/`: met.