package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/not-empty/bridge/platform/bootstrap"
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

	deps, closeDeps, err := bootstrap.MountDeps(context.Background())
	if err != nil {
		return err
	}

	defer closeDeps()

	handler, ok := bootstrap.MountQueues(deps, modules)[cfg.Queue]
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
