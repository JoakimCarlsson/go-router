/*
Package openapi provides OpenAPI 3.0 specification generation and route documentation
for the go-router package.

# Overview

This package provides:
  - Type-safe route documentation options
  - Automatic schema generation from Go types
  - OpenAPI 3.0 specification generation
  - Custom type handler registration
  - SSE (Server-Sent Events) documentation

# Documenting Routes

Add documentation to routes using the With* functions:

	r.GET("/users/{id}", getUser,
		openapi.WithSummary("Get user by ID"),
		openapi.WithDescription("Retrieves a user by their unique identifier"),
		openapi.WithTags("Users"),
		openapi.WithPathParam("id", "string", true, "User ID", "123"),
		openapi.WithJSONResponse[User](200, "User found"),
		openapi.WithJSONResponse[ErrorResponse](404, "User not found"),
	)

	r.POST("/users", createUser,
		openapi.WithSummary("Create user"),
		openapi.WithTags("Users"),
		openapi.WithJSONRequestBody[CreateUserRequest](true, "User data"),
		openapi.WithJSONResponse[User](201, "User created"),
		openapi.WithJSONResponse[ErrorResponse](400, "Invalid input"),
	)

# Schema Generation

Schemas are automatically generated from Go types:

	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name" validate:"required"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	}

The validate:"required" tag marks fields as required in the OpenAPI schema.

# OpenAPI Generator

Create a generator and configure it:

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "My API",
		Version:     "1.0.0",
		Description: "API description",
	})

	// Add security schemes
	generator.WithBearerAuth("bearerAuth", "JWT Bearer token")
	generator.WithAPIKey("apiKey", "API Key", "header", "X-API-Key")

	// Add servers
	generator.WithServer("https://api.example.com", "Production")
	generator.WithServer("https://staging.example.com", "Staging")

# Custom Type Handlers

Register custom schema handlers for your types:

	openapi.RegisterTypeHandler("main.EmailAddress", func(t reflect.Type) openapi.Schema {
		return openapi.Schema{
			Type:    "string",
			Format:  "email",
			Example: "user@example.com",
		}
	})

	openapi.RegisterTypeHandler("main.Currency", func(t reflect.Type) openapi.Schema {
		return openapi.Schema{
			Type:    "string",
			Pattern: "^[A-Z]{3}$",
			Example: "USD",
		}
	})

# SSE Documentation

Document Server-Sent Events endpoints:

	r.GET("/events", eventHandler,
		openapi.WithSummary("Event stream"),
		openapi.WithSSEResponse("Real-time event stream"),
		openapi.WithSSEEvent[MessageEvent]("message", "New message received"),
		openapi.WithSSEEvent[StatusEvent]("status", "Status update"),
	)

# Security

Add authentication requirements to routes:

	// Bearer token authentication
	r.GET("/protected", handler,
		openapi.WithBearerAuth(),
	)

	// API key authentication
	r.GET("/api/data", handler,
		openapi.WithAPIKey(),
	)

	// OAuth2 with scopes
	r.GET("/user/profile", handler,
		openapi.WithOAuth2Scopes("read:profile", "read:email"),
	)

# Parameter Types

Various parameter documentation options:

	// Path parameters
	openapi.WithPathParam("id", "string", true, "Resource ID", "123")
	openapi.WithUUIDPathParam("id", true, "Resource UUID", "...")
	openapi.WithIntegerPathParam("id", true, "Numeric ID", 42, nil, nil)

	// Query parameters
	openapi.WithQueryParam("page", "integer", false, "Page number", 1)
	openapi.WithEnumQueryParam("status", false, "Status filter", "active",
		[]interface{}{"active", "inactive", "pending"})

	// Header parameters
	openapi.WithHeaderParam("X-Request-ID", false, "Request tracking ID", nil)

# Excluding Routes

Exclude routes from documentation:

	r.GET("/health", healthCheck, openapi.ExcludeFromDocs())
	r.GET("/metrics", metrics, openapi.ExcludeFromDocs())

For Swagger UI integration, see the swaggerui package.

For more information: https://github.com/JoakimCarlsson/go-router
*/
package openapi
