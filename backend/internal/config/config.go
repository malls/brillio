// Package config loads the API's runtime configuration from environment
// variables (docs/SPEC.md section 11), applying defaults and validating every
// value once at startup.
//
// An unset, empty or whitespace-only variable takes its default, and Load
// rejects anything else that breaks a rule: LOG_LEVEL must be one of debug,
// info, warn or error (any case); SHUTDOWN_TIMEOUT must be a positive Go
// duration such as 10s; PORT must be an integer from 1 to 65535; and
// CORS_ORIGINS must hold at least one comma-separated entry, each a bare
// http(s)://host[:port] origin with no path, trailing slash, query, fragment,
// userinfo, "*" wildcard or bare trailing colon (a colon with no port), and
// each is stored with its scheme and host lowercased, as a browser sends them.
// DB_PATH and LISTINGS_FILE pass through unchecked.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Defaults from docs/SPEC.md section 11.
const (
	defaultPort            = 8080
	defaultDBPath          = "./data/listings.db"
	defaultListingsFile    = "./data/sample_listings.json"
	defaultLogLevel        = slog.LevelInfo
	defaultCORSOrigins     = "http://localhost:5173"
	defaultShutdownTimeout = 10 * time.Second
)

// logLevels maps every accepted LOG_LEVEL value. It is explicit rather than
// slog.Level.UnmarshalText, which also accepts offsets such as "INFO+2".
var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// Config is the API's validated runtime configuration.
type Config struct {
	// Port is the TCP port the server listens on (PORT).
	Port int
	// DBPath is the SQLite database file, or ":memory:" (DB_PATH).
	DBPath string
	// ListingsFile is the JSON file the database is seeded from (LISTINGS_FILE).
	ListingsFile string
	// LogLevel is the minimum level logged (LOG_LEVEL).
	LogLevel slog.Level
	// CORSOrigins are the exact origins allowed cross-origin access, with
	// scheme and host lowercased (CORS_ORIGINS).
	CORSOrigins []string
	// ShutdownTimeout bounds the graceful drain on SIGINT or SIGTERM (SHUTDOWN_TIMEOUT).
	ShutdownTimeout time.Duration
}

// Addr returns the listen address for Port on all interfaces, such as ":8080".
func (c Config) Addr() string {
	return ":" + strconv.Itoa(c.Port)
}

// Load reads the configuration through getenv (os.Getenv in production, a map
// lookup in tests). It reports every invalid variable at once, joined, each
// message prefixed with the variable's name.
func Load(getenv func(string) string) (Config, error) {
	get := func(name string) string { return strings.TrimSpace(getenv(name)) }
	cfg := Config{
		Port:            defaultPort,
		DBPath:          defaultDBPath,
		ListingsFile:    defaultListingsFile,
		LogLevel:        defaultLogLevel,
		ShutdownTimeout: defaultShutdownTimeout,
	}
	var errs []error

	if v := get("PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil || port < 1 || port > 65535 {
			errs = append(errs, fmt.Errorf("PORT: %q is not an integer from 1 to 65535", v))
		} else {
			cfg.Port = port
		}
	}

	if v := get("DB_PATH"); v != "" {
		cfg.DBPath = v
	}

	if v := get("LISTINGS_FILE"); v != "" {
		cfg.ListingsFile = v
	}

	if v := get("LOG_LEVEL"); v != "" {
		level, ok := logLevels[strings.ToLower(v)]
		if !ok {
			errs = append(errs, fmt.Errorf("LOG_LEVEL: %q is not one of debug|info|warn|error", v))
		} else {
			cfg.LogLevel = level
		}
	}

	origins := get("CORS_ORIGINS")
	if origins == "" {
		origins = defaultCORSOrigins
	}
	corsOrigins, originErrs := parseOrigins(origins)
	cfg.CORSOrigins = corsOrigins
	errs = append(errs, originErrs...)

	if v := get("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT: %q is not a Go duration such as 10s", v))
		case d <= 0:
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT: %q must be greater than zero", v))
		default:
			cfg.ShutdownTimeout = d
		}
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// parseOrigins splits a comma-separated origin list, trims each entry and drops
// empty ones. It returns a CORS_ORIGINS error for every entry that is not a
// bare origin: a browser's Origin header is exactly scheme://host[:port], so
// anything more (even a trailing slash or a colon with no port) could never
// match. Scheme and host are compared case-insensitively and stored lowercased,
// because a browser always sends them lowercase.
func parseOrigins(raw string) ([]string, []error) {
	var (
		origins []string
		errs    []error
	)
	for entry := range strings.SplitSeq(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == "*" {
			errs = append(errs, errors.New(`CORS_ORIGINS: "*" is not allowed; list each origin explicitly`))
			continue
		}
		u, err := url.Parse(entry)
		// url.Parse lowercases the scheme but keeps the host as written, so the
		// rebuilt origin matches the entry only up to case.
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" ||
			!strings.EqualFold(u.Scheme+"://"+u.Host, entry) {
			errs = append(errs, fmt.Errorf(
				"CORS_ORIGINS: %q is not a bare http(s)://host[:port] origin "+
					"(no path, trailing slash, query, fragment or userinfo)", entry))
			continue
		}
		if strings.HasSuffix(u.Host, ":") {
			errs = append(errs, fmt.Errorf(
				"CORS_ORIGINS: %q ends in a colon with no port; remove the colon or add a port", entry))
			continue
		}
		origins = append(origins, strings.ToLower(u.Scheme)+"://"+strings.ToLower(u.Host))
	}
	if len(origins) == 0 && len(errs) == 0 {
		errs = append(errs, errors.New("CORS_ORIGINS: at least one origin is required"))
	}
	return origins, errs
}
