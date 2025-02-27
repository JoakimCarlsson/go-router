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
	// DarkMode enables dark mode UI theme when true
	DarkMode bool
	// PersistAuthorization preserves the authorization data between browser sessions
	PersistAuthorization bool
	// DefaultModelsExpandDepth sets the default expansion depth for models
	DefaultModelsExpandDepth int
	// DeepLinking enables deeplinking for tags and operations
	DeepLinking bool
	// DocExpansion controls the default expansion setting for the operations
	// Allowed values are: "list" (expands only tags), "full" (expands everything), "none" (expands nothing)
	DocExpansion string
	// Filter enables filtering of operations
	Filter bool
	// AdditionalQueryParams allows adding query params to the OpenAPI spec URL
	AdditionalQueryParams map[string]string
	// DisplayRequestDuration shows how long API calls take
	DisplayRequestDuration bool
	// MaxDisplayedTags limits the number of tagged operations shown
	MaxDisplayedTags int
	// ShowExtensions displays vendor extensions
	ShowExtensions bool
	// TryItOutEnabled enables the "Try it out" feature by default
	TryItOutEnabled bool
	// RequestSnippetsEnabled enables the request snippets section
	RequestSnippetsEnabled bool
	// DefaultModelRendering controls how models are displayed
	// Possible values: "example" or "model"
	DefaultModelRendering string
	// CustomCSS allows injecting additional CSS styles
	CustomCSS string
	// CustomJS allows injecting custom JavaScript
	CustomJS string
}

// DefaultSwaggerUIConfig returns a default configuration for Swagger UI
func DefaultSwaggerUIConfig() SwaggerUIConfig {
	return SwaggerUIConfig{
		Title:                    "API Documentation",
		SpecURL:                  "/openapi.json",
		SwaggerVersion:           "5.20.0",
		DarkMode:                 false,
		PersistAuthorization:     true,
		DefaultModelsExpandDepth: 1,
		DeepLinking:              true,
		DocExpansion:             "list",
		Filter:                   false,
		AdditionalQueryParams:    make(map[string]string),
		DisplayRequestDuration:   true,
		MaxDisplayedTags:         0,
		ShowExtensions:           false,
		TryItOutEnabled:          false,
		RequestSnippetsEnabled:   true,
		DefaultModelRendering:    "model",
		CustomCSS:                "",
		CustomJS:                 "",
	}
}

// ServeSwaggerUI returns a handler for serving Swagger UI using CDN-hosted resources
func (r *Router) ServeSwaggerUI(config SwaggerUIConfig) HandlerFunc {
	const swaggerTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@{{.SwaggerVersion}}/swagger-ui.css" />
  {{if .DarkMode}}
  <!-- Using jsDelivr CDN to serve the SwaggerDark CSS with proper MIME type -->
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/Amoenus/SwaggerDark@master/SwaggerDark.css" />
  {{end}}
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: {{if .DarkMode}}#1a1a1a{{else}}#fafafa{{end}}; }
    .topbar { display: none; }
    {{.CustomCSS}}
  </style>
</head>
<body>
  <div id="swagger-ui"></div>

  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@{{.SwaggerVersion}}/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@{{.SwaggerVersion}}/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      // Build the URL with additional query parameters if provided
      let specUrl = "{{.SpecURL}}";
      {{ if .AdditionalQueryParams }}
      if (specUrl.includes("?")) {
        specUrl += "&";
      } else {
        specUrl += "?";
      }
      {{ range $key, $value := .AdditionalQueryParams }}
      specUrl += "{{ $key }}={{ $value }}&";
      {{ end }}
      // Remove trailing & if present
      if (specUrl.endsWith("&")) {
        specUrl = specUrl.slice(0, -1);
      }
      {{ end }}

      const ui = SwaggerUIBundle({
        url: specUrl,
        dom_id: '#swagger-ui',
        deepLinking: {{.DeepLinking}},
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        defaultModelsExpandDepth: {{.DefaultModelsExpandDepth}},
        displayRequestDuration: {{.DisplayRequestDuration}},
        docExpansion: "{{.DocExpansion}}",
        filter: {{.Filter}},
        persistAuthorization: {{.PersistAuthorization}},
        syntaxHighlight: {
          activate: true,
          theme: "{{if .DarkMode}}agate{{else}}default{{end}}"
        },
        {{if gt .MaxDisplayedTags 0}}
        maxDisplayedTags: {{.MaxDisplayedTags}},
        {{end}}
        showExtensions: {{.ShowExtensions}},
        tryItOutEnabled: {{.TryItOutEnabled}},
        requestSnippetsEnabled: {{.RequestSnippetsEnabled}},
        defaultModelRendering: "{{.DefaultModelRendering}}"
      });
      window.ui = ui;
      
      {{.CustomJS}}
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
			Title                    string
			SpecURL                  string
			SwaggerVersion           string
			DarkMode                 bool
			PersistAuthorization     bool
			DefaultModelsExpandDepth int
			DeepLinking              bool
			DocExpansion             string
			Filter                   bool
			AdditionalQueryParams    map[string]string
			DisplayRequestDuration   bool
			MaxDisplayedTags         int
			ShowExtensions           bool
			TryItOutEnabled          bool
			RequestSnippetsEnabled   bool
			DefaultModelRendering    string
			CustomCSS                string
			CustomJS                 string
		}{
			Title:                    config.Title,
			SpecURL:                  config.SpecURL,
			SwaggerVersion:           config.SwaggerVersion,
			DarkMode:                 config.DarkMode,
			PersistAuthorization:     config.PersistAuthorization,
			DefaultModelsExpandDepth: config.DefaultModelsExpandDepth,
			DeepLinking:              config.DeepLinking,
			DocExpansion:             config.DocExpansion,
			Filter:                   config.Filter,
			AdditionalQueryParams:    config.AdditionalQueryParams,
			DisplayRequestDuration:   config.DisplayRequestDuration,
			MaxDisplayedTags:         config.MaxDisplayedTags,
			ShowExtensions:           config.ShowExtensions,
			TryItOutEnabled:          config.TryItOutEnabled,
			RequestSnippetsEnabled:   config.RequestSnippetsEnabled,
			DefaultModelRendering:    config.DefaultModelRendering,
			CustomCSS:                config.CustomCSS,
			CustomJS:                 config.CustomJS,
		}

		c.SetHeader("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		tmpl.Execute(c.Writer, data)
	}
}
