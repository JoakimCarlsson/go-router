package router

import (
	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/metadata"
)

// Route represents a single route with its method, path, handler, and metadata
type Route struct {
	Method   string
	Path     string
	Handler  HandlerFunc
	Metadata *metadata.RouteMetadata
}

// RouteOption is a function that configures route metadata
type RouteOption = docs.RouteOption

// RouteConfig is used to provide configuration options for routes
type RouteConfig struct {
	// Core routing properties
	Method  string
	Path    string
	Handler HandlerFunc

	// Optional route metadata
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Deprecated  bool
}

// NewRoute creates a new route with the given configuration
func NewRoute(config RouteConfig) Route {
	metadata := &metadata.RouteMetadata{
		Method:      config.Method,
		Path:        config.Path,
		OperationID: config.OperationID,
		Summary:     config.Summary,
		Description: config.Description,
		Tags:        config.Tags,
		Deprecated:  config.Deprecated,
		Responses:   make(map[string]metadata.Response),
	}

	return Route{
		Method:   config.Method,
		Path:     config.Path,
		Handler:  config.Handler,
		Metadata: metadata,
	}
}
