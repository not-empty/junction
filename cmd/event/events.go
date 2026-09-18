package main

import (
	userlistener "github.com/not-empty/bridge/internal/modules/user/listener"
	userrepository "github.com/not-empty/bridge/internal/modules/user/repository"
	userservice "github.com/not-empty/bridge/internal/modules/user/service"
	"github.com/not-empty/bridge/platform/database"
	"github.com/not-empty/bridge/platform/event"
)

func newRegistry(db *database.DB) event.Registry {
	registry := event.Registry{}

	userlistener.RegisterEvents(registry, userlistener.NewUserListener(
		userservice.NewUserService(
			userrepository.NewUserRepository(db),
		),
	))

	return registry
}
