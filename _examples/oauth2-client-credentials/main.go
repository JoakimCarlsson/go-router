package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// ServiceMetrics represents API metrics
type ServiceMetrics struct {
	Uptime          string           `json:"uptime"`
	RequestCount    int64            `json:"requestCount"`
	ErrorCount      int64            `json:"errorCount"`
	AverageLatency  float64          `json:"averageLatency"`
	EndpointMetrics map[string]int64 `json:"endpointMetrics"`
}

// ServiceStatus represents API status information
type ServiceStatus struct {
	Status      string         `json:"status"`
	Version     string         `json:"version"`
	Environment string         `json:"environment"`
	Metrics     ServiceMetrics `json:"metrics"`
}

func main() {
	r := router.New()

	// Public endpoints
	r.GET("/health", healthCheck,
		docs.WithTags("Health"),
		docs.WithSummary("Basic health check"),
		docs.WithDescription("Public endpoint to check if service is running"),
		docs.WithResponse(200, "Service is healthy"),
	)

	// Protected endpoints for service-to-service communication
	r.GET("/status", getStatus,
		docs.WithTags("Status"),
		docs.WithSummary("Get detailed service status"),
		docs.WithDescription("Returns detailed status information (requires service authentication)"),
		docs.WithResponse(200, "Status information retrieved"),
		docs.WithJSONResponse[ServiceStatus](200, "Service status details"),
		docs.WithOAuth2Scopes("status:read"),
	)

	r.POST("/maintenance/start", startMaintenance,
		docs.WithTags("Maintenance"),
		docs.WithSummary("Start maintenance mode"),
		docs.WithDescription("Puts the service into maintenance mode (requires service authentication)"),
		docs.WithResponse(200, "Maintenance mode activated"),
		docs.WithOAuth2Scopes("maintenance:write"),
	)

	r.POST("/maintenance/end", endMaintenance,
		docs.WithTags("Maintenance"),
		docs.WithSummary("End maintenance mode"),
		docs.WithDescription("Takes the service out of maintenance mode (requires service authentication)"),
		docs.WithResponse(200, "Maintenance mode deactivated"),
		docs.WithOAuth2Scopes("maintenance:write"),
	)

	// Create OpenAPI generator
	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Service API with Client Credentials Flow",
		Version:     "1.0.0",
		Description: "API demonstrating OAuth2 Client Credentials Flow for service-to-service authentication",
	})

	// Configure OAuth2 Client Credentials Flow
	generator.WithOAuth2ClientCredentialsFlow(
		"oauth2",
		"OAuth2 Client Credentials Flow",
		"https://your-auth-server.com/token",
		map[string]string{
			"status:read":       "Read status information",
			"maintenance:write": "Perform maintenance operations",
		},
	)

	// Configure OAuth2 for Swagger UI
	oauth2Config := metadata.NewOAuth2Config().
		WithClientID("your-service-client-id").
		WithClientSecret("your-service-client-secret").
		WithScopes("status:read", "maintenance:write")

	// Configure Swagger UI
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.OAuth2Config = oauth2Config
	uiConfig.Title = "Service API Documentation"
	uiConfig.TryItOutEnabled = true

	// Set up Swagger UI integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handler implementations

func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func getStatus(c *router.Context) {
	// In a real application, you would validate the client credentials
	// token and ensure it has the 'status:read' scope before providing data

	status := ServiceStatus{
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
	}

	c.JSON(http.StatusOK, status)
}

var maintenanceMode bool = false

func startMaintenance(c *router.Context) {
	// In a real application, you would validate the client credentials
	// token and ensure it has the 'maintenance:write' scope

	maintenanceMode = true

	c.JSON(http.StatusOK, map[string]interface{}{
		"maintenanceMode": true,
		"timestamp":       time.Now(),
		"message":         "Maintenance mode activated",
	})
}

func endMaintenance(c *router.Context) {
	// In a real application, you would validate the client credentials
	// token and ensure it has the 'maintenance:write' scope

	maintenanceMode = false

	c.JSON(http.StatusOK, map[string]interface{}{
		"maintenanceMode": false,
		"timestamp":       time.Now(),
		"message":         "Maintenance mode deactivated",
	})
}
