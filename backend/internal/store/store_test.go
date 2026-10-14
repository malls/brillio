package store_test

import (
	"path/filepath"
	"strings"
	"testing"

	"brillio/backend/internal/store"
)

func TestPing(t *testing.T) {
	s, err := store.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := s.Ping(t.Context()); err != nil {
		t.Errorf("Ping() on an open store = %v, want nil", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := s.Ping(t.Context()); err == nil {
		t.Error("Ping() after Close = nil, want an error")
	}
}

func TestOpenErrors(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing", "listings.db")
	tests := []struct {
		name string
		path string
		want string // substring of the error
	}{
		{name: "parent directory missing", path: missingDir, want: missingDir},
		{name: "query string", path: "x.db?mode=ro", want: "DB_PATH must be a file path or :memory:"},
		{name: "file URI", path: "file:x.db", want: "DB_PATH must be a file path or :memory:"},
		{name: "empty", path: "", want: "DB_PATH must be a file path or :memory:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := store.Open(t.Context(), tt.path)
			if err == nil {
				_ = s.Close()
				t.Fatalf("Open(%q) error = nil, want one containing %q", tt.path, tt.want)
			}
			t.Logf("Open(%q) error: %v", tt.path, err)
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Open(%q) error = %q, want it to contain %q", tt.path, err, tt.want)
			}
		})
	}
}
