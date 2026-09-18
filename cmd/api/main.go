package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/not-empty/bridge/platform/database"
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

	db, err := database.Open(ctx, database.Config{
		DSN: cfg.DatabaseDSN,
	})

	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	defer db.Close()

	router := newRouter(db)

	app := httpserver.NewAPIServer(
		":"+cfg.Port,
		router,
	)

	app.UseMiddleware(httpserver.RecoverMiddleware)
	app.UseMiddleware(httpserver.NewCorsMiddleware(cfg.CORSAllowedOrigins))
	app.UseMiddleware(httpserver.NewBodyLimitMiddleware(cfg.MaxBodyBytes))

	return app.Run(ctx)
}
