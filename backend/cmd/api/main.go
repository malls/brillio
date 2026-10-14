// Command api serves the listing search HTTP API. It only wires things
// together: configuration, logging, the store (opened, migrated and seeded
// from LISTINGS_FILE at startup), the router, the HTTP server and graceful
// shutdown on SIGINT or SIGTERM. Startup failures, including an invalid seed
// record, exit with status 1.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brillio/backend/internal/config"
	"brillio/backend/internal/httpapi"
	"brillio/backend/internal/store"
	"brillio/backend/internal/store/seed"
)

// HTTP server timeouts. Fixed protections, not deployment settings.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.New(slog.NewJSONHandler(os.Stdout, nil)).Error("fatal", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

// run starts the server and blocks until it stops. It returns nil after a
// clean shutdown triggered by SIGINT or SIGTERM.
func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Validate the seed file before opening the database, so a bad file
	// never creates a database file.
	listings, err := seed.Load(cfg.ListingsFile)
	if err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	st, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			logger.Error("store close failed", slog.String("err", err.Error()))
		}
	}()
	n, err := st.Upsert(ctx, listings)
	if err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	logger.Info(fmt.Sprintf("seeded %d listings", n),
		slog.Int("count", n),
		slog.String("listings_file", cfg.ListingsFile),
		slog.String("db_path", cfg.DBPath),
	)

	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	srv := &http.Server{
		Handler: httpapi.NewRouter(httpapi.Deps{
			Logger:      logger,
			CORSOrigins: cfg.CORSOrigins,
			Pinger:      st,
		}),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}

	logger.Info("listening",
		slog.String("addr", ln.Addr().String()),
		slog.String("log_level", cfg.LogLevel.String()),
		slog.Any("cors_origins", cfg.CORSOrigins),
		slog.String("shutdown_timeout", cfg.ShutdownTimeout.String()),
	)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	// Restore default signal handling so a second signal kills the process
	// instead of waiting out the drain.
	stop()

	logger.Info("shutting down", slog.String("timeout", cfg.ShutdownTimeout.String()))
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return errors.Join(fmt.Errorf("shutdown: %w", err), srv.Close())
	}
	logger.Info("server stopped")
	return nil
}
