/*
Package integration provides adapters and utilities for integrating OpenAPI and Swagger UI
with the router. It offers both low-level adapters for fine-grained control and high-level
setup functions for quick configuration.

# Quick Setup

For basic usage, use the Setup function with default options:

	err := integration.Setup(router, integration.DefaultSetupOptions())
	if err != nil {
		log.Fatal(err)
	}

Custom configuration:

	err := integration.Setup(router, integration.SetupOptions{
		Title:        "My API",
		Version:      "1.0.0",
		Description:  "API documentation",
		SpecPath:     "/api-spec.json",
		DocsPath:     "/api-docs",
		DarkMode:     true,
		UseBearerAuth: true,
	})

# Advanced Usage

For more control, use the individual components:

	// Create OpenAPI adapter
	generator := openapi.NewGenerator(...)
	adapter := NewRouterOpenAPIAdapter(router, generator)

	// Configure Swagger UI
	swaggerUI := NewSwaggerUIIntegration(router, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(router, "/openapi.json", "/docs")

The integration package is designed to be optional - the router works without it,
and you can add API documentation features when needed.
*/
package integration
