package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultMigrationsTable = "schema_migrations"

	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
)

type Config struct {
	DatabaseDriver  string
	DatabaseDSN     string
	MigrationsTable string
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("loading .env: %w", err)
	}

	driver, err := Driver()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		DatabaseDriver:  driver,
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

// Driver reads DATABASE_DRIVER from the environment, defaulting to MySQL. It
// does not load .env: the caller decides whether it is loaded.
func Driver() (string, error) {
	driver := os.Getenv("DATABASE_DRIVER")

	switch driver {
	case "":
		return DriverMySQL, nil
	case DriverMySQL, DriverPostgres:
		return driver, nil
	default:
		return "", fmt.Errorf("invalid DATABASE_DRIVER %q: must be %q or %q", driver, DriverMySQL, DriverPostgres)
	}
}
