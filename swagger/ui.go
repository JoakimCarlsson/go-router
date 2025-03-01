package swagger

import (
	"html/template"
	"net/http"

	"github.com/joakimcarlsson/go-router/metadata"
)

// UIConfig holds configuration options for serving Swagger UI
type UIConfig struct {
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
	// OAuth2Config contains OAuth2 configuration for Swagger UI
	OAuth2Config *metadata.OAuth2Config
}

// DefaultUIConfig returns a default configuration for Swagger UI
func DefaultUIConfig() UIConfig {
	return UIConfig{
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
		OAuth2Config:             nil,
	}
}

// Handler returns an http.HandlerFunc that serves the Swagger UI
func Handler(config UIConfig) http.HandlerFunc {
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
        {{if .OAuth2Config}},
        initOAuth: {
          clientId: "{{.OAuth2Config.ClientID}}",
          {{if .OAuth2Config.ClientSecret}}clientSecret: "{{.OAuth2Config.ClientSecret}}",{{end}}
          {{if .OAuth2Config.Realm}}realm: "{{.OAuth2Config.Realm}}",{{end}}
          {{if .OAuth2Config.AppName}}appName: "{{.OAuth2Config.AppName}}",{{end}}
          {{if .OAuth2Config.ScopeSeparator}}scopeSeparator: "{{.OAuth2Config.ScopeSeparator}}",{{end}}
          {{if .OAuth2Config.Scopes}}scopes: {{.OAuth2Config.Scopes}},{{end}}
          {{if .OAuth2Config.AdditionalQueryParams}}
          additionalQueryStringParams: {
            {{range $key, $value := .OAuth2Config.AdditionalQueryParams}}
            "{{$key}}": "{{$value}}"{{if not (last $key $.OAuth2Config.AdditionalQueryParams)}},{{end}}
            {{end}}
          },
          {{end}}
          usePkceWithAuthorizationCodeGrant: {{.OAuth2Config.UsePkceWithAuthorizationCodeGrant}},
          useBasicAuthenticationWithAccessCodeGrant: {{.OAuth2Config.UseBasicAuthenticationWithAccessCodeGrant}}
        }
        {{end}}
      });
      window.ui = ui;
      
      {{.CustomJS}}
    };
  </script>
</body>
</html>`

	tmpl, err := template.New("swagger-ui").Funcs(template.FuncMap{
		"last": func(key string, m map[string]string) bool {
			// Get all keys and find if this is the last one
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			return len(keys) > 0 && keys[len(keys)-1] == key
		},
	}).Parse(swaggerTemplate)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
			OAuth2Config             *metadata.OAuth2Config
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
			OAuth2Config:             config.OAuth2Config,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, data)
	}
}
