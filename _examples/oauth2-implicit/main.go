package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// Resource represents a protected resource
type Resource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	r := router.New()

	// Public endpoints
	r.GET("/health", healthCheck,
		docs.WithTags("Health"),
		docs.WithSummary("Health check endpoint"),
		docs.WithDescription("Returns the health status of the API"),
		docs.WithResponse(200, "Service is healthy"),
	)

	// Protected endpoints that require authentication
	r.GET("/resources", listResources,
		docs.WithTags("Resources"),
		docs.WithSummary("List all resources"),
		docs.WithDescription("Returns a list of all resources (requires authentication)"),
		docs.WithResponse(200, "Resources retrieved successfully"),
		docs.WithJSONResponse[[]Resource](200, "List of resources"),
		docs.WithOAuth2Scopes("read"),
	)

	r.GET("/resources/{id}", getResource,
		docs.WithTags("Resources"),
		docs.WithSummary("Get resource by ID"),
		docs.WithDescription("Retrieves a resource by its unique identifier (requires authentication)"),
		docs.WithPathParam("id", "string", true, "Resource ID", nil),
		docs.WithResponse(200, "Resource found"),
		docs.WithJSONResponse[Resource](200, "Resource details"),
		docs.WithResponse(404, "Resource not found"),
		docs.WithOAuth2Scopes("read"),
	)

	// Create OpenAPI generator with API information
	generator := openapi.NewGenerator(openapi.Info{
		Title:       "OAuth2 Implicit Flow API",
		Version:     "1.0.0",
		Description: "API demonstrating OAuth2 Implicit Flow authentication with go-router",
	})

	// Configure OAuth2 Implicit Flow
	generator.WithOAuth2ImplicitFlow(
		"oauth2",
		"OAuth2 Implicit Flow",
		"https://your-auth-server.com/authorize",
		map[string]string{
			"read":  "Read access to resources",
			"write": "Write access to resources",
		},
	)

	// Configure OAuth2 for Swagger UI
	oauth2Config := metadata.NewOAuth2Config().
		WithClientID("your-implicit-client-id").
		WithScopes("read", "write").
		WithAppName("OAuth2 Implicit Demo App")

	// Configure Swagger UI
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.OAuth2Config = oauth2Config
	uiConfig.Title = "OAuth2 Implicit Flow Demo"
	uiConfig.TryItOutEnabled = true
	uiConfig.PersistAuthorization = true

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
		"status":  "ok",
		"version": "1.0.0",
	})
}

func listResources(c *router.Context) {
	// In a real application, you would validate the access token
	// and ensure it has the 'read' scope before providing data

	resources := []Resource{
		{ID: "1", Name: "Resource 1", Description: "First resource"},
		{ID: "2", Name: "Resource 2", Description: "Second resource"},
		{ID: "3", Name: "Resource 3", Description: "Third resource"},
	}
	c.JSON(http.StatusOK, resources)
}

func getResource(c *router.Context) {
	// In a real application, you would validate the access token
	// and ensure it has the 'read' scope before providing data

	id := c.Param("id")

	// Simple example - in production, fetch from a database
	resource := Resource{
		ID:          id,
		Name:        "Resource " + id,
		Description: "This is resource " + id,
	}

	c.JSON(http.StatusOK, resource)
}
