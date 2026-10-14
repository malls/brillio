# BC-4: SQLite store: schema migrations, JSON seed loader, filtered search query, cities query

SQLite persistence: open, migrate, seed from JSON, filtered search query, cities query, ping. Spec: docs/SPEC.md §4.1, §4.2, §5.2, §6.1, §8 (store list), §11.

Why: the handout wants real logic over the dataset; SQLite gives a real WHERE clause (filters) while ranking stays in Go (BC-3). Seeding must be idempotent because the container restarts on every hot reload.

Scope (package internal/store, plus seed loading):
- `Open(ctx, dsn string) (*Store, error)` using modernc.org/sqlite (pure Go). Pragmas: busy_timeout=5000, foreign_keys=on, journal_mode=WAL when not :memory:. Single connection for :memory: so the schema is shared.
- Embedded migrations under internal/store/migrations/ (0001_listings.sql per §4.1 including schema_migrations and the four indexes). `Migrate(ctx)` applies pending files in order inside a transaction; running it twice is a no-op.
- `seed.Load(path string) ([]domain.Listing, error)`: strict JSON decode (DisallowUnknownFields), per-record validation and normalization per §4.2; errors name the record index and field. `(*Store).Upsert(ctx, []domain.Listing) (int, error)` in one transaction with ON CONFLICT(source, id) DO UPDATE.
- `(*Store).Search(ctx, domain.SearchFilter) ([]domain.Listing, error)`: builds the WHERE clause per §6.1 with parameterized args only; keyword tokens use `instr(lower(description), ?) > 0`; status uses IN with one placeholder per value; ORDER BY source, id.
- `(*Store).Cities(ctx) ([]string, error)` distinct case-insensitively, sorted, first-seen casing.
- `(*Store).Ping(ctx) error`; wire into main.go so /healthz returns 503 {"status":"degraded"} when it fails. main.go opens the store, migrates, seeds from LISTINGS_FILE, and exits non-zero with a clear log on any failure.
- Tests per §8 store list against :memory: using the real data/sample_listings.json: 12 rows after seed, 12 after seeding twice, each filter alone and combined, keyword tokens containing % and _ are literal, multi-token AND, status sets, city case-insensitive and trimmed, Cities output exact.

Acceptance:
- `go run ./cmd/api` against a fresh DB_PATH logs "seeded 12 listings" and /healthz is ok.
- Every filter in §6.1 has a passing test; `make check` green for the backend.
- No string concatenation of user input into SQL anywhere (reviewer greps for it).

Validation: start the binary with DB_PATH=/tmp/x.db, confirm 12 rows (sqlite3 CLI or a one-off go run), curl /healthz, restart and confirm still 12; attach the output with --role validation.
