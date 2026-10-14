# BC-4 Plan: SQLite store (open, migrate, seed, search, cities, ping) and main wiring

Branch `feat/BC-4-sqlite-store`, cut from `feat/BC-2-backend-scaffold` @ 5da9b80 (BC-2 has since added
a9b7f73, `config.go` only: no overlap; merge `master` once BC-2 lands, never rebase). Paths are
worktree-relative. Spec: docs/SPEC.md §4, §4.1, §4.2, §5.1, §5.2, §6.1, §8 (store), §11.
Open questions: none. No `internal/domain` change is needed; BC-2 types are used as they are.

## Scope

In: new `backend/internal/store` (open, embedded migrations, upsert, search, cities, ping), new
`backend/internal/store/seed` (strict JSON loader), `modernc.org/sqlite` in go.mod/go.sum, store and
seed wiring in `cmd/api/main.go`, a real-store /healthz test, one §4.2 SPEC line.
Out: query parsing, search/cities handlers and new `Deps` fields (BC-5), ranking (BC-3), dedupe
(§9.1), deleting rows that vanished from the JSON (upsert never deletes; SPEC behaviour).

## Decisions

1. **Packages.** `internal/store` (SQL) and `internal/store/seed` (pure decode, validate, normalize; no
   DB). SPEC §3 puts "seed" under `internal/store/`, the ticket names `seed.Load`, and the split keeps
   file parsing testable without SQLite. `seed` must not import `store`; store tests import `seed`.
2. **Driver.** `modernc.org/sqlite` v1.60.1 (newest, checked with `go list -m -versions`), the only
   direct dependency; blank-imported in `store.go`; driver name `"sqlite"`. Indirect modules it pulls
   (modernc libc/mathutil/memory, x/sys, ...) are expected; `go mod tidy` decides the list.
3. **Pragmas via DSN, so they hold on every pooled connection.** Verified in v1.60.1 source and a
   scratch probe: `_pragma=` values run on each new connection; a DSN without the `file:` prefix is
   split at the first `?` and the prefix is opened as the path. `Open` builds:
   - memory (`path == ":memory:"`): `:memory:?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_txlock=immediate`
   - file: `<path>?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_txlock=immediate`
   `_txlock=immediate` makes write transactions take the write lock at BEGIN, avoiding WAL lock-upgrade
   SQLITE_BUSY that busy_timeout cannot fix. Memory: `db.SetMaxOpenConns(1)` (every connection to
   `:memory:` is a separate database); never set a conn lifetime/idle timeout. A path containing `?`
   is rejected (`DB_PATH must be a file path or :memory:`). The parent directory is not created: the
   compose mount and the repo-root `data/` exist; a missing directory fails `Open` with the path named.
   `Open` calls `db.PingContext` so a bad path fails at startup, then `Migrate`.
4. **Migrations.** `//go:embed migrations/*.sql` into `embed.FS`; `Migrate` runs
   `migrate(ctx, fs.Sub(..., "migrations"))`. Order = `fs.ReadDir` order (sorted by filename), `.sql`
   only. `0001_listings.sql` creates `schema_migrations` itself (per the ticket), so the applied set is
   read as: if `sqlite_master` has no table `schema_migrations`, nothing is applied; else
   `SELECT filename FROM schema_migrations`. Rows are closed before any transaction begins (single
   memory connection: a query left open deadlocks the next one). Each pending file: BEGIN; Exec the whole
   file (modernc runs multi-statement Exec; probed); `INSERT INTO schema_migrations (filename, applied_at)
   VALUES (?, ?)` with UTC RFC 3339; COMMIT. Any error rolls back that file and returns
   `migrate <filename>: <err>`. Reconciling SPEC §4.1 ("on open") with the ticket's `Migrate(ctx)`:
   `Open` calls the exported, idempotent `Migrate`, so a `*Store` is never returned unmigrated and tests
   can call `Migrate` again to prove the no-op. main calls only `Open`.
5. **Seed record index is 0-based** (`jq '.[3]'` is record index 3, which is `MLS_B/B8`), printed as
   `record index N`. File-level errors are prefixed with the path; main adds `seed: `. Exact text for the
   validation case: `seed: <path>: record index 3: field "status": invalid listing status: "gone"`.
6. **Strict decode with per-record index.** `os.ReadFile`; reject anything whose first non-space byte
   is not `[` (`expected a JSON array of listing records`; covers `null`, objects); `json.Unmarshal` into
   `[]json.RawMessage` (rejects malformed JSON and trailing data); then per element a
   `json.NewDecoder` + `DisallowUnknownFields()` into the unexported DTO `record`, whose 15 fields are all
   pointers (`*string`, `*float64`, `*int`) with the wire tags (`listedDate` and `status` are `*string`).
   Missing key and JSON `null` both leave nil, reported as `missing or null`, so a missing `bedrooms` is
   distinct from `0`. `*json.UnmarshalTypeError` (e.g. `bedrooms: 2.5`, `"2"`) maps to `Field =
   te.Field`; unknown keys and non-object elements get `Field = ""` and keep the json error text.
   First failure wins (fail fast); checks run decode, then presence in §4 field order, then values.
7. **Normalize then validate** (§4.2). `strings.TrimSpace` every string (id, source, address, zip,
   description, listedDate, status too); city and state: `strings.Join(strings.Fields(s), " ")`; state
   `strings.ToUpper`. Address keeps internal runs (only city/state collapse). Then: `status` via
   `domain.ParseStatus` (exact, case-sensitive; wraps `domain.ErrInvalidStatus`); `listedDate` via
   `time.Parse(domain.DateLayout, s)` (midnight UTC); `price`, `bedrooms`, `sqft` `>= 0`; plus the
   identity checks below. No bathrooms/lat/long range rules (SPEC has none).
8. **SPEC edit: §4.2** — add "`id` and `source` non-empty after trimming; `(source, id)` unique within
   the file" to the validation list. Why: §4 makes `(source, id)` the identity; an empty component is no
   identity, and an in-file duplicate would silently last-write-win while "seeded N" over-counts rows.
   Sample data passes. Record the edit in a `lattice comment` on BC-4.
9. **Dates in SQL.** Write `l.ListedDate.Format(domain.DateLayout)` (never bind a `time.Time`: the driver
   would store its own format); the column is declared `TEXT`, so it scans as a string, parsed back with
   `domain.DateLayout`. Scanned `status` goes through `domain.ParseStatus` (corrupt row fails loudly).
10. **Search semantics** for `domain.SearchFilter` as it exists: pointer bounds apply when non-nil (a
    pointer to 0 is a real bound: `MaxPrice=&0` matches nothing). `City`: `strings.TrimSpace`; empty
    after trim means absent. Keyword: only `KeywordTokens` drive SQL (`Keyword` is echo-only, BC-5
    tokenizes); each token is trimmed and `strings.ToLower`ed, empty tokens skipped, one `instr` clause
    per token AND-ed. `Statuses`: one placeholder per element; nil/empty means no status clause (all
    three, matching "absent means all"). Clause order fixed as the §6.1 table. Result is a non-nil
    empty slice when nothing matches. ORDER BY is text order: `B10 < B11 < B7 < B8 < B9`.
11. **Cities first-seen = lowest rowid** (first insertion; on a fresh DB the seed file order). Upsert's
    DO UPDATE keeps the rowid. Grouping uses `COLLATE NOCASE`, the same equivalence as the city filter
    (ASCII-only, as SQLite's). Empty cities excluded. Sample data has no casing or whitespace variants;
    expected `["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]`.
12. **main log line.** `logger.Info(fmt.Sprintf("seeded %d listings", n), slog.Int("count", n),
    slog.String("listings_file", ...), slog.String("db_path", ...))` so the line contains the literal
    `seeded 12 listings`. `Deps` gains nothing: only `Pinger: st` is wired; BC-5 adds search/cities.

## SQL

`backend/internal/store/migrations/0001_listings.sql`:
```sql
CREATE TABLE schema_migrations (filename TEXT PRIMARY KEY, applied_at TEXT NOT NULL);
CREATE TABLE listings (
  id TEXT NOT NULL, source TEXT NOT NULL, address TEXT NOT NULL, city TEXT NOT NULL,
  state TEXT NOT NULL, zip TEXT NOT NULL, price REAL NOT NULL, bedrooms INTEGER NOT NULL,
  bathrooms REAL NOT NULL, sqft INTEGER NOT NULL, latitude REAL NOT NULL, longitude REAL NOT NULL,
  listed_date TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active','pending','sold')),
  description TEXT NOT NULL,
  PRIMARY KEY (source, id)
);
CREATE INDEX listings_city_nocase ON listings (city COLLATE NOCASE);
CREATE INDEX listings_price ON listings (price);
CREATE INDEX listings_bedrooms ON listings (bedrooms);
CREATE INDEX listings_status ON listings (status);
```
Upsert (prepared once per transaction, exec per listing): `INSERT INTO listings (<15 columns>) VALUES
(<15 ?>) ON CONFLICT(source, id) DO UPDATE SET address = excluded.address, city = excluded.city, ...`
for all 13 non-key columns. One transaction for the whole slice; returns `len(listings)` on commit; any
failure rolls back everything and returns `upsert <source>/<id>: <err>`.
Search, all filters set (2 tokens, 2 statuses), exact text from `buildSearch`:
```sql
SELECT id, source, address, city, state, zip, price, bedrooms, bathrooms, sqft, latitude, longitude,
listed_date, status, description FROM listings WHERE price >= ? AND price <= ? AND bedrooms >= ?
AND city = ? COLLATE NOCASE AND instr(lower(description), ?) > 0 AND instr(lower(description), ?) > 0
AND status IN (?, ?) ORDER BY source, id
```
(single line in code; no filters means no `WHERE`). Built from constant fragments joined with
`" AND "` and `strings.Repeat`-style `?` lists; never `fmt.Sprintf` and never a user value in the text.
Cities:
```sql
SELECT city FROM listings
WHERE rowid IN (SELECT min(rowid) FROM listings WHERE city <> '' GROUP BY city COLLATE NOCASE)
ORDER BY city COLLATE NOCASE
```

## Public API

```go
package store
type Store struct{ /* db *sql.DB */ }
func Open(ctx context.Context, path string) (*Store, error) // path = DB_PATH (the ticket's "dsn")
func (s *Store) Migrate(ctx context.Context) error
func (s *Store) Close() error
func (s *Store) Ping(ctx context.Context) error // db.PingContext; driver runs "select 1"; satisfies httpapi.Pinger
func (s *Store) Upsert(ctx context.Context, listings []domain.Listing) (int, error)
func (s *Store) Search(ctx context.Context, f domain.SearchFilter) ([]domain.Listing, error)
func (s *Store) Cities(ctx context.Context) ([]string, error) // non-nil empty when no rows
// unexported: dsn(path) string; (s *Store) migrate(ctx, fs.FS) error;
// buildSearch(domain.SearchFilter) (string, []any); scanListing(*sql.Rows) (domain.Listing, error)

package seed
func Load(path string) ([]domain.Listing, error) // file order preserved
type RecordError struct{ Index int; Field string; Err error } // Field is the JSON name, "" if none
func (e *RecordError) Error() string // `record index 3: field "status": <Err>` or `record index 3: <Err>`
func (e *RecordError) Unwrap() error
```

## Files

| File | Purpose |
|---|---|
| `backend/go.mod`, `backend/go.sum` (new) | `require modernc.org/sqlite v1.60.1` + indirects via `go mod tidy` |
| `backend/internal/store/store.go` | package doc, driver import, `Store`, `Open`, `dsn`, `Close`, `Ping` |
| `backend/internal/store/migrate.go` | embed FS, `Migrate`, `migrate(fs.FS)` |
| `backend/internal/store/migrations/0001_listings.sql` | schema above |
| `backend/internal/store/listings.go` | `Upsert`, `Search`, `buildSearch`, `scanListing`, `Cities` |
| `backend/internal/store/store_internal_test.go` | package `store`: pragmas, migrate (MapFS), `buildSearch` text |
| `backend/internal/store/store_test.go` | package `store_test`: open/ping/upsert/search/cities on real data |
| `backend/internal/store/seed/seed.go` | package doc, `record` DTO, `Load`, `RecordError`, normalize/validate |
| `backend/internal/store/seed/seed_test.go` | package `seed_test`: real file, normalization, failure table |
| `backend/internal/httpapi/health_store_test.go` (new) | /healthz against a real `:memory:` store; `var _ httpapi.Pinger = (*store.Store)(nil)` |
| `backend/cmd/api/main.go` | seed.Load, store.Open, Upsert, log, `Pinger: st`, close on exit |
| `docs/SPEC.md` | §4.2 one-line identity-check edit (decision 8) |

main.go order inside `run()`: config, logger, `signal.NotifyContext`, then **`seed.Load` before
`store.Open`** (a bad file never creates a DB file), `store.Open(ctx, cfg.DBPath)`, a deferred closure
that closes the store and logs a close error at error level, `st.Upsert`, the "seeded" log, then the
existing `net.Listen` / server code with `Pinger: st` in `Deps`. Errors: `seed: %w`, `store: %w`,
`seed: %w` (upsert), returned to `main`, which already logs `fatal` and exits 1. Update the package
doc to mention the store. Store-side errors carry no `store:` prefix (callers add context, as
`config.Load` does).

## Tests (stdlib `testing`, table-driven, `go test -race`)

Store tests seed the real file `../../../data/sample_listings.json` (seed tests:
`../../../../data/...`) into `:memory:` via a helper `newSeededStore(t)` (Open, Load, Upsert, assert 12,
`t.Cleanup` closes and `t.Error`s a close error). IDs below are `(source, id)` text order.

| # | Test | Asserts | Covers |
|---|---|---|---|
| T1 | `TestSeedTwiceIdempotent` | first and second Upsert return 12; `Search({})` has 12 rows both times | §8 seed twice; ticket "12 after seeding twice" |
| T2 | `TestRoundTrip` | `Search({})` deep-equals `seed.Load` output sorted by (source,id), incl. `ListedDate` | upsert/scan columns, date TEXT |
| T3 | `TestUpsertUpdates` | re-upsert A1 with price 1 and new city casing: still 12 rows, values updated | DO UPDATE SET all non-key cols |
| T4 | `TestUpsertAtomic` | fresh store, Upsert [valid, status "gone"]: error names `upsert`, 0 rows after | one transaction, CHECK |
| T5 | `TestSearchFilters` (table) | none → all 12 `[A1 A2 A3 A4 A5 A6 A7 B10 B11 B7 B8 B9]`; minPrice 500000 → `[A2 A4 A7 B10 B8]`; minPrice 450000 → 9 rows incl. A1 (inclusive); maxPrice 450000 → `[A1 A3 A6 B9]`; min 450000 + max 470000 → `[A1 A5 B11 B7]`; maxPrice &0 → none; minPrice &0 / minBedrooms &0 → 12; minBedrooms 4 → `[A4 A7]`; city "springfield" → `[A1 A2 B7 B8]`; city "  FAIRFAX  " → `[A3 B9]`; city "   " → 12; city "Spring" → none; tokens ["pets"] → `[A3 A4 A5 A6 A7 B7 B9]`; ["PETS"] same; ["top-floor"] → `[A1]`; ["fenced","kitchen"] → `[A2 B8]`; ["metro","quiet"] → `[A5]`; ["fenced","pool"] → none; ["%"], ["_"] → none; statuses [pending] → `[A7]`, [sold] → none, [active] → 11, [active,pending] → 12, [pending,sold] → `[A7]`, nil → 12; combined city Springfield + minBedrooms 3 + maxPrice 526000 → `[A2]`; combined 400000..550000 + minBedrooms 3 + ["pets"] + [active] → `[A5 A6]`, same with [active,pending] → `[A5 A6 A7]` | every §6.1 filter alone and combined; multi-token AND; status sets; city case-insensitive + trimmed; empty set semantics |
| T6 | `TestSearchLiteralWildcards` | add TEST/T1 description "100% move-in ready, pet_friendly": ["100%"], ["%"], ["_"], ["pet_"], ["pet_friendly","100%"] → exactly `[TEST/T1]` (LIKE would also hit "pets") | §6.1 instr, `%`/`_` literal |
| T7 | `TestSearchInjection` | city `Springfield' OR '1'='1` → none; token `') or 1=1 --` → none; no-match result is non-nil | parameterization |
| T8 | `TestCities` | sample → exact 6-city slice; + TEST/T2 city "SPRINGFIELD" → unchanged; fresh store [vienna, Vienna] → `["vienna"]`, reversed → `["Vienna"]`; + "arlington" → sorts first; empty store → `[]` | §5.2 distinct/sort/first-seen |
| T9 | `TestPing` | nil on open store; error after `Close` | Ping |
| T10 | `TestReopenFile` | file in `t.TempDir()`: Open, seed, Close, Open, seed → 12 rows | restart persistence |
| T11 | `TestOpenErrors` | path under a missing dir → error containing the path; `x.db?mode=ro` → rejected | Open failures |
| I1 | `TestOpenPragmas` (internal) | memory: busy_timeout 5000, foreign_keys 1, journal_mode `memory`, `Stats().MaxOpenConnections == 1`; file: two simultaneous `db.Conn`s each report `wal`, 5000, 1 | pragmas on every conn |
| I2 | `TestMigrateTwice` (internal) | after Open, `Migrate` again: nil; `schema_migrations` = exactly `0001_listings.sql`; four indexes present in `sqlite_master` | ticket "migrate twice no-op", §4.1 |
| I3 | `TestMigrateOrderAndRollback` (internal, `fstest.MapFS`) | 0002 inserts into 0001's table (order); a broken 0003 returns an error naming `0003_bad.sql`, its table absent, no row for it; rerun with 0003 fixed applies only 0003 | ordering, per-file tx |
| I4 | `TestBuildSearch` (internal) | all-filters case equals the SQL text above and args `[min max beds city tok1 tok2 st1 st2]`; no-filter case has no WHERE; no user value appears in the SQL text | ticket "no concatenation" |
| S1 | `TestLoadSample` | 12 records in file order; A1 fully checked (ListedDate `2026-08-29` UTC, StatusActive, "VA") | §4.2 happy path |
| S2 | `TestLoadNormalizes` | id " X1 " → "X1"; city "  Falls   Church " → "Falls Church"; state " va " → "VA"; address "1  Main St " → "1  Main St"; status/date padded → parsed | §4.2 normalization |
| S3 | `TestLoadRecordErrors` (table; bad record at index 2 of 3) | `errors.As` `*RecordError` with `Index == 2` and Field: missing `bedrooms`, null `price`, unknown key `garage` (Field "", message has `garage`), status "gone" (`errors.Is ErrInvalidStatus`), status "Active", listedDate "2026-13-01"/"08/29/2026"/"2026-08-29T00:00:00Z", price -1, bedrooms -1, sqft -1, bedrooms 2.5, bedrooms "2", id "   ", source "", element `"x"` (Field ""), duplicate (source,id) of index 0 (Field "", message has `record index 0`); price/bedrooms/sqft 0 are valid | §4.2 validation naming index + field |
| S4 | `TestLoadFileErrors` | missing file → `errors.Is fs.ErrNotExist`, path in message; `{}`, `null`, malformed → error; `[]` → 0 listings, nil | file-level |
| H1 | `TestHealthzRealStore` | `:memory:` store: GET /healthz 200 `{"status":"ok"}`; after `Close`: 503 `{"status":"degraded","error":"database unavailable"}` (reuses `newTestRouter`) | §5.1, card note 3 |

Existing `TestHealthz` (failing fake Pinger) stays untouched and green.

## Lint notes (golangci-lint 2.14, no exclusion presets)

Probed with the repo `.golangci.yml`: errcheck flags `defer rows.Close()`, `defer db.Close()`,
`stmt.Close()` and `os.RemoveAll` even in defers. Use `defer func() { _ = rows.Close() }()` with
`rows.Err()` checked after the loop, `defer func() { _ = tx.Rollback() }()` (ErrTxDone after commit),
same for `stmt.Close()`; tests close via `t.Cleanup` and report errors. revive: package comments for
`store` and `seed`, doc comments on every exported identifier, `_` for unused params. gofmt.

## Gate

`go mod tidy`, then from the worktree root `make lint-api test-api` (frontend half is still the BC-1
stub; run full `make check` only if BC-7 has landed on master by hand-off).

## Validation (native, port 18004)

From `backend/`, after `lsof -nP -iTCP:18004 -sTCP:LISTEN` shows nothing and `rm -f /tmp/brillio-bc4*`:
1. `rtk proxy grep -rnE 'Sprintf|\+' internal/store --include='*.go' | grep -v _test.go` and inspect
   every hit plus `buildSearch`: only constant fragments and `?` lists reach SQL text (the DSN concat
   in `dsn()` is a driver DSN from operator config, not SQL). Zero `Sprintf` hits expected.
2. Run 1, fresh DB: `PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store/data/sample_listings.json go run ./cmd/api > /tmp/brillio-bc4-run1.out 2>&1 &`;
   wait for `"msg":"listening"`; confirm the `seeded 12 listings` line;
   `sqlite3 /tmp/brillio-bc4.db 'select count(*) from listings'` → 12; `'select filename from schema_migrations'`
   → `0001_listings.sql`; `'pragma journal_mode'` → `wal`; `curl -si localhost:18004/healthz` → 200
   `{"status":"ok"}` with `X-Request-ID`. Stop with `kill -TERM $(lsof -t -nP -iTCP:18004 -sTCP:LISTEN)`
   (kills the real server, not the `go run` wrapper); confirm `server stopped`.
3. Run 2, same DB: same command to `run2.out`; `seeded 12 listings` again; count still 12;
   schema_migrations still one row; healthz 200; stop as above.
4. Bad seed: `jq '.[3].status = "gone"' ../data/sample_listings.json > /tmp/brillio-bc4-bad.json`; run with
   `DB_PATH=/tmp/brillio-bc4-bad.db LISTINGS_FILE=/tmp/brillio-bc4-bad.json`; `echo "exit=$?"` is non-zero;
   the `fatal` line contains `record index 3: field \"status\": invalid listing status: \"gone\"`;
   `test ! -e /tmp/brillio-bc4-bad.db` (seed is checked before the DB is opened).
5. Cleanup: nothing left on 18004; `rm -f /tmp/brillio-bc4*`. Attach the transcript with
   `lattice attach BC-4 --role validation`.

## Shared files (run-state table)

- `cmd/api/main.go`: BC-4 is the second serial writer (after BC-2, before BC-5); edits confined to
  imports, package doc, the seed/store block before `net.Listen`, `Pinger: st`, and the close.
- `httpapi`: `Deps`/`router.go`/`health_test.go` untouched; one new test file only.
- `internal/domain`: untouched (BC-3 adds files there in parallel).
- `Makefile`: untouched (BC-2's `lint-api`/`test-api` already cover new packages).
- `go.mod`/`go.sum`: BC-4 owns the dependency add; BC-3 is stdlib-only. On a merge conflict, `go mod tidy`.
- `.gitignore`: untouched; `*.db` and `*.db-*` already ignore `data/listings.db` and WAL/SHM files
  (checked with `git -c core.excludesFile=/dev/null check-ignore -v`).
- `docs/SPEC.md`: §4.2 line only (union with BC-5's §5.4 edit). Never commit `.lattice/`.

## Risks

- Memory DB lives on one connection: any nested query while `Rows` is open deadlocks (Migrate reads
  its applied set before BEGIN; Search/Cities never query while iterating).
- `lower()` and `NOCASE` fold ASCII only: non-ASCII case variants will not match (fine for the data;
  BC-15 README can note it). Go's `strings.ToLower` on tokens is a superset, so ASCII is consistent.
- Upsert never deletes: a record removed from the JSON stays in a persistent `DB_PATH` (SPEC upsert).
- `go run` does not reliably forward SIGTERM to its child: kill by listening port.
- `-race` with the pure-Go driver is slower; the suite still runs in seconds.

## Size

Production ~530 LOC (store.go 90, migrate.go 90, listings.go 170, SQL 20, seed.go 150, main.go +30);
tests ~750 LOC. Suggested commits (conventional, no emoji): `build(backend): add modernc.org/sqlite`
with store open/migrations; `feat(store): seed loader`; `feat(store): upsert, search and cities`;
`feat(api): open, migrate and seed the store at startup` (+ SPEC line).

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)

Source: plan-review art_01M4W2MFYZ0TVC255R5K8WKHP0 (PASS, seven minor items). Triage by agent:opus-5.5-delegator.

| # | Concern | Bucket | Resolution | Section |
|---|---|---|---|---|
| 1 | "BEGIN; Exec the whole file; COMMIT" could be read as literal SQL `BEGIN` (DEFERRED, may hop pooled connections, drops `_txlock=immediate`) | obvious | Every transaction uses `tx, err := s.db.BeginTx(ctx, nil)`; `tx.ExecContext(ctx, fileSQL)`; `tx.ExecContext(ctx, insertSchemaMigration, name, appliedAt)`; `tx.Commit()`, with `defer func() { _ = tx.Rollback() }()`. Migration files must not contain `BEGIN`/`COMMIT`. Upsert likewise: `BeginTx` + `tx.PrepareContext` (statement closed with `defer func() { _ = stmt.Close() }()`) + `Commit`. Never send `BEGIN`/`COMMIT` as SQL text. | Decisions §4, Upsert |
| 2 | I3 fixture bootstrap of `schema_migrations` ambiguous | obvious | Option (b): I3 builds a raw `:memory:` `*sql.DB` (`SetMaxOpenConns(1)`, same DSN pragmas) wrapped as `&Store{db: db}`; the MapFS fixture's `0001_*.sql` creates `schema_migrations` itself plus a table; `0002_*.sql` inserts into it (ordering); `0003_bad.sql` fails; the "rerun with 0003 fixed" step is a second `migrate` call with a second MapFS against the same store and must apply only 0003. | Tests I3 |
| 3 | I2 counts indexes; `PRIMARY KEY` adds `sqlite_autoindex_listings_1` | obvious | Query `SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='listings'` and assert the set contains the four named indexes (do not assert a count). | Tests I2 |
| 4 | `DB_PATH=file:...` would be parsed as a URI by the driver | obvious | `Open` rejects a path that contains `?` OR has prefix `file:` with the same error (`DB_PATH must be a file path or :memory:`). Add a `file:x.db` case to T11. | Decisions §3, Tests T11 |
| 5 | Cities first-seen relies on the implicit rowid, which VACUUM may renumber | obvious | Add to Risks: "Cities first-seen order depends on the implicit rowid; a VACUUM could renumber it. Not run anywhere in v1." Also a one-line comment at the Cities query. | Risks, Cities |
| 6 | T2 compares `time.Time` via `reflect.DeepEqual` | obvious | Compare field by field, `ListedDate` with `.Equal`. | Tests T2 |
| 7 | Seed RecordError could carry `(source, id)` | evolutionary | Skipped for BC-4: the agreed message text (`record index 3: field "status": ...`) is what validation and the ticket card check; the index already locates the record. Recorded in a lattice comment; not ticketed (low value). | Decisions §5/§6 |

Contract: Decision 8's SPEC §4.2 edit (non-empty `id`/`source` after trim; `(source, id)` unique within one file) is accepted by the delegator: it enforces SPEC §4's existing identity rule, the sample data passes it, and the plan review raised no objection. The implementer edits docs/SPEC.md §4.2 in-branch (one sentence) and the delegator records a `lattice comment` per SPEC's deviation rule.

Base movement: `feat/BC-2-backend-scaffold` gained a9b7f73 (`config.go` only) before implementation. The implementer starts by running `git merge feat/BC-2-backend-scaffold` (merge, never rebase) in the worktree.
