package main

import (
	"net/http"

	healthcontroller "github.com/not-empty/bridge/internal/modules/health/controller"
	usercontroller "github.com/not-empty/bridge/internal/modules/user/controller"
	userrepository "github.com/not-empty/bridge/internal/modules/user/repository"
	userservice "github.com/not-empty/bridge/internal/modules/user/service"
	"github.com/not-empty/bridge/platform/database"
)

func newRouter(db *database.DB) *http.ServeMux {
	router := http.NewServeMux()

	healthcontroller.RegisterRoutes(router, healthcontroller.NewController())

	usercontroller.RegisterRoutes(router, usercontroller.NewUserController(
		userservice.NewUserService(
			userrepository.NewUserRepository(db),
		),
	))

	return router
}
