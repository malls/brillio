package store

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"brillio/backend/internal/domain"
)

// openMemory opens a migrated ":memory:" store and closes it at cleanup.
func openMemory(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.Context(), memoryPath)
	if err != nil {
		t.Fatalf("Open(:memory:) error = %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return s
}

// pragmaQuerier is satisfied by *sql.DB and *sql.Conn.
type pragmaQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// checkPragmas asserts the per-connection settings Open's DSN configures.
func checkPragmas(t *testing.T, q pragmaQuerier, wantJournal string) {
	t.Helper()
	ctx := t.Context()
	var busy, fk int
	var journal string
	if err := q.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busy); err != nil {
		t.Fatalf("PRAGMA busy_timeout: %v", err)
	}
	if err := q.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if err := q.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journal); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if busy != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busy)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
	if journal != wantJournal {
		t.Errorf("journal_mode = %q, want %q", journal, wantJournal)
	}
}

func TestOpenPragmas(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		s := openMemory(t)
		checkPragmas(t, s.db, "memory")
		if got := s.db.Stats().MaxOpenConnections; got != 1 {
			t.Errorf("MaxOpenConnections = %d, want 1", got)
		}
	})

	t.Run("file, every pooled connection", func(t *testing.T) {
		ctx := t.Context()
		s, err := Open(ctx, filepath.Join(t.TempDir(), "listings.db"))
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		t.Cleanup(func() {
			if err := s.Close(); err != nil {
				t.Errorf("Close() error = %v", err)
			}
		})
		// Hold two connections at once so the pool must open a second one.
		var conns []*sql.Conn
		for range 2 {
			c, err := s.db.Conn(ctx)
			if err != nil {
				t.Fatalf("Conn() error = %v", err)
			}
			conns = append(conns, c)
		}
		for i, c := range conns {
			t.Run(fmt.Sprintf("conn %d", i+1), func(t *testing.T) { checkPragmas(t, c, "wal") })
		}
		for _, c := range conns {
			if err := c.Close(); err != nil {
				t.Errorf("Conn.Close() error = %v", err)
			}
		}
	})
}

// queryStrings returns the single text column of every row of query.
func queryStrings(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), query)
	if err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("scan %q: %v", query, err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows %q: %v", query, err)
	}
	return out
}

func TestMigrateTwice(t *testing.T) {
	s := openMemory(t) // Open has migrated once already
	if err := s.Migrate(t.Context()); err != nil {
		t.Fatalf("second Migrate() error = %v, want nil", err)
	}
	if got, want := queryStrings(t, s.db, "SELECT filename FROM schema_migrations ORDER BY filename"),
		[]string{"0001_listings.sql"}; !slices.Equal(got, want) {
		t.Errorf("schema_migrations = %q, want %q", got, want)
	}
	indexes := queryStrings(t, s.db,
		"SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'listings'")
	for _, want := range []string{"listings_city_nocase", "listings_price", "listings_bedrooms", "listings_status"} {
		if !slices.Contains(indexes, want) {
			t.Errorf("index %q missing; have %q", want, indexes)
		}
	}
}

func TestMigrateOrderAndRollback(t *testing.T) {
	ctx := t.Context()
	// A raw, unmigrated store, so the fixtures below bootstrap the schema.
	db, err := sql.Open(driverName, dsn(memoryPath))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	file := func(sql string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(sql)} }
	first := fstest.MapFS{
		"0001_init.sql": file(`CREATE TABLE schema_migrations (filename TEXT PRIMARY KEY, applied_at TEXT NOT NULL);
CREATE TABLE t (v TEXT NOT NULL);`),
		// Inserting into 0001's table only works if 0001 ran first.
		"0002_insert.sql": file(`INSERT INTO t (v) VALUES ('two');`),
		// The second statement fails, so the first must roll back too.
		"0003_bad.sql":   file(`CREATE TABLE bad (x INTEGER); INSERT INTO missing_table VALUES (1);`),
		"README.md":      file(`not a migration`),
		"sub/0000_x.sql": file(`not applied: in a subdirectory`),
	}
	err = s.migrate(ctx, first)
	if err == nil || !strings.Contains(err.Error(), "0003_bad.sql") {
		t.Fatalf("migrate() error = %v, want one naming 0003_bad.sql", err)
	}
	if got := queryStrings(t, db, "SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'bad'"); len(got) != 0 {
		t.Errorf("table bad exists after its migration failed")
	}
	if got, want := queryStrings(t, db, "SELECT filename FROM schema_migrations ORDER BY filename"),
		[]string{"0001_init.sql", "0002_insert.sql"}; !slices.Equal(got, want) {
		t.Errorf("schema_migrations after failure = %q, want %q", got, want)
	}

	fixed := fstest.MapFS{
		"0001_init.sql":   first["0001_init.sql"],
		"0002_insert.sql": first["0002_insert.sql"],
		"0003_bad.sql":    file(`CREATE TABLE bad (x INTEGER); INSERT INTO bad (x) VALUES (1);`),
	}
	if err := s.migrate(ctx, fixed); err != nil {
		t.Fatalf("migrate() with 0003 fixed error = %v", err)
	}
	if got, want := queryStrings(t, db, "SELECT v FROM t"), []string{"two"}; !slices.Equal(got, want) {
		t.Errorf("t rows = %q, want %q (0002 must not run twice)", got, want)
	}
	if got, want := queryStrings(t, db, "SELECT CAST(x AS TEXT) FROM bad"), []string{"1"}; !slices.Equal(got, want) {
		t.Errorf("bad rows = %q, want %q", got, want)
	}
	if got, want := queryStrings(t, db, "SELECT filename FROM schema_migrations ORDER BY filename"),
		[]string{"0001_init.sql", "0002_insert.sql", "0003_bad.sql"}; !slices.Equal(got, want) {
		t.Errorf("schema_migrations = %q, want %q", got, want)
	}
}

func TestBuildSearch(t *testing.T) {
	ptr := func(v float64) *float64 { return &v }
	beds := 3

	t.Run("all filters", func(t *testing.T) {
		f := domain.SearchFilter{
			MinPrice:      ptr(123456),
			MaxPrice:      ptr(654321),
			MinBedrooms:   &beds,
			City:          "  Zzcity' OR 1=1 --  ",
			Keyword:       "ZZTOKEN-ONE zztoken%two",
			KeywordTokens: []string{"ZZTOKEN-ONE", " ", "zztoken%two"},
			Statuses:      []domain.ListingStatus{domain.StatusActive, domain.StatusPending},
		}
		query, args := buildSearch(f)
		wantQuery := "SELECT id, source, address, city, state, zip, price, bedrooms, bathrooms, sqft, " +
			"latitude, longitude, listed_date, status, description FROM listings " +
			"WHERE price >= ? AND price <= ? AND bedrooms >= ? AND city = ? COLLATE NOCASE " +
			"AND instr(lower(description), ?) > 0 AND instr(lower(description), ?) > 0 " +
			"AND status IN (?, ?) ORDER BY source, id"
		if query != wantQuery {
			t.Errorf("query =\n%s\nwant\n%s", query, wantQuery)
		}
		wantArgs := []any{123456.0, 654321.0, 3, "Zzcity' OR 1=1 --", "zztoken-one", "zztoken%two", "active", "pending"}
		if !slices.Equal(args, wantArgs) {
			t.Errorf("args = %#v, want %#v", args, wantArgs)
		}
		for _, v := range []string{"123456", "654321", "Zzcity", "zztoken", "ZZTOKEN", "active", "pending", "'", "%"} {
			if strings.Contains(query, v) {
				t.Errorf("query text contains the filter value %q: %s", v, query)
			}
		}
	})

	t.Run("no filters", func(t *testing.T) {
		query, args := buildSearch(domain.SearchFilter{City: "   ", Keyword: "x", KeywordTokens: []string{"", " "}})
		want := "SELECT id, source, address, city, state, zip, price, bedrooms, bathrooms, sqft, " +
			"latitude, longitude, listed_date, status, description FROM listings ORDER BY source, id"
		if query != want {
			t.Errorf("query = %q, want %q", query, want)
		}
		if len(args) != 0 {
			t.Errorf("args = %#v, want none", args)
		}
	})

	t.Run("one status", func(t *testing.T) {
		query, args := buildSearch(domain.SearchFilter{Statuses: []domain.ListingStatus{domain.StatusSold}})
		if !strings.HasSuffix(query, " FROM listings WHERE status IN (?) ORDER BY source, id") {
			t.Errorf("query = %q, want a single-placeholder status clause", query)
		}
		if !slices.Equal(args, []any{"sold"}) {
			t.Errorf("args = %#v, want [sold]", args)
		}
	})
}
