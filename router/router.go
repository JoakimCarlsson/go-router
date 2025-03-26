package router

import (
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/joakimcarlsson/go-router/metadata"
)

// HandlerFunc defines a function to process HTTP requests in the context of the router.
// It receives a Context which encapsulates the HTTP request and response writer.
type HandlerFunc func(*Context)

// MiddlewareFunc defines a function that wraps a HandlerFunc for middleware processing.
// Middleware functions can perform pre-processing before calling the next handler,
// or post-processing after the handler returns.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// route represents an internal route definition with its HTTP method, path pattern,
// handler function and metadata for documentation.
type route struct {
	method   string
	path     string
	handler  HandlerFunc
	metadata *metadata.RouteMetadata
}

// Router is the main HTTP router that registers routes and dispatches requests to handlers.
// It supports middleware, route groups, and OpenAPI documentation generation.
type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []func(http.Handler) http.Handler
	parent      *Router
	routes      []route
	mu          sync.RWMutex
	tags        []string
	security    []metadata.SecurityRequirement
	// maxMultipartMemory is the max memory used to parse multipart forms in bytes
	maxMultipartMemory int64
}

// New creates a new Router instance with default configuration.
// The returned router is ready to register routes and handle HTTP requests.
func New() *Router {
	return &Router{
		mux:                http.NewServeMux(),
		prefix:             "",
		middlewares:        make([]func(http.Handler) http.Handler, 0),
		routes:             make([]route, 0),
		tags:               make([]string, 0),
		security:           make([]metadata.SecurityRequirement, 0),
		maxMultipartMemory: 32 << 20, // 32 MB
	}
}

// WithTags adds OpenAPI tags to a router group.
// Tags are used to group operations in the OpenAPI documentation.
// Returns the router for method chaining.
func (r *Router) WithTags(tags ...string) *Router {
	r.tags = append(r.tags, tags...)
	return r
}

// WithSecurity adds security requirements to a router group.
// All routes registered with this router will inherit these security requirements.
// Returns the router for method chaining.
func (r *Router) WithSecurity(requirements ...map[string][]string) *Router {
	for _, req := range requirements {
		secReq := make(metadata.SecurityRequirement)
		for k, v := range req {
			secReq[k] = v
		}
		r.security = append(r.security, secReq)
	}
	return r
}

// Use adds standard HTTP middleware to the router.
//
// It accepts middleware that follows the standard Go HTTP middleware pattern:
// func(http.Handler) http.Handler
//
// Example usage:
//
//	// Built-in middleware
//	r.Use(cors.Default())
//
//	// Third-party middleware
//	r.Use(nosurf.New)
//
//	// Multiple middleware
//	r.Use(logger, recovery, cors.Default())
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// Group creates a new router group with a specific path prefix.
// The provided function is called with the new group as an argument,
// allowing routes to be registered within the group.
func (r *Router) Group(path string, fn func(*Router)) {
	group := &Router{
		mux:         r.mux,
		prefix:      r.prefix + path,
		middlewares: slices.Clone(r.middlewares),
		parent:      r,
		routes:      make([]route, 0),
		tags:        make([]string, 0),
		security:    make([]metadata.SecurityRequirement, 0),
	}
	fn(group)

	r.mu.Lock()
	r.routes = append(r.routes, group.routes...)
	r.mu.Unlock()
}

// Handle registers a new route with the given pattern and handler.
// The pattern must be in the format "METHOD /path".
// Route options can be provided to add OpenAPI documentation to the route.
func (r *Router) Handle(pattern string, handler HandlerFunc, opts ...RouteOption) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		panic("invalid route pattern format, expected 'METHOD /path'")
	}
	method, subpath := parts[0], parts[1]

	fullpath := normalizePath(path.Join(r.prefix, subpath))

	metadata := &metadata.RouteMetadata{
		Method:     method,
		Path:       fullpath,
		Parameters: make([]metadata.Parameter, 0),
		Tags:       make([]string, 0),
		Responses:  make(map[string]metadata.Response),
		Security:   make([]metadata.SecurityRequirement, 0),
	}

	if len(r.tags) > 0 {
		metadata.Tags = append(metadata.Tags, r.tags...)
	}

	if len(r.security) > 0 {
		metadata.Security = append(metadata.Security, r.security...)
	}

	for _, opt := range opts {
		opt(metadata)
	}

	r.mu.Lock()
	r.routes = append(r.routes, route{
		method:   method,
		path:     fullpath,
		handler:  handler,
		metadata: metadata,
	})
	r.mu.Unlock()

	// Create a handler chain with middleware
	var httpHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := acquireContext(w, req)
		ctx.maxMultipartMemory = r.maxMultipartMemory
		defer releaseContext(ctx)
		handler(ctx)
	})

	// Apply middleware in reverse order so that the first middleware
	// in the list is the outermost wrapper around the handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		httpHandler = r.middlewares[i](httpHandler)
	}

	r.mux.Handle(method+" "+fullpath, httpHandler)
}

// GET registers a new GET route with the specified path and handler.
// Options can be provided to add OpenAPI documentation to the route.
func (r *Router) GET(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("GET "+path, handler, opts...)
}

// POST registers a new POST route with the specified path and handler.
// Options can be provided to add OpenAPI documentation to the route.
func (r *Router) POST(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("POST "+path, handler, opts...)
}

// PUT registers a new PUT route with the specified path and handler.
// Options can be provided to add OpenAPI documentation to the route.
func (r *Router) PUT(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PUT "+path, handler, opts...)
}

// DELETE registers a new DELETE route with the specified path and handler.
// Options can be provided to add OpenAPI documentation to the route.
func (r *Router) DELETE(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("DELETE "+path, handler, opts...)
}

// PATCH registers a new PATCH route with the specified path and handler.
// Options can be provided to add OpenAPI documentation to the route.
func (r *Router) PATCH(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PATCH "+path, handler, opts...)
}

// WithMultipartConfig sets the maximum memory allocation for multipart form data parsing.
// This affects how much of a file upload will be stored in memory before being written to disk.
// Default is 32MB if not specified.
func (r *Router) WithMultipartConfig(maxMemory int64) *Router {
	r.maxMultipartMemory = maxMemory
	return r
}

// ServeHTTP implements the http.Handler interface.
// This allows the router to be used directly with http.ListenAndServe.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Routes returns all registered routes.
// This is used primarily for OpenAPI documentation generation.
func (r *Router) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]Route, 0, len(r.routes))
	for _, rt := range r.routes {
		routes = append(routes, Route{
			Method:   rt.method,
			Path:     rt.path,
			Handler:  rt.handler,
			Metadata: rt.metadata,
		})
	}
	return routes
}

// AutoRegisterOptions automatically registers OPTIONS handlers for all routes in the current router.
// This ensures that preflight requests for CORS will be handled properly.
// Call this method after registering all your routes and before starting the server.
//
// For CORS to work correctly with browsers, the server must respond to preflight OPTIONS requests.
// Without this method or manually registered OPTIONS handlers, browsers will block cross-origin
// requests to your API endpoints.
//
// Example usage:
//
//	r := router.New()
//	r.Use(cors.Handler(...))
//
//	// Register your routes
//	r.GET("/api/users", getUsersHandler)
//	r.POST("/api/users", createUserHandler)
//
//	// Auto-register OPTIONS handlers for all routes
//	r.AutoRegisterOptions()
//
//	http.ListenAndServe(":8080", r)
func (r *Router) AutoRegisterOptions() *Router {
	r.mu.RLock()

	// Only collect paths from the current router instance, not all router instances
	routes := make(map[string]bool)

	for _, route := range r.routes {
		// Only include routes registered directly on this router instance
		// by checking the route's path against the router's prefix
		if r.parent == nil || strings.HasPrefix(route.path, r.prefix) {
			// Extract the path relative to this router's prefix for comparing
			routePath := route.path
			routes[routePath] = true
		}
	}
	r.mu.RUnlock()

	// For each registered path, add an OPTIONS handler if one doesn't already exist
	for path := range routes {
		found := false

		// Check if OPTIONS handler already exists for this path
		r.mu.RLock()
		for _, route := range r.routes {
			if route.method == "OPTIONS" && route.path == path {
				found = true
				break
			}
		}
		r.mu.RUnlock()

		// If no OPTIONS handler exists, register an empty one
		if !found {
			r.Handle("OPTIONS "+path, func(c *Context) {
				// Empty handler - the CORS middleware will handle the response
				c.Status(http.StatusNoContent)
			})
		}
	}

	return r
}

// normalizePath ensures the path starts with a slash and is cleaned.
// It handles edge cases like empty paths and relative paths.
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

// ToHTTPHandlerFunc converts a router.HandlerFunc to a standard http.HandlerFunc.
// This enables using router handlers with standard Go HTTP servers or middleware.
//
// Example usage:
//
//	http.Handle("/api/users", router.ToHTTPHandlerFunc(myRouterHandler))
func ToHTTPHandlerFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := acquireContext(w, r)
		defer releaseContext(ctx)
		h(ctx)
	}
}

// FromHTTPHandler converts a standard http.Handler to a router.HandlerFunc.
// This allows you to use existing http.Handler implementations with this router.
//
// Example usage:
//
//	r := router.New()
//	fileServer := http.FileServer(http.Dir("./static"))
//	r.GET("/static/*filepath", router.FromHTTPHandler(fileServer))
func FromHTTPHandler(handler http.Handler) HandlerFunc {
	return func(c *Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
