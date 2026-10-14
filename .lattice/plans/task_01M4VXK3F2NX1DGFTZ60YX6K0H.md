# BC-13: Docker Compose dev stack with hot reload, health gating, and make smoke

Spec: docs/SPEC.md §2, §3, §10, §11 (smoke checks §5.1-§5.5). Worktree:
`/Users/forrest/Code/brillio-worktrees/bc-13-compose-stack` on branch `feat/BC-13-compose-stack`,
which has BC-2 (from master) and BC-7 (still at review). Planner: agent:opus-5.5-planner.

## Scope
`docker compose up` starts two services, gated on health. `api` runs Go under air with hot reload.
`web` runs the Vite dev server with HMR and proxies `/api` and `/healthz` to api. `make up`,
`make down` and `make smoke` fill the BC-13 Makefile block. `scripts/smoke.sh` runs every assertion
the ticket lists. No Go or frontend source changes. `vite.config.ts` stays untouched: it already
reads `VITE_API_PROXY_TARGET` through `loadEnv(mode, cwd, 'VITE_')`, where the process env wins,
and it already proxies `/api` and `/healthz` (both verified).

## Decisions (verified at plan time)
- **air v1.67.4.** Newest tag (2026-08-01), per `go list -m -versions` and the proxy's `@latest`.
  Needs Go 1.26; upstream builds it with CGO_ENABLED=0.
- **Polling for file changes, in both services.** air sets `poll = true` (500 ms). Vite gets
  `CHOKIDAR_USEPOLLING=true` and `CHOKIDAR_INTERVAL=300` from compose env.
  - Why: VirtioFS on this host usually forwards inotify events, but SPEC §12 hands the stack to a
    stranger whose file sharing is unknown (gRPC FUSE, Colima). Polling is deterministic, and
    both source trees are tiny.
  - Checked: air's poller (hugo `filenotify`) watches the files inside each watched directory.
  - Checked: Vite 8.3.4 bundles chokidar 3, and these env vars override its options. So
    `vite.config.ts` does not change.
- **No `frontend/Dockerfile.dev`.** The stock `node:24-alpine` image plus `command:` is enough.
  **SPEC edit: §3.** Drop `Dockerfile.dev` from the `frontend/` listing (§10 and the ticket run
  plain `node:24-alpine`) and add `scripts/smoke.sh`. Commit it in this branch and add a
  `lattice comment` on BC-13.
- **No `.dockerignore`.** Dockerfile.dev has no COPY or ADD (the source is bind-mounted), so
  BuildKit sends no build context.
- **No `.gitignore` change.** `/backend/tmp/` (air output), `*.db`, `*.db-*` and `node_modules/`
  are already ignored. The 4 new paths are not ignored by repo or host rules (checked
  `check-ignore -v`, with and without `-c core.excludesFile=/dev/null`).
- **`init: true` on both services.** tini runs as PID 1 and forwards SIGTERM, so `down` is fast and
  air's `clean_on_exit` runs.
- **Go caches in named volumes.** `gomod` holds /go/pkg/mod (the ticket asks for it). `gobuild`
  holds /root/.cache/go-build, so rebuilds stay warm when a container is recreated. `down -v`
  deletes both on purpose (binding decision 4). The first build after `make down` is cold, and so
  is every `make smoke` run.
- **Make recipes.**
  - `up`: `docker compose up --build`. Foreground; `--build` picks up Dockerfile.dev edits (about
    1 s when cached).
  - `down`: exactly `docker compose down -v`.
  - `smoke`: `sh scripts/smoke.sh`.
- **`name: brillio`** fixes the Compose project name, so every command acts only on `brillio-*`
  containers, `brillio_default` and `brillio_*` volumes. Never prune, and never touch another
  project's containers, images or volumes (other projects run on this host).

## Files
Create `compose.yaml`, `backend/Dockerfile.dev`, `backend/.air.toml` and `scripts/smoke.sh` (git mode
100755). Modify `Makefile` (BC-13 block only) and `docs/SPEC.md` (§3 only).

**backend/Dockerfile.dev** (header comment cites SPEC §10):
```dockerfile
FROM golang:1.27-alpine
ARG AIR_VERSION=v1.67.4
ENV CGO_ENABLED=0 GOMODCACHE=/go/pkg/mod GOCACHE=/root/.cache/go-build
# Clear caches in this layer: small image, and the empty gomod/gobuild named volumes are not
# seeded with air's whole dependency tree (hugo) when first mounted.
RUN go install github.com/air-verse/air@${AIR_VERSION} && go clean -cache -modcache
WORKDIR /app
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]
```

**backend/.air.toml** (commented). The cwd is `/app`, which is the bind-mounted `backend/`. The
env paths are absolute (`/data/...`), so they resolve from any cwd.
```toml
root = "."
tmp_dir = "tmp"                     # host backend/tmp/ (gitignored); bin/ stays native make dev-api's
[build]
cmd = "go build -o ./tmp/api ./cmd/api"
entrypoint = ["./tmp/api"]
include_ext = ["go", "sql", "yaml"] # sql/yaml: embedded migrations (§4.1) and openapi.yaml (§5.7)
exclude_dir = ["tmp", "bin", "testdata", "vendor"]
exclude_regex = ["_test\\.go$"]     # test edits do not restart the server
poll = true
poll_interval = 500
delay = 200
send_interrupt = true               # SIGINT first so the server drains...
kill_delay = "2s"                   # ...then SIGKILL
log = "build-errors.log"            # lands in tmp/
[log]
time = true
[screen]
clear_on_rebuild = false            # keep compose logs readable
[misc]
clean_on_exit = true                # remove backend/tmp when the container stops
```
`.golangci.yml` is not watched (`yml` is not in `include_ext`). `stop_on_error` stays false (the
default), so a failed build keeps the last good binary serving.

**compose.yaml.** The header comment covers SPEC §10, `make up|down|smoke`, and why polling. Write
the real file in block style; flow style here is only for brevity.
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
    # 127.0.0.1: Vite binds IPv4 only; busybox may resolve localhost to ::1
    healthcheck: { test: ["CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1:5173/"],
                   interval: 10s, timeout: 3s, retries: 3, start_period: 180s, start_interval: 2s }
volumes: { gomod: {}, gobuild: {} }
```
- **Config.** These §11 values pass BC-2's `internal/config`, which validates PORT, LOG_LEVEL,
  SHUTDOWN_TIMEOUT and bare `http(s)://host[:port]` CORS origins.
- **Health timing.** Inside `start_period`, Docker probes every 2 s and failures don't count. api
  goes healthy about 2 s after its first good build. 180 s covers a cold api build after BC-4
  (module download plus the modernc sqlite compile) and a cold `npm ci`. An api that never comes
  up is marked unhealthy after about 210 s, and `depends_on` then fails `up`.
- **node_modules.** `npm ci` (npm 8+) empties a mount-point directory instead of removing it, so the
  anonymous volume works. It shadows the host's macOS-native modules. The lockfile already lists
  the `linux-*-musl` rolldown and lightningcss bindings (checked). Docker leaves an empty
  `frontend/node_modules` mount point on the host; that is the case BC-7's Makefile sentinel
  comment describes.
- **Additive, not SPEC deviations:** the web healthcheck, `init`, the polling and npm env vars, and
  `--port 5173 --strictPort`.

**Makefile.** In `# --- Compose stack (BC-13) ---`, replace only the three stub recipes and add a
one-line comment that the project name is fixed in compose.yaml. Recipe lines start with a TAB;
no GNU extensions:
- `up:` → `docker compose up --build`
- `down:` → `docker compose down -v`
- `smoke:` → `sh scripts/smoke.sh`

**scripts/smoke.sh** (about 130 lines). `#!/bin/sh` with `set -eu`; POSIX only, using curl, grep,
sed and tr (no jq, no `grep -o`). It does `cd "$(dirname "$0")/.."`, then sets
`API=http://localhost:8080` and `WEB=http://localhost:5173`.
1. **Preflight**, with no trap yet, so a stack the user already has running is never torn down.
   Exit with a clear message if any check fails:
   - `curl` is available, and `docker info` and `docker compose version` succeed.
   - `docker compose ps -a -q` is empty. Otherwise print "a brillio stack already exists; run make
     down first".
   - Ports 8080 and 5173 are free. For each port, run `rc=0; curl -s -o /dev/null --max-time 2
     http://localhost:$p/ || rc=$?`. Any rc other than 7 (connection refused) means "port $p is
     in use".
2. **Trap.** `tmp=$(mktemp -d)`, then `trap cleanup EXIT`, `trap 'exit 130' INT` and
   `trap 'exit 143' TERM`. `cleanup` saves `$?`, clears the traps, prints
   `teardown: docker compose down -v` and runs it (warn on failure, but keep the saved status),
   removes `$tmp`, then exits with the saved status. Teardown runs on pass, fail and interrupt.
3. **Bring-up.** `docker compose up -d --build --wait --wait-timeout 300 --quiet-pull`. If that
   fails, print `FAIL  stack did not become healthy`, `docker compose ps -a` and
   `docker compose logs --no-color --tail=60`, then `exit 1`. Then wait on the host side, bounded:
   `wait_for URL` runs `curl -fs -o /dev/null --max-time 2` up to 30 times, 1 s apart. Use it on
   `$API/healthz`, then `$WEB/`.
4. **Assertions.** All of them run even if earlier ones fail; none relies on errexit. Helpers:
   - `fetch`: `code=$(curl -s --max-time 10 -o "$tmp/body" -w '%{http_code}' "$1") || code=000`.
   - `compact`: `tr -d '[:space:]' < "$tmp/body"`.
   - All matching uses `grep -qF`.
   - Each assertion prints `PASS  <name>`, or `FAIL  <name> (HTTP <code>: <first 200 bytes, via dd>)`.

   | # | Request | Expect |
   |---|---|---|
   | a1 | `$API/healthz` | 200, compact body contains `"status":"ok"` |
   | a2 | `$API/api/cities` | 200, body contains `"Springfield"` |
   | a3 | `$API/api/listings/search?targetBudget=450000&city=Springfield` | 200, compact body contains `"items":[{` |
   | a4 | same body as a3 | first rank is 1: take `compact \| tr ',{' '\n\n' \| sed -n 's/^"rank":\([0-9][0-9]*\).*$/\1/p' \| head -n 1` and test `= 1` (items[0] holds the first `"rank"` in any field order) |
   | a5 | `$API/api/listings/search?minPrice=5&maxPrice=1` | 400, compact body contains `"code":"min_gt_max"` |
   | a6 | `$WEB/` | 200, body contains `<title>Brillio Listings</title>` |
   | a7 | `$WEB/healthz` (through the Vite proxy) | 200, `"status":"ok"` |
   | a8 | `$WEB/api/cities` (through the Vite proxy) | 200, `"Springfield"` |

   a7 and a8 cheaply prove "web proxies to api".
5. **Exit.** Print `smoke: N passed, M failed`, then `exit 0` only if M is 0, else `exit 1`. The
   trap prints the teardown after that.

## Shared files and commits
- **Makefile:** union rule from run-state. Edit only the BC-13 block, and keep every other block
  byte-identical through merges.
- **SPEC.md:** union. Make only the §3 edit, then post a `lattice comment`
  (`LATTICE_ROOT=/Users/forrest/Code/brillio`, `--actor agent:opus-5.5-impl`).
- **Commits:** never commit `.lattice/`. Use conventional messages, e.g.
  `build(dev): add air dev image and compose stack`, `test(dev): add make smoke script`,
  `docs(spec): drop frontend Dockerfile.dev, list scripts/smoke.sh (BC-13)`. End each with
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Gate (implementer, before review)
- Full `make check` is green. There is no stub exception: both halves are present. The first run
  may `npm ci` on the host through BC-7's sentinel.
- `sh -n scripts/smoke.sh` passes, and `shellcheck -s sh scripts/smoke.sh` is clean (shellcheck is
  installed).
- `docker compose config -q` exits 0, and `docker compose build api` builds (air installs).
- `make -n up down smoke` shows the three recipes, and `cat -et Makefile` shows `^I` on the BC-13
  recipe lines.
- `git -c core.excludesFile=/dev/null status --porcelain` lists only the intended files, and
  `git ls-files -s scripts/smoke.sh` shows mode 100755.
- An optional self-check run is allowed only after `lsof -nP -iTCP:8080 -sTCP:LISTEN` and the same
  for `:5173` print nothing. Always end it with `docker compose down -v`.

## Acceptance mapping
| Ticket item | Proven by |
|---|---|
| Clean checkout `docker compose up`: api healthy, web on :5173 proxying to api | Stage 1 steps 1-3 (fresh clone); smoke a7 and a8 |
| `make smoke` passes and exits 0 | Stage 2 step 2 (8 of 8 PASS); stage 1 step 7 covers the failure path |
| `make down` leaves no containers or volumes | Stage 1 step 6 and stage 2 step 4 (container, network and volume diffs) |
| Go and TSX edits reload without restarting compose | Stage 1 steps 4-5 (`StartedAt` unchanged); stage 2 re-checks Go |
| Validation: smoke output and a browser search, attached `--role validation` | Stage 2 steps 2-3 |

## Validation (validator sub-agent; `$S` = its scratchpad)
**Before each stage:**
- `lsof -nP -iTCP:8080 -sTCP:LISTEN` and `lsof -nP -iTCP:5173 -sTCP:LISTEN` print nothing.
- `docker compose ls -a` lists no `brillio` project.
- Save a volume snapshot: `docker volume ls -q | sort > $S/vol-before`.

Kill everything you started before you report.

**Stage 1 (now, before BC-5).**
1. Clone fresh: `git clone -q --branch feat/BC-13-compose-stack /Users/forrest/Code/brillio $S/clean`.
   A fresh clone also catches files that the host gitignore hides. If the `brillio-api` image
   exists, remove it with `docker image rm brillio-api` (this project's image only) to force a
   cold build.
2. In `$S/clean`, start `docker compose up --build` in the background, logging to `$S/up.log`. Poll
   `docker compose ps --format '{{.Service}} {{.Health}}'` until both services are healthy
   (give up after 5 min).
3. Collect evidence:
   - `docker compose ps`;
   - `curl -si localhost:8080/healthz`: 200, `{"status":"ok"}`, an `X-Request-ID` header;
   - `curl -s localhost:5173/`: contains the title;
   - `curl -si localhost:5173/healthz`: 200 through the proxy;
   - the api `listening` log line shows `:8080` and `cors_origins`;
   - `docker compose exec api env` shows `DB_PATH`, `LISTINGS_FILE` and `CORS_ORIGINS`.
4. Go reload:
   - Record `docker inspect -f '{{.State.StartedAt}}' brillio-api-1`.
   - In the clone, change the 200 branch of `backend/internal/httpapi/health.go` to
     `Status: "ok-reload"`.
   - Within about 10 s, `curl -s localhost:8080/healthz` returns `{"status":"ok-reload"}`.
   - The api logs show air's changed, building and running lines, then a new `listening` line.
   - `StartedAt` is unchanged and the container is still healthy.
5. TSX reload:
   - In a c11 browser at http://localhost:5173, evaluate `window.__bc13 = 1`.
   - Change `scope="Northern Virginia"` to `scope="Northern Virginia (HMR)"` in the clone's
     `frontend/src/App.tsx`.
   - The header updates in place, and `window.__bc13` is still 1, so there was no full reload.
   - `docker compose logs web` contains `hmr update /src/App.tsx`, and the web `StartedAt` is
     unchanged.
   - Without a browser, a weaker fallback: `curl -s localhost:5173/src/App.tsx` shows the new text.
6. Tear down: `docker compose down -v` in the clone. Then check that all of these print nothing:
   - `docker ps -a --filter label=com.docker.compose.project=brillio -q`;
   - the same filter on `docker network ls`;
   - `comm -13 $S/vol-before <(docker volume ls -q | sort)` (run it in bash or zsh).

   Then delete `$S/clean`.
7. Failure path, in the worktree: run `make smoke`. Expect PASS on a1, a6 and a7, and FAIL on a2-a5
   and a8 (404s, because BC-5 is missing). The exit status must be non-zero, and the step 6
   checks must come back clean.
8. Guard, in the worktree: run `docker compose up -d --wait`, then `make smoke`. It must fail at
   once with "already exists" and leave the stack running. Then run `make down` and the step 6
   checks.

Record stage 1 with `--role validation`, labelled "stage 1 of 2". The delegator holds at
`in_validation` and sends no LAND-READY.

**Stage 2 (after BC-5, with BC-3 and BC-4, lands on master).**
1. `git merge master` (no rebase), then re-run the gate.
2. `make smoke 2>&1 | tee $S/smoke.txt`: 8 PASS, `smoke: 8 passed, 0 failed`, exit 0. Paste the
   output. `data/listings.db` now exists, and `git status` is clean.
3. Run `make up` in the background. Once healthy, open http://localhost:5173 in a c11 browser and
   perform one search, then take a screenshot:
   - if SearchPage (BC-12) is on master, search through the UI;
   - otherwise, load
     `http://localhost:5173/api/listings/search?targetBudget=450000&city=Springfield` in the
     browser, which searches through the web origin's proxy.

   Then `touch backend/cmd/api/main.go`: air must rebuild with `StartedAt` unchanged.
4. Run `make down`, repeat the step 6 checks, and confirm both ports are free. Attach the smoke
   output and the browser evidence with `--role validation`.

## Risks
- **Bind-mount file watching on macOS.** Polling in both services mitigates it; stage 1 steps 4-5
  prove it.
- **Cold start.** The air image build takes about 1-2 min, once. After BC-4, the first api build
  also compiles modernc sqlite. `npm ci` runs on every start, and web waits for a healthy api.
  The start_period and wait-timeout cover this. Expect `make smoke` to take about 1-3 min,
  because `down -v` wipes the caches.
- **Ports or project already in use.** The smoke preflight fails fast without tearing anything
  down. BC-17 uses the same project name, so the two must never run at the same time.
- **SQLite WAL on a VirtioFS bind mount.** BC-4 uses `journal_mode(WAL)`, so this can only surface
  in stage 2. If it fails, set needs-human. Moving the DB to a named volume would be a SPEC
  §10/§11 change.
- **BC-7 not on master yet.** This branch relies on its `vite.config.ts` and its `index.html`
  title. Re-verify after any BC-7 rework is merged in.
- **Other.**
  - The stack needs network access (`npm ci`, Go modules, the air install).
  - On Linux hosts, files the containers write to `backend/tmp` and `data/` are owned by root.
  - A `COMPOSE_PROJECT_NAME` env var overrides `name:`.
  - `make down` leaves behind the `brillio-api` image, the empty `node_modules` mount point and
    `data/listings.db`. None of these is a container or a volume.

## Size
About 270 LOC: compose.yaml ~60, Dockerfile.dev ~20, .air.toml ~30, smoke.sh ~130, Makefile ~8,
SPEC ~3. Complexity: medium, and most of the effort is validation.
