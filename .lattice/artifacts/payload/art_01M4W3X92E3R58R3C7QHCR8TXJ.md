Lattice-Reviewed-Commit: 318ebc5005eb1f72b3b849a02ed9c6626c24def7
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store
Lattice-Reviewed-Base: master (7ab182d6817d3418d45ebe6dee860540e49f7730)
Lattice-Reviewed-Head: feat/BC-4-sqlite-store (318ebc5005eb1f72b3b849a02ed9c6626c24def7)

# Code Review: BC-4 — SQLite store (open, migrate, seed, search, cities, ping)

Reviewed: `feat/BC-4-sqlite-store` @ 318ebc5 against `master` @ 7ab182d, in worktree
`/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store`. Reviewer: agent:claude-fable-5.1-reviewer.

## 1. Verdict

**PASS**

## 2. Summary

The diff adds `internal/store` (open with DSN pragmas, embedded migrations, upsert, parameterized
search, cities, ping), `internal/store/seed` (strict JSON loader with per-record errors), the
`modernc.org/sqlite` dependency, startup wiring in `cmd/api/main.go`, a one-sentence SPEC §4.2 edit, and
~900 lines of tests. The implementation follows the plan closely, including every plan-review resolution
(BeginTx not literal BEGIN, `file:` rejection, index-set assertion, `.Equal` date comparison, VACUUM
caveat). All acceptance criteria hold: `go test -race ./...` passes 188 tests, `golangci-lint run` reports
0 issues, `go mod tidy` is a no-op, the SQL-concatenation grep shows only constant fragments, and a live run
seeds 12 listings on a fresh DB, keeps 12 after restart, and exits 1 on a bad seed without creating the DB
file. Only two minor hardening notes; nothing blocks advancing to `in_validation`.

### What was verified

| Check | Result |
|---|---|
| `go build ./... && go vet ./...` | clean |
| `go test -race -count=1 ./...` | 188 passed, 6 packages |
| `golangci-lint run` (v2.14.0, repo config) | 0 issues |
| `go mod tidy` | no change to go.mod / go.sum |
| `grep -rnE 'Sprintf\|\+ ' internal/store --include='*.go'` (non-test) | 6 hits: `buildSearch` joins constant fragments + `?` lists (listings.go:156,158); `RecordError.Error` message text (seed.go:67,69); `dsn()` appends constant pragma params to operator-config path (store.go:77,79). No caller value ever reaches SQL text. |
| Driver claims (modernc.org/sqlite v1.60.1 source) | `conn.go:61-73` splits a non-`file:` DSN at the first `?`; `conn.go:151` calls `applyQueryParams` inside `newConn`, so `_pragma` runs on every pooled connection. Matches plan decision 3 and `TestOpenPragmas`. |
| Live run 1 (fresh `DB_PATH`, port 18004) | log `"msg":"seeded 12 listings"`; `GET /healthz` → `200 {"status":"ok"}`; `select count(*) from listings` → 12; `schema_migrations` → `0001_listings.sql`; `pragma journal_mode` → `wal`; SIGTERM → `server stopped`, exit 0 |
| Live run 2 (same `DB_PATH`) | seeded 12 again; count still 12; one migration row; healthz 200 |
| Bad seed (`.[3].status = "gone"`) | exit 1; `"err":"seed: /tmp/...bad.json: record index 3: field \"status\": invalid listing status: \"gone\""`; no DB file created |
| §6.1 filters each tested alone and combined | `TestSearchFilters` (31 cases), `TestSearchLiteralWildcards`, `TestSearchInjection`; expected ID sets hand-checked against `data/sample_listings.json` |

## 3. Issues

**[minor] backend/internal/store/migrate.go:46 — Applied-migration set is read outside the migration transaction, so two processes bootstrapping the same new file race.**
`appliedMigrations` runs before any `BeginTx`. If two API processes open the same brand-new file concurrently
(e.g. two containers on one volume), both see nothing applied; the second blocks on the `_txlock=immediate`
BEGIN until the first commits, then its `CREATE TABLE schema_migrations` fails with "table already exists"
and the process exits 1. v1 runs a single API process, so this is theoretical today, and a restart fixes it.
**Fix:** Inside `applyMigration`, after `BeginTx`, re-check `SELECT count(*) FROM schema_migrations WHERE filename = ?`
(guarded by the sqlite_master existence check) and return nil if already present; or, simpler, make
`0001_listings.sql` use `CREATE TABLE IF NOT EXISTS` for `schema_migrations` and `INSERT OR IGNORE` for the
bookkeeping row. Optional for BC-4; a one-line note in the Risks section of the plan is acceptable instead.

**[minor] backend/internal/store/listings.go:72 — `ListedDate.Format` uses the Time's own location.**
`l.ListedDate.Format(domain.DateLayout)` formats the wall-clock date in whatever location the `time.Time`
carries. `seed.Load` always produces UTC midnight so the stored text is correct today, but a `Listing`
built elsewhere (BC-3/BC-5 tests, future importers) with a non-UTC location near midnight would write a
different calendar date than its UTC instant, and `scanListing` reads it back as UTC. Cheap to make robust.
**Fix:** `l.ListedDate.UTC().Format(domain.DateLayout)`.

Notes, no action required:
- `store.go:59` — `:memory:` runs on one connection, so a `/healthz` Ping issued during an in-flight
  `Search` waits for it (bounded by the handler's 2s timeout). `:memory:` is dev/test only; the plan's Risks
  section already records the single-connection constraint.
- `cmd/api/main.go:69` — `Upsert` runs on the signal context, so a SIGTERM during seeding exits 1 with
  `seed: ... context canceled` rather than a clean shutdown. Startup was interrupted, so a non-zero exit is
  defensible.
- `seed.go:110-114` — `errors.AsType` is Go 1.26+; `go.mod` says 1.27, fine.

## 4. Positive Observations

- **Parameterization is enforced, not just intended.** Every SQL fragment is a named constant; `buildSearch`
  only joins constants and `?` lists. `TestBuildSearch` pins the exact query text for the all-filters case
  and asserts no filter value substring appears in it; `TestSearchInjection` exercises city, token and status
  payloads against the real engine and confirms the table is intact afterwards.
- **Pragmas hold on every pooled connection** by riding the DSN, and the test proves it by holding two
  `*sql.Conn`s simultaneously and checking `busy_timeout`, `foreign_keys` and `journal_mode` on each. The
  `_txlock=immediate` choice to avoid WAL lock-upgrade `SQLITE_BUSY` is explained in a comment.
- **Single-connection `:memory:` is handled carefully**: `appliedMigrations` fully drains and closes its rows
  before any `BeginTx`, and the comment says why. `Search`/`Cities` never issue nested queries.
- **Migrations are genuinely idempotent and atomic per file.** `TestMigrateOrderAndRollback` uses a
  `fstest.MapFS` with a two-statement failing file to prove the first statement rolled back, that earlier
  files stay applied, that non-`.sql` and subdirectory entries are skipped, and that a rerun applies only the
  fixed file. `TestMigrateTwice` asserts the index set by name rather than count (plan-review item 3).
- **Seed loader is precise.** Pointer DTO distinguishes missing/null from zero; `DisallowUnknownFields`;
  `*json.UnmarshalTypeError` mapped to the field name; presence checked in SPEC field order; first-seen
  `(source, id)` duplicate detection names the earlier index. `TestLoadRecordErrors` covers 18 failure
  shapes with `errors.As`/`errors.Is`, and `TestLoadStatusMessage` pins the exact log text the validation
  step greps for.
- **Startup ordering is right**: `seed.Load` before `store.Open` means a bad file never creates a DB, which
  the smoke run confirmed. The "seeded N listings" message is both a literal and structured attrs.
- **errcheck-clean resource handling**: `defer func() { _ = rows.Close() }()` with `rows.Err()` checked,
  `Rollback` deferred and harmless after `Commit`, `PingContext` failure closes the DB before returning.
- **Cities query** documents the `min(rowid)` first-seen trick and its VACUUM caveat inline, and the test
  table covers first-seen casing in both directions, case-insensitive sort, empty-city exclusion and the
  empty store.
- **Scope discipline**: no `Deps` changes beyond `Pinger`, `domain` untouched, Makefile untouched, SPEC edit
  is the single agreed sentence. `go.mod`/`go.sum` are tidy.

Routing recommendation: advance `review -> in_validation`. The two minor items can be fixed inline in
validation or deferred with a comment; neither warrants a rework cycle.
