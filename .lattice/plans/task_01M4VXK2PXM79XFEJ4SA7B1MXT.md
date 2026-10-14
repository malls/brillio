# BC-7: Frontend scaffold: Vite 8 + React 19 + TS 7, lint, Vitest/RTL, API proxy, BRANDING tokens, app shell, Header

Create the frontend project with the exact toolchain, the design tokens, the fonts, the pinned app shell and the Header. Spec: docs/SPEC.md §2, §3, §7.6, §10 (proxy). Design: BRANDING.md §2.1 (tokens), §2.5, §3.1 (fonts), §3.2 (scale), §4 (grid, shell, breakpoints), §5.1 (Header). Reference rendering: docs/branding-demo.html.

Why: every component task (BC-9, BC-10, BC-11) needs tokens, the type scale and the test harness in place, and the shell grid is what makes "app, not document" true.

Scope:
- `npm create vite@latest frontend -- --template react-ts`, then pin latest stable: React 19, Vite 8, TypeScript 7, Vitest 5, @testing-library/react, @testing-library/user-event, @testing-library/jest-dom, jsdom, typescript-eslint, eslint-plugin-react-hooks, eslint-plugin-jsx-a11y, prettier. No other runtime or dev dependencies.
- tsconfig strict plus noUncheckedIndexedAccess and exactOptionalPropertyTypes. ESLint flat config. .prettierrc. vitest config with jsdom and src/test/setup.ts (jest-dom matchers).
- vite.config.ts: proxy /api and /healthz to VITE_API_PROXY_TARGET, default http://localhost:8080.
- index.html: `<title>Brillio Listings</title>`, the BRANDING §3.1 Google Fonts link, lang="en".
- src/styles/tokens.css: the §2.1 :root block verbatim, plus `--font-display`, `--font-ui`, the §3.2 sizes/leading as custom properties, spacing (8px grid, 24/16px gutters, 1040px content width, 760px breakpoint noted in a comment since media queries cannot read custom properties).
- src/styles/global.css: html/body height 100% and overflow hidden, paper background, ink text, Figtree body at 15/22, `font-optical-sizing: auto`, `color-scheme: light`, a `.tabular` utility, reset for buttons/inputs to inherit font.
- src/styles/shell.module.css: §4.2 grid with rows `auto auto auto minmax(0, 1fr) auto`, results region `overflow-y: auto; overscroll-behavior: contain`, 1040px centered content, gutters per §4.1.
- src/components/Header: wordmark BRILLIO LISTINGS per §3.2 (Figtree 13/600, +0.16em, uppercase), optional scope label at right in ink-2, 56px tall, hairline below. Test: renders wordmark and optional scope.
- App.tsx renders the shell with Header and placeholder regions labelled by role.
- Makefile: dev-web (npm run dev), frontend halves of test, lint (eslint + tsc --noEmit + prettier --check), check.

Acceptance:
- `npm run dev` shows the header on paper ground with the fonts loaded (check computed font-family in devtools).
- `npm run build`, `npm test`, and `make lint` all pass.
- package.json contains only the listed dependencies.
- No emoji anywhere (BRANDING §8).

Validation: open http://localhost:5173 in a browser (c11-browser skill), confirm the header renders and the document itself does not scroll; attach a screenshot with --role validation.
