package swaggerui

import (
	"encoding/json"
	"net/http"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router/v2"
)

// Setup provides integration between the router, OpenAPI generator, and Swagger UI.
type Setup struct {
	router    *router.Router
	generator *openapi.Generator
	uiConfig  UIConfig
}

// NewSetup creates a new Swagger UI setup.
func NewSetup(r *router.Router, generator *openapi.Generator) *Setup {
	return &Setup{
		router:    r,
		generator: generator,
		uiConfig:  DefaultUIConfig(),
	}
}

// WithUIConfig sets the Swagger UI configuration.
func (s *Setup) WithUIConfig(config UIConfig) *Setup {
	s.uiConfig = config
	return s
}

// RegisterRoutes registers the OpenAPI spec and Swagger UI routes.
func (s *Setup) RegisterRoutes(r *router.Router, specPath, docsPath string) {
	s.uiConfig.SpecURL = specPath

	r.GET(specPath, func(c *router.Context) {
		routes := s.collectRouteInfo()
		spec := s.generator.Generate(routes)

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(c.Writer).Encode(spec)
	}, openapi.ExcludeFromDocs())

	r.GET(docsPath, router.FromHTTPHandler(http.HandlerFunc(Handler(s.uiConfig))),
		openapi.ExcludeFromDocs(),
	)

	r.GET(docsPath+"/oauth2-redirect.html",
		router.FromHTTPHandler(http.HandlerFunc(OAuth2RedirectHandler())),
		openapi.ExcludeFromDocs(),
	)
}

func (s *Setup) collectRouteInfo() []openapi.RouteInfo {
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
