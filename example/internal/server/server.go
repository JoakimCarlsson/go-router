package server

import (
	"log/slog"
	"net/http"

	"github.com/joakimcarlsson/go-router/example/internal/middleware"
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
	s.router.Use(middleware.Logger(s.slog))

	return s.router
}

func (s *Server) HTTP() *http.Server {
	return &http.Server{
		Addr:    ":6784",
		Handler: s.RegisterRoutes(),
	}
}
