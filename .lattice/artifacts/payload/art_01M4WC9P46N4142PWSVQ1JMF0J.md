# BC-12 SearchPage: end-to-end validation

- **Verdict: PASS (7/7)**
- HEAD: `bb78854286750e47db7b21028b7848fb232e408f` (branch `feat/BC-12-search-page`, worktree
  `/Users/forrest/Code/brillio-worktrees.noindex/bc-12-search-page`, `git status --porcelain` empty before and after)
- Run: 2026-10-14 00:50–00:55 EDT (04:50–04:55Z)
- Validator: `agent:opus-5.5-validator` (c11 panel:29), browser: c11 WKWebView panel:36 in area:11 (workspace:3),
  viewport 763×430 CSS px

## Servers (native)

```
cd $WT/backend && go build -o bin/api ./cmd/api            # backend/bin/ is gitignored (.gitignore:16)
cd $WT && PORT=18012 DB_PATH=/tmp/brillio-bc12.db CORS_ORIGINS=http://localhost:15112 \
  ./backend/bin/api >>/tmp/brillio-bc12-api.log 2>&1          # PID 14447, later 30624 after the item 6 restart
cd $WT/frontend && VITE_API_PROXY_TARGET=http://localhost:18012 npx vite --port 15112 --strictPort \
  >/tmp/brillio-bc12-web.log 2>&1                             # npm exec PID 14489, vite node PID 14555 (Vite v8.3.4)
```

Checks: `curl localhost:18012/healthz` gave `{"status":"ok"}`, and `curl localhost:15112/api/cities` (through the proxy)
gave `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}`. The API log shows
`seeded 12 listings`.

## Interaction method (tooling note)

c11's WKWebView browser does not support trusted keyboard input: `browser input keyboard` returns
`not_supported: browser.input_keyboard is not supported on WKWebView`. Its `type` and `fill` set `.value` directly and then
dispatch untrusted `input`/`change` events, which React's value tracker ignores. A probe confirmed this: the field kept no
React state and was wiped on blur. So **text entry** used a minimal eval shim that does what a keystroke does at React's
level: `el.focus()`, then the native `HTMLInputElement.prototype.value` setter, then a bubbling `input` event (the same as
RTL's `fireEvent.change`). Everything else used c11's own commands: `focus` (real focus and blur, which drives the
on-blur currency formatting), `select` (City), `click` (Search, Clear, pager, Try again), `press Enter`, `back` (history),
`scroll`, and `goto`. Eval was used only for oracles and the item 6 MutationObserver. This is a limit of the tool, not
the app.

## Item 1: Default load (PASS)

- Action: open `http://localhost:15112/`.
- Observed: URL `/`. Status `12 results · newest first` | `Page 1 of 2`. `main[aria-label=Results]` has `aria-busy="false"`.
  There are 10 cards, in order: `55 Elm Ct No. 1` (Vienna), `22 Birch Ln No. 2`, `456 Oak Ave No. 3`, `789 Pine Rd No. 4`,
  `456 Oak Avenue No. 5`, `123 Main St, Apt 4B No. 6`, `123 Main Street, Unit 4B No. 7`, `789 Pine Rd No. 8`,
  `100 Maple Dr No. 9`, `55 Elm Court No. 10`. Pager: `‹ Previous[disabled] 1[current] 2 Next ›`. City options:
  `Any, Chantilly, Fairfax, Manassas, Reston, Springfield, Vienna`.
- Requests: on the very first load only one search and one cities request reached the API (the StrictMode duplicate was
  aborted before the proxy forwarded it). On a later `goto /` (item 7) both duplicate searches reached the API. Both are
  expected.
- Screenshot: `BC-12-item1-default.png`

## Item 2: Target budget 450000 + City Springfield (PASS)

- Action: enter `450000` in Target budget, then `focus` City (blur). The field read `$450,000`. `select` City
  `Springfield`, `click` Search.
- Observed: URL `/?city=Springfield&targetBudget=450000`. Status `4 results · ranked for $450,000` | `Page 1 of 1`. Cards:
  `123 Main St, Apt 4B No. 1` (seed id **A1**, $450,000, score 0.74), `123 Main Street, Unit 4B No. 2` (B7),
  `456 Oak Ave No. 3` (A2), `456 Oak Avenue No. 4` (B8). Pager: Previous and Next disabled, `1` current. The API search
  count went from 2 to 3 (one request).
- Screenshot: `BC-12-item2-budget-city.png`

## Item 3: Min price 500000 > Max price 400000, no request (PASS)

- Action: `click` Clear. URL `/`, all 7 controls empty, default search ran. API search count **N = 4**. Enter Min `500000`,
  `focus` Max (Min blurs to `$500,000`), enter Max `400000`, `focus` Min bedrooms (Max blurs to `$400,000`).
- Observed: Min `aria-invalid="true"` and Max `aria-invalid="true"`. Both have bottom border `rgb(122, 31, 43)` = `--oxblood: #7A1F2B`
  (an untouched field is `rgb(110, 108, 102)`). Both have `aria-describedby="…-maxPrice-msg"`. Exactly one message,
  `<p id="…-maxPrice-msg">Min price must be less than max price.</p>`, sits inside Max price's field and not Min's.
  Search `disabled === true`.
- No-request proof: `focus` Max, then `press Enter`. URL stays `/`, `aria-busy="false"`, and the count is still **4**.
  Stronger check: `form.requestSubmit()` forced a `submit` event (`submit:true` was recorded). The URL stayed `/`, the
  status was unchanged (`12 results · newest first | Page 1 of 2`), and the count was still **4**. The API log has no
  search between 00:52:22 (Clear) and 00:53:04 (item 4's search).
- Screenshot: `BC-12-item3-min-gt-max.png` (Max shows raw `400000` only because it holds focus for the Enter test.
  Blurred, it reads `$400,000`.)

## Item 4: Keyword `pets` emphasis (PASS)

- Action: `click` Clear. Fields and invalid marks reset. The URL was already `/`, so no request was sent: identical params
  are a no-op by design. Enter `pets` in Keyword, then `press Enter` (this submitted).
- Observed: URL `/?keyword=pets`. Status `7 results · newest first` | `Page 1 of 1`. Cards: 55 Elm Ct, 22 Birch Ln,
  789 Pine Rd, 123 Main Street Unit 4B, 789 Pine Rd, 300 Cedar Blvd, 42 Willow Way (No. 1–7).
  `[...document.querySelectorAll('main em')].map(e=>[e.textContent,getComputedStyle(e).fontWeight])` returned
  `[["pets","600"],["Pets","600"],["pets","600"],["Pets","600"],["pets","600"],["pets","600"],["pets","600"]]`, one `<em>`
  per description. Description body weight is `400`.
- Screenshot: `BC-12-item4-keyword-pets.png`

## Item 5: pageSize=3, page 2, Back (PASS)

- Action: `goto http://localhost:15112/?pageSize=3`.
- Observed: `12 results · newest first` | `Page 1 of 4`, cards No. 1–3 (55 Elm Ct, 22 Birch Ln, 456 Oak Ave). Pager
  `‹ Previous[disabled] 1[current] 2 3 4 Next ›`.
- Action: `click` pager button `Page 2`.
- Observed: URL `/?page=2&pageSize=3`. `Page 2 of 4`. Cards `789 Pine Rd No. 4`, `456 Oak Avenue No. 5`,
  `123 Main St, Apt 4B No. 6`. `document.activeElement.id === 'results-heading'` is `true` (an H2 reading "Results").
- Action: `back` (browser history).
- Observed: URL `/?pageSize=3`. First card `55 Elm Ct No. 1`. `Page 1 of 4`. Pager `1` current. The API log shows the
  page-1 refetch (00:53:43).
- Screenshots: `BC-12-item5-page2.png`, `BC-12-item5-back-page1.png`

## Item 6: API down → error → restart → Try again (PASS)

- Action: `goto http://localhost:15112/` (12 results), `kill 14447` (API log: `server stopped`), `curl healthz` gave `000`,
  no listener on 18012. Installed a MutationObserver that logs each distinct render as
  `aria-busy ; status ; #cards ; pager? ; errorText? ; TryAgain?`. `select` City `Vienna`, `click` Search.
- Observed: URL `/?city=Vienna`. Main text is exactly `Couldn't load results.` / `Try again`. Both status halves are
  `""`. 0 cards, no Pagination nav. Console:
  `[error] Listing search failed {"kind":"http","status":502,...}`. Render log: `false ; 12 results… ; 10 ; true ; false`,
  then `true ; Loading | Page 1 of 2 ; 10 ; true ; false` (previous list kept while loading), then
  `false ;  ; 0 ; false ; true ; true`.
- **Deviation from the plan text (not a defect):** the Vite 8 proxy answered **502** (`http proxy error … ECONNREFUSED`),
  not the 500 the plan predicted. The app handles both as `kind:"http"` and shows the ErrorState.
- Retry while still down (checks resolution 2's guard): `click` Try again. Render log: error view, then
  `true ; Loading ; 0 ; false ; true ; true` (StatusLine reads `Loading` alone, no cards, no pager, error copy kept), then
  back to the error view. The old list never flashed back.
- Restart: same command (new PID 30624), healthz `{"status":"ok"}`. The page stayed in the error view and did not refetch
  on its own. `click` Try again. Render log: `false ;  ; 0 ; false ; true ; true`, then
  `true ; Loading ; 0 ; false ; true ; true`, then `false ; 2 results · newest first | Page 1 of 1 ; 2 ; true ; false ; false`.
  The cards are `55 Elm Ct No. 1` and `55 Elm Court No. 2`, with the pager `1` current. **No intermediate render showed the
  old 10-card list.**
- Screenshots: `BC-12-item6-error.png`, `BC-12-item6-recovered.png`

## Item 7: Only the results region scrolls (PASS)

- Action: `goto http://localhost:15112/`. `main` has scrollHeight 1155 and clientHeight 127, so the tall-panel fallback was
  not needed. The plan's oracle eval (sets `main.scrollTop=10000`) returned `[0,0,true,"hidden","auto",0,true]`, which
  matches the expected `[0, 0, true, 'hidden', 'auto', 0, true]` exactly. The document's scrollHeight equals its
  clientHeight (430), and `html` overflow is `hidden`.
- Wheel check: reset `main.scrollTop=0`, then `scroll --selector main --dy 260` and a page-level `scroll --dy 400`.
  Result: `main.scrollTop 260`, `document.scrollingElement.scrollTop 0`, `scrollY 0`, header top `0`, pager bottom
  `430 === innerHeight`.
- Screenshot (mid-scroll, header + FilterBar + StatusLine + pager pinned, cards scrolled to No. 3):
  `BC-12-item7-results-scroll.png`

## API request trace (`/tmp/brillio-bc12-api.log`, search/cities only)

```
00:50:17 cities, search            initial load (item 1)
00:51:57 cities, search            reload before item 2
00:52:06 search                    item 2 Springfield + budget
00:52:22 search                    item 3 Clear (N=4); none until 00:53:04 -> item 3 sent nothing
00:53:04 search                    item 4 keyword=pets
00:53:29 cities, search            item 5 ?pageSize=3
00:53:36 search                    item 5 page 2
00:53:43 search                    item 5 Back -> page 1 refetch
00:53:58 search, cities            item 6 start at /
(00:54:06 server stopped; 00:54:13 and 00:54:23 Vite proxy ECONNREFUSED -> 502 for ?city=Vienna)
00:54:35 search                    item 6 Try again after restart
00:54:48 search x2, cities         item 7 goto / (StrictMode duplicate)
```

## Teardown

- `kill 30624 14555 14489` (API, vite node, npm exec). `lsof -nP -iTCP:18012 -sTCP:LISTEN` and `…:15112` both printed
  nothing (exit 1). Ports are free.
- Browser panel: confirmed via `c11 tree` that panel:36 was the `[browser]` child of panel:29 at
  `http://localhost:15112/`, then ran `c11 close-panel --panel panel:36`. area:11 keeps panel:15 (placeholder), panel:33,
  and panel:29.
- `git status --porcelain` in the worktree: empty. HEAD is still `bb78854`. The only build output is the gitignored
  `backend/bin/api`.

## Observations for BC-14 (not failures)

- When the list is scrolled, cards clip flush against the StatusLine row with no separator or fade (see the item 7
  screenshot). This is a possible branding polish point for the conformance pass.
