package router

import (
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/joakimcarlsson/go-router/pkg/http/openapi"
)

type HandlerFunc func(*Context)
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type route struct {
	method   string
	path     string
	handler  HandlerFunc
	metadata *openapi.RouteMetadata
}

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []MiddlewareFunc
	parent      *Router
	routes      []route
	mu          sync.RWMutex
	tags        []string // New field for group-level tags
}

func New() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		prefix: "",
		routes: make([]route, 0),
		tags:   make([]string, 0),
	}
}

// WithTags adds tags to a router group
func (r *Router) WithTags(tags ...string) *Router {
	r.tags = append(r.tags, tags...)
	return r
}

func (r *Router) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) Group(path string, fn func(*Router)) {
	group := &Router{
		mux:         r.mux,
		prefix:      r.prefix + path,
		middlewares: slices.Clone(r.middlewares),
		parent:      r,
		routes:      make([]route, 0),
		tags:        make([]string, 0), // Initialize tags slice
	}
	fn(group)

	// Add group's routes to parent
	r.mu.Lock()
	r.routes = append(r.routes, group.routes...)
	r.mu.Unlock()
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

func (r *Router) Handle(pattern string, handler HandlerFunc, opts ...openapi.RouteOption) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		panic("invalid route pattern format, expected 'METHOD /path'")
	}
	method, subpath := parts[0], parts[1]

	fullpath := normalizePath(path.Join(r.prefix, subpath))
	finalHandler := r.buildMiddlewareChain(handler)

	metadata := &openapi.RouteMetadata{
		Method:      method,
		Path:        fullpath,
		Parameters:  make([]openapi.Parameter, 0),
		Tags:        make([]string, 0),
		Responses:   make(map[string]openapi.Response),
		RequestBody: nil,
	}

	// Apply group tags first
	if len(r.tags) > 0 {
		metadata.Tags = append(metadata.Tags, r.tags...)
	}

	// Then apply route-specific tags
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

// Modified HTTP method handlers to accept options
func (r *Router) GET(path string, handler HandlerFunc, opts ...openapi.RouteOption) {
	r.Handle("GET "+path, handler, opts...)
}

func (r *Router) POST(path string, handler HandlerFunc, opts ...openapi.RouteOption) {
	r.Handle("POST "+path, handler, opts...)
}

func (r *Router) PUT(path string, handler HandlerFunc, opts ...openapi.RouteOption) {
	r.Handle("PUT "+path, handler, opts...)
}

func (r *Router) DELETE(path string, handler HandlerFunc, opts ...openapi.RouteOption) {
	r.Handle("DELETE "+path, handler, opts...)
}

func (r *Router) PATCH(path string, handler HandlerFunc, opts ...openapi.RouteOption) {
	r.Handle("PATCH "+path, handler, opts...)
}

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

// Static serves files from the given file system root.
// The path must end with "/*filepath" where the matched path will be used to serve files.
func (r *Router) Static(urlPath string, root string) {
	if !strings.HasSuffix(urlPath, "/*filepath") {
		panic("static path must end with /*filepath")
	}

	handler := http.StripPrefix(
		strings.TrimSuffix(urlPath, "/*filepath"),
		http.FileServer(http.Dir(root)),
	)

	pattern := "GET " + urlPath
	r.Handle(pattern, func(c *Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// collectRoutesRecursively gathers all routes from this router and its groups
func (r *Router) collectRoutesRecursively() []openapi.RouteMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]openapi.RouteMetadata, 0)

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
func (r *Router) ServeOpenAPI(info openapi.Info) HandlerFunc {
	return func(c *Context) {
		generator := openapi.NewGenerator(info)
		routes := r.collectRoutesRecursively()
		spec := generator.Generate(routes)
		c.JSON(200, spec)
	}
}
