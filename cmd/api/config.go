package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/not-empty/bridge/platform/config"
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

func loadConfig() (Config, error) {
	baseCfg, err := config.Load()
	if err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	cfg := Config{
		Port:         defaultPort,
		MaxBodyBytes: defaultMaxBodyBytes,
		DatabaseDSN:  baseCfg.DatabaseDSN,
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

	return cfg, nil
}
