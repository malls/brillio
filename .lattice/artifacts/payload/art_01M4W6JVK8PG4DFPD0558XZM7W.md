# Plan Review: BC-9 (FilterBar)

Reviewer: agent:fable-plan-reviewer. Reviewed against `docs/SPEC.md` (§5.3, §5.5, §7.1–7.6, §13.1), `BRANDING.md` (§2.2, §3.2, §4.4, §5.2, §5.8, §5.9, §7.2), `docs/branding-demo.html`, the BC-8 code now on `master` (`src/api/types.ts`, `src/api/client.ts`, `src/format/money.ts`, `src/hooks/*`, `src/test/fixtures.ts`, `src/styles/*`), the BC-5 plan's message table, the BC-12 plan (consumer of this component), and the BC-9 orchestration card.

## 1. Verdict

**PASS** — with two major amendments to apply to the plan file before moving to `in_progress` (Issues 1 and 2). Both are small, local corrections; neither requires re-planning.

## 2. Summary

The plan is unusually complete: every requirement in the task description maps to a concrete design decision, the copy table and client rules match BRANDING §7.2 and BC-5's pinned server strings exactly, the state model avoids setState-in-render/effect, and the test matrix covers every line of the task's test list. Every BC-8 dependency it names (`formatPrice`, `buildSearchQuery`, `LISTING_STATUSES`, `cities`, `validationErrorBody`, `@testing-library/user-event` 14.6.8) exists on `master` under exactly those names, and `master`'s `frontend/src` is byte-identical to the plan's anchor commit, so risk 5 is closed. The key concerns are (1) the wide-layout paragraph describes an arrangement its own CSS does not produce, so the reviewer and implementer would disagree about what "correct" looks like, and (2) the content-keyed server-error dismissal can hide a server mark after a resubmit that returns the identical 400.

## 3. Issues

**[MAJOR] Approach → CSS → "Layout reconciliation" — the described wide layout contradicts the CSS**
The plan says that at the 1040px content width "row one is the demo's six fields in demo order, row two is Status then Search and Clear." With `grid-template-columns: repeat(auto-fit, minmax(130px, 1fr))` and `gap: 20px`, the track count at 1040px is `floor((1040 + 20) / 150) = 7` (seven cells need 7×130 + 6×20 = 1030px, which fits). Auto-placement therefore puts all seven fields on row one and the actions cell alone on row two, left-aligned under Min price. The six-plus-two arrangement the plan describes only occurs between roughly 880px and 1029px of content width (viewport ~928–1077px). The plan's own arithmetic ("eight cells need 1180px") is right, but it stopped one step short. This matters because the acceptance criterion is a visual side-by-side with the demo and the plan is the reviewer's only statement of what to expect; as written, a correct implementation would be flagged.
**Recommendation:** Either (a) keep the CSS and rewrite the paragraph to state the true outcome (at ≥1030px content: seven fields on row one, actions alone on row two; 880–1029px: six fields, then Status + actions; narrower: more rows; never below 130px), and decide explicitly whether the lone actions cell stays left (auto-placement default) or is pushed right to echo the demo — note that any definite `grid-column` on `.actions` is placed before auto items and would move it onto row one; or (b) change the CSS to produce the six-plus-two layout the plan wants. Whichever is chosen, add one sentence to the plan's acceptance telling the reviewer what the bar looks like at 1040px.

**[MAJOR] Approach → "Server error mapping" → "Clearing server marks" — dismissals resurrect on an identical repeat 400**
Dismissals are stored as `{ key: serverKey, fields }` and consulted only when `serverKey` matches; a different key is read as "nothing dismissed" but the stored record is never overwritten. BC-12 passes `serverErrors` only in the `invalid` state, so a resubmit goes K1 → `[]` (loading) → K1 again when the server returns the same body. The stored `{ key: K1, fields }` still matches, so the mark the user dismissed by editing stays hidden even though the server has just re-reported it. Reproducible from the UI: a keyword over 200 characters → `too_long` → edit the keyword but keep it too long → Enter → the same 400 → no mark under Keyword. The planned test "new serverErrors restore marks" would pass (it uses different content) without catching this.
**Recommendation:** Clear the dismissal set in `submit()` (the moment the API is asked again) in addition to the content key. Keep the content key for the "parent rebuilds the array each render" case. Add one test: dismiss by editing, submit, receive the identical `serverErrors` array content → the mark is back.

**[MINOR] Scope / Files → `FilterBarProps` — `serverErrors?: FieldError[]` under `exactOptionalPropertyTypes`**
`tsconfig.app.json` sets `exactOptionalPropertyTypes: true`. With `serverErrors?: FieldError[]`, BC-12's natural wiring `serverErrors={state.status === 'invalid' ? state.error.fields : undefined}` is a type error (explicit `undefined` is not assignable). BC-12 would have to conditionally spread the prop.
**Recommendation:** Declare `serverErrors?: FieldError[] | undefined` in `FilterBarProps`. This is still "exactly SPEC §7.4" in meaning and saves BC-12 a workaround. Same consideration does not apply to `loading` (required boolean).

**[MINOR] Approach → "Client rules" → "Timing" — a disabled Search gives mouse users no explanation until blur**
Marks reveal on blur, Enter, or seed, but Search is disabled as soon as a draft is invalid. A user who types `-5` and clicks Search sees a disabled button and no message: clicking a disabled button does not move focus in Chromium or Firefox, so no blur fires. Keyboard users get Enter; mouse users get a dead end until they click elsewhere.
**Recommendation:** Add a `pointerdown` handler on `.actions` (or a wrapper around the disabled Search) that reveals every field when `!canSubmit && !loading`. Three lines, and one test ("clicking the disabled Search reveals the message"). If the planner prefers to accept the gap, say so in the plan so the reviewer does not raise it.

**[MINOR] Approach → "Parsing and display" — the `type="text"` decision is a recorded deviation from BRANDING §5.2's "number input"**
The reasoning is sound (`type="number"` cannot render `$400,000`, shows a spinner glyph outside §5.10's allowed set, and hides bad text from script), and `inputMode` preserves the mobile keypad. But BRANDING is design truth and says "number input" and "step 1"; the plan does not say where this deviation is recorded, so the branding conformance pass (BC-14) or the reviewer may flag it.
**Recommendation:** Add it to the delegator's `lattice comment` list alongside the Clear/oxblood note (risk 1), and consider one clause in the SPEC §7.4 edit: "Numeric fields are text inputs with a numeric `inputMode`; `type=number` cannot show the §7.1 price format." No BRANDING edit.

**[MINOR] Risks → item 1 — the task's acceptance line excludes §2.2 job 2 but Clear is a text button in oxblood**
The plan already resolves this correctly (BRANDING §2.2 job 2 and §5.8 name Clear explicitly; the task text is the error). Listed here so the triage is explicit.
**Recommendation:** Amend the plan's own acceptance to read "no oxblood outside §2.2 jobs 1–4" and keep the delegator comment. No task-description edit needed.

**[MINOR] Approach → "Submit, Enter, Clear" — `value.status` with several entries shows Any; inconsistent with the City principle**
The plan adds an extra `<option>` so the City select "always shows the applied filter," but a multi-status URL (which BC-8's `parseSearchParams` preserves and sends to the API) shows "Any" while results are filtered. The SPEC §7.4 edit documents the choice, so it is not silent, and the UI never writes such a URL.
**Recommendation:** Accept as planned (not blocking); the open point is correctly flagged for a human. If the planner wants consistency at low cost, the City mechanism (one extra selected option labelled "Active, Pending", submit sends the same array) is about ten lines.

**[MINOR] Header / Shared-file handling — stale anchor note**
BC-8 has landed on `master` (`ccd1669`); the worktree is still at `43e8824`. `git diff 43e8824 master -- frontend/src` is empty, so nothing renames, but the plan's "merge master once it lands" is now due before the first commit, and REVIEW_BASE switches to `master` per the orchestration card.
**Recommendation:** Add "merge `master` first" as step 0 of implementation. Also note for the implementer: the form's `onSubmit` must `preventDefault()` (implied, not stated), and `parseNumber` is deliberately stricter than the server (`1e5`, `+5` rejected client-side where `ParseFloat` accepts them); one sentence saying this is intentional will stop a reviewer asking.

## 4. Positive Observations

- **Contract fidelity.** Every string in the copy table was checked: the three BRANDING §7.2 sentences are verbatim, and the seven fallbacks match BC-5's rules table character for character, in BC-5's precedence order (number → sign → whole). The decision to copy BC-5's sentences so a value "reads the same whether the client or the API catches it" is the right call and risk 4 names the coupling honestly.
- **State model.** Keyed local state with derive-in-render (no effect, no setState-in-render) is the correct React 19 pattern and the plan explains why in terms the react-hooks 7 rules will enforce. Keying the seed on the seven filter fields via `buildSearchQuery` so page changes never wipe typing is a precise, well-justified choice.
- **Pixel reasoning.** The 35px control box in every state (6+22+6+1 at rest, 6+22+5+2 focused/invalid), the fixed-height `.actions` cell that keeps underlines aligned when a message appears (an improvement on the demo's `align-items: end`), and the 1px bottom padding that puts the button baseline on the underline all show the demo was read closely, not approximated.
- **Accessibility.** Real `<label htmlFor>`, `aria-invalid` only when a message shows, `aria-describedby` pointing Min price at Max price's message in a conflict, `role="alert"` for the unmapped line, `useId()` per instance, and the `content: "▾" / ""` alt-text form so the glyph is not announced. Nothing here is boilerplate.
- **Enter handling.** The plan correctly identifies that implicit submission is skipped when the submit button is disabled and never fires from a `<select>`, and handles `isComposing`. This is exactly the kind of thing that would otherwise surface as a flaky integration test in BC-12.
- **Scope discipline.** Out-of-scope list is explicit (App.tsx, shell styles, tokens, BC-8 helpers consumed not forked), the one SPEC edit is justified by a real contradiction (§7.2 "only Any" vs an applied city filter), BRANDING is not touched, and the private model file is argued for rather than assumed.
- **Test matrix.** Twenty-eight component cases plus a model table map one-to-one onto the task's test list and the acceptance criteria, including the subtle ones (equal object or page-only change keeps typing; two instances with distinct ids; Enter submits exactly once per field).
