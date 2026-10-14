# BC-4 Validation Addendum: re-validation on HEAD 532996f

- **Validator:** agent:opus-5.5-validator
- **HEAD:** `532996fd5caa87fc8eb96868d7d896ecb413b5a5`, branch `feat/BC-4-sqlite-store` (a merge of master: BC-7 `aa65e61` + board sync `83952b0` on top of the previously validated `318ebc5`)
- **Run window (UTC):** 2026-10-14T02:43:35Z to 2026-10-14T02:44:35Z
- **Environment:** go1.27.1 darwin/arm64, sqlite3 3.54.0, node v24.21.0, npm 11.19.0
- **Prior evidence:** `BC-4-validation.md` (artifact `art_01M4W46HFFNKS31ZVRFVYQSMYJ`, HEAD 318ebc5)
- **Pre-check:** `lsof -nP -iTCP:18004 -sTCP:LISTEN` returned rc=1 (free); no `/tmp/brillio-bc4*` files; `git status --porcelain` 0 lines.

`<scratch>` = `/private/tmp/claude-503/-Users-forrest-Code-brillio-worktrees-bc-4-sqlite-store/ae5d964d-93a5-4e15-8361-cda482968d38/scratchpad`

## (a) HEAD and backend/data diff: PASS

```sh
git rev-parse HEAD; git log --oneline -4
git diff 318ebc5 532996f -- backend data | wc -c
git diff --name-only 318ebc5 532996f | grep -v '^frontend/'
```

```
532996fd5caa87fc8eb96868d7d896ecb413b5a5
532996f Merge branch 'master' into feat/BC-4-sqlite-store
83952b0 chore(lattice): sync board after BC-7
aa65e61 Merge branch 'feat/BC-7-frontend-scaffold' (BC-7)
318ebc5 Merge branch 'master' into feat/BC-4-sqlite-store
0                                   # bytes of diff under backend/ and data/
```

Non-frontend paths changed: `.lattice/**` (board, events, artifacts, BC-7 evidence), `Makefile`
(the frontend block), and `docs/SPEC.md`. The overall diff is 55 files, +8047/-71, almost all under `frontend/`.

## (b) Items 1-5 re-run on a binary rebuilt from 532996f

Build: `cd backend && go build -o <scratch>/api ./cmd/api` (rc=0; the old binary was deleted first).
Each run used `PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=<worktree>/data/sample_listings.json <scratch>/api`,
started, polled until listening, stopped with `kill -TERM $P; wait $P`.

### 1. Fresh DB start: PASS

```
started pid=35048 listening pid=35048
"msg":"seeded 12 listings","count":12
```

### 2. DB contents: PASS

```
count=12
migrations=1 0001_listings.sql@2026-10-14T02:43:49Z
journal_mode=wal
listed_date sample=A1 2026-08-29 text
```

### 3. Healthz: PASS

```sh
curl -si localhost:18004/healthz | grep -E '^HTTP|^X-Request-Id|^\{'
```

```
HTTP/1.1 200 OK
X-Request-Id: bfb5e751a7c345ea2462a30af4b6f847
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
started pid=35115 listening pid=35115
"msg":"seeded 12 listings","count":12
count=12
migrations=1 0001_listings.sql@2026-10-14T02:43:49Z      # same row and applied_at: not re-applied
HTTP/1.1 200 OK
X-Request-Id: d3549955a9cb80530d457e7b33812162
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
PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=<scratch>/bad.json <scratch>/api > <scratch>/bad2.log 2>&1; echo "exit=$?"
```

```
{"id":"B8","source":"MLS_B","status":"gone"}
exit=1
{"time":"2026-10-13T22:43:57.113952-04:00","level":"ERROR","msg":"fatal","err":"seed: <scratch>/bad.json: record index 3: field \"status\": invalid listing status: \"gone\""}
no /tmp/brillio-bc4.db created
18004 lsof rc=1
```

## (c) Full gate, `make check`: PASS

```sh
cd <worktree> && make check; echo "make check rc=$?"
```

`frontend/node_modules/.package-lock.json` was already newer than `frontend/package-lock.json`, so `npm ci` did not run.

```
cd backend && golangci-lint run
0 issues.
cd frontend && npm run lint && npm run typecheck && npm run format:check
> brillio-listings-web@0.0.0 lint
> eslint . --max-warnings 0
> brillio-listings-web@0.0.0 typecheck
> tsc -b --noEmit
> brillio-listings-web@0.0.0 format:check
> prettier --check .
Checking formatting...
All matched files use Prettier code style!
cd backend && go test -race ./...
?   	brillio/backend/cmd/api	[no test files]
ok  	brillio/backend/internal/config	(cached)
ok  	brillio/backend/internal/domain	(cached)
ok  	brillio/backend/internal/httpapi	(cached)
ok  	brillio/backend/internal/store	(cached)
ok  	brillio/backend/internal/store/seed	(cached)
cd frontend && npm test
> brillio-listings-web@0.0.0 test
> vitest run
 RUN  v5.0.4 /Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store/frontend
 Test Files  2 passed (2)
      Tests  4 passed (4)
   Duration  797ms (environment 74%, setup 13%, tests 11%, transform 1%)
make check rc=0
```

The backend results above came from the Go test cache, so I also ran an uncached pass on this HEAD:

```sh
cd backend && go test -race -count=1 ./... -v
```

```
go test rc=0
top-level PASS=36 incl-subtests PASS=188 FAIL=0
ok  	brillio/backend/internal/config	1.184s
ok  	brillio/backend/internal/domain	1.133s
ok  	brillio/backend/internal/httpapi	1.251s
ok  	brillio/backend/internal/store	1.526s
ok  	brillio/backend/internal/store/seed	1.183s
```

Items 6 (degraded health via `TestHealthzRealStore` / `TestHealthz/pinger_fails`) and 8 (parameterized-SQL
review) depend only on `backend/`, which is byte-identical to 318ebc5 per (a). The uncached run above
re-passed the Healthz tests as part of the 36 top-level PASS lines.

## Cleanup

```
lsof -nP -iTCP:18004 -sTCP:LISTEN     -> rc=1 (free)
pgrep -fl "<scratch>/api"             -> rc=1 (none running)
rm -f /tmp/brillio-bc4.db{,-wal,-shm}; ls /tmp | grep -c '^brillio-bc4'   -> 0
git -C <worktree> status --porcelain  -> 0 lines
git -C <worktree> rev-parse HEAD      -> 532996fd5caa87fc8eb96868d7d896ecb413b5a5
```

**Overall: PASS**
