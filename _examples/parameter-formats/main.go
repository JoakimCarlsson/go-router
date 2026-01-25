package main

import (
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

func main() {
	r := router.New()

	r.GET("/users/{id}", getUserById,
		openapi.WithTags("Users"),
		openapi.WithSummary("Get user by ID"),
		openapi.WithIntegerPathParam("id", true, "User ID number", 42, nil, nil),
	)

	r.GET("/users/by-uuid/{uuid}", getUserByUUID,
		openapi.WithTags("Users"),
		openapi.WithSummary("Get user by UUID"),
		openapi.WithUUIDPathParam("uuid", true, "User UUID", "123e4567-e89b-12d3-a456-426614174000"),
	)

	r.GET("/products/{code}", getProductByCode,
		openapi.WithTags("Products"),
		openapi.WithSummary("Get product by code"),
		openapi.WithRegexPathParam("code", "^[A-Z]{2}[0-9]{4}$", true, "Product code (XX0000)", "AB1234"),
	)

	r.GET("/orders/{status}", getOrdersByStatus,
		openapi.WithTags("Orders"),
		openapi.WithSummary("Get orders by status"),
		openapi.WithEnumPathParam("status", true, "Order status", "pending", []interface{}{"pending", "shipped", "delivered", "canceled"}),
	)

	r.GET("/products", getProducts,
		openapi.WithTags("Products"),
		openapi.WithSummary("List products"),
		openapi.WithFormattedQueryParam("category", "string", false, "Product category", "electronics"),
		openapi.WithNumericPathParam("minPrice", false, "Minimum price", 10.0, floatPtr(0.0), nil),
		openapi.WithNumericPathParam("maxPrice", false, "Maximum price", 100.0, nil, nil),
		openapi.WithDateQueryParam("updatedSince", false, "Filter by update date", "2023-01-01"),
		openapi.WithRegexQueryParam("search", "^[\\w\\s]{3,50}$", false, "Search term", "bluetooth speaker"),
	)

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Advanced Parameter Formats API",
		Version:     "1.0.0",
		Description: "API demonstrating various parameter formats",
	})

	uiConfig := swaggerui.DefaultUIConfig()
	uiConfig.Title = "Parameter Formats API"
	uiConfig.DocExpansion = "list"

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(uiConfig)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func getUserById(c *router.Context) {
	c.JSON(200, map[string]interface{}{"id": c.Param("id"), "name": "Example User"})
}

func getUserByUUID(c *router.Context) {
	c.JSON(200, map[string]interface{}{"uuid": c.Param("uuid"), "name": "UUID User"})
}

func getProductByCode(c *router.Context) {
	c.JSON(200, map[string]interface{}{"code": c.Param("code"), "name": "Example Product"})
}

func getOrdersByStatus(c *router.Context) {
	c.JSON(200, []map[string]interface{}{{"id": "order-1", "status": c.Param("status")}, {"id": "order-2", "status": c.Param("status")}})
}

func getProducts(c *router.Context) {
	c.JSON(200, []map[string]interface{}{
		{"id": "product-1", "name": "Product 1", "category": c.QueryDefault("category", "all")},
		{"id": "product-2", "name": "Product 2", "category": c.QueryDefault("category", "all")},
	})
}

func floatPtr(v float64) *float64 { return &v }
