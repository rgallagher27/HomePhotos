package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rgallagher/homephotos/config"
	"github.com/rgallagher/homephotos/ports/rest"
	"github.com/rgallagher/homephotos/services/cache"
	"github.com/rgallagher/homephotos/services/scanner"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var cfg config.Config
	if err := envconfig.Process("homephotos", &cfg); err != nil {
		slog.ErrorContext(ctx, "load config", "error", err)
		os.Exit(1)
	}

	if err := run(ctx, cfg); err != nil {
		slog.ErrorContext(ctx, "server exited with error", "error", err)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	httpServer, scannerSvc, cacheSvc, err := rest.NewRestServer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("setup server: %w", err)
	}

	// Start scanner scheduler in background
	sched := scanner.NewScheduler(scannerSvc, cfg.ScanInterval, cfg.ScanOnStartup)
	go sched.Start(ctx)

	// Start cache worker pool in background
	wp := cache.NewWorkerPool(cacheSvc, cfg.CacheWorkers)
	go wp.Start(ctx)

	serverErr := make(chan error, 1)

	go func() {
		slog.InfoContext(ctx, "starting server", "addr", httpServer.Addr)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		slog.Info("server shutdown gracefully")
		return nil
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}
