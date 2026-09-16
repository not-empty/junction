package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	defaultPort         = "8080"
	defaultMaxBodyBytes = 1 << 20
)

type Config struct {
	Port               string
	CORSAllowedOrigins []string
	MaxBodyBytes       int64
	DatabaseDSN        string
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("loading .env: %w", err)
	}

	cfg := Config{
		Port:         defaultPort,
		MaxBodyBytes: defaultMaxBodyBytes,
	}

	if port := os.Getenv("PORT"); port != "" {
		_, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return Config{}, fmt.Errorf("PORT is invalid: %q", port)
		}

		cfg.Port = port
	}

	for _, origin := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)

		if origin != "" {
			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, origin)
		}
	}

	if len(cfg.CORSAllowedOrigins) == 0 {
		return Config{}, errors.New("CORS_ALLOWED_ORIGINS is not configured")
	}

	if rawMaxBody := os.Getenv("MAX_BODY_BYTES"); rawMaxBody != "" {
		maxBody, err := strconv.ParseInt(rawMaxBody, 10, 64)
		if err != nil || maxBody <= 0 {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES is invalid: %q", rawMaxBody)
		}

		cfg.MaxBodyBytes = maxBody
	}

	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	if cfg.DatabaseDSN == "" {
		return Config{}, errors.New("DATABASE_DSN is not configured")
	}

	return cfg, nil
}
