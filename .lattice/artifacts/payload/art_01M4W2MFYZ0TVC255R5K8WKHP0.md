# Plan Review: BC-4 — SQLite store (open, migrate, seed, search, cities, ping)

Reviewer: Claude Fable 5.1 (fresh context, headless plan review). Inputs read: the ticket card and
BC-2 comment on BC-4, docs/SPEC.md §2–§4.2, §5.1–5.2, §6.1, §8, §9.1, §11, the BC-2 scaffold at
5da9b80 in the BC-4 worktree (domain, config, httpapi, main.go, Makefile, .golangci.yml,
.gitignore), data/sample_listings.json, run-state.md (shared-file table, ports, toolchain), and the
modernc.org/sqlite v1.60.1 source in the module cache.

### 1. Verdict

**PASS** — implementation can proceed. The minor items below are clarifications the implementer
can fold in without a plan rework.

### 2. Summary

The plan covers every scope bullet and acceptance criterion in the ticket with a concrete file
list, exact SQL, a public API that satisfies `httpapi.Pinger` and what BC-5 will need, and a test
table that maps one-to-one onto the §8 store list. I independently recomputed every expected ID
set in T5, T6 and T8 against `data/sample_listings.json` (including the `(source, id)` text order
`B10 < B11 < B7 < B8 < B9`, the `pets` substring set, the combined-filter cases and the six-city
list) and all of them are correct. I also verified the driver claims against the v1.60.1 source:
a bare `:memory:?...` DSN is split at the first `?`, `_pragma=` values run on every new
connection with `busy_timeout` forced first, `_txlock=immediate` is accepted, `Ping` runs
`select 1`, and v1.60.1 is the newest tag. The only concern is wording in two places (how the
migration transaction is opened, how the MapFS migration test bootstraps `schema_migrations`)
where a literal reading could lead an implementer astray.

### 3. Issues

**[minor] Decisions §4 (Migrations) — "BEGIN; Exec the whole file; COMMIT" must mean `db.BeginTx`, not a literal `BEGIN`**
The text reads as three statements. If an implementer sends `BEGIN` as SQL via `db.ExecContext`,
the next statement can land on a different pooled connection for a file DB, and a literal `BEGIN`
is DEFERRED, which silently drops the `_txlock=immediate` guarantee the plan relies on. The lint
section's `tx.Rollback()` implies `sql.Tx`, but the migration section itself does not say so.
**Recommendation:** State explicitly: `tx, err := s.db.BeginTx(ctx, nil)`; `tx.ExecContext(ctx,
fileSQL)`; `tx.ExecContext(ctx, insert, ...)`; `tx.Commit()`; migration files must not contain
`BEGIN`/`COMMIT`. Same wording for Upsert (`BeginTx` + `tx.PrepareContext`).

**[minor] Tests I3 (`TestMigrateOrderAndRollback`) — say how the MapFS fixture bootstraps `schema_migrations`**
The runner's "applied set" rule is: no `schema_migrations` table means nothing applied, and the
embedded `0001_listings.sql` is what creates that table. The I3 fixture therefore has two valid
shapes and the plan does not pick one: (a) run `migrate` on a store from `Open` — then
`schema_migrations` already exists and holds `0001_listings.sql`, so the fixture must not create
the table and its filenames must not collide; or (b) construct `&Store{db: sql.Open(...)}` directly —
then the fixture's first file must create `schema_migrations` itself, or every run fails on the
bookkeeping insert. An implementer who mixes the two gets "table schema_migrations already exists"
or "no such table" and loses time.
**Recommendation:** Pick (b): a raw `:memory:` `*sql.DB` wrapped in `&Store{}` with the fixture's
`0001_*.sql` creating `schema_migrations`. It exercises the bootstrap path independently of the
embedded files, and the "rerun with 0003 fixed" step is then a second `migrate` call with a second
MapFS against the same store.

**[minor] Tests I2 (`TestMigrateTwice`) — assert the four indexes by name, not by count**
`PRIMARY KEY (source, id)` on a rowid table creates `sqlite_autoindex_listings_1`, so
`sqlite_master` holds five index rows, not four.
**Recommendation:** Query `SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='listings'`
and assert the set contains the four named indexes.

**[minor] Decisions §3 (Open) — the `?` guard should also reject a `file:` prefix**
The driver opens with `SQLITE_OPEN_URI` and keeps the whole string when it starts with `file:`,
so `DB_PATH=file:x.db` would be parsed as a URI (percent-decoding, `#` fragment handling) rather
than as the plain path the plan's `DB_PATH must be a file path or :memory:` message promises.
**Recommendation:** Extend the same guard with `strings.HasPrefix(path, "file:")`; add it to T11.
One line; optional if the orchestrator prefers to leave operator-config edge cases alone.

**[minor] Decisions §11 (Cities) — name the implicit-rowid dependency in Risks**
"First-seen = lowest rowid" is correct for this schema (composite TEXT key, so rowid is implicit
and `DO UPDATE` preserves it), but SQLite documents that `VACUUM` may renumber implicit rowids.
Nothing in this project vacuums, so it is not a defect, only an unstated assumption.
**Recommendation:** One line under Risks: "Cities first-seen order depends on the implicit rowid;
a VACUUM could renumber it. Not run anywhere in v1."

**[minor] Tests T2 (`TestRoundTrip`) — compare `ListedDate` with `Equal`, not `reflect.DeepEqual`**
Both sides come from `time.Parse(DateLayout, ...)` so `DeepEqual` will pass today, but a `time.Time`
deep-equality is brittle (location pointer, monotonic bits) and the failure mode is a confusing
diff on otherwise identical dates.
**Recommendation:** Compare every field, using `ListedDate.Equal` for the date; or zero the
monotonic reading on both sides first. Trivial.

**[minor, optional] Decisions §5/§6 (RecordError) — consider carrying `(source, id)` when they decoded**
For a validation failure (as opposed to a decode failure) the record's `source` and `id` are known,
and `record index 3 (MLS_B/B8): field "status": ...` is far more actionable in a fatal log than an
index alone. This would change the agreed message text that validation step 4 and the ticket card
grep for (`record index 3: field "status"`), so it is an evolutionary change, not a gap.
**Recommendation:** Defer unless the orchestrator wants it now; if adopted, append the identity
after the field segment (`... "gone" (MLS_B/B8)`) so the existing substring checks still hold.

### 4. Positive Observations

- **Expected values are right.** Every ID set in the T5/T6/T8 table, the 0-based index 3 = `MLS_B/B8`
  mapping, and the ORDER BY text order were recomputed from the sample file and match. This is the
  single biggest de-risking a plan can do for a table-driven test suite and it was done carefully.
- **Driver behaviour verified, not assumed.** The DSN split, per-connection pragmas, `_txlock`,
  multi-statement `Exec`, TEXT-column scanning and the version check all hold against v1.60.1
  source. The plan says what was probed and why each setting matters (`_txlock=immediate` for WAL
  lock upgrades, `busy_timeout` ordering).
- **Single-connection deadlock awareness.** Reading the applied set before `BeginTx` and never
  querying while a `Rows` is open is exactly the trap a `:memory:` store with `MaxOpenConns(1)` sets,
  and it is called out in both Decisions and Risks.
- **Boundary discipline matches the BC-2 comment.** Dates are formatted and parsed with
  `domain.DateLayout` at the SQL boundary and in a seed DTO; `Listing` is never JSON-encoded; a
  bound `time.Time` is explicitly avoided.
- **Failure-first startup order.** `seed.Load` before `store.Open` so a bad file never creates a DB,
  with validation step 4 proving it (`test ! -e` on the DB path) and the exact fatal text derived
  from `ParseStatus` plus JSON escaping.
- **Acceptance mapped to proof.** I4 pins the exact SQL text and arg order, validation step 1 greps
  for concatenation, T7 adds injection probes, and T6 proves `instr` vs `LIKE` with a planted
  `100% ... pet_friendly` row.
- **Shared-file discipline.** The plan reproduces the run-state table: main.go confined to the
  seed/store block and `Pinger: st`, `Deps`/router untouched, domain untouched, go.mod ownership,
  SPEC edit by the deviation rule with a `lattice comment`, `.lattice/` never committed, and the
  BC-2 press-ahead merge rule (merge master, never rebase).
- **Proportional.** About 265 lines, inside the brief's ~300-line cap for a large ticket, with a
  LOC estimate and four conventional commits.
