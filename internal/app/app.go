package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/grigory/url-shortener/internal/adapters/http"
	"github.com/grigory/url-shortener/internal/adapters/memory"
	"github.com/grigory/url-shortener/internal/adapters/postgres"
	"github.com/grigory/url-shortener/internal/config"
	"github.com/grigory/url-shortener/internal/ports"
	"github.com/grigory/url-shortener/internal/service"
)

func Run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	repo, err := newRepository(cfg, logger)
	if err != nil {
		return err
	}

	svc := service.New(repo, logger)
	handler := httpadapter.NewHandler(svc, logger)
	router := httpadapter.NewRouter(handler)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", slog.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("server stopped")
	return nil
}

func newRepository(cfg *config.Config, logger *slog.Logger) (ports.Repository, error) {
	switch cfg.Storage {
	case "memory":
		logger.Info("using in-memory storage")
		return memory.New(), nil

	case "postgres":
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		repo, err := postgres.NewRepository(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("postgres storage: %w", err)
		}
		logger.Info("using postgres storage", slog.String("dsn", maskDSN(cfg.DatabaseURL)))
		return repo, nil

	default:
		return nil, fmt.Errorf("unknown storage type %q: choose memory or postgres", cfg.Storage)
	}
}

func maskDSN(dsn string) string {
	for i, c := range dsn {
		if c == '@' {
			return "***@" + dsn[i+1:]
		}
	}
	return "***"
}
