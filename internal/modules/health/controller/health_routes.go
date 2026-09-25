package controller

import (
	"net/http"

	"github.com/not-empty/bridge/platform/bootstrap"
)

func Module(_ bootstrap.Deps, mux *http.ServeMux) {
	registerRoutes(mux, NewController())
}

// Kept apart from Module so a test can mount the routes itself.
func registerRoutes(mux *http.ServeMux, controller *Controller) {
	mux.HandleFunc("GET /health", controller.Check)
}
