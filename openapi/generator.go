package openapi

import (
	"reflect"
	"strings"
)

// Generator handles OpenAPI specification generation
type Generator struct {
	info            Info
	securitySchemes map[string]SecurityScheme
	servers         []Server
	schemas         map[string]Schema
	routeMetadata   []RouteMetadata // Track routes for schema collection
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(info Info) *Generator {
	return &Generator{
		info:            info,
		securitySchemes: make(map[string]SecurityScheme),
		servers:         make([]Server, 0),
		schemas:         make(map[string]Schema),
		routeMetadata:   make([]RouteMetadata, 0),
	}
}

// WithSecurityScheme adds a security scheme to the OpenAPI specification
func (g *Generator) WithSecurityScheme(name string, scheme SecurityScheme) {
	g.securitySchemes[name] = scheme
}

// WithServer adds a server to the OpenAPI specification
func (g *Generator) WithServer(url string, description string) {
	g.servers = append(g.servers, Server{
		URL:         url,
		Description: description,
	})
}

// collectSchemas recursively collects schemas from route metadata
func (g *Generator) collectSchemas() {
	for _, route := range g.routeMetadata {
		// Collect from request bodies
		if route.RequestBody != nil {
			for _, mediaType := range route.RequestBody.Content {
				g.collectSchemaComponents(mediaType.Schema)
			}
		}

		// Collect from responses
		for _, response := range route.Responses {
			if response.Content != nil {
				for _, mediaType := range response.Content {
					g.collectSchemaComponents(mediaType.Schema)
				}
			}
		}
	}
}

// collectSchemaComponents recursively collects component schemas
func (g *Generator) collectSchemaComponents(schema Schema) {
	// If it's a struct type, register it as a component
	if schema.Type == "object" && schema.Properties != nil {
		name := g.generateSchemaName(schema)
		if name != "" {
			g.schemas[name] = schema
		}

		// Recurse into properties
		for _, prop := range schema.Properties {
			g.collectSchemaComponents(prop)
		}
	}

	// Recurse into array items
	if schema.Items != nil {
		g.collectSchemaComponents(*schema.Items)
	}
}

// generateSchemaName generates a name for a schema based on its structure
func (g *Generator) generateSchemaName(schema Schema) string {
	if schema.TypeName != "" {
		// For arrays, we only want the element type name
		if strings.HasPrefix(schema.TypeName, "[]") {
			return strings.TrimPrefix(schema.TypeName, "[]")
		}
		return schema.TypeName
	}
	return ""
}

// convertSchemaToRef converts a schema to a reference if it exists in components
func (g *Generator) convertSchemaToRef(schema Schema) Schema {
	if schema.Type == "object" && schema.Properties != nil {
		name := g.generateSchemaName(schema)
		if name != "" && g.schemas[name].Properties != nil {
			return Schema{
				Ref: "#/components/schemas/" + name,
			}
		}
	}
	if schema.Items != nil {
		if schema.Type == "array" {
			itemsRef := g.convertSchemaToRef(*schema.Items)
			schema.Items = &itemsRef
		}
	}
	return schema
}

// RouteMetadata contains OpenAPI documentation for a route
type RouteMetadata struct {
	Method      string                `json:"-"`
	Path        string                `json:"-"`
	OperationID string                `json:"operationId,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
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
func WithParameter(name, in, typ string, required bool, description string) RouteOption {
	return func(m *RouteMetadata) {
		m.Parameters = append(m.Parameters, Parameter{
			Name:        name,
			In:          in,
			Required:    required,
			Description: description,
			Schema: Schema{
				Type: typ,
			},
		})
	}
}

// WithQueryParam adds a query parameter to the route
func WithQueryParam(name, typ string, required bool, description string) RouteOption {
	return WithParameter(name, "query", typ, required, description)
}

// WithPathParam adds a path parameter to the route
func WithPathParam(name, typ string, required bool, description string) RouteOption {
	return WithParameter(name, "path", typ, required, description)
}

// WithResponse adds a response to the route
func WithResponse(statusCode, description string, contentType string, schema Schema) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]Response)
		}
		m.Responses[statusCode] = Response{
			Description: description,
			Content: map[string]MediaType{
				contentType: {Schema: schema},
			},
		}
	}
}

// WithEmptyResponse adds a response without any content schema
func WithEmptyResponse(statusCode, description string) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]Response)
		}
		m.Responses[statusCode] = Response{
			Description: description,
		}
	}
}

// WithResponseType adds a response with schema inferred from the provided type
// It automatically detects if the type is a slice/array
func WithResponseType[T any](statusCode, description string, _ T) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]Response)
		}

		t := reflect.TypeOf((*T)(nil)).Elem()

		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			itemSchema := SchemaFromType(t.Elem())
			m.Responses[statusCode] = Response{
				Description: description,
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Type:  "array",
							Items: &itemSchema,
						},
					},
				},
			}
		} else {
			schema := SchemaFromType(t)
			m.Responses[statusCode] = Response{
				Description: description,
				Content: map[string]MediaType{
					"application/json": {Schema: schema},
				},
			}
		}
	}
}

// WithRequestBody adds a request body schema to the route
func WithRequestBody[T any](description string, required bool, _ T) RouteOption {
	return func(m *RouteMetadata) {
		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := SchemaFromType(t)

		m.RequestBody = &RequestBody{
			Description: description,
			Required:    required,
			Content: map[string]MediaType{
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
			m.Security = make([]SecurityRequirement, 0)
		}
		for _, req := range requirements {
			secReq := make(SecurityRequirement)
			for k, v := range req {
				secReq[k] = v
			}
			m.Security = append(m.Security, secReq)
		}
	}
}

// Generate creates an OpenAPI specification from the collected route metadata
func (g *Generator) Generate(routes []RouteMetadata) *Spec {
	g.routeMetadata = routes
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
		pathItem, ok := spec.Paths[route.Path]
		if (!ok) {
			pathItem = PathItem{}
		}

		// Convert request body schema to ref if possible
		if route.RequestBody != nil {
			for contentType, mediaType := range route.RequestBody.Content {
				refSchema := g.convertSchemaToRef(mediaType.Schema)
				route.RequestBody.Content[contentType] = MediaType{Schema: refSchema}
			}
		}

		// Convert response schemas to refs if possible
		for statusCode, response := range route.Responses {
			if response.Content != nil {
				for contentType, mediaType := range response.Content {
					refSchema := g.convertSchemaToRef(mediaType.Schema)
					response.Content[contentType] = MediaType{Schema: refSchema}
				}
				route.Responses[statusCode] = response
			}
		}

		operation := &Operation{
			OperationID: route.OperationID,
			Summary:     route.Summary,
			Description: route.Description,
			Tags:        route.Tags,
			Parameters:  route.Parameters,
			RequestBody: route.RequestBody,
			Responses:   route.Responses,
			Security:    route.Security,
		}

		switch route.Method {
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

		spec.Paths[route.Path] = pathItem
	}

	delete(spec.Paths, "/openapi.json")

	return spec
}
