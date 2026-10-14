# BC-2: Backend scaffold: Go 1.27 HTTP server with health, slog, request IDs, CORS, graceful shutdown, config, lint

Stand up the Go service skeleton that BC-3, BC-4 and BC-5 build inside. Spec: docs/SPEC.md §2, §3, §5.1, §5.5 (envelope), §5.6, §11.

Why: every backend task needs the module, the router, the middleware chain and the shared domain types to exist first. Defining the shared types here lets BC-3 (domain logic) and BC-4 (store) run in parallel.

Scope:
- backend/go.mod, module path `brillio/backend`, `go 1.27`. Stdlib only.
- cmd/api/main.go: load config, build router, http.Server with Read/Write/Idle timeouts, graceful shutdown on SIGINT/SIGTERM honoring SHUTDOWN_TIMEOUT, slog JSON logger to stdout at LOG_LEVEL.
- internal/config: parse §11 env vars with defaults; invalid LOG_LEVEL or SHUTDOWN_TIMEOUT is a startup error.
- internal/domain: types only. `Listing` per §4, `ListingStatus` with constants and `ParseStatus`, `SearchFilter{MinPrice, MaxPrice *float64; MinBedrooms *int; City, Keyword string; KeywordTokens []string; Statuses []ListingStatus}`. No logic.
- internal/httpapi: `NewRouter(deps Deps) http.Handler` using Go 1.22 ServeMux patterns. Middleware in order: request ID (echo X-Request-ID or generate 16 random bytes hex), request log line (method, path, status, duration, request id, remote addr), panic recovery to a 500 envelope, CORS per §5.6. JSON helpers `writeJSON` and `writeError` producing the §5.5 envelope. 404 and 405 handlers that return the envelope. GET /healthz returning {"status":"ok"} (DB ping arrives with BC-4; leave a `Pinger` interface in Deps that may be nil).
- backend/.golangci.yml: default linters plus errcheck, govet, staticcheck, revive.
- Root Makefile: fill in dev-api (go run ./cmd/api), and the backend halves of test (go test -race ./...), lint (golangci-lint run), check.
- Tests: config table; middleware (request id echoed and generated, panic becomes 500 envelope with no stack in body, CORS preflight 204 with the right headers, unknown route 404 envelope, wrong method 405 envelope).

Out of scope: SQLite, the search handler, OpenAPI, Docker.

Acceptance:
- `go run ./cmd/api` serves GET /healthz 200 with Content-Type application/json and an X-Request-ID header.
- `make lint` and `make test` pass for the backend.
- SIGTERM produces a "shutting down" log line and exit code 0.
- internal/domain compiles with the shared types and zero logic.

Validation: run the binary, curl /healthz with and without an X-Request-ID header, send SIGTERM, paste the output as evidence with --role validation.
