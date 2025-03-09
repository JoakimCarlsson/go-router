package health

import (
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// Handler provides a health check endpoint
func Handler() router.HandlerFunc {
	return func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

// RouteOptions returns the route options for this handler
func RouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("System"),
		docs.WithSummary("Health check endpoint"),
		docs.WithDescription("Returns the health status of the API"),
		docs.WithResponse(200, "API is healthy"),
	}
}
