package v1

import (
	"net/http"

	"github.com/joakimcarlsson/go-router/pkg/router"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleHealthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
