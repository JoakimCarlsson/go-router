package router

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
