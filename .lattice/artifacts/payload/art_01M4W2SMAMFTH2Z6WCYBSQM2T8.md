Lattice-Reviewed-Commit: 1d6008a32c663486c6fbed355d320aa80ec240ca
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold
Lattice-Reviewed-Base: master (2afc005e1fe0b4a3347e95cc540b53244ffd4cdd)
Lattice-Reviewed-Head: feat/BC-7-frontend-scaffold (1d6008a32c663486c6fbed355d320aa80ec240ca)

# Code Review: BC-7 (frontend scaffold: Vite 8, React 19, tokens, shell, Header)

Reviewed range: `master` 2afc005 -> `feat/BC-7-frontend-scaffold` 1d6008a (2 commits, 22 files,
488 insertions excluding `package-lock.json`). Reviewer: agent:claude-fable-5.1-reviewer, 2026-10-13,
fresh context. The prompt's inline diff was truncated inside `package-lock.json`, so this review read the
full diff with `git diff 2afc005..1d6008a` in the worktree and re-ran every gate itself.

## 1. Verdict

**PASS**

No blocking findings. Two minor, non-blocking observations below for the orchestrator to route
(one is a one-token tsconfig addition that BC-8 will probably need; the other is a documented layout
trade-off that needs no code change in this ticket).

## 2. Summary

The diff delivers exactly what the plan (including the Cycle 1 resolutions R1-R10) specifies: the pinned
dependency table with nothing extra, strict tsconfigs with the three required flags on top of the
unmodified create-vite 9.2.1 react-ts template, type-aware ESLint 9 with hooks and jsx-a11y, Prettier,
Vitest + RTL + jest-dom with explicit cleanup, the `/api` and `/healthz` proxy driven by
`VITE_API_PROXY_TARGET`, the BRANDING 2.1 token block byte-for-byte plus the 3.2 scale and 4.x spacing
as custom properties, the 4.2 shell grid, a Header that matches `docs/branding-demo.html`, the four
planned tests, the Makefile BC-7 block with the R1 sentinel, and the two R3 SPEC edits. Every acceptance
check (`make lint-web test-web`, `npm run build`, dependency list, emoji scan, tracking diff, proxy
behaviour, `tsc -b` catching errors where plain `tsc` does not) was re-executed and passed. Code quality
is high: small files, comments that cite BRANDING sections, no `!` assertions, named exports, no
disabled lint rules.

## 3. Issues

**[MINOR] frontend/tsconfig.app.json:5 — `lib` omits `DOM.Iterable`**
`"lib": ["ES2023", "DOM"]` is the create-vite template's value (verified against the 9.2.1 tarball), and
nothing in BC-7 needs more. But without `DOM.Iterable`, `for (const [k, v] of params)` over
`URLSearchParams`, `Object.fromEntries(params)`, `[...form.elements]`, and `for...of` over `NodeList`
fail with TS2488/TS2802 under `verbatimModuleSyntax` + strict. BC-8's `useSearchParamsState` round-trips
every URL param and is the first likely caller. Not a defect in this ticket; flagged so BC-8 does not lose
a cycle on it.
**Fix:** either add `"DOM.Iterable"` to `lib` in `tsconfig.app.json` now (additive one-line commit on this
branch is allowed per the shared-file rules), or put a one-line note in the BC-8 hand-off and let BC-8 add
it. No test or lint change needed; `tsc -b --noEmit` stays green either way.

**[MINOR] frontend/src/styles/shell.module.css:25 — results column is narrower than the pinned rows by the scrollbar width on classic-scrollbar platforms**
`.results` is the scroll container and also carries `padding-inline: var(--inline-pad)`. A visible
vertical scrollbar (Windows, Linux, macOS with "always show scroll bars") takes 15-17px inside the border
box, so the right edge of content inside `.results` sits that far left of the Header's right edge; left
edges still align, which is the edge BRANDING 4.5 cares about ("everything left-aligned except the price
block"). The price block (BC-10) is right-aligned, so on those platforms it will end short of the scope
label's right edge. macOS overlay scrollbars hide this entirely. Plan decision 8 chose this deliberately
("the results scroller keeps its scrollbar at the viewport edge"), and the alternatives
(`scrollbar-gutter: stable both-edges`, or a nested scroller) misalign the left edge or move the
scrollbar off the viewport edge, so the current choice is the right one.
**Fix:** no code change for BC-7. Record the trade-off in the BC-10/BC-14 hand-off so the right-aligned
price block is checked on a classic-scrollbar platform during BC-14 polish. Suggested routing: ignore with
justification.

## 4. Positive Observations

- **Plan fidelity, including the review-cycle resolutions.** R1 (sentinel target
  `frontend/node_modules/.package-lock.json`), R3 (both SPEC edits, wording matches the plan), R6 (the
  `.row` comment block in `shell.module.css`), R9 (T2 asserts `textContent` instead of a child count) are
  all in the diff exactly as resolved. The `.PHONY` line lists only the three web targets; no other
  Makefile block was touched.
- **Dependency discipline.** `package.json` holds the 18 names in the plan's table, every one an exact
  pin, nothing from the template's extras (`oxlint`, `@eslint/js`, `globals`, `react-refresh`) survived.
  Installed versions match the pins (`typescript 6.0.3`, `eslint 9.39.5`, `vite 8.3.4`, `react 19.3.0`).
- **Template kept intact, then tightened.** Both tsconfigs are byte-identical to create-vite 9.2.1's
  react-ts template apart from the inserted `strict` / `noUncheckedIndexedAccess` /
  `exactOptionalPropertyTypes` block. That keeps future template diffs readable.
- **The lint config is actually type-aware.** `eslint --print-config src/App.tsx` shows 123 rules with
  `@typescript-eslint/no-floating-promises`, `no-unsafe-assignment`, `react-hooks/rules-of-hooks`,
  `react-hooks/set-state-in-effect`, `jsx-a11y/alt-text` all at error level; `--max-warnings 0` makes
  warnings fail the gate.
- **`tsc -b` decision is correct and now documented in SPEC 8.** Verified: a deliberate type error in
  `src/` fails `tsc -b --noEmit` (exit 2) and passes plain `tsc --noEmit` (exit 0) on the solution file.
- **Tokens match BRANDING to the character.** The 2.1 block is byte-for-byte under `prettier-ignore`;
  every 3.2 row (price, address, monogram, body, description, meta, button, wordmark, label, rank) is
  present with the right size, leading and tracking; the 3.1 fallback stacks and the Google Fonts href
  are exact; the 760px breakpoint uses `max-width: 760px`, matching 4.4's "<= 760px".
- **Header matches the demo, not just the prose.** 56px with the hairline inside the box (the demo also
  uses `* { box-sizing: border-box }`), `-webkit-font-smoothing: antialiased` on body (demo line 14),
  scope label styled like the demo's `.lbl`. DOM text `Brillio Listings` uppercased by CSS, as the demo
  does, so assistive tech reads words. The wordmark is the page `<h1>` inside a `<header>` banner landmark.
- **Shell and scroll model are per 4.2.** `html, body, #root { height: 100% }`, `html, body { overflow:
  hidden }`, grid rows `auto auto auto minmax(0, 1fr) auto`, `.results { min-height: 0; overflow-y: auto;
  overscroll-behavior: contain }`. The `--inline-pad` `max(gutter, (100% - 1040px) / 2)` trick centres a
  1040px column while letting hairlines run full-bleed, with a comment explaining the `100%` resolution.
- **Proxy verified end-to-end, not just by reading.** With a dummy HTTP server on 127.0.0.1:18199 and
  `VITE_API_PROXY_TARGET=http://127.0.0.1:18199`, `GET /healthz` and `GET /api/cities` on the Vite port
  returned the dummy's 404 and both paths appeared in its access log; `GET /` returned 200. `loadEnv`
  with `||` fallback handles unset and empty values; `changeOrigin: true` is right for `api:8080` under
  Compose.
- **Tests assert behaviour, not styles.** Roles, accessible names, heading level, order of the five shell
  children, and `aria-live="polite"` on `main` (SPEC 7.6). `setup.ts` registers jest-dom and an explicit
  `cleanup()` because globals are off; `restoreMocks` and `unstubGlobals` are preset for BC-8's fetch stubs.
- **Hygiene.** No emoji in any tracked file under the R10 regex; `git ls-files frontend` equals the files
  on disk (nothing untracked, nothing ignored by accident); `git status` clean after the build; no
  `.lattice/` changes in the diff; commit messages are conventional and name the ticket.

## Verification record

| Check | Command (from the worktree) | Result |
|---|---|---|
| Lint + typecheck + format | `make lint-web` | eslint clean, `tsc -b --noEmit` clean, Prettier "All matched files use Prettier code style" |
| Unit tests | `make test-web` | Vitest 5.0.4, 2 files, 4 tests passed |
| Build | `cd frontend && npm run build` | `tsc -b` clean; `dist/` 0.77 kB html, 2.56 kB css, 220 kB js |
| Versions | `node -p` on installed package.json files | typescript 6.0.3, eslint 9.39.5, vite 8.3.4, react 19.3.0 |
| Dependency list | read `frontend/package.json` | 18 names, exact pins, equals the plan table |
| Emoji scan | `git ls-files -z frontend Makefile docs/SPEC.md \| xargs -0 perl -CSD -ne '... R10 regex ...'` | no output |
| Tracking | `diff <(git ls-files frontend) <(find frontend ...)` | empty, exit 0 |
| `tsc -b` vs plain `tsc` | temp file with a type error, then removed | `-b` exit 2 with TS2322; plain exit 0 |
| Lint rules active | `npx eslint --print-config src/App.tsx` | 123 rules; type-aware, hooks and a11y rules at level 2 |
| Proxy | dummy server on 18199, `vite --port 15199`, curl `/healthz`, `/api/cities`, `/` | 404, 404 (both logged by the dummy), 200 |
| Template parity | `npm pack create-vite@9.2.1` into a temp dir, diffed tsconfigs | identical apart from the added strict block |

## Notes for the orchestrator

- Both minor items are non-blocking. Recommended routing: item 1 either fix inline as an additive
  commit or hand to BC-8; item 2 ignore with justification (document in the BC-10/BC-14 hand-off).
- Plan R3 also asks the delegator for a `lattice comment` listing the two SPEC edits and the six
  non-allowlisted packages with justifications, plus an FYI mail about ESLint 9. Those live outside the
  diff and were not verified here; the SPEC edits themselves are present and match the agreed wording.
- The browser-level claims (fonts loaded, computed styles, document does not scroll) are the validation
  stage's job per the plan's R2/R5 steps; nothing in this review substitutes for them.
