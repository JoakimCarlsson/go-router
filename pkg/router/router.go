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

type routeCacheKey struct {
	method string
	path   string
}

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []MiddlewareFunc
	parent      *Router
	routeCache  sync.Map
}

func New() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		prefix: "",
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

func (r *Router) Handle(pattern string, handler HandlerFunc) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		panic("invalid route pattern format, expected 'METHOD /path'")
	}
	method, subpath := parts[0], parts[1]

	fullpath := normalizePath(path.Join(r.prefix, subpath))

	finalHandler := r.buildMiddlewareChain(handler)

	cacheKey := routeCacheKey{
		method: method,
		path:   fullpath,
	}
	r.routeCache.Store(cacheKey, finalHandler)

	r.mux.HandleFunc(method+" "+fullpath, func(w http.ResponseWriter, req *http.Request) {
		ctx := acquireContext(w, req)
		defer releaseContext(ctx)

		if handler, ok := r.routeCache.Load(routeCacheKey{
			method: req.Method,
			path:   req.URL.Path,
		}); ok {
			handler.(HandlerFunc)(ctx)
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

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
