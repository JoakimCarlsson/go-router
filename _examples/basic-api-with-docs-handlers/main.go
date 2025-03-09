package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/handlers/create"
	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/handlers/delete"
	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/handlers/edit"
	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/handlers/get"
	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/handlers/get_all"
	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/handlers/health"
	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/store"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

func main() {
	productStore := store.NewProductStore()
	r := router.New()

	// Logger middleware
	r.Use(loggerMiddleware)

	// Public endpoints
	r.GET("/health", health.Handler(), health.RouteOptions()...)

	// Products endpoints
	r.GET("/products", get_all.Handler(productStore), get_all.RouteOptions()...)
	r.POST("/products", create.Handler(productStore), create.RouteOptions()...)
	r.GET("/products/{id}", get.Handler(productStore), get.RouteOptions()...)
	r.PUT("/products/{id}", edit.Handler(productStore), edit.RouteOptions()...)
	r.DELETE("/products/{id}", delete.Handler(productStore), delete.RouteOptions()...)

	// Create OpenAPI generator
	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Product Catalog API",
		Version:     "1.0.0",
		Description: "A sample product catalog API built with go-router",
	})

	// Configure Swagger UI with specific settings for pointer fields
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.DefaultModelRendering = "example"
	uiConfig.Title = "Product Catalog API"

	// Set up the integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Middleware for logging requests
func loggerMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		start := time.Now()

		// Process request
		next(c)

		// Log after request is processed
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s - %d (%v)\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.StatusCode,
			duration,
		)
	}
}
