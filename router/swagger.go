package router

import (
	"html/template"
	"net/http"
)

// SwaggerUIConfig holds configuration options for serving Swagger UI
type SwaggerUIConfig struct {
	// Title is the page title for the Swagger UI page
	Title string
	// SpecURL is the URL to the OpenAPI specification JSON
	SpecURL string
	// SwaggerVersion is the version of Swagger UI to use from the CDN
	SwaggerVersion string
}

// DefaultSwaggerUIConfig returns a default configuration for Swagger UI
func DefaultSwaggerUIConfig() SwaggerUIConfig {
	return SwaggerUIConfig{
		Title:          "API Documentation",
		SpecURL:        "/openapi.json",
		SwaggerVersion: "5.17.5",
	}
}

// ServeSwaggerUI returns a handler for serving Swagger UI using CDN-hosted resources
func (r *Router) ServeSwaggerUI(config SwaggerUIConfig) HandlerFunc {
	const swaggerTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@{{.SwaggerVersion}}/swagger-ui.css" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>

  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@{{.SwaggerVersion}}/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@{{.SwaggerVersion}}/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "{{.SpecURL}}",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        validatorUrl: null
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`

	tmpl, err := template.New("swagger-ui").Parse(swaggerTemplate)
	if err != nil {
		panic(err)
	}

	return func(c *Context) {
		data := struct {
			Title          string
			SpecURL        string
			SwaggerVersion string
		}{
			Title:          config.Title,
			SpecURL:        config.SpecURL,
			SwaggerVersion: config.SwaggerVersion,
		}

		c.SetHeader("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		tmpl.Execute(c.Writer, data)
	}
}
