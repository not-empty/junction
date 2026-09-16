package routes

import (
	"net/http"

	"github.com/not-empty/bridge/database"
	"github.com/not-empty/bridge/domains/user"
)

func SetUserRoutes(router *http.ServeMux, db *database.DB) *http.ServeMux {
	userController := user.NewController(
		user.NewService(
			user.NewRepository(db),
		),
	)

	router.HandleFunc("POST /user/add", userController.Create)

	return router
}
