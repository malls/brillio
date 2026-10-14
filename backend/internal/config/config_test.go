package config_test

import (
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"brillio/backend/internal/config"
)

// lookup returns a getenv that reads only from m, so tests never touch the
// real process environment.
func lookup(m map[string]string) func(string) string {
	return func(name string) string { return m[name] }
}

func TestLoad(t *testing.T) {
	defaults := config.Config{
		Port:            8080,
		DBPath:          "./data/listings.db",
		ListingsFile:    "./data/sample_listings.json",
		LogLevel:        slog.LevelInfo,
		CORSOrigins:     []string{"http://localhost:5173"},
		ShutdownTimeout: 10 * time.Second,
	}
	with := func(change func(*config.Config)) config.Config {
		c := defaults
		c.CORSOrigins = slices.Clone(defaults.CORSOrigins)
		change(&c)
		return c
	}

	tests := []struct {
		name string
		env  map[string]string
		want config.Config
		// wantErr lists substrings the error must contain; empty means success.
		wantErr []string
	}{
		{name: "all defaults", env: nil, want: defaults},
		{
			name: "whitespace-only values mean default",
			env: map[string]string{
				"PORT": " ", "DB_PATH": "\t", "LISTINGS_FILE": " ", "LOG_LEVEL": " ",
				"CORS_ORIGINS": "  ", "SHUTDOWN_TIMEOUT": " ",
			},
			want: defaults,
		},
		{
			name: "full valid override",
			env: map[string]string{
				"PORT":             " 18002 ",
				"DB_PATH":          ":memory:",
				"LISTINGS_FILE":    "/srv/listings.json",
				"LOG_LEVEL":        "debug",
				"CORS_ORIGINS":     "https://app.test",
				"SHUTDOWN_TIMEOUT": "30s",
			},
			want: config.Config{
				Port:            18002,
				DBPath:          ":memory:",
				ListingsFile:    "/srv/listings.json",
				LogLevel:        slog.LevelDebug,
				CORSOrigins:     []string{"https://app.test"},
				ShutdownTimeout: 30 * time.Second,
			},
		},

		{name: "log level debug", env: map[string]string{"LOG_LEVEL": "debug"},
			want: with(func(c *config.Config) { c.LogLevel = slog.LevelDebug })},
		{name: "log level info", env: map[string]string{"LOG_LEVEL": "info"}, want: defaults},
		{name: "log level warn", env: map[string]string{"LOG_LEVEL": "warn"},
			want: with(func(c *config.Config) { c.LogLevel = slog.LevelWarn })},
		{name: "log level error", env: map[string]string{"LOG_LEVEL": "error"},
			want: with(func(c *config.Config) { c.LogLevel = slog.LevelError })},
		{name: "log level is case-insensitive", env: map[string]string{"LOG_LEVEL": "WARN"},
			want: with(func(c *config.Config) { c.LogLevel = slog.LevelWarn })},
		{name: "log level verbose", env: map[string]string{"LOG_LEVEL": "verbose"},
			wantErr: []string{`LOG_LEVEL: "verbose" is not one of debug|info|warn|error`}},
		{name: "log level warning", env: map[string]string{"LOG_LEVEL": "warning"},
			wantErr: []string{`LOG_LEVEL: "warning"`}},
		{name: "log level with offset", env: map[string]string{"LOG_LEVEL": "INFO+2"},
			wantErr: []string{`LOG_LEVEL: "INFO+2"`}},

		{name: "shutdown timeout 500ms", env: map[string]string{"SHUTDOWN_TIMEOUT": "500ms"},
			want: with(func(c *config.Config) { c.ShutdownTimeout = 500 * time.Millisecond })},
		{name: "shutdown timeout not a duration", env: map[string]string{"SHUTDOWN_TIMEOUT": "abc"},
			wantErr: []string{`SHUTDOWN_TIMEOUT: "abc" is not a Go duration`}},
		{name: "shutdown timeout zero", env: map[string]string{"SHUTDOWN_TIMEOUT": "0s"},
			wantErr: []string{`SHUTDOWN_TIMEOUT: "0s" must be greater than zero`}},
		{name: "shutdown timeout negative", env: map[string]string{"SHUTDOWN_TIMEOUT": "-1s"},
			wantErr: []string{`SHUTDOWN_TIMEOUT: "-1s" must be greater than zero`}},
		{name: "shutdown timeout missing unit", env: map[string]string{"SHUTDOWN_TIMEOUT": "10"},
			wantErr: []string{`SHUTDOWN_TIMEOUT: "10" is not a Go duration`}},

		{name: "port 18002", env: map[string]string{"PORT": "18002"},
			want: with(func(c *config.Config) { c.Port = 18002 })},
		{name: "port 65535", env: map[string]string{"PORT": "65535"},
			want: with(func(c *config.Config) { c.Port = 65535 })},
		{name: "port not a number", env: map[string]string{"PORT": "abc"},
			wantErr: []string{`PORT: "abc" is not an integer from 1 to 65535`}},
		{name: "port zero", env: map[string]string{"PORT": "0"},
			wantErr: []string{`PORT: "0"`}},
		{name: "port too large", env: map[string]string{"PORT": "65536"},
			wantErr: []string{`PORT: "65536"`}},

		{
			name: "cors origins trimmed and empties dropped",
			env:  map[string]string{"CORS_ORIGINS": " http://a.test , http://b.test:8080 ,"},
			want: with(func(c *config.Config) { c.CORSOrigins = []string{"http://a.test", "http://b.test:8080"} }),
		},
		{name: "cors origin with trailing slash", env: map[string]string{"CORS_ORIGINS": "http://a.test/"},
			wantErr: []string{`CORS_ORIGINS: "http://a.test/" is not a bare http(s)://host[:port] origin`}},
		{name: "cors origin with path", env: map[string]string{"CORS_ORIGINS": "http://a.test/app"},
			wantErr: []string{`CORS_ORIGINS: "http://a.test/app"`}},
		{name: "cors origin with query", env: map[string]string{"CORS_ORIGINS": "http://a.test?x=1"},
			wantErr: []string{`CORS_ORIGINS: "http://a.test?x=1"`}},
		{name: "cors origin with userinfo", env: map[string]string{"CORS_ORIGINS": "http://u@a.test"},
			wantErr: []string{`CORS_ORIGINS: "http://u@a.test"`}},
		{name: "cors origin wildcard", env: map[string]string{"CORS_ORIGINS": "*"},
			wantErr: []string{`CORS_ORIGINS: "*" is not allowed`}},
		{name: "cors origin without scheme", env: map[string]string{"CORS_ORIGINS": "localhost:5173"},
			wantErr: []string{`CORS_ORIGINS: "localhost:5173"`}},
		{name: "cors origin with other scheme", env: map[string]string{"CORS_ORIGINS": "ftp://a.test"},
			wantErr: []string{`CORS_ORIGINS: "ftp://a.test"`}},
		{name: "cors origins all empty", env: map[string]string{"CORS_ORIGINS": " , "},
			wantErr: []string{"CORS_ORIGINS: at least one origin is required"}},

		{
			name:    "every bad variable is reported",
			env:     map[string]string{"PORT": "abc", "LOG_LEVEL": "verbose", "CORS_ORIGINS": ","},
			wantErr: []string{`PORT: "abc"`, `LOG_LEVEL: "verbose"`, "CORS_ORIGINS: at least one origin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.Load(lookup(tt.env))
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("Load() error = nil, want one containing %q (config %+v)", tt.wantErr, got)
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("Load() error = %q, want it to contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() =\n  %+v\nwant\n  %+v", got, tt.want)
			}
		})
	}
}

func TestConfigAddr(t *testing.T) {
	if got := (config.Config{Port: 18002}).Addr(); got != ":18002" {
		t.Errorf("Addr() = %q, want %q", got, ":18002")
	}
}
