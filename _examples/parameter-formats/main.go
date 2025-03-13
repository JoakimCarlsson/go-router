package main

import (
	"fmt"
	"net/http"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

func main() {
	r := router.New()

	r.GET("/users/{id}", getUserById,
		docs.WithTags("Users"),
		docs.WithSummary("Get user by ID"),
		docs.WithDescription("Retrieves a user by their unique identifier"),
		docs.WithIntegerPathParam("id", true, "User ID number", 42, nil, nil),
	)

	r.GET("/users/by-uuid/{uuid}", getUserByUUID,
		docs.WithTags("Users"),
		docs.WithSummary("Get user by UUID"),
		docs.WithDescription("Retrieves a user by their UUID"),
		docs.WithUUIDPathParam("uuid", true, "User UUID", "123e4567-e89b-12d3-a456-426614174000"),
	)

	r.GET("/products/{code}", getProductByCode,
		docs.WithTags("Products"),
		docs.WithSummary("Get product by code"),
		docs.WithDescription("Retrieves a product using its unique product code"),
		docs.WithRegexPathParam("code", "^[A-Z]{2}[0-9]{4}$", true, "Product code in format XX0000", "AB1234"),
	)

	r.GET("/orders/{status}", getOrdersByStatus,
		docs.WithTags("Orders"),
		docs.WithSummary("Get orders by status"),
		docs.WithDescription("Retrieves orders with a specific status"),
		docs.WithEnumPathParam("status", true, "Order status", "pending", []interface{}{"pending", "shipped", "delivered", "canceled"}),
	)

	r.GET("/products", getProducts,
		docs.WithTags("Products"),
		docs.WithSummary("List products"),
		docs.WithDescription("Retrieves a list of products with optional filtering"),
		docs.WithFormattedQueryParam("category", "string", false, "Product category", "electronics"),
		docs.WithNumericPathParam("minPrice", false, "Minimum product price", 10.0, floatPtr(0.0), nil),
		docs.WithNumericPathParam("maxPrice", false, "Maximum product price", 100.0, nil, nil),
		docs.WithDateQueryParam("updatedSince", false, "Filter by last update date", "2023-01-01"),
		docs.WithRegexQueryParam("search", "^[\\w\\s]{3,50}$", false, "Search term (3-50 alphanumeric chars)", "bluetooth speaker"),
	)

	// Create the OpenAPI generator
	generator := openapi.NewGenerator(metadata.Info{
		Title:       "Advanced Parameter Formats API",
		Version:     "1.0.0",
		Description: "API demonstrating various parameter formats and patterns in OpenAPI specification",
	})

	// Configure Swagger UI
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(swagger.UIConfig{
		Title:                 "Parameter Formats API Documentation",
		SpecURL:               "/openapi.json",
		SwaggerVersion:        "4.18.3",
		DefaultModelRendering: "example",
		DeepLinking:           true,
		DocExpansion:          "list",
	})

	// Register OpenAPI and Swagger UI routes
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	// Start the server
	fmt.Println("Server starting on :8080")
	fmt.Println("Documentation available at http://localhost:8080/docs")
	http.ListenAndServe(":8080", r)
}

// Handlers
func getUserById(c *router.Context) {
	id := c.Param("id")
	c.JSON(200, map[string]interface{}{
		"id":   id,
		"name": "Example User",
	})
}

func getUserByUUID(c *router.Context) {
	uuid := c.Param("uuid")
	c.JSON(200, map[string]interface{}{
		"uuid": uuid,
		"name": "UUID User",
	})
}

func getProductByCode(c *router.Context) {
	code := c.Param("code")
	c.JSON(200, map[string]interface{}{
		"code":        code,
		"name":        "Example Product",
		"description": "Product with code: " + code,
	})
}

func getOrdersByStatus(c *router.Context) {
	status := c.Param("status")
	c.JSON(200, []map[string]interface{}{
		{
			"id":     "order-1",
			"status": status,
		},
		{
			"id":     "order-2",
			"status": status,
		},
	})
}

func getProducts(c *router.Context) {
	category := c.QueryDefault("category", "all")
	search := c.QueryDefault("search", "")
	c.JSON(200, []map[string]interface{}{
		{
			"id":       "product-1",
			"name":     "Example Product 1",
			"category": category,
			"search":   search,
		},
		{
			"id":       "product-2",
			"name":     "Example Product 2",
			"category": category,
			"search":   search,
		},
	})
}

// Helper function to create float pointer
func floatPtr(v float64) *float64 {
	return &v
}
