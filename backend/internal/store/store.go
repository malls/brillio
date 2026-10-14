// Package store is the service's SQLite persistence (docs/SPEC.md sections
// 4.1, 4.2, 5.2 and 6.1), built on the pure-Go modernc.org/sqlite driver.
//
// Open connects and applies the embedded schema migrations, so a *Store is
// never handed out unmigrated. Upsert writes seeded listings, Search runs the
// filtered listing query, Cities lists the distinct city names and Ping backs
// GET /healthz. Every query is parameterized: SQL text is assembled only from
// constant fragments and "?" placeholders, never from caller values. The
// subpackage seed decodes and validates the listings file.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// driverName is the database/sql driver modernc.org/sqlite registers.
const driverName = "sqlite"

// memoryPath is the DB_PATH value for a private in-memory database.
const memoryPath = ":memory:"

// DSN query parameters. modernc.org/sqlite runs every _pragma on each new
// connection, so the settings hold across the whole pool; _txlock=immediate
// makes write transactions take the write lock at BEGIN, which avoids the WAL
// lock-upgrade SQLITE_BUSY that busy_timeout cannot retry.
const (
	dsnMemoryParams = "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_txlock=immediate"
	dsnFileParams   = "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_txlock=immediate"
)

// Store is a migrated SQLite listings database. It is safe for concurrent
// use.
type Store struct {
	db *sql.DB
}

// Open opens the SQLite database at path (DB_PATH), which is a file path or
// ":memory:", checks the connection and applies any pending migrations. The
// file is created if missing, but its parent directory is not. A path that
// is empty, contains "?" or starts with "file:" is rejected, since the
// driver would read those as a URI or DSN parameters rather than a file name.
//
// A ":memory:" database lives on a single connection: each connection to
// ":memory:" is a separate, empty database.
func Open(ctx context.Context, path string) (*Store, error) {
	if path == "" || strings.Contains(path, "?") || strings.HasPrefix(path, "file:") {
		return nil, fmt.Errorf("DB_PATH must be a file path or :memory:, got %q", path)
	}
	db, err := sql.Open(driverName, dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if path == memoryPath {
		db.SetMaxOpenConns(1)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	s := &Store{db: db}
	if err := s.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// dsn returns the driver data source name for path: the path followed by the
// pragma parameters. path is operator configuration, never request input.
func dsn(path string) string {
	if path == memoryPath {
		return path + dsnMemoryParams
	}
	return path + dsnFileParams
}

// Close closes the database. For ":memory:" this discards all data.
func (s *Store) Close() error {
	return s.db.Close()
}

// Ping reports whether the database is reachable. It satisfies
// httpapi.Pinger for GET /healthz.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
