package openapi

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/metadata"
)

// Generator handles OpenAPI specification generation
type Generator struct {
	info            metadata.Info
	securitySchemes map[string]metadata.SecurityScheme
	servers         []metadata.Server
	schemas         map[string]metadata.Schema
	routeInfo       []RouteInfo
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(info metadata.Info) *Generator {
	return &Generator{
		info:            info,
		securitySchemes: make(map[string]metadata.SecurityScheme),
		servers:         make([]metadata.Server, 0),
		schemas:         make(map[string]metadata.Schema),
		routeInfo:       make([]RouteInfo, 0),
	}
}

// WithSecurityScheme adds a security scheme to the OpenAPI specification
func (g *Generator) WithSecurityScheme(name string, scheme metadata.SecurityScheme) {
	g.securitySchemes[name] = scheme
}

// WithBasicAuth adds a basic authentication security scheme
func (g *Generator) WithBasicAuth(name, description string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "http",
		Scheme:      "basic",
		Description: description,
	})
}

// WithBearerAuth adds a bearer token authentication security scheme
func (g *Generator) WithBearerAuth(name, description string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: description,
	})
}

// WithAPIKey adds an API key authentication security scheme
func (g *Generator) WithAPIKey(name, description, in, paramName string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "apiKey",
		Description: description,
		Name:        paramName,
		In:          in,
	})
}

// WithOAuth2ImplicitFlow adds an OAuth2 security scheme with implicit flow
func (g *Generator) WithOAuth2ImplicitFlow(name, description, authorizationURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &metadata.OAuthFlows{
			Implicit: &metadata.OAuthFlow{
				AuthorizationURL: authorizationURL,
				Scopes:           scopes,
			},
		},
	})
}

// WithOAuth2PasswordFlow adds an OAuth2 security scheme with password flow
func (g *Generator) WithOAuth2PasswordFlow(name, description, tokenURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &metadata.OAuthFlows{
			Password: &metadata.OAuthFlow{
				TokenURL: tokenURL,
				Scopes:   scopes,
			},
		},
	})
}

// WithOAuth2ClientCredentialsFlow adds an OAuth2 security scheme with client credentials flow
func (g *Generator) WithOAuth2ClientCredentialsFlow(name, description, tokenURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &metadata.OAuthFlows{
			ClientCredentials: &metadata.OAuthFlow{
				TokenURL: tokenURL,
				Scopes:   scopes,
			},
		},
	})
}

// WithOAuth2AuthorizationCodeFlow adds an OAuth2 security scheme with authorization code flow
func (g *Generator) WithOAuth2AuthorizationCodeFlow(name, description, authorizationURL, tokenURL string, scopes map[string]string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:        "oauth2",
		Description: description,
		Flows: &metadata.OAuthFlows{
			AuthorizationCode: &metadata.OAuthFlow{
				AuthorizationURL: authorizationURL,
				TokenURL:         tokenURL,
				Scopes:           scopes,
			},
		},
	})
}

// WithOpenIDConnect adds an OpenID Connect security scheme
func (g *Generator) WithOpenIDConnect(name, description, openIDConnectURL string) {
	g.WithSecurityScheme(name, metadata.SecurityScheme{
		Type:             "openIdConnect",
		Description:      description,
		OpenIDConnectURL: openIDConnectURL,
	})
}

// WithServer adds a server to the OpenAPI specification
func (g *Generator) WithServer(url string, description string) {
	g.servers = append(g.servers, metadata.Server{
		URL:         url,
		Description: description,
	})
}

// collectSchemas recursively collects schemas from route info
func (g *Generator) collectSchemas() {
	for _, route := range g.routeInfo {
		// Collect from request bodies
		if reqBody := route.RequestBody(); reqBody != nil {
			for _, mediaType := range reqBody.Content {
				g.collectSchemaComponents(mediaType.Schema)
			}
		}

		// Collect from responses
		for _, response := range route.Responses() {
			if response.Content != nil {
				for _, mediaType := range response.Content {
					g.collectSchemaComponents(mediaType.Schema)
				}
			}
		}
	}
}

// collectSchemaComponents recursively collects component schemas
func (g *Generator) collectSchemaComponents(schema metadata.Schema) {
	// If it's an array type, process the item type
	if schema.Type == "array" && schema.Items != nil {
		// Register the array item type if it's an object
		if schema.Items.Type == "object" && schema.Items.Properties != nil && schema.Items.TypeName != "" {
			name := metadata.SanitizeSchemaName(schema.Items.TypeName)
			g.schemas[name] = *schema.Items
		}

		// Continue processing the items schema
		g.collectSchemaComponents(*schema.Items)
		return
	}

	// If it's a struct type, register it as a component
	if schema.Type == "object" && schema.Properties != nil && schema.TypeName != "" {
		name := g.generateSchemaName(schema)
		if name != "" {
			g.schemas[name] = schema
		}

		// Recurse into properties
		for _, prop := range schema.Properties {
			g.collectSchemaComponents(prop)
		}
	}
}

// generateSchemaName generates a name for a schema based on its structure
func (g *Generator) generateSchemaName(schema metadata.Schema) string {
	if schema.TypeName != "" {
		// For arrays, we only want the element type name
		if strings.HasPrefix(schema.TypeName, "[]") {
			return metadata.SanitizeSchemaName(strings.TrimPrefix(schema.TypeName, "[]"))
		}
		return metadata.SanitizeSchemaName(schema.TypeName)
	}
	return ""
}

// createSchemaReference creates a reference to a schema component
func (g *Generator) createSchemaReference(schemaName string) *metadata.Reference {
	return &metadata.Reference{
		Ref: "#/components/schemas/" + schemaName,
	}
}

// RouteMetadata contains OpenAPI documentation for a route
// This structure remains for backward compatibility
type RouteMetadata struct {
	Method      string                         `json:"-"`
	Path        string                         `json:"-"`
	OperationID string                         `json:"operationId,omitempty"`
	Summary     string                         `json:"summary,omitempty"`
	Description string                         `json:"description,omitempty"`
	Tags        []string                       `json:"tags,omitempty"`
	Parameters  []metadata.Parameter           `json:"parameters,omitempty"`
	RequestBody *metadata.RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]metadata.Response   `json:"responses"`
	Security    []metadata.SecurityRequirement `json:"security,omitempty"`
	Deprecated  bool                           `json:"deprecated,omitempty"`
}

// WithOperationID sets the operationId for the route
func WithOperationID(operationId string) RouteOption {
	return func(m *RouteMetadata) {
		m.OperationID = operationId
	}
}

// RouteOption is a function that configures route metadata
type RouteOption func(*RouteMetadata)

// WithSummary sets the route summary
func WithSummary(summary string) RouteOption {
	return func(m *RouteMetadata) {
		m.Summary = summary
	}
}

// WithDescription sets the route description
func WithDescription(description string) RouteOption {
	return func(m *RouteMetadata) {
		m.Description = description
	}
}

// WithTags adds tags to the route
func WithTags(tags ...string) RouteOption {
	return func(m *RouteMetadata) {
		m.Tags = append(m.Tags, tags...)
	}
}

// WithParameter adds a parameter to the route
func WithParameter(name, in, typ string, required bool, description string, example interface{}) RouteOption {
	return func(m *RouteMetadata) {
		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          in,
			Required:    required,
			Description: description,
			Schema: metadata.Schema{
				Type:    typ,
				Example: example,
			},
		})
	}
}

// WithQueryParam adds a query parameter to the route
func WithQueryParam(name, typ string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "query", typ, required, description, example)
}

// WithPathParam adds a path parameter to the route
func WithPathParam(name, typ string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "path", typ, required, description, example)
}

// WithResponse adds a response to the route
func WithResponse(statusCode int, description string, contentType string, schema metadata.Schema) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}
		m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
			Description: description,
			Content: map[string]metadata.MediaType{
				contentType: {Schema: schema},
			},
		}
	}
}

// WithEmptyResponse adds a response without any content schema
func WithEmptyResponse(statusCode int, description string) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}
		m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
			Description: description,
		}
	}
}

// WithJSONResponse adds a JSON response with schema inferred from the provided type T
// It automatically handles both array and non-array types
func WithJSONResponse[T any](statusCode int, description string) RouteOption {
	return func(m *RouteMetadata) {
		t := reflect.TypeOf((*T)(nil)).Elem()

		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}

		// Special handling for array types
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			elemType := t.Elem()
			// Register the element type to ensure it appears in components
			itemTypeName := metadata.RegisterType(elemType)
			sanitizedName := metadata.SanitizeSchemaName(itemTypeName)

			m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
				Description: description,
				Content: map[string]metadata.MediaType{
					"application/json": {
						Schema: metadata.Schema{
							Type: "array",
							Items: &metadata.Schema{
								Ref: "#/components/schemas/" + sanitizedName,
							},
						},
					},
				},
			}
			return
		}

		// For non-array types
		schema := docs.SchemaFromType(t)
		if schema.Type == "object" && schema.Properties != nil && schema.TypeName != "" {
			// Use reference for object types
			m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
				Description: description,
				Content: map[string]metadata.MediaType{
					"application/json": {
						SchemaRef: &metadata.Reference{
							Ref: "#/components/schemas/" + metadata.SanitizeSchemaName(schema.TypeName),
						},
					},
				},
			}
		} else {
			// Use inline schema for primitive types
			m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
				Description: description,
				Content: map[string]metadata.MediaType{
					"application/json": {
						Schema: schema,
					},
				},
			}
		}
	}
}

// WithResponseType adds a response with schema inferred from the provided type
// It automatically detects if the type is a slice/array
func WithResponseType[T any](statusCode int, description string, _ T) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}

		t := reflect.TypeOf((*T)(nil)).Elem()

		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			elemType := t.Elem()
			itemSchema := docs.SchemaFromType(elemType)

			if itemSchema.Type == "object" && itemSchema.Properties != nil && itemSchema.TypeName != "" {
				m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
					Description: description,
					Content: map[string]metadata.MediaType{
						"application/json": {
							Schema: metadata.Schema{
								Type: "array",
								Items: &metadata.Schema{
									Ref: "#/components/schemas/" + itemSchema.TypeName,
								},
							},
						},
					},
				}
			} else {
				// For primitive type arrays, use the schema directly
				m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
					Description: description,
					Content: map[string]metadata.MediaType{
						"application/json": {
							Schema: metadata.Schema{
								Type:  "array",
								Items: &itemSchema,
							},
						},
					},
				}
			}
		} else {
			schema := docs.SchemaFromType(t)
			schemaName := schema.TypeName

			if schema.Type == "object" && schema.Properties != nil && schemaName != "" {
				m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
					Description: description,
					Content: map[string]metadata.MediaType{
						"application/json": {
							SchemaRef: &metadata.Reference{
								Ref: "#/components/schemas/" + schemaName,
							},
						},
					},
				}
			} else {
				// For primitive types, use the schema directly
				m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
					Description: description,
					Content: map[string]metadata.MediaType{
						"application/json": {Schema: schema},
					},
				}
			}
		}
	}
}

// WithRequestBody adds a request body schema to the route
func WithRequestBody[T any](description string, required bool, _ T) RouteOption {
	return func(m *RouteMetadata) {
		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := docs.SchemaFromType(t)

		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required:    required,
			Content: map[string]metadata.MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}
	}
}

// WithSecurity adds security requirements to a route
func WithSecurity(requirements ...map[string][]string) RouteOption {
	return func(m *RouteMetadata) {
		if m.Security == nil {
			m.Security = make([]metadata.SecurityRequirement, 0)
		}
		for _, req := range requirements {
			secReq := make(metadata.SecurityRequirement)
			for k, v := range req {
				secReq[k] = v
			}
			m.Security = append(m.Security, secReq)
		}
	}
}

// WithDeprecated marks an endpoint as deprecated
func WithDeprecated(message string) RouteOption {
	return func(m *RouteMetadata) {
		m.Deprecated = true
		if message != "" {
			if m.Description != "" {
				m.Description += "\n\n"
			}
			m.Description += "DEPRECATED: " + message
		}
	}
}

// WithResponseExample adds a response with a specific example
func WithResponseExample[T any](statusCode int, description string, example T) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}

		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := docs.SchemaFromType(t)
		schema.Example = example

		m.Responses[strconv.Itoa(statusCode)] = metadata.Response{
			Description: description,
			Content: map[string]metadata.MediaType{
				"application/json": {Schema: schema},
			},
		}
	}
}

// WithRequestBodyExample adds a request body schema with example to the route
func WithRequestBodyExample[T any](description string, required bool, example T) RouteOption {
	return func(m *RouteMetadata) {
		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := docs.SchemaFromType(t)
		schema.Example = example

		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required:    required,
			Content: map[string]metadata.MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}
	}
}

// Generate creates an OpenAPI specification from the collected route information
func (g *Generator) Generate(routes []RouteInfo) *metadata.Spec {
	g.routeInfo = routes
	g.collectSchemas()

	spec := &metadata.Spec{
		OpenAPI: "3.0.0",
		Info:    g.info,
		Paths:   make(map[string]metadata.PathItem),
		Components: &metadata.Components{
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
			pathItem = metadata.PathItem{}
		}

		var requestBody *metadata.RequestBody
		if rb := route.RequestBody(); rb != nil {
			requestBody = rb

			for contentType, mediaType := range requestBody.Content {
				schemaName := g.generateSchemaName(mediaType.Schema)
				if schemaName != "" && g.schemas[schemaName].Type != "" {
					mediaType.SchemaRef = g.createSchemaReference(schemaName)
					mediaType.Schema = metadata.Schema{}
					requestBody.Content[contentType] = mediaType
				}
			}
		}

		// Convert responses
		responses := make(map[string]metadata.Response)
		for statusCode, response := range route.Responses() {
			for contentType, mediaType := range response.Content {
				schemaName := g.generateSchemaName(mediaType.Schema)
				if schemaName != "" && g.schemas[schemaName].Type != "" {
					mediaType.SchemaRef = g.createSchemaReference(schemaName)
					mediaType.Schema = metadata.Schema{}
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

		// Properly copy parameters from the route
		routeParams := route.Parameters()
		parameters := make([]metadata.Parameter, len(routeParams))
		copy(parameters, routeParams)

		// Convert security requirements
		security := make([]metadata.SecurityRequirement, len(route.Security()))
		for i, sec := range route.Security() {
			secReq := make(metadata.SecurityRequirement)
			for k, v := range sec {
				secReq[k] = v
			}
			security[i] = secReq
		}

		operation := &metadata.Operation{
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

	delete(spec.Paths, "/openapi.json")

	return spec
}
