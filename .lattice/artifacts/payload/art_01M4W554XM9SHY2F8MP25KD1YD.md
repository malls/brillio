Lattice-Reviewed-Commit: 9daad85836b1fe6b517b2c6bc5e76ea3f9a85709
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store
Lattice-Reviewed-Base: master (c91fa7f79dfce9029dcc63f0db6b9dc4768fa6b4)
Lattice-Reviewed-Head: feat/BC-4-sqlite-store (9daad85836b1fe6b517b2c6bc5e76ea3f9a85709)

# Code Review: BC-4 — SQLite store (open, migrate, seed, search, cities, ping)

Reviewed commit: 9daad85836b1fe6b517b2c6bc5e76ea3f9a85709 (`feat/BC-4-sqlite-store`) against
`master` @ c91fa7f, in worktree `/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store`.
Reviewer: agent:claude-fable-5.1-reviewer (independent, cold context).

## 1. Verdict

**PASS**

The implementation matches the plan (including all seven plan-review resolutions) and every
acceptance criterion on the ticket. The two findings below are minor and non-blocking; neither
requires rework before `in_validation`.

## 2. Summary

Reviewed the new `internal/store` and `internal/store/seed` packages, the embedded migration,
the `main.go` wiring, the real-store `/healthz` test and the one-line SPEC §4.2 edit. Quality
is high: SQL is assembled only from constants and `?` lists, every DB resource is closed with
`rows.Err()` checked, the `:memory:` single-connection deadlock hazard is handled, and the test
suite covers every §6.1 filter alone and combined plus the literal-wildcard, injection, casing
and idempotency cases the ticket asks for. All gates verified green by me, not taken from the
diff: `go build`, `go vet`, `go test -race ./...`, `golangci-lint run` (0 issues),
`go mod tidy -diff` (no changes), and a native end-to-end run of the binary.

Verification results:

| Check | Result |
|---|---|
| `go build ./... && go vet ./...` | OK |
| `go test -race -count=1 ./...` | all packages pass |
| `make lint-api` (golangci-lint 2.14.0) | 0 issues |
| `go mod tidy -diff` | no diff |
| grep `Sprintf\|+` in non-test store files | only constant fragments, `?` lists and the DSN path concat in `dsn()` |
| Run 1, fresh `DB_PATH` | `"msg":"seeded 12 listings"`, `/healthz` 200 `{"status":"ok"}` with `X-Request-Id`, `count(*)` = 12, `schema_migrations` = `0001_listings.sql`, `journal_mode` = wal, clean `server stopped` on SIGTERM |
| Run 2, same `DB_PATH` | identical: still 12 rows, one migration row, healthz 200 |
| Bad seed (`.[3].status = "gone"`) | exit 1, `fatal` line `seed: <path>: record index 3: field "status": invalid listing status: "gone"`, no DB file created |

Driver claims in the plan were checked against `modernc.org/sqlite@v1.60.1` source: `newConn`
splits the DSN at the first `?` and strips the query for non-`file:` names (`conn.go:62-72`),
`_pragma`/`_txlock` are applied per connection, and `conn.Ping` exists, so `db.PingContext`
is a real round-trip rather than a pool check.

## 3. Issues

**[MINOR] backend/internal/store/migrate.go:46 — Applied set is read outside the migration transaction**
`migrate` reads `schema_migrations` once, then opens a `BEGIN IMMEDIATE` transaction per pending
file. If two processes ever migrate the same file concurrently (two containers on one volume),
both see "nothing applied"; the second waits on the first's write lock, then its
`CREATE TABLE schema_migrations` fails with "table already exists" and that process exits 1 at
startup. v1 runs a single `api` container and the plan does not require multi-instance safety,
so this is informational, not a defect in scope.
**Fix:** None needed for BC-4. If multi-instance ever matters, re-read the applied set inside
each transaction (after `BeginTx`, which already serializes writers via `_txlock=immediate`)
and skip the file when its name is present. Suggested routing: ignore with justification.

**[MINOR] backend/internal/store/store_test.go:106 and :355 — Re-seed does not assert Cities order is preserved**
The ticket's motivation is idempotent seeding on every container restart. `Cities` relies on
`ON CONFLICT DO UPDATE` keeping the original rowid (plan decision 11, comment at
`listings.go:49`). `TestSeedTwiceIdempotent` and `TestReopenFile` prove the row count after a
second seed but never call `Cities`, so a future change to `INSERT OR REPLACE` (which deletes
and re-inserts, assigning new rowids) would silently reorder display casing across restarts
without failing a test. `TestCities` only covers casing variants added as new rows.
**Fix:** One assertion: after the second `upsert` in `TestSeedTwiceIdempotent` (or after
`upsert(t, second, listings...)` at `store_test.go:372`), assert `s.Cities()` equals
`["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]`. Cheap to fix inline;
also acceptable to skip with a comment.

No critical or major issues found.

Notes with no action required:
- `lower()` and `COLLATE NOCASE` fold ASCII only, while `strings.ToLower` on tokens folds
  Unicode. A description "CAFÉ" will not match token "café". The plan lists this as a known
  risk for the BC-15 README; the sample data is ASCII.
- A SIGTERM arriving during the millisecond-long seed is reported as `fatal` (context canceled)
  with exit 1 rather than a clean stop. Harmless at this data size.

## 4. Positive Observations

- **Parameterization is structural, not incidental.** Every SQL fragment in `listings.go` is a
  named constant; `buildSearch` only joins constants and `?` lists. `TestBuildSearch` pins the
  exact query text and additionally asserts that no filter value appears in the SQL string,
  which guards the acceptance criterion directly rather than by convention.
- **All seven plan-review resolutions landed verbatim:** `BeginTx` instead of SQL `BEGIN`
  (`migrate.go:99`, `listings.go` Upsert), the raw-`*sql.DB` MapFS fixture for
  `TestMigrateOrderAndRollback`, index assertion by name rather than count, `file:` prefix
  rejection with a T11 case, the VACUUM comment at the Cities query, field-by-field
  `ListedDate.Equal` comparison via `sameListing`, and the deferred RecordError enrichment.
- **The `:memory:` single-connection hazard is designed around**, not discovered later:
  `appliedMigrations` fully drains and closes its rows before any `BeginTx`, `Search` and
  `Cities` never issue a nested query while iterating, and the constraint is documented at
  `Open` and `appliedMigrations`.
- **Seed loader error model is precise.** Pointer DTO fields distinguish missing/null from zero,
  `null` elements are caught before decode (which would otherwise silently succeed),
  `UnmarshalTypeError.Field` is mapped to the JSON name, `errors.Is(err, domain.ErrInvalidStatus)`
  works through the `RecordError.Unwrap` chain, and `TestLoadStatusMessage` pins the exact
  startup log text the validation step greps for.
- **Startup ordering is deliberate:** `seed.Load` runs before `store.Open`, so a bad seed file
  never creates a database file. Verified end-to-end above (`test ! -e` on the bad DB path).
- **Test coverage maps one-to-one onto the ticket and §8.** Every §6.1 filter alone and in
  combination, `%`/`_` literal via a listing that `LIKE` would mis-match (`pet_` vs `pets`),
  four injection shapes with a post-check that all 12 rows survive, city trim and case, status
  sets including `nil`, idempotent seed, file reopen, pragma presence on two simultaneous
  pooled connections, and per-file migration rollback with a second-pass rerun.
- **Lint discipline matches the repo's no-exclusions config**: `defer func() { _ = x.Close() }()`
  with `rows.Err()` after every loop, doc comments on every exported identifier, and
  `go mod tidy` leaves no diff.
- **Scope held.** No `Deps` fields beyond `Pinger`, `domain` untouched, Makefile untouched,
  SPEC edit confined to the one agreed §4.2 sentence.

Recommended routing for the orchestrator: advance to `in_validation`. Issue 2 is a one-line
inline fix if desired; issue 1 is an ignore-with-justification comment.
