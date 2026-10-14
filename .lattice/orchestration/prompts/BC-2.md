# Ticket card — BC-2

```
TICKET=BC-2
TASK_UUID=task_01M4VXK222MWVR1FDJ20AS3NYR
TITLE=Backend scaffold: Go 1.27 HTTP server with health, slog, request IDs, CORS, graceful shutdown, config, lint
WORKTREE=/Users/forrest/Code/brillio-worktrees/bc-2-backend-scaffold
BRANCH=feat/BC-2-backend-scaffold
REVIEW_BASE=master
ANCHOR=none (based on master after BC-1)
PORTS=API 18002 (PORT=18002, DB not used yet)
AREA=area:11
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-2 notes (binding)

1. **Gate (run-state Gate rule):** BC-7 runs in parallel and owns the web halves. Your gate is
   `make lint-api test-api` green. Full `make check` may fail only on the BC-1 web stubs if BC-7 has not landed
   yet; if BC-7 has landed by the time you hand off, merge master and full `make check` must be green.
2. **Makefile:** fill only `dev-api`, `test-api` (`cd backend && go test -race ./...`), `lint-api`
   (`cd backend && golangci-lint run`). Do not touch the web targets or the `test`/`lint`/`check` composition
   that BC-1 created; BC-7 edits the web lines at the same time and the Orchestrator unions on landing.
3. **golangci-lint 2.14.0** is on the host. `.golangci.yml` uses the v2 schema (`version: "2"`), default
   linters plus errcheck, govet, staticcheck, revive.
4. **Shared types are yours to define** exactly as the ticket states (`Listing`, `ListingStatus` + `ParseStatus`,
   `SearchFilter`). BC-3 (domain logic) and BC-4 (store) branch from your branch the moment you reach
   `review` and build on these names — after `review` never rename or rebase; additive commits only.
5. `httpapi.Deps` carries a `Pinger` interface (may be nil) for BC-4, and leaves room for BC-5's `Searcher`,
   `Now func() time.Time` and cities source. Keep `NewRouter` the single place routes are registered so BC-5 and
   BC-6 add one line each.
6. Validation needs the real binary: `PORT=18002 go run ./cmd/api` (or the built binary), curl `/healthz`
   with and without `X-Request-ID`, a 404 and a 405 envelope, a CORS preflight, then SIGTERM and capture the
   "shutting down" line and exit code 0. Kill the process afterwards.
7. **Forwarded from BC-1 review (art_01M4VZQ546WYVMW5DWJH6HMBPN), two nits:** (a) a plain `go build ./cmd/api`
   inside `backend/` writes an unignored `backend/api` binary; add an anchored `.gitignore` rule for wherever you
   actually build (`/backend/bin/` is already covered — prefer `go build -o bin/api`). (b) repo `.gitignore`
   covers only `coverage.out` and `*.coverprofile`; add whatever filename your `test-api` recipe writes
   (`cover.out`, `coverage.html`) — the host's global `*.out` hides the gap locally but not in a fresh clone.
   Edit only the backend lines of `.gitignore`.
