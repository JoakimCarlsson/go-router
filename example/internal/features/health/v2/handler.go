package v2

import (
	"net/http"

	"github.com/joakimcarlsson/go-router/pkg/router"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleHealthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": "2.0",
		"details": map[string]string{
			"environment": "development",
			"timestamp":   http.TimeFormat,
		},
	})
}
