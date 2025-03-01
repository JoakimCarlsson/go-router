# Go Router

A modular HTTP router for Go with built-in OpenAPI and Swagger UI support.

## Overview

This router is designed with modularity in mind, allowing you to use only the components you need. The project is structured into several packages, each with a specific responsibility:

### Core Packages

- **router**: The core HTTP routing functionality
  - Path parameter support
  - Middleware support
  - Router groups
  - HTTP method helpers

- **metadata**: Shared type definitions
  - OpenAPI/Swagger shared types
  - OAuth2 configuration
  - Common utilities

### Documentation Packages

- **docs**: API documentation utilities
  - Type-safe route documentation
  - Request/response schema generation
  - Parameter and security documentation
  - Validation tag support

- **openapi**: OpenAPI specification generation
  - OpenAPI 3.0 support
  - Schema generation from Go types
  - Security scheme configuration
  - Server and info configuration

- **swagger**: Swagger UI configuration and serving
  - Customizable UI
  - Dark mode support
  - OAuth2 configuration
  - Custom CSS/JS support

### Integration

- **integration**: Component integration
  - OpenAPI adapter
  - Swagger UI integration
  - Clean separation of concerns

## Installation

```bash
go get github.com/joakimcarlsson/go-router
```

## Basic Usage

```go
package main

import (
    "github.com/joakimcarlsson/go-router/router"
)

func main() {
    r := router.New()
    
    r.GET("/hello", func(c *router.Context) {
        c.String(200, "Hello, World!")
    })
    
    r.Run(":8080")
}
```

## Documentation Support

Add OpenAPI documentation to your routes:

```go
import "github.com/joakimcarlsson/go-router/docs"

r.GET("/users/:id", getUser,
    docs.WithSummary("Get user by ID"),
    docs.WithPathParam("id", "string", true, "User ID", nil),
    docs.WithJSONResponse[User](200, "User found"),
)
```

## Swagger UI Integration

Add interactive API documentation:

```go
import (
    "github.com/joakimcarlsson/go-router/integration"
    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/swagger"
)

// Create OpenAPI generator
generator := openapi.NewGenerator(openapi.Info{
    Title:   "My API",
    Version: "1.0.0",
})

// Configure Swagger UI
swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")
```

## Examples

See the `_examples` directory for complete examples:

- Basic routing and middleware
- OpenAPI documentation
- OAuth2 authentication
- Swagger UI integration
- Complete refactored example

## Design Goals

1. **Modularity**: Use only what you need
2. **Type Safety**: Leveraging Go's type system
3. **Clean API**: Intuitive and consistent interfaces
4. **Extensibility**: Easy to add new features
5. **Documentation**: First-class OpenAPI support

## Contributing

Contributions are welcome! Please read our contributing guidelines and code of conduct.

## License

MIT License - see LICENSE file for details