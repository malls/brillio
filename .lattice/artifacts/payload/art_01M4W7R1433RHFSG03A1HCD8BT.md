Lattice-Reviewed-Commit: 0e70187ac5d4838851087f944ee2e6784f396a4c
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-11-pager-status-loading
Lattice-Reviewed-Base: master (e5f49bb9d424d29e304faf33dca345917cdc657e)
Lattice-Reviewed-Head: feat/BC-11-pager-status-loading (0e70187ac5d4838851087f944ee2e6784f396a4c)

# Code Review: BC-11 — Pager, StatusLine and LoadingLine

Reviewed commit: `0e70187ac5d4838851087f944ee2e6784f396a4c` (branch `feat/BC-11-pager-status-loading`, base master `e5f49bb`).
Reviewer: agent:fable-5.1-reviewer, cold read of the diff against `.lattice/plans/task_01M4VXK372WGG8ZBYFVE97TXH3.md` (including the authoritative Cycle 1 resolutions).

## 1. Verdict

**PASS**

## 2. Summary

Nine new files (three components, three CSS modules, three test suites; 667 lines, no existing file touched) were reviewed against the plan, SPEC §7.4–§7.6 and BRANDING §2.2/§3.2/§5.3/§5.7/§5.8/§5.9/§6/§7.2. The implementation follows the plan and its Cycle 1 resolutions line for line: every StatusLine string is verbatim, oxblood is confined to the pager text buttons, the focus ring, the current page and the loading segment, and the full gate is green. No correctness, security or scope defects were found; the only finding is a cosmetic duplicated comment.

## 3. Issues

**[minor] frontend/src/components/StatusLine/StatusLine.tsx:14 — Duplicated comment**
The sentence above `summarize` ("Left: what the list on screen is. Right: where that list sits in the pages.") is restated almost word for word in the comment above `describeState` (lines 19–20). Harmless, but it reads as a leftover.
**Fix:** Delete line 14 (keep the fuller comment on `describeState`, which also cites the BRANDING/SPEC sections). Optional; no rework cycle warranted.

No other issues found.

## 4. Positive Observations

**Verification performed (all at the reviewed head, worktree clean before and after):**

| Check | Result |
|---|---|
| `make check` (golangci-lint, eslint `--max-warnings 0`, `tsc -b`, `prettier --check`, `go test -race`, vitest) | Green: 0 lint issues, 13 files / 228 tests passed |
| `npx vitest run src/components/Pager src/components/StatusLine src/components/LoadingLine` | 3 files / 31 tests passed |
| `npm run build` | Succeeds. Note: the three new CSS modules are not in `dist/` because nothing imports the components until BC-12, so the production build does not yet exercise them. |
| Vite `transformRequest` on the three `.module.css` files | `@keyframes slide` is localized to `_slide_16eo8_1` and the `animation:` shorthand on `.segment` references the same scoped name; the `prefers-reduced-motion` block targets the scoped `.segment` class. The CSS-modules concern in the plan is confirmed sound. |
| `grep oxblood` on the three stylesheets | Pager `.button` color, `:focus-visible` outline, `[aria-current='page']` color; LoadingLine `.segment` background. StatusLine: none. Matches acceptance criterion 2 exactly. |
| Hex literals / emoji on the nine files | None. Non-ASCII is limited to `‹`, `›` and `·`. |
| `global.css` | `button { font: inherit; color: inherit; letter-spacing: inherit }` so the pager buttons take Figtree and inherit `tabular-nums` from `.pager`; universal `box-sizing: border-box` so the 48px pager includes its hairline (same as Header's 56px). |
| `useSearchParamsState` | `page` is parsed as an integer ≥ 1 and `setPage` rejects non-integers, so the Pager's `Math.min(page - 1, totalPages)` and past-end handling only ever see sane integers. |
| `src/test/setup.ts` | Registers RTL `cleanup` in `afterEach`, so the StatusLine `halves()` helper's `getByRole('status')` is safe across cases. |
| Lattice board | The plan-review triage comment (2026-10-14T03:24:50Z) records the transparent-resting-track decision and the one-line flip, as resolution 2 required. |

**Correctness.** The `describeState` switch covers all six `SearchState` members with no default and type-checks under `strict` + `noFallthroughCasesInSwitch`, so a new status would fail the build rather than render silently. The invalid-with-previous branch keeps describing the list still on screen, error returns empty even with a previous list, loading with a previous list keeps the page half; all three match the demo panels and SPEC §7.5. The Pager's past-end behaviour (no `aria-current`, Next disabled, Previous → last page) is both implemented and tested. Disabled ends use the real `disabled` attribute, which the tab-order tests prove.

**Design decisions worth keeping.** Styling the current page via `.button[aria-current='page']` ties the look to the semantics so they cannot drift. The current page stays a focusable button that no-ops, which keeps it discoverable to screen readers. The inactive LoadingLine renders the same 2px box with no role, name or segment, so AT never meets an idle progressbar and the layout never shifts. The reduced-motion branch is pure CSS, which is the only approach that actually proves the shipped stylesheet rather than a JS mock.

**Tests.** Every case in the plan's S1–S12, P1–P8 and L1–L5 maps to a named test, and the strings are asserted with exact `toBe`/`toEqual` including the U+00B7 middle dot. The Pager suite exercises mouse, Enter and Space, and both tab-order directions. The CSS-source tests are implementation-coupled by design; the plan documents why (jsdom evaluates no media queries, Vitest does not process CSS) and defers visual proof to BC-12's browser validation. Resolution 4's hardening (non-empty check, regex anchored to `.segment` inside the `@media` block) is present.

**Hand-offs confirmed for the orchestrator / BC-12.**
- Validation is N/A for this ticket per the description; the validation comment should cite the 31-test vitest output above at `0e70187`.
- BRANDING §3.2 and §3.3 say buttons carry +0.02em tracking while §5.8 and the demo's `.link` carry none. The implementation follows the authoritative plan resolution (no tracking). This is a BRANDING.md inconsistency for the branding owner, not a BC-11 defect, and BC-9/BC-10 should receive the same recipe as the plan notes.
- BC-12 must keep the LoadingLine mounted, move focus after `onChange`, query page buttons as `Page N`, and expect both `role="status"` and the results live region to announce.
