package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runeforge/control-plane/internal/api"
	"github.com/runeforge/control-plane/internal/config"
	"github.com/runeforge/control-plane/internal/executor/remote"
	"github.com/runeforge/control-plane/internal/scheduler"
	"github.com/runeforge/control-plane/internal/store/postgres"
	"go.uber.org/zap"
)

func main() {
	// --- Logger ---
	log, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	// --- Config ---
	cfg := config.Load()
	log.Info("starting runeforge control-plane",
		zap.String("port", cfg.Port),
		zap.String("bun_executor", cfg.BunExecutorURL),
		zap.String("python_executor", cfg.PythonExecutorURL),
	)

	// --- Postgres (with retry for docker-compose startup ordering) ---
	ctx := context.Background()
	store, err := connectWithRetry(ctx, cfg.DatabaseURL, log, 10, 2*time.Second)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer store.Close()
	log.Info("postgres connected and migrations applied")

	// --- Executor ---
	exec := remote.New(cfg.BunExecutorURL, cfg.PythonExecutorURL)

	// --- Scheduler ---
	sched := scheduler.New(store, exec)

	// --- Router ---
	router := api.NewRouter(store, sched, log)

	// --- HTTP server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute, // allow long invocations
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
	log.Info("server stopped")
}

// connectWithRetry attempts to open a Postgres connection up to maxAttempts
// times, sleeping between each attempt. This handles the race where the control
// plane container starts before Postgres is fully ready.
func connectWithRetry(ctx context.Context, dsn string, log *zap.Logger, maxAttempts int, delay time.Duration) (*postgres.Store, error) {
	var lastErr error
	for i := 1; i <= maxAttempts; i++ {
		store, err := postgres.New(ctx, dsn)
		if err == nil {
			return store, nil
		}
		lastErr = err
		log.Warn("postgres not ready, retrying",
			zap.Int("attempt", i),
			zap.Int("max", maxAttempts),
			zap.Error(err),
		)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("gave up after %d attempts: %w", maxAttempts, lastErr)
}
