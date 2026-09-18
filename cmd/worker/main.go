package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/not-empty/bridge/platform/database"
	"github.com/not-empty/bridge/platform/queue"
	"github.com/not-empty/omniq-go/src/omniq"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	err := run()
	if err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	db, err := database.Open(context.Background(), database.Config{
		DSN: cfg.DatabaseDSN,
	})

	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	defer db.Close()

	handler, ok := newRegistry(db)[cfg.Queue]
	if !ok {
		return fmt.Errorf("queue %q is not registered", cfg.Queue)
	}

	client, err := omniq.NewClient(omniq.ClientOpts{
		Host: cfg.RedisHost,
		Port: cfg.RedisPort,
	})

	if err != nil {
		return fmt.Errorf("creating omniq client: %w", err)
	}

	slog.Info("worker started", "queue", cfg.Queue)

	return client.Consume(omniq.ConsumeOpts{
		Queue:            cfg.Queue,
		Handler:          queue.OmniqHandler(handler, cfg.JobTimeout),
		PollIntervalS:    cfg.PollIntervalSeconds,
		PromoteIntervalS: cfg.PromoteIntervalSeconds,
		ReapIntervalS:    cfg.ReapIntervalSeconds,
		Drain:            true,
	})
}
