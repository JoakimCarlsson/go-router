package server

import (
	"log/slog"
	"net/http"

	healthv1 "github.com/joakimcarlsson/go-router/example/internal/features/health/v1"
	healthv2 "github.com/joakimcarlsson/go-router/example/internal/features/health/v2"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

type Server struct {
	router *router.Router
	slog   *slog.Logger
}

func NewServer(slog *slog.Logger) *Server {
	return &Server{
		router: router.New(),
		slog:   slog,
	}
}

func (s *Server) RegisterRoutes() http.Handler {
	// V1 API routes
	s.router.Group("/api/v1", func(v1 *router.Router) {
		healthv1.NewRouter().Register(v1)
	})

	// V2 API routes
	s.router.Group("/api/v2", func(v2 *router.Router) {
		healthv2.NewRouter().Register(v2)
	})

	return s.router
}

func (s *Server) HTTP() *http.Server {
	return &http.Server{
		Addr:    ":6784",
		Handler: s.RegisterRoutes(),
	}
}
