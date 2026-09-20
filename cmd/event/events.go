package main

import (
	"github.com/not-empty/bridge/platform/database"
	"github.com/not-empty/bridge/platform/event"
)

func newRegistry(db *database.DB) event.Registry {
	registry := event.Registry{}

	return registry
}
