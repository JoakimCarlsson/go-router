package v2

import (
	"github.com/joakimcarlsson/go-router/pkg/router"
)

type Router struct {
	handler *Handler
}

func NewRouter() *Router {
	return &Router{
		handler: NewHandler(),
	}
}

func (r *Router) Register(rtr *router.Router) {
	rtr.GET("/health", r.handler.HandleHealthCheck)
}
