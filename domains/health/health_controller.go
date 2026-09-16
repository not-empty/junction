package health

import (
	"net/http"

	"github.com/not-empty/bridge/request"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Check(w http.ResponseWriter, r *http.Request) {
	request.Response(
		w,
		http.StatusOK,
		map[string]string{"status": "Ok"},
		nil,
	)
}
