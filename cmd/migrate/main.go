package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/not-empty/junction/platform/config"
	"github.com/not-empty/junction/platform/migration"
)

const usage = `migrate - applies the versioned schema in migrations/

usage:
  migrate up       apply every pending migration
  migrate down     roll back the last migration
  migrate status   print the current version
`

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	err := run(os.Args[1:])
	if err != nil {
		slog.Error("migration stopped", "error", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usage)
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch args[0] {
	case "up":
		return up(cfg.DatabaseDSN)
	case "down":
		return down(cfg.DatabaseDSN)
	case "status":
		return status(cfg.DatabaseDSN)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func up(dsn string) error {
	before, _, err := migration.Version(dsn)
	if err != nil {
		return err
	}

	err = migration.Up(dsn)
	if errors.Is(err, migration.ErrNoChange) {
		slog.Info("schema already up to date", "version", before)
		return nil
	}

	if err != nil {
		return err
	}

	after, _, err := migration.Version(dsn)
	if err != nil {
		return err
	}

	slog.Info("migrations applied", "from", before, "to", after)
	return nil
}

func down(dsn string) error {
	before, _, err := migration.Version(dsn)
	if err != nil {
		return err
	}

	err = migration.Down(dsn)
	if errors.Is(err, migration.ErrNoChange) {
		slog.Info("nothing to roll back")
		return nil
	}

	if err != nil {
		return err
	}

	after, _, err := migration.Version(dsn)
	if err != nil {
		return err
	}

	slog.Info("migration rolled back", "from", before, "to", after)
	return nil
}

func status(dsn string) error {
	version, dirty, err := migration.Version(dsn)
	if err != nil {
		return err
	}

	slog.Info("schema status", "version", version, "dirty", dirty)
	return nil
}
