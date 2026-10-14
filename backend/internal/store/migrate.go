package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"time"
)

// migrationFiles holds the schema migrations, applied in filename order. Name
// new files NNNN_name.sql; a migration must not contain BEGIN or COMMIT,
// because Migrate wraps each file in its own transaction.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

const (
	// countSchemaMigrations tells whether the schema_migrations table exists
	// yet; 0001_listings.sql creates it.
	countSchemaMigrations = "SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'"
	selectAppliedFiles    = "SELECT filename FROM schema_migrations"
	insertSchemaMigration = "INSERT INTO schema_migrations (filename, applied_at) VALUES (?, ?)"
)

// Migrate applies every embedded migration not yet recorded in
// schema_migrations, in filename order, each in its own transaction. Running
// it again is a no-op. Open already calls it.
func (s *Store) Migrate(ctx context.Context) error {
	sub, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return s.migrate(ctx, sub)
}

// migrate applies the pending *.sql files at the root of fsys in filename
// order. A failing file is rolled back and named in the error; files applied
// before it stay applied.
func (s *Store) migrate(ctx context.Context, fsys fs.FS) error {
	entries, err := fs.ReadDir(fsys, ".") // sorted by filename
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	applied, err := s.appliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || path.Ext(name) != ".sql" || applied[name] {
			continue
		}
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("migrate %s: %w", name, err)
		}
		if err := s.applyMigration(ctx, name, string(body)); err != nil {
			return fmt.Errorf("migrate %s: %w", name, err)
		}
	}
	return nil
}

// appliedMigrations returns the set of migration filenames already applied.
// Its rows are closed before it returns, so no query is left open when a
// migration transaction begins (a ":memory:" store has a single connection).
func (s *Store) appliedMigrations(ctx context.Context) (map[string]bool, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, countSchemaMigrations).Scan(&n); err != nil {
		return nil, err
	}
	applied := map[string]bool{}
	if n == 0 {
		return applied, nil
	}
	rows, err := s.db.QueryContext(ctx, selectAppliedFiles)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return applied, nil
}

// applyMigration runs one migration file and records it in
// schema_migrations, both in a single transaction.
func (s *Store) applyMigration(ctx context.Context, name, body string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // ErrTxDone after Commit
	if _, err := tx.ExecContext(ctx, body); err != nil {
		return err
	}
	appliedAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, insertSchemaMigration, name, appliedAt); err != nil {
		return err
	}
	return tx.Commit()
}
