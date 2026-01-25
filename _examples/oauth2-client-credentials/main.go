package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

type ServiceMetrics struct {
	Uptime          string           `json:"uptime"`
	RequestCount    int64            `json:"requestCount"`
	ErrorCount      int64            `json:"errorCount"`
	AverageLatency  float64          `json:"averageLatency"`
	EndpointMetrics map[string]int64 `json:"endpointMetrics"`
}

type ServiceStatus struct {
	Status      string         `json:"status"`
	Version     string         `json:"version"`
	Environment string         `json:"environment"`
	Metrics     ServiceMetrics `json:"metrics"`
}

var maintenanceMode bool

func main() {
	r := router.New()

	r.GET("/health", healthCheck,
		openapi.WithTags("Health"),
		openapi.WithSummary("Basic health check"),
		openapi.WithResponse(http.StatusOK, "Service is healthy"),
	)

	r.GET("/status", getStatus,
		openapi.WithTags("Status"),
		openapi.WithSummary("Get detailed service status"),
		openapi.WithJSONResponse[ServiceStatus](http.StatusOK, "Service status"),
		openapi.WithOAuth2Scopes("status:read"),
	)

	r.POST("/maintenance/start", startMaintenance,
		openapi.WithTags("Maintenance"),
		openapi.WithSummary("Start maintenance mode"),
		openapi.WithOAuth2Scopes("maintenance:write"),
	)

	r.POST("/maintenance/end", endMaintenance,
		openapi.WithTags("Maintenance"),
		openapi.WithSummary("End maintenance mode"),
		openapi.WithOAuth2Scopes("maintenance:write"),
	)

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Service API with Client Credentials Flow",
		Version:     "1.0.0",
		Description: "API demonstrating OAuth2 Client Credentials Flow",
	})

	generator.WithOAuth2ClientCredentialsFlow(
		"oauth2",
		"OAuth2 Client Credentials Flow",
		"https://your-auth-server.com/token",
		map[string]string{
			"status:read":       "Read status information",
			"maintenance:write": "Perform maintenance operations",
		},
	)

	uiConfig := swaggerui.DefaultUIConfig()
	uiConfig.Title = "Service API Documentation"
	uiConfig.TryItOutEnabled = true
	uiConfig.OAuth2Config = &swaggerui.OAuth2Config{
		ClientID:     "your-service-client-id",
		ClientSecret: "your-service-client-secret",
		Scopes:       `"status:read maintenance:write"`,
	}

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(uiConfig)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func getStatus(c *router.Context) {
	c.JSON(http.StatusOK, ServiceStatus{
		Status:      "operational",
		Version:     "1.2.0",
		Environment: "production",
		Metrics: ServiceMetrics{
			Uptime:         "2d 3h 45m",
			RequestCount:   45892,
			ErrorCount:     23,
			AverageLatency: 42.7,
			EndpointMetrics: map[string]int64{
				"/api/v1/users":    15234,
				"/api/v1/products": 30658,
			},
		},
	})
}

func startMaintenance(c *router.Context) {
	maintenanceMode = true
	c.JSON(http.StatusOK, map[string]interface{}{"maintenanceMode": true, "timestamp": time.Now(), "message": "Maintenance mode activated"})
}

func endMaintenance(c *router.Context) {
	maintenanceMode = false
	c.JSON(http.StatusOK, map[string]interface{}{"maintenanceMode": false, "timestamp": time.Now(), "message": "Maintenance mode deactivated"})
}
