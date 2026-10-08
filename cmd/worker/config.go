package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/not-empty/junction/platform/config"
)

const (
	pollIntervalSeconds    = 1.0
	promoteIntervalSeconds = 5.0
	reapIntervalSeconds    = 5.0
	jobTimeout             = 30 * time.Second
)

type Config struct {
	RedisPort              int
	RedisHost              string
	Queue                  string
	PollIntervalSeconds    float64
	PromoteIntervalSeconds float64
	ReapIntervalSeconds    float64
	JobTimeout             time.Duration
}

func loadConfig() (Config, error) {
	_, err := config.Load()
	if err != nil {
		return Config{}, err
	}

	rawRedisPort := os.Getenv("REDIS_PORT")
	redisPort, err := strconv.Atoi(rawRedisPort)
	if err != nil {
		return Config{}, fmt.Errorf("REDIS_PORT is invalid: %q", rawRedisPort)
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		return Config{}, errors.New("REDIS_HOST is not configured")
	}

	queue := flag.String("queue", "", "Queue name")
	flag.Parse()

	if *queue == "" {
		return Config{}, errors.New("queue flag is required")
	}

	cfg := Config{
		RedisPort:              redisPort,
		RedisHost:              redisHost,
		Queue:                  *queue,
		PollIntervalSeconds:    pollIntervalSeconds,
		PromoteIntervalSeconds: promoteIntervalSeconds,
		ReapIntervalSeconds:    reapIntervalSeconds,
		JobTimeout:             jobTimeout,
	}

	return cfg, nil
}
