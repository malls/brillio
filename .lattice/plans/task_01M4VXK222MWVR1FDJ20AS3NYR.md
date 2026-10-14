# BC-2 plan: Backend scaffold (Go 1.27 server, config, domain types, middleware, lint)

Spec: docs/SPEC.md §2, §3, §4, §5.1, §5.5, §5.6, §8 (httpapi + Lint), §11. Binding notes:
`.lattice/orchestration/prompts/BC-2.md` and `run-state.md` (shared files, gate, ports).
Planner checks on the host: go1.27.1 (CGO_ENABLED=1, clang present, so `-race` works),
golangci-lint 2.14.0, and the planned `.golangci.yml` passes `golangci-lint config verify`.

## Scope

In: `backend/` Go module (stdlib only), `cmd/api` wiring, `internal/config`, `internal/domain` (shared
types only), `internal/httpapi` (router, middleware, JSON and error helpers, `/healthz`), `backend/.golangci.yml`,
the Makefile BC-2 block, the backend lines of `.gitignore`, and tests.
Out: SQLite and store (BC-4), scoring/ranking/paging (BC-3), search and cities handlers (BC-5), OpenAPI (BC-6),
Docker/air (BC-13), README (BC-15). No `go.sum` (no dependencies yet).

## Files (worktree-relative)

| File | Purpose |
|---|---|
| `backend/go.mod` | `module brillio/backend`, line exactly `go 1.27` (run `go mod edit -go=1.27` if `go mod init` writes a patch version). |
| `backend/.golangci.yml` | v2 schema: standard linters plus errcheck, govet, staticcheck, revive; gofmt formatter. |
| `backend/cmd/api/main.go` | Wiring only: config, logger, router, listener, `http.Server`, signal-driven graceful shutdown, exit code. |
| `backend/internal/config/config.go` | `Config`, `Load(getenv)`, §11 defaults and validation. |
| `backend/internal/config/config_test.go` | Table test for `Load`. |
| `backend/internal/domain/listing.go` | Package doc, `DateLayout`, `ListingStatus` + constants, `ErrInvalidStatus`, `ParseStatus`, `AllStatuses`, `Listing`. |
| `backend/internal/domain/filter.go` | `SearchFilter`, with field semantics documented in comments. |
| `backend/internal/domain/listing_test.go` | `ParseStatus`/`AllStatuses` table, `Listing` JSON key set. |
| `backend/internal/httpapi/router.go` | Package doc, `Pinger`, `Deps`, `NewRouter`, unexported `api` struct, 404/405 envelope fallback. |
| `backend/internal/httpapi/middleware.go` | Request ID (+ `RequestIDFrom`), request log, panic recovery, CORS, `statusWriter`. |
| `backend/internal/httpapi/respond.go` | `FieldError`, envelope types, `writeJSON`, `writeError`, error code/message constants. |
| `backend/internal/httpapi/health.go` | `GET /healthz` handler. |
| `backend/internal/httpapi/{router,middleware,health}_test.go` | External package `httpapi_test`, run through `NewRouter` with `httptest`. |
| `Makefile` | BC-2 block only: `dev-api`, `lint-api`, `test-api`. |
| `.gitignore` | `# Go (backend)` section only: three anchored lines. |

Do not create `query.go`, `search.go`, `cities.go` (BC-5) or `scoring.go`/`rank*.go`/`paginate*.go` (BC-3).

## Design decisions (binding; do not re-derive)

**D1 Config.** `type Config struct { Port int; DBPath, ListingsFile string; LogLevel slog.Level; CORSOrigins []string;
ShutdownTimeout time.Duration }`, `func Load(getenv func(string) string) (Config, error)`, `func (c Config) Addr() string`
(`":"+port`). main passes `os.Getenv`; tests pass a map lookup and never touch the real environment. Every value is
`TrimSpace`d, and empty means default (§11: 8080, `./data/listings.db`, `./data/sample_listings.json`, info,
`http://localhost:5173`, 10s). Rules: PORT is an integer in 1..65535. LOG_LEVEL is lowercased and must be one of
debug|info|warn|error, mapped explicitly (not `slog.Level.UnmarshalText`, which also accepts `INFO+2`). SHUTDOWN_TIMEOUT
must pass `time.ParseDuration` and be > 0. CORS_ORIGINS is split on `,`, each entry trimmed, empties dropped. Each entry
must `url.Parse` to scheme http|https with a non-empty Host and no path, query, fragment or userinfo. The check exists
because a trailing slash would silently never match a browser's `Origin`. Zero origins is an error, and `*` is rejected
because §5.6 has no wildcard. DB_PATH and LISTINGS_FILE pass through unvalidated (BC-4 owns them). Every problem is
collected and returned with `errors.Join`, each prefixed with the variable name (`LOG_LEVEL: "verbose" is not one of
debug|info|warn|error`).

**D2 Domain types (frozen once the task reaches `review`, because BC-3 and BC-4 branch from it).**
`Listing{ID, Source, Address, City, State, Zip string; Price float64; Bedrooms int; Bathrooms float64; Sqft int;
Latitude, Longitude float64; ListedDate time.Time; Status ListingStatus; Description string}` with JSON tags exactly per
§4 (`id source address city state zip price bedrooms bathrooms sqft latitude longitude listedDate status description`).
`type ListingStatus string` with `StatusActive="active"`, `StatusPending="pending"`, `StatusSold="sold"`.
`var ErrInvalidStatus`. `ParseStatus(s string) (ListingStatus, error)` does an exact, case-sensitive match with no
trimming (callers trim), and its error is `fmt.Errorf("%w: %q", ErrInvalidStatus, s)`. `AllStatuses() []ListingStatus`
returns a fresh slice in the order active, pending, sold. ParseStatus and BC-5's "absent means all three" share this one
list, so it is not hard-coded in two places. `SearchFilter` is exactly the ticket's shape. Doc comments carry its
semantics: a nil pointer means absent, `City` is trimmed and `""` means absent, `Keyword` is the raw value echoed in
appliedFilters, `KeywordTokens` are the lowercased tokens, and `Statuses` is never empty after parsing.
`const DateLayout = "2006-01-02"`.
*ListedDate tension (flagged, not a SPEC edit):* §4 says `time.Time` holding a date only, and BC-3's
`AgeDays(listedDate, now time.Time)` needs that type. Encoded as-is, `time.Time` produces RFC 3339 and rejects
`"2026-08-29"` on decode. So `Listing` is the in-memory model, and the wire date is parsed and formatted with
`DateLayout` at the boundaries: BC-4's seed decodes into its own record struct (it needs per-field errors anyway), and
BC-5 maps the response through its own DTO. State this in the `ListedDate` doc comment. The delegator forwards a one-line
note to BC-4 and BC-5. No custom `MarshalJSON` (keeps zero logic).

**D3 httpapi shape.** `type Pinger interface{ Ping(ctx context.Context) error }`.
`type Deps struct { Logger *slog.Logger; CORSOrigins []string; Pinger Pinger }`. A nil Logger becomes
`slog.New(slog.DiscardHandler)`, and a nil Pinger means no DB check. A neutral comment in `Deps` says that search
dependencies (searcher, cities source, `Now func() time.Time`) are added as fields. Do not declare them now, and put no
ticket numbers in shipped code. Handlers are methods on an unexported `type api struct{ deps Deps; log *slog.Logger }`.
`NewRouter` is the only place routes are registered, as plain `mux.HandleFunc("GET /healthz", a.healthz)`, so BC-5 and
BC-6 add one line each and get envelope 404/405 automatically.

**D4 Middleware order (outermost to innermost):** `requestID → logRequests → recoverPanic → cors → fallbacks(mux)`.
Request ID is outermost, so every response carries it: preflights, 404s, 405s and 500s. The log wraps recovery, so a
recovered panic logs status 500. CORS sits inside recovery and answers preflights before routing.
- *requestID:* use the inbound `X-Request-ID` verbatim when non-empty. Otherwise generate 16 bytes from `crypto/rand`
  and hex-encode them (32 chars); `rand.Read` cannot fail since Go 1.24, so assign `_, _ =` with a comment. Set the
  response header before calling next, and store the ID in the context under `type requestIDKey struct{}`. Export
  `func RequestIDFrom(ctx context.Context) string` for handlers and logging. There is no length cap, because the
  contract says to echo. The value is safe because net/http rejects CR/LF in inbound headers and slog's JSON handler
  escapes it. Go writes the header on the wire as `X-Request-Id` (canonical form); header names are case-insensitive, so
  this is not a deviation.
- *logRequests:* wrap in `statusWriter{http.ResponseWriter; status int; wrote bool}`, which has `Unwrap()` so
  `http.ResponseController` keeps working and defaults to 200. Emit one `LogAttrs` info line with msg `"request"` and
  keys `method`, `path` (URL.Path), `status`, `duration_ms` (float, microsecond precision), `request_id`, `remote_addr`.
- *recoverPanic:* wrap its own `statusWriter`, then `defer` a recover. Re-panic on `http.ErrAbortHandler`. Log at error
  level with msg `"panic recovered"`, `request_id`, `panic` (`fmt.Sprint`) and `stack` (`debug.Stack()`). Write the 500
  `internal_error` envelope only if nothing was written yet. The body never contains the panic value or the stack.
- *cors:* always `Header().Add("Vary","Origin")`. A request is allowed when `Origin` is non-empty and exactly equals one
  of `Deps.CORSOrigins`; it then gets `Access-Control-Allow-Origin: <origin>` and
  `Access-Control-Expose-Headers: X-Request-ID`. A preflight is `OPTIONS` + `Origin` + `Access-Control-Request-Method`.
  If allowed, add `Access-Control-Allow-Methods: GET, OPTIONS` and
  `Access-Control-Allow-Headers: Content-Type, X-Request-ID`. Every preflight returns 204 with no body and never reaches
  the mux. A disallowed origin gets 204 with no Allow-* headers, so the browser blocks it (no 403, because §5.5 has no
  such code and §5.6 says preflight returns 204). Non-preflight requests from a disallowed origin, or with no Origin,
  are served normally with no CORS headers. An OPTIONS that is not a preflight falls through to the mux and gets a 405
  envelope. No credentials, no Max-Age, no wildcard (none are in the contract).

**D5 404/405 envelopes with the Go 1.22+ ServeMux (chosen approach: classify via `ServeMux.Handler`).**
```go
func (a *api) fallbacks(mux *http.ServeMux) http.Handler // sketch of behavior:
// h, pattern := mux.Handler(r); if pattern != "" { mux.ServeHTTP(w, r); return }  // real route or canonical redirect
// probe := &probeWriter{header: http.Header{}}; h.ServeHTTP(probe, r)              // stdlib's own 404/405/redirect handler
// 404 -> writeError(404 not_found); 405 -> copy probe "Allow" to w, writeError(405 method_not_allowed); else h.ServeHTTP(w, r)
```
Go 1.27 documents `Handler` this way: "If there is no registered handler that applies to the request, Handler returns a
'page not found' or 'method not supported' handler and an empty pattern." With an empty pattern, `h` is always stdlib
code, never a user handler, so probing it into a discard writer has no side effects. The stdlib still computes `Allow`
(including HEAD). Matched requests go through `mux.ServeHTTP`, so `r.Pattern` and `PathValue` are populated. Planner
probe on go1.27.1 gave: `POST /healthz`→405 with pattern "" and Allow "GET, HEAD"; `GET /nope`→404 ""; `GET /a/../nope`→307
"" (handled by the default branch, so the stdlib redirect is kept); `HEAD /healthz`→matches `GET /healthz`.
Rejected alternatives:
- A catch-all `"/"` pattern matches every method, so `POST /healthz` would route to it and the 405 would be lost.
- A method-less twin pattern per path re-implements the stdlib's Allow bookkeeping and forces every route through a
  helper.
- Rewriting any 404/405 by wrapping the writer cannot tell the mux's 404 from a handler's deliberate one.
Cost: a second route lookup per request, which is negligible.

**D6 Response helpers.** `type FieldError struct{ Field, Code, Message string }` with tags `field`, `code`, `message`,
exported for BC-5. Envelope: `{"error":{"code","message","fields"}}`, where `fields` is `omitempty` (400 only).
`func (a *api) writeJSON(w, r, status int, v any)` calls `json.Marshal` first. On failure (for example NaN) it logs at
error level with `request_id` and writes a pre-encoded 500 envelope. Otherwise it sets
`Content-Type: application/json; charset=utf-8`, calls `WriteHeader(status)` and writes the body plus `"\n"`
(`_, _ = w.Write`: the client is gone). `func (a *api) writeError(w, r, status int, code, message string, fields []FieldError)`
builds the envelope. Constants: `not_found` "Resource not found.", `method_not_allowed` "Method not allowed.",
`internal_error` "Internal server error.", following the plain-sentence voice of §5.5.

**D7 /healthz (implement the 503 branch now; the interface exists and is testable with fakes).** With a nil Pinger it
returns 200 `{"status":"ok"}`. Otherwise it calls `Ping` with `context.WithTimeout(r.Context(), 2*time.Second)`. On error
it logs at error level (`"health check failed"`, request_id, err) and returns 503
`{"status":"degraded","error":"database unavailable"}`. The body is generic because §5.1's `"..."` is unspecified and
§5.5 keeps error details in the log. After this, BC-4 only sets `Pinger: store` in main.

**D8 main.go.** `main()` calls `run() error`. On error it logs `"fatal"` with `err` through a bootstrap JSON logger on
stdout and calls `os.Exit(1)`; this covers config errors and bind errors. `run`:
1. `config.Load(os.Getenv)`; on error return `config: %w`.
2. Build `slog.New(slog.NewJSONHandler(os.Stdout, &HandlerOptions{Level: cfg.LogLevel}))`.
3. `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`.
4. `net.Listen("tcp", cfg.Addr())` synchronously; a bind failure exits 1.
5. Log info `"listening"` with addr, log_level, cors_origins, shutdown_timeout.
6. Run `srv.Serve(ln)` in a goroutine that writes to a buffered errCh. The server has `ReadHeaderTimeout` 5s,
   `ReadTimeout` 10s, `WriteTimeout` 15s, `IdleTimeout` 60s (constants in main, not §11 config) and
   `ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelWarn)`.
7. `select` on errCh (return error) or `ctx.Done()`.
8. Call `stop()` so a second signal force-quits.
9. Log info **`"shutting down"`** with `timeout`, then call `srv.Shutdown` with a `cfg.ShutdownTimeout` deadline. On
   error: `srv.Close()`, return the error, exit 1.
10. Log `"server stopped"`, return nil, exit 0.

**D9 Lint config:** `version: "2"`, `linters.default: standard`, `linters.enable: [errcheck, govet, staticcheck, revive]`
(listed explicitly to mirror §8), and `formatters.enable: [gofmt]`. Add no exclusion presets. Instead, every package
gets a package comment, every exported identifier gets a doc comment, and unused parameters are named `_`. Revive
defaults enforce exactly these rules; the planner probe confirmed it.

**D10 Makefile (BC-2 block only; keep its `.PHONY` line):**
`dev-api` runs `cd backend && go build -o bin/api ./cmd/api` and then, on its own recipe line from the repo root,
`./backend/bin/api`. Running from the root lets §11's relative defaults (`./data/...`) resolve to the repo's `data/`
once BC-4 reads them. `go run` from `backend/` would look in `backend/data/` and break `make dev-api` after BC-4.
`lint-api` runs `cd backend && golangci-lint run`; `test-api` runs `cd backend && go test -race ./...`. The ticket's
"(go run ./cmd/api)" acceptance is still met and validated directly from `backend/`.
**`.gitignore`:** append `/backend/api` (stray plain `go build`), `/backend/cover.out` and `/backend/coverage.html`
under `# Go (backend)`. `test-api` writes no profile; these lines cover ad-hoc coverage runs in a fresh clone.

## Tests (stdlib `testing`, table-driven, no testify)

| Ticket / SPEC item | Test |
|---|---|
| Config table | `TestLoad`. Rows: all defaults. Full valid override. Each LOG_LEVEL, plus `WARN`. Invalid `verbose`/`warning`. SHUTDOWN_TIMEOUT `500ms` OK; `abc`/`0s`/`-1s`/`10` fail. PORT `18002` OK; `abc`/`0`/`65536` fail. CORS ` http://a.test , http://b.test:8080 ,` gives 2; `http://a.test/`, `*`, `localhost:5173`, ` , ` fail. Two bad vars give one error naming both. |
| Domain compiles, zero logic | `go build ./...`. `TestParseStatus`: 3 OK; `Active`, ` active`, `""`, `closed` all satisfy `errors.Is(ErrInvalidStatus)`. `TestAllStatuses`: order, and the returned slice is independent. `TestListingJSONKeys`: marshal a `Listing`; its key set equals the 15 §4 names. |
| healthz 200 + CT + X-Request-ID; §5.1 | `TestHealthz`. Rows: nil Pinger 200 `{"status":"ok"}`. OK Pinger 200. Failing Pinger 503 degraded body, with the real error in the log. HEAD 200 with empty body. Every row checks CT `application/json; charset=utf-8` and a non-empty X-Request-ID. |
| Request ID echoed / generated; §8 | `TestRequestID`. Echoed: `abc-123` exactly. Generated: matches `^[0-9a-f]{32}$`, and two requests differ. Present on 404, 405, 500 and 204 (shared assert helper). |
| Panic gives 500 envelope, no stack | `TestPanicRecovery` via a panicking fake Pinger through `NewRouter` (no test-only route seam). Body exactly `{"error":{"code":"internal_error","message":"Internal server error."}}`, with no `goroutine`, `.go:` or panic value. Log has `panic recovered` with `request_id` equal to the header and a non-empty `stack`, and the request line has status 500. |
| CORS preflight 204; §5.6 | `TestCORS`. Rows: allowed preflight gives 204, the exact ACAO / Allow-Methods / Allow-Headers / Expose values, `Vary: Origin` and an empty body. Disallowed preflight gives 204 with no Access-Control-Allow-*. Allowed GET gets ACAO + Expose. Disallowed GET and no-Origin GET get no ACAO. Second configured origin matches. Non-preflight OPTIONS gives a 405 envelope. |
| 404 / 405 envelopes; §5.5 | `TestErrorEnvelopes`. Rows: `GET /nope` and `GET /api/listings` give 404 `not_found`. `POST` and `DELETE /healthz` give 405 `method_not_allowed` with `Allow: GET, HEAD`. `GET /x/../nope` gives a 307 stdlib redirect. Bodies decode with `DisallowUnknownFields`. |
| §11 request log line | `TestRequestLog`: exactly one `msg:"request"` JSON line with method, path, status, duration_ms, request_id (== header) and remote_addr, for a 200 and a 404. |
| SIGTERM line + exit 0 | Validation only. No in-process signal or port test, per CLAUDE.md guidance on concurrent-process tests. |

Test logger helper: `slog.New(slog.NewJSONHandler(&buf, nil))`, with lines split and `json.Unmarshal`ed per line.
Gate: `make lint-api test-api` green. Also run `go vet ./...`, and run `git -c core.excludesFile=/dev/null status --porcelain`
to confirm every new file shows up regardless of the host's global ignore.

## Validation (one Bash call: the cwd resets between calls, and `wait` needs the same shell)

1. `lsof -nP -iTCP:18002 -sTCP:LISTEN` must be empty. Then `cd backend && go build -o bin/api ./cmd/api`.
2. Start `PORT=18002 ./bin/api >"$SCRATCH/api.log" 2>&1 &` and record `PID=$!`. Poll `/healthz` up to about 5 seconds.
3. Run `curl -si` for:
   - `/healthz`, expecting a generated 32-hex X-Request-Id.
   - `/healthz` with `-H 'X-Request-ID: bc2-val-1'`, expecting the ID echoed.
   - `/nope`, expecting a 404 envelope.
   - `-X POST /healthz`, expecting a 405 envelope with Allow.
   - `-X OPTIONS` with `-H 'Origin: http://localhost:5173' -H 'Access-Control-Request-Method: GET'
     -H 'Access-Control-Request-Headers: X-Request-ID'`, expecting 204 with the full header set.
   - The same preflight with `Origin: http://evil.test`, expecting 204 with no Allow-*.
4. `kill -TERM $PID; wait $PID; echo "exit=$?"` must print `exit=0`. Then `cat api.log`, showing the request lines,
   `"msg":"shutting down"` and `"server stopped"`.
5. Config error: `LOG_LEVEL=verbose PORT=18002 ./bin/api; echo "exit=$?"` gives a fatal JSON line and a non-zero exit.
6. Ticket bullet: `PORT=18002 go run ./cmd/api &`, then curl `/healthz` (200, CT, X-Request-Id). Then
   `lsof -ti tcp:18002 | xargs kill`. The `go run` wrapper does not forward a `kill` to its child, so the SIGTERM check
   uses the built binary.
7. Final `lsof` shows nothing on 18002. Attach the transcript with `--role validation`.

## Shared files and downstream

- Makefile: only the three BC-2 recipe lines change; web targets and the `check`/`lint`/`test` composition stay
  untouched (BC-7 edits in parallel). `.gitignore`: only the `# Go (backend)` section changes.
- BC-2 owns `main.go`, `httpapi` Deps and the router.
  - BC-4 adds store open/migrate/seed in `run()` and `Pinger: st`.
  - BC-5 adds Deps fields plus two route lines.
  - BC-6 adds one route line.
  Keep `run()` linear so each insert is a contiguous block.
- `internal/domain`: BC-3 adds files only, and type changes go through a SPEC edit. After `review`, commits are
  additive only: no renames, no rebase. Never stage `.lattice/`.

## Risks

1. Revive with no exclusions flags every missing doc or package comment and every unused parameter (probe-confirmed).
   Write comments up front.
2. `ServeMux.Handler` empty-pattern behavior depends on the 1.22+ mux. `GODEBUG=httpmuxgo121=1` would disable 405
   detection; it is not set, and the `TestErrorEnvelopes` rows would catch it.
3. ListedDate encoding (D2): downstream tickets must not encode or decode `Listing` straight to the wire. The doc
   comment plus the forwarded note mitigate this.
4. The host's global `core.excludesFile` hides `*.out`/`*.log` locally. Verify with
   `core.excludesFile=/dev/null` (above). Write logs to the scratchpad, never the repo.
5. A port 18002 collision, or a leftover process if the script aborts. Check with `lsof` before and after.
6. `go mod init` may write `go 1.27.1`. Pin it to `go 1.27`.

SPEC edits proposed: none. The `dev-api` recipe refines the ticket's "(go run ./cmd/api)" parenthetical (D10); it does
not touch SPEC.

## Size

About 500 LOC of code: config 90, domain 75, httpapi 260, main 85, plus config/Makefile/.gitignore around 20. About 450
LOC of tests. Roughly 950 to 1000 lines in total, about 3 commits:
1. module + config + domain
2. httpapi + main
3. lint/Makefile/.gitignore

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Source: plan-review artifact art_01M4W0VJ7731216A8G38XW3DTG (verdict PASS). Triage by agent:opus-5.5-delegator.

| # | Finding | Bucket | Resolution | Section |
|---|---|---|---|---|
| R1 | [MAJOR] ListedDate wire-format decision needs a durable downstream handoff | obvious | Keep D2's design (`time.Time` + `DateLayout`, DTOs at the boundaries). The `Listing.ListedDate` doc comment and the `Listing` type doc comment are **prescriptive**: "Listing is the in-memory model; do not json.Marshal/Unmarshal it (or any struct embedding it) to the wire. ListedDate is a calendar date; parse and format it with DateLayout in a boundary record/DTO." The handoff itself is the delegator's job, not the implementer's: at the `review` transition the delegator posts `lattice comment`s on BC-3, BC-4 and BC-5 with that rule and mails the Orchestrator to add it to the run-state `internal/domain` shared-files row. | D2, Shared files and downstream |
| R2 | [MINOR] httptest mode unstated; HEAD and 307 rows need opposite modes | obvious | All httpapi tests use `httptest.NewRecorder` against `NewRouter(...)` (no `httptest.NewServer`). The `TestHealthz` HEAD row asserts status 200, Content-Type and a non-empty X-Request-ID only; it does NOT assert an empty body (body suppression for HEAD is net/http's job; the validation run shows it on the real server). The 307 row stays as written. | Tests table, Files row for `*_test.go` |
| R3 | [MINOR] /healthz 503 branch overlaps BC-4's card | obvious | Keep D7. The delegator's `review`-time `lattice comment` on BC-4 states that the 503 branch and the failing-Pinger test already exist, so BC-4 only wires `Pinger: st` in `run()` and may extend the test with the real store. No implementer change. | D7, Shared files and downstream |
| R4 | [MINOR] D1 startup validation broader than the ticket | obvious | Keep D1. The `internal/config` package doc comment lists every validation rule in one sentence (LOG_LEVEL in debug/info/warn/error; SHUTDOWN_TIMEOUT a positive Go duration; PORT an integer 1..65535; each CORS_ORIGINS entry a bare `http(s)://host[:port]` with no path, trailing slash or `*`, at least one entry) so BC-13 and BC-15 can quote it. The delegator mentions the rule in the LAND-READY handoff. | D1 |

No evolutionary or complex findings; no needs-human.

## Review Cycle 1 Findings (code review art_01M4W1SKATQMGBTAQ2BWVXMEK5, PASS @ 5da9b80)

- **[MINOR] backend/internal/config/config.go:149 — CORS origin validation admits dead values.** `http://LocalHost:5173`
  (mixed-case host) and `http://a.test:` (empty port after colon) pass the exact `scheme://host == entry` check, but a
  browser's `Origin` is always lowercase and never ends in a bare colon, so either is silently dead configuration (the
  failure class D1 exists to prevent). **Fix (binding):** after the structural check, store the origin normalized to
  `strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)`, and reject an entry whose host ends in `:` with an empty
  port. Update the `internal/config` package doc sentence listing the rules (R4). Add `TestLoad` rows:
  `http://LocalHost:5173` loads as `http://localhost:5173`; `http://a.test:` is a startup error.
- **Cleanup (implementer's own note):** delete the now-redundant `backend/.gitkeep` in the same commit.
- Routing: fix (delegator decision; verdict was PASS, fix is additive, ~10 LOC + 2 test rows). Re-review must name the new HEAD.

## Reset 2026-10-14 by agent:opus-5.5-delegator
