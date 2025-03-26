package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/router/middleware"
)

func main() {
	// Create a new router instance
	r := router.New()

	// Add CORS middleware with default configuration
	// This will allow all origins with default methods
	r.Use(middleware.CORS())

	// Root group with default CORS
	r.GET("/", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Hello from go-router with CORS middleware!",
		})
	})

	// API group with custom CORS configuration
	apiRouter := router.New()

	// Custom CORS configuration for API routes
	// This is more restrictive and allows specific origins
	apiCorsConfig := middleware.CORSConfig{
		AllowOrigins:     []string{"https://example.com", "https://api.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}

	apiRouter.Use(middleware.CORSWithConfig(apiCorsConfig))

	apiRouter.GET("/users", func(c *router.Context) {
		c.JSON(http.StatusOK, []map[string]string{
			{"id": "1", "name": "John Doe"},
			{"id": "2", "name": "Jane Doe"},
		})
	})

	apiRouter.POST("/users", func(c *router.Context) {
		c.JSON(http.StatusCreated, map[string]string{
			"id":      "3",
			"message": "User created successfully",
		})
	})

	// Mount the API router
	r.Group("/api", func(api *router.Router) {
		// Routes from apiRouter will be available under /api
		// Their CORS settings will be respected

		api.GET("/public", func(c *router.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"message": "This is public API with different CORS settings",
			})
		})
	})

	// Create a third group with wildcard origin support
	adminRouter := router.New()
	adminCorsConfig := middleware.CORSConfig{
		AllowOrigins:     []string{"https://*.admin.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"*"}, // Allow all headers
		AllowCredentials: true,
	}
	adminRouter.Use(middleware.CORSWithConfig(adminCorsConfig))

	// Mount the admin router
	r.Group("/admin", func(admin *router.Router) {
		admin.GET("/dashboard", func(c *router.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"message": "Admin dashboard with domain wildcard CORS",
			})
		})
	})

	// Start the server
	fmt.Println("CORS example server starting on http://localhost:8080")
	fmt.Println("Try requests from different origins to see CORS in action")
	fmt.Println("Default CORS (all origins): http://localhost:8080/")
	fmt.Println("Restricted CORS (specific origins): http://localhost:8080/api/users")
	fmt.Println("Wildcard CORS (*.admin.example.com): http://localhost:8080/admin/dashboard")

	log.Fatal(http.ListenAndServe(":8080", apiRouter))
}
