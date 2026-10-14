# BC-7 Validation Addendum: re-proof at the merged HEAD

- Validator: `agent:opus-5.5-validator`
- Worktree: `/Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold` (branch `feat/BC-7-frontend-scaffold`)
- Merged HEAD: `0d1c2a2a85c4ef23fb3da77452fa1387095f4986` ("chore: merge master (BC-2 backend scaffold) into feat/BC-7-frontend-scaffold"), confirmed with `git rev-parse HEAD`
- Previously validated HEAD: `1d6008a32c663486c6fbed355d320aa80ec240ca` (full evidence: `BC-7-validation.md`, screenshot `BC-7-header.png`)
- Date: 2026-10-14T02:17Z to 02:19Z (2026-10-13 22:17 to 22:19 EDT)
- Ports (run-state): web 15107 `--strictPort`, dummy proxy target 127.0.0.1:18007. Both confirmed free before start and after cleanup.

## Overall verdict: PASS

## Reasoning

After the full validation passed at `1d6008a`, master (BC-2 backend scaffold) was merged in. The merge touched only
`backend/`, `.gitignore`, `.lattice/` and the BC-2 block of `Makefile`. The `frontend/` and `docs/` trees are
byte-identical between the two commits (same git tree objects, below). The browser-level checks in
`BC-7-validation.md` (computed styles, loaded fonts, gutter, shell landmarks, screenshot, no document scroll) depend
only on `frontend/` source and its installed dependencies, so they carry over unchanged. This addendum re-proves the
parts the merge could plausibly affect: tree identity, ignore rules after the `.gitignore` change, the served app and
dev proxy at the merged HEAD, and the frontend gate (`make lint-web test-web`, which now sits next to a changed BC-2
Makefile block, plus `npm run build`).

## 1. Tree identity: PASS

```bash
git diff --stat 1d6008a32c663486c6fbed355d320aa80ec240ca HEAD -- frontend docs
git rev-parse 1d6008a...:frontend HEAD:frontend ; git rev-parse 1d6008a...:docs HEAD:docs
git diff --stat 1d6008a32c663486c6fbed355d320aa80ec240ca HEAD -- . ':!.lattice'
git diff 1d6008a32c663486c6fbed355d320aa80ec240ca HEAD -- Makefile .gitignore
git -c core.excludesFile=/dev/null check-ignore -v frontend/src/App.tsx frontend/package-lock.json \
  frontend/src/styles/tokens.css frontend/src/components/Header/Header.tsx frontend/index.html frontend/package.json
```

| Check | Expected | Observed | Result |
|---|---|---|---|
| `git diff --stat ... -- frontend docs` | empty | empty (0 bytes, raw output via `rtk proxy`) | PASS |
| `frontend` tree object | identical | `0b59ddb689d6c24e69a219b087a8454d62eb10ea` at both commits | PASS |
| `docs` tree object | identical | `6e148675642b40519c60735efec5fe0ffe9f25b8` at both commits | PASS |
| Non-`.lattice` diff | only `backend/`, `.gitignore`, `Makefile` | 19 files: `.gitignore` (+5 -1), `Makefile` (+5 -3), 17 files under `backend/`; 1639 insertions, 4 deletions | PASS |
| Makefile hunk | only the BC-2 block | single hunk `@@ -14,12 +14,14 @@` inside `# --- Backend (BC-2) ---`: adds a comment line and replaces the three "not yet implemented" stubs for `dev-api`, `lint-api`, `test-api` with real recipes. The `# --- Frontend (BC-7) ---` section is context only, unchanged | PASS |
| `.gitignore` hunk | backend-only | adds `/backend/api`, `/backend/cover.out`, `/backend/coverage.html` and rewords the Go comment; nothing under the Node section | PASS |
| `check-ignore -v` on 6 frontend source paths | prints nothing | no output, exit 1 (none ignored) | PASS |

## 2. Live smoke at the merged HEAD: PASS

```bash
lsof -nP -iTCP:15107 -sTCP:LISTEN ; lsof -nP -iTCP:18007 -sTCP:LISTEN        # both empty
python3 -m http.server 18007 --bind 127.0.0.1 --directory <scratch>/empty > <scratch>/dummy.log 2>&1 &   # PID 26399
cd frontend && VITE_API_PROXY_TARGET=http://127.0.0.1:18007 ./node_modules/.bin/vite --port 15107 --strictPort > <scratch>/vite.log 2>&1 &   # PID 26400
#   vite.log: "VITE v8.3.4  ready in 182 ms  ->  Local: http://localhost:15107/"; lsof: node 26400 on [::1]:15107, Python 26399 on 127.0.0.1:18007
curl -s -o index.out   -w '%{http_code}' http://localhost:15107/
curl -s -o healthz.out -w '%{http_code}' http://localhost:15107/healthz
curl -s -o cities.out  -w '%{http_code}' http://localhost:15107/api/cities
curl -s -o tokens.out  -w '%{http_code} %{content_type}' http://localhost:15107/src/styles/tokens.css
curl -s -o header.out  -w '%{http_code} %{content_type}' http://localhost:15107/src/components/Header/Header.tsx
kill 26400 26399
```

| Check | Expected | Observed | Result |
|---|---|---|---|
| `GET /` | 200 | `200` | PASS |
| `<title>` | `Brillio Listings` | `<title>Brillio Listings</title>` | PASS |
| `lang` | `en` | `<html lang="en"` | PASS |
| Google Fonts href | byte-identical to BRANDING.md section 3.1 | `grep -F -c` on the full expected `href="..."` = `1`; extracted: `https://fonts.googleapis.com/css2?family=Bodoni+Moda:ital,opsz,wght@0,6..96,400..700;1,6..96,400..700&family=Figtree:wght@400..600&display=swap` | PASS |
| `GET /healthz` via 15107 | dummy's 404 | `404`, body is Python http.server's error page | PASS |
| `GET /api/cities` via 15107 | dummy's 404 | `404`, body `<title>Error response</title>` (Python http.server) | PASS |
| Dummy log shows both paths | both | `[13/Oct/2026 22:18:19] "GET /healthz HTTP/1.1" 404 -` and `[13/Oct/2026 22:18:19] "GET /api/cities HTTP/1.1" 404 -` | PASS |
| `GET /src/styles/tokens.css` | 200, contains `--paper: #F6F5F1` | `200 text/javascript` (Vite's CSS module wrapper), 1 match; context `--paper: #F6F5F1; --stone: #E8E7E1; --line: #D9D7D0;` | PASS |
| `GET /src/components/Header/Header.tsx` | 200, Vite-transformed | `200 text/javascript`, 3912 bytes; starts with `import { createHotContext as __vite__createHotContext } from "/@vite/client"`, uses `_jsxDEV` from `react/jsx-dev-runtime` (4 matching lines), imports `/src/components/Header/Header.module.css`, contains `Brillio Listings` | PASS |
| Shutdown | both PIDs gone, ports free | `ps -p 26399,26400` empty; `lsof` on 15107 and 18007 empty | PASS |

## 3. Gate at the merged HEAD: PASS

```bash
make lint-web test-web            # from the worktree root
cd frontend && npm run build
rm -rf frontend/dist
```

| Check | Expected | Observed | Result |
|---|---|---|---|
| `make lint-web test-web` exit | 0 | `0` | PASS |
| lint-web | eslint, typecheck, prettier clean | `eslint . --max-warnings 0` clean; `tsc -b --noEmit` clean; `prettier --check .`: "All matched files use Prettier code style!" | PASS |
| test-web | all pass | vitest v5.0.4: Test Files 2 passed (2), Tests 4 passed (4) | PASS |
| `npm run build` exit | 0 | `0`; `tsc -b --noEmit && vite build` (vite v8.3.4): 20 modules; `dist/index.html` 0.77 kB, `dist/assets/index-Bi9deQIM.css` 2.56 kB, `dist/assets/index-C_W2cHaM.js` 220.30 kB (gzip 68.89 kB); built in 102 ms | PASS |
| `frontend/dist` removed | absent | `ls` reports No such file or directory | PASS |

Note: a `frontend/dist/` from an earlier gate run (mtime 22:06:38 EDT, gitignored) was present before this session's
build; the build overwrote it and it was then removed as instructed.

## Carried over from `BC-7-validation.md` (at 1d6008a; frontend/ and docs/ unchanged since)

Browser checks in the c11 panel: paper/ink computed colors, 56px header with 1px hairline, h1 Figtree 600 13px
uppercase tracked, Northern Virginia scope label in ink-2, Figtree and Bodoni Moda faces `loaded`, 24px gutter, shell
landmark order (banner, search "Filters", status, main "Results" aria-live polite, navigation "Pagination"), header
screenshot `BC-7-header.png`, and the document-does-not-scroll probe. Not re-run: they are a function of the identical
`frontend/` tree, and steps 2 and 3 above show the same tree still serves and builds at the merged HEAD.

## Cleanup

Vite (26400) and the dummy server (26399) killed by captured PID; ports 15107 and 18007 free. Scratch directory
removed. `git status --porcelain` in the worktree empty. No source edits, commits or Lattice status changes.
