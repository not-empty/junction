package main

import (
	"net/http"

	healthcontroller "github.com/not-empty/bridge/internal/modules/health/controller"
	"github.com/not-empty/bridge/platform/database"
)

func newRouter(db *database.DB) *http.ServeMux {
	router := http.NewServeMux()

	healthcontroller.RegisterRoutes(router, healthcontroller.NewController())

	return router
}
