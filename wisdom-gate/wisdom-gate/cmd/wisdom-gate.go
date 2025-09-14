package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	logger.Info("repo Pool has been initialized")

	connForMigrations := stdlib.OpenDBFromPool(repo)
	if err = goose.Up(connForMigrations, cfg.Repo.MigrationPath); err != nil {
		log.Fatalf("failed to run migrations: %s", err)
	}

	server, err := tcp.NewServer(cfg, logger, repo)
	if err != nil {
		logger.Error("Failed to create wisdom-gate", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting wisdom-gate...")
	if err := server.Start(ctx); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
