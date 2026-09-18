package controller

import "net/http"

func RegisterRoutes(router *http.ServeMux, userController *UserController) {
	router.HandleFunc("POST /user/add", userController.Create)
	router.HandleFunc("GET /user/list", userController.List)
	router.HandleFunc("GET /user/{id}", userController.GetOne)
	router.HandleFunc("PATCH /user/{id}", userController.Update)
	router.HandleFunc("DELETE /user/{id}", userController.Delete)
}
