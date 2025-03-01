package router

import (
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/joakimcarlsson/go-router/openapi"
)

// HandlerFunc defines a function to process HTTP requests in the context of the router
// MiddlewareFunc defines a function to wrap HandlerFunc for middleware processing
// route represents a single route with its method, path, handler, and metadata
// Router represents the main router with its configuration and routes

type HandlerFunc func(*Context)
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type route struct {
	method   string
	path     string
	handler  HandlerFunc
	metadata *RouteMetadata
}

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []MiddlewareFunc
	parent      *Router
	routes      []route
	mu          sync.RWMutex
	tags        []string
	security    []SecurityRequirement
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		mux:      http.NewServeMux(),
		prefix:   "",
		routes:   make([]route, 0),
		tags:     make([]string, 0),
		security: make([]SecurityRequirement, 0),
	}
}

// WithTags adds tags to a router group
func (r *Router) WithTags(tags ...string) *Router {
	r.tags = append(r.tags, tags...)
	return r
}

// WithSecurity adds security requirements to a router group
func (r *Router) WithSecurity(requirements ...map[string][]string) *Router {
	for _, req := range requirements {
		secReq := make(SecurityRequirement)
		for k, v := range req {
			secReq[k] = v
		}
		r.security = append(r.security, secReq)
	}
	return r
}

// Use adds middleware functions to the router
func (r *Router) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// Group creates a new router group with a specific path prefix and applies the provided function to it
func (r *Router) Group(path string, fn func(*Router)) {
	group := &Router{
		mux:         r.mux,
		prefix:      r.prefix + path,
		middlewares: slices.Clone(r.middlewares),
		parent:      r,
		routes:      make([]route, 0),
		tags:        make([]string, 0),
		security:    make([]SecurityRequirement, 0),
	}
	fn(group)

	r.mu.Lock()
	r.routes = append(r.routes, group.routes...)
	r.mu.Unlock()
}

// normalizePath cleans and normalizes the given path
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

// Handle registers a new route with the given pattern, handler, and options
func (r *Router) Handle(pattern string, handler HandlerFunc, opts ...RouteOption) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		panic("invalid route pattern format, expected 'METHOD /path'")
	}
	method, subpath := parts[0], parts[1]

	fullpath := normalizePath(path.Join(r.prefix, subpath))
	finalHandler := r.buildMiddlewareChain(handler)

	metadata := &RouteMetadata{
		Method:      method,
		Path:        fullpath,
		Parameters:  make([]Parameter, 0),
		Tags:        make([]string, 0),
		Responses:   make(map[string]Response),
		Security:    make([]SecurityRequirement, 0),
		RequestBody: nil,
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
		defer releaseContext(ctx)
		finalHandler(ctx)
	})
}

// GET registers a new GET route
func (r *Router) GET(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("GET "+path, handler, opts...)
}

// POST registers a new POST route
func (r *Router) POST(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("POST "+path, handler, opts...)
}

// PUT registers a new PUT route
func (r *Router) PUT(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PUT "+path, handler, opts...)
}

// DELETE registers a new DELETE route
func (r *Router) DELETE(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("DELETE "+path, handler, opts...)
}

// PATCH registers a new PATCH route
func (r *Router) PATCH(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PATCH "+path, handler, opts...)
}

// buildMiddlewareChain builds the middleware chain for a handler
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

// ServeHTTP implements the http.Handler interface for the router
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// collectRoutesRecursively gathers all routes from this router and its groups
func (r *Router) collectRoutesRecursively() []RouteMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]RouteMetadata, 0)

	// Collect routes from current router
	for _, route := range r.routes {
		if route.metadata != nil {
			metadata := *route.metadata
			routes = append(routes, metadata)
		}
	}

	return routes
}

// ServeOpenAPI serves the OpenAPI specification as JSON
func (r *Router) ServeOpenAPI(generator *openapi.Generator) HandlerFunc {
	return func(c *Context) {
		routes := r.collectRoutesRecursively()
		spec := generator.Generate(routes)
		c.JSON(200, spec)
	}
}
