/*
Package router provides a high-performance, modular HTTP router for Go with built-in OpenAPI 3.0 and Swagger UI support.

# Philosophy

go-router is designed with modularity, type safety, and performance in mind. You can use only the components you need:
- Core routing without documentation (zero-dependency routing)
- Add OpenAPI documentation when needed
- Include Swagger UI for interactive API exploration

# Quick Start

Basic HTTP server with routing:

	package main

	import (
		"github.com/joakimcarlsson/go-router/router"
		"log"
		"net/http"
	)

	func main() {
		r := router.New()
		
		r.GET("/", func(c *router.Context) {
			c.String(200, "Hello, World!")
		})
		
		r.GET("/users/{id}", func(c *router.Context) {
			userID := c.Param("id")
			c.JSON(200, map[string]string{"user": userID})
		})
		
		log.Fatal(http.ListenAndServe(":8080", r))
	}

# Core Features

HTTP Routing:
- RESTful HTTP methods (GET, POST, PUT, DELETE, PATCH)
- Path parameters with Go 1.22+ patterns (/users/{id}, /files/{*path})
- Query parameter helpers with type conversion
- Route groups for logical organization
- Flexible middleware system using standard http.Handler interface

Context and Request Handling:
- Rich context object with request/response utilities
- JSON, XML, and plain text response helpers
- File upload support with multipart form handling
- Server-Sent Events (SSE) support
- Content negotiation
- Request body binding (JSON, XML, forms)

Performance Optimizations:
- Object pooling to minimize allocations
- Zero-allocation routing in hot paths
- Efficient string building and response writing
- Concurrent request handling

# Middleware

The router uses standard Go HTTP middleware patterns, making it compatible with the entire ecosystem:

	import (
		"github.com/joakimcarlsson/go-router/router"
		"github.com/joakimcarlsson/go-router/router/middleware/cors"
	)

	r := router.New()
	
	// Built-in CORS middleware
	r.Use(cors.Default())
	
	// Any standard HTTP middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Pre-processing
			next.ServeHTTP(w, req)
			// Post-processing
		})
	})

# Route Groups

Organize routes with common prefixes and middleware:

	r.Group("/api/v1", func(api *router.Router) {
		api.Use(authMiddleware)
		api.WithTags("Users")
		
		api.GET("/users", listUsers)
		api.POST("/users", createUser)
		
		api.Group("/admin", func(admin *router.Router) {
			admin.Use(adminMiddleware)
			admin.DELETE("/users/{id}", deleteUser)
		})
	})

# File Uploads

Built-in support for multipart form handling:

	type Upload struct {
		File        *multipart.FileHeader `form:"file" file:"true" required:"true"`
		Description string                `form:"description"`
	}

	r.POST("/upload", func(c *router.Context) {
		var upload Upload
		if err := c.BindForm(&upload); err != nil {
			c.JSON(400, map[string]string{"error": err.Error()})
			return
		}
		
		err := c.SaveUploadedFile(upload.File, "./uploads/" + upload.File.Filename)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
			return
		}
		
		c.JSON(200, map[string]string{"status": "uploaded"})
	})

# API Documentation

Add OpenAPI 3.0 documentation to your routes using the docs package:

	import "github.com/joakimcarlsson/go-router/docs"

	r.GET("/users/{id}", getUser,
		docs.WithSummary("Get user by ID"),
		docs.WithDescription("Retrieves a user by their unique identifier"),
		docs.WithTags("Users"),
		docs.WithPathParam("id", "string", true, "User ID", "123"),
		docs.WithJSONResponse[User](200, "User found"),
		docs.WithJSONResponse[ErrorResponse](404, "User not found"),
	)

Type-safe request/response documentation:

	r.POST("/users", createUser,
		docs.WithSummary("Create user"),
		docs.WithJSONRequestBody[CreateUserRequest](true, "User data"),
		docs.WithJSONResponse[User](201, "User created"),
		docs.WithJSONResponse[ValidationError](400, "Invalid input"),
	)

# Swagger UI Integration

Serve interactive API documentation:

	import "github.com/joakimcarlsson/go-router/integration"

	err := integration.Setup(r, integration.SetupOptions{
		Title:       "My API",
		Version:     "1.0.0",
		Description: "REST API for my application",
		SpecPath:    "/openapi.json",  // OpenAPI spec endpoint
		DocsPath:    "/docs",          // Swagger UI endpoint
	})
	
	// Now visit http://localhost:8080/docs for interactive documentation

# Custom Type Handlers

Register custom OpenAPI schemas for your types:

	import (
		"reflect"
		"github.com/joakimcarlsson/go-router/metadata"
	)

	type UserID string

	// Register custom schema
	metadata.RegisterTypeHandler("main.UserID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:    "string",
			Format:  "uuid",
			Example: "550e8400-e29b-41d4-a716-446655440000",
		}
	})

# Performance

The router is designed for high performance:
- Zero allocations in hot routing paths
- Object pooling for contexts and encoders
- Efficient path parameter extraction
- Minimal middleware overhead
- Support for thousands of routes

Benchmark comparisons with Go's standard library show excellent performance characteristics
while providing significantly more features.

# Server-Sent Events

Built-in SSE support for real-time applications:

	r.GET("/events", func(c *router.Context) {
		c.InitSSE()
		
		for i := 0; i < 10; i++ {
			c.SSE(router.SSEEvent{
				Event: "message",
				Data:  fmt.Sprintf("Event %d", i),
				ID:    fmt.Sprintf("msg-%d", i),
			})
			time.Sleep(time.Second)
		}
	})

# Architecture

The codebase is organized into focused packages:

Core Packages:
- router: HTTP routing, middleware, context handling
- router/middleware/cors: CORS middleware with extensive configuration

Documentation Packages:
- docs: Type-safe route documentation utilities
- openapi: OpenAPI 3.0 specification generation
- swagger: Swagger UI configuration and serving
- integration: High-level setup functions
- metadata: Shared types and utilities

This modular design allows you to include only what you need, keeping your binary size minimal
when you don't need documentation features.

# Compatibility

- Go 1.22+ (uses new routing patterns)
- Compatible with standard http.Handler middleware
- Works with existing Go HTTP ecosystem
- Can be used alongside other routers or HTTP frameworks
*/
package router
