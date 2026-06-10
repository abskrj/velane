package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/api"
	"github.com/abskrj/velane/services/control-plane/internal/api/handlers"
	"github.com/abskrj/velane/services/control-plane/internal/auth"
	"github.com/abskrj/velane/services/control-plane/internal/config"
	"github.com/abskrj/velane/services/control-plane/internal/executor"
	"github.com/abskrj/velane/services/control-plane/internal/executor/firecracker"
	"github.com/abskrj/velane/services/control-plane/internal/executor/remote"
	"github.com/abskrj/velane/services/control-plane/internal/nango"
	"github.com/abskrj/velane/services/control-plane/internal/observability"
	"github.com/abskrj/velane/services/control-plane/internal/platformlibs"
	"github.com/abskrj/velane/services/control-plane/internal/scheduler"
	"github.com/abskrj/velane/services/control-plane/internal/store/postgres"
	redisstore "github.com/abskrj/velane/services/control-plane/internal/store/redis"
	"github.com/abskrj/velane/services/control-plane/internal/worker"
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
	log.Info("starting velane control-plane",
		zap.String("port", cfg.Port),
		zap.String("bun_executor", cfg.BunExecutorURL),
		zap.String("python_executor", cfg.PythonExecutorURL),
		zap.String("redis_url", cfg.RedisURL),
		zap.Int("worker_count", cfg.WorkerCount),
	)

	// --- Encryption key ---
	encKey := cfg.EncryptionKeyBytes(log)

	// --- Context (cancelled on shutdown signal) ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Postgres (with retry for docker-compose startup ordering) ---
	store, err := connectWithRetry(ctx, cfg.DatabaseURL, log, 10, 2*time.Second)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer store.Close()
	log.Info("postgres connected and migrations applied")

	// --- Platform libraries (embedded in binary) ---
	platLibs, err := platformlibs.Load()
	if err != nil {
		log.Fatal("failed to load platform libraries", zap.Error(err))
	}
	log.Info("platform libraries loaded", zap.Int("count", len(platLibs)))

	// --- Auth provider (needed early for bootstrap) ---
	privKey, pubKey := cfg.JWTKeyPair(log)
	authProvider := auth.NewJWTProvider(store, privKey, "https://api.velane.io")

	// --- Bootstrap first admin (no-op if users already exist or env vars not set) ---
	bootstrapAdminIfNeeded(ctx, store, authProvider, cfg, log)

	// --- Nango client ---
	nangoClient := nango.New(cfg.NangoInternalURL, cfg.NangoSecretKey)
	if cfg.NangoSecretKey == "" {
		log.Warn("NANGO_SECRET_KEY not set — integration OAuth connections will not function")
	}
	log.Info("nango client configured", zap.String("url", cfg.NangoInternalURL))

	// --- Executor ---
	var exec executor.Executor
	switch cfg.ExecutorType {
	case "firecracker":
		fcExec, err := firecracker.New(firecracker.Config{
			FirecrackerBin: cfg.FirecrackerBinary,
			JailerBin:      cfg.FirecrackerJailerBinary,
			BunRootfs:      cfg.FirecrackerBunRootfs,
			PythonRootfs:   cfg.FirecrackerPythonRootfs,
			KernelImage:    cfg.FirecrackerKernelImage,
		}, log)
		if err != nil {
			log.Fatal("firecracker executor init failed", zap.Error(err))
		}
		exec = fcExec
	default:
		exec = remote.New(cfg.BunExecutorURL, cfg.PythonExecutorURL)
	}

	// --- Redis (optional — if unreachable we fall back to sync-only mode) ---
	var redisClient *redisstore.Client
	var sched *scheduler.Scheduler
	observer := observability.NewPipelineObserver(nil, nil, nil)

	rc, err := redisstore.New(cfg.RedisURL)
	if err != nil {
		log.Warn("redis not available, running in sync-only mode (no async/queue)",
			zap.String("redis_url", cfg.RedisURL),
			zap.Error(err),
		)
		sched = scheduler.New(store, exec, encKey, platLibs)
	} else {
		redisClient = rc
		log.Info("redis connected", zap.String("redis_url", cfg.RedisURL))
		sched = scheduler.NewWithQueue(store, exec, redisClient, encKey, platLibs)
		sched.SetEventStream(redisClient)
	}
	sched.WithInternalProxyURL(cfg.InternalProxyURL)
	sched.SetObserver(observer)

	// --- Background worker (only if Redis is available) ---
	if redisClient != nil {
		w := worker.New(redisClient, store, exec, log, cfg.WorkerCount)
		w.SetObserver(observer)
		w.SetEventPublisher(redisClient)
		go w.Run(ctx)
		log.Info("background worker started", zap.Int("workers", cfg.WorkerCount))

		// Reaper: mark invocations stuck in pending/running (worker never
		// picked them up or crashed mid-run) as timed out.
		go runStaleInvocationReaper(ctx, store, log)
	}

	// --- Router ---
	oauthCfg := handlers.OAuthConfig{
		PublicBaseURL:           cfg.PublicBaseURL,
		GoogleOAuthClientID:     cfg.GoogleOAuthClientID,
		GoogleOAuthClientSecret: cfg.GoogleOAuthClientSecret,
		GitHubOAuthClientID:     cfg.GitHubOAuthClientID,
		GitHubOAuthClientSecret: cfg.GitHubOAuthClientSecret,
	}
	router := api.NewRouterWithJWT(store, sched, log, encKey, authProvider, pubKey, nangoClient, cfg.NangoInternalURL, cfg.NangoConnectURL, cfg.NangoApiURL, cfg.NangoWebhookSecret, cfg.NangoSecretKey, cfg.MCPPublicURL, platLibs, oauthCfg)

	// --- HTTP server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute, // allow long invocations and streams
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

	// Cancel ctx first so background workers stop pulling new jobs.
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Warn("redis close", zap.Error(err))
		}
	}

	log.Info("server stopped")
}

// runStaleInvocationReaper periodically fails invocations stuck in
// pending/running longer than the staleness threshold. The threshold is set
// well above the maximum streaming wait so it never races a live invocation.
func runStaleInvocationReaper(ctx context.Context, store *postgres.Store, log *zap.Logger) {
	const (
		interval   = 1 * time.Minute
		staleAfter = 10 * time.Minute
	)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := store.FailStaleInvocations(ctx, staleAfter)
			if err != nil {
				log.Warn("reaper: fail stale invocations", zap.Error(err))
				continue
			}
			if n > 0 {
				log.Info("reaper: marked stale invocations as timed out", zap.Int64("count", n))
			}
		}
	}
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
