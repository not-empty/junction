package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseDSN string
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("loading .env: %w", err)
	}

	cfg := Config{
		DatabaseDSN: os.Getenv("DATABASE_DSN"),
	}

	if cfg.DatabaseDSN == "" {
		return Config{}, errors.New("DATABASE_DSN is not configured")
	}

	return cfg, nil
}
