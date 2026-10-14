# Validation Report

Source spec: [docs/SPEC.md](../../docs/SPEC.md) (as amended in-branch during the run)
Design contract: [BRANDING.md](../../BRANDING.md)
Source build plan: none. No BUILDPLAN.md exists; the ticket descriptions plus [.lattice/context.md](../context.md) are the plan (run-state decision 00:47Z)
Source validation plan: [validation-plan.md](./validation-plan.md) (56 rows)
Result Validator: agent:opus-5.5-result-validator (c11 panel:40), in a fresh session with five in-process audit sub-agents (one per bucket)
Date: 2026-10-14
Run completed: 2026-10-14T07:11Z (run-state RUN COMPLETE). Audited head: `master` 61228c7 ("chore(lattice): sync board after BC-17")
Audit checkout: isolated detached worktree `/Users/forrest/Code/brillio-worktrees.noindex/rv-audit` at 61228c7. Nothing was edited or committed in either checkout, and no ticket was mutated.

## Summary

- Total criteria audited: **52**. These are every `pre-merge-static` and `pre-merge-runtime` row. The 4 `post-merge-smoke` rows are listed for the operator at the end.
- Pass: **52**
- Partial: **0**
- Fail: **0**
- Blocked: **0**

Three browser rows (35, 44, 45) passed with an adapted method because of c11 browser-panel limits: viewport width, real key presses and real wheel input. Their pass conditions were all observed at runtime. The gap is described under "What I couldn't verify".

**Overall verdict: GREEN.** The assembled master meets every audited criterion. Specifically:

- `make check` is green.
- The backend runtime contract holds byte-for-byte against SPEC §5 and §6, with scores recomputed independently at asOf 2026-10-14.
- The UI renders every §7.5 state in a real browser.
- Compose and `make smoke` pass.
- The zip rebuilt from 61228c7 is **byte-identical** to the shipped `dist/brillio-listings-20261014.zip` (329238 B, 135 files, same SHA-256), and `verify-dist` passes all 13 checks.

Residual risk is concentrated in design conformance: fifteen open owner questions (O1–O15) record places where the build and BRANDING.md's text disagree, and the operator's felt pass (row 55) is where those land.

## Per-criterion results

Evidence logs are in the validator's session scratchpad (`.../scratchpad/{a,b,c,d1,d2}`, plus `make-check.txt`).

| # | Criterion (ID) | Result | Notes |
|---|---|---|---|
| 1 | §3 layout | Pass | `git ls-files` lists 523 files. All 76 required paths are present, covering the row's list plus the fuller §3 tree (scripts, docs/branding-demo.html, CLAUDE.md, agents.md, each named src file, and the 3 files in each of the 11 component folders). 0 missing. The extras are listed under Drift. |
| 2 | §2 stack + rules | Pass | go.mod: `go 1.27`, and the only direct require is `modernc.org/sqlite v1.60.1`. Imports are stdlib plus sqlite only, with no router, testify or ORM. package.json has 18 packages and matches the BC-7 plan's final dependency table exactly. Versions: vite 8.3.4, react 19.3.0, typescript 6.0.3, vitest 5.0.4, eslint 9.39.5. |
| 3 | §2 git conventions | Pass | There are 16 first-parent merges, one per ticket BC-2..BC-17. All are 2-parent (`--no-ff`) and the second parent is exactly the `feat/BC-N-*` tip. The root is `f4f8a0b chore: bootstrap repository` (BC-1). All 75 non-merge commits are conventional. The merge subjects are not; see Drift. |
| 4 | BRANDING §8 no emoji | Pass | Shipped files (git ls-files minus `.lattice/` and both PDFs, 135 files): 0 matches for U+1F300–1FAFF / U+2600–27BF. The scan used a perl equivalent of `grep -P`, which macOS grep lacks; it was self-tested on planted characters. Commit messages also have 0 hits. |
| 5 | §8 make check | Pass | Run by the validator on the clean 61228c7 checkout: exit 0 in 15 s. Steps: golangci-lint 0 issues; eslint `--max-warnings 0`; `tsc -b --noEmit`; prettier `--check` clean; `go test -race ./...` uncached, 5 packages ok; vitest 23 files, 398 tests passed. |
| 6 | §8 backend test inventory | Pass | `go test -race -v -count=1 ./...`: 481 tests and subtests PASS, 0 FAIL. Every §8 domain, store and httpapi item maps to a named test, for example `TestRank/tied_scores_tiebreak_order`, `TestPaginate/{page_1,…,rank_continuity_across_pages}`, `TestBudgetFit/asymmetry_at_matching_percentages`, `TestSearchLiteralWildcards`, `TestSeedTwiceIdempotent`, `TestCORS/allowed_preflight`, `TestSearchPaging/page=99`. The §5.3 validation-rule tests live in httpapi (Drift). |
| 7 | §8 frontend test inventory | Pass | `npm test -- --reporter=verbose`: 23 files, 398 tests, 0 failed. Coverage: every §7.4 component has a test file; every BRANDING §7.1 row has a formatter case, including `1 result` and the not-current-year date; useSearchParamsState covers the round-trip of all 9 params, popstate, page reset and pageSize kept. SearchPage T1 (success), T2 (empty), T3 (min_gt_max on both fields), T5 (network error and Try again) and T6 (abort, with `No. 11`, `No. 12`). |
| 8 | §4.1 schema | Pass | `0001_listings.sql`: `PRIMARY KEY (source, id)` :25, `listed_date TEXT` :22, status CHECK :23, the 4 indexes :28-31, and `schema_migrations` :4-7. `migrate.go` embeds `migrations/*.sql` (:16) and runs each file and its record in one transaction (:98-111). |
| 9 | §4.2 seed idempotent | Pass | Fresh `/tmp/brillio-rv.db`: the log reads `seeded 12 listings` on both starts. sqlite3 `count(*)` is 12 after start 1 and after start 2, and `schema_migrations` has 1 row. |
| 10 | §4.2 invalid seed fails | Pass | Index 3 set to `"status":"gone"`: exit 1, log `record index 3: field "status": invalid listing status: "gone"`. The index is 0-based and labelled as an index, so it is unambiguous. |
| 11 | §5.1 healthz | Pass | 200, `{"status":"ok"}`, `Content-Type: application/json; charset=utf-8`, and `X-Request-Id` present (Go's canonical spelling). |
| 12 | §5.1 degraded | Pass | `health.go:25-32` returns 503 `{"status":"degraded","error":"database unavailable"}`. Tests: `health_test.go:28-31` uses a failing Pinger, asserting the exact body and that the raw error is logged but not echoed; `health_store_test.go:26-33` uses a closed real store. |
| 13 | §5.2 cities | Pass | The body is exactly `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}`. |
| 14 | §5.3/§5.5 field validation | Pass | All 12 single-fault requests return 400 `validation_error` with the expected code: negative, not_a_number, not_an_integer, not_positive, invalid_enum, out_of_range ×3, too_long ×2, a single min_gt_max on maxPrice, and unknown_param. The BRANDING §7.2 strings are exact where they exist. The combined `minPrice=-1&pageSize=0&foo=1` returns all 3 errors in §5.3 order, independent of input order. |
| 15 | §5.3 empty value absent | Pass | `?minPrice=` gives 200, `totalItems` 12 (10 on page 1) and `appliedFilters.minPrice` null. |
| 16 | §5.3 page past end | Pass | `?page=99` gives 200, `items: []`, `{"page":99,"pageSize":10,"totalItems":12,"totalPages":2}`. |
| 17 | §5.3 empty city | Pass | `?city=Atlantis` gives 200, `totalItems` 0, `totalPages` 0. |
| 18 | §5.3 status param | Pass | `status=pending` returns only A7. The comma form and the repeated form are byte-identical, both as written and at `pageSize=100`: 12 items, 11 active and 1 pending. |
| 19 | §5.4 ranked response | Pass | A1, B7, A2, B8 at ranks 1–4. `scoring` is 450000/0.6/0.4/30 with asOf `2026-10-14`. Recomputed from §6.2 at that asOf: A1 is age 46, recency 0.3455, score **0.7382**; B7 0.7266; A2 0.5516; B8 0.5348. All match the API to four decimals. `appliedFilters` echoes city with nulls for the rest. |
| 20 | §5.4 no budget | Pass | `scoring.targetBudget` null, weights 0/1, all 12 `budgetFit` null. Order is strictly newest first: A5, A4, A2, A3, B8, A1, B7, B9, B10, B11, A6, A7, with score = recency on every row. |
| 21 | §5.5 404/405 | Pass | 404 `{"error":{"code":"not_found",…}}` and 405 `{"error":{"code":"method_not_allowed",…}}` with `Allow: GET, HEAD`. Both are JSON and carry a request ID. |
| 22 | §5.5 500 hygiene | Pass | `middleware.go:83-105` recovers the panic, logs it with `request_id`, the panic and the stack, and writes a generic `internal_error`. `middleware_test.go:86-113` asserts the exact body, checks for absence of `goroutine`/`.go:`/the panic text, and checks the request ID in the log. |
| 23 | §5.6 CORS | Pass | Allowed origin: 204 with the exact allow-origin, methods `GET, OPTIONS`, headers `Content-Type, X-Request-ID` and expose `X-Request-ID`. `http://evil.test`: 204 with no `Access-Control-*` headers. |
| 24 | §5 request id | Pass | `rv-123` is echoed (and logged). With no header sent, a 32-character lowercase hex ID is generated (two samples). |
| 25 | §5.7 OpenAPI | Pass | 200 `application/yaml`, and the body is byte-identical to `backend/openapi.yaml`. Redocly 2.63.0 reported 0 errors (BC-6 artifact art_01M4WBDAG3WZDYG9TAK2SNFFDD), and the file is unchanged since it was linted. `TestOpenAPISearchParamsMatchHandler` and five more `TestOpenAPI*` tests pass with `-count=1`. The example score is 0.7414 (openapi.yaml:261). |
| 26 | §6 pipeline readability | Pass | `httpapi/search.go:22-54` `searchListings` runs parse, `Searcher.Search`, the dedupe-slot comment, `domain.Rank`, `domain.Paginate` and `writeJSON`, one call and early return per stage. |
| 27 | §6.1 SQL safety | Pass | No `fmt.Sprintf` SQL. The only concatenations join constant fragments. The keyword clause is `instr(lower(description), ?) > 0` with a lowercased token, and status `IN` uses placeholders only (`listings.go:146-151`). |
| 28 | §6.1 literal `%` | Pass | `?keyword=%25` and `?keyword=_` each return 200 with 0 items. |
| 29 | §6.2 scoring | Pass | Exported `BudgetWeight` 0.6, `RecencyWeight` 0.4, `RecencyHalfLifeDays` 30 (`scoring.go:25-35`). Budget fit is asymmetric and recency is `0.5^(age/30)` over whole UTC days with future dates clamped. The formulas are in doc comments. There is no `time.Now` in domain, and its imports are stdlib only. |
| 30 | §6.3 ordering | Pass | `rank.go:74-86` uses `cmp.Or` on unrounded score desc, listedDate desc, price asc, source asc, id asc, with a stable sort. `tied_scores_tiebreak_order` asserts the full order; one fixture pair is decided by each key. There is also an all-720-permutations test. |
| 31 | §6.4 pagination | Pass | `?pageSize=3&page=2` gives ranks 4 (A3), 5 (B8), 6 (A1) and `totalPages` 4. |
| 32 | §11 config + logging | Pass | `LOG_LEVEL=loud` exits 1 with `LOG_LEVEL: "loud" is not one of debug\|info\|warn\|error`. One request produces exactly one JSON line with method, path, status, duration_ms, request_id and remote_addr. SIGTERM logs `shutting down` and `server stopped`, then exits 0. |
| 33 | §7.3 types/client | Pass | `types.ts` matches §7.3 field for field: names, optionality, nullability and unions. The only extra export is `LISTING_STATUSES`. `client.ts` omits undefined values (:17-29), repeats status (:31-39), maps non-2xx to ApiError via `parseErrorResponse` (:91-93) and maps fetch failure or abort to `network` (:84-88). |
| 34 | §7.2 abort/previous | Pass | One AbortController per request (`useListingSearch.ts:63`) and abort on cleanup (:89-91). Aborted-request guards are on both paths (:66-68, :75-77). `previous` is kept in loading, invalid and error (:36-49). SearchPage T6 shows the old signal aborted and the new one live; T8 shows a late aborted page-2 body never rendering. |
| 35 | §7.4 FilterBar fields | Pass (adapted) | The panel was **763px** wide, not 1280; the c11 WKWebView cannot emulate a viewport. 763 is above 760, so the same CSS applies. The DOM shows 7 fields in the specified order, each with exactly 1 `<label>`. Labels and placeholders are verbatim from BRANDING §7.2 and SPEC §7.4. City offers Any plus the 6 cities; Status offers Any/Active/Pending/Sold. A simulated 1280 layout puts 7 fields in row 1 and the actions in row 2. |
| 36 | §7.4 client validation | Pass | Both price fields get `aria-invalid` and a 2px oxblood underline. The exact string sits under Max price and both fields describe-by it. Search is disabled. The API log's search count is 2 before and 2 after, so no request was issued. |
| 38 | §7.5 default load | Pass | `12 results · newest first`, `Page 1 of 2`. Cards No. 1–10 are dated Sep 4 → Aug 15 and page 2 has Aug 10 and Jul 28, which is strict newest first. |
| 39 | §7.5 ranked search | Pass | URL `?city=Springfield&targetBudget=450000`, `4 results · ranked for $450,000`. The first card is `123 Main St, Apt 4B` · MLS A · `$450,000` · `No. 1` · 0.74. |
| 40 | §7.4 card + emphasis | Pass | `?keyword=pets` gives 7 results. Each match is wrapped at computed weight 600 in the same `rgb(22,22,22)` color as the description. The meta line on A7 reads `PENDING · Chantilly, VA 20151 · 4 bd · 2.5 ba · 1,950 sqft · Listed Jul 28 · MLS A`, with PENDING in the Label style. |
| 41 | §7.1 URL + pager | Pass | Clicking page 2 gives URL `?page=2&pageSize=3` with first card `No. 4` and `document.activeElement` = `H2#results-heading`. Back returns `?pageSize=3`, page 1, and `aria-current` on 1. |
| 42 | §7.5 error + recovery | Pass | Search with unchanged filters correctly sends no request (§7.2), so Min bedrooms was set to 1. With the API down the page shows `Couldn't load results.` and `Try again`, the status line is empty and the pager hidden. After restarting the API, Try again gave `12 results · newest first`. |
| 43 | §7.5 empty state | Pass | `?city=Atlantis` shows `No results.` / `Adjust your filters.` / `Clear filters`, the status `0 results`, and no pagination nav. Clear filters returns to `/` with 12 results. |
| 44 | §7.6 a11y + keyboard | Pass (adapted) | All the pass-condition attributes were observed at runtime: every input is labelled; `aria-live="polite"`; `aria-busy="true"` during loading (caught by a MutationObserver), then false; `aria-current="page"` on the current page only; disabled Previous/Next carry `disabled` and are absent from the tab order; `role="progressbar"` with `aria-label="Loading"`. **Method gap:** c11 browser key events are synthetic, so a Tab key press does not move focus. The keyboard-only pass was emulated instead: the computed tab order was walked with `focus()` (fields, Search, Clear, results region, pager, Try again), Enter in a field and in the Status select submitted, and the focused buttons were activated. |
| 45 | BRANDING §4.2 shell | Pass (adapted) | `?pageSize=100`: html and body are `overflow:hidden`, and the results region is `overflow-y:auto; overscroll-behavior:contain` (scrollHeight 1385 against clientHeight 127). Document scroll attempts (scrollBy, scrollTop, scrollTo, c11 scroll) left `scrollY` at 0. The results region scrolled to its bottom, and the header, bar, status and pager rects never moved. No real wheel or trackpad input was possible. |
| 46 | BRANDING conformance | Pass | The checklist comment ev_01M4WFA51XH4EQCJCF92K392DK covers §0–§8 and §10 as pass or fixed (fixes D1–D7), with §9 n/a and a reason. Artifact art_01M4WF7QGZ00YPVXSFZ5E4N5FT has 8 PNGs, all opened: default, no-results, invalid and error at 1280 and at 375. No frontend commit landed after the evidence head 10133f8. |
| 47 | BRANDING §6 motion | Pass | `LoadingLine.module.css:30` holds the only `animation` in `frontend/src`. It is static and full-width under `prefers-reduced-motion: reduce` (:44-48), and the test asserts both rules. There is no `transition` anywhere. |
| 48 | §10 compose up | Pass | Ran `docker compose up -d` with the operator's three Docker env vars unset. api was healthy, and web came up after `npm ci`. `curl :5173/api/cities` returned the exact six-city body, proxied (the api log shows the web container's address with the same request ID). `make down` left no containers, volumes or networks, and both ports were free. |
| 49 | §10 make smoke | Pass | Exit 0. Eight assertions printed PASS: api health, cities, search non-empty, rank 1 first, 400 min_gt_max, web shell, web proxies /healthz and /api/cities. Then teardown `down -v`; nothing was left behind. |
| 51 | README accuracy | Pass | All required sections are present (Run :5-48 through Testing :227-249). Every number and formula matches `internal/domain`, config and compose: 0.6/0.4/30, the budget-fit formula, page size 10/100, tie order, env defaults. The worked example was recomputed (0.7414 at 2026-10-13, 0.7382 at 2026-10-14). No emoji and no `!`. |
| 52 | EXTENSIONS.md | Pass | 221 lines, with sections 1–8 matching SPEC §9 items 1–8. All 162 backticked tokens resolve; the only ones absent from the tree are explicit proposals or example values. The rehearsal covers dedupe (:138-159), geo radius (:161-180) and sort (:182-201), and its numbers were recomputed and hold. |
| 53 | §12 dist | Pass | `make dist` on 61228c7 produced 329238 B and 135 files. `unzip -l` shows no `.git/`, `node_modules`, `*.db`, `.lattice`, PDFs or `dist/` entries, and README.md, CLAUDE.md and agents.md are present (operator ruling 05:40Z). `verify-dist.sh` printed `PASS (13 checks)` in about 65 s, including its own compose up, smoke and `make check` from the extracted tree. **It is byte-identical to the shipped `/Users/forrest/Code/brillio/dist/brillio-listings-20261014.zip`, with the same SHA-256.** An emoji scan of the extracted zip found 0 hits. |

Rows 50, 54, 55 and 56 are `post-merge-smoke` rows, staged below for the operator and not attempted here.

## Drift from the build plan

There is no BUILDPLAN.md. Drift is measured against the ticket descriptions, `.lattice/context.md`, and SPEC as it stood at intake. Every contract amendment made during the run is reflected in SPEC's text, which is what this audit judges against. None of the items below is a defect.

- **Contract amendments made in-branch (all in SPEC now).** TypeScript 6.0.3 rather than 7 (operator ruling); ESLint 9.39.5 rather than 10 (jsx-a11y peer); the `malformed_query` field code (BC-5); the §5.4 example corrected from 0.8731 to 0.7414; the expanded §5.3 parsing rules; the §7.6 focusable results region (BC-14 D1); and the §13 note 5 copy question.
- **Extra files beyond the SPEC §3 tree.** These are `internal/store/seed/` (a subpackage), `backend/openapi.go` (an embed shim), `backend/.dockerignore`, `frontend/src/format/keyword.ts`, `components/FilterBar/filterBarModel.ts`, `tsconfig.app.json`, `tsconfig.node.json` and `.prettierignore`. All are reasonable decompositions; the README documents the backend ones. §3 could list `keyword.ts` and the seed package.
- **Frontend dependency allowlist.** BC-7's description listed 12 packages and said "no others". It shipped 18: react-dom, @vitejs/plugin-react, @types/react, @types/react-dom, @types/node and eslint. Each is justified in the BC-7 plan's final table, which plan review accepted. package.json matches that table exactly.
- **Error transport in the client.** The client rejects with an `ApiRequestError` (an `Error` subclass carrying `.apiError` and `.requestId`) rather than a bare `ApiError` object, because typescript-eslint forbids rejecting plain objects. SPEC §7.3 is satisfied but silent on this wrapper. A 2xx response with a malformed body maps to `{kind:'http'}`, which SPEC does not mention.
- **Dedupe key in EXTENSIONS.md.** It uses the normalized address plus coordinates rounded to 4 decimals instead of SPEC §9.1's "address + zip" candidate, because A2 and B8 have different zips. SPEC calls these "key candidates", and the doc explains the choice. This is good interview material.
- **Budget-fit float form.** The code computes `price/target`, while §6.2 writes `1 - pctUnder` and notes "equals price/target". These differ at the last bit in about 37% of sampled pairs; rounding to four decimals hides it. SPEC declares them equal, so this is accepted.
- **Merge commit subjects.** The 16 ticket merges use `Merge branch 'feat/BC-N-…' (BC-N)`, and 36 back-merges or anchor merges use git's default subject. SPEC §2 says "Conventional commit messages" without carving out merges. Conventional Commits tooling ignores merges by default; every non-merge commit conforms.
- **Process drift (not code).** The cross-family reviewer rule could not be honored because no Codex or Grok was available. Reviews moved from headless `lattice code-review` to fresh-context Fable sub-agents after five watchdog deaths. Fable reviewed Opus work throughout, which keeps a cross-model check.
- **Loading-state status line.** While loading, the status line's right half keeps the previous `Page 1 of 2` next to `Loading`. This was a deliberate BC-14 decision (plan S2), and BRANDING §5.3 specifies only the left half.

## Gaps

None. Every audited SPEC criterion is satisfied by merged code on master 61228c7.

The places where the build departs from **BRANDING.md's text** are recorded, not silent. They sit in BC-14's open owner questions, and several of them are real text-versus-build mismatches the operator should see:

- **O1.** §4.4 says "one row: six fields plus Search and Clear", but there are now seven fields with the actions on row 2.
- **O6.** The loading segment is oxblood, which is not one of the five oxblood jobs in §2.2.
- **O7.** §5.9 says the first load shows "the status line alone", but the build shows the loading line on every load.
- **O12.** The results-region focus ring is a new oxblood use.
- **O14.** In phone landscape (667×375) the results region is 0px tall and the pinned pager is unreachable.
- **O15.** §4.4 says "Search full width" at ≤760, but the build has Search flex:1 beside Clear.

The build follows SPEC and the reasoning recorded in BC-14 on each. The owner's decision closes them.

## Recommendations

- **Fix-back-in-flight:** none required. No criterion failed.
- **New tickets:** (the validator mints none; this is for the Orchestrator and operator to route)
  - **Shipped CLAUDE.md and agents.md.** Their content is Lattice-mandatory ("You MUST use Lattice … `.lattice/` exists"), but `.lattice/` is excluded from the zip. A reader of the zip, or an agent working from it, is told to use a tracker that is not there. The operator ruled to ship both files, so this is not a defect. Consider a short preface in the shipped copies, or a dist-time note, so the disclosure reads as intended.
  - **O14 (phone landscape).** The pager is unreachable at 667×375. This is the one owner question with real usability impact. If the interviewer demos on a phone or in a small window, it shows. It is worth deciding before the live session.
  - Existing backlog BC-18..BC-24 already covers the other run findings, including BC-20 (Unicode keyword with SQLite ASCII `lower()`) and BC-22 (focus after Try again and Clear filters).
- **Accept-as-is:**
  - The merge subjects are git defaults.
  - The `price/target` float form.
  - The Vite proxy returns a 502 (`kind:'http'`, `requestId: null`) when the native API is down in dev, rather than `kind:'network'`.
  - The `X-Request-Id` header spelling is Go's canonical form.
  - `make smoke` unsets only `DOCKER_DEFAULT_PLATFORM`, so the classic builder still runs with the operator's `DOCKER_BUILDKIT=0`. It builds and passes.
  - Two unknown params collapse to a single "Unknown parameter." line in the UI.
  - EXTENSIONS.md line 4 still cites commit 0b24829; every name still resolves at 61228c7.
- **Before the live demo on this laptop:** use `make up`, or unset `DOCKER_DEFAULT_PLATFORM` (and ideally `DOCKER_BUILDKIT` and `COMPOSE_DOCKER_CLI_BUILD`) first. Plain `docker compose up` with the operator's `~/.zshrc` exports is the known footgun (run-state). The README says so, but a live session is where it would bite.

## What I couldn't verify

- **A real keyboard (row 44).** The c11 browser panel delivers synthetic key events, so a Tab press does not move focus and Enter does not activate a button. The tab order and activation were emulated with `focus()` and `click()`, and all aria attributes were observed live. Separately, WebKit and Safari skip buttons on Tab unless "Press Tab to highlight each item" is enabled, so a real Safari keyboard user reaches Search and the pager only with Option-Tab. That is platform behavior, but worth knowing for a demo. **Suggested:** the operator does a 30-second real-keyboard Tab pass in Chrome during row 55.
- **A 1280px viewport (row 35).** The panel was 763px wide and WKWebView cannot emulate a viewport. The DOM checks are width-independent and 763 is already in the above-760 layout. A 1280 layout was simulated by widening `html`/`body`. BC-14's 1280 screenshots (row 46) cover the real width.
- **Real wheel or trackpad scrolling (row 45).** The test used c11 `scroll` plus script scroll APIs. The document never moved under any of them.
- **A Linux Docker host.** It was not available. BC-23 records that on Linux, `make up` leaves root-owned bind-mount paths that break a later native `make check`. Row 54 would hit this on a Linux machine; the README carries a note.
- **Runtime tie-breaks.** All 12 sample listings have distinct `listedDate`s, so no API request exercises the §6.3 tie rules. Only the unit tests do (row 30).
- **Feedback on the validation plan itself, for the contract author:**
  - The curl shorthand `curl :18099/...` is a malformed URL on curl 8.7.1. Write `http://localhost:18099/...`.
  - Row 15's "12 total items" means `page.totalItems`, since page 1 holds 10.
  - Row 18's comma and repeat comparison should use `pageSize=100` to compare full sets.
  - Row 4's `grep -P` does not exist on macOS (no ggrep installed).
  - Rows 35, 44 and 45 name a viewport and input methods that c11 browser panels cannot produce. Name the observable instead, or tag these rows `post-merge-smoke`.
  - Row 9 names `/tmp/rv.db` while the Result Validator brief names `/tmp/brillio-rv.db`; the latter was used.

## Operator smoke-pass checklist (post-merge)

Every `runnable_at: post-merge-smoke` row in validation-plan.md, copied verbatim. The Result Validator did not attempt these.

| # | Criterion (ID) | Verification method | Artifact to inspect | Pass condition |
|---|---|---|---|---|
| 50 | §10 hot reload (BC-13) | With compose up, edit a Go file and a TSX file | BC-13 | api restarts inside the container; page hot-reloads; no compose restart |
| 54 | §12 fresh machine | Operator unzips on a clean directory (ideally another machine with only Docker) and follows only the README | BC-17 | `docker compose up` and `make check` both work as written |
| 55 | Felt: product look | Operator compares the running UI with docs/branding-demo.html side by side at both breakpoints | BC-14 | Operator judges it indistinguishable apart from the Status field |
| 56 | Felt: interview readiness | Operator rehearses the three EXTENSIONS.md mini-plans aloud against the code | BC-16 | Each seam is where the doc says, and the narration holds |

Notes for the operator:

- **Row 54:** on a Linux host, see BC-23 and the README note.
- **Row 55:** O1, O6, O7, O12, O14 and O15 above are exactly where "indistinguishable apart from the Status field" may not hold. Decide them as you look.
- **Shipped zip:** `dist/brillio-listings-20261014.zip` is verified byte-identical to a fresh `make dist` of 61228c7.

## Open human items (run-state "Final handoff checklist")

- **BC-14 needs_human: branding-owner questions O1–O15.** These go through the BRANDING §9 process and are non-blocking. The full text is in the BC-14 plan §4 (O1–O14) and the comment ev_01M4WFBQHHP9R5P1MTEX209RXR (O15). The ticket's needs_human reason string lists only O1–O14; O15 was added in that later comment.
- **Backlog tickets created during the run:**
  - BC-18: store hardening nits.
  - BC-19: text-button tracking and hover, for the branding owner.
  - BC-20: Unicode keyword versus SQLite ASCII `lower()`.
  - BC-21: a11y live feedback.
  - BC-22: focus after Try again and Clear filters.
  - BC-23: Linux Docker host root-owned bind-mount paths. The README note shipped; the real fix is pending.
  - BC-24: verify-dist signal-edge hardening.
- **Docker env footgun (run-state).** The live demo on this laptop must use `make up` or unset `DOCKER_DEFAULT_PLATFORM`.
- **The closeout audit is not done** (lessons routing to LESSONS.md, CLAUDE.md, the skill or tickets). It belongs to the Orchestrator after this report, and this validator did not perform it.
