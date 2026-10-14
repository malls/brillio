Lattice-Reviewed-Commit: a9b7f73ada5002f000ffe209da79af8fdb7756a3
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-2-backend-scaffold
Lattice-Reviewed-Base: master (2afc005e1fe0b4a3347e95cc540b53244ffd4cdd)
Lattice-Reviewed-Head: feat/BC-2-backend-scaffold (a9b7f73ada5002f000ffe209da79af8fdb7756a3)

# Code Review: BC-2 — Backend scaffold

Reviewed-Commit: a9b7f73ada5002f000ffe209da79af8fdb7756a3 (`feat/BC-2-backend-scaffold`)
Reviewed-Base: master (2afc005e1fe0b4a3347e95cc540b53244ffd4cdd)
Reviewer: agent:claude-fable-5.1-reviewer (cold read; did not write this code)

## 1. Verdict

**PASS**

Two minor, non-blocking findings below. Neither requires a rework cycle; both can be fixed inline as additive commits or routed as "ignore with justification" / follow-up task by the orchestrator.

## 2. Summary

Reviewed the full BC-2 diff (18 files, ~1,640 lines) against the plan, its cycle-1 resolutions (R1–R4 and the CORS-normalization finding), and SPEC §2, §3, §4, §5.1, §5.5, §5.6, §8, §11. The implementation matches the plan's binding design decisions D1–D10 point for point, every gate is green on this exact head, and the live binary does what the acceptance criteria demand (verified below, not assumed). Code quality is high: small, well-named functions, doc comments that carry real contracts for downstream tickets, and tests that assert observable behavior through the public `NewRouter` rather than internals. The one correctness-adjacent gap is a residual dead-configuration class in CORS origin validation (default and non-canonical ports), which is the same failure class the cycle-1 fix targeted but did not fully close.

### Gates run on a9b7f73

| Gate | Result |
|---|---|
| `go build ./...` | OK |
| `go vet ./...` | OK |
| `go test -race -count=1 ./...` | 82 tests pass, 4 packages |
| `golangci-lint config verify` | OK |
| `golangci-lint run` (`make lint-api`) | 0 issues |
| `make test-api` | OK |
| `make check` | fails only on `lint-web` BC-1 stub (expected per run-state gate exception) |
| `git -c core.excludesFile=/dev/null status --porcelain` | clean; all 18 files tracked |

### Live validation performed (built binary, port 18077/18078)

- `GET /healthz` → 200, `Content-Type: application/json; charset=utf-8`, generated 32-hex `X-Request-Id`, body `{"status":"ok"}`.
- `GET /healthz` with `X-Request-ID: bc2-rev-1` → echoed verbatim.
- `HEAD /healthz` → 200, `Content-Length: 16`, no body (net/http suppresses it).
- `GET /nope` → 404 `{"error":{"code":"not_found","message":"Resource not found."}}`.
- `POST /healthz` → 405 with `Allow: GET, HEAD` and the `method_not_allowed` envelope.
- `OPTIONS /healthz` with no Origin → 405 envelope (falls through to mux, per D4).
- Allowed preflight (`Origin: http://localhost:5173`) → 204, ACAO echoed, `Allow-Methods: GET, OPTIONS`, `Allow-Headers: Content-Type, X-Request-ID`, `Expose-Headers: X-Request-ID`, `Vary: Origin`, empty body.
- Disallowed preflight (`Origin: http://evil.test`) → 204 with only `Vary: Origin` and `X-Request-Id`; no Allow-* headers.
- `GET /x/../nope` and `GET //healthz` (sent with `--path-as-is`) → 307 stdlib redirect, request ID present.
- `kill -TERM` → log lines `"shutting down"` then `"server stopped"`, **exit=0**. `kill -INT` → same, exit=0.
- `LOG_LEVEL=verbose SHUTDOWN_TIMEOUT=abc` → one `"fatal"` JSON line naming both variables, exit=1.
- Second instance on the same port → `"fatal" ... bind: address already in use`, exit=1.
- Request log lines carry `method`, `path`, `status`, `duration_ms`, `request_id` (equal to the response header), `remote_addr`.

### Cycle-1 finding verification

The binding fix from the cycle-1 review is fully present in a9b7f73: scheme/host compared case-insensitively and stored lowercased; a host ending in a bare `:` is a startup error; package doc updated (R4); two new `TestLoad` rows plus an uppercase-scheme row; `backend/.gitkeep` removed. Doc comments on `Listing` and `Listing.ListedDate` are prescriptive as R1 required.

## 3. Issues

**[MINOR] backend/internal/config/config.go:161 — CORS origin validation still admits dead values for default and non-canonical ports**
`http://localhost:80`, `https://app.test:443` and `http://a.test:08080` all pass `parseOrigins` and are stored verbatim (verified with a `net/url` probe: all three accepted, ports `"80"`, `"443"`, `"08080"`). A browser serializes `Origin` per the URL spec: the port is omitted when it is the scheme's default, and is always a decimal integer with no leading zeros. So none of these can ever equal a real `Origin` header, which is exactly the "silently dead configuration" class that D1 and the cycle-1 fix exist to prevent. Low real-world likelihood (nobody configures `:80` on purpose), so this does not block the PASS.
**Fix:** after the bare-colon check, when `u.Port() != ""`: parse it with `strconv.Atoi`; reject if it is not in 1..65535 or if `strconv.Itoa(p) != u.Port()` (leading zeros); and either reject or normalize away `:80` for `http` and `:443` for `https` (normalizing is friendlier and matches the lowercasing precedent). Extend the package doc sentence (R4) and add `TestLoad` rows: `http://localhost:80` → `http://localhost` (or error), `https://app.test:443` → `https://app.test` (or error), `http://a.test:08080` → error. About 8 LOC plus 3 test rows; additive, no API change.

**[MINOR] backend/internal/httpapi/router_test.go — the nil-`Logger` branch of `NewRouter` is untested**
D3 makes "a nil Logger becomes `slog.DiscardHandler`" a contract that BC-5 and BC-6 tests may rely on when they construct `Deps{}` without a logger. Every current test passes a real logger via `newTestRouter`, so the `if log == nil` branch in `NewRouter` has no coverage and a future regression (e.g. someone moving the nil check) would surface only as a nil-pointer panic in a downstream ticket's tests.
**Fix:** one subtest, e.g. in `TestHealthz` or a new `TestNewRouterDefaults`: `httpapi.NewRouter(httpapi.Deps{})` served `GET /healthz` returns 200 and a request ID without panicking. Three or four lines.

No critical or major issues found.

## 4. Positive Observations

- **404/405 via `ServeMux.Handler` probing (D5) is implemented exactly as designed and the reasoning is sound.** `go doc` on go1.27.1 confirms the empty-pattern contract, the code routes matched requests through `mux.ServeHTTP` so `r.Pattern`/`PathValue` work, the stdlib's own `Allow` header is preserved (`GET, HEAD` observed live), and non-canonical paths keep the stdlib 307. The doc comment on `fallbacks` explains the invariant that makes the probe side-effect-free, which is the one thing a future maintainer needs to know.
- **Middleware order is correct and tested at the boundaries.** Request ID is outermost so it appears on 204/404/405/500 (asserted by a shared table in `TestRequestID`), the log wraps recovery so a recovered panic logs status 500 (asserted), and `http.ErrAbortHandler` is re-panicked rather than swallowed (asserted, including that it is not logged as a recovered panic).
- **Panic recovery leaks nothing.** `TestPanicRecovery` asserts the body byte-for-byte and additionally greps for `goroutine`, `.go:` and the panic value, while asserting the stack and request ID are in the log. This is precisely §5.5's "details go to the log with the request ID, never to the client".
- **`writeJSON` encodes before writing**, so an unencodable value becomes a clean 500 instead of a truncated 200. The fallback body is pre-encoded once and never mutated (no `append` aliasing). The internal test for the NaN path is a nice touch given no route can trigger it yet.
- **Config loader is testable by construction** (`Load(getenv)`), collects every error with `errors.Join` under the variable name, and the live run confirms a single fatal line naming both bad variables. The package doc sentence enumerating every rule (R4) is exactly what BC-13/BC-15 need to quote.
- **Domain types freeze cleanly.** `allStatuses` is the single source for both `ParseStatus` and `AllStatuses`, the latter returns a clone (tested for independence), and the `Listing`/`ListedDate` doc comments carry the wire-format rule prescriptively so BC-3/BC-4/BC-5 cannot miss it. `TestListingJSONKeys` pins the 15 §4 names.
- **main.go is linear and small** (D8), with fixed server timeouts kept out of §11 config, `signal.NotifyContext` plus an explicit `stop()` so a second signal force-quits (commented), `ErrorLog` bridged into slog, and the shutdown error path closing the server. Exit codes verified: 0 on SIGTERM/SIGINT, 1 on config error, bind failure.
- **Tests assert behavior, not implementation.** All httpapi tests go through `NewRouter` with `httptest.NewRecorder` (R2), use a panicking fake `Pinger` instead of a test-only route seam, decode envelopes with `DisallowUnknownFields`, and check `fields` is absent outside validation errors. Test names read as specifications.
- **Shared-file discipline respected.** Only the three BC-2 recipe lines in the Makefile changed (`.PHONY` kept), only the `# Go (backend)` section of `.gitignore` changed, no `.lattice/` staged, `go.mod` is exactly `go 1.27`, no `go.sum` (no deps), no `query.go`/`search.go`/`scoring.go` created. Commits are conventional and emoji-free per SPEC §2.
- **Accepted-by-plan notes (not findings):** the request ID has no length cap (D4, bounded by `MaxHeaderBytes`); the 307 redirect body is stdlib HTML rather than an envelope (D5 keeps it); the `dev-api` recipe runs the built binary from the repo root instead of `go run` (D10, so `./data/...` defaults resolve after BC-4).

## Routing recommendation

PASS. Both minor findings are additive and safe after the `review` freeze (no renames, no type changes). Recommended: fix inline in one small commit (`fix(config): normalize default ports in CORS origins; test nil-logger default`), re-run `make lint-api test-api`, then advance to `in_validation`. Acceptable alternative: route finding 1 as a follow-up task and finding 2 as "ignore with justification" (trivial branch), and advance now.
