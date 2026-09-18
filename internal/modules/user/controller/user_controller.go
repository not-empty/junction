package controller

import (
	"net/http"

	"github.com/not-empty/bridge/internal/modules/user/domain"
	"github.com/not-empty/bridge/internal/modules/user/service"
	"github.com/not-empty/bridge/platform/httpx"
)

type UserController struct {
	service service.UserServiceInterface
}

func NewUserController(service service.UserServiceInterface) *UserController {
	return &UserController{
		service: service,
	}
}

func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var userRegister domain.UserRegister

	err := httpx.ValidateData(r.Body, &userRegister)

	if err != nil {
		httpx.ResponseError(w, err)
		return
	}

	response, err := c.service.Create(r.Context(), userRegister)
	httpx.Response(w, http.StatusCreated, response, err)
}

func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	var userUpdate domain.UserUpdate

	err := httpx.ValidateData(r.Body, &userUpdate)

	if err != nil {
		httpx.ResponseError(w, err)
		return
	}

	err = c.service.Update(r.Context(), r.PathValue("id"), userUpdate)
	httpx.Response(w, http.StatusNoContent, nil, err)
}

func (c *UserController) GetOne(w http.ResponseWriter, r *http.Request) {
	response, err := c.service.GetOne(r.Context(), r.PathValue("id"))
	httpx.Response(w, http.StatusOK, response, err)
}

func (c *UserController) List(w http.ResponseWriter, r *http.Request) {
	response, err := c.service.List(r.Context())
	httpx.Response(w, http.StatusOK, response, err)
}

func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	err := c.service.Delete(r.Context(), r.PathValue("id"))
	httpx.Response(w, http.StatusNoContent, nil, err)
}
