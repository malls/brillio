# Ticket card — BC-4

```
TICKET=BC-4
TASK_UUID=task_01M4VXK2A1R3W99HGSVPYAJGRX
TITLE=SQLite store: schema migrations, JSON seed loader, filtered search query, cities query
WORKTREE=/Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store
BRANCH=feat/BC-4-sqlite-store
REVIEW_BASE=feat/BC-2-backend-scaffold
ANCHOR=feat/BC-2-backend-scaffold (BC-2 at review 01:47Z, head 5da9b80; not yet landed)
PORTS=API 18004 (PORT=18004, DB_PATH=/tmp/brillio-bc4.db)
AREA=area:9
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-4 notes (binding)

1. **Press-ahead anchor.** If `ANCHOR` is `feat/BC-2-backend-scaffold`, your branch was cut from BC-2's in-review
   branch. Review your own diff against `REVIEW_BASE`. When BC-2 lands, `git merge master` (never rebase) and
   switch `REVIEW_BASE` to `master`. If BC-2 gains commits before it lands, `git merge feat/BC-2-backend-scaffold`.
2. **BC-3 runs in parallel on the same anchor** and adds files under `internal/domain` only. You own
   `internal/store/**`, the seed loader, and the store/seed/Ping wiring in `cmd/api/main.go` and `httpapi.Deps`.
   Use the BC-2 domain types as they are; a needed type change goes to the Orchestrator first.
3. `modernc.org/sqlite` newest stable (1.60+) is the **only** direct Go dependency the backend may have
   (SPEC §2). Parameterized SQL only: building `IN (?, ?, ?)` placeholder text from a count is fine; any user
   value concatenated into SQL is a review FAIL (validation-plan row 27 greps for it).
4. `/healthz` 503 `{"status":"degraded","error":"..."}` and a failing-Pinger test already exist in BC-2's branch;
   you wire `Deps.Pinger` to the store in `main.go` and keep that test green. Read the BC-2 comment on your ticket:
   `Listing` is the in-memory model (never JSON-encode it on the wire); `ListedDate` is `time.Time`, parsed and
   formatted with `domain.DateLayout` at boundaries (the seed loader decodes into a boundary DTO, then converts).
5. Store tests seed the real `data/sample_listings.json` (resolve the path from the test file, for example
   `../../../data/sample_listings.json` from `internal/store`), into `:memory:` with a single connection.
6. Gate: `make lint-api test-api` (full `make check` if BC-7 has landed on master by hand-off).
7. Validation per the ticket: `PORT=18004 DB_PATH=/tmp/brillio-bc4.db LISTINGS_FILE=<worktree>/data/sample_listings.json go run ./cmd/api`
   from `backend/`; confirm the "seeded 12 listings" log, 12 rows via `sqlite3 /tmp/brillio-bc4.db 'select count(*) from listings'`,
   `/healthz` ok, restart and still 12; also start once with a copy of the data whose record 3 has
   `"status":"gone"` and capture the non-zero exit naming index 3. Remove `/tmp/brillio-bc4.db*` and kill the
   process afterwards.
