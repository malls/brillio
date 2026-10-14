# Brillio Listings - developer entry points (docs/SPEC.md sections 3, 8, 10, 12).
# `make` alone runs `make check` (lint, then tests, both halves).
# A stub prints "not yet implemented: <target> (lands in BC-N)" to stderr and
# fails, so the gate fails loudly until ticket BC-N replaces its recipe.
# One block per owning ticket: edit only your own block (including its own
# .PHONY line) so parallel branches merge cleanly. Plain POSIX make, no GNU
# extensions (host make is GNU Make 3.81). Recipe lines start with a TAB.

# --- Gate (BC-1): composite targets, prerequisites only, no recipes ---
.PHONY: check lint test
check: lint test
lint: lint-api lint-web
test: test-api test-web

# --- Backend (BC-2) ---
.PHONY: dev-api lint-api test-api
# dev-api runs the binary from the repo root so the ./data/... defaults resolve.
dev-api:
	cd backend && go build -o bin/api ./cmd/api
	./backend/bin/api
lint-api:
	cd backend && golangci-lint run
test-api:
	cd backend && go test -race ./...

# --- Frontend (BC-7) ---
# Recipes run in frontend/. The sentinel frontend/node_modules/.package-lock.json (written
# by npm) runs `npm ci` when missing or older than the lockfile; an empty node_modules
# directory (e.g. a Compose anonymous volume's mount point) does not count as installed.
.PHONY: dev-web lint-web test-web
dev-web: frontend/node_modules/.package-lock.json
	cd frontend && npm run dev
lint-web: frontend/node_modules/.package-lock.json
	cd frontend && npm run lint && npm run typecheck && npm run format:check
test-web: frontend/node_modules/.package-lock.json
	cd frontend && npm test
frontend/node_modules/.package-lock.json: frontend/package-lock.json
	cd frontend && npm ci
	touch frontend/node_modules/.package-lock.json

# --- Compose stack (BC-13) ---
# compose.yaml fixes the project name (brillio), so these touch only its own objects.
# up unsets DOCKER_DEFAULT_PLATFORM so the stack runs native (see compose.yaml).
.PHONY: up down smoke
up:
	unset DOCKER_DEFAULT_PLATFORM; docker compose up --build
down:
	docker compose down -v
smoke:
	@echo 'not yet implemented: smoke (lands in BC-13)' >&2; exit 1

# --- Release (BC-17) ---
.PHONY: dist
dist:
	@echo 'not yet implemented: dist (lands in BC-17)' >&2; exit 1
