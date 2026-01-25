# Go Router

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue.svg)](https://golang.org/dl/)

A high-performance, modular HTTP router for Go with built-in **OpenAPI 3.0** and **Swagger UI** support.

| Module | Reference | Report Card |
|--------|-----------|-------------|
| router | [![Go Reference](https://pkg.go.dev/badge/github.com/joakimcarlsson/go-router/router.svg)](https://pkg.go.dev/github.com/joakimcarlsson/go-router/router) | [![Go Report Card](https://goreportcard.com/badge/github.com/joakimcarlsson/go-router/router)](https://goreportcard.com/report/github.com/joakimcarlsson/go-router/router) |
| openapi | [![Go Reference](https://pkg.go.dev/badge/github.com/joakimcarlsson/go-router/openapi.svg)](https://pkg.go.dev/github.com/joakimcarlsson/go-router/openapi) | [![Go Report Card](https://goreportcard.com/badge/github.com/joakimcarlsson/go-router/openapi)](https://goreportcard.com/report/github.com/joakimcarlsson/go-router/openapi) |
| swaggerui | [![Go Reference](https://pkg.go.dev/badge/github.com/joakimcarlsson/go-router/swaggerui.svg)](https://pkg.go.dev/github.com/joakimcarlsson/go-router/swaggerui) | [![Go Report Card](https://goreportcard.com/badge/github.com/joakimcarlsson/go-router/swaggerui)](https://goreportcard.com/report/github.com/joakimcarlsson/go-router/swaggerui) |

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

## Installation

### Router Only (no documentation)

For basic HTTP routing without OpenAPI documentation:

```bash
go get github.com/joakimcarlsson/go-router/router@latest
```

### Full Stack (with Swagger UI)

For routing with auto-generated API documentation and Swagger UI:

```bash
go get github.com/joakimcarlsson/go-router/router@latest
go get github.com/joakimcarlsson/go-router/openapi@latest
go get github.com/joakimcarlsson/go-router/swaggerui@latest
```

### OpenAPI Only (no UI)

For routing with OpenAPI spec generation but serving your own UI:

```bash
go get github.com/joakimcarlsson/go-router/router@latest
go get github.com/joakimcarlsson/go-router/openapi@latest
```

## Modules

This router is split into three independent modules, each with its own versioning:

### router

Core HTTP routing functionality.

- Path parameter support
- Middleware support
- Router groups
- HTTP method helpers
- Multipart form data handling
- File upload support
- Server-Sent Events
- Built-in CORS middleware

### openapi

OpenAPI 3.0 specification generation and route documentation.

- Type-safe route documentation
- Request/response schema generation
- Parameter and security documentation
- Validation tag support
- Custom type schema generation
- SSE event documentation

### swaggerui

Swagger UI serving and integration.

- Customizable Swagger UI
- Dark mode support
- OAuth2 configuration
- Custom CSS/JS support
- Easy setup with router and openapi

## Quick Start

### Basic HTTP Server (router only)

```go
package main

import (
    "log"
    "net/http"

    "github.com/joakimcarlsson/go-router/router"
)

func main() {
    r := router.New()
    
    r.GET("/", func(c *router.Context) {
        c.JSON(200, map[string]string{"message": "Hello, World!"})
    })
    
    r.GET("/users/{id}", func(c *router.Context) {
        userID := c.Param("id")
        c.JSON(200, map[string]string{"user_id": userID})
    })
    
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

### With Auto-Generated API Documentation

```go
package main

import (
    "log"
    "net/http"

    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/swaggerui"
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
    
    // Create OpenAPI generator
    generator := openapi.NewGenerator(openapi.Info{
        Title:       "My API",
        Version:     "1.0.0",
        Description: "A sample API with auto-generated documentation",
    })
    
    // Documented route with type safety
    r.GET("/users/{id}", getUser,
        openapi.WithSummary("Get user by ID"),
        openapi.WithDescription("Retrieves a user by their unique identifier"),
        openapi.WithTags("Users"),
        openapi.WithPathParam("id", "string", true, "User ID", "123"),
        openapi.WithJSONResponse[User](200, "User found"),
        openapi.WithJSONResponse[ErrorResponse](404, "User not found"),
    )
    
    // Setup Swagger UI
    setup := swaggerui.NewSetup(r, generator)
    setup.RegisterRoutes(r, "/openapi.json", "/docs")
    
    log.Printf("Server starting on :8080")
    log.Printf("API docs available at: http://localhost:8080/docs")
    log.Fatal(http.ListenAndServe(":8080", r))
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
    "net/http"
    "strconv"
    "time"

    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/swaggerui"
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

type ErrorResponse struct {
    Error string `json:"error"`
}

func main() {
    r := router.New()
    
    // Create OpenAPI generator
    generator := openapi.NewGenerator(openapi.Info{
        Title:       "Task Manager API",
        Version:     "1.0.0",
        Description: "A RESTful API for managing tasks",
    })
    
    // API routes with documentation
    r.Group("/api/v1", func(api *router.Router) {
        api.WithOptions(openapi.WithTags("Tasks"))
        
        // List tasks
        api.GET("/tasks", listTasks,
            openapi.WithSummary("List all tasks"),
            openapi.WithQueryParam("completed", "boolean", false, "Filter by completion status", nil),
            openapi.WithQueryParam("limit", "integer", false, "Number of tasks to return", 10),
            openapi.WithJSONResponse[[]Task](200, "List of tasks"),
        )
        
        // Create task
        api.POST("/tasks", createTask,
            openapi.WithSummary("Create a new task"),
            openapi.WithJSONRequestBody[CreateTaskRequest](true, "Task data"),
            openapi.WithJSONResponse[Task](201, "Task created"),
            openapi.WithJSONResponse[ErrorResponse](400, "Invalid input"),
        )
        
        // Get task
        api.GET("/tasks/{id}", getTask,
            openapi.WithSummary("Get task by ID"),
            openapi.WithPathParam("id", "integer", true, "Task ID", 1),
            openapi.WithJSONResponse[Task](200, "Task found"),
            openapi.WithJSONResponse[ErrorResponse](404, "Task not found"),
        )
        
        // Update task
        api.PUT("/tasks/{id}", updateTask,
            openapi.WithSummary("Update a task"),
            openapi.WithPathParam("id", "integer", true, "Task ID", 1),
            openapi.WithJSONRequestBody[UpdateTaskRequest](true, "Updated task data"),
            openapi.WithJSONResponse[Task](200, "Task updated"),
            openapi.WithJSONResponse[ErrorResponse](404, "Task not found"),
        )
        
        // Delete task
        api.DELETE("/tasks/{id}", deleteTask,
            openapi.WithSummary("Delete a task"),
            openapi.WithPathParam("id", "integer", true, "Task ID", 1),
            openapi.WithResponse(204, "Task deleted"),
            openapi.WithJSONResponse[ErrorResponse](404, "Task not found"),
        )
    })
    
    // Setup Swagger UI
    setup := swaggerui.NewSetup(r, generator)
    setup.RegisterRoutes(r, "/openapi.json", "/docs")
    
    http.ListenAndServe(":8080", r)
}

func listTasks(c *router.Context) {
    c.JSON(200, []Task{})
}

func createTask(c *router.Context) {
    var req CreateTaskRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: "Invalid JSON"})
        return
    }
    
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
    
    c.JSON(200, Task{ID: id})
}

func deleteTask(c *router.Context) {
    c.Status(204)
}
```

## CORS Middleware

Configure Cross-Origin Resource Sharing (CORS) with the built-in middleware:

```go
import (
    "net/http"

    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/router/middleware/cors"
)

func main() {
    r := router.New()
    
    // Use default CORS settings (allow all origins)
    r.Use(cors.Default())
    
    // Or use custom CORS configuration
    r.Use(cors.Handler(cors.Options{
        AllowOrigins:     []string{"https://example.com", "https://*.trusted-domain.com"},
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           86400,
    }))
}
```

## Standard Middleware Compatibility

The router uses standard HTTP middleware, making it compatible with the ecosystem:

```go
import (
    "net/http"

    "github.com/joakimcarlsson/go-router/router"
)

func main() {
    r := router.New()
    
    // Use any standard HTTP middleware
    r.Use(loggingMiddleware, cors.Default())
    
    // Convert a standard http.Handler to a router.HandlerFunc
    fileServer := http.FileServer(http.Dir("./static"))
    r.GET("/static/*filepath", router.FromHTTPHandler(fileServer))
    
    // Convert a router.HandlerFunc to a standard http.HandlerFunc
    customHandler := func(c *router.Context) {
        c.JSON(200, map[string]string{"message": "Hello"})
    }
    http.Handle("/api/hello", router.ToHTTPHandlerFunc(customHandler))
}
```

## File Uploads

Handle file uploads with built-in multipart form support:

```go
import (
    "mime/multipart"
    "path/filepath"

    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/router"
)

type FileUpload struct {
    File        *multipart.FileHeader `form:"file" file:"true" required:"true" description:"The file to upload"`
    Name        string                `form:"name" description:"Optional name for the file"`
    Description string                `form:"description" description:"Description of the file"`
}

func main() {
    r := router.New()
    
    r.POST("/upload", func(c *router.Context) {
        var upload FileUpload
        if err := c.BindForm(&upload); err != nil {
            c.JSON(400, map[string]string{"error": err.Error()})
            return
        }

        dst := filepath.Join("uploads", upload.File.Filename)
        if err := c.SaveUploadedFile(upload.File, dst); err != nil {
            c.JSON(500, map[string]string{"error": err.Error()})
            return
        }

        c.JSON(201, map[string]string{
            "message": "File uploaded successfully",
            "path":    dst,
        })
    },
        openapi.WithSummary("Upload file"),
        openapi.WithMultipartFormStruct[FileUpload]("File upload with metadata"),
    )
    
    // Configure upload size limit
    r.WithMultipartConfig(32 << 20) // 32 MB
}
```

## Server-Sent Events

```go
import (
    "fmt"
    "time"

    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/router"
)

type EventData struct {
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}

func main() {
    r := router.New()
    
    r.GET("/events", func(c *router.Context) {
        c.InitSSE()
        
        for i := 0; i < 10; i++ {
            err := c.SSEJson("message", EventData{
                Message:   fmt.Sprintf("Event %d", i),
                Timestamp: time.Now(),
            }, fmt.Sprintf("msg-%d", i))
            if err != nil {
                break
            }
            time.Sleep(time.Second)
        }
    },
        openapi.WithSummary("Event stream"),
        openapi.WithSSEResponse("Real-time event stream"),
        openapi.WithSSEEvent[EventData]("message", "Periodic message event"),
    )
}
```

## Custom Type Handlers

Register custom OpenAPI schema handlers for your types:

```go
import (
    "reflect"

    "github.com/joakimcarlsson/go-router/openapi"
)

type EmailAddress string

func init() {
    openapi.RegisterTypeHandler("main.EmailAddress", func(t reflect.Type) openapi.Schema {
        return openapi.Schema{
            Type:        "string",
            Format:      "email",
            Example:     "user@example.com",
            Description: "Email address in standard format",
        }
    })
}

type User struct {
    Email EmailAddress `json:"email"`
}
```

## Swagger UI Configuration

```go
import (
    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/swaggerui"
)

func main() {
    r := router.New()
    
    generator := openapi.NewGenerator(openapi.Info{
        Title:   "My API",
        Version: "1.0.0",
    })
    
    // Configure Swagger UI
    config := swaggerui.DefaultUIConfig()
    config.Title = "My API Documentation"
    config.DarkMode = true
    config.TryItOutEnabled = true
    
    setup := swaggerui.NewSetup(r, generator)
    setup.WithUIConfig(config)
    setup.RegisterRoutes(r, "/openapi.json", "/docs")
}
```

## Authentication & Security

```go
import (
    "github.com/joakimcarlsson/go-router/openapi"
    "github.com/joakimcarlsson/go-router/router"
    "github.com/joakimcarlsson/go-router/swaggerui"
)

func main() {
    r := router.New()
    
    generator := openapi.NewGenerator(openapi.Info{
        Title:   "Secure API",
        Version: "1.0.0",
    })
    
    // Add OAuth2 security scheme
    generator.WithOAuth2ImplicitFlow("oauth2", "OAuth2 authentication",
        "https://auth.example.com/oauth/authorize",
        map[string]string{
            "read":  "Read access",
            "write": "Write access",
        },
    )
    
    // Add Bearer auth security scheme
    generator.WithBearerAuth("bearerAuth", "JWT Bearer token")
    
    // Protected route
    r.GET("/protected", protectedHandler,
        openapi.WithSummary("Protected resource"),
        openapi.WithBearerAuth(),
    )
    
    setup := swaggerui.NewSetup(r, generator)
    setup.RegisterRoutes(r, "/openapi.json", "/docs")
}
```

## API Reference

### Router Methods

- `New()` - Create new router
- `GET/POST/PUT/DELETE/PATCH(path, handler, ...options)` - Register routes
- `Group(prefix, func)` - Create route groups
- `Use(middleware...)` - Add middleware
- `WithOptions(options...)` - Add options to route groups
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
- `InitSSE()` - Initialize SSE stream
- `SSEJson(event, data, id)` - Send SSE event

### Documentation Options (openapi package)

- `WithSummary/WithDescription` - Basic documentation
- `WithTags` - Group operations
- `WithPathParam/WithQueryParam` - Document parameters
- `WithJSONRequestBody[T]` - Type-safe request body
- `WithJSONResponse[T]` - Type-safe response
- `WithSecurity/WithBearerAuth` - Authentication requirements
- `WithSSEResponse/WithSSEEvent[T]` - SSE documentation

## Examples Directory

Explore the `_examples` directory for complete, runnable examples:

- **minimal-api** - Basic routing without documentation
- **basic-api-with-docs** - Full OpenAPI and Swagger UI integration
- **file-upload** - Multipart form handling with documentation
- **parameter-formats** - Path, query, and header parameter examples
- **cors-middleware** - Cross-origin resource sharing configuration
- **custom-middleware** - Building your own middleware (logging, auth, request ID)
- **builtin-middleware** - Recovery and security headers middleware
- **server-sent-events** - Real-time event streaming
- **static-files** - Serving static files, embedded files, and SPA fallback
- **oauth2-auth-code-pkce** - OAuth2 Authorization Code + PKCE flow
- **oauth2-client-credentials** - OAuth2 Client Credentials flow
- **oauth2-implicit** - OAuth2 Implicit flow

## Migration from Previous Versions

If upgrading from the monolithic version:

| Old Import | New Import |
|------------|------------|
| `github.com/joakimcarlsson/go-router/docs` | `github.com/joakimcarlsson/go-router/openapi` |
| `github.com/joakimcarlsson/go-router/metadata` | `github.com/joakimcarlsson/go-router/openapi` |
| `github.com/joakimcarlsson/go-router/swagger` | `github.com/joakimcarlsson/go-router/swaggerui` |
| `github.com/joakimcarlsson/go-router/integration` | `github.com/joakimcarlsson/go-router/swaggerui` |

| Old Usage | New Usage |
|-----------|-----------|
| `docs.WithSummary(...)` | `openapi.WithSummary(...)` |
| `metadata.RegisterTypeHandler(...)` | `openapi.RegisterTypeHandler(...)` |
| `integration.NewSwaggerUIIntegration(...)` | `swaggerui.NewSetup(...)` |
| `swaggerUI.SetupRoutes(...)` | `setup.RegisterRoutes(...)` |

## Contributing

We welcome contributions! Please see:

- **Issues**: Report bugs or request features
- **Pull Requests**: Submit improvements
- **Documentation**: Help improve examples and guides

### Development

```bash
# Run tests for all modules
cd router && go test -v ./...
cd ../openapi && go test -v ./...
cd ../swaggerui && go test -v ./...

# Run linter
golangci-lint run ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details
