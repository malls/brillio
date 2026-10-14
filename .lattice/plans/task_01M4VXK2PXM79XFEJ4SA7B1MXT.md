# BC-7 plan: frontend scaffold (Vite 8, React 19, TS 6.0.3, lint, Vitest/RTL, proxy, tokens, shell, Header)

Planner: agent:opus-5.5-planner, 2026-10-14. Worktree `/Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold`,
branch `feat/BC-7-frontend-scaffold` (base master after BC-1). Paths below are worktree-relative.
Card notes override the ticket text: **TypeScript is exactly 6.0.3, not 7**; port 15107; only the three web Make targets.

## Scope

In scope: `frontend/` project (toolchain, configs, tokens, global CSS, shell grid, Header, placeholder App, test harness)
and the BC-7 block of `Makefile`. Out of scope: `src/api|hooks|format|pages`, `src/test/fixtures.ts` (BC-8+), real
FilterBar/StatusLine/Pager (BC-9..11), SearchPage (BC-12), Dockerfile.dev (BC-13), README (BC-15), BRANDING.md (never).
No SPEC edit needed.

## Decisions (verified 2026-10-14 by a clean install of the exact table below in a scratch dir: no ERESOLVE,
then lint + `tsc -b --noEmit` + prettier + vitest + `vite build` all green)

1. **Scaffold, then prune.** From the worktree root: `git rm -q frontend/.gitkeep`, then
   `npm create vite@latest frontend -- --template react-ts --no-interactive --no-immediate` (resolves to create-vite
   9.2.1; its react-ts template ships `typescript ~6.0.2` and **oxlint, not ESLint**). Delete: `.oxlintrc.json`,
   `.gitignore` (root `.gitignore` already covers `node_modules/`, `/frontend/dist/`, `*.tsbuildinfo`,
   `/frontend/coverage/`, `.env.local`), `README.md`, `public/` (favicon.svg, icons.svg), `src/App.css`,
   `src/index.css`, `src/assets/` (hero.png, react.svg, vite.svg). Keep `tsconfig.json` unchanged (solution file).
   Rewrite the rest as specified below. Then replace the dependency blocks with the table and run a clean `npm install`.
2. **ESLint 9.39.5, not 10.12.0.** `eslint-plugin-jsx-a11y@6.10.2` (latest, last published 2024) peers
   `eslint ^3 … ^9`; eslint 10 fails with ERESOLVE. 9.39.5 is the newest compatible release (npm prints
   "This version is no longer supported" on install, expected). Same rule as the TS ruling: no peer workarounds.
3. **Typecheck = `tsc -b --noEmit`.** Plain `tsc --noEmit` on the solution `tsconfig.json` (`files: []`) checks
   nothing (verified: exit 0 with type errors present). `-b` checks `tsconfig.app.json` (all of `src/`, tests
   included) and `tsconfig.node.json` (`vite.config.ts`). tsbuildinfo goes to `node_modules/.tmp/` (ignored).
4. **Type-aware lint:** `tseslint.configs.recommendedTypeChecked` (catches floating promises and unsafe `any`,
   which matter for BC-8/BC-12 fetch and abort code) rather than `strictTypeChecked` (its template-literal rules
   fight the §7.1 formatters). `projectService` resolves through the solution tsconfig (verified on src and vite.config.ts).
5. **Vitest config lives in `vite.config.ts`** (`defineConfig` from `vitest/config`): one config, the React plugin
   applies to tests. No vitest globals: tests import `describe/it/expect` from `vitest`, so RTL's auto-cleanup does
   not register; `setup.ts` calls `cleanup()` in `afterEach`. CSS is not processed in tests (Vitest default); CSS
   module imports return a proxy yielding `_<class>_<hash>`, so tests assert roles and text, never styles.
6. **Proxy target via `loadEnv(mode, process.cwd(), 'VITE_')`**: it returns `VITE_API_PROXY_TARGET` from the process
   env (Compose sets it, SPEC §10) and also from `frontend/.env*.local` for native dev; process env wins. Fallback
   uses `||` so an empty value also falls back to `http://localhost:8080`. `changeOrigin: true` (Host header matches
   `api:8080` under Compose; the Go API does not route on Host).
7. **Wordmark DOM text is `Brillio Listings`, uppercased by CSS** (exactly as `docs/branding-demo.html` does), so
   screen readers say words, not letters; it is the page's `<h1>`. Scope label uses the demo's Label style.
8. **Centering without wrappers:** `--inline-pad: max(var(--gutter), calc((100% - var(--content-width)) / 2))`.
   Rows stay full-bleed (hairlines span the viewport) and content sits in a centered 1040px column with 24px/16px
   minimum gutters; the results scroller keeps its scrollbar at the viewport edge.
9. **Make targets self-install:** a `frontend/node_modules` file target runs `npm ci` when missing or older than
   the lockfile, so `make check` works from a fresh extract (SPEC §12) with no manual step.
10. Exact pins everywhere (`"x.y.z"`, no `^`/`~`): TS must be exactly 6.0.3 and eslint must stay on 9.

## Final dependency table (`frontend/package.json`; nothing else)

| Package | Version | Kind | Justification |
|---|---|---|---|
| react | 19.3.0 | runtime | allowlist (React 19) |
| react-dom | 19.3.0 | runtime | toolchain: React's DOM renderer (`createRoot`); required peer of @testing-library/react |
| vite | 8.3.4 | dev | allowlist (Vite 8) |
| @vitejs/plugin-react | 6.1.2 | dev | template: React JSX transform + Fast Refresh for Vite 8 (hot reload, SPEC §2/§10) |
| typescript | 6.0.3 | dev | allowlist, pinned by operator ruling (SPEC §2) |
| @types/react | 19.3.0 | dev | toolchain: React ships no types; strict `tsc` cannot compile TSX without them |
| @types/react-dom | 19.3.0 | dev | toolchain: types for `react-dom/client` |
| @types/node | 24.19.2 | dev | template: `tsconfig.node.json` `types: ["node"]` for `vite.config.ts` (`process`); 24.x matches Node 24 (SPEC §10), not latest 26.6.5 |
| vitest | 5.0.4 | dev | allowlist (Vitest 5; peers vite ^8) |
| jsdom | 30.1.2 | dev | allowlist |
| @testing-library/react | 16.3.3 | dev | allowlist |
| @testing-library/user-event | 14.6.8 | dev | allowlist (first used by BC-9) |
| @testing-library/jest-dom | 7.0.1 | dev | allowlist |
| eslint | 9.39.5 | dev | toolchain: the runner every listed lint plugin peers on; 9.x per decision 2 |
| typescript-eslint | 8.71.2 | dev | allowlist (peers `typescript >=4.8.4 <6.1.0`, eslint ^9) |
| eslint-plugin-react-hooks | 7.1.1 | dev | allowlist |
| eslint-plugin-jsx-a11y | 6.10.2 | dev | allowlist |
| prettier | 3.9.10 | dev | allowlist |

Dropped: `oxlint` (template), `@eslint/js`, `globals`, `eslint-plugin-react-refresh` (optional extras; TS files
need no globals and type-aware typescript-eslint covers what `js.configs.recommended` would add).
`@testing-library/dom@10.4.2` is **not** listed: npm 11 auto-installs it as the required peer of RTL, user-event and
jest-dom; it appears only in `package-lock.json`.

## Files

Create/rewrite under `frontend/`: `package.json`, `package-lock.json` (committed), `index.html`, `vite.config.ts`,
`tsconfig.app.json`, `tsconfig.node.json`, `eslint.config.js`, `.prettierrc`, `.prettierignore`, `src/main.tsx`,
`src/App.tsx`, `src/App.test.tsx`, `src/test/setup.ts`, `src/styles/tokens.css`, `src/styles/global.css`,
`src/styles/shell.module.css`, `src/components/Header/{Header.tsx,Header.module.css,Header.test.tsx}`
(no `index.ts`: SPEC §7.4 folder = three files). Keep: `frontend/tsconfig.json`. Delete: `frontend/.gitkeep`
plus the template files in decision 1. Modify: `Makefile` (BC-7 block only).
All paths verified not ignored under both the repo rules and the host `~/.gitignore` (`git check-ignore -v`).

## Contents

**package.json:** name `brillio-listings-web`, `private`, `version 0.0.0`, `type module`, deps per table. Scripts:
`dev: vite` · `build: tsc -b --noEmit && vite build` · `preview: vite preview` · `test: vitest run` ·
`lint: eslint . --max-warnings 0` · `typecheck: tsc -b --noEmit` · `format: prettier --write .` ·
`format:check: prettier --check .`

**tsconfig.app.json / tsconfig.node.json:** template files plus `"strict": true, "noUncheckedIndexedAccess": true,
"exactOptionalPropertyTypes": true` in both. Keep the template's `types` (`["vite/client"]` / `["node"]`; TS 6 defaults
`types` to `[]`), `lib`, `jsx: react-jsx`, `noUnused*`, `erasableSyntaxOnly`, `verbatimModuleSyntax`.

**vite.config.ts:** `react()` plugin; `server.proxy` = `'/api'` and `'/healthz'`, each `{ target, changeOrigin: true }`
(two literals) with `target` per decision 6; `test: { environment: 'jsdom', setupFiles: ['./src/test/setup.ts'],
include: ['src/**/*.test.{ts,tsx}'], restoreMocks: true, unstubGlobals: true }` (the last two stop `vi.stubGlobal('fetch')`
leaking between later tickets' tests). Default port 5173 stays (Compose, CORS default).

**eslint.config.js:** `defineConfig([globalIgnores(['dist', 'coverage']), { files: ['**/*.{ts,tsx}'], extends:
[tseslint.configs.recommendedTypeChecked, reactHooks.configs.flat.recommended, jsxA11y.flatConfigs.recommended],
languageOptions: { parserOptions: { projectService: true, tsconfigRootDir: import.meta.dirname } } }])`, imports from
`eslint/config`, `typescript-eslint`, `eslint-plugin-react-hooks`, `eslint-plugin-jsx-a11y`.

**.prettierrc:** `{ "singleQuote": true, "printWidth": 100 }` (indent comes from root `.editorconfig`).
**.prettierignore:** `dist`, `coverage`, `package-lock.json`.

**index.html:** template skeleton with `lang="en"`, charset, viewport, `<link rel="icon" href="data:," />` (no favicon
file, no 404), `<link rel="preconnect" href="https://fonts.googleapis.com" />`,
`<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />`, the BRANDING §3.1 stylesheet with this href
byte-for-byte: `https://fonts.googleapis.com/css2?family=Bodoni+Moda:ital,opsz,wght@0,6..96,400..700;1,6..96,400..700&family=Figtree:wght@400..600&display=swap`,
`<title>Brillio Listings</title>`, `#root`, `/src/main.tsx`.

**src/main.tsx:** imports `./styles/tokens.css` then `./styles/global.css`; `getElementById('root')`, throw
`new Error('Missing #root element')` if null (no `!`); `createRoot(...).render(<StrictMode><App /></StrictMode>)`.
Named exports, extensionless imports.

**src/styles/tokens.css** (BC-7 owns; BC-14 fixes only). First block is the §2.1 block byte-for-byte, preceded by
`/* prettier-ignore */` (Prettier otherwise lowercases the hex and splits lines; verified):
```css
/* prettier-ignore */
:root {
  --paper: #F6F5F1; --stone: #E8E7E1; --line: #D9D7D0;
  --ink: #161616; --ink-2: #6E6C66; --ink-3: #A3A19A;
  --oxblood: #7A1F2B; --oxblood-deep: #5E1620;
  color-scheme: light;
}
```
Second `:root` block, one declaration per line, commented by BRANDING section:
- §3.1 `--font-display: 'Bodoni Moda', 'Bodoni 72', Didot, Georgia, serif;`
  `--font-ui: 'Figtree', system-ui, -apple-system, 'Segoe UI', sans-serif;`
- §3.2 (`--text-*` size, `--leading-*` line-height, `--tracking-*` letter-spacing): price 28px/31px/-0.01em;
  address 22px/26px (state heading reuses it); monogram 26px/1; body 15px/22px; description 14px/20px;
  meta 13px/18px (status line, pager, score); button 13px/1/0.02em; wordmark 13px/16px/0.16em;
  label 11px/14px/0.08em; rank 11px/1.3/0.04em. Page heading 24/29 omitted (BRANDING: not used in the app).
- §4.1/§4.2/§5.1/§5.7 `--grid: 8px; --gap-label: 4px; --gutter: 24px; --content-width: 1040px;`
  `--inline-pad: max(var(--gutter), calc((100% - var(--content-width)) / 2));`
  `--card-padding-block: 20px; --card-column-gap: 20px; --monogram-width: 72px; --header-height: 56px; --pager-height: 48px;`

Then: `/* Breakpoint 760px (BRANDING §4.4). Media queries cannot read custom properties, so components write
@media (max-width: 760px) literally. */` and `@media (max-width: 760px) { :root { --gutter: 16px; } }`.

**src/styles/global.css:** `*, *::before, *::after { box-sizing: border-box }` · `html, body, #root { height: 100% }` ·
`html, body { overflow: hidden }` · `html { color-scheme: light }` (ticket lists it here too) · `body { margin: 0;
background: var(--paper); color: var(--ink); font-family: var(--font-ui); font-size: var(--text-body);
line-height: var(--leading-body); font-optical-sizing: auto; -webkit-font-smoothing: antialiased }` ·
`button, input, select, textarea { font: inherit; color: inherit; letter-spacing: inherit; border-radius: 0 }` ·
`.tabular { font-variant-numeric: tabular-nums }`.

**src/styles/shell.module.css** (BRANDING §4.2): `.shell { display: grid; grid-template-rows: auto auto auto
minmax(0, 1fr) auto; grid-template-columns: minmax(0, 1fr); height: 100% }` · `.row { padding-inline:
var(--inline-pad) }` (filter, status, pager rows) · `.results { min-height: 0; overflow-y: auto;
overscroll-behavior: contain; padding-inline: var(--inline-pad) }`. Hairlines belong to the row components
(§5.1, §5.2, §5.7), not the shell. BC-11's 2px loading line goes inside row 2 or 3; the five-row template stays.

**Header** (`Header({ scope?: string })`, SPEC §7.4; `export interface HeaderProps { scope?: string }`): `<header
className={styles.header}>` (banner landmark) containing `<h1 className={styles.wordmark}>Brillio Listings</h1>` and,
only when `scope !== undefined`, `<span className={styles.scope}>{scope}</span>`. Header.module.css:
`.header { display: flex; align-items: center; justify-content: space-between; gap: 16px; height: var(--header-height);
padding-inline: var(--inline-pad); border-bottom: 1px solid var(--line) }` (56px includes the hairline, as in the demo) ·
`.wordmark { margin: 0; font-size: var(--text-wordmark); line-height: var(--leading-wordmark); font-weight: 600;
letter-spacing: var(--tracking-wordmark); text-transform: uppercase; color: var(--ink) }` · `.scope { font-size:
var(--text-label); line-height: var(--leading-label); font-weight: 500; letter-spacing: var(--tracking-label);
text-transform: uppercase; color: var(--ink-2) }`.

**src/App.tsx** (BC-7 placeholder; BC-12 replaces with `<SearchPage />`): `<div className={shell.shell}>` with five
children in §4.2 order: `<Header scope="Northern Virginia" />` (BRANDING §5.1) · `<div className={shell.row}
role="search" aria-label="Filters" />` · `<div className={shell.row} role="status" />` · `<main className=
{shell.results} aria-label="Results" aria-live="polite" />` (SPEC §7.6) · `<nav className={shell.row}
aria-label="Pagination" />`. Accessible names only; no visible copy beyond §5.1.

**src/test/setup.ts:** `import '@testing-library/jest-dom/vitest'`; `afterEach(() => { cleanup(); })` with
`cleanup` from `@testing-library/react`, `afterEach` from `vitest`. (BC-8 adds `fixtures.ts` beside it.)

**Makefile, BC-7 block only** (TAB-indented recipes; api block, composite lines and other blocks untouched):
```make
# --- Frontend (BC-7) ---
# Recipes run in frontend/. frontend/node_modules runs `npm ci` when missing or older than the lockfile.
.PHONY: dev-web lint-web test-web
dev-web: frontend/node_modules
	cd frontend && npm run dev
lint-web: frontend/node_modules
	cd frontend && npm run lint && npm run typecheck && npm run format:check
test-web: frontend/node_modules
	cd frontend && npm test
frontend/node_modules: frontend/package-lock.json
	cd frontend && npm ci
	touch frontend/node_modules
```

## Steps

1. Scaffold and prune (decision 1). 2. Write package.json deps/scripts; `rm -rf node_modules package-lock.json &&
npm install` (clean install keeps every platform's rolldown/lightningcss bindings in the lock, needed by BC-13's
`npm ci` on node:24-alpine; verified present). 3. Configs, styles, Header, App, tests. 4. `npm run format`, then the
gate. 5. Check tracking: `diff <(git ls-files frontend | sort) <(find frontend -type f -not -path 'frontend/node_modules/*'
-not -path 'frontend/dist/*' | sort)` prints nothing after `git add frontend Makefile`. Never stage `.lattice/`.
6. One conventional commit, e.g. `feat(frontend): scaffold Vite app shell, tokens and Header (BC-7)`; no emoji.

## Tests and checks, mapped to the ticket's Acceptance

| Acceptance / SPEC | Proof |
|---|---|
| A1 header on paper ground, fonts loaded | Validation (below). Unit support: Header T1–T3, App T4 |
| A2 `npm run build`, `npm test`, `make lint` pass | Gate (below) |
| A3 only listed deps | `node -p "Object.keys({...require('./package.json').dependencies,...require('./package.json').devDependencies}).sort().join('\n')"` equals the table's 18 names |
| A4 no emoji (BRANDING §8) | emoji scan below prints nothing (macOS grep has no `-P`, so perl) |
| SPEC §8 "each component: renders from props, roles/labels, copy verbatim" | T1–T3 |
| SPEC §8 lint: eslint, tsc, prettier | `make lint-web` |
| `npx tsc --version` | prints `Version 6.0.3` |

Emoji scan (from the worktree root):
`git ls-files -z frontend Makefile | xargs -0 perl -CSD -ne 'print "$ARGV:$.\n" if /[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]/; close ARGV if eof'`

- **T1** Header: `getByRole('banner')` contains `getByRole('heading', { level: 1, name: 'Brillio Listings' })`.
- **T2** Header without `scope`: the banner has exactly one child element (no scope label).
- **T3** Header with `scope="Northern Virginia"`: `within(banner).getByText('Northern Virginia')`.
- **T4** App: the shell's children are, in order, banner, search "Filters", status, main "Results", navigation
  "Pagination" (`expect(Array.from(shell.children)).toEqual([...])`), and main has `aria-live="polite"`.

## Gate (run-state Gate rule)

In `frontend/`: `npx tsc --version` (6.0.3), `npm run build`. From the root: `make lint-web test-web` exits 0.
If BC-2 has not landed, `make -k check` may fail only on the `lint-api`/`test-api` "not yet implemented" stubs; if it
has, merge master and full `make check` must be green.

## Validation (no API needed; ports per run-state: web 15107, dummy target 18107)

1. In the worktree `frontend/`: start a dummy target `python3 -m http.server 18107 --bind 127.0.0.1 --directory <empty
   scratch dir>` and `VITE_API_PROXY_TARGET=http://127.0.0.1:18107 npm run dev -- --port 15107 --strictPort`
   (use 127.0.0.1: Node may resolve `localhost` to ::1). `curl -s -o /dev/null -w '%{http_code}' localhost:15107/healthz`
   and `/api/cities` return the dummy server's 404 (its log shows both paths), proving the §10 proxy and env var.
2. Load the c11-browser skill; open a browser panel in **area:9** (`c11 new-panel --type browser --area area:9`) at
   `http://localhost:15107/`. Confirm by JS eval: body background `rgb(246, 245, 241)`; header height 56 with a 1px
   `rgb(217, 215, 208)` bottom border; h1 computed `font-family` starts with `Figtree`, weight 600, uppercase, letter
   spacing about 2.08px; scope "Northern Virginia" in `rgb(110, 108, 102)`; `document.fonts.check('600 13px Figtree')`
   is true; after `await document.fonts.load('400 22px "Bodoni Moda"')`, `document.fonts.check('400 22px "Bodoni Moda"')`
   is true (Bodoni downloads only when used and the placeholder has no Bodoni text); header `padding-left` equals
   `max(24, (innerWidth - 1040) / 2)` px.
3. Screenshot to `/Users/forrest/Code/brillio/.lattice/orchestration/evidence/BC-7-header.png`.
4. Scroll proof: append a 3000px-tall div to `main`, `window.scrollTo(0, 1000)`; `window.scrollY` and
   `document.scrollingElement.scrollTop` stay 0, header top stays 0, and `main.scrollTop = 500` reads back 500.
5. Write `EVIDENCE_DIR/BC-7-validation.md` (commands, observed values, PASS/FAIL, HEAD sha); attach it and the
   screenshot with `--role validation`. Kill the dev server and the dummy server; close only the browser panel.

## Shared-file handling (run-state table)

`Makefile`: BC-7 block only; BC-2 edits the api block in parallel, so the union merges cleanly. `tokens.css`,
`global.css`: BC-7 creates, BC-14 fixes only. `App.tsx`: placeholder, BC-12 replaces. `shell.module.css` and
`src/test/setup.ts`: BC-7 owns; later tickets import, not fork. BC-8 branches from this branch at `review`: after
`review`, additive commits only, no rebase or amend. Do not create `src/format/*` or `src/test/fixtures.ts`.

## Risks

- ESLint 9.39.5 is end of life (npm deprecation warning). Upgrade path: eslint 10 once jsx-a11y declares support
  (`overrides` would work today; `jsx-a11y/alt-text` fires under eslint 10.12.0 in a scratch check, but that is a workaround).
- TS 6 defaults: `types` is `[]`, so every global type package is listed in `types`; `strict` is set explicitly.
- jsdom loads no fonts and Vitest applies no CSS: style and font claims are proven only in the browser step.
- Host `~/.gitignore` (`*.d`, `*.tmp`, `node_modules`, `notes.md`, ...): no planned path matches; step 5 diff catches
  any surprise (fix with `git add -f`). `vite-env.d.ts` is not created (template uses `types: ["vite/client"]`).
- `make lint-web` on a fresh clone needs network for `npm ci` (one time).

## Size

About 480 LOC excluding `package-lock.json`: configs ~150, CSS ~150, TSX/tests ~150, Makefile ~12, index.html ~20.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Review: fallback sub-agent (headless `lattice plan-review` hit the 600 s watchdog), artifact
`art_01M4W1VMFHB6FMYMB8P8981EAT`, verdict APPROVE-WITH-CHANGES. Triage by agent:opus-5.5-delegator.

| # | Concern | Resolution | Section |
|---|---|---|---|
| R1 (Major) | `frontend/node_modules` dir target is satisfied by an empty dir (Compose anonymous volume creates one on the host), so `npm ci` is skipped and lint fails | **Fixed.** Sentinel target is `frontend/node_modules/.package-lock.json: frontend/package-lock.json`, recipe `cd frontend && npm ci` then `touch frontend/node_modules/.package-lock.json`. `dev-web`, `lint-web`, `test-web` depend on the sentinel. Stays inside the BC-7 block; `.PHONY` line lists only the three web targets. | Makefile, Decision 9 |
| R2 (Major) | `document.fonts.check()` is true when no FontFace matches, so the "fonts loaded" proof is vacuous | **Fixed.** Validation step 2 now: `await document.fonts.ready`; assert `(await document.fonts.load('600 13px Figtree')).length > 0` and `(await document.fonts.load('400 22px "Bodoni Moda"')).length > 0`; assert `[...document.fonts]` has a face with family `Figtree` and one with `Bodoni Moda`, each `status === 'loaded'`; record the `fonts.gstatic.com` woff2 entries from `performance.getEntriesByType('resource')` in the evidence. Computed `font-family` of the h1 is still recorded. | Validation step 2 |
| R3 (Major) | Version and command deviations not recorded per the SPEC preamble | **Fixed.** Replace "No SPEC edit needed" with two in-branch SPEC edits made by the implementer: **SPEC edit: §2** — add a note under the TypeScript note: "ESLint version (2026-10-14): eslint-plugin-jsx-a11y 6.10.2 (latest) peers `eslint ^3 … ^9`, so the frontend pins ESLint 9.39.5, the newest 9.x, with no peer overrides (typescript-eslint and react-hooks already accept ^10). Moving to ESLint 10 is a one-line bump once jsx-a11y supports it. `@types/node` tracks the Node 24 runtime (24.x), not the newest major." **SPEC edit: §8** — the lint line reads `tsc -b --noEmit` instead of `tsc --noEmit`, with a parenthetical: "plain `tsc --noEmit` on the solution tsconfig checks no files". The delegator records both edits in a `lattice comment` and lists the non-allowlisted packages (react-dom, @vitejs/plugin-react, @types/react, @types/react-dom, @types/node, eslint) with their table justifications in the completion comment, and mails the Orchestrator an FYI about ESLint 9. | Scope, Decisions 2-3, dependency table |
| R4 (Minor) | react-hooks 7.1.1 `configs.flat.recommended` errors on `set-state-in-effect`, which a `setState('loading')`-in-effect `useListingSearch` would hit | **Decided: keep the recommended config unchanged** (allowlisted plugin, its default rules, no disables). Hand-off note for BC-8 (in the completion comment): implement SPEC §7.2 with `useReducer` and `dispatch` inside the effect (verified to pass), no inline disables. | Decision 4, eslint.config.js |
| R5 (Minor) | Padding check ignores the 760px breakpoint | **Fixed.** Expected header `padding-left` = `max(innerWidth <= 760 ? 16 : 24, (innerWidth - 1040) / 2)` px; record `innerWidth` in the evidence. | Validation step 2 |
| R6 (Minor) | `shell.row` usage by later row components is ambiguous | **Fixed.** Add a comment at the top of `shell.module.css` and state in the hand-off: row components (Header, FilterBar, StatusLine, Pager) are direct children of `.shell` and apply `padding-inline: var(--inline-pad)` and their own hairlines themselves, as Header does; `.row` exists only for BC-7's placeholder regions and is not to be wrapped around real components. | shell.module.css, Decision 8 |
| R7 (Nit) | Dummy proxy port 18107 drifts from run-state's 18000+N | **Fixed.** Use 18007 (`VITE_API_PROXY_TARGET=http://127.0.0.1:18007`). | Validation step 1 |
| R8 (Nit) | Peer-range notes imprecise | **Fixed.** vitest 5.0.4 peers vite `^6.4 \|\| ^7 \|\| ^8`; typescript-eslint 8.71.2 peers eslint `^8.57 \|\| ^9 \|\| ^10`. jsx-a11y is the only package blocking ESLint 10. | Dependency table, Decision 2 |
| R9 (Nit) | T2 depends on the exact child count | **Fixed.** T2 asserts `banner` `textContent` is exactly `Brillio Listings` (no scope text) and `queryByText` for any scope is absent. | Tests T2 |
| R10 (Nit) | Emoji scan gaps | **Fixed.** Regex is `/[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{2B00}-\x{2BFF}\x{FE0F}]/`. | Emoji scan |

Size estimate unchanged (~480 LOC plus ~6 lines of SPEC.md).
