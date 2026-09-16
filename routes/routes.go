package routes

import (
	"net/http"

	"github.com/not-empty/bridge/database"
	"github.com/not-empty/bridge/domains/health"
)

func SetRoutes(db *database.DB) *http.ServeMux {
	router := http.NewServeMux()

	healthController := health.NewController()
	router.HandleFunc("GET /health", healthController.Check)

	SetUserRoutes(router, db)

	return router
}
