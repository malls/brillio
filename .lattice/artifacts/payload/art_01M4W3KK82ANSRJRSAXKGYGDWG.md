Lattice-Reviewed-Commit: 0d1c2a2a85c4ef23fb3da77452fa1387095f4986
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold
Lattice-Reviewed-Base: master (7ab182d6817d3418d45ebe6dee860540e49f7730)
Lattice-Reviewed-Head: feat/BC-7-frontend-scaffold (0d1c2a2a85c4ef23fb3da77452fa1387095f4986)

Lattice-Reviewed-Commit: 0d1c2a2a85c4ef23fb3da77452fa1387095f4986
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold
Lattice-Reviewed-Base: master (7ab182d6817d3418d45ebe6dee860540e49f7730)
Lattice-Reviewed-Head: feat/BC-7-frontend-scaffold (0d1c2a2a85c4ef23fb3da77452fa1387095f4986)

# Code Review: BC-7 (frontend scaffold, re-review at the merged HEAD)

Reviewed range: `master` 7ab182d -> `feat/BC-7-frontend-scaffold` 0d1c2a2 (merge of master/BC-2 into the
branch; 22 files, 488 insertions excluding `package-lock.json`). Reviewer: agent:claude-fable-5.1-reviewer,
2026-10-13, fresh context. The prompt's inline diff was truncated inside `package-lock.json`, so this review read
the full diff with `git diff master...HEAD -- . ':!frontend/package-lock.json'` in the worktree and re-ran every
gate itself. The earlier review artifact (art_01M4W2SMAMFTH2Z6WCYBSQM2T8, at 1d6008a) and the delegator's routing
comment were read only after the independent pass, to check for already-routed findings.

## 1. Verdict

**PASS**

No blocking findings. One new minor, non-blocking observation on test assertion strength (section 3). The two
minors from the previous review were routed by the delegator; both routings are confirmed correct here with
independent evidence (section 5).

## 2. Summary

The diff delivers what the plan and its Cycle 1 resolutions specify: the 18-package pinned dependency table with
nothing extra, strict tsconfigs with the three required flags on top of the create-vite react-ts template,
type-aware ESLint 9 with react-hooks and jsx-a11y, Prettier, Vitest + RTL + jest-dom with explicit cleanup, the
`/api` and `/healthz` proxy driven by `VITE_API_PROXY_TARGET`, the BRANDING section 2.1 token block byte-for-byte
plus the 3.2 scale and 4.x spacing as custom properties, the 4.2 shell grid, a Header that matches
`docs/branding-demo.html`, four behavioural tests, the Makefile BC-7 block with the R1 sentinel, and the two R3
SPEC edits. The merge commit changed nothing under `frontend/` or `docs/` (identical tree objects) and only
brought in BC-2's Makefile block, so the union of the two blocks is now exercised: the full `make check`
(lint-api, lint-web, test-api, test-web) exits 0 at this HEAD. Code quality is high: small files, comments that
cite BRANDING sections by number, no non-null assertions, named exports, no disabled lint rules.

## 3. Issues

**[MINOR] frontend/src/App.test.tsx:14 — `toEqual` on DOM nodes is structural, not identity**
Vitest's `equals` special-cases DOM nodes: when both sides are nodes it returns `a.isEqualNode(b)`
(`node_modules/vitest/dist/chunks/index.uwTEEBsB.js:1305`). So `expect(Array.from(shell.children)).toEqual([...])`
checks that each child is structurally equal to the expected landmark, not that it is the same element. Today the
five placeholders are structurally distinct (different tag, role or label), and `getByRole` throws on duplicate
matches, so the assertion still proves the BRANDING 4.2 order and cannot pass with the rows swapped. It is a
robustness gap rather than a defect: a future refactor that makes two rows structurally identical would weaken
the test without failing it.
**Fix:** compare by identity, e.g. `expected.forEach((el, i) => expect(shell.children[i]).toBe(el))`, or
`expect(shell.children[0]).toBe(screen.getByRole('banner'))` per row. No rework needed for BC-7; BC-12 replaces
`App.tsx` and rewrites this test, so the cheapest route is a one-line note in the BC-12 hand-off. Suggested
routing: ignore with justification, or fold into BC-12.

No other issues found.

## 4. Positive Observations

- **Plan fidelity, including the review-cycle resolutions.** R1 (sentinel target
  `frontend/node_modules/.package-lock.json`, with `touch` after `npm ci`), R3 (both SPEC edits present with the
  agreed wording in section 2 and section 8), R6 (the `.row` ownership comment in `shell.module.css`), R9 (T2
  asserts `textContent`, not a child count) are all in the diff. The `.PHONY` line lists only the three web
  targets and no other Makefile block was touched by BC-7; the BC-2 merge touched only its own block.
- **Dependency discipline.** `package.json` holds exactly the 18 names in the plan table, every one an exact pin;
  `npm ls --depth=0` shows the same 18 and nothing extraneous; `@testing-library/dom` appears only as a transitive
  peer in the lockfile. The lockfile carries the linux-musl rolldown and lightningcss bindings BC-13 will need
  for `npm ci` on `node:24-alpine`.
- **Tokens match BRANDING to the character.** The 2.1 block is byte-for-byte under `prettier-ignore`; every 3.2
  row (price 28/31 -0.01em, address 22/26, monogram 26/1, body 15/22, description 14/20, meta 13/18, button 13/1
  +0.02em, wordmark 13/16 +0.16em, label 11/14 +0.08em, rank 11/1.3 +0.04em) is present; the 3.1 fallback stacks
  and the Google Fonts href are exact; 4.1 spacing (8px grid, 4px label gap, 24px gutter dropping to 16px at the
  760px breakpoint, 1040px content width, 20px card padding and column gap, 72px monogram) is all named.
- **Header matches the demo, not just the prose.** 56px including the hairline (both the demo and the app use
  border-box sizing), `-webkit-font-smoothing: antialiased` on body as the demo does, scope label styled like the
  demo's `.lbl`. DOM text is `Brillio Listings` uppercased by CSS, exactly as the demo does it, so assistive tech
  reads words. The wordmark is the page `<h1>` inside a `<header>` banner landmark.
- **Shell and scroll model per 4.2.** `html, body, #root { height: 100% }`, `html, body { overflow: hidden }`,
  rows `auto auto auto minmax(0, 1fr) auto`, `.results { min-height: 0; overflow-y: auto; overscroll-behavior:
  contain }`. The `--inline-pad: max(gutter, (100% - 1040px) / 2)` trick centres the column while hairlines run
  full-bleed, with a comment explaining where the `100%` resolves.
- **Proxy verified end-to-end, not just by reading.** With a dummy HTTP server on 127.0.0.1:18077 and
  `VITE_API_PROXY_TARGET=http://127.0.0.1:18077`, `GET /healthz` and `GET /api/cities` on the Vite port returned the
  dummy's 404 and both paths appeared in its access log; `GET /` returned 200 with `lang="en"`, the title and the
  byte-identical fonts href. `loadEnv` with `||` handles unset and empty values; `changeOrigin: true` is right for
  `api:8080` under Compose.
- **Tests assert behaviour, not styles.** Roles, accessible names, heading level, order of the five shell
  children, and `aria-live="polite"` on `main` (SPEC 7.6). `setup.ts` registers jest-dom and an explicit `cleanup()`
  because globals are off; `restoreMocks` and `unstubGlobals` are preset for BC-8's fetch stubs.
- **The lint config is type-aware and strict.** `recommendedTypeChecked`, hooks and a11y rules on every TS/TSX
  file, `projectService` resolving through the solution tsconfig, `--max-warnings 0`.
- **Hygiene.** No emoji in any tracked file under the R10 regex; `git ls-files frontend` equals the files on disk;
  no `.lattice/` changes in the BC-7 diff; working tree clean after the gate; conventional commit messages naming
  the ticket.

## 5. Previously routed findings, re-checked

| Finding (art_01M4W2SMAMFTH2Z6WCYBSQM2T8) | Delegator routing | Independent check here |
|---|---|---|
| `tsconfig.app.json` lib omits `DOM.Iterable` | Skipped: TS 6.0 `lib.dom` already includes the iterable declarations | Confirmed. A temporary `src/__review_probe.ts` with `for...of` over `URLSearchParams` and `NodeListOf`, `Object.fromEntries(params)` and `[...list]` passed `tsc -b --noEmit` (exit 0) with `lib: ["ES2023", "DOM"]`. Probe removed; tree clean. |
| `.results` content narrower than pinned rows by the scrollbar width on classic-scrollbar platforms | Skipped with justification (plan decision 8); hand-off note to BC-10/BC-14 | Agree. Left edges align, which is the edge BRANDING 4.5 cares about; the alternatives move the scrollbar off the viewport edge. Nothing to change in BC-7. |

## Verification record

| Check | Command (from the worktree) | Result |
|---|---|---|
| Full gate at merged HEAD | `make check` | exit 0: golangci-lint clean, eslint clean, `tsc -b --noEmit` clean, Prettier "All matched files use Prettier code style", `go test -race ./...` ok, Vitest 2 files / 4 tests passed |
| Lint + typecheck + format | `make lint-web` | clean |
| Unit tests | `make test-web` | Vitest 5.0.4, 2 files, 4 tests passed |
| Build | `cd frontend && npm run build` | `tsc -b` clean; `dist/` 0.77 kB html, 2.56 kB css, 220.30 kB js (gzip 68.89 kB) |
| TypeScript version | `npx tsc --version` | 6.0.3 |
| Dependency list | `node -p` over package.json, `npm ls --depth=0` | 18 names, exact pins, no extraneous packages |
| Lockfile platform bindings | grep for musl rolldown/lightningcss entries | present (3 hits) |
| Emoji scan | `git ls-files -z frontend Makefile docs/SPEC.md \| xargs -0 perl -CSD -ne '... R10 regex ...'` | no output |
| Tracking | `diff <(git ls-files frontend) <(find frontend ...)` | empty, exit 0 |
| Merge hygiene | `git rev-parse 1d6008a:frontend HEAD:frontend 1d6008a:docs HEAD:docs`; `git diff --stat 1d6008a HEAD -- frontend docs Makefile` | tree objects identical; only `Makefile` (BC-2 block) differs |
| BC-7 touches `.lattice`? | `git diff --stat master...HEAD -- .lattice` | empty |
| Proxy | dummy server on 18077, `vite --port 15177 --strictPort`, curl `/healthz`, `/api/cities`, `/` | 404, 404 (both logged by the dummy), 200; processes killed, scratch removed |
| DOM iterables under current `lib` | temporary probe file, `tsc -b --noEmit`, removed | exit 0 |
| Vitest DOM equality | grep `isEqualNode` in `node_modules/vitest/dist` | `equals` uses `a.isEqualNode(b)` for nodes (basis of the section 3 note) |

## Notes for the orchestrator

- The single minor item is non-blocking and does not require a rework cycle. Recommended routing: ignore with
  justification (BC-12 rewrites the test) or add a one-line note to the BC-12 hand-off.
- This review covers the merged HEAD. The browser-level claims (fonts loaded, computed styles, document does not
  scroll) are proven in the validation artifacts at 1d6008a plus the addendum at 0d1c2a2; nothing in this review
  substitutes for them, and nothing here contradicts them.
- `frontend/dist/` was present before this session (gitignored, another agent's gate output) and was rebuilt by
  the build step; it was left in place. No source edits, commits or Lattice status changes were made.
