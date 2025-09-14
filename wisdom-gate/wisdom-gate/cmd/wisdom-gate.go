package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wisdom-gate/internal/adapters/postgres"
	"wisdom-gate/internal/config"
	"wisdom-gate/internal/delivery/tcp"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	repo, err := postgres.NewPostgresDBPool(ctx, cfg.Repo.ConnectionString)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	logger.Info("Database pool has been initialized")

	// Выполнение миграций
	connForMigrations := stdlib.OpenDBFromPool(repo)
	if err = goose.Up(connForMigrations, cfg.Repo.MigrationPath); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Создание сервера
	server, err := tcp.NewServer(cfg, logger, repo)
	if err != nil {
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("Starting wisdom-gate server...")
		serverErr <- server.Start(ctx)
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received, starting graceful shutdown...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Graceful shutdown failed", "error", err)
		}
		logger.Info("Graceful shutdown completed")

	case err := <-serverErr:
		if err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
