# BC-4 Validation: SQLite store (migrations, JSON seed loader, filtered search, cities)

- **Validator:** agent:opus-5.5-validator
- **Worktree / branch:** `/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store`, `feat/BC-4-sqlite-store`
- **HEAD:** `318ebc5005eb1f72b3b849a02ed9c6626c24def7` (`git rev-parse HEAD`; `git status --porcelain` clean before and after)
- **Run window (UTC):** 2026-10-14T02:32:04Z to 2026-10-14T02:33:59Z (local logs show 2026-10-13 22:32 to 22:33 -04:00)
- **Environment:** go1.27.1 darwin/arm64; sqlite3 3.54.0 (2026-04-09); golangci-lint 2.14.0; jq 1.8.2; macOS (Darwin 27.0.0)
- **Binding:** `PORT=18004`, `DB_PATH=/tmp/brillio-bc4.db`. Pre-check: `lsof -nP -iTCP:18004 -sTCP:LISTEN` returned nothing (rc=1), no `/tmp/brillio-bc4.db*` files existed.
- **Binary:** built once with `cd backend && go build -o <scratch>/api ./cmd/api` (rc=0) and run directly, so SIGTERM reaches the server process. The listening pid matched the started pid on every run.

`<scratch>` = `/private/tmp/claude-503/-Users-forrest-Code-brillio-worktrees-bc-4-sqlite-store/ae5d964d-93a5-4e15-8361-cda482968d38/scratchpad`

---

## 1. Fresh DB start: PASS

```sh
cd backend && PORT=18004 DB_PATH=/tmp/brillio-bc4.db \
  LISTINGS_FILE=/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store/data/sample_listings.json \
  nohup <scratch>/api > <scratch>/run1.log 2>&1 &      # started pid=85667
# poll lsof up to 30x1s -> listening on the 1st poll
lsof -nP -iTCP:18004 -sTCP:LISTEN
```

```
api     85667 forrest    8u  IPv6 0xc5ba094a32bbb4ad      0t0  TCP *:18004 (LISTEN)
```

`run1.log`:

```
{"time":"2026-10-13T22:32:20.554018-04:00","level":"INFO","msg":"seeded 12 listings","count":12,"listings_file":"/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store/data/sample_listings.json","db_path":"/tmp/brillio-bc4.db"}
{"time":"2026-10-13T22:32:20.554452-04:00","level":"INFO","msg":"listening","addr":"[::]:18004","log_level":"INFO","cors_origins":["http://localhost:5173"],"shutdown_timeout":"10s"}
```

## 2. Database contents: PASS

```sh
sqlite3 /tmp/brillio-bc4.db 'select count(*) from listings'
sqlite3 /tmp/brillio-bc4.db 'select * from schema_migrations'
sqlite3 /tmp/brillio-bc4.db 'pragma journal_mode'
sqlite3 -header -column /tmp/brillio-bc4.db 'select source, id, city, state, listed_date, typeof(listed_date) as t, status from listings order by source, id'
sqlite3 /tmp/brillio-bc4.db '.schema'
```

```
--- count
12
--- schema_migrations
0001_listings.sql|2026-10-14T02:32:20Z
--- journal_mode
wal
--- rows
source  id      city      state  listed_date   t    status
------  ---  -----------  -----  -----------  ----  -------
MLS_A   A1   Springfield  VA     2026-08-29   text  active
MLS_A   A2   Springfield  VA     2026-09-02   text  active
MLS_A   A3   Fairfax      VA     2026-09-01   text  active
MLS_A   A4   Reston       VA     2026-09-03   text  active
MLS_A   A5   Vienna       VA     2026-09-04   text  active
MLS_A   A6   Manassas     VA     2026-08-10   text  active
MLS_A   A7   Chantilly    VA     2026-07-28   text  pending
MLS_B   B10  Reston       VA     2026-08-20   text  active
MLS_B   B11  Vienna       VA     2026-08-15   text  active
MLS_B   B7   Springfield  VA     2026-08-27   text  active
MLS_B   B8   Springfield  VA     2026-08-30   text  active
MLS_B   B9   Fairfax      VA     2026-08-25   text  active
```

Schema excerpt (from `.schema`): `listed_date TEXT NOT NULL, -- YYYY-MM-DD`,
`status TEXT NOT NULL CHECK (status IN ('active', 'pending', 'sold'))`, `PRIMARY KEY (source, id)`,
indexes `listings_city_nocase (city COLLATE NOCASE)`, `listings_price`, `listings_bedrooms`, `listings_status`.

The source JSON field `listedDate` (for example `"listedDate": "2026-08-29"` on A1) lands in
`listed_date` as `text` `YYYY-MM-DD`. Note: the sample data's city/state/status values are already
in canonical form, so this run shows they pass through unchanged. It does not exercise trimming or
case normalization of messy input; the unit tests cover that.

## 3. Healthz: PASS

```sh
curl -si localhost:18004/healthz
```

```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: aaf0efc2df2c4a4a0adf16b30adb8a5e
Date: Wed, 14 Oct 2026 02:32:26 GMT
Content-Length: 16

{"status":"ok"}
```

Matching access-log line: `{"msg":"request","method":"GET","path":"/healthz","status":200,"duration_ms":0.112,"request_id":"aaf0efc2df2c4a4a0adf16b30adb8a5e","remote_addr":"[::1]:63959"}`

## 4. Graceful stop, restart on the same DB (idempotent seed and migrations): PASS

Stop run 1:

```sh
P=$(lsof -t -nP -iTCP:18004 -sTCP:LISTEN); kill -TERM $P   # P=85667 (= started pid)
# poll kill -0 up to 10 s
```

```
listening pid=85667 (started pid=85667)
pid 85667 exited
lsof rc=1
{"time":"2026-10-13T22:32:39.327185-04:00","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-13T22:32:39.327298-04:00","level":"INFO","msg":"server stopped"}
-rw-r--r--@ 1 forrest  wheel  36864 Oct 13 22:32 /tmp/brillio-bc4.db      # -wal/-shm gone after clean close
```

Run 2, same `DB_PATH` (started pid=87110, listening on the 1st poll):

```
{"time":"2026-10-13T22:32:44.541905-04:00","level":"INFO","msg":"seeded 12 listings","count":12,...,"db_path":"/tmp/brillio-bc4.db"}
{"time":"2026-10-13T22:32:44.542199-04:00","level":"INFO","msg":"listening","addr":"[::]:18004",...}
--- count
12
--- schema_migrations   (select count(*), group_concat(filename||"@"||applied_at) from schema_migrations)
1|0001_listings.sql@2026-10-14T02:32:20Z
--- healthz
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 1dcd75e8f98d2869075c283821d24cbe
Content-Length: 16

{"status":"ok"}
```

`applied_at` is still run 1's timestamp, so the migration was not re-applied; the upsert kept 12 rows (no duplicates).

Stop run 2:

```
listening pid=87110 (started pid=87110)
pid 87110 exited
lsof rc=1
{"time":"2026-10-13T22:32:58.713943-04:00","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-13T22:32:58.714049-04:00","level":"INFO","msg":"server stopped"}
```

Extra check, shutdown exit code (run 3, fresh DB, start/kill/`wait` in one shell so the status is visible):

```sh
<scratch>/api > <scratch>/run3.log 2>&1 & P=$!; <poll>; curl ... /healthz; kill -TERM $P; wait $P; echo "exit=$?"
```

```
healthz=200
exit=0
"msg":"seeded 12 listings"
"msg":"listening"
"msg":"request"
"msg":"shutting down"
"msg":"server stopped"
```

## 5. Bad seed fails fast and creates no DB: PASS

```sh
jq '.[3].status = "gone"' <worktree>/data/sample_listings.json > <scratch>/bad.json
jq -c '.[3] | {id, source, status}' <scratch>/bad.json
rm -f /tmp/brillio-bc4.db /tmp/brillio-bc4.db-wal /tmp/brillio-bc4.db-shm     # 0 files remained
cd backend && PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=<scratch>/bad.json <scratch>/api > <scratch>/bad.log 2>&1; echo "exit=$?"
test ! -e /tmp/brillio-bc4.db && echo "no /tmp/brillio-bc4.db created"
```

```
{"id":"B8","source":"MLS_B","status":"gone"}
exit=1
{"time":"2026-10-13T22:33:06.777388-04:00","level":"ERROR","msg":"fatal","err":"seed: /private/tmp/claude-503/-Users-forrest-Code-brillio-worktrees-bc-4-sqlite-store/ae5d964d-93a5-4e15-8361-cda482968d38/scratchpad/bad.json: record index 3: field \"status\": invalid listing status: \"gone\""}
no /tmp/brillio-bc4.db created
lsof rc=1     # nothing ever listened on 18004
```

The fatal line names `record index 3` and field `"status"`, and the exit is non-zero. `main.go` calls `seed.Load` before `store.Open`, which is why no DB file exists.

## 6. Degraded health: PASS (via tests; real-system probe skipped)

I skipped the live probe. `Store.Ping` is `db.PingContext` on a pool that already holds an open
SQLite connection. Deleting, renaming or chmod-ing the file on a live process would not make the
ping fail, because the fd stays open. Forcing a failure would take a hack (killing the fd, injecting
faults), so I used the tests the brief names instead:

```sh
cd backend && go test -race -count=1 -run 'Healthz' -v ./internal/httpapi/
```

```
=== RUN   TestHealthzRealStore
--- PASS: TestHealthzRealStore (0.01s)
=== RUN   TestHealthz
=== RUN   TestHealthz/no_pinger
=== RUN   TestHealthz/pinger_ok
=== RUN   TestHealthz/pinger_fails
=== RUN   TestHealthz/HEAD
--- PASS: TestHealthz (0.00s)
    --- PASS: TestHealthz/no_pinger (0.00s)
    --- PASS: TestHealthz/pinger_ok (0.00s)
    --- PASS: TestHealthz/pinger_fails (0.00s)
    --- PASS: TestHealthz/HEAD (0.00s)
PASS
ok  	brillio/backend/internal/httpapi	1.357s
```

`TestHealthzRealStore` (`internal/httpapi/health_store_test.go`) opens a real `store.Open(ctx, ":memory:")`.
It asserts `200 {"status":"ok"}`, then calls `st.Close()` and asserts
`503 {"status":"degraded","error":"database unavailable"}` through the real router.

## 7. Gate on HEAD: PASS

```sh
cd <worktree> && make lint-api test-api; echo "make rc=$?"
```

```
cd backend && golangci-lint run
0 issues.
cd backend && go test -race ./...
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	(cached)
ok  	brillio/backend/internal/domain	(cached)
ok  	brillio/backend/internal/httpapi	(cached)
ok  	brillio/backend/internal/store	(cached)
ok  	brillio/backend/internal/store/seed	(cached)
make rc=0
```

Uncached verbose run:

```sh
cd backend && go test -race -count=1 ./... -v > <scratch>/gotest.log 2>&1; echo "go test rc=$?"
grep -c '^--- PASS' <scratch>/gotest.log; grep -cE -- '--- FAIL|^FAIL' <scratch>/gotest.log
```

```
go test rc=0
PASS lines (top-level '^--- PASS'): 36
PASS lines incl. subtests: 188
FAIL lines: 0
SKIP lines: 0
ok  	brillio/backend/internal/config	1.131s
ok  	brillio/backend/internal/domain	1.132s
ok  	brillio/backend/internal/httpapi	1.233s
ok  	brillio/backend/internal/store	1.540s
ok  	brillio/backend/internal/store/seed	1.143s
```

## 8. Parameterized SQL (no user input concatenated into SQL): PASS

```sh
cd backend && grep -rnE 'Sprintf|\+ ' internal/store --include='*.go' | grep -v _test.go
```

```
internal/store/listings.go:156:		query += sqlWhere + strings.Join(clauses, sqlAnd)
internal/store/listings.go:158:	return query + orderBySourceID, args
internal/store/seed/seed.go:67:		return fmt.Sprintf("record index %d: %v", e.Index, e.Err)
internal/store/seed/seed.go:69:	return fmt.Sprintf("record index %d: field %q: %v", e.Index, e.Field, e.Err)
internal/store/store.go:77:		return path + dsnMemoryParams
internal/store/store.go:79:	return path + dsnFileParams
```

The `\+ ` pattern misses a `+` with no space after it, so I also ran `grep -rnE '\+'`. It found two more lines:

```
internal/store/listings.go:26:	selectListings = "SELECT id, source, ... sqft, " +
internal/store/listings.go:151:		clauses = append(clauses, whereStatusIn+strings.Join(marks, placeholderSep)+whereStatusInEnd)
```

What each hit is:
- `listings.go:26`: concatenation of two string literals inside a `const` block. Compile-time constant SQL.
- `listings.go:151` (`buildSearch`): the constants `"status IN ("` + n copies of the constant `"?"` joined by `", "` + `")"`. Status values go to `args`, never into the text.
- `listings.go:156` (`buildSearch`): `" WHERE "` + a join of clause constants (`"price >= ?"`, `"city = ? COLLATE NOCASE"`, `"instr(lower(description), ?) > 0"`, ...) with `" AND "`. Every filter value is appended to `args`.
- `listings.go:158` (`buildSearch`): appends the constant `" ORDER BY source, id"`.
- `seed.go:67,69`: `RecordError.Error()` message text. Not SQL.
- `store.go:77,79` (`dsn()`): driver DSN = operator-config `DB_PATH` + constant `_pragma` params. Not SQL text, and never request input. `Open` rejects paths containing `?` or starting with `file:`.

Other SQL execution sites (`grep -nE 'Exec|Query'`) all pass constant statements: `upsertListing`
(prepared, 15 `?`), `selectCities`, `countSchemaMigrations`, `selectAppliedFiles`,
`insertSchemaMigration` (with args). Migration bodies come from `//go:embed migrations/*.sql`, so they
are compile-time content. `fmt.Sprintf` appears nowhere in SQL construction.

---

## Cleanup

```
lsof -nP -iTCP:18004 -sTCP:LISTEN          -> rc=1 (nothing listening)
pgrep -fl "<scratch>/api"                  -> rc=1 (no scratch binaries running)
rm -f /tmp/brillio-bc4.db{,-wal,-shm}; ls /tmp | grep -c '^brillio-bc4'   -> 0
git -C <worktree> status --porcelain       -> 0 lines (clean)
git -C <worktree> rev-parse HEAD           -> 318ebc5005eb1f72b3b849a02ed9c6626c24def7
```

Validation only used port 18004 (never 8080 or 5173). The binary was built into the scratch dir,
not `backend/bin/`, so the worktree was not touched.

**Overall: PASS**
