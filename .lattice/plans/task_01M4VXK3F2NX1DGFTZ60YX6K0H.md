# BC-13: Docker Compose dev stack with hot reload, health gating, and make smoke

**Spec:** docs/SPEC.md §2, §3, §10 and §11 (the smoke assertions follow §5.1-§5.5).
**Worktree:** `/Users/forrest/Code/brillio-worktrees/bc-13-compose-stack`, branch
`feat/BC-13-compose-stack`. It carries BC-2 from master and BC-7, which is still at review.
**Planner:** agent:opus-5.5-planner.

## Scope
`docker compose up` starts two health-gated services:
- `api` runs Go under air and reloads on Go edits.
- `web` runs the Vite dev server with HMR and proxies `/api` and `/healthz` to `api`.

This ticket fills the BC-13 Makefile block (`up`, `down`, `smoke`) and adds `scripts/smoke.sh`,
which runs every assertion the ticket lists. No Go or frontend source changes. `vite.config.ts`
stays untouched: it already reads `VITE_API_PROXY_TARGET` through
`loadEnv(mode, cwd, 'VITE_')` (the process env wins) and proxies both `/api` and `/healthz`
(both verified).

## Decisions (verified at plan time)
- **air v1.67.4.** It is the newest tag (2026-08-01) per `go list -m -versions` and the module
  proxy's `@latest`. It needs Go 1.26 or later, and upstream builds it with `CGO_ENABLED=0`.
- **Polling in both services.** air sets `poll = true` (500 ms). Vite gets
  `CHOKIDAR_USEPOLLING=true` and `CHOKIDAR_INTERVAL=300` from the compose env.
  - **Why:** this host's VirtioFS usually forwards inotify events, but SPEC §12 hands the stack
    to a stranger with unknown file sharing (gRPC FUSE, Colima). Both trees are tiny, so
    polling costs little and behaves the same everywhere.
  - **Checked:** air's poller (hugo `filenotify`) also watches the files inside each watched
    directory. Vite 8.3.4 bundles chokidar 3, whose env vars override its options, so
    `vite.config.ts` does not change.
- **No `frontend/Dockerfile.dev`.** The stock `node:24-alpine` image plus a `command:` is enough.
  - **SPEC edit: §3.** Drop `Dockerfile.dev` from the `frontend/` listing, because §10 and the
    ticket run plain `node:24-alpine`. Add `scripts/smoke.sh` to the listing.
  - Commit the edit in this branch and record it in a `lattice comment`.
- **No `.dockerignore`.** Dockerfile.dev has no COPY or ADD (the source is bind-mounted), so
  BuildKit sends no build context.
- **No `.gitignore` change.** `/backend/tmp/` (air output), `*.db`, `*.db-*` and `node_modules/` are
  already ignored. None of the four new paths is ignored by repo rules or by the host's global
  ignore file. Checked with `check-ignore -v`, with and without `-c core.excludesFile=/dev/null`.
- **`init: true` on both services.** tini runs as PID 1 and forwards SIGTERM, so `down` is fast
  and air's `clean_on_exit` runs.
- **Go caches in named volumes.**
  - `gomod` at /go/pkg/mod, which the ticket asks for.
  - `gobuild` at /root/.cache/go-build, so rebuilds stay warm when a container is recreated.

  `down -v` deletes both on purpose (binding decision 4). So the first build after
  `make down`, and every `make smoke`, starts cold.
- **Make recipes.**
  - `up`: `docker compose up --build`. It runs in the foreground, and `--build` picks up
    Dockerfile.dev edits for about 1 s when cached.
  - `down`: exactly `docker compose down -v`.
  - `smoke`: `sh scripts/smoke.sh`.
- **Fixed project name: `name: brillio`.** Commands touch only `brillio-*` containers,
  `brillio_default` and `brillio_*` volumes. Never prune, and never touch another project's
  containers, images or volumes: other projects run on this host.

## Files
Create `compose.yaml`, `backend/Dockerfile.dev`, `backend/.air.toml` and `scripts/smoke.sh` (git
mode 100755). Modify only the BC-13 block of `Makefile` and §3 of `docs/SPEC.md`.

**backend/Dockerfile.dev** (its header comment cites SPEC §10):
```dockerfile
FROM golang:1.27-alpine
ARG AIR_VERSION=v1.67.4
ENV CGO_ENABLED=0 GOMODCACHE=/go/pkg/mod GOCACHE=/root/.cache/go-build
# Clear caches in this same layer. The image stays small, and the empty gomod/gobuild named
# volumes are not seeded with air's whole dependency tree (hugo) when first mounted.
RUN go install github.com/air-verse/air@${AIR_VERSION} && go clean -cache -modcache
WORKDIR /app
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]
```

**backend/.air.toml** (commented). The working directory is `/app`, the bind-mounted `backend/`.
Env paths are absolute (`/data/...`), so they resolve from any cwd.
```toml
root = "."
tmp_dir = "tmp"                     # host backend/tmp/, gitignored; bin/ stays with native make dev-api
[build]
cmd = "go build -o ./tmp/api ./cmd/api"
entrypoint = ["./tmp/api"]
include_ext = ["go", "sql", "yaml"] # sql/yaml: embedded migrations (§4.1) and openapi.yaml (§5.7)
exclude_dir = ["tmp", "bin", "testdata", "vendor"]
exclude_regex = ["_test\\.go$"]     # editing a test does not restart the server
poll = true
poll_interval = 500
delay = 200
send_interrupt = true               # SIGINT first, so the server drains gracefully
kill_delay = "2s"                   # SIGKILL if it is still running after 2s
log = "build-errors.log"            # written inside tmp/
[misc]
clean_on_exit = true                # remove backend/tmp when the container stops
```
- Also set `[log] time = true` and `[screen] clear_on_rebuild = false`, which keeps compose logs
  readable.
- `.golangci.yml` is not watched, because its extension is `yml`, not `yaml`.
- Leave `stop_on_error` at its default (false), so a failed build keeps the last good binary
  serving.

**compose.yaml.** The header comment covers SPEC §10, `make up|down|smoke` and the reason for
polling. Write the real file in block style; flow style here only keeps the plan short.
```yaml
name: brillio
services:
  api:
    build: { context: ./backend, dockerfile: Dockerfile.dev }
    init: true
    environment: # every SPEC §11 var, explicit
      { PORT: "8080", DB_PATH: /data/listings.db, LISTINGS_FILE: /data/sample_listings.json,
        LOG_LEVEL: info, CORS_ORIGINS: http://localhost:5173, SHUTDOWN_TIMEOUT: 10s }
    volumes: [./backend:/app, ./data:/data, gomod:/go/pkg/mod, gobuild:/root/.cache/go-build]
    ports: ["8080:8080"]
    healthcheck: { test: ["CMD", "wget", "-qO-", "http://localhost:8080/healthz"],
                   interval: 10s, timeout: 3s, retries: 3, start_period: 180s, start_interval: 2s }
  web:
    image: node:24-alpine
    init: true
    working_dir: /app
    command: ["sh", "-c", "npm ci && exec npm run dev -- --host 0.0.0.0 --port 5173 --strictPort"]
    environment:
      { VITE_API_PROXY_TARGET: http://api:8080, CHOKIDAR_USEPOLLING: "true", CHOKIDAR_INTERVAL: "300",
        NPM_CONFIG_AUDIT: "false", NPM_CONFIG_FUND: "false", NPM_CONFIG_UPDATE_NOTIFIER: "false" }
    volumes: [./frontend:/app, /app/node_modules]
    ports: ["5173:5173"]
    depends_on: { api: { condition: service_healthy } }
    # 127.0.0.1, not localhost: Vite binds IPv4 only, and busybox may resolve localhost to ::1
    healthcheck: { test: ["CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1:5173/"],
                   interval: 10s, timeout: 3s, retries: 3, start_period: 180s, start_interval: 2s }
volumes: { gomod: {}, gobuild: {} }
```
- **Config check.** These values pass BC-2's `internal/config`: PORT, LOG_LEVEL and
  SHUTDOWN_TIMEOUT are valid, and the CORS entry is a bare origin.
- **Health timing.** During `start_period`, probes run every 2 s and failures do not count, so
  `api` turns healthy about 2 s after its first good build. The 180 s covers a cold post-BC-4
  build (modernc sqlite download and compile) and a cold `npm ci`. An `api` that never comes up
  turns unhealthy after about 210 s, and `depends_on` then fails `up`.
- **node_modules volume.** `npm ci` (npm 8 and later) empties a mount-point directory rather than
  removing it, so the anonymous volume works.
  - The volume shadows the host's macOS-native modules.
  - The lockfile already lists the `linux-*-musl` rolldown and lightningcss bindings (checked).
  - Docker leaves an empty `frontend/node_modules` mount point on the host, which is exactly the
    case BC-7's Makefile sentinel comment covers.
- **Not SPEC deviations.** The web healthcheck, `init`, the polling and npm env vars, and
  `--port 5173 --strictPort` are all additive.

**Makefile.** Edit only the `# --- Compose stack (BC-13) ---` block:
- Replace the three stub recipes:

  | Target | Recipe |
  |---|---|
  | `up` | `docker compose up --build` |
  | `down` | `docker compose down -v` |
  | `smoke` | `sh scripts/smoke.sh` |
- Add a one-line comment saying the project name is fixed in compose.yaml.
- Start every recipe line with a TAB, and use no GNU extensions.

**scripts/smoke.sh** (about 130 lines). It starts with `#!/bin/sh` and `set -eu` and stays POSIX:
only curl, grep, sed and tr, with no jq and no `grep -o`. It runs `cd "$(dirname "$0")/.."` and sets
`API=http://localhost:8080` and `WEB=http://localhost:5173`.

1. **Preflight.** No trap is set yet, so a stack the user already has running is never torn down.
   Each check exits with a clear message:
   - curl is installed, and `docker info` and `docker compose version` succeed.
   - `docker compose ps -a -q` is empty. Otherwise exit with "a brillio stack already exists;
     run make down first".
   - Ports 8080 and 5173 are free. For each port `p`, run
     `rc=0; curl -s -o /dev/null --max-time 2 http://localhost:$p/ || rc=$?`. Any rc other than 7
     (connection refused) means "port $p is in use".
2. **Trap.** Create `tmp=$(mktemp -d)`, then set `trap cleanup EXIT`, `trap 'exit 130' INT` and
   `trap 'exit 143' TERM`. `cleanup` must:
   - save `$?` and clear the traps;
   - print `teardown: docker compose down -v` and run it; on failure, warn but keep the saved
     status;
   - remove `$tmp` and exit with the saved status.

   Teardown therefore runs on pass, on fail and on interrupt.
3. **Bring-up.** Run `docker compose up -d --build --wait --wait-timeout 300 --quiet-pull`. If it
   fails, print `FAIL  stack did not become healthy`, then the output of `docker compose ps -a`
   and `docker compose logs --no-color --tail=60`, and `exit 1`. Then run the host-side
   `wait_for URL` check (`curl -fs -o /dev/null --max-time 2`, up to 30 tries, 1 s apart) on
   `$API/healthz` and `$WEB/`.
4. **Assertions.** All of them run, and none relies on errexit.
   - `fetch` sets `code=$(curl -s --max-time 10 -o "$tmp/body" -w '%{http_code}' "$1") || code=000`.
   - `compact` is `tr -d '[:space:]' < "$tmp/body"`.
   - Matching uses `grep -qF`.
   - Each assertion prints `PASS  <name>` or `FAIL  <name> (HTTP <code>: <first 200 bytes, via dd>)`.

   | # | Request | Expect |
   |---|---|---|
   | a1 | `$API/healthz` | 200, compact body contains `"status":"ok"` |
   | a2 | `$API/api/cities` | 200, contains `"Springfield"` |
   | a3 | `$API/api/listings/search?targetBudget=450000&city=Springfield` | 200, compact body contains `"items":[{` |
   | a4 | same body as a3 | first rank is 1 (see below) |
   | a5 | `$API/api/listings/search?minPrice=5&maxPrice=1` | 400, compact body contains `"code":"min_gt_max"` |
   | a6 | `$WEB/` | 200, contains `<title>Brillio Listings</title>` |
   | a7 | `$WEB/healthz` via the Vite proxy | 200, `"status":"ok"` |
   | a8 | `$WEB/api/cities` via the Vite proxy | 200, `"Springfield"` |

   For a4, take the first rank in the document, which belongs to `items[0]` in any field
   order, and require `= 1`:
   `compact | tr ',{' '\n\n' | sed -n 's/^"rank":\([0-9][0-9]*\).*$/\1/p' | head -n 1`.
   a7 and a8 prove "web proxies to api" at almost no extra cost.
5. **Exit.** Print `smoke: N passed, M failed`, then `exit 0` only if M is 0, else `exit 1`.

## Shared files and commits
- **Makefile.** Follow the run-state union rule: change the BC-13 block only, and keep every other
  block byte-identical across merges.
- **docs/SPEC.md.** Make the §3 edit only, then post a `lattice comment`
  (`LATTICE_ROOT=/Users/forrest/Code/brillio`, `--actor agent:opus-5.5-impl`).
- **Commits.** Never commit `.lattice/`. Use conventional messages, for example:
  - `build(dev): add air dev image and compose stack`
  - `test(dev): add make smoke script`
  - `docs(spec): drop frontend Dockerfile.dev, list scripts/smoke.sh (BC-13)`

  End each message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Gate (implementer, before review)
1. Run the full `make check`. There is no stub exception, because both halves exist.
2. `sh -n scripts/smoke.sh` passes, and `shellcheck -s sh scripts/smoke.sh` reports nothing
   (shellcheck is installed).
3. `docker compose config -q` passes, and `docker compose build api` succeeds.
4. `make -n up down smoke` prints the three recipes, and `cat -et Makefile` shows `^I` on the
   BC-13 recipe lines.
5. `git -c core.excludesFile=/dev/null status --porcelain` lists only the intended files, and
   `git ls-files -s scripts/smoke.sh` shows mode 100755.
6. An optional self-check run is allowed only after `lsof -nP -iTCP:8080 -sTCP:LISTEN` and the
   same check on 5173 print nothing. Always end it with `docker compose down -v`.

## Acceptance mapping
| Ticket item | Proven by |
|---|---|
| A clean checkout's `docker compose up` gives a healthy api and web on :5173 proxying to api | Stage 1 steps 1-3 (fresh clone); smoke a7 and a8 |
| `make smoke` passes and exits 0 | Stage 2 step 2 (8/8 PASS); stage 1 step 7 exercises the failure path |
| `make down` leaves no containers or volumes | Stage 1 step 6 and stage 2 step 4 |
| Go and TSX edits reload without restarting compose | Stage 1 steps 4-5 (container `StartedAt` unchanged); stage 2 step 3 |
| Validation: smoke output and a browser search, attached with `--role validation` | Stage 2 steps 2-3 |

## Validation (validator sub-agent; `$S` = its scratchpad)
**Before each stage:**
- `lsof -nP -iTCP:8080 -sTCP:LISTEN` and the same check on 5173 print nothing.
- `docker compose ls -a` shows no `brillio` project.
- Record the volumes: `docker volume ls -q | sort > $S/vol-before`.

Kill everything you started before reporting.

**Stage 1 (now, before BC-5).**
1. **Fresh clone.** Run
   `git clone -q --branch feat/BC-13-compose-stack /Users/forrest/Code/brillio $S/clean`. A fresh
   clone also catches files hidden by the host gitignore. If the `brillio-api` image exists,
   remove it with `docker image rm brillio-api` to force a cold build (it is this project's
   image only).
2. **Start the stack.** In `$S/clean`, run `docker compose up --build` in the background, logging
   to `$S/up.log`. Poll `docker compose ps --format '{{.Service}} {{.Health}}'` until both
   services are healthy, for at most 5 minutes.
3. **Evidence:**
   - `docker compose ps`;
   - `curl -si localhost:8080/healthz` returns 200, `{"status":"ok"}` and an `X-Request-ID`;
   - `curl -s localhost:5173/` contains the page title;
   - `curl -si localhost:5173/healthz` returns 200 through the proxy;
   - the api `listening` log line shows `:8080` and `cors_origins`;
   - `docker compose exec api env` shows `DB_PATH`, `LISTINGS_FILE` and `CORS_ORIGINS`.
4. **Go reload.**
   - Record `docker inspect -f '{{.State.StartedAt}}' brillio-api-1`.
   - In the clone, make the 200 branch of `backend/internal/httpapi/health.go` return
     `Status: "ok-reload"`.
   - Within about 10 s, `/healthz` returns `{"status":"ok-reload"}`.
   - The logs show air's change, build and run lines and a new `listening` line.
   - `StartedAt` is unchanged, and `api` is still healthy.
5. **TSX reload.**
   - Open http://localhost:5173 in a c11 browser and evaluate `window.__bc13 = 1`.
   - In the clone's `frontend/src/App.tsx`, change `scope="Northern Virginia"` to
     `scope="Northern Virginia (HMR)"`.
   - The header updates in place and `window.__bc13` is still 1, so there was no full reload.
   - `docker compose logs web` contains `hmr update /src/App.tsx`, and the web container's
     `StartedAt` is unchanged.
   - Weaker fallback if no browser is available: `curl -s localhost:5173/src/App.tsx` shows the
     new text.
6. **Teardown.** Run `docker compose down -v` in the clone. Then all three of these must print
   nothing:
   - `docker ps -a --filter label=com.docker.compose.project=brillio -q`;
   - `docker network ls` with the same filter;
   - `comm -13 $S/vol-before <(docker volume ls -q | sort)`, run in bash or zsh.

   Then delete `$S/clean`.
7. **Failure path.** In the worktree, run `make smoke`. Expect PASS on a1, a6 and a7, and FAIL on
   a2-a5 and a8, which return 404 because BC-5 is missing. The exit status is non-zero, and the
   step 6 checks are clean afterwards.
8. **Guard.** In the worktree, run `docker compose up -d --wait`, then `make smoke`. Smoke must
   fail at once with "already exists" and leave the stack running. Then run `make down` and the
   step 6 checks.

Record stage 1 with `--role validation`, labeled "stage 1 of 2". The delegator then holds at
`in_validation` without sending LAND-READY.

**Stage 2 (after BC-5, with BC-3 and BC-4, lands on master).**
1. Run `git merge master` (no rebase), then re-run the gate.
2. Run `make smoke 2>&1 | tee $S/smoke.txt`. Expect 8 PASS, `smoke: 8 passed, 0 failed` and exit 0.
   Paste the output. `data/listings.db` should now exist, and `git status` should be clean.
3. Run `make up` in the background. Once healthy, open http://localhost:5173 in a c11 browser and
   do one search: use the UI if SearchPage (BC-12) is on master. Otherwise load
   `http://localhost:5173/api/listings/search?targetBudget=450000&city=Springfield` in the
   browser, which searches through the web origin's proxy. Take a screenshot. Then run
   `touch backend/cmd/api/main.go`: air rebuilds, and `StartedAt` is unchanged.
4. Run `make down`, repeat the step 6 checks and confirm both ports are free. Attach the smoke
   output and the browser evidence with `--role validation`.

## Risks
- **Bind-mount file watching on macOS.** Polling in both services mitigates it; stage 1 steps 4-5
  prove it works.
- **Cold start.**
  - The air image build takes 1-2 minutes, but only once.
  - After BC-4, the first api build also compiles modernc sqlite.
  - `npm ci` runs on every start, and `web` waits for `api` to be healthy.

  A 180 s start period and a 300 s wait timeout cover this. Expect `make smoke` to take
  1-3 minutes, because `down -v` wipes the caches.
- **Ports or project in use.** Smoke fails fast and does not tear anything down. BC-17 uses the
  same project name, so the two must never run at the same time.
- **SQLite WAL on a VirtioFS bind mount.** BC-4 uses `journal_mode(WAL)`, so this can only show up
  in stage 2. If it fails, set needs-human: moving the DB to a named volume would be a SPEC
  §10/§11 change.
- **BC-7 is not on master yet.** This ticket relies on BC-7's `vite.config.ts` and its
  `index.html` title. Re-verify after any BC-7 rework is merged in.
- **Other.**
  - The stack needs network access for `npm ci`, Go modules and the air install.
  - On Linux hosts, root owns the files the containers write to `backend/tmp` and `data/`.
  - A `COMPOSE_PROJECT_NAME` env var overrides `name:`.
  - After `make down`, the `brillio-api` image, the empty `node_modules` mount point and
    `data/listings.db` remain. None of these is a container or a volume.

## Size
About 270 LOC: compose.yaml ~60, Dockerfile.dev ~20, .air.toml ~30, smoke.sh ~130, Makefile ~8,
SPEC ~3. Complexity: medium; most of the work is validation.
