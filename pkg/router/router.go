package router

import (
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"
)

type HandlerFunc func(*Context)
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type route struct {
	method  string
	path    string
	handler HandlerFunc
}

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []MiddlewareFunc
	parent      *Router
	routes      []route
	mu          sync.RWMutex
}

func New() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		prefix: "",
		routes: make([]route, 0),
	}
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
	}
	fn(group)
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

func (r *Router) findRoute(method, path string) (HandlerFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		if route.method == method && route.path == path {
			return route.handler, true
		}
	}
	return nil, false
}

func (r *Router) Handle(pattern string, handler HandlerFunc) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		panic("invalid route pattern format, expected 'METHOD /path'")
	}
	method, subpath := parts[0], parts[1]

	fullpath := normalizePath(path.Join(r.prefix, subpath))
	finalHandler := r.buildMiddlewareChain(handler)

	r.mu.Lock()
	r.routes = append(r.routes, route{
		method:  method,
		path:    fullpath,
		handler: finalHandler,
	})
	r.mu.Unlock()

	r.mux.HandleFunc(method+" "+fullpath, func(w http.ResponseWriter, req *http.Request) {
		ctx := acquireContext(w, req)
		defer releaseContext(ctx)

		if handler, ok := r.findRoute(req.Method, req.URL.Path); ok {
			handler(ctx)
			return
		}

		finalHandler(ctx)
	})
}

func (r *Router) GET(path string, handler HandlerFunc) {
	r.Handle("GET "+path, handler)
}

func (r *Router) POST(path string, handler HandlerFunc) {
	r.Handle("POST "+path, handler)
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	r.Handle("PUT "+path, handler)
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.Handle("DELETE "+path, handler)
}

func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.Handle("PATCH "+path, handler)
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
