# go-router Middleware

This directory contains official middleware implementations for use with the go-router package.

## Available Middleware

### CORS Middleware

A dedicated CORS middleware is available in the [cors package](cors/README.md).

```go
import "github.com/joakimcarlsson/go-router/router/middleware/cors"

// Default configuration
r.Use(cors.Default())

// Custom configuration
r.Use(cors.Handler(cors.Options{
    AllowOrigins: []string{"https://example.com"},
    // Other options...
}))
```

#### Configuration Options

The `CORSOptions` struct provides the following configuration options:

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `AllowOrigins` | `[]string` | List of allowed origins. Can contain wildcards like `https://*.example.com`. Use `*` to allow all origins. | `["*"]` |
| `AllowMethods` | `[]string` | HTTP methods allowed for CORS requests. | `[GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH]` |
| `AllowHeaders` | `[]string` | HTTP headers clients can use in requests. Use `*` to allow all headers. | `[]` (none) |
| `ExposeHeaders` | `[]string` | HTTP headers that should be exposed to clients. | `[]` (none) |
| `AllowCredentials` | `bool` | Allow cookies and credentials to be sent with requests. | `false` |
| `MaxAge` | `int` | How long (in seconds) the results of a preflight request can be cached. | `0` (no caching) |
| `OptionsPassthrough` | `bool` | Pass OPTIONS requests to handlers instead of responding automatically. | `false` |

#### Advanced Usage Examples

**Allowing Multiple Origins**

```go
r.Use(middleware.Handler(middleware.CORSOptions{
    AllowOrigins: []string{
        "https://example.com",
        "https://api.example.com",
        "https://admin.example.com",
    },
}))
```

**Domain Wildcard Support**

```go
r.Use(middleware.Handler(middleware.CORSOptions{
    AllowOrigins: []string{
        "https://*.example.com", // Allows any subdomain of example.com
    },
}))
```

**Different CORS Settings for Different Route Groups**

```go
// Main router with default CORS
r := router.New()
r.Use(middleware.CORS())

// API group with stricter CORS
r.Group("/api", func(api *router.Router) {
    api.Use(middleware.Handler(middleware.CORSOptions{
        AllowOrigins:     []string{"https://api.example.com"},
        AllowCredentials: true,
    }))
    
    // API routes...
})
```

#### Security Considerations

1. Avoid using `AllowOrigins: []string{"*"}` with `AllowCredentials: true` as this is not secure and not allowed by browsers.
2. Specify exact origins rather than using wildcard `*` in production environments.
3. Only expose headers that are necessary for your client applications.
4. Limit allowed methods to only those your API actually supports. 