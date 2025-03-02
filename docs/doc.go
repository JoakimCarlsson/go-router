/*
Package docs provides utilities for documenting API routes using OpenAPI-compatible options.
It offers a rich set of functions to describe routes, parameters, request bodies, and responses
using Go's type system and reflection.

Example usage:

	router.GET("/users/:id", handler,
		docs.WithSummary("Get user by ID"),
		docs.WithDescription("Retrieves a user's details by their unique identifier"),
		docs.WithTags("Users"),
		docs.WithPathParam("id", "string", true, "The user's unique identifier", nil),
		docs.WithResponse(200, "User found successfully"),
		docs.WithJSONResponse[User](200, "User details"),
	)

The package supports:
  - Route documentation with summaries and descriptions
  - Parameter documentation (path, query, header)
  - Request body documentation with schema inference
  - Response documentation with status codes and schemas
  - Security requirements (Basic Auth, Bearer Token, OAuth2)
  - Tags for grouping operations
  - Schema generation from Go types
*/
package docs
