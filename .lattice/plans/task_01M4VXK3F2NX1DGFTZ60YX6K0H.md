# BC-13: Docker Compose dev stack with hot reload, health gating, and make smoke

Docker Compose dev stack with hot reload for both halves, health-gated startup, and a smoke target. Spec: docs/SPEC.md §10, §11, §2 (versions).

Why: the owner asked for Docker Compose on the latest versions, and the interview addendum says the environment must be running on the laptop with no setup time lost. One command brings up a reloading stack; `make smoke` proves it in seconds.

Scope:
- backend/Dockerfile.dev: golang:1.27-alpine, CGO_ENABLED=0, installs air, workdir /app, module cache in a named volume. backend/.air.toml watching *.go and rebuilding ./cmd/api.
- Web service on node:24-alpine (a Dockerfile.dev only if a plain image command is not enough): bind-mount ./frontend, anonymous volume for node_modules, `npm ci && npm run dev -- --host 0.0.0.0`.
- compose.yaml per §10: api with env from §11 (DB_PATH=/data/listings.db, LISTINGS_FILE=/data/sample_listings.json, CORS_ORIGINS=http://localhost:5173), ./data bind-mounted, port 8080, healthcheck wget /healthz; web with VITE_API_PROXY_TARGET=http://api:8080, port 5173, depends_on api service_healthy.
- Makefile: up (docker compose up), down (docker compose down -v), smoke.
- scripts/smoke.sh: bring the stack up detached, wait for api health (bounded retries), GET /api/cities and assert it contains Springfield, GET /api/listings/search?targetBudget=450000&city=Springfield and assert items non-empty and the first rank is 1, GET with minPrice=5&maxPrice=1 and assert 400 and code min_gt_max, GET http://localhost:5173 and assert 200 with the title Brillio Listings, then tear down. Plain sh with curl and grep; no jq dependency.
- Confirm that saving a Go file restarts api inside the container and saving a TSX file hot-reloads the page.

Acceptance:
- From a clean checkout `docker compose up` reaches healthy api and a reachable web on :5173 that proxies to api.
- `make smoke` passes and exits 0; `make down` leaves no containers or volumes behind.
- Go and TSX edits reload without restarting compose.

Validation: run `make smoke` and paste the output; open http://localhost:5173 served by compose and perform one search in the browser; attach both with --role validation.
