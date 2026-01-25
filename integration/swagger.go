package integration

import (
	"encoding/json"
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

// SwaggerUIIntegration provides integration between the router, OpenAPI generator, and Swagger UI.
type SwaggerUIIntegration struct {
	router    *router.Router
	generator *openapi.Generator
	uiConfig  swaggerui.UIConfig
}

// NewSwaggerUIIntegration creates a new SwaggerUIIntegration.
func NewSwaggerUIIntegration(r *router.Router, generator *openapi.Generator) *SwaggerUIIntegration {
	return &SwaggerUIIntegration{
		router:    r,
		generator: generator,
		uiConfig:  swaggerui.DefaultUIConfig(),
	}
}

// WithUIConfig sets the Swagger UI configuration.
func (s *SwaggerUIIntegration) WithUIConfig(config swaggerui.UIConfig) *SwaggerUIIntegration {
	s.uiConfig = config
	return s
}

// SetupRoutes registers the OpenAPI spec and Swagger UI routes.
func (s *SwaggerUIIntegration) SetupRoutes(r *router.Router, specPath, docsPath string) {
	s.uiConfig.SpecURL = specPath

	r.GET(specPath, func(c *router.Context) {
		routes := s.collectRouteInfo()
		spec := s.generator.Generate(routes)

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(c.Writer).Encode(spec)
	}, openapi.ExcludeFromDocs())

	r.GET(docsPath, router.FromHTTPHandler(http.HandlerFunc(swaggerui.Handler(s.uiConfig))),
		openapi.ExcludeFromDocs(),
	)

	r.GET(docsPath+"/oauth2-redirect.html",
		router.FromHTTPHandler(http.HandlerFunc(swaggerui.OAuth2RedirectHandler())),
		openapi.ExcludeFromDocs(),
	)
}

func (s *SwaggerUIIntegration) collectRouteInfo() []openapi.RouteInfo {
	routes := s.router.Routes()
	routeInfoList := make([]openapi.RouteInfo, 0, len(routes))

	for _, route := range routes {
		metadata := openapi.MetadataFromOptions(route.Method, route.Path, route.Options)

		if metadata.ExcludeFromDocs {
			continue
		}

		routeInfoList = append(routeInfoList, openapi.RouteInfoFromMetadata(*metadata))
	}

	return routeInfoList
}
