# BC-7 Validation: frontend scaffold (end-to-end, running dev server, c11 browser)

- Validator: `agent:opus-5.5-validator`
- Worktree: `/Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold` (branch `feat/BC-7-frontend-scaffold`)
- HEAD: `1d6008a32c663486c6fbed355d320aa80ec240ca` (confirmed with `git rev-parse HEAD`; working tree clean before and after)
- Date: 2026-10-14T02:11Z to 02:16Z (2026-10-13 22:11 to 22:16 EDT)
- Toolchain observed: Vite 8.3.4 (node v24.21.0), TypeScript 6.0.3, React 19.3.0; c11 1.1.1 (134) browser panel (WKWebView), macOS 27.0.1
- Ports (run-state): web 15107 `--strictPort`, dummy proxy target 127.0.0.1:18007. Both confirmed free before start and after cleanup.

## Overall verdict: PASS

Every check passed. One sub-item of R2 (listing `fonts.gstatic.com` entries from Resource Timing) could not be
produced as literally specified because WebKit emits no Resource Timing entries for CSS `@font-face` loads; this
is an engine limitation, not an implementation defect. It is documented below with substitute proof (cold-profile
fetch of the same woff2 URLs, FontFace `loaded` status for the matching latin subsets, and glyph-metric deltas).
The R2 assertions themselves (`fonts.load(...).length > 0`, faces `status === 'loaded'`) pass.

## Commands run

```bash
# ports free
lsof -nP -iTCP:15107 -sTCP:LISTEN ; lsof -nP -iTCP:18007 -sTCP:LISTEN        # both empty

# dummy target (empty scratch dir)
python3 -m http.server 18007 --bind 127.0.0.1 --directory <scratch>/empty > <scratch>/dummy.log 2>&1 &

# dev server (identical to `npm run dev -- --port 15107 --strictPort`; package.json "dev": "vite")
cd frontend && VITE_API_PROXY_TARGET=http://127.0.0.1:18007 ./node_modules/.bin/vite --port 15107 --strictPort > <scratch>/vite.log 2>&1 &
#   log: "VITE v8.3.4  ready in 174 ms  ->  Local: http://localhost:15107/"  (listening on [::1]:15107)

curl -s -o /dev/null -w '%{http_code}' http://localhost:15107/healthz
curl -s -o /dev/null -w '%{http_code}' http://localhost:15107/api/cities
curl -s -w '%{http_code}' http://localhost:15107/

# browser
c11 --json new-panel --type browser --area area:9 --url http://localhost:15107/      # -> panel:22
c11 rename-panel --panel panel:22 "BC-7 Validation Browser"
c11 browser panel:22 wait --load-state complete ; c11 browser panel:22 wait --selector "header h1"
c11 --json browser panel:22 eval --script "<step 2 script>"        # computed styles, fonts, gutter, shell children
c11 --json browser panel:22 screenshot --out .../evidence/BC-7-header.png
c11 --json browser panel:22 eval --script "<step 4 script>"        # scroll probe
# supplementary (R2 resource-timing gap): temporary profile bc7-validation-cold, panel:23, then closed and deleted
```

## Step 1: Dev proxy (SPEC section 10) and index.html: PASS

| Check | Expected | Observed | Result |
|---|---|---|---|
| `GET /healthz` via 15107 | dummy's 404 | `404` | PASS |
| `GET /api/cities` via 15107 | dummy's 404 | `404` | PASS |
| Dummy log shows both paths | both | `"GET /healthz HTTP/1.1" 404 -` and `"GET /api/cities HTTP/1.1" 404 -` (22:11:56) | PASS |
| `GET /` | 200 | `200` | PASS |
| `<title>` | `Brillio Listings` | `<title>Brillio Listings</title>` | PASS |
| `lang` | `en` | `<html lang="en">` | PASS |
| Google Fonts href equals BRANDING.md section 3.1 (line 113) | identical | `https://fonts.googleapis.com/css2?family=Bodoni+Moda:ital,opsz,wght@0,6..96,400..700;1,6..96,400..700&family=Figtree:wght@400..600&display=swap` (byte-identical) | PASS |

This proves `VITE_API_PROXY_TARGET` is read from the process env and both `/api` and `/healthz` are proxied.

## Step 2: Browser (c11 panel:22, area:9, `http://localhost:15107/`): PASS

Computed values via `getComputedStyle` in the live page:

| Check | Expected | Observed | Result |
|---|---|---|---|
| body `background-color` (paper) | `rgb(246, 245, 241)` | `rgb(246, 245, 241)` | PASS |
| body `color` (ink) | `rgb(22, 22, 22)` | `rgb(22, 22, 22)` | PASS |
| header height | 56 | `getBoundingClientRect().height` 56, `offsetHeight` 56 | PASS |
| header border-bottom | 1px `rgb(217, 215, 208)` | `1px solid rgb(217, 215, 208)` | PASS |
| h1 text | `Brillio Listings` | `Brillio Listings` (a11y snapshot: heading "BRILLIO LISTINGS") | PASS |
| h1 `font-family` | starts with `Figtree` | `Figtree, system-ui, -apple-system, Segoe UI, sans-serif` | PASS |
| h1 `font-weight` | 600 | `600` | PASS |
| h1 `text-transform` | uppercase | `uppercase` | PASS |
| h1 `letter-spacing` | about 2.08px (0.16em x 13px) | `2.08px` | PASS |
| h1 `font-size` | 13px | `13px` | PASS |
| scope label `Northern Virginia` color (ink-2) | `rgb(110, 108, 102)` | `<span>` text `Northern Virginia`, color `rgb(110, 108, 102)` (font-size 11px, Figtree) | PASS |

### Fonts actually loaded (R2): PASS

| Check | Observed | Result |
|---|---|---|
| `await document.fonts.ready` | resolved | PASS |
| `(await document.fonts.load('600 13px Figtree')).length > 0` | `1` | PASS |
| `(await document.fonts.load('400 22px "Bodoni Moda"')).length > 0` | `1` | PASS |
| `[...document.fonts]` has family `Figtree` with `status === 'loaded'` | 1 face: `Figtree`, normal, weight `400 600`, unicode-range `U+0-ff, U+131, ...` (latin), `loaded` | PASS |
| `[...document.fonts]` has family `Bodoni Moda` with `status === 'loaded'` | 1 face: `Bodoni Moda`, normal, weight `400 700`, unicode-range `U+0-ff, U+131, ...` (latin), `loaded` | PASS |
| Other faces | 10 FontFace objects total from the Google CSS (8 Bodoni Moda across subsets and italic, 2 Figtree); the non-latin and italic subsets stay `unloaded`, as expected for unused ranges | info |
| `fonts.gstatic.com` entries in `performance.getEntriesByType('resource')` | **none** (`[]`), in both the Default profile and a fresh cold profile. The googleapis CSS entry is present (`initiatorType: link`; 145 ms in the cold profile, so it came from the network). | see note |

Note on the Resource Timing sub-item (engine limitation, not a defect): WKWebView does not create Resource Timing
entries for fonts fetched by CSS `@font-face`. Proof that the engine records gstatic entries when it is the
initiator: after the page `fetch()`ed the two woff2 URLs below, two `fonts.gstatic.com` entries appeared with
`initiatorType: "fetch"`, while the `@font-face` loads that had already set the faces to `loaded` produced none. To
rule out caching, the same check was run in a temporary, empty browser profile (`bc7-validation-cold`, panel:23,
since closed and deleted): same result.

Substitute evidence that the woff2 files are fetched and in use (cold profile, identical in the Default profile):

| Item | Observed |
|---|---|
| Figtree latin woff2 (from the page's own Google CSS, `U+0000-00FF` block) | `https://fonts.gstatic.com/s/figtree/v9/_Xms-HUzqDCFdgfMm4S9DaRvzig.woff2`: HTTP 200, `font/woff2`, 20184 bytes, signature `wOF2` |
| Bodoni Moda latin woff2 | `https://fonts.gstatic.com/s/bodonimoda/v28/aFTQ7PxzY382XsXX63LUYJSKSKjWXFBP.woff2`: HTTP 200, `font/woff2`, 46248 bytes, signature `wOF2` |
| Glyph metrics, `BRILLIO LISTINGS` at 600 13px | Figtree 107.96 px vs system-ui 114.01 px (the Figtree face renders, not the fallback) |
| Glyph metrics, `Brillio Listings 0123456789` at 400 22px | Bodoni Moda 270.98 px vs serif 249.35 px (the Bodoni Moda face renders) |

### Gutter (R5): PASS

| Check | Observed | Result |
|---|---|---|
| `innerWidth` | `763` (innerHeight 430) | info |
| expected `max(innerWidth <= 760 ? 16 : 24, (innerWidth - 1040) / 2)` | `max(24, -138.5)` = `24px` | info |
| header `padding-left` | `24px` (padding-right `24px`) | PASS |

Only the `> 760px` branch is observable at this panel width (WKWebView has no viewport emulation). The `<= 760px`
branch is the `@media (max-width: 760px)` rule in `src/styles/tokens.css` (not exercised here).

### Shell children, in order: PASS

The shell is `div._shell_1ltl3_5` (`display: grid`, rows `56px 0px 0px 374px 0px`, height 430 = viewport), a
direct child of `div#root`.

| # | Expected | Observed | Result |
|---|---|---|---|
| 1 | banner | `<header>` (not inside article/aside/main/nav/section, so it is the banner landmark; the only `header`, the only `h1`) | PASS |
| 2 | search "Filters" | `<div role="search" aria-label="Filters">` | PASS |
| 3 | status | `<div role="status">` | PASS |
| 4 | main "Results", aria-live polite | `<main aria-label="Results" aria-live="polite">` | PASS |
| 5 | navigation "Pagination" | `<nav aria-label="Pagination">` | PASS |

## Step 3: Screenshot: PASS

`/Users/forrest/Code/brillio/.lattice/orchestration/evidence/BC-7-header.png` (1526x860 PNG, 2x of the 763x430
viewport). Looked at it: the warm off-white paper ground fills the panel; "BRILLIO LISTINGS" sits at the left in
dark tracked uppercase sans; "NORTHERN VIRGINIA" sits at the right in a muted grey; one thin hairline runs the full
width under the 56px header; the area below is empty paper (placeholder regions).

## Step 4: The document does not scroll: PASS

Probe: a 3000px-tall div appended to `main`, then `window.scrollTo(0, 1000)` and
`document.scrollingElement.scrollTop = 1000`, then `main.scrollTop = 500`. The probe was removed and `main.scrollTop`
reset to 0 afterwards (confirmed `probeRemoved: true`).

| Check | Expected | Observed | Result |
|---|---|---|---|
| document scrollHeight vs clientHeight with the probe present | equal (nothing to scroll) | 430 / 430 | PASS |
| main scrollHeight vs clientHeight with the probe | main overflows | 3000 / 374 | info |
| `window.scrollY` after scrollTo(0, 1000) | 0 | `0` | PASS |
| `document.scrollingElement.scrollTop` (`html`) | 0 | `0` | PASS |
| header `getBoundingClientRect().top` | 0 | `0` | PASS |
| `main.scrollTop = 500` reads back | 500 | `500` (main `overflow-y: auto`) | PASS |
| `window.scrollY` / header top after main scrolled | 0 / 0 | `0` / `0` | PASS |
| `getComputedStyle(html).overflow` | hidden | `hidden` | PASS |
| `getComputedStyle(body).overflow` | hidden | `hidden` | PASS |

Run note: the panel reported `document.visibilityState === "hidden"`, so `requestAnimationFrame` never fired and the
first probe attempt timed out (c11 10s eval limit), leaving the probe div in place. The rerun removed any leftover
probe first and waited with `setTimeout(50)` instead; the values above come from that clean rerun.

## Mapping to the ticket

| Ticket item | Evidence | Result |
|---|---|---|
| Acceptance: "`npm run dev` shows the header on paper ground with the fonts loaded (check computed font-family in devtools)" | Step 1 (dev server via the `dev` script's binary, title/lang/fonts link), Step 2 (paper/ink, header metrics, h1 computed `font-family` Figtree, R2 faces loaded for Figtree and Bodoni Moda), Step 3 screenshot | PASS |
| Validation: "confirm the header renders and the document itself does not scroll; attach a screenshot" | Step 3 screenshot (attached), Step 4 scroll probe | PASS |
| Card note: header on paper ground | body `rgb(246, 245, 241)`, screenshot | PASS |
| Card note: Bodoni Moda / Figtree loaded via computed font-family | h1 and scope computed `font-family` start with `Figtree`; both families have a `loaded` FontFace; `fonts.load()` returns 1 face each; metrics differ from fallbacks | PASS |
| Card note: document does not scroll | Step 4 | PASS |
| Card note: screenshot | `BC-7-header.png` | PASS |
| SPEC section 10 proxy, `VITE_API_PROXY_TARGET` | Step 1 | PASS |
| Ticket text says port 5173 | run-state binding port 15107 used instead (5173 reserved), per the plan's Validation section | info |

Out of scope for this validation (covered by the Gate and the review artifact `art_01M4W2SMAMFTH2Z6WCYBSQM2T8`):
`npm run build`, `npm test`, `make lint`, the dependency list and the emoji scan.

## Cleanup

Vite dev server and dummy server killed by captured PID; `lsof` confirms 15107 and 18007 free. Browser panels
panel:22 and panel:23 (temporary) closed; temporary profile `bc7-validation-cold` deleted; no other panels touched.
Scratch directory deleted. `git status --porcelain` in the worktree empty. This validation did not run a build; the
existing `frontend/dist/` (mtime 22:06:38 EDT, before this session started at 22:11, gitignored by `/frontend/dist/`)
was left untouched as another agent's gate output.
