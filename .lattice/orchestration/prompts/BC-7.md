# Ticket card — BC-7

```
TICKET=BC-7
TASK_UUID=task_01M4VXK2PXM79XFEJ4SA7B1MXT
TITLE=Frontend scaffold: Vite 8 + React 19 + TS 6.0, lint, Vitest/RTL, API proxy, BRANDING tokens, app shell, Header
WORKTREE=/Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold
BRANCH=feat/BC-7-frontend-scaffold
REVIEW_BASE=master
ANCHOR=none (based on master after BC-1)
PORTS=web 15107 (vite --port 15107 --strictPort); no API needed
AREA=area:9
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-7 notes (binding)

1. **TypeScript is pinned to exactly `6.0.3`** (operator ruling 2026-10-14, SPEC §2 amended; see the BC-7 comment).
   TS 7 exports no compiler API and typescript-eslint does not support it. Every other dependency: newest stable.
   `create-vite` templates may default to TS 7 — override it. Confirm `npx tsc --version` prints 6.0.3.
2. **Dependency allowlist is a hard limit** (ticket Scope bullet 1). Anything else needed (for example
   `@vitejs/plugin-react`, `@eslint/js`, `globals`, `eslint-plugin-react-refresh`, `@types/react`,
   `@types/react-dom`) must be justified in the plan as a template/toolchain requirement and listed in the
   completion comment; the reviewer checks package.json against the ticket.
3. **Gate (run-state Gate rule):** BC-2 runs in parallel and owns the api halves. Your gate is
   `make lint-web test-web` green plus `npm run build`. Full `make check` may fail only on BC-1's api stubs if
   BC-2 has not landed; if it has, merge master and full `make check` must be green.
4. **Makefile:** fill only `dev-web`, `test-web`, `lint-web` (eslint + `tsc --noEmit` + `prettier --check`).
   Do not touch the api targets or the composition lines.
5. **You own** `src/styles/tokens.css`, `global.css`, `shell.module.css`, the test harness (`src/test/setup.ts`) and
   `App.tsx` (placeholder regions; BC-12 replaces them). BC-8 branches from your branch at `review`: after
   `review`, additive commits only, no rebase.
6. Design truth is BRANDING.md (§2.1 tokens verbatim, §3.1 fonts link, §3.2 scale, §4 shell, §5.1 header) with
   `docs/branding-demo.html` as the reference rendering. Do not edit BRANDING.md.
7. Validation: run `npm run dev -- --port 15107 --strictPort` in the worktree's `frontend/`, open it in a c11
   browser panel in your AREA, confirm the header renders on paper ground with Bodoni Moda / Figtree loaded
   (computed font-family) and that the document does not scroll; save a screenshot to EVIDENCE_DIR and attach it.
   Kill the dev server afterwards.
