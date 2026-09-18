package main

import (
	userconsumer "github.com/not-empty/bridge/internal/modules/user/consumer"
	userrepository "github.com/not-empty/bridge/internal/modules/user/repository"
	userservice "github.com/not-empty/bridge/internal/modules/user/service"
	"github.com/not-empty/bridge/platform/database"
	"github.com/not-empty/bridge/platform/queue"
)

func newRegistry(db *database.DB) queue.Registry {
	registry := queue.Registry{}

	userconsumer.RegisterQueues(registry, userconsumer.NewUserConsumer(
		userservice.NewUserService(
			userrepository.NewUserRepository(db),
		),
	))

	return registry
}
