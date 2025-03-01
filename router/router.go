package router

import (
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/joakimcarlsson/go-router/metadata"
)

// HandlerFunc defines a function to process HTTP requests in the context of the router
type HandlerFunc func(*Context)

// MiddlewareFunc defines a function to wrap HandlerFunc for middleware processing
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type route struct {
	method   string
	path     string
	handler  HandlerFunc
	metadata *metadata.RouteMetadata
}

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []MiddlewareFunc
	parent      *Router
	routes      []route
	mu          sync.RWMutex
	tags        []string
	security    []metadata.SecurityRequirement
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		mux:      http.NewServeMux(),
		prefix:   "",
		routes:   make([]route, 0),
		tags:     make([]string, 0),
		security: make([]metadata.SecurityRequirement, 0),
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
		secReq := make(metadata.SecurityRequirement)
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

// Group creates a new router group with a specific path prefix
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

// Handle registers a new route with the given pattern and handler
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

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Routes returns all registered routes
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

// normalizePath ensures the path starts with a slash and is cleaned
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}
