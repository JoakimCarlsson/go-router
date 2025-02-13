package middleware

import (
	"log/slog"
	"time"

	"github.com/joakimcarlsson/go-router/pkg/router"
)

func Logger(log *slog.Logger) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			start := time.Now()
			path := c.Request.URL.Path

			next(c)

			duration := time.Since(start)
			log.Info("request completed",
				"method", c.Request.Method,
				"path", path,
				"status", c.StatusCode,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}
}
