# BC-4 Validation Addendum 2: re-validation on HEAD 9daad85

- **Validator:** agent:opus-5.5-validator
- **HEAD:** `9daad85836b1fe6b517b2c6bc5e76ea3f9a85709`, branch `feat/BC-4-sqlite-store` (a merge of master: BC-3 `dc35d1a` + board sync `c91fa7f` on top of `532996f`)
- **Run window (UTC):** 2026-10-14T02:46:11Z to 2026-10-14T02:47:05Z
- **Environment:** go1.27.1 darwin/arm64, sqlite3 3.54.0, node v24.21.0, npm 11.19.0
- **Prior evidence:** `BC-4-validation.md` (`art_01M4W46HFFNKS31ZVRFVYQSMYJ`, 318ebc5), `BC-4-validation-addendum.md` (`art_01M4W4S56AVRBPQ1GWFCXGK1KR`, 532996f)
- **Pre-check:** `lsof -nP -iTCP:18004 -sTCP:LISTEN` returned rc=1 (free); no `/tmp/brillio-bc4*` files; `git status --porcelain` 0 lines.

`<scratch>` = `/private/tmp/claude-503/-Users-forrest-Code-brillio-worktrees-bc-4-sqlite-store/ae5d964d-93a5-4e15-8361-cda482968d38/scratchpad`

## (a) HEAD and diffs: PASS

```sh
git rev-parse HEAD; git log --oneline -4
git diff 532996f HEAD --stat -- backend data
git diff 318ebc5 HEAD -- backend/internal/store backend/cmd backend/internal/httpapi data | wc -c
git diff --name-only 532996f HEAD | grep -v '^\.lattice/'
```

```
9daad85836b1fe6b517b2c6bc5e76ea3f9a85709
9daad85 Merge branch 'master' into feat/BC-4-sqlite-store
c91fa7f chore(lattice): sync board after BC-3
dc35d1a Merge branch 'feat/BC-3-domain-ranking' (BC-3)
532996f Merge branch 'master' into feat/BC-4-sqlite-store

 backend/internal/domain/paginate.go      |  86 +++++++++++
 backend/internal/domain/paginate_test.go | 148 ++++++++++++++++++
 backend/internal/domain/rank.go          |  87 +++++++++++
 backend/internal/domain/rank_test.go     | 255 +++++++++++++++++++++++++++++++
 backend/internal/domain/scoring.go       |  88 +++++++++++
 backend/internal/domain/scoring_test.go  | 211 +++++++++++++++++++++++++
 6 files changed, 875 insertions(+)

0                       # bytes of diff 318ebc5..HEAD under store/, cmd/, httpapi/, data/
```

`--name-status` shows all six as `A` (new files). Outside `.lattice/`, those six domain files are the only paths changed since 532996f.

## (b) Items 1-5 re-run on a binary rebuilt from 9daad85

Build: `cd backend && go build -o <scratch>/api ./cmd/api` (rc=0; the old binary was deleted first).
Each run used `PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=<worktree>/data/sample_listings.json <scratch>/api`,
started, polled until listening, stopped with `kill -TERM $P; wait $P`.

### 1. Fresh DB start: PASS

```
started pid=46938 listening pid=46938
"msg":"seeded 12 listings","count":12
```

### 2. DB contents: PASS

```
count=12
migrations=1 0001_listings.sql@2026-10-14T02:46:24Z
journal_mode=wal
```

### 3. Healthz: PASS

```
HTTP/1.1 200 OK
X-Request-Id: 70205a8995a3b85d1f2b0e91516a5107
{"status":"ok"}
```

### 4. Graceful stop and restart on the same DB: PASS

```
# stop run A
exit=0
18004 lsof rc=1 (1=free)
"msg":"shutting down"
"msg":"server stopped"
# run B, same DB_PATH
started pid=46996 listening pid=46996
"msg":"seeded 12 listings","count":12
count=12
migrations=1 0001_listings.sql@2026-10-14T02:46:24Z      # same row and applied_at: not re-applied
HTTP/1.1 200 OK
X-Request-Id: a80566f71f21ab64f32d54731a5787f0
{"status":"ok"}
# stop run B
exit=0
18004 lsof rc=1 (1=free)
"msg":"shutting down"
"msg":"server stopped"
```

### 5. Bad seed: PASS

```sh
jq '.[3].status = "gone"' <worktree>/data/sample_listings.json > <scratch>/bad.json
rm -f /tmp/brillio-bc4.db /tmp/brillio-bc4.db-wal /tmp/brillio-bc4.db-shm     # 0 files before
PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=<scratch>/bad.json <scratch>/api > <scratch>/bad3.log 2>&1; echo "exit=$?"
```

```
{"id":"B8","source":"MLS_B","status":"gone"}
exit=1
{"time":"2026-10-13T22:46:25.192782-04:00","level":"ERROR","msg":"fatal","err":"seed: <scratch>/bad.json: record index 3: field \"status\": invalid listing status: \"gone\""}
no /tmp/brillio-bc4.db created
18004 lsof rc=1
```

## (c) Uncached backend tests and `make check`: PASS

```sh
cd backend && go test -race -count=1 ./... -v > <scratch>/gotest3.log 2>&1; echo "go test rc=$?"
```

```
go test rc=0
top-level PASS=43 incl-subtests PASS=244 FAIL=0 SKIP=0
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	1.090s
ok  	brillio/backend/internal/domain	1.104s
ok  	brillio/backend/internal/httpapi	1.226s
ok  	brillio/backend/internal/store	1.530s
ok  	brillio/backend/internal/store/seed	1.156s
```

The counts rose from 36/188 at 532996f to 43/244 here. The increase comes from the new BC-3 domain tests.

```sh
cd <worktree> && make check; echo "make check rc=$?"
```

```
make check rc=0
0 issues.
> eslint . --max-warnings 0
> tsc -b --noEmit
> prettier --check .
All matched files use Prettier code style!
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	(cached)
ok  	brillio/backend/internal/domain	(cached)
ok  	brillio/backend/internal/httpapi	(cached)
ok  	brillio/backend/internal/store	(cached)
ok  	brillio/backend/internal/store/seed	(cached)
 Test Files  2 passed (2)
      Tests  4 passed (4)
```

Backend packages show `(cached)` because the uncached `-count=1` run above, against the same code, came just before. `npm ci` did not run because the node_modules sentinel was already current.

Items 6 and 8 still hold. `internal/store`, `cmd/`, `internal/httpapi` and `data/` are byte-identical to 318ebc5. The health tests re-passed in the uncached run.

## Cleanup

```
lsof -nP -iTCP:18004 -sTCP:LISTEN     -> rc=1 (free)
pgrep -fl "<scratch>/api"             -> rc=1 (none running)
rm -f /tmp/brillio-bc4.db{,-wal,-shm}; ls /tmp | grep -c '^brillio-bc4'   -> 0
git -C <worktree> status --porcelain  -> 0 lines
git -C <worktree> rev-parse HEAD      -> 9daad85836b1fe6b517b2c6bc5e76ea3f9a85709
```

**Overall: PASS**
