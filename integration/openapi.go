package integration

import (
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
)

// RouterOpenAPIAdapter combines Router with OpenAPI generation.
// This adapter connects the router's route definitions with
// the OpenAPI generator to produce OpenAPI specifications.
type RouterOpenAPIAdapter struct {
	// Router is the router containing the routes to document
	Router *router.Router
	// Generator is the OpenAPI generator used to create the specification
	Generator *openapi.Generator
}

// NewRouterOpenAPIAdapter creates a new adapter.
// It initializes the adapter with a router and OpenAPI generator.
//
// Parameters:
//   - r: The router containing the routes to document
//   - generator: The OpenAPI generator to use
func NewRouterOpenAPIAdapter(r *router.Router, generator *openapi.Generator) *RouterOpenAPIAdapter {
	return &RouterOpenAPIAdapter{
		Router:    r,
		Generator: generator,
	}
}

// ExtractRouteInfo extracts OpenAPI route information from the router.
// It converts the router's route metadata to the format expected by
// the OpenAPI generator.
func (a *RouterOpenAPIAdapter) ExtractRouteInfo() []openapi.RouteInfo {
	routes := a.Router.Routes()
	routeInfos := make([]openapi.RouteInfo, 0, len(routes))

	for _, route := range routes {
		// Convert RouteMetadata to RouteInfo
		if route.Metadata != nil {
			routeInfos = append(routeInfos, openapi.RouteInfoFromMetadata(*route.Metadata))
		}
	}

	return routeInfos
}

// GenerateOpenAPISpec generates an OpenAPI specification from the router's routes.
// This creates a complete OpenAPI specification document based on the
// route metadata and configuration in the generator.
func (a *RouterOpenAPIAdapter) GenerateOpenAPISpec() *openapi.Spec {
	routeInfos := a.ExtractRouteInfo()
	return a.Generator.Generate(routeInfos)
}

// ServeHTTP implements http.Handler interface.
// This allows the adapter to be used as an HTTP handler to serve
// the OpenAPI specification as JSON.
func (a *RouterOpenAPIAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	spec := a.GenerateOpenAPISpec()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := openapi.WriteJSON(w, spec); err != nil {
		http.Error(w, "Failed to write OpenAPI spec", http.StatusInternalServerError)
	}
}
