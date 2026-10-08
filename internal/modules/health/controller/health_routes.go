package controller

import (
	"net/http"

	"github.com/not-empty/junction/platform/bootstrap"
)

func Module(_ bootstrap.Deps, mux *http.ServeMux) {
	registerRoutes(mux, NewController())
}

func registerRoutes(mux *http.ServeMux, controller *Controller) {
	mux.HandleFunc("GET /health", controller.Check)
}
