// Package router provides a lightweight HTTP router with OpenAPI support.
//
// go-router is designed to be modular and extensible, allowing users to select
// only the components they need. The router supports standard HTTP routing features
// like path parameters, middleware, route groups, and more.
//
// Core Features:
//   - HTTP method-based routing with path parameters
//   - Middleware support
//   - Route grouping
//   - OpenAPI 3.0 specification generation
//   - Swagger UI integration
//
// Example usage:
//
//	r := router.New()
//	r.GET("/users", handleUsers)
//	r.POST("/users", createUser)
//	http.ListenAndServe(":8080", r)
//
// For OpenAPI documentation, you can use the docs package:
//
//	r.GET("/users/:id", getUser,
//	    docs.WithSummary("Get user by ID"),
//	    docs.WithDescription("Retrieves a user by their ID"),
//	    docs.WithPathParam("id", "string", true, "User ID", nil),
//	    docs.WithJSONResponse[User](200, "User found")
//	)
//
// To serve OpenAPI documentation with Swagger UI:
//
//	integration.Setup(r, integration.DefaultSetupOptions())
package router
