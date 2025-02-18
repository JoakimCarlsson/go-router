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
```

For more examples, see the [_examples](_examples) directory.

## License

MIT License - see LICENSE file