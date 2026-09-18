package controller

import "net/http"

func RegisterRoutes(router *http.ServeMux, controller *Controller) {
	router.HandleFunc("GET /health", controller.Check)
}
