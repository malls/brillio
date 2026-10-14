# BC-2 Validation: Backend scaffold (Go HTTP server, healthz, slog, request IDs, CORS, graceful shutdown, config, lint)

- **Validator:** agent:opus-5.5-validator
- **Worktree:** `/Users/forrest/Code/brillio-worktrees/bc-2-backend-scaffold`, branch `feat/BC-2-backend-scaffold`
- **HEAD:** `a9b7f73ada5002f000ffe209da79af8fdb7756a3` (verified with `git rev-parse HEAD`; working tree clean)
- **Date/time:** 2026-10-13 22:05–22:07 EDT
- **Toolchain:**
  - `go version` → `go version go1.27.1 darwin/arm64`
  - `golangci-lint --version` → `golangci-lint has version 2.14.0 built with go1.27.1 from 114493f on 2026-09-24T10:54:53Z`
- **Port:** 18002 (API, `PORT=18000+N`). `lsof -nP -iTCP:18002 -sTCP:LISTEN` returned nothing (exit 1) before the run and after every run.
- **Scratch logs:** `/private/tmp/claude-503/bc2-validate/` (`gate.log`, `test-fresh.log`, `check.log`, `gorun.log`, `api.log`)

---

## 1. Gate: `make lint-api test-api`, `make check`, domain package

### 1a. `make lint-api test-api`: PASS

```
$ make lint-api test-api; echo EXIT=$?
cd backend && golangci-lint run
0 issues.
cd backend && go test -race ./...
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	(cached)
ok  	brillio/backend/internal/domain	(cached)
ok  	brillio/backend/internal/httpapi	(cached)
EXIT=0
```

The results came from the test cache, so I also ran the tests uncached:

```
$ (cd backend && go test -race -count=1 ./...); echo EXIT=$?
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	1.285s
ok  	brillio/backend/internal/domain	1.224s
ok  	brillio/backend/internal/httpapi	1.264s
EXIT=0
```

### 1b. `make check`: PASS (fails only on the BC-1 web stub, as allowed)

`check: lint test`, `lint: lint-api lint-web`. Make stops at the first failing prerequisite, which is the web stub. The backend half (`lint-api`) runs first and passes.

```
$ make check; echo CHECK_EXIT=$?
cd backend && golangci-lint run
0 issues.
not yet implemented: lint-web (lands in BC-7)
make: *** [lint-web] Error 1
CHECK_EXIT=2
```

### 1c. `internal/domain` compiles, holds types only: PASS

```
$ cd backend && go build ./internal/domain && go vet ./internal/domain; echo BUILD_VET_EXIT=$?
BUILD_VET_EXIT=0

$ go list -f '{{join .Imports " "}}' ./internal/domain
errors fmt slices time

$ grep -n '^func ' internal/domain/*.go | grep -v _test.go
internal/domain/listing.go:37:func AllStatuses() []ListingStatus {
internal/domain/listing.go:44:func ParseStatus(s string) (ListingStatus, error) {
```

Exported identifiers from `go doc -all ./internal/domain` (doc comments trimmed):

```
package domain // import "brillio/backend/internal/domain"

CONSTANTS
const DateLayout = "2006-01-02"

VARIABLES
var ErrInvalidStatus = errors.New("invalid listing status")

TYPES
type Listing struct { ID, Source, Address, City, State, Zip string; Price float64;
    Bedrooms int; Bathrooms float64; Sqft int; Latitude, Longitude float64;
    ListedDate time.Time; Status ListingStatus; Description string }   // json tags on all fields
type ListingStatus string
const ( StatusActive ListingStatus = "active"; StatusPending = "pending"; StatusSold = "sold" )
func AllStatuses() []ListingStatus
func ParseStatus(s string) (ListingStatus, error)
type SearchFilter struct { MinPrice, MaxPrice *float64; MinBedrooms *int; City, Keyword string;
    KeywordTokens []string; Statuses []ListingStatus }
```

The package declares only types, constants and a sentinel error. Its only funcs are `AllStatuses` and `ParseStatus`, with no methods and no I/O. Imports are standard library only.

Minor note (not a failure): the package doc comment says the package holds "core types and its pure business logic". That wording anticipates later tickets. There is no business logic in it today.

---

## 2. `go run ./cmd/api` acceptance: PASS

Command (from `backend/`): `PORT=18002 go run ./cmd/api > gorun.log 2>&1 &`, then polled `curl -s localhost:18002/healthz` until it answered (ready after about 2 s).

```
$ curl -si http://localhost:18002/healthz
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 0b4390f678a9e0cc1fe62055d12dc5b0
Date: Wed, 14 Oct 2026 02:06:15 GMT
Content-Length: 16

{"status":"ok"}
```

Stop: `go run` (pid 75196) had spawned the child `/var/folders/.../go-build3924650387/b001/exe/api` (pid 75245, ppid 75196), which was the listener on 18002. I sent SIGTERM to the child. The child shut down gracefully and the `go run` parent exited by itself, so no extra kill was needed. Afterwards `lsof` found no listener (exit 1) and `pgrep -fl 'go run ./cmd/api|cmd/api|bin/api'` found no process (exit 1).

`gorun.log`:
```
{"time":"2026-10-13T22:06:14.926197-04:00","level":"INFO","msg":"listening","addr":"[::]:18002","log_level":"INFO","cors_origins":["http://localhost:5173"],"shutdown_timeout":"10s"}
{"time":"2026-10-13T22:06:15.420668-04:00","level":"INFO","msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.092,"request_id":"ba2a957f0146f82d637d8c2b483d9f43","remote_addr":"[::1]:62417"}
{"time":"2026-10-13T22:06:15.430947-04:00","level":"INFO","msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.055,"request_id":"0b4390f678a9e0cc1fe62055d12dc5b0","remote_addr":"[::1]:62418"}
{"time":"2026-10-13T22:06:15.606569-04:00","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-13T22:06:15.606652-04:00","level":"INFO","msg":"server stopped"}
```
(The first request line is the readiness poll.)

---

## 3. Built binary run: the authoritative check

Build: `cd backend && go build -o bin/api ./cmd/api` printed `BUILD_OK`. `git check-ignore -v backend/bin/api` returned `.gitignore:16:/backend/bin/`, so the binary is ignored.
Run (repo root): `PORT=18002 ./backend/bin/api > api.log 2>&1 & PID=$!` (PID 76335), then polled until ready.

### 3A. GET /healthz without X-Request-ID: PASS
Status 200, `Content-Type: application/json; charset=utf-8`, a generated 32-hex `X-Request-Id`, body `{"status":"ok"}`.
```
$ curl -si http://localhost:18002/healthz
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 5e50017108c01c304856ccd550baa1c0
Date: Wed, 14 Oct 2026 02:06:26 GMT
Content-Length: 16

{"status":"ok"}
```

### 3B. GET /healthz with X-Request-ID: PASS (ID echoed)
```
$ curl -si -H 'X-Request-ID: bc2-validation-abc123' http://localhost:18002/healthz
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: bc2-validation-abc123
Date: Wed, 14 Oct 2026 02:06:26 GMT
Content-Length: 16

{"status":"ok"}
```

### 3C. 404 envelope: PASS
```
$ curl -si http://localhost:18002/nope
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: bc99ab4c812c9fa63da1afda160a567d
Date: Wed, 14 Oct 2026 02:06:26 GMT
Content-Length: 63

{"error":{"code":"not_found","message":"Resource not found."}}
```

### 3D. 405 envelope: PASS (`Allow: GET, HEAD`)
```
$ curl -si -X POST http://localhost:18002/healthz
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: b31a6eaee3dd97d7dc2d386e6e056567
Date: Wed, 14 Oct 2026 02:06:26 GMT
Content-Length: 72

{"error":{"code":"method_not_allowed","message":"Method not allowed."}}
```

### 3E. CORS preflight from the allowed origin: PASS
```
$ curl -si -X OPTIONS -H 'Origin: http://localhost:5173' -H 'Access-Control-Request-Method: GET' -H 'Access-Control-Request-Headers: X-Request-ID' http://localhost:18002/healthz
HTTP/1.1 204 No Content
Access-Control-Allow-Headers: Content-Type, X-Request-ID
Access-Control-Allow-Methods: GET, OPTIONS
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Expose-Headers: X-Request-ID
Vary: Origin
X-Request-Id: 5a771a51af1b137b84ec92773749c30a
Date: Wed, 14 Oct 2026 02:06:26 GMT
```

### 3F. CORS preflight from a disallowed origin: PASS (no Access-Control-Allow-* headers)
```
$ curl -si -X OPTIONS -H 'Origin: http://evil.test' -H 'Access-Control-Request-Method: GET' -H 'Access-Control-Request-Headers: X-Request-ID' http://localhost:18002/healthz
HTTP/1.1 204 No Content
Vary: Origin
X-Request-Id: c4befff10c37f8eea1f33e3597257a3f
Date: Wed, 14 Oct 2026 02:06:26 GMT
```

### 3G. Simple GET from the allowed origin: PASS (ACAO and Expose-Headers present)
```
$ curl -si -H 'Origin: http://localhost:5173' http://localhost:18002/healthz
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Expose-Headers: X-Request-ID
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 819506297e6c71909059b07615f6937d
Date: Wed, 14 Oct 2026 02:06:26 GMT
Content-Length: 16

{"status":"ok"}
```

### 3H. SIGTERM: PASS (exit 0, "shutting down" logged)
```
$ kill -TERM $PID; wait $PID; echo "exit=$?"
exit=0
$ lsof -nP -iTCP:18002 -sTCP:LISTEN; echo "post_lsof_exit=$?"
post_lsof_exit=1
```

### 3I. Structured log (`api.log`): PASS
The log has one JSON line per request, each with method, path, status, duration_ms, request_id and remote_addr. The request IDs match the response headers above, including the echoed `bc2-validation-abc123`. The run ends with `"msg":"shutting down"` and `"msg":"server stopped"`. The first request line is the readiness poll.
```
{"time":"2026-10-13T22:06:25.712352-04:00","level":"INFO","msg":"listening","addr":"[::]:18002","log_level":"INFO","cors_origins":["http://localhost:5173"],"shutdown_timeout":"10s"}
{"time":"2026-10-13T22:06:26.082481-04:00","level":"INFO","msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.087,"request_id":"33cd4371333693f7f72b9649c5f34ec2","remote_addr":"[::1]:62430"}
{"time":"2026-10-13T22:06:26.09378-04:00","level":"INFO","msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.05,"request_id":"5e50017108c01c304856ccd550baa1c0","remote_addr":"[::1]:62431"}
{"time":"2026-10-13T22:06:26.104301-04:00","level":"INFO","msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.022,"request_id":"bc2-validation-abc123","remote_addr":"[::1]:62432"}
{"time":"2026-10-13T22:06:26.114704-04:00","level":"INFO","msg":"request","method":"GET","path":"/nope","status":404,"duration_ms":0.056,"request_id":"bc99ab4c812c9fa63da1afda160a567d","remote_addr":"[::1]:62433"}
{"time":"2026-10-13T22:06:26.125339-04:00","level":"INFO","msg":"request","method":"POST","path":"/healthz","status":405,"duration_ms":0.023,"request_id":"b31a6eaee3dd97d7dc2d386e6e056567","remote_addr":"[::1]:62434"}
{"time":"2026-10-13T22:06:26.137188-04:00","level":"INFO","msg":"request","method":"OPTIONS","path":"/healthz","status":204,"duration_ms":0.041,"request_id":"5a771a51af1b137b84ec92773749c30a","remote_addr":"[::1]:62435"}
{"time":"2026-10-13T22:06:26.148303-04:00","level":"INFO","msg":"request","method":"OPTIONS","path":"/healthz","status":204,"duration_ms":0.009,"request_id":"c4befff10c37f8eea1f33e3597257a3f","remote_addr":"[::1]:62436"}
{"time":"2026-10-13T22:06:26.160619-04:00","level":"INFO","msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.023,"request_id":"819506297e6c71909059b07615f6937d","remote_addr":"[::1]:62437"}
{"time":"2026-10-13T22:06:26.161806-04:00","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-13T22:06:26.161865-04:00","level":"INFO","msg":"server stopped"}
```

---

## 4. Config error path: PASS

One fatal JSON line names both bad variables and the process exits non-zero (1). Nothing was left listening.
```
$ LOG_LEVEL=verbose SHUTDOWN_TIMEOUT=abc PORT=18002 ./backend/bin/api; echo "exit=$?"
{"time":"2026-10-13T22:06:29.856267-04:00","level":"ERROR","msg":"fatal","err":"config: LOG_LEVEL: \"verbose\" is not one of debug|info|warn|error\nSHUTDOWN_TIMEOUT: \"abc\" is not a Go duration such as 10s"}
exit=1
$ lsof -nP -iTCP:18002 -sTCP:LISTEN; echo "lsof_exit=$?"
lsof_exit=1
```

---

## 5. Cleanup: PASS

```
$ lsof -nP -iTCP:18002 -sTCP:LISTEN; echo "lsof_exit=$?"
lsof_exit=1
$ pgrep -fl 'bin/api|cmd/api'; echo "pgrep_exit=$?"
pgrep_exit=1
```
Port 18002 is free and none of the validator's `bin/api` or `go run` processes remain. The only file I created in the worktree is the git-ignored `backend/bin/api`.

---

## Acceptance mapping

| Acceptance bullet | Evidence | Result |
|---|---|---|
| `go run ./cmd/api` serves GET /healthz 200 with `Content-Type: application/json` and an `X-Request-ID` header | §2 (go run), §3A (built binary) | PASS |
| `make lint` and `make test` pass for the backend (gate `make lint-api test-api`; `make check` may fail only on the web stub) | §1a, §1b | PASS |
| SIGTERM produces a "shutting down" log line and exit code 0 | §3H, §3I (also §2 for go run) | PASS |
| `internal/domain` compiles with the shared types and zero logic | §1c | PASS |
| Validation block: run binary, curl /healthz with and without X-Request-ID, send SIGTERM, paste output | §3A, §3B, §3H, §3I | PASS |
| Card addition: 404 envelope | §3C | PASS |
| Card addition: 405 envelope | §3D | PASS |
| Card addition: CORS preflight (allowed origin, disallowed origin, simple GET) | §3E, §3F, §3G | PASS |
| Config error path (fatal line naming both variables, non-zero exit) | §4 | PASS |
| Port 18002 free and no processes left | §5 | PASS |

Overall: PASS
