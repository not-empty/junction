package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/not-empty/bridge/platform/config"
	"github.com/not-empty/bridge/platform/database"
)

type Deps struct {
	DB *database.DB
}

// A nil closer means the dependency is not configured: its field stays zero.
type provider struct {
	name string
	open func(ctx context.Context, cfg config.Config, deps *Deps) (func() error, error)
}

// Declare a new dependency here and on Deps. No cmd/ changes.
var providers = []provider{
	{name: "database", open: openDatabase},
}

func MountDeps(ctx context.Context) (Deps, func() error, error) {
	cfg, err := config.Load()
	if err != nil {
		return Deps{}, nil, err
	}

	var (
		deps    Deps
		closers []func() error
		mounted []string
	)

	closeAll := func() error {
		var closeErr error

		for i := len(closers) - 1; i >= 0; i-- {
			closeErr = errors.Join(closeErr, closers[i]())
		}

		return closeErr
	}

	for _, item := range providers {
		closer, err := item.open(ctx, cfg, &deps)
		if err != nil {
			closeAll()
			return Deps{}, nil, fmt.Errorf("mounting %s: %w", item.name, err)
		}

		if closer == nil {
			continue
		}

		closers = append(closers, closer)
		mounted = append(mounted, item.name)
	}

	slog.Info("dependencies mounted", "deps", mounted)

	return deps, closeAll, nil
}

func openDatabase(ctx context.Context, cfg config.Config, deps *Deps) (func() error, error) {
	db, err := database.Open(ctx, database.Config{
		DSN: cfg.DatabaseDSN,
	})

	if err != nil {
		return nil, err
	}

	deps.DB = db

	return db.Close, nil
}
