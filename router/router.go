package router

import (
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"
)

// HandlerFunc defines a function to process HTTP requests in the context of the router.
// It receives a Context which encapsulates the HTTP request and response writer.
type HandlerFunc func(*Context)

// MiddlewareFunc defines a function that wraps a HandlerFunc for middleware processing.
// Middleware functions can perform pre-processing before calling the next handler,
// or post-processing after the handler returns.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// route represents an internal route definition with its HTTP method, path pattern,
// handler function and options for documentation.
type route struct {
	method  string
	path    string
	handler HandlerFunc
	options []RouteOption
}

// handlerWrapper wraps a HandlerFunc to be compatible with http.Handler.
// This is reused to avoid allocation overhead of anonymous functions.
type handlerWrapper struct {
	handler            HandlerFunc
	maxMultipartMemory int64
}

func (hw *handlerWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := contextPool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = req
	ctx.ctx = req.Context()
	ctx.startTimeSet = false
	ctx.StatusCode = 200
	ctx.maxMultipartMemory = hw.maxMultipartMemory
	ctx.sseInitialized = false
	ctx.queryCache = nil
	ctx.statusWritten = false

	defer func() {
		if err := recover(); err != nil {
			panic(err)
		}
	}()

	defer func() {
		ctx.Writer = nil
		ctx.Request = nil
		ctx.queryCache = nil
		ctx.statusWritten = false
		if len(ctx.store) > 0 {
			clearInterfaceMap(ctx.store)
		}
		contextPool.Put(ctx)
	}()

	hw.handler(ctx)
}

// Router is the main HTTP router that registers routes and dispatches requests to handlers.
// It supports middleware, route groups, and OpenAPI documentation generation.
type Router struct {
	mux                *http.ServeMux
	prefix             string
	middlewares        []func(http.Handler) http.Handler
	parent             *Router
	routes             []route
	mu                 sync.RWMutex
	groupOptions       []RouteOption
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
		groupOptions:       make([]RouteOption, 0),
		maxMultipartMemory: 32 << 20,
	}
}

// WithOptions adds route options to a router group.
// Options are used for OpenAPI documentation configuration.
// Returns the router for method chaining.
func (r *Router) WithOptions(opts ...RouteOption) *Router {
	r.groupOptions = append(r.groupOptions, opts...)
	return r
}

// Use adds standard HTTP middleware to the router.
//
// It accepts middleware that follows the standard Go HTTP middleware pattern:
// func(http.Handler) http.Handler
//
// Example usage:
//
//	r.Use(cors.Default())
//	r.Use(logger, recovery, cors.Default())
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// Group creates a route group with a path prefix and shared middleware.
//
// Example usage:
//
//	r.Group("/api/v1", func(api *router.Router) {
//	    api.Use(authMiddleware)
//	    api.GET("/users", listUsers)
//	    api.POST("/users", createUser)
//	})
func (r *Router) Group(path string, fn func(*Router)) {
	group := &Router{
		mux:                r.mux,
		prefix:             r.prefix + path,
		middlewares:        slices.Clone(r.middlewares),
		parent:             r,
		routes:             make([]route, 0),
		groupOptions:       slices.Clone(r.groupOptions),
		maxMultipartMemory: r.maxMultipartMemory,
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

	if len(fullpath) > 1 && fullpath[len(fullpath)-1] == '/' {
		fullpath = fullpath[:len(fullpath)-1]
	}

	allOpts := make([]RouteOption, 0, len(r.groupOptions)+len(opts))
	allOpts = append(allOpts, r.groupOptions...)
	allOpts = append(allOpts, opts...)

	r.mu.Lock()
	r.routes = append(r.routes, route{
		method:  method,
		path:    fullpath,
		handler: handler,
		options: allOpts,
	})
	r.mu.Unlock()

	var httpHandler http.Handler = &handlerWrapper{
		handler:            handler,
		maxMultipartMemory: r.maxMultipartMemory,
	}

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		httpHandler = r.middlewares[i](httpHandler)
	}

	r.mux.Handle(method+" "+fullpath, httpHandler)
}

// GET registers a GET route with the given path and handler.
// Route options can be provided to add OpenAPI documentation.
func (r *Router) GET(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("GET "+path, handler, opts...)
}

// POST registers a POST route with the given path and handler.
// Route options can be provided to add OpenAPI documentation.
func (r *Router) POST(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("POST "+path, handler, opts...)
}

// PUT registers a PUT route with the given path and handler.
// Route options can be provided to add OpenAPI documentation.
func (r *Router) PUT(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PUT "+path, handler, opts...)
}

// DELETE registers a DELETE route with the given path and handler.
// Route options can be provided to add OpenAPI documentation.
func (r *Router) DELETE(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("DELETE "+path, handler, opts...)
}

// PATCH registers a PATCH route with the given path and handler.
// Route options can be provided to add OpenAPI documentation.
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
// This is used by external modules (e.g., OpenAPI) to collect route information.
func (r *Router) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]Route, 0, len(r.routes))
	for _, rt := range r.routes {
		routes = append(routes, Route{
			Method:  rt.method,
			Path:    rt.path,
			Handler: rt.handler,
			Options: rt.options,
		})
	}
	return routes
}

// AutoRegisterOptions automatically registers OPTIONS handlers for all routes.
// This is useful for CORS preflight requests.
// Returns the router for method chaining.
func (r *Router) AutoRegisterOptions() *Router {
	r.mu.RLock()

	routes := make(map[string]bool)

	for _, route := range r.routes {
		if r.parent == nil || strings.HasPrefix(route.path, r.prefix) {
			routePath := route.path
			routes[routePath] = true
		}
	}
	r.mu.RUnlock()

	for path := range routes {
		found := false

		r.mu.RLock()
		for _, route := range r.routes {
			if route.method == "OPTIONS" && route.path == path {
				found = true
				break
			}
		}
		r.mu.RUnlock()

		if !found {
			r.Handle("OPTIONS "+path, func(c *Context) {
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
// This allows router handlers to be used in contexts that expect standard HTTP handlers.
func ToHTTPHandlerFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := acquireContext(w, r)
		defer releaseContext(ctx)
		h(ctx)
	}
}

// FromHTTPHandler converts a standard http.Handler to a router.HandlerFunc.
// This allows standard HTTP handlers to be used with the router.
func FromHTTPHandler(handler http.Handler) HandlerFunc {
	return func(c *Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
