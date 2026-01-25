package main

import (
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

type Resource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	r := router.New()

	r.GET("/health", healthCheck,
		openapi.WithTags("Health"),
		openapi.WithSummary("Health check endpoint"),
		openapi.WithResponse(http.StatusOK, "Service is healthy"),
	)

	r.GET("/resources", listResources,
		openapi.WithTags("Resources"),
		openapi.WithSummary("List all resources"),
		openapi.WithJSONResponse[[]Resource](http.StatusOK, "List of resources"),
		openapi.WithOAuth2Scopes("read"),
	)

	r.GET("/resources/{id}", getResource,
		openapi.WithTags("Resources"),
		openapi.WithSummary("Get resource by ID"),
		openapi.WithPathParam("id", "string", true, "Resource ID", nil),
		openapi.WithJSONResponse[Resource](http.StatusOK, "Resource details"),
		openapi.WithOAuth2Scopes("read"),
	)

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "OAuth2 Implicit Flow API",
		Version:     "1.0.0",
		Description: "API demonstrating OAuth2 Implicit Flow",
	})

	generator.WithOAuth2ImplicitFlow(
		"oauth2",
		"OAuth2 Implicit Flow",
		"https://your-auth-server.com/authorize",
		map[string]string{
			"read":  "Read access to resources",
			"write": "Write access to resources",
		},
	)

	uiConfig := swaggerui.DefaultUIConfig()
	uiConfig.Title = "OAuth2 Implicit Flow Demo"
	uiConfig.TryItOutEnabled = true
	uiConfig.OAuth2Config = &swaggerui.OAuth2Config{
		ClientID: "your-implicit-client-id",
		AppName:  "OAuth2 Implicit Demo App",
		Scopes:   `"read write"`,
	}

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(uiConfig)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}

func listResources(c *router.Context) {
	c.JSON(http.StatusOK, []Resource{
		{ID: "1", Name: "Resource 1", Description: "First resource"},
		{ID: "2", Name: "Resource 2", Description: "Second resource"},
		{ID: "3", Name: "Resource 3", Description: "Third resource"},
	})
}

func getResource(c *router.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, Resource{ID: id, Name: "Resource " + id, Description: "This is resource " + id})
}
