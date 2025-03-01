package integration

import (
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// SwaggerUIIntegration combines OpenAPI specification with Swagger UI
type SwaggerUIIntegration struct {
	OpenAPIAdapter *RouterOpenAPIAdapter
	UIConfig       swagger.UIConfig
}

// NewSwaggerUIIntegration creates a new Swagger UI integration
func NewSwaggerUIIntegration(r *router.Router, generator *openapi.Generator) *SwaggerUIIntegration {
	return &SwaggerUIIntegration{
		OpenAPIAdapter: NewRouterOpenAPIAdapter(r, generator),
		UIConfig:       swagger.DefaultUIConfig(),
	}
}

// WithUIConfig updates the Swagger UI configuration
func (s *SwaggerUIIntegration) WithUIConfig(config swagger.UIConfig) *SwaggerUIIntegration {
	s.UIConfig = config
	return s
}

// SetupRoutes sets up the OpenAPI JSON and Swagger UI routes on the router
func (s *SwaggerUIIntegration) SetupRoutes(r *router.Router, specPath, uiPath string) {
	// Serve OpenAPI JSON
	r.GET(specPath, wrapHandler(s.OpenAPIAdapter.ServeHTTP))

	// Configure UI to use the correct spec path
	s.UIConfig.SpecURL = specPath

	// Serve Swagger UI
	r.GET(uiPath, wrapHandler(swagger.Handler(s.UIConfig)))
}

// helper to convert http.HandlerFunc to router.HandlerFunc
func wrapHandler(h http.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		h(c.Writer, c.Request)
	}
}
