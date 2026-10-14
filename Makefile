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
dev-api:
	@echo 'not yet implemented: dev-api (lands in BC-2)' >&2; exit 1
lint-api:
	@echo 'not yet implemented: lint-api (lands in BC-2)' >&2; exit 1
test-api:
	@echo 'not yet implemented: test-api (lands in BC-2)' >&2; exit 1

# --- Frontend (BC-7) ---
.PHONY: dev-web lint-web test-web
dev-web:
	@echo 'not yet implemented: dev-web (lands in BC-7)' >&2; exit 1
lint-web:
	@echo 'not yet implemented: lint-web (lands in BC-7)' >&2; exit 1
test-web:
	@echo 'not yet implemented: test-web (lands in BC-7)' >&2; exit 1

# --- Compose stack (BC-13) ---
.PHONY: up down smoke
up:
	@echo 'not yet implemented: up (lands in BC-13)' >&2; exit 1
down:
	@echo 'not yet implemented: down (lands in BC-13)' >&2; exit 1
smoke:
	@echo 'not yet implemented: smoke (lands in BC-13)' >&2; exit 1

# --- Release (BC-17) ---
.PHONY: dist
dist:
	@echo 'not yet implemented: dist (lands in BC-17)' >&2; exit 1
