/*
Package swaggerui provides Swagger UI serving and integration with the go-router
and openapi packages.

# Quick Start

	import (
		"net/http"

		"github.com/joakimcarlsson/go-router/openapi"
		"github.com/joakimcarlsson/go-router/router"
		"github.com/joakimcarlsson/go-router/swaggerui"
	)

	func main() {
		r := router.New()

		// Create OpenAPI generator
		generator := openapi.NewGenerator(openapi.Info{
			Title:       "My API",
			Version:     "1.0.0",
			Description: "API documentation",
		})

		// Add your routes with documentation
		r.GET("/users", listUsers,
			openapi.WithSummary("List users"),
			openapi.WithJSONResponse[[]User](200, "List of users"),
		)

		// Setup Swagger UI
		setup := swaggerui.NewSetup(r, generator)
		setup.RegisterRoutes(r, "/openapi.json", "/docs")

		http.ListenAndServe(":8080", r)
	}

# UI Configuration

Customize the Swagger UI appearance and behavior:

	config := swaggerui.DefaultUIConfig()
	config.Title = "My API Documentation"
	config.DarkMode = true
	config.TryItOutEnabled = true
	config.DocExpansion = "list"  // "list", "full", or "none"
	config.DefaultModelRendering = "example"  // "example" or "model"

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(config)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

# Configuration Options

UIConfig provides the following options:

	Title                    string  // Page title
	SpecURL                  string  // URL to OpenAPI spec (set automatically)
	SwaggerVersion           string  // Swagger UI version from CDN
	DarkMode                 bool    // Enable dark theme
	PersistAuthorization     bool    // Remember auth between sessions
	DefaultModelsExpandDepth int     // Model expansion depth
	DeepLinking              bool    // Enable deep linking
	DocExpansion             string  // Default expansion: "list", "full", "none"
	Filter                   bool    // Enable filtering
	DisplayRequestDuration   bool    // Show request timing
	TryItOutEnabled          bool    // Enable "Try it out" by default
	RequestSnippetsEnabled   bool    // Show request snippets
	DefaultModelRendering    string  // "example" or "model"
	CustomCSS                string  // Additional CSS
	CustomJS                 string  // Additional JavaScript

# OAuth2 Configuration

Configure OAuth2 for the Swagger UI:

	config := swaggerui.DefaultUIConfig()
	config.OAuth2Config = &swaggerui.OAuth2Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",  // Optional
		AppName:      "My App",
		Scopes:       []string{"read", "write"},
		UsePkceWithAuthorizationCodeGrant: true,
	}

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(config)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

# Standalone Handlers

Use the handlers directly without the Setup helper:

	// Serve Swagger UI at a custom path
	http.Handle("/swagger/", http.StripPrefix("/swagger",
		http.HandlerFunc(swaggerui.Handler(swaggerui.DefaultUIConfig()))))

	// OAuth2 redirect handler
	http.HandleFunc("/oauth2-redirect.html", swaggerui.OAuth2RedirectHandler())

# Security Schemes

The Swagger UI automatically picks up security schemes defined in the OpenAPI generator:

	generator := openapi.NewGenerator(openapi.Info{
		Title:   "Secure API",
		Version: "1.0.0",
	})

	// Add security schemes
	generator.WithBearerAuth("bearerAuth", "JWT token")
	generator.WithOAuth2AuthorizationCodeFlow(
		"oauth2",
		"OAuth2 authentication",
		"https://auth.example.com/authorize",
		"https://auth.example.com/token",
		map[string]string{
			"read":  "Read access",
			"write": "Write access",
		},
	)

	// Routes using security
	r.GET("/protected", handler,
		openapi.WithBearerAuth(),
	)

For route documentation, see the openapi package.
For core routing, see the router package.

For more information: https://github.com/JoakimCarlsson/go-router
*/
package swaggerui
