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
	options []RouteOption
}

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

func (r *Router) WithOptions(opts ...RouteOption) *Router {
	r.groupOptions = append(r.groupOptions, opts...)
	return r
}

func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

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

func (r *Router) GET(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("GET "+path, handler, opts...)
}

func (r *Router) POST(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("POST "+path, handler, opts...)
}

func (r *Router) PUT(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PUT "+path, handler, opts...)
}

func (r *Router) DELETE(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("DELETE "+path, handler, opts...)
}

func (r *Router) PATCH(path string, handler HandlerFunc, opts ...RouteOption) {
	r.Handle("PATCH "+path, handler, opts...)
}

func (r *Router) WithMultipartConfig(maxMemory int64) *Router {
	r.maxMultipartMemory = maxMemory
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

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

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

func ToHTTPHandlerFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := acquireContext(w, r)
		defer releaseContext(ctx)
		h(ctx)
	}
}

func FromHTTPHandler(handler http.Handler) HandlerFunc {
	return func(c *Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
