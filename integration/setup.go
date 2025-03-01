package integration

import (
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// SetupOptions holds configuration for setting up API documentation.
// It provides a single place to configure both OpenAPI and Swagger UI.
type SetupOptions struct {
	// OpenAPI configuration
	Title       string // API title for OpenAPI info section
	Version     string // API version for OpenAPI info section
	Description string // API description for OpenAPI info section

	// Route paths
	SpecPath string // Path to serve OpenAPI JSON (default: /openapi.json)
	DocsPath string // Path to serve Swagger UI (default: /docs)

	// UI customization
	DarkMode bool   // Enable dark mode in Swagger UI
	UITitle  string // Custom title for Swagger UI page (defaults to Title if not set)

	// Security schemes
	UseBasicAuth  bool // Add basic auth security scheme
	UseBearerAuth bool // Add bearer token security scheme
	UseAPIKey     bool // Add API key security scheme
}

// DefaultSetupOptions returns default setup options for API documentation.
// It sets sensible defaults for title, version, paths, and UI options.
func DefaultSetupOptions() SetupOptions {
	return SetupOptions{
		Title:       "API Documentation",
		Version:     "1.0.0",
		Description: "API documentation powered by OpenAPI and Swagger UI",
		SpecPath:    "/openapi.json",
		DocsPath:    "/docs",
		DarkMode:    false,
	}
}

// Setup configures OpenAPI generation and Swagger UI for a router.
// It's a convenience function that handles the integration between
// the router, OpenAPI generator, and Swagger UI components.
//
// Example:
//
//	err := integration.Setup(router, integration.DefaultSetupOptions())
//	if err != nil {
//	    log.Fatal(err)
//	}
func Setup(r *router.Router, opts SetupOptions) error {
	// Create OpenAPI generator
	generator := openapi.NewGenerator(openapi.Info{
		Title:       opts.Title,
		Version:     opts.Version,
		Description: opts.Description,
	})

	// Add requested security schemes
	if opts.UseBasicAuth {
		generator.WithBasicAuth("basicAuth", "Basic authentication")
	}
	if opts.UseBearerAuth {
		generator.WithBearerAuth("bearerAuth", "Bearer token authentication")
	}
	if opts.UseAPIKey {
		generator.WithAPIKey("apiKey", "API key authentication", "header", "X-API-Key")
	}

	// Configure Swagger UI
	uiConfig := swagger.DefaultUIConfig()
	if opts.UITitle != "" {
		uiConfig.Title = opts.UITitle
	} else {
		uiConfig.Title = opts.Title
	}
	uiConfig.DarkMode = opts.DarkMode

	// Set up the integration
	swaggerUI := NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)

	// Set up routes with provided paths
	swaggerUI.SetupRoutes(r, opts.SpecPath, opts.DocsPath)

	return nil
}
