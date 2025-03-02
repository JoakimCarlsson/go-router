package integration

import (
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// SwaggerUIIntegration combines OpenAPI specification with Swagger UI.
// It provides a clean way to connect the OpenAPI generator to the router
// and serve a Swagger UI interface for API documentation.
type SwaggerUIIntegration struct {
	// OpenAPIAdapter provides the OpenAPI specification for the UI
	OpenAPIAdapter *RouterOpenAPIAdapter
	// UIConfig contains configuration for the Swagger UI
	UIConfig swagger.UIConfig
}

// NewSwaggerUIIntegration creates a new Swagger UI integration.
// It initializes the integration with the provided router and OpenAPI generator.
//
// Parameters:
//   - r: The router containing the routes to document
//   - generator: The OpenAPI generator to use for creating the specification
func NewSwaggerUIIntegration(r *router.Router, generator *openapi.Generator) *SwaggerUIIntegration {
	return &SwaggerUIIntegration{
		OpenAPIAdapter: NewRouterOpenAPIAdapter(r, generator),
		UIConfig:       swagger.DefaultUIConfig(),
	}
}

// WithUIConfig updates the Swagger UI configuration.
// This allows customizing the UI appearance, behavior, and features.
//
// Parameters:
//   - config: The Swagger UI configuration to use
//
// Returns the SwaggerUIIntegration for method chaining.
func (s *SwaggerUIIntegration) WithUIConfig(config swagger.UIConfig) *SwaggerUIIntegration {
	s.UIConfig = config
	return s
}

// SetupRoutes sets up the OpenAPI JSON and Swagger UI routes on the router.
// This registers two routes:
//  1. A route to serve the OpenAPI JSON specification
//  2. A route to serve the Swagger UI that consumes the specification
//
// Parameters:
//   - r: The router to register routes on
//   - specPath: The path to serve the OpenAPI JSON specification (e.g., "/openapi.json")
//   - uiPath: The path to serve the Swagger UI (e.g., "/docs")
func (s *SwaggerUIIntegration) SetupRoutes(r *router.Router, specPath, uiPath string) {
	// Serve OpenAPI JSON
	r.GET(specPath, wrapHandler(s.OpenAPIAdapter.ServeHTTP))

	// Configure UI to use the correct spec path
	s.UIConfig.SpecURL = specPath

	// Serve Swagger UI
	r.GET(uiPath, wrapHandler(swagger.Handler(s.UIConfig)))
}

// wrapHandler converts an http.HandlerFunc to a router.HandlerFunc.
// This is a helper function to bridge between the standard library's http
// package and the router's custom handler type.
func wrapHandler(h http.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		h(c.Writer, c.Request)
	}
}
