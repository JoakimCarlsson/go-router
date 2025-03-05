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
	middlewares []MiddlewareFunc
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

// Use adds middleware functions to the router.
// Middleware functions are executed in the order they are added,
// and apply to all routes registered after this call.
func (r *Router) Use(middlewares ...MiddlewareFunc) {
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
	finalHandler := r.buildMiddlewareChain(handler)

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
		handler:  finalHandler,
		metadata: metadata,
	})
	r.mu.Unlock()

	r.mux.HandleFunc(method+" "+fullpath, func(w http.ResponseWriter, req *http.Request) {
		ctx := acquireContext(w, req)
		ctx.maxMultipartMemory = r.maxMultipartMemory
		defer releaseContext(ctx)
		finalHandler(ctx)
	})
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

// buildMiddlewareChain builds the middleware chain for a handler.
// It applies each middleware in reverse order so that the first middleware
// in the list is the outermost wrapper around the handler.
func (r *Router) buildMiddlewareChain(handler HandlerFunc) HandlerFunc {
	if len(r.middlewares) == 0 {
		return handler
	}

	h := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
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
