package controller

import (
	"net/http"

	"github.com/not-empty/bridge/platform/httpx"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Check(w http.ResponseWriter, r *http.Request) {
	httpx.Response(
		w,
		http.StatusOK,
		map[string]string{"status": "Ok"},
		nil,
	)
}
