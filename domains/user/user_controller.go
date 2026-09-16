package user

import (
	"net/http"

	"github.com/not-empty/bridge/request"
)

type Controller struct {
	service ServiceInterface
}

func NewController(service ServiceInterface) *Controller {
	return &Controller{
		service: service,
	}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var userRegister UserRegister

	err := request.ValidateData(r.Body, &userRegister)

	if err != nil {
		request.ResponseError(w, err)
		return
	}

	user, err := c.service.Create(r.Context(), userRegister)
	request.Response(w, http.StatusCreated, NewUserResponse(user), err)
}
