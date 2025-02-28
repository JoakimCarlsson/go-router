# Go Router

A lightweight HTTP router for Go with built-in OpenAPI support.

## Installation

```bash
go get github.com/joakimcarlsson/go-router
```

## Features

- Lightweight HTTP routing
- Built-in OpenAPI 3.0 specification generation
- Route grouping with middleware support
- Type-safe request/response documentation
- Security scheme support

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/openapi"
)

func main() {
    r := router.New()
    
    r.GET("/hello", func(c *router.Context) {
        c.JSON(200, map[string]string{"message": "Hello, World!"})
    })

    http.ListenAndServe(":8080", r)
}
```

## OpenAPI Support

```go
generator := openapi.NewGenerator(openapi.Info{
    Title: "My API",
    Version: "1.0.0",
})

r.GET("/users", listUsers,
    openapi.WithSummary("List users"),
    openapi.WithResponseType("200", "Success", []User{}),
)

// Serve OpenAPI specification
r.GET("/openapi.json", r.ServeOpenAPI(generator))

// Serve Swagger UI documentation
swaggerConfig := router.DefaultSwaggerUIConfig()
swaggerConfig.Title = "API Documentation"
swaggerConfig.DarkMode = true  // Enable dark mode
r.GET("/docs", r.ServeSwaggerUI(swaggerConfig))
```

### Swagger UI Configuration Options

The router includes built-in Swagger UI support with many configuration options:

```go
config := router.DefaultSwaggerUIConfig()
config.Title = "My API Documentation"           // Page title
config.DarkMode = true                          // Enable dark mode
config.SpecURL = "/openapi.json"                // Path to OpenAPI spec
config.DocExpansion = "list"                    // "list", "full", or "none"
config.TryItOutEnabled = true                   // Enable "Try it out" by default
config.PersistAuthorization = true              // Save auth between page reloads
config.DefaultModelsExpandDepth = 2             // Expand nested models
config.RequestSnippetsEnabled = true            // Show code snippets for requests
config.CustomCSS = "/* Add your custom CSS */"  // Customize styling

r.GET("/docs", r.ServeSwaggerUI(config))
```

For more examples, see the [_examples](_examples) directory.

## License

MIT License - see LICENSE file