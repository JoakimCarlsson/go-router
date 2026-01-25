package router

import "time"

// RouteOption is an interface for route configuration options.
// It allows for fluent API-style configuration of routes with documentation.
// Concrete option types are defined in the openapi package.
type RouteOption interface{}

// Route represents a single route with its method, path, handler, and options.
// It provides a public interface to access route information, primarily used for
// OpenAPI documentation generation.
type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
	Options []RouteOption
}

// RouteConfig is used to provide configuration options for routes.
// It contains both core routing properties and optional documentation metadata.
type RouteConfig struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Deprecated  bool
}

// NewRoute creates a new route with the given configuration.
// It initializes the route with the provided configuration options
// and returns a fully configured Route instance.
func NewRoute(config RouteConfig) Route {
	return Route{
		Method:  config.Method,
		Path:    config.Path,
		Handler: config.Handler,
		Options: []RouteOption{},
	}
}

// RouteRegistration represents a registered route that can be configured with additional options.
// It provides a fluent API for configuring output caching and other route-level features.
type RouteRegistration struct {
	router *Router
	method string
	path   string
}

// CacheConfig holds the output cache configuration for a route.
type CacheConfig struct {
	Duration time.Duration
	Options  []interface{}
}

// WithOutputCache enables output caching for this route with the specified duration and options.
// Returns the RouteRegistration for method chaining.
//
// Example:
//
//	r.GET("/products", handler).WithOutputCache(time.Minute)
func (rr *RouteRegistration) WithOutputCache(duration time.Duration, opts ...interface{}) *RouteRegistration {
	rr.router.setCacheConfig(rr.method, rr.path, CacheConfig{
		Duration: duration,
		Options:  opts,
	})
	return rr
}

// WithProfile enables output caching for this route using a named cache profile.
// Returns the RouteRegistration for method chaining.
//
// Example:
//
//	r.GET("/products", handler).WithProfile("aggressive")
func (rr *RouteRegistration) WithProfile(profileName string) *RouteRegistration {
	rr.router.setCacheProfile(rr.method, rr.path, profileName)
	return rr
}

// GetCacheProfile retrieves the cache profile name for a specific route.
// Returns empty string if no profile is configured.
func (r *Router) GetCacheProfile(method, path string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := method + " " + path
	if profile, ok := r.cacheProfiles[key]; ok {
		return profile
	}
	return ""
}

func (r *Router) setCacheProfile(method, path, profileName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cacheProfiles == nil {
		r.cacheProfiles = make(map[string]string)
	}

	key := method + " " + path
	r.cacheProfiles[key] = profileName
}

// GetCacheConfig retrieves the cache configuration for a specific route.
// Returns nil if no cache configuration is set for the route.
func (r *Router) GetCacheConfig(method, path string) *CacheConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	key := method + " " + path
	if cfg, ok := r.cacheConfigs[key]; ok {
		return &cfg
	}
	return nil
}

func (r *Router) setCacheConfig(method, path string, config CacheConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.cacheConfigs == nil {
		r.cacheConfigs = make(map[string]CacheConfig)
	}
	
	key := method + " " + path
	r.cacheConfigs[key] = config
}
