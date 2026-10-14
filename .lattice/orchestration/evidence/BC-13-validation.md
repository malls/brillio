# BC-13 validation: Docker Compose dev stack (hot reload, health gating, make smoke)

- **Ticket:** BC-13 (task_01M4VXK3F2NX1DGFTZ60YX6K0H)
- **Branch / HEAD:** `feat/BC-13-compose-stack` @ `21cd239cc4ead693bc80d0411eae00a0473f4422` (confirmed with `git rev-parse HEAD` in the worktree and in the fresh clone)
- **Validator:** agent:opus-5.5-validator
- **Date:** 2026-10-14 (run 04:20-04:24 UTC)
- **Host:** macOS 27.0.1, arm64; Docker client/server 29.8.1 (arm64); Docker Compose v5.5.1
- **Operator shell env (recorded as found):** `DOCKER_DEFAULT_PLATFORM=linux/amd64`, `DOCKER_BUILDKIT=0`, `COMPOSE_DOCKER_CLI_BUILD=0`.
  Every raw `docker compose up|build` below ran as `env -u DOCKER_DEFAULT_PLATFORM docker compose ...` (accepted deviation, plan "Implementation Notes"). `make up` / `make smoke` unset the variable themselves. The classic builder (BUILDKIT=0) was used throughout, as the operator's shell dictates.
- **Overall verdict: PASS** (all steps and all acceptance items).

---

## Step 0. Pre-checks: PASS

```
$ git rev-parse HEAD                       -> 21cd239cc4ead693bc80d0411eae00a0473f4422
$ git -c core.excludesFile=/dev/null status --porcelain   -> (empty)
$ lsof -nP -iTCP:8080 -sTCP:LISTEN         -> (empty)
$ lsof -nP -iTCP:5173 -sTCP:LISTEN         -> (empty)
$ docker compose ls -a                     -> asc-chat running(1), whatever-sticks exited(1)  (no brillio)
$ docker ps -a      --filter label=com.docker.compose.project=brillio -q   -> (empty)
$ docker volume ls -q --filter label=com.docker.compose.project=brillio    -> (empty)
$ docker network ls -q --filter label=com.docker.compose.project=brillio   -> (empty)
```

## Step 1. Fresh clone + cold build: PASS

```
$ git clone -q --branch feat/BC-13-compose-stack /Users/forrest/Code/brillio $S/clean
$ git -C $S/clean rev-parse HEAD           -> 21cd239cc4ead693bc80d0411eae00a0473f4422
$ env -u DOCKER_DEFAULT_PLATFORM docker compose build --no-cache api   (in $S/clean)
rc=0 elapsed=10s
Sending build context to Docker daemon     788B
Step 1/8 : FROM golang:1.27-alpine
Step 4/8 : RUN go install github.com/air-verse/air@${AIR_VERSION}     && go clean -cache -modcache
  (23 "go: downloading" lines; zero "Using cache" lines -> genuinely cold)
Successfully tagged brillio-api:latest
$ docker image inspect brillio-api   -> sha256:5d60f855c02e... arm64
```

## Step 2. `docker compose up` from the clean checkout reaches healthy: PASS

```
$ nohup env -u DOCKER_DEFAULT_PLATFORM docker compose up > $S/up.log 2>&1 &    (PID 11652)
$ poll: docker compose ps --format '{{.Service}} {{.Health}}'  (every 2s, max 150 iterations)
healthy after 21s: api healthy web healthy
$ docker compose ps
NAME            IMAGE            COMMAND                  SERVICE   STATUS                    PORTS
brillio-api-1   brillio-api      "air -c .air.toml"       api       Up 20 seconds (healthy)   0.0.0.0:8080->8080/tcp
brillio-web-1   node:24-alpine   "docker-entrypoint.s..." web       Up 6 seconds (healthy)    0.0.0.0:5173->5173/tcp
```
Health gating observed: web started only after api reported Healthy (`Container brillio-api-1 Healthy` precedes `Container brillio-web-1 Started` in up.log; web is 14 s younger than api).

## Step 3. Stack evidence: PASS

```
$ curl -si http://localhost:8080/healthz
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Request-Id: 5ae972651ab3ef6e8a202dbfed07338b
{"status":"ok"}

$ curl -s http://localhost:5173/ | grep -o '<title>[^<]*</title>'
<title>Brillio Listings</title>

$ curl -si http://localhost:5173/healthz          (through the Vite proxy)
HTTP/1.1 200 OK
x-request-id: 838dcf8e45951cda48c6d46534d37e4f
{"status":"ok"}

$ curl -s http://localhost:5173/api/cities         (through the Vite proxy)
{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}

$ curl -s 'http://localhost:5173/api/listings/search?targetBudget=450000&city=Springfield' | head -c 600
{"items":[{"rank":1,"id":"A1","source":"MLS_A","address":"123 Main St, Apt 4B","city":"Springfield","state":"VA",
"zip":"22150","price":450000,...,"relevance":{"score":0.7382,"budgetFit":1,"recency":0.3455}},{"rank":2,"id":"B7",...

api log (air):
[04:20:48] building...
[04:21:01] running...
{"level":"INFO","msg":"seeded 12 listings","count":12,"listings_file":"/data/sample_listings.json","db_path":"/data/listings.db"}
{"level":"INFO","msg":"listening","addr":"[::]:8080","log_level":"INFO","cors_origins":["http://localhost:5173"],"shutdown_timeout":"10s"}
Proxied requests arrive from the web container (remote_addr 172.18.0.3), e.g. path=/api/cities status=200.

$ docker compose exec api env | grep -E 'DB_PATH|LISTINGS_FILE|CORS_ORIGINS|PORT|LOG_LEVEL|SHUTDOWN_TIMEOUT'
LOG_LEVEL=info
CORS_ORIGINS=http://localhost:5173
SHUTDOWN_TIMEOUT=10s
PORT=8080
DB_PATH=/data/listings.db
LISTINGS_FILE=/data/sample_listings.json

$ ls $S/clean/data
listings.db  listings.db-shm  listings.db-wal  sample_listings.json
```
**SQLite WAL on the macOS bind mount works** (closes plan Resolution 5): the api seeded 12 listings into `/data/listings.db` on the `./data` bind mount, the `-wal`/`-shm` files appeared on the host, and search/cities reads succeeded. No separate probe was needed.

## Step 4. Go hot reload: PASS (~2 s, no container restart)

```
StartedAt before: 2026-10-14T04:20:47.916617419Z
edit: backend/internal/httpapi/health.go 200 branch -> healthResponse{Status: "ok-reload"}
poll curl -s http://localhost:8080/healthz -> {"status":"ok-reload"}   elapsed=2s
api log:
[04:21:23] internal/httpapi/health.go has changed
[04:21:23] building...
{"msg":"shutting down","timeout":"10s"}
{"msg":"server stopped"}
[04:21:24] running...
{"msg":"seeded 12 listings",...}
{"msg":"listening","addr":"[::]:8080",...,"cors_origins":["http://localhost:5173"]}
StartedAt after:  2026-10-14T04:20:47.916617419Z  (UNCHANGED)
api Up 37 seconds (healthy)

revert: git -C $S/clean checkout -- backend/internal/httpapi/health.go
poll -> {"status":"ok"} after 2s; StartedAt still 2026-10-14T04:20:47.916617419Z; api healthy
```

## Step 5. TSX hot reload (HMR) in the c11 browser: PASS

```
c11 browser panel opened at http://localhost:5173/ (header text "Brillio ListingsNorthern Virginia")
web StartedAt: 2026-10-14T04:21:02.538830218Z
eval: window.__bc13 = 1
edit: frontend/src/App.tsx  scope="Northern Virginia" -> scope="Northern Virginia (HMR)"
browser header after 1s: "Brillio ListingsNorthern Virginia (HMR)"
eval: {"bc13":1,"header":"Brillio ListingsNorthern Virginia (HMR)"}   -> window.__bc13 survived: no full reload
docker compose logs web: 4:22:06 AM [vite] (client) hmr update /src/App.tsx
web StartedAt after: 2026-10-14T04:21:02.538830218Z (UNCHANGED)
screenshot: BC-13-hmr.png (header reads "NORTHERN VIRGINIA (HMR)"; CSS uppercases it)

revert: git -C $S/clean checkout -- frontend/src/App.tsx
browser header after 3s: "Brillio ListingsNorthern Virginia"; window.__bc13 still 1
docker compose logs web: 4:22:18 AM [vite] (client) hmr update /src/App.tsx
web StartedAt unchanged; both services healthy
```

## Step 6. Browser search (ticket Validation line): PASS

**browser search fallback: BC-12 SearchPage not on master; searched via the 5173 proxy URL.**

```
c11 browser goto http://localhost:5173/api/listings/search?targetBudget=450000&city=Springfield
get url -> http://localhost:5173/api/listings/search?targetBudget=450000&city=Springfield
eval -> {"count":4,"ranks":[1,2,3,4],"first":{"rank":1,"id":"A1","price":450000,"city":"Springfield"},
         "keys":["items","page","scoring","appliedFilters"]}
api log: {"msg":"request","method":"GET","path":"/api/listings/search","status":200,"remote_addr":"172.18.0.3:46796"}
         (remote_addr is the web container -> the request went browser -> Vite proxy -> api)
screenshot: BC-13-browser-search.png (rank 1 A1 first, then B7, A2, B8; page.totalItems 4)
```

## Step 7. Teardown of the clone stack (`down -v`): PASS

```
kill -INT 11652 (compose up) -> "Gracefully Stopping", both containers Stopped, process exited
$ docker compose down -v  (in $S/clean)  rc=0
 Container brillio-web-1 Removed / Container brillio-api-1 Removed
 Volume brillio_gomod Removed / Volume brillio_gobuild Removed / Network brillio_default Removed
brillio-label containers / volumes / networks: (empty) (empty) (empty)
ports 8080 / 5173: free;  docker compose ls -a: no brillio
rm -rf $S/clean
```

## Step 8. Real `make smoke` in the worktree: PASS (8/8, exit 0, 24 s)

Pre: ports free, no brillio project, worktree porcelain empty.
```
$ cd /Users/forrest/Code/brillio-worktrees/bc-13-compose-stack && make smoke > BC-13-smoke.txt 2>&1; echo "exit=$?"
exit=0 elapsed=24s
```
Full output (also attached as BC-13-smoke.txt):
```
sh scripts/smoke.sh
smoke: docker compose up -d --build --wait
 Image brillio-api Building
Sending build context to Docker daemon     796B
Step 1/8 : FROM golang:1.27-alpine
 ---> b14800def2ec
Step 2/8 : ARG AIR_VERSION=v1.67.4
 ---> Using cache
Step 3/8 : ENV CGO_ENABLED=0     GOMODCACHE=/go/pkg/mod     GOCACHE=/root/.cache/go-build
 ---> Using cache
Step 4/8 : RUN go install github.com/air-verse/air@${AIR_VERSION}     && go clean -cache -modcache
 ---> Using cache
Step 5/8 : WORKDIR /app
 ---> Using cache
Step 6/8 : EXPOSE 8080
 ---> Using cache
Step 7/8 : CMD ["air", "-c", ".air.toml"]
 ---> Using cache
Step 8/8 : LABEL com.docker.compose.image.builder=classic
 ---> Using cache
 ---> 5d60f855c02e
Successfully built 5d60f855c02e
Successfully tagged brillio-api:latest
 Image brillio-api Built
 Network brillio_default Created
 Volume brillio_gomod Created
 Volume brillio_gobuild Created
 Container brillio-api-1 Created
 Container brillio-web-1 Created
 Container brillio-api-1 Started
 Container brillio-api-1 Healthy
 Container brillio-web-1 Started
 Container brillio-web-1 Healthy
PASS  a1 api /healthz is ok
PASS  a2 api /api/cities lists Springfield
PASS  a3 api search returns items
PASS  a4 api search ranks the first item 1
PASS  a5 api search rejects minPrice > maxPrice
PASS  a6 web serves the app shell
PASS  a7 web proxies /healthz to api
PASS  a8 web proxies /api/cities to api
smoke: 8 passed, 0 failed
teardown: docker compose down -v
 Container brillio-web-1 Removed
 Container brillio-api-1 Removed
 Volume brillio_gobuild Removed
 Network brillio_default Removed
 Volume brillio_gomod Removed
```
(Compose's duplicated progress lines collapsed above; the attached .txt is verbatim.)
After: brillio-label containers/volumes/networks empty; ports 8080/5173 free; no brillio project; `git -c core.excludesFile=/dev/null status --porcelain` empty.

## Step 9. Guard + `make down`: PASS

```
$ env -u DOCKER_DEFAULT_PLATFORM docker compose up -d --wait     rc=0, 22s, api+web healthy
web anonymous volume: 9c040e401f76... -> /app/node_modules  (labels: {"com.docker.volume.anonymous":""}, no project label)
$ make smoke
sh scripts/smoke.sh
smoke: a brillio stack already exists; run make down first
make: *** [smoke] Error 1          (exit 2 from make; elapsed 1s)
$ docker compose ps  -> api Up 25 seconds (healthy) / web Up 10 seconds (healthy)   (stack left running)
$ make down
docker compose down -v
 Container brillio-web-1 Removed / Container brillio-api-1 Removed
 Volume brillio_gomod Removed / Volume brillio_gobuild Removed / Network brillio_default Removed
down exit=0
brillio-label containers / volumes / networks: (empty) (empty) (empty)
anonymous node_modules volume 9c040e401f76...: removed
dangling anonymous volumes created 2026-10-14: none  (so the clone-stack and smoke runs left none either)
ports 8080 / 5173: free;  no brillio project;  worktree porcelain empty;  HEAD still 21cd239
```
Note: the anonymous `/app/node_modules` volume carries no `com.docker.compose.project` label, so the plan's label filter alone cannot see it; it was checked by name and by the dangling-anonymous sweep, and `down -v` removes it.

---

## Acceptance table

| Ticket item | Result | Proven by |
|---|---|---|
| From a clean checkout, `docker compose up` reaches healthy api and a reachable web on :5173 that proxies to api | **PASS** | Steps 1-3 (fresh clone at 21cd239, healthy in 21 s, `/healthz`, `/api/cities`, search all 200 through :5173); smoke a6-a8 (step 8) |
| `make smoke` passes and exits 0 | **PASS** | Step 8: `smoke: 8 passed, 0 failed`, exit 0 |
| `make down` leaves no containers or volumes behind | **PASS** | Step 9 (and steps 7, 8 for `down -v`): label filters empty, anonymous volume removed, ports free |
| Go and TSX edits reload without restarting compose | **PASS** | Step 4 (Go, 2 s, api StartedAt unchanged) and step 5 (TSX HMR, 1 s, `window.__bc13` kept, web StartedAt unchanged) |
| Validation: `make smoke` output + one search in the browser at :5173, attached `--role validation` | **PASS** | Step 8 (BC-13-smoke.txt) and step 6 (BC-13-browser-search.png), via the fallback below |

**browser search fallback: BC-12 SearchPage not on master; searched via the 5173 proxy URL.**

## Deviations / observations

- Raw compose commands used the `env -u DOCKER_DEFAULT_PLATFORM` prefix (accepted deviation; operator shell exports `linux/amd64`).
- `c11 browser wait --text "Northern Virginia"` timed out because the header is CSS-uppercased (innerText is "NORTHERN VIRGINIA"); the header was read with `eval` on `textContent` instead. Not a product issue.
- `make smoke` took 24 s here because the `brillio-api` image was already built by step 1; the Go module and build caches were still cold (`down -v` had wiped gomod/gobuild), as the plan intends.
- The anonymous `node_modules` volume is not project-labeled (see step 9); verified removed by other means.

## Overall verdict: PASS
