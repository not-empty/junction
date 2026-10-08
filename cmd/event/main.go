package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/not-empty/junction/platform/bootstrap"
	"github.com/not-empty/junction/platform/event"
	"github.com/not-empty/junction/platform/event/kafka"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deps, closeDeps, err := bootstrap.MountDeps(ctx)
	if err != nil {
		return err
	}

	defer closeDeps()

	registry := bootstrap.MountEvents(deps, modules)

	subscriber := kafka.NewSubscriber(kafka.Config{
		Brokers:     cfg.KafkaBrokers,
		GroupID:     cfg.KafkaGroupID,
		MaxAttempts: defaultMaxAttempts,
		Backoff:     defaultBackoff,
		DLQSuffix:   defaultDLQSuffix,
	}, registry.Topics())

	defer subscriber.Close()

	slog.Info("event consumer started", "topics", registry.Topics(), "group", cfg.KafkaGroupID)

	return subscriber.Run(ctx, event.Dispatcher(registry, cfg.HandlerTimeout))
}
