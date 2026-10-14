Lattice-Reviewed-Commit: 5da9b806f79ba0c7e3db2eb8a4cecbfd781e9f0c
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-2-backend-scaffold
Lattice-Reviewed-Base: master (2afc005e1fe0b4a3347e95cc540b53244ffd4cdd)
Lattice-Reviewed-Head: feat/BC-2-backend-scaffold (5da9b806f79ba0c7e3db2eb8a4cecbfd781e9f0c)

# Code Review: BC-2 — Backend scaffold

Reviewed commit: `5da9b806f79ba0c7e3db2eb8a4cecbfd781e9f0c` (`feat/BC-2-backend-scaffold`) against base `2afc005`.
Reviewer: `agent:claude-fable-5.1-reviewer`, cold read of the plan, the diff and the spec (§2, §3, §4, §5.1, §5.5, §5.6, §8, §11).

## 1. Verdict

**PASS**

No implementation-level or plan-level rework is needed. One minor, non-blocking finding is recorded below for the orchestrator to route (fix inline or defer with justification).

## 2. Summary

I reviewed the 18-file diff (module, config loader, domain types, httpapi router/middleware/respond/health, main, lint config, Makefile, .gitignore, and six test files) against the plan's binding decisions D1–D10 and the four acceptance criteria. The implementation follows the plan closely and faithfully, with no unexplained deviations; the code is idiomatic, well-commented, stdlib-only, and the middleware chain, envelope fallbacks and graceful shutdown are correct. Every acceptance criterion was verified: gates pass locally (`go build`, `go vet`, `golangci-lint run` clean, `golangci-lint config verify` ok, `go test -race ./...` 79 tests green in 4 packages), and a live smoke run of the built binary confirmed `/healthz` 200 with `Content-Type: application/json; charset=utf-8` and both a generated 32-hex and an echoed `X-Request-ID`, 404/405 envelopes with `Allow: GET, HEAD`, allowed and disallowed CORS preflights, a `"shutting down"` log line followed by `"server stopped"` and exit code 0 on SIGTERM, and a fatal JSON line with exit 1 on an invalid `LOG_LEVEL`/`SHUTDOWN_TIMEOUT`.

## 3. Issues

**[MINOR] backend/internal/config/config.go:149 — Origin validation still admits two values that can never match a browser's `Origin`**

D1's stated reason for validating `CORS_ORIGINS` is that "a trailing slash would silently never match a browser's `Origin`." The check `u.Scheme+"://"+u.Host == entry` rejects paths, query, fragment, userinfo and uppercase schemes (because `url.Parse` lowercases the scheme), but it accepts a mixed-case host (`http://LocalHost:5173`) and an empty port after a colon (`http://a.test:`). I confirmed both pass the exact check used in the file. Browsers serialize the `Origin` host in lowercase and never emit a bare trailing colon, and `cors` compares with `slices.Contains` exactly, so either value is silently dead configuration — the same failure class D1 set out to prevent. This is outside the rule list the plan enumerated (and that the package doc now quotes per R4), so it is not a plan violation; it is a small gap in the rationale's coverage.

**Fix:** Either normalize before storing (`entry = strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)` after the structural check, and reject when `u.Port() == "" && strings.HasSuffix(u.Host, ":")`), or add a lowercase-host requirement to the rejection branch. Add two table rows to `TestLoad` (`http://LocalHost:5173`, `http://a.test:`). If the orchestrator prefers to keep D1's rule list frozen, record the skip with that justification; nothing downstream depends on it.

No other issues found.

## 4. Positive Observations

- **Plan fidelity.** Every binding decision is implemented as written: the `Config` shape and `Load(getenv)` signature (D1), the frozen domain types with prescriptive `Listing`/`ListedDate` doc comments per R1 (D2), `Pinger`/`Deps`/unexported `api` with a neutral extension comment and no ticket numbers in shipped code (D3, grep-verified), middleware order `requestID → logRequests → recoverPanic → cors → fallbacks` (D4), the `ServeMux.Handler` probe for 404/405 classification (D5), pre-encoded 500 fallback in `writeJSON` (D6), the 503 branch with a bounded ping (D7), the linear `run()` with synchronous `net.Listen`, `stop()` before drain, and `Shutdown`+`Close` on error (D8), the v2 lint config (D9), and the Makefile/.gitignore blocks (D10).
- **The 404/405 fallback is sound.** `fallbacks` only probes when `mux.Handler` returns an empty pattern, so the probed handler is always stdlib code with no side effects; matched requests go through `mux.ServeHTTP` so `r.Pattern` and path values are populated. I confirmed against the Go 1.27.1 source that the cleaned-path redirect is `StatusTemporaryRedirect`, so `TestNonCanonicalPathRedirects` asserts real behavior, and the live run showed `Allow: GET, HEAD` copied onto the 405 envelope.
- **Status tracking is careful.** `statusWriter` ignores 1xx when recording the final status, records implicit 200 on first `Write`, and exposes `Unwrap()` so `http.ResponseController` keeps working. `recoverPanic` re-panics `http.ErrAbortHandler`, only writes the envelope when nothing has been sent, and because it sits inside `logRequests` the request line correctly shows status 500 (tested).
- **Secrets stay out of responses.** `TestPanicRecovery` asserts the exact 500 body and scans it for `goroutine`, `.go:` and the panic text; `TestHealthz` asserts the ping error appears in the log but not the body. This matches §5.5's "details go to the log with the request ID, never to the client."
- **Tests verify behavior through the public surface.** All httpapi tests drive `NewRouter` with `httptest.NewRecorder`, decode envelopes with `DisallowUnknownFields`, parse log lines as JSON, and cover every row in the plan's test table plus useful extras (preflight to an unknown path, exact-origin match with a trailing slash, non-preflight OPTIONS → 405, `RequestIDFrom` on an empty context, `ErrAbortHandler` propagation). The single internal-package test (`respond_test.go`) is a justified, commented exception: no route can yet produce an unencodable value.
- **Config loader is pure and exhaustive.** `Load` takes a `getenv` function so tests never touch the process environment; every failing variable is reported at once via `errors.Join` with the variable name prefixed, and the live run showed exactly that in the fatal line.
- **Commit hygiene.** Three conventional commits matching the plan's split, no `.lattice/` files staged, `go.mod` pinned to `go 1.27` with no dependencies and no `go.sum`, working tree clean.
