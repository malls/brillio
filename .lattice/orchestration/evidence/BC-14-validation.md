# BC-14 Branding conformance pass: end-to-end validation

- **Verdict: PASS** (V0-V7 pass, C14 pass for shipped files, gate green; teardown V8 confirmed at the end)
- HEAD: `10133f83f5eb6fdb034280a4e80aea413f8e6854` (branch `feat/BC-14-branding-pass`, worktree
  `/Users/forrest/Code/brillio-worktrees.noindex/bc-14-branding-pass`; `git rev-parse HEAD` checked before the run,
  `git status --porcelain` empty before and after)
- Run: 2026-10-14 05:34Z-05:47Z. Validator `agent:opus-5.5-validator` (sub-agent of the BC-14 delegator, c11 panel:37).
- Plan: `.lattice/plans/task_01M4VXK3K5KDHX0P4DPET2S05V.md` section 8 (V0-V8) with Resolutions R3, R4, R6, R8, R9, R10,
  and the Orchestrator ruling of 05:22Z for C14.
- Companion files: `BC-14-keyboard.md` (V6 transcript), eight state screenshots, `BC-14-default-375x667.png` (O14),
  `BC-14-results-ring-1280.png` (D1 ring, O12), `BC-14-c11-panel-763x430.png` (V1 cross-check).
- Drivers (outside the worktree): `<scratchpad>/validator/{lib,oracles,states,parity,keyboard,extra,extra2}.mjs`, Node
  v24.21.0, built-in WebSocket and fetch, no npm installs.

## V0 Servers (PASS)

```
cd $WT/backend && go build -o bin/api ./cmd/api              # go1.27.1; backend/bin/ is gitignored (.gitignore:16)
cd $WT && PORT=18014 DB_PATH=<scratch>/validator/brillio-bc14.db CORS_ORIGINS=http://localhost:15114 \
  ./backend/bin/api >> <scratch>/validator/api.log 2>&1          # first PID 70835
cd $WT/frontend && VITE_API_PROXY_TARGET=http://localhost:18014 npx vite --port 15114 --strictPort   # Vite v8.3.4
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new --remote-debugging-port=19214 \
  --user-data-dir=<scratch>/validator/chrome-bc14 --no-first-run --no-default-browser-check about:blank
```

- `curl -s localhost:18014/healthz` -> `{"status":"ok"}`
- `curl -s localhost:15114/api/cities` (through the vite proxy) -> `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}` (6 cities)
- API log: `seeded 12 listings`, `listening addr=[::]:18014 cors_origins=[http://localhost:15114]`.
- The API was stopped and restarted five times by the drivers for the error states (SIGTERM to the listener on
  18014, then the same command with the same env); every restart answered `/healthz` before the run continued.
  While it was down, vite logged `http proxy error ... ECONNREFUSED` and the browser console logged
  `Listing search failed { kind: "http", status: 502, ... }`; nothing reached the UI but the error state (BRANDING 5.9).

## V1 c11 browser panel (PASS, visual cross-check only)

- `c11 new-panel --type browser --area area:9 --url http://localhost:15114/` -> `OK panel:39 area:9 workspace:3`.
- In panel:39 (WKWebView, 763x430 CSS px): status `12 results · newest first` / `Page 1 of 2`, `main` tabindex `0`,
  the only `[aria-live]` element is `MAIN`, `document.fonts.check` true for Bodoni 400 22px and Figtree 400 15px, the
  document does not scroll. Screenshot: `BC-14-c11-panel-763x430.png` (at 763px the bar is four columns with Search
  and Clear beside Status). WKWebView cannot emulate widths or send trusted keys, so all evidence below comes from
  headless Chrome. (macOS WebKit does not expose `-webkit-tap-highlight-color` in computed style; D6 is read in Chrome.)
- Closed with `c11 close-panel --panel panel:39` -> `OK panel:39 workspace:3` (ref checked != panel:37 first);
  `c11 list-panels` afterwards shows panel:37 still alive and no panel:39.

## V2 Evidence browser (PASS)

Headless Chrome 155.0.8059.40 over CDP on 19214. Viewports via `Emulation.setDeviceMetricsOverride`: 1280x800 DPR 2
(desktop) and 375x812 DPR 2 (`mobile: true`, so the page's `width=device-width` viewport applies).
`Emulation.setFocusEmulationEnabled` on, `Page.captureScreenshot` for PNGs, `Runtime.evaluate` (awaitPromise,
returnByValue) for oracles, `Input.dispatchKeyEvent` / `Input.insertText` / `Input.dispatchMouseEvent` for trusted
input, `Fetch.enable` on `*/api/listings/search*` to hold a search, `Emulation.setEmulatedMedia` for
`prefers-reduced-motion: reduce`. Every shot waited for `document.fonts.ready`, the expected status text and
`aria-busy="false"`.

## V3 Screenshots (PASS)

| File | State | What it shows |
|---|---|---|
| BC-14-default-1280.png | `/` | `12 results · newest first` / `Page 1 of 2`, ten cards, pager `‹ Previous` (ink-3) `1` (current, underlined) `2` `Next ›`. Seven fields on row one, Search + Clear on row two (O1). |
| BC-14-default-375.png | `/` | Two-column bar, Search (flex) + Clear on their own row, price block under the meta line, meta lines wrap only after a dot (D3). |
| BC-14-noresults-1280.png / -375.png | `/?city=Atlantis` | `0 results` / (empty), centred italic `No results.`, `Adjust your filters.`, `Clear filters`; heading 96px below the top of the results region; pager hidden. City shows the extra option `Atlantis`. |
| BC-14-invalid-1280.png / -375.png | `/`, then Min price `500000`, Tab, Max price `400000`, Tab | Min `$500,000` and Max `$400,000` both `aria-invalid="true"` with `2px solid rgb(122, 31, 43)` underlines; one message `Min price must be less than max price.` (oxblood) inside the Max price field; Search disabled (stone fill, ink-3 text, opacity 1); list and status unchanged (10 cards, `12 results · newest first`). Min bedrooms also shows the 2px underline because it holds focus after the second Tab (job 3). |
| BC-14-error-1280.png / -375.png | API stopped, Keyword `pets`, Enter | URL `/?keyword=pets`, status line empty on both halves, centred italic `Couldn't load results.` and `Try again`, pager hidden. Keyword keeps its focus underline. |
| BC-14-default-375x667.png | `/` at 375x667 (O14) | Chrome is 504px, the results region 163px: less than one card (a card is about 175px at 375). |
| BC-14-results-ring-1280.png | `/`, results region focused by a trusted Shift+Tab | The D1 ring: 2px oxblood, inset (offset -2px), flush under the status row and above the pager hairline (O12). |

Invalid state, no request: API search-request count (lines with `"path":"/api/listings/search"` in the API log) was
17 before typing, 17 after the second Tab and still 17 after a further Enter in Min bedrooms at 1280 (12/12/12 at
375); the URL stayed `/`.

## V4 Runtime oracles (PASS at both widths)

Census method (c): for every element and its `::before`/`::after`, report own computed `color` equal to
rgb(122, 31, 43) or rgb(94, 22, 32) when it differs from the parent's, `background-color`, each border side with
width > 0 and style not none, `outline` with style not none, and `text-decoration-color` when a line is drawn.
Class names are shown without the CSS-module hash.

### Width 1280 (viewport 1280x800, DPR 2)

```
(a) default  before: scrollingElement.scrollHeight=800 innerHeight=800 scrollY=0 main.scrollTop=0 (main.scrollHeight=1155, clientHeight=497)
             after main.scrollTop=main.scrollHeight and window.scrollTo(0,99999): scrollHeight=800 scrollY=0 main.scrollTop=658 -> a_pass=True
(a) noresults: scrollHeight=800 innerHeight=800 scrollY=0 -> a_pass=True
(a) error: scrollHeight=800 innerHeight=800 scrollY=0 -> a_pass=True
    horizontal: documentElement.scrollWidth=1280 clientWidth=1280 main.scrollWidth=1280 main.clientWidth=1280
(b) rects [x,y,w,h] before: [["header", [0, 0, 1280, 56]], ["form", [0, 56, 1280, 165]], ["[role=status]", [0, 223, 1280, 32]], ["nav", [0, 752, 1280, 48]]]
    rects after scroll:     [["header", [0, 0, 1280, 56]], ["form", [0, 56, 1280, 165]], ["[role=status]", [0, 223, 1280, 32]], ["nav", [0, 752, 1280, 48]]]  -> unchanged=True
(c) census, idle default: ["button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]
(c) census, invalid (focus on Min bedrooms): ["input.control -> border-bottom 2px", "input.control -> border-bottom 2px", "p.message \"Min price must be less t\" -> color", "input.control -> border-bottom 2px", "button.clear \"Clear\" -> color", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]
(c) census, held search (Fetch paused): ["input.control -> border-bottom 2px", "button.clear \"Clear\" -> color", "div.segment -> background", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]
(c) census, no results: ["button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "button.action \"Clear filters\" -> color"]
(c) census, error (focus on Keyword): ["input.control -> border-bottom 2px", "button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "button.action \"Try again\" -> color"]
(d) non-zero radius: 10 elements, all span.rank (rank pills on screen: 10), e.g. "span.rank \"No. 1\" 999px 999px 999px 999px" -> d_pass=True; no results state radius count=0
(e) box-shadow/background-image/filter/backdrop-filter/text-shadow/opacity<1 hits: []; ::placeholder opacity ["1"]; img/svg/canvas/picture/video/iframe count=0 -> e_pass=True (no results: True)
(f) Bodoni text elements (distinct sizes): ['22px', '26px', '28px']; min=22px; no-results heading ["\"Bodoni Moda\"", "22px", "26px", "400", "italic", true]
    document.fonts.check per used face: {"normal 400 13px \"Figtree\"": true, "normal 400 14px \"Figtree\"": true, "normal 400 15px \"Figtree\"": true, "normal 400 22px \"Bodoni Moda\"": true, "normal 400 26px \"Bodoni Moda\"": true, "normal 500 11px \"Figtree\"": true, "normal 500 13px \"Figtree\"": true, "normal 500 28px \"Bodoni Moda\"": true, "normal 600 13px \"Figtree\"": true, "normal 700 22.5px \"Figtree\"": true}
    no results: {"italic 400 22px \"Bodoni Moda\"": true, "normal 400 13px \"Figtree\"": true, "normal 400 15px \"Figtree\"": true, "normal 500 11px \"Figtree\"": true, "normal 500 13px \"Figtree\"": true, "normal 600 13px \"Figtree\"": true} -> f_pass=True
(g) uppercase (topmost elements whose computed text-transform is uppercase): ["h1.wordmark \"Brillio Listings\"", "span.scope \"Northern Virginia\"", "label.label \"Min price\"", "label.label \"Max price\"", "label.label \"Min bedrooms\"", "label.label \"City\"", "label.label \"Keyword\"", "label.label \"Target budget\"", "label.label \"Status\""]
(h) idle getAnimations(): []; held: ["_slide_pu83o_1 on segment 1200ms linear xInfinity"]; segment 512px of 1280px track, animation '1.2s linear infinite _slide_pu83o_1'; after release: []
    reduced motion (matchMedia reduce=True), held: getAnimations()=[], segment 1280px of 1280px, animation 'none'
(i) controls (labels count:text): ["1:Min price", "1:Max price", "1:Min bedrooms", "1:City", "1:Keyword", "1:Target budget", "1:Status"] -> exactly one each=True
    main aria-live=polite tabindex=0; live regions in DOM: ["main=polite"]
    aria-busy idle=false, held=true, after release=false; progressbar held="progressbar \"Loading\"", idle=null
    aria-current: ["Page 1=page"]; pager ends: ["‹ Previous [disabled]", "Next › [enabled]"]
    held status: ["Loading", "Page 1 of 2"], cards still on screen=10, Search disabled=True look [bg, color, opacity]=["rgb(232, 231, 225)", "rgb(163, 161, 154)", "1"]; paused requests=1
(j) html -webkit-tap-highlight-color: rgba(0, 0, 0, 0)
(k) meta lines: cards=10, wrapped lines=[], items broken inside=[], overflowing=[] -> k_pass=True
layout: header=56 filter bar=165 line=2 status=32 results=497 pager=48; actions box [left,right,width]=[120, 251, 131] search=[120, 202] clear=[220, 251] clear overflows actions=False
```

### Width 375 (viewport 375x812, DPR 2, mobile emulation)

```
(a) default  before: scrollingElement.scrollHeight=812 innerHeight=812 scrollY=0 main.scrollTop=0 (main.scrollHeight=1748, clientHeight=308)
             after main.scrollTop=main.scrollHeight and window.scrollTo(0,99999): scrollHeight=812 scrollY=0 main.scrollTop=1440 -> a_pass=True
(a) noresults: scrollHeight=812 innerHeight=812 scrollY=0 -> a_pass=True
(a) error: scrollHeight=812 innerHeight=812 scrollY=0 -> a_pass=True
    horizontal: documentElement.scrollWidth=375 clientWidth=375 main.scrollWidth=375 main.clientWidth=375
(b) rects [x,y,w,h] before: [["header", [0, 0, 375, 56]], ["form", [0, 56, 375, 366]], ["[role=status]", [0, 424, 375, 32]], ["nav", [0, 764, 375, 48]]]
    rects after scroll:     [["header", [0, 0, 375, 56]], ["form", [0, 56, 375, 366]], ["[role=status]", [0, 424, 375, 32]], ["nav", [0, 764, 375, 48]]]  -> unchanged=True
(c) census, idle default: ["button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]
(c) census, invalid (focus on Min bedrooms): ["input.control -> border-bottom 2px", "input.control -> border-bottom 2px", "p.message \"Min price must be less t\" -> color", "input.control -> border-bottom 2px", "button.clear \"Clear\" -> color", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]
(c) census, held search (Fetch paused): ["input.control -> border-bottom 2px", "button.clear \"Clear\" -> color", "div.segment -> background", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]
(c) census, no results: ["button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "button.action \"Clear filters\" -> color"]
(c) census, error (focus on Keyword): ["input.control -> border-bottom 2px", "button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "button.action \"Try again\" -> color"]
(d) non-zero radius: 10 elements, all span.rank (rank pills on screen: 10), e.g. "span.rank \"No. 1\" 999px 999px 999px 999px" -> d_pass=True; no results state radius count=0
(e) box-shadow/background-image/filter/backdrop-filter/text-shadow/opacity<1 hits: []; ::placeholder opacity ["1"]; img/svg/canvas/picture/video/iframe count=0 -> e_pass=True (no results: True)
(f) Bodoni text elements (distinct sizes): ['22px', '26px', '28px']; min=22px; no-results heading ["\"Bodoni Moda\"", "22px", "26px", "400", "italic", true]
    document.fonts.check per used face: {"normal 400 13px \"Figtree\"": true, "normal 400 14px \"Figtree\"": true, "normal 400 15px \"Figtree\"": true, "normal 400 22px \"Bodoni Moda\"": true, "normal 400 26px \"Bodoni Moda\"": true, "normal 500 11px \"Figtree\"": true, "normal 500 13px \"Figtree\"": true, "normal 500 28px \"Bodoni Moda\"": true, "normal 600 13px \"Figtree\"": true, "normal 700 22.5px \"Figtree\"": true}
    no results: {"italic 400 22px \"Bodoni Moda\"": true, "normal 400 13px \"Figtree\"": true, "normal 400 15px \"Figtree\"": true, "normal 500 11px \"Figtree\"": true, "normal 500 13px \"Figtree\"": true, "normal 600 13px \"Figtree\"": true} -> f_pass=True
(g) uppercase (topmost elements whose computed text-transform is uppercase): ["h1.wordmark \"Brillio Listings\"", "span.scope \"Northern Virginia\"", "label.label \"Min price\"", "label.label \"Max price\"", "label.label \"Min bedrooms\"", "label.label \"City\"", "label.label \"Keyword\"", "label.label \"Target budget\"", "label.label \"Status\""]
(h) idle getAnimations(): []; held: ["_slide_pu83o_1 on segment 1200ms linear xInfinity"]; segment 150px of 375px track, animation '1.2s linear infinite _slide_pu83o_1'; after release: []
    reduced motion (matchMedia reduce=True), held: getAnimations()=[], segment 375px of 375px, animation 'none'
(i) controls (labels count:text): ["1:Min price", "1:Max price", "1:Min bedrooms", "1:City", "1:Keyword", "1:Target budget", "1:Status"] -> exactly one each=True
    main aria-live=polite tabindex=0; live regions in DOM: ["main=polite"]
    aria-busy idle=false, held=true, after release=false; progressbar held="progressbar \"Loading\"", idle=null
    aria-current: ["Page 1=page"]; pager ends: ["‹ Previous [disabled]", "Next › [enabled]"]
    held status: ["Loading", "Page 1 of 2"], cards still on screen=10, Search disabled=True look [bg, color, opacity]=["rgb(232, 231, 225)", "rgb(163, 161, 154)", "1"]; paused requests=1
(j) html -webkit-tap-highlight-color: rgba(0, 0, 0, 0)
(k) meta lines: cards=10, wrapped lines=["55 Elm Ct (2 lines)", "22 Birch Ln (2 lines)", "456 Oak Ave (2 lines)", "789 Pine Rd (2 lines)", "456 Oak Avenue (2 lines)", "123 Main St, Apt 4B (2 lines)", "123 Main Street, Unit 4B (2 lines)", "789 Pine Rd (2 lines)", "100 Maple Dr (2 lines)", "55 Elm Court (2 lines)"], items broken inside=[], overflowing=[] -> k_pass=True
layout: header=56 filter bar=366 line=2 status=32 results=308 pager=48; actions box [left,right,width]=[16, 359, 343] search=[16, 310] clear=[328, 359] clear overflows actions=False
```

### Oxblood: static complement and job mapping

Every CSS rule in the running page (`document.styleSheets`, vite dev) whose declarations name oxblood:

```
.action { color: var(--oxblood) }                                        NoResults: job 2
.action:focus-visible { outline: 2px solid var(--oxblood) }              NoResults: job 3
.control:focus, .control[aria-invalid="true"] { border-bottom: 2px solid var(--oxblood) }   jobs 3 and 4
.message { color: var(--oxblood) }                                       job 4
.search { background: var(--oxblood) }                                   job 1
.search:hover:not(:disabled), .search:active:not(:disabled) { background: var(--oxblood-deep) }   job 1
.search:focus-visible { outline: 2px solid var(--oxblood) }              job 3
.clear { color: var(--oxblood) }                                         job 2
.clear:focus-visible { outline: 2px solid var(--oxblood) }               job 3
.segment { background: var(--oxblood) }                                  loading segment (BRANDING 5.9, 6; O6)
.action { color: var(--oxblood) }                                        ErrorState: job 2
.action:focus-visible { outline: 2px solid var(--oxblood) }              ErrorState: job 3
.button { color: var(--oxblood) }                                        Pager: job 2
.button:focus-visible { outline: 2px solid var(--oxblood) }              Pager: job 3
.button[aria-current="page"] { color: var(--oxblood) }                   job 5
.results:focus-visible { outline: 2px solid var(--oxblood) }             D1 ring on the results region (O12)
```

Runtime census with keyboard focus on the results region (1280, `/`, after a trusted Tab / Shift+Tab):
`["button.search \"Search\" -> background", "button.clear \"Clear\" -> color", "main.results[Results] -> outline 2px solid offset -2px", "button.button[Page 1] \"1\" -> color, text-decoration underline 1px", "button.button[Page 2] \"2\" -> color", "button.button.next -> color"]`.

Nothing else is oxblood in any state: no price, address, rank, score, monogram, status label or hairline.
Result for (c): only the five jobs, plus the segment while loading (O6) and the D1 ring while the results region
has keyboard focus (O12).


Note on (f): the only Figtree 700 text is the visually hidden `Results` heading (`#results-heading`, UA h2 bold
22.5px, 1x1px box, `clip-path: inset(50%)`, absolute). It never shows, which matches A7 ("the default bold is
overridden everywhere it shows").

## V5 Text-button parity (PASS)

1280x800. Idle values from `getComputedStyle`; hover after a trusted `mouseMoved` onto the button (`:hover` true);
focus ring after reaching the button with trusted Tab keys (`:focus-visible` true). States: `/` for Clear, Next and
Page 2; `/?page=2` for Previous (enabled); `/?city=Atlantis` for Clear filters; API stopped for Try again.
`font-variant-numeric` deliberately not compared (R8: page numbers inherit tabular-nums from `.pager`). The current
page is listed for reference only and not compared.

| Button (state) | font-size | line-height | weight | letter-spacing | color | padding | text-decoration idle | hover | cursor | keyboard focus ring |
|---|---|---|---|---|---|---|---|---|---|---|
| Clear | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | none | underline, offset 3px, rgb(122, 31, 43) | pointer | 2px solid rgb(122, 31, 43), offset 3px |
| Next | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | none | underline, offset 3px, rgb(122, 31, 43) | pointer | 2px solid rgb(122, 31, 43), offset 3px |
| Page 2 (non-current) | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | none | underline, offset 3px, rgb(122, 31, 43) | pointer | 2px solid rgb(122, 31, 43), offset 3px |
| Page 1 (current, reference only) | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | underline 1px, offset 3px | n/a | default | n/a |
| Previous | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | none | underline, offset 3px, rgb(122, 31, 43) | pointer | 2px solid rgb(122, 31, 43), offset 3px |
| Clear filters | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | none | underline, offset 3px, rgb(122, 31, 43) | pointer | 2px solid rgb(122, 31, 43), offset 3px |
| Try again | 13px | 13px | 500 | normal | rgb(122, 31, 43) | 0px 0px 0px 0px | none | underline, offset 3px, rgb(122, 31, 43) | pointer | 2px solid rgb(122, 31, 43), offset 3px |

Also identical across the six compared buttons: font-family `Figtree, system-ui, -apple-system, "Segoe UI", sans-serif`, font-style normal, text-transform none, background rgba(0, 0, 0, 0), border 0px, white-space nowrap, radius 0px. Differences found by the script: none.

### Mouse click inside `<main>` shows no ring (PASS)

1280, `/`: a trusted left click (`Input.dispatchMouseEvent` pressed/released) on the first card's description.
Result: `{"activeElement":"MAIN[tabindex=0]","focusVisible":false,"outlineStyle":"none"}`, so the click focuses the
region (tabindex 0) without `:focus-visible` and without the ring. A trusted Tab then Shift+Tab back to the region:
`{"activeElement":"MAIN","focusVisible":true,"outline":"2px solid rgb(122, 31, 43)","outlineOffset":"-2px"}`.

## V6 Keyboard-only pass (PASS)

Full transcript (59 key lines plus the R4 line): `BC-14-keyboard.md`. Summary, 1280x800, trusted keys only:

1. From `/`, Tab x13: Min price > Max price > Min bedrooms > City > Keyword > Target budget > Status > Search >
   Clear > results region > Page 1 > Page 2 > Next. Previous (disabled) skipped. Fields show the 2px oxblood
   underline, buttons a 2px oxblood ring (Search offset 2px, text buttons 3px), the region the inset ring.
   R4 at the region stop: `tagName=MAIN`, `getAttribute('tabindex') === '0'` true, `outline-color=rgb(122, 31, 43)`,
   `:focus-visible` true. WebKit path: SearchPage jsdom test T13 (passes at this HEAD, see the keyboard file).
2. Shift+Tab x7 to Target budget, type `450000`, Enter: `/?targetBudget=450000`, `12 results · ranked for $450,000`.
3. Tab x4 to the region, PageDown: `main.scrollTop` 0 -> 457 (smooth), ArrowDown -> 497, `scrollY` 0 throughout.
4. Tab to Page 2, Enter: `/?targetBudget=450000&page=2`, `Page 2 of 2`, focus on `#results-heading`, first rank
   `No. 11`. Tab: Previous > Page 1 > Page 2 > (Next skipped, focus leaves the page). Shift+Tab back to Previous,
   Enter: `/?targetBudget=450000`, `Page 1 of 2`, first rank `No. 1`.
5. (R3) Shift+Tab to Clear, Enter: `/`, `12 results · newest first`. API stopped. Shift+Tab to Keyword, type `pets`,
   Enter: `/?keyword=pets`, error state, status empty. API restarted. Tab x6 to Try again (Target budget, Status,
   Search, Clear, region, Try again), Enter: `7 results · newest first`.

## O14 measurements (measure only; owner question)

| Viewport | header | filter bar | line | status | results region | pager | pager on screen | document scrollHeight |
|---|---|---|---|---|---|---|---|---|
| 1280x800 | 56 | 165 | 2 | 32 | 497 | 48 | yes (752-800) | 800 |
| 375x812 | 56 | 366 | 2 | 32 | 308 | 48 | yes (764-812) | 812 |
| 375x667 | 56 | 366 | 2 | 32 | 163 | 48 | yes (619-667) | 667 |
| 667x375 | 56 | 366 | 2 | 32 | 0 | 48 | no (456-504) | 375 |

- 375x812 and 375x667: the pinned chrome is 504px, as the plan estimated. At 375x667 the results region (163px) is
  smaller than one card (about 175px).
- 667x375 (phone landscape): the results region is 0px tall and the pager sits below the fold. By touch or mouse it
  cannot be reached (html and body are `overflow: hidden`). By keyboard it can: Tab to Search, Clear, the region and
  the pager scrolls `<body>` (an `overflow: hidden` box is still scrolled by focus) by 29px and then 129px, pushing
  the header off the top (`header.top=-129`, nav 327-375) while `window.scrollY` and `html.scrollTop` stay 0. So at
  this size the "document never scrolls" rule is broken by focus, in a way oracle (a) cannot see. Tab back to Min
  price restores body.scrollTop 0. Feeds O14.
- >760 (A13): at 1280 the actions column is 131px (x 120-251); Search 120-202 and Clear 220-251, so Clear's right
  edge lands exactly on the column edge: no overflow.

## C14 Emoji check (PASS for shipped files)

Scope per R6 and the Orchestrator ruling (05:22Z): `git ls-files` minus `.lattice/`, the two PDFs, and SPEC section
12's other exclusions (`*.db`, `node_modules/`; `.git/` is never listed).

```
cd $WT && git ls-files -z | grep -zv '^\.lattice/' | grep -zv '\.pdf$' | grep -zv '\.db$' | grep -zv 'node_modules/' \
  | xargs -0 perl -CSD -ne 'print "$ARGV:$.\n" if /[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]/; close ARGV if eof'
```

- Result: no output, perl exit 0. 132 files in scope (of 434 tracked). A per-file loop over the same 132 files
  (one perl per file) also gave 0 hits and 0 non-zero exits, so no file aborted the scan.
- `.lattice/` reported separately (300 tracked files): 724 x U+2713 (the vitest pass mark) in 9 files
  (BC-8-vitest-verbose.txt and its artifact payload, BC-9-validation.md and its payload, BC-10-vitest.txt and its
  payload, one BC-9 review payload, BC-8-validation.md and its payload), plus 44 PNGs that perl -CSD cannot decode
  as text (binary). These are immutable records, not shipped (SPEC section 12). This evidence summarises vitest
  output and contains no characters from the checked ranges (checked with the same perl expression).

## Gate (plan section 9) at this HEAD

- `cd $WT && make check` -> exit 0: golangci-lint `0 issues.`, `eslint . --max-warnings 0` clean, typecheck clean,
  prettier `All matched files use Prettier code style!`, `go test -race ./...` ok for config, domain, httpapi, store,
  store/seed (cached), vitest 23 test files passed, 398 tests passed.
- `cd $WT/frontend && npm run build` -> exit 0 (dist/index.html 0.77 kB, CSS 12.19 kB, JS 241.36 kB; `frontend/dist/`
  is gitignored, .gitignore:25).
- Targeted: `npx vitest run src/pages/SearchPage.test.tsx -t "puts the results region in the Tab order"` -> 1 passed,
  14 skipped (T13).

## Verdict per item

| Item | Result | Evidence |
|---|---|---|
| V0 Servers | PASS | healthz `{"status":"ok"}`, 6 cities through the proxy |
| V1 c11 panel | PASS (cross-check) | panel:39 opened, inspected, closed with `--panel panel:39` |
| V2 Evidence browser | PASS | CDP driver ran every step; no fallback needed |
| V3 Screenshots | PASS | 8 required + 375x667 + results ring + c11 panel |
| V4 (a) document never scrolls | PASS | 1280 and 375: scrollHeight == innerHeight, scrollY 0 after scrolling results to the bottom |
| V4 (b) pinned rows | PASS | header/form/status/nav rects identical before and after |
| V4 (c) oxblood census | PASS | five jobs only, plus segment while loading (O6) and D1 ring on keyboard focus (O12) |
| V4 (d) radius | PASS | only the 10 rank pills (999px) |
| V4 (e) no shadow/gradient/filter/opacity/images | PASS | 0 hits, placeholder opacity 1, 0 media elements |
| V4 (f) Bodoni floor, fonts loaded | PASS | min 22px; fonts.check true for every used face |
| V4 (g) uppercase | PASS | wordmark, scope, seven field labels; plus the card status label on page 2 (`PENDING`, ink-2 11px/500 +0.88px, the only non-active seed listing; same on `/?status=pending`) |
| V4 (h) animations | PASS | idle []; held only `slide` 1.2s linear on the segment (40% width); reduced motion [] and full width |
| V4 (i) a11y row | PASS | 7 controls x 1 label, main polite and the only live region (D2), busy true/false, progressbar "Loading", aria-current page, disabled ends |
| V4 (j) tap highlight | PASS | `rgba(0, 0, 0, 0)` at both widths (D6) |
| V4 (k) meta items whole | PASS | 375: all 10 meta lines wrap to 2 lines, 0 items broken, 0 overflow (D3) |
| V5 text-button parity | PASS | 6 buttons identical idle, hover and focus ring |
| Mouse click in main | PASS | focus on MAIN, no :focus-visible, outline none |
| V6 keyboard pass | PASS | 10/10 checks, transcript in BC-14-keyboard.md |
| C14 emoji (shipped) | PASS | 0 hits in 132 shipped files |
| Gate | PASS | make check exit 0; npm run build exit 0 |
| V7 evidence | done | this file, BC-14-keyboard.md, screenshots attached with --role validation; checklist comment posted |
| V8 teardown | done | see below |

**Overall: PASS.** No implementation-level or plan-level failure.

## Findings beyond the plan's O1-O14

1. **Focus falls to `<body>` after Try again and after Clear filters** (keyboard row 59; extra probe for Clear
   filters). Both buttons unmount with their state when the new list arrives. Keyboard use still works (the next Tab
   starts at Min price), and SPEC section 7.4 names focus handling only for page changes. Not a BRANDING item;
   candidate follow-up: move focus to `#results-heading` after a retry or a clear, as after a page change.
2. **O14 refinement**: at 667x375, focus scrolls the `overflow: hidden` body by up to 129px, so the header leaves the
   screen and the pager comes into view for keyboard users only; touch users cannot reach the pager and the results
   region is 0px tall. At 375x667 the results region is 163px, less than one card.
3. **O11 confirmed at runtime**: a hovered non-current page number (oxblood, 500, underline thickness `auto` at 3px
   offset) is near-identical to the current page (oxblood, 500, 1px underline at 3px offset); the cursor
   (pointer vs default) is the only other difference.
4. Wording only (owner may ignore): BRANDING 4.4 says "Search full width" at <=760; the build, like the demo
   (demo:98-100), gives Search `flex: 1` beside Clear (x 16-310 and 328-359 at 375).

## V8 Teardown (done, 05:47Z)

- Killed the API (PID 97782, the last restart, on 18014), vite (node PID 70969 and its `npm exec` parent 70945, on
  15114) and headless Chrome (PID 71054, on 19214). `lsof -nP -iTCP:18014 -iTCP:15114 -iTCP:19214 -sTCP:LISTEN`
  afterwards: no output (exit 1); `pgrep` finds no `chrome-bc14` or `backend/bin/api` process.
- `rm -f <scratch>/validator/brillio-bc14.db*` (the DB lived in the session scratchpad per R9, never in /tmp).
- c11 browser panel closed earlier with `c11 close-panel --panel panel:39`.
- Worktree: `git status --porcelain` empty, HEAD still `10133f83f5eb6fdb034280a4e80aea413f8e6854`. Build outputs
  `backend/bin/api` and `frontend/dist/` are gitignored.
