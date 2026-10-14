Lattice-Reviewed-Commit: 532996fd5caa87fc8eb96868d7d896ecb413b5a5
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store
Lattice-Reviewed-Base: master (83952b0a8cb700f40fcc2dc50f785784215cd561)
Lattice-Reviewed-Head: feat/BC-4-sqlite-store (532996fd5caa87fc8eb96868d7d896ecb413b5a5)

Lattice-Reviewed-Commit: 532996fd5caa87fc8eb96868d7d896ecb413b5a5
Lattice-Reviewed-Worktree: /Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store
Lattice-Reviewed-Base: master (83952b0a8cb700f40fcc2dc50f785784215cd561)
Lattice-Reviewed-Head: feat/BC-4-sqlite-store (532996fd5caa87fc8eb96868d7d896ecb413b5a5)

# Code Review: BC-4 — SQLite store (open, migrate, seed, search, cities, ping)

Reviewed: `feat/BC-4-sqlite-store` @ 532996f against `master` @ 83952b0, in worktree
`/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store`. Reviewer: agent:claude-fable-5.1-reviewer
(fresh context; did not write this code).

## 1. Verdict

**PASS**

## 2. Summary

The diff adds `internal/store` (open with DSN-carried pragmas, embedded migrations, upsert, parameterized
search, cities, ping), `internal/store/seed` (strict JSON loader with indexed record errors), the
`modernc.org/sqlite` dependency, startup wiring in `cmd/api/main.go`, one agreed SPEC §4.2 sentence, and
about 950 lines of tests. Every acceptance criterion holds under independent verification: 188 tests pass
under `-race`, lint is clean, `go mod tidy` is a no-op, the SQL-text grep shows only constant fragments, and
a live run seeds 12 listings on a fresh DB, keeps 12 after a restart, and exits 1 on a bad seed without
creating the DB file. The backend code at this head is byte-identical to the code passed at 318ebc5
(`git diff 318ebc5..532996f -- backend` is empty; the merge from master brought only frontend files and a
non-BC-4 SPEC paragraph). No new blocking findings; the two minor hardening items from the prior review are
already tracked in BC-18.

### What was verified

| Check | Result |
|---|---|
| `go build ./... && go vet ./...` | clean |
| `go test -race -count=1 ./...` | 188 passed, 6 packages |
| `make lint-api` (golangci-lint 2.14.0, repo `.golangci.yml`) | 0 issues |
| `go mod tidy` | no change to go.mod / go.sum |
| `grep -rnE 'Sprintf\|\+' internal/store --include='*.go'` (non-test) | 8 hits, all benign: `selectListings` constant split across lines (listings.go:26); `buildSearch` joins constants and `?` lists (listings.go:151,156,158); `RecordError.Error` message text, not SQL (seed.go:67,69); `dsn()` appends constant pragma params to the operator-config path, which `Open` has already rejected if it contains `?` or starts with `file:` (store.go:77,79). No caller value ever reaches SQL text. |
| Live run 1, fresh `DB_PATH`, port 18944 | `"msg":"seeded 12 listings"`; `GET /healthz` → 200 `{"status":"ok"}` with `X-Request-Id`; `select count(*) from listings` → 12; `schema_migrations` → `0001_listings.sql`; `pragma journal_mode` → `wal`; SIGTERM → `server stopped` |
| Live run 2, same `DB_PATH` | seeded 12 again; count still 12; one `schema_migrations` row; healthz 200 |
| Bad seed (`jq '.[3].status = "gone"'`) | exit 1; `"err":"seed: /tmp/...bad.json: record index 3: field \"status\": invalid listing status: \"gone\""`; DB file not created |
| Expected ID sets in `TestSearchFilters` | hand-checked against `data/sample_listings.json` for minPrice, maxPrice, city, `pets` token (7 rows: A3 A4 A5 A6 A7 B7 B9; A1 "Pet friendly" and B11 "Pet restrictions" correctly excluded), status sets |
| Edge probes (throwaway test, removed) | `Upsert(nil)` and `Upsert([]domain.Listing{})` → `(0, nil)`; zero `ListedDate` round-trips as `0001-01-01` UTC |
| Prior review routing | art_01M4W3X92E3R58R3C7QHCR8TXJ PASS at 318ebc5; its two minor items are ticketed as BC-18 and recorded in a BC-4 comment |

## 3. Issues

No new issues that warrant a rework cycle. For completeness and routing, the two prior minor findings
remain in the code and are listed here so the orchestrator can confirm they are tracked, followed by
three new no-action notes.

**[minor] backend/internal/store/listings.go:72 — `ListedDate.Format` uses the Time's own location (already tracked: BC-18).**
`l.ListedDate.Format(domain.DateLayout)` writes the wall-clock date of whatever location the `time.Time`
carries. `seed.Load` always yields UTC midnight (verified in `TestLoadSample`), so production is correct,
and `scanListing` reads back as UTC, which matches the BC-3 note that `domain.AgeDays` needs UTC midnight.
A `Listing` built elsewhere with a non-UTC location near midnight would store a shifted calendar date.
**Fix:** `l.ListedDate.UTC().Format(domain.DateLayout)`. One token; BC-18 owns it. No action for BC-4.

**[minor] backend/internal/store/migrate.go:46 — Applied-migration set is read outside the migration transaction (already tracked: BC-18).**
Two processes bootstrapping the same brand-new file would both see nothing applied; the second blocks on
the `_txlock=immediate` BEGIN, then fails on `CREATE TABLE schema_migrations` and exits 1. v1 runs one API
process and a restart recovers. **Fix:** re-check inside the transaction, or `CREATE TABLE IF NOT EXISTS` /
`INSERT OR IGNORE` for the bookkeeping. BC-18 owns it. No action for BC-4.

Notes, no action required:

- `backend/internal/store/listings.go:46` — `selectCities` excludes `city <> ''` but not a whitespace-only
  city, and `whereCity` trims the filter input but naturally not stored values. Both are only reachable by a
  caller that bypasses `seed`, which collapses whitespace in `city` before any row is written; the store
  trusts normalized input by design (the seed loader is the validation boundary). A probe confirmed a
  `"   "` city written directly does appear in `Cities`. Not worth a `trim()` in SQL for v1.
- `backend/internal/store/listings.go:33` — SQLite's built-in `lower()` folds ASCII only while Go's
  `strings.ToLower` on the token is Unicode-aware, so a description containing a non-ASCII uppercase letter
  (e.g. `CAFÉ`) will not match the token `café`. Fine for the sample data; already recorded in the plan's
  Risks section for the BC-15 README.
- `docs/SPEC.md` §4.1 says migrations run "inside a transaction on open"; the implementation uses one
  transaction per file (plan decision 4, proven by `TestMigrateOrderAndRollback`). That is the stronger,
  standard reading and matches the ticket text. No SPEC edit needed.
- Test coverage gap, trivial: `Upsert` with an empty slice is not covered. The probe shows it returns
  `(0, nil)` as the doc comment implies.

## 4. Positive Observations

- **Parameterization is proven, not promised.** Every SQL fragment is a named constant; `buildSearch` only
  joins constants and `?` lists. `TestBuildSearch` pins the exact all-filters query text, asserts no filter
  value substring appears in it, and `TestSearchInjection` fires city, token, DROP TABLE and status payloads
  at the real engine and confirms the row count afterwards.
- **Pragmas ride the DSN** so they apply to every pooled connection, and `TestOpenPragmas` proves it by
  holding two `*sql.Conn`s at once and checking `busy_timeout`, `foreign_keys` and `journal_mode` on each.
  The `_txlock=immediate` rationale (WAL lock-upgrade `SQLITE_BUSY` that `busy_timeout` cannot retry) is in
  a comment where the next reader will find it.
- **The single-connection `:memory:` constraint is respected everywhere it matters.** `appliedMigrations`
  drains and closes its rows before any `BeginTx` and says why; `Search` and `Cities` never nest a query
  inside row iteration.
- **Migrations are idempotent and atomic per file.** `TestMigrateOrderAndRollback` uses a `fstest.MapFS`
  with a two-statement failing file to prove the first statement rolled back, earlier files stay applied,
  non-`.sql` and subdirectory entries are skipped, and a rerun applies only the fixed file. `TestMigrateTwice`
  asserts the four indexes by name (plan-review item 3).
- **The seed loader is exact.** Pointer DTO distinguishes missing or null from zero; `DisallowUnknownFields`;
  `*json.UnmarshalTypeError` mapped to the JSON field name; presence checked in SPEC field order; `(source, id)`
  duplicates name the earlier index. `TestLoadRecordErrors` covers 18 failure shapes through `errors.As` and
  `errors.Is`, and `TestLoadStatusMessage` pins the exact text the validation step greps for.
- **Startup ordering is right.** `seed.Load` runs before `store.Open`, so a bad file never creates a DB; the
  smoke run confirmed no file was left behind. The "seeded N listings" line is both a literal message and
  structured attributes.
- **errcheck-clean resource handling** throughout: `rows.Close()` deferred with `rows.Err()` checked,
  `Rollback` deferred and harmless after `Commit`, `PingContext` and `Migrate` failures close the DB before
  returning, tests close through `t.Cleanup` and report close errors.
- **Cities first-seen casing** is implemented with a documented `min(rowid)` trick, its VACUUM caveat is
  inline, and the test table covers first-seen casing in both directions, case-insensitive sort,
  empty-city exclusion and the empty store.
- **Scope discipline.** `Deps` gains only `Pinger`, `domain` and the Makefile are untouched, the SPEC edit is
  the single agreed sentence, and go.mod/go.sum are tidy. The compile-time check
  `var _ httpapi.Pinger = (*store.Store)(nil)` lives in the httpapi test package, where the contract is.

Routing recommendation: the task is already `in_validation` with validation evidence attached at 318ebc5.
Since the backend code is unchanged at 532996f, no rework is needed; proceed to `pr_open` once the
validator confirms the existing evidence covers this head (or re-attaches a run at 532996f).
