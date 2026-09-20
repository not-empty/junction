package main

import (
	"github.com/not-empty/bridge/platform/database"
	"github.com/not-empty/bridge/platform/queue"
)

func newRegistry(db *database.DB) queue.Registry {
	registry := queue.Registry{}

	return registry
}
