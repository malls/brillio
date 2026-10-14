# Code review — BC-9 FilterBar (reviewer: fable sub-agent (decision 04:02Z))
Verdict: PASS
Reviewed HEAD: 13ae060fdab81e47bc29ef2f3c6ba46c82dfc33b
Base: master @ 76d6a97bf7f18da629e9294b139ae65147212ca3 (contained in HEAD; product diff vs master is 6 files: frontend/src/components/FilterBar/{FilterBar.tsx,FilterBar.module.css,FilterBar.test.tsx,filterBarModel.ts,filterBarModel.test.ts} and docs/SPEC.md)
Gate: make check exit 0 (golangci-lint 0 issues; eslint, `tsc -b --noEmit`, prettier clean; `go test -race` ok; vitest 22 files, 382/382 passed); `npx vitest run src/components/FilterBar --reporter=verbose` exit 0 (2 files, 113/113, no React/act warnings in the log); npm run build exit 0 (tsc -b + vite). `git status --porcelain` empty before and after.

## Summary
The head matches the plan as amended by the Cycle 1 resolutions, the Cycle 1 review routing and the 03:25Z text-button recipe. All copy lives in one `as const` table (`filterBarModel.ts:38-63`) and matches BRANDING §7.2 and BC-5's landed sentences in `backend/internal/httpapi/query.go:93-108` word for word. The client rules, min/max placement, server mapping (`min_gt_max`, unmapped `pageSize`/`minprice`/`query`+`malformed_query` on one `role="alert"` line), payload shape, Enter/IME handling, Clear, the reveal timing and the aria wiring are correct and each has a test that would fail if its feature were removed (the IME test's first and third keydowns differ only by `keyCode`, so the 229 guard is what the test proves). All five prior fix items are in place, with the pointerdown reveal now scoped to a Search-only wrapper and Clear keeping focus in the field on press; no regressions found in the phone layout, keyboard or touch paths. Scope is clean: the branch's own commits touch only the FilterBar folder and the planned SPEC §7.2/§7.4 sentences, and no `.lattice/` file. Nothing actionable remains; the three nits below are optional observations.

## Prior findings (art_01M4W8RJEX9PZGNYXK1279SVG3)
- 1 (Minor, pointerdown reveal also fired on Clear): fixed — `onPointerDown` moved to `<span className={styles.searchSlot}>` wrapping Search only (`FilterBar.tsx:286-290`, handler `:184-188`), Clear gets `onMouseDown` → `preventDefault()` (`FilterBar.tsx:295-297`); `.searchSlot { display: flex }` (`FilterBar.module.css:139-141`) and `.searchSlot, .search { flex: 1 }` at ≤760px (`:232-235`) keep Search full width on phone. Tests: `reveals the messages when the disabled Search is pressed` (`FilterBar.test.tsx:357-364`) and `reveals nothing while Clear is pressed, so Clear stays put until the click lands` (`:366-380`; removing the preventDefault moves focus, blurs, reveals and fails it).
- 2 (Nit, CSS comment): fixed — `FilterBar.module.css:125-126` reads "a message under a field in the same row never moves the buttons".
- 3 (Nit, WebKit IME keyCode 229): fixed — `isPlainEnter` adds `event.nativeEvent.keyCode !== 229` with a comment (`FilterBar.tsx:63-75`). Test `ignores the Enter that ends an IME composition` (`FilterBar.test.tsx:533-542`).
- 4 (Nit, `numberText` exponent for non-integers): fixed — `filterBarModel.ts:104-114` writes non-integers via `toLocaleString('en-US', { useGrouping: false, maximumFractionDigits: 20 })`, keeping the exponent form only when the digits do not round-trip, so a >20-decimal value is marked rather than silently rounded. Tests `filterBarModel.test.ts:161-170` and `normalizeDraft` cases (`:206-210`).
- 5 (Nit, Clear hid the unmapped line): fixed — `Dismissals.all` removed; `notice = server.unmapped` follows `serverErrors` alone (`FilterBar.tsx:115-117`), Clear dismisses field marks only (`:162-167`). Test `calls onClear and empties drafts and field marks at once, keeping the unmapped line` (`FilterBar.test.tsx:560-591`).
- 6 (Nit, live-region feedback): routed to BC-21; not in this diff, correctly.

## Findings
### Critical
- None
### Major
- None
### Minor
- None
### Nit
- `frontend/src/components/FilterBar/FilterBar.tsx:149-156` — `submit()` returns on `loading` before the `invalid` → `revealAll()` branch, so Enter during a load on an invalid bar reveals nothing (pointerdown reveal gated the same way at `:185`). Exactly the plan's order; loading is brief. Optional, no action needed.
- `frontend/src/components/FilterBar/FilterBar.tsx:115-116` and `FilterBar.test.tsx:581` — the rationale "Clear keeps pageSize, so the 400 behind the line may still apply" cites a case BC-8's `parseSearchParams` cannot produce (it drops out-of-range `pageSize`, unknown params and `page` < 1). The behaviour (the notice tracks the response, not the drafts) is the right contract-level one regardless. Optional, no action needed; a comment tweak would be more accurate.
- `frontend/src/components/FilterBar/FilterBar.tsx:176-181` — `preventDefault()` on a plain Enter on the two `<select>`s suppresses the native "open the picker on Enter" some macOS browsers offer; that is what SPEC §7.4 "Enter in any field submits" asks for and the plan chose it deliberately; Space / Alt+Down still open the picker. Optional, no action needed; worth one click in BC-12's browser e2e (Firefox in-content dropdown: confirm choosing an option with Enter does not also submit).

## Acceptance check
- Labels, placeholders, messages verbatim: met — `FIELDS` (`filterBarModel.ts:21-29`), `COPY.messages` (`:51-62`); tests `FilterBar.test.tsx:81-104`, `:285-330`.
- Seven fields in order; Status options Any/Active/Pending/Sold from `LISTING_STATUSES`: met — `FilterBar.test.tsx:81-89`, `:122-132`.
- Field anatomy and states (label, Body 15 value, 1px ink-2 underline; focus/invalid 2px oxblood; message Meta 13 oxblood; disabled ink-3 + line; select ▾ ink-2; placeholder ink-2; 35px box in every state): met — `FilterBar.module.css:32-121`.
- Blur formatting: met — `filterBarModel.ts:147-169`, `FilterBar.tsx:260-264`; tests `:210-250`.
- Client rules and placement (number → sign → whole; min>max marks both, message under Max price, Min price described by it; equal valid; budget ≤ 0): met — `filterBarModel.ts:221-245, 319-351`; tests `FilterBar.test.tsx:254-330`, `filterBarModel.test.ts:55-105, 271-329`.
- Server error mapping incl. `min_gt_max` and the unmapped line (`pageSize`, `minprice`, `query`/`malformed_query`); Search stays enabled: met — `filterBarModel.ts:287-310`; tests `FilterBar.test.tsx:403-442`, `filterBarModel.test.ts:230-269`.
- Dismissals (edit clears, Min price clears `min_gt_max`, identical content stays dismissed, new content restores, submit resets): met — `FilterBar.tsx:124-138, 157-158`; tests `:444-481`.
- Payload shape: met — `filterBarModel.ts:254-274`; tests `FilterBar.test.tsx:485-522`, `filterBarModel.test.ts:107-134`.
- Clear (onClear, drafts and field marks emptied at once, also when value already empty; notice kept; enabled while loading): met — `FilterBar.tsx:162-167`; tests `:560-604`, `:551`.
- Enter submits exactly once from every field incl. both selects; not during IME composition incl. keyCode 229: met — `FilterBar.tsx:63-75, 169-181`; tests `:524-542`.
- Search disabled while invalid or loading; fields stay enabled: met — `FilterBar.tsx:112, 287`; tests `:205, :338, :544-556`.
- pointerdown reveal on Search only: met (Prior finding 1).
- aria-invalid / aria-describedby wiring and unique ids: met — `FilterBar.tsx:190-211`; tests `:254-271, :437-442, :608-626`.
- Layout >760px (auto-fit 130px grid, gap 20; at 1040px content seven fields on row one, actions row two at the left; never a field under 130px): met by reading the CSS (`FilterBar.module.css:14-22`; 7×130 + 6×20 = 1030 ≤ 1040 < 1180). Not rendered (validation N/A until BC-12).
- Layout ≤760px (two columns, actions full-width row, Search fills it, Clear at right, padding 16px): met — `FilterBar.module.css:220-236`, mirrors demo lines 98-100.
- Tokens only via `var(--…)`, no hex, no transitions/animation: met.
- Oxblood only in §2.2 jobs 1-4: met — focus/invalid underline (jobs 3/4), message (job 4), Search fill (job 1) with `oxblood-deep` hover/pressed only, focus rings (job 3), Clear (job 2); unmapped notice is `ink` (§2.3).
- Clear matches the shared text-button recipe (Pager `.button`): met — `FilterBar.module.css:180-208`.
- No emoji, no icons beyond `▾` (with `/ ''` alt text): met — test `:134-142`.
- Consumed helpers used, not forked: met — `buildSearchQuery`, `LISTING_STATUSES`, `formatPrice`, fixtures; private `parseNumber` deliberately stricter than the API, commented.
- `serverErrors?: FieldError[] | undefined`; form `onSubmit` preventDefault: met.
- SPEC edits (§7.2, §7.4): met — +8/-3 as planned.
- Scope (FilterBar folder + SPEC only; no `.lattice/` in branch-own commits): met — `git log master..HEAD --no-merges --name-only` lists only the six files.