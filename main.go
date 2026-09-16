package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/not-empty/bridge/config"
	"github.com/not-empty/bridge/database"
	"github.com/not-empty/bridge/middleware"
	"github.com/not-empty/bridge/routes"
	"github.com/not-empty/bridge/server"
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
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
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

	router := routes.SetRoutes(db)

	app := server.NewAPIServer(
		":"+cfg.Port,
		router,
	)

	app.UseMiddleware(middleware.RecoverMiddleware)
	app.UseMiddleware(middleware.NewCorsMiddleware(cfg.CORSAllowedOrigins))
	app.UseMiddleware(middleware.NewBodyLimitMiddleware(cfg.MaxBodyBytes))

	return app.Run(ctx)
}
