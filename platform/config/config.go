package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

const defaultMigrationsTable = "schema_migrations"

type Config struct {
	DatabaseDSN     string
	MigrationsTable string
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("loading .env: %w", err)
	}

	cfg := Config{
		DatabaseDSN:     os.Getenv("DATABASE_DSN"),
		MigrationsTable: os.Getenv("MIGRATIONS_TABLE"),
	}

	if cfg.MigrationsTable == "" {
		cfg.MigrationsTable = defaultMigrationsTable
	}

	if cfg.DatabaseDSN == "" {
		return Config{}, errors.New("DATABASE_DSN is not configured")
	}

	return cfg, nil
}
