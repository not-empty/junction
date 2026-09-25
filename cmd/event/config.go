package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/not-empty/bridge/platform/config"
)

const (
	defaultGroupID     = "bridge"
	defaultMaxAttempts = 5
	defaultBackoff     = 1 * time.Second
	defaultDLQSuffix   = ".dlq"
	handlerTimeout     = 30 * time.Second
)

type Config struct {
	KafkaBrokers   []string
	KafkaGroupID   string
	HandlerTimeout time.Duration
}

func loadConfig() (Config, error) {
	// Loads .env, which the os.Getenv calls below depend on.
	_, err := config.Load()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		KafkaGroupID:   defaultGroupID,
		HandlerTimeout: handlerTimeout,
	}

	for _, broker := range strings.Split(os.Getenv("KAFKA_BROKERS"), ",") {
		broker = strings.TrimSpace(broker)
		if broker != "" {
			cfg.KafkaBrokers = append(cfg.KafkaBrokers, broker)
		}
	}

	if len(cfg.KafkaBrokers) == 0 {
		return Config{}, errors.New("KAFKA_BROKERS is not configured")
	}

	if groupID := os.Getenv("KAFKA_GROUP_ID"); groupID != "" {
		cfg.KafkaGroupID = groupID
	}

	return cfg, nil
}
