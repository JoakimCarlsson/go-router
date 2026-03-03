package openapi

import (
	"encoding/json"
	"strings"

	"github.com/joakimcarlsson/go-router/router/v2"
)

// Generator handles OpenAPI 3.0 specification generation from router routes.
type Generator struct {
	info            Info
	securitySchemes map[string]SecurityScheme
	servers         []Server
	schemas         map[string]Schema
	routeInfo       []RouteInfo
}

// NewGenerator creates a new OpenAPI generator with the provided API information.
func NewGenerator(info Info) *Generator {
	return &Generator{
		info:            info,
		securitySchemes: make(map[string]SecurityScheme),
		servers:         make([]Server, 0),
		schemas:         make(map[string]Schema),
		routeInfo:       make([]RouteInfo, 0),
	}
}

// WithSecurityScheme adds a security scheme to the OpenAPI specification.
func (g *Generator) WithSecurityScheme(name string, scheme SecurityScheme) {
	g.securitySchemes[name] = scheme
}

// WithBasicAuth adds a basic authentication security scheme.
func (g *Generator) WithBasicAuth(name, description string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "http",
		Scheme:      "basic",
		Description: description,
	})
}

// WithBearerAuth adds a bearer token authentication security scheme.
func (g *Generator) WithBearerAuth(name, description string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: description,
	})
}

// WithAPIKey adds an API key authentication security scheme.
func (g *Generator) WithAPIKey(name, description, in, paramName string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "apiKey",
		Description: description,
		Name:        paramName,
		In:          in,
	})
}

// WithOAuth2ImplicitFlow adds an OAuth2 security scheme with implicit flow.
func (g *Generator) WithOAuth2ImplicitFlow(name, description, authorizationURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &OAuthFlows{
			Implicit: &OAuthFlow{
				AuthorizationURL: authorizationURL,
				Scopes:           scopes,
			},
		},
	})
}

// WithOAuth2PasswordFlow adds an OAuth2 security scheme with password flow.
func (g *Generator) WithOAuth2PasswordFlow(name, description, tokenURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &OAuthFlows{
			Password: &OAuthFlow{
				TokenURL: tokenURL,
				Scopes:   scopes,
			},
		},
	})
}

// WithOAuth2ClientCredentialsFlow adds an OAuth2 security scheme with client credentials flow.
func (g *Generator) WithOAuth2ClientCredentialsFlow(name, description, tokenURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &OAuthFlows{
			ClientCredentials: &OAuthFlow{
				TokenURL: tokenURL,
				Scopes:   scopes,
			},
		},
	})
}

// WithOAuth2AuthorizationCodeFlow adds an OAuth2 security scheme with authorization code flow.
func (g *Generator) WithOAuth2AuthorizationCodeFlow(name, description, authorizationURL, tokenURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &OAuthFlows{
			AuthorizationCode: &OAuthFlow{
				AuthorizationURL: authorizationURL,
				TokenURL:         tokenURL,
				Scopes:           scopes,
			},
		},
	})
}

// WithOpenIDConnect adds an OpenID Connect security scheme.
func (g *Generator) WithOpenIDConnect(name, description, openIDConnectURL string) {
	g.WithSecurityScheme(name, SecurityScheme{
		Type:             "openIdConnect",
		Description:      description,
		OpenIDConnectURL: openIDConnectURL,
	})
}

// WithServer adds a server to the OpenAPI specification.
func (g *Generator) WithServer(url string, description string) {
	g.servers = append(g.servers, Server{
		URL:         url,
		Description: description,
	})
}

func (g *Generator) collectSchemas() {
	for _, route := range g.routeInfo {
		if reqBody := route.RequestBody(); reqBody != nil {
			for _, mediaType := range reqBody.Content {
				g.collectSchemaComponents(mediaType.Schema)
			}
		}

		for _, response := range route.Responses() {
			if response.Content != nil {
				for _, mediaType := range response.Content {
					g.collectSchemaComponents(mediaType.Schema)
				}
			}
		}
	}
}

func (g *Generator) collectSchemaComponents(schema Schema) {
	if schema.Type == "array" && schema.Items != nil {
		if schema.Items.Type == "object" && schema.Items.Properties != nil && schema.Items.TypeName != "" {
			name := SanitizeSchemaName(schema.Items.TypeName)
			g.schemas[name] = *schema.Items
		}
		g.collectSchemaComponents(*schema.Items)
		return
	}

	if schema.Type == "object" && schema.Properties != nil && schema.TypeName != "" {
		name := g.generateSchemaName(schema)
		if name != "" {
			g.schemas[name] = schema
		}

		for _, prop := range schema.Properties {
			g.collectSchemaComponents(prop)
		}
	}
}

func (g *Generator) generateSchemaName(schema Schema) string {
	if schema.TypeName != "" {
		if strings.HasPrefix(schema.TypeName, "[]") {
			return SanitizeSchemaName(strings.TrimPrefix(schema.TypeName, "[]"))
		}
		return SanitizeSchemaName(schema.TypeName)
	}
	return ""
}

func (g *Generator) createSchemaReference(schemaName string) *Reference {
	return &Reference{
		Ref: "#/components/schemas/" + schemaName,
	}
}

// WithResponseSchema adds a response with content schema to the route.
func WithResponseSchema(statusCode int, description string, contentType string, schema Schema) router.RouteOption {
	return ResponseSchemaOption{
		StatusCode:  statusCode,
		Description: description,
		ContentType: contentType,
		Schema:      schema,
	}
}

// ResponseSchemaOption represents a response with a custom schema.
type ResponseSchemaOption struct {
	StatusCode  int
	Description string
	ContentType string
	Schema      Schema
}

// WithEmptyResponse adds a response without any content schema.
func WithEmptyResponse(statusCode int, description string) router.RouteOption {
	return ResponseOption{
		StatusCode:  statusCode,
		Description: description,
	}
}

// WithJSONResponseAdvanced adds a JSON response with schema inferred from the provided type T.
func WithJSONResponseAdvanced[T any](statusCode int, description string) router.RouteOption {
	return JSONResponseAdvancedOption{
		StatusCode:  statusCode,
		Description: description,
		Type:        GetTypeFromGeneric[T](),
	}
}

// JSONResponseAdvancedOption represents an advanced JSON response option.
type JSONResponseAdvancedOption struct {
	StatusCode  int
	Description string
	Type        interface{}
}

// WithResponseType adds a response with schema inferred from the provided type.
func WithResponseType[T any](statusCode int, description string, _ T) router.RouteOption {
	return WithJSONResponseAdvanced[T](statusCode, description)
}

// WithRequestBodyAdvanced adds a request body schema to the route.
func WithRequestBodyAdvanced[T any](description string, required bool, _ T) router.RouteOption {
	return JSONRequestBodyOption{
		Type:        GetTypeFromGeneric[T](),
		Required:    required,
		Description: description,
	}
}

// WithResponseExample adds a response with a specific example.
func WithResponseExample[T any](statusCode int, description string, example T) router.RouteOption {
	return ResponseExampleOption{
		StatusCode:  statusCode,
		Description: description,
		Type:        GetTypeFromGeneric[T](),
		Example:     example,
	}
}

// ResponseExampleOption represents a response with an example.
type ResponseExampleOption struct {
	StatusCode  int
	Description string
	Type        interface{}
	Example     interface{}
}

// WithRequestBodyExample adds a request body schema with example to the route.
func WithRequestBodyExample[T any](description string, required bool, example T) router.RouteOption {
	return RequestBodyExampleOption{
		Description: description,
		Required:    required,
		Type:        GetTypeFromGeneric[T](),
		Example:     example,
	}
}

// RequestBodyExampleOption represents a request body with an example.
type RequestBodyExampleOption struct {
	Description string
	Required    bool
	Type        interface{}
	Example     interface{}
}

// Generate creates an OpenAPI specification from the collected route information.
func (g *Generator) Generate(routes []RouteInfo) *Spec {
	g.routeInfo = routes
	g.collectSchemas()

	spec := &Spec{
		OpenAPI: "3.0.0",
		Info:    g.info,
		Paths:   make(map[string]PathItem),
		Components: &Components{
			SecuritySchemes: g.securitySchemes,
			Schemas:         g.schemas,
		},
	}

	if len(g.servers) > 0 {
		spec.Servers = g.servers
	}

	for _, route := range routes {
		pathItem, ok := spec.Paths[route.Path()]
		if !ok {
			pathItem = PathItem{}
		}

		var requestBody *RequestBody
		if rb := route.RequestBody(); rb != nil {
			requestBody = rb

			for contentType, mediaType := range requestBody.Content {
				schemaName := g.generateSchemaName(mediaType.Schema)
				if schemaName != "" && g.schemas[schemaName].Type != "" {
					mediaType.SchemaRef = g.createSchemaReference(schemaName)
					mediaType.Schema = Schema{}
					requestBody.Content[contentType] = mediaType
				}
			}
		}

		responses := make(map[string]Response)
		for statusCode, response := range route.Responses() {
			for contentType, mediaType := range response.Content {
				schemaName := g.generateSchemaName(mediaType.Schema)
				if schemaName != "" && g.schemas[schemaName].Type != "" {
					mediaType.SchemaRef = g.createSchemaReference(schemaName)
					mediaType.Schema = Schema{}
					response.Content[contentType] = mediaType
				} else if mediaType.Schema.Type == "array" && mediaType.Schema.Items != nil {
					itemSchemaName := g.generateSchemaName(*mediaType.Schema.Items)
					if itemSchemaName != "" && g.schemas[itemSchemaName].Type != "" {
						mediaType.Schema.Items.Ref = "#/components/schemas/" + itemSchemaName
						mediaType.Schema.Items.Type = ""
						mediaType.Schema.Items.Properties = nil
						mediaType.Schema.Items.Example = nil
						mediaType.Schema.Items.Required = nil
						response.Content[contentType] = mediaType
					}
				}
			}

			responses[statusCode] = response
		}

		if route.IsSSE() && len(route.SSEEvents()) > 0 {
			if resp, ok := responses["200"]; ok {
				if mediaType, ok := resp.Content[ContentTypeEventStream]; ok {
					mediaType.Examples = g.buildSSEExamples(route.SSEEvents())
					resp.Content[ContentTypeEventStream] = mediaType
					responses["200"] = resp
				}
			}
		}

		routeParams := route.Parameters()
		parameters := make([]Parameter, len(routeParams))
		copy(parameters, routeParams)

		security := make([]SecurityRequirement, len(route.Security()))
		for i, sec := range route.Security() {
			secReq := make(SecurityRequirement)
			for k, v := range sec {
				secReq[k] = v
			}
			security[i] = secReq
		}

		operation := &Operation{
			OperationID: route.OperationID(),
			Summary:     route.Summary(),
			Description: route.Description(),
			Tags:        route.Tags(),
			Parameters:  parameters,
			RequestBody: requestBody,
			Responses:   responses,
			Security:    security,
			Deprecated:  route.IsDeprecated(),
		}

		switch route.Method() {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		}

		spec.Paths[route.Path()] = pathItem
	}

	return spec
}

func (g *Generator) buildSSEExamples(events []SSEEventSchema) map[string]Example {
	if len(events) == 0 {
		return nil
	}

	var sb strings.Builder
	for i, event := range events {
		exampleData := event.Schema.Example
		if exampleData == nil {
			exampleData = map[string]interface{}{}
		}

		jsonBytes, _ := json.Marshal(exampleData)

		sb.WriteString("event: ")
		sb.WriteString(event.EventName)
		sb.WriteString("\ndata: ")
		sb.Write(jsonBytes)
		sb.WriteString("\n")

		if i < len(events)-1 {
			sb.WriteString("\n")
		}
	}

	return map[string]Example{
		"eventStream": {
			Summary: "Example event stream",
			Value:   sb.String(),
		},
	}
}
