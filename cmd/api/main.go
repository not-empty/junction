package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/not-empty/bridge/platform/bootstrap"
	"github.com/not-empty/bridge/platform/httpserver"
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
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deps, closeDeps, err := bootstrap.MountDeps(ctx)
	if err != nil {
		return err
	}

	defer closeDeps()

	app := httpserver.NewAPIServer(
		":"+cfg.Port,
		bootstrap.MountHTTP(deps, modules),
	)

	app.UseMiddleware(httpserver.RecoverMiddleware)
	app.UseMiddleware(httpserver.NewCorsMiddleware(cfg.CORSAllowedOrigins))
	app.UseMiddleware(httpserver.NewBodyLimitMiddleware(cfg.MaxBodyBytes))

	return app.Run(ctx)
}
