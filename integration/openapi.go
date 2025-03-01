package integration

import (
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
)

// RouterOpenAPIAdapter combines Router with OpenAPI generation
type RouterOpenAPIAdapter struct {
	Router    *router.Router
	Generator *openapi.Generator
}

// NewRouterOpenAPIAdapter creates a new adapter
func NewRouterOpenAPIAdapter(r *router.Router, generator *openapi.Generator) *RouterOpenAPIAdapter {
	return &RouterOpenAPIAdapter{
		Router:    r,
		Generator: generator,
	}
}

// ExtractRouteInfo extracts OpenAPI route information from the router
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

// GenerateOpenAPISpec generates an OpenAPI specification from the router's routes
func (a *RouterOpenAPIAdapter) GenerateOpenAPISpec() *openapi.Spec {
	routeInfos := a.ExtractRouteInfo()
	return a.Generator.Generate(routeInfos)
}

// ServeHTTP implements http.Handler interface
func (a *RouterOpenAPIAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	spec := a.GenerateOpenAPISpec()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := openapi.WriteJSON(w, spec); err != nil {
		http.Error(w, "Failed to write OpenAPI spec", http.StatusInternalServerError)
	}
}
