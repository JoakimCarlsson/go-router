# Go Router

[![Go Reference](https://pkg.go.dev/badge/github.com/JoakimCarlsson/go-router.svg)](https://pkg.go.dev/github.com/JoakimCarlsson/go-router)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue.svg)](https://golang.org/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/joakimcarlsson/go-router)](https://goreportcard.com/report/github.com/joakimcarlsson/go-router)

A high-performance, modular HTTP router for Go with built-in **OpenAPI 3.0** and **Swagger UI** support.

## Features

- **Auto Documentation**: Built-in OpenAPI 3.0 spec generation with type safety
- **Interactive UI**: Integrated Swagger UI for API exploration and testing
- **Modular Design**: Use only what you need - core routing or full documentation stack
- **Standard Compatible**: Works with any `http.Handler` middleware from the ecosystem
- **Modern Go**: Built for Go 1.22+ with new routing patterns and features
- **Type Safe**: Compile-time type safety for request/response documentation
- **Server-Sent Events**: Built-in SSE support for real-time applications
- **File Uploads**: Multipart form handling with validation
- **Content Negotiation**: Automatic JSON/XML response selection
- **Route Groups**: Organize routes with prefixes and shared middleware
- **Custom Types**: Register custom OpenAPI schemas for your types

## Overview

This router is designed with modularity in mind, allowing you to use only the components you need. The project is structured into several packages, each with a specific responsibility:

### Core Packages

- **router**: The core HTTP routing functionality
  - Path parameter support
  - Middleware support
  - Router groups
  - HTTP method helpers
  - Multipart form data handling
  - File upload support

- **middleware**: Built-in middleware components
  - CORS middleware with extensive configuration options
  - Support for custom headers, origins, and methods
  - Wildcard support for domain matching

- **metadata**: Shared type definitions
  - OpenAPI/Swagger shared types
  - OAuth2 configuration
  - Common utilities
  - Custom type handler registry

### Documentation Packages

- **docs**: API documentation utilities
  - Type-safe route documentation
  - Request/response schema generation
  - Parameter and security documentation
  - Validation tag support
  - Custom type schema generation

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

## Quick Start

### Basic HTTP Server

```go
package main

import (
    "github.com/joakimcarlsson/go-router/router"
    "log"
)

func main() {
    r := router.New()
    
    // Simple routes
    r.GET("/", func(c *router.Context) {
        c.JSON(200, map[string]string{"message": "Hello, World!"})
    })
    
    r.GET("/users/{id}", func(c *router.Context) {
        userID := c.Param("id")
        c.JSON(200, map[string]string{"user_id": userID})
    })
    
    // Start server
    log.Fatal(r.Run(":8080"))
}
```

### With Auto-Generated API Documentation

```go
package main

import (
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/docs"
    "github.com/joakimcarlsson/go-router/integration"
    "log"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

func main() {
    r := router.New()
    
    // Setup auto-documentation
    err := integration.Setup(r, integration.SetupOptions{
        Title:       "My API",
        Version:     "1.0.0",
        Description: "A sample API with auto-generated documentation",
        SpecPath:    "/openapi.json",
        DocsPath:    "/docs",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Documented route with type safety
    r.GET("/users/{id}", getUser,
        docs.WithSummary("Get user by ID"),
        docs.WithDescription("Retrieves a user by their unique identifier"),
        docs.WithTags("Users"),
        docs.WithPathParam("id", "string", true, "User ID", "123"),
        docs.WithJSONResponse[User](200, "User found"),
        docs.WithJSONResponse[ErrorResponse](404, "User not found"),
    )
    
    log.Printf("Server starting on :8080")
    log.Printf("API docs available at: http://localhost:8080/docs")
    log.Fatal(r.Run(":8080"))
}

func getUser(c *router.Context) {
    id := c.Param("id")
    user := User{ID: id, Name: "John Doe", Email: "john@example.com"}
    c.JSON(200, user)
}
```

## Comprehensive Examples

### RESTful API with Full Documentation

```go
package main

import (
    "strconv"
    "time"
    
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/docs"
    "github.com/joakimcarlsson/go-router/integration"
)

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
    Title       string `json:"title" validate:"required"`
    Description string `json:"description"`
}

type UpdateTaskRequest struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
    Completed   *bool   `json:"completed,omitempty"`
}

func main() {
    r := router.New()
    
    // Setup documentation
    integration.Setup(r, integration.SetupOptions{
        Title:       "Task Manager API",
        Version:     "1.0.0",
        Description: "A RESTful API for managing tasks",
    })
    
    // API routes with documentation
    api := r.Group("/api/v1", func(api *router.Router) {
        api.WithTags("Tasks")
        
        // List tasks
        api.GET("/tasks", listTasks,
            docs.WithSummary("List all tasks"),
            docs.WithQueryParam("completed", "boolean", false, "Filter by completion status", nil),
            docs.WithQueryParam("limit", "integer", false, "Number of tasks to return", 10),
            docs.WithJSONResponse[[]Task](200, "List of tasks"),
        )
        
        // Create task
        api.POST("/tasks", createTask,
            docs.WithSummary("Create a new task"),
            docs.WithJSONRequestBody[CreateTaskRequest](true, "Task data"),
            docs.WithJSONResponse[Task](201, "Task created"),
            docs.WithJSONResponse[ErrorResponse](400, "Invalid input"),
        )
        
        // Get task
        api.GET("/tasks/{id}", getTask,
            docs.WithSummary("Get task by ID"),
            docs.WithPathParam("id", "integer", true, "Task ID", 1),
            docs.WithJSONResponse[Task](200, "Task found"),
            docs.WithJSONResponse[ErrorResponse](404, "Task not found"),
        )
        
        // Update task
        api.PUT("/tasks/{id}", updateTask,
            docs.WithSummary("Update a task"),
            docs.WithPathParam("id", "integer", true, "Task ID", 1),
            docs.WithJSONRequestBody[UpdateTaskRequest](true, "Updated task data"),
            docs.WithJSONResponse[Task](200, "Task updated"),
            docs.WithJSONResponse[ErrorResponse](404, "Task not found"),
        )
        
        // Delete task
        api.DELETE("/tasks/{id}", deleteTask,
            docs.WithSummary("Delete a task"),
            docs.WithPathParam("id", "integer", true, "Task ID", 1),
            docs.WithResponse(204, "Task deleted"),
            docs.WithJSONResponse[ErrorResponse](404, "Task not found"),
        )
    })
    
    r.Run(":8080")
}

// Handler implementations
func listTasks(c *router.Context) {
    completed := c.QueryBoolDefault("completed", false)
    limit := c.QueryIntDefault("limit", 10)
    
    // Implementation here
    c.JSON(200, []Task{})
}

func createTask(c *router.Context) {
    var req CreateTaskRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: "Invalid JSON"})
        return
    }
    
    // Implementation here
    task := Task{
        ID:          1,
        Title:       req.Title,
        Description: req.Description,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    c.JSON(201, task)
}

func getTask(c *router.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(400, ErrorResponse{Error: "Invalid task ID"})
        return
    }
    
    // Implementation here
    c.JSON(200, Task{ID: id})
}

func updateTask(c *router.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(400, ErrorResponse{Error: "Invalid task ID"})
        return
    }
    
    var req UpdateTaskRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: "Invalid JSON"})
        return
    }
    
    // Implementation here
    c.JSON(200, Task{ID: id})
}

func deleteTask(c *router.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(400, ErrorResponse{Error: "Invalid task ID"})
        return
    }
    
    // Implementation here
    c.Status(204)
}
```

## CORS Middleware

Configure Cross-Origin Resource Sharing (CORS) with the built-in middleware:

```go
import (
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/router/middleware/cors"
)

func main() {
    r := router.New()
    
    // Use default CORS settings (allow all origins)
    r.Use(cors.Default())
    
    // Or use custom CORS configuration with the simple API
    r.Use(cors.Handler(cors.Options{
        AllowOrigins:     []string{"https://example.com", "https://*.trusted-domain.com"},
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           86400, // Cache preflight response for 24 hours
    }))
    
    // Different CORS settings for specific route groups
    r.Group("/api", func(api *router.Router) {
        api.Use(cors.Handler(cors.Options{
            AllowOrigins: []string{"https://api.example.com"},
            // Other options...
        }))
        
        // API routes...
    })
}
```

## Standard Middleware Compatibility

The router exclusively uses standard HTTP middleware, making it compatible with the vast ecosystem of existing Go middleware:

```go
import (
    "github.com/joakimcarlsson/go-router/router"
    "github.com/justinas/nosurf"  // Example of a standard middleware package
)

func main() {
    r := router.New()
    
    // Use any standard HTTP middleware
    r.Use(loggingMiddleware, cors.Default())
    
    // Standard middleware is any function with signature:
    // func(http.Handler) http.Handler
    
    // Convert a standard http.Handler to a router.HandlerFunc
    fileServer := http.FileServer(http.Dir("./static"))
    r.GET("/static/*filepath", router.FromHTTPHandler(fileServer))
    
    // Convert a router.HandlerFunc to a standard http.HandlerFunc
    customHandler := func(c *router.Context) {
        c.JSON(200, map[string]string{"message": "Hello"})
    }
    
    // Use with standard http package
    http.Handle("/api/hello", router.ToHTTPHandlerFunc(customHandler))
}
```

## File Uploads

Handle file uploads with built-in multipart form support:

```go
// Define your upload struct with form tags
type FileUpload struct {
    File        *multipart.FileHeader `form:"file" file:"true" required:"true" description:"The file to upload"`
    Name        string                `form:"name" description:"Optional name for the file"`
    Description string                `form:"description" description:"Description of the file"`
}

// Handle single file upload
r.POST("/upload", func(c *router.Context) {
    var upload FileUpload
    if err := c.BindForm(&upload); err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }

    // Save the file
    dst := filepath.Join("uploads", upload.File.Filename)
    if err := c.SaveUploadedFile(upload.File, dst); err != nil {
        c.JSON(500, map[string]string{"error": err.Error()})
        return
    }

    c.JSON(201, map[string]string{
        "message": "File uploaded successfully",
        "name": upload.Name,
        "path": dst,
    })
})

// Configure upload size limit
r.WithMultipartConfig(32 << 20) // 32 MB
```

## Documentation Support

Add OpenAPI documentation to your routes:

```go
import "github.com/joakimcarlsson/go-router/docs"

// Document a JSON endpoint
r.GET("/users/{id}", getUser,
    docs.WithSummary("Get user by ID"),
    docs.WithPathParam("id", "string", true, "User ID", nil),
    docs.WithJSONResponse[User](200, "User found"),
)

// Document a file upload endpoint
r.POST("/upload", uploadHandler,
    docs.WithSummary("Upload a file"),
    docs.WithMultipartFormStruct[FileUpload]("File upload with metadata"),
    docs.WithJSONResponse[UploadResponse](201, "File uploaded successfully"),
)
```

## Custom Type Handlers

Register custom OpenAPI schema handlers for your own types:

```go
import (
    "reflect"
    "github.com/joakimcarlsson/go-router/metadata"
)

// Define a custom type
type EmailAddress string

// Register a type handler
metadata.RegisterTypeHandler("mypackage.EmailAddress", func(t reflect.Type) metadata.Schema {
    return metadata.Schema{
        Type:        "string",
        Format:      "email",
        Example:     "user@example.com",
        Description: "Email address in standard format",
    }
})

// Use it in your models
type User struct {
    Email EmailAddress `json:"email"`
    // Other fields...
}
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

## Advanced Features

### File Uploads

```go
type FileUpload struct {
    File        *multipart.FileHeader `form:"file" file:"true" required:"true"`
    Name        string                `form:"name"`
    Description string                `form:"description"`
}

r.POST("/upload", func(c *router.Context) {
    var upload FileUpload
    if err := c.BindForm(&upload); err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }

    // Save the file
    dst := filepath.Join("uploads", upload.File.Filename)
    if err := c.SaveUploadedFile(upload.File, dst); err != nil {
        c.JSON(500, map[string]string{"error": err.Error()})
        return
    }

    c.JSON(201, map[string]string{"message": "File uploaded successfully"})
},
docs.WithSummary("Upload file"),
docs.WithMultipartFormStruct[FileUpload]("File upload with metadata"),
docs.WithJSONResponse[map[string]string](201, "Upload successful"),
)
```

### Server-Sent Events

```go
r.GET("/events", func(c *router.Context) {
    c.InitSSE()
    
    // Send events
    for i := 0; i < 10; i++ {
        err := c.SSE(router.SSEEvent{
            Event: "message",
            Data:  fmt.Sprintf("Event %d", i),
            ID:    fmt.Sprintf("msg-%d", i),
        })
        if err != nil {
            break
        }
        time.Sleep(time.Second)
    }
})
```

### Custom Type Handlers

```go
import (
    "reflect"
    "github.com/joakimcarlsson/go-router/metadata"
)

type UserID string

// Register custom schema for OpenAPI documentation
metadata.RegisterTypeHandler("main.UserID", func(t reflect.Type) metadata.Schema {
    return metadata.Schema{
        Type:    "string",
        Format:  "uuid",
        Example: "550e8400-e29b-41d4-a716-446655440000",
    }
})

type User struct {
    ID   UserID `json:"id"`
    Name string `json:"name"`
}
```

### Authentication & Security

```go
// Setup OAuth2 security scheme
generator := openapi.NewGenerator(openapi.Info{
    Title:   "Secure API",
    Version: "1.0.0",
})

generator.WithOAuth2ImplicitFlow("oauth2", "OAuth2 authentication",
    "https://auth.example.com/oauth/authorize",
    map[string]string{
        "read":  "Read access",
        "write": "Write access",
    },
)

// Add security to routes
r.GET("/protected", protectedHandler,
    docs.WithSummary("Protected resource"),
    docs.WithOAuth2Scopes("read"),
    docs.WithBearerAuth(),
)
```

## Performance

Go Router is designed for high performance:

- **Zero allocations** in hot routing paths
- **Object pooling** for contexts and encoders
- **Efficient path matching** using Go 1.22+ patterns
- **Minimal middleware overhead**
- **Concurrent request handling**

### Benchmarks

Run benchmarks to see performance characteristics:

```bash
# Core router benchmarks
cd benchmarks && go test -bench=BenchmarkRouter_ -benchmem

# Comparison with standard library
cd benchmarks && go test -bench=BenchmarkComparison_ -benchmem

# Memory allocation tests
cd benchmarks && go test -bench=BenchmarkRouter_MemoryAllocation -benchmem
```

Typical results show excellent performance compared to standard library while providing significantly more features.

## Migration Guides

### From Gin

```go
// Gin
gin.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

// Go Router
r.GET("/users/{id}", func(c *router.Context) {
    id := c.Param("id")
    c.JSON(200, map[string]string{"id": id})
})
```

### From Echo

```go
// Echo
e.GET("/users/:id", func(c echo.Context) error {
    id := c.Param("id")
    return c.JSON(200, map[string]string{"id": id})
})

// Go Router
r.GET("/users/{id}", func(c *router.Context) {
    id := c.Param("id")
    c.JSON(200, map[string]string{"id": id})
})
```

### From Chi

```go
// Chi
r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    json.NewEncoder(w).Encode(map[string]string{"id": id})
})

// Go Router
r.GET("/users/{id}", func(c *router.Context) {
    id := c.Param("id")
    c.JSON(200, map[string]string{"id": id})
})
```

## API Reference

### Router Methods

- `New()` - Create new router
- `GET/POST/PUT/DELETE/PATCH(path, handler, ...options)` - Register routes
- `Group(prefix, func)` - Create route groups
- `Use(middleware...)` - Add middleware
- `Run(addr)` - Start HTTP server
- `ServeHTTP(w, r)` - Implement http.Handler

### Context Methods

- `Param(key)` - Get path parameter
- `Query()` - Get query parameters
- `JSON/XML/String(code, obj)` - Send responses
- `BindJSON/BindXML/BindForm(obj)` - Parse request body
- `Status(code)` - Set status code
- `SetHeader/GetHeader(key, value)` - Manage headers
- `File(path)` - Serve files
- `Redirect(code, url)` - HTTP redirects

### Documentation Options

- `docs.WithSummary/WithDescription` - Basic documentation
- `docs.WithTags` - Group operations
- `docs.WithPathParam/WithQueryParam` - Document parameters
- `docs.WithJSONRequest/Response[T]` - Type-safe schemas
- `docs.WithSecurity/WithAuth` - Authentication requirements

## Examples Directory

Explore the `_examples` directory for complete, runnable examples:

- **minimal-api** - Basic routing without documentation
- **basic-api-with-docs** - Full OpenAPI integration
- **file-upload** - Multipart form handling
- **cors** - Cross-origin resource sharing
- **oauth2-*** - Various OAuth2 flows
- **server-sent-events** - Real-time event streaming
- **custom-type-handlers** - Custom OpenAPI schemas

## Design Goals

1. **Modularity**: Use only the components you need
2. **Type Safety**: Leverage Go's type system for documentation
5. **Developer Experience**: Intuitive APIs with comprehensive examples
6. **Extensibility**: Easy to add custom functionality

## Contributing

We welcome contributions! Please see:

- **Issues**: Report bugs or request features
- **Pull Requests**: Submit improvements
- **Documentation**: Help improve examples and guides
- **Testing**: Add test cases or benchmarks

### Development

```bash
# Run tests
go test ./...

# Run benchmarks
cd benchmarks && go test -bench=. -benchmem

# Run linter
golangci-lint run

# Check all examples
find _examples -name "*.go" -exec go run {} \;
```

## License

MIT License - see [LICENSE](LICENSE) file for details

## Support

- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/JoakimCarlsson/go-router)
- **Examples**: See `_examples/` directory
- **Issues**: [GitHub Issues](https://github.com/JoakimCarlsson/go-router/issues)
