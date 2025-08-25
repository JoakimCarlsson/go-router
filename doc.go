/*
Package router provides a high-performance, modular HTTP router for Go with built-in OpenAPI 3.0 and Swagger UI support.

# Quick Start

	r := router.New()
	
	r.GET("/users/{id}", func(c *router.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"user_id": id})
	})
	
	r.Run(":8080")

# Features

- RESTful HTTP routing with path parameters (/users/{id})
- Standard middleware compatibility (http.Handler)
- Route groups for organization
- JSON, XML, and plain text responses
- File upload support with multipart forms
- Server-Sent Events (SSE) for real-time communication
- OpenAPI 3.0 spec generation with type safety
- Interactive Swagger UI documentation
- High performance with zero-allocation hot paths
- Object pooling for minimal GC pressure

# API Documentation

Add OpenAPI documentation to routes:

	r.GET("/users/{id}", getUser,
		docs.WithSummary("Get user by ID"),
		docs.WithPathParam("id", "string", true, "User ID", "123"),
		docs.WithJSONResponse[User](200, "User found"),
	)

Setup interactive documentation:

	integration.Setup(r, integration.SetupOptions{
		Title:    "My API",
		Version:  "1.0.0",
		DocsPath: "/docs",
	})

# Middleware

Standard HTTP middleware support:

	r.Use(cors.Default())
	
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Custom middleware logic
			next.ServeHTTP(w, req)
		})
	})

# Route Groups

	r.Group("/api/v1", func(api *router.Router) {
		api.Use(authMiddleware)
		api.GET("/users", listUsers)
		api.POST("/users", createUser)
	})

For comprehensive documentation, examples, and guides, see: https://github.com/JoakimCarlsson/go-router
*/
package router
