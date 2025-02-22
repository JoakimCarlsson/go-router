package openapi

import "reflect"

// Generator handles OpenAPI specification generation
type Generator struct {
	info            Info
	securitySchemes map[string]SecurityScheme
	servers         []Server
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(info Info) *Generator {
	return &Generator{
		info:            info,
		securitySchemes: make(map[string]SecurityScheme),
		servers:         make([]Server, 0),
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

// RouteMetadata contains OpenAPI documentation for a route
type RouteMetadata struct {
	Method      string                `json:"-"` // Used internally, not part of OpenAPI spec
	Path        string                `json:"-"` // Used internally, not part of OpenAPI spec
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
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
func WithResponseType[T any](statusCode, description string, _ T) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]Response)
		}

		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := SchemaFromType(t)

		m.Responses[statusCode] = Response{
			Description: description,
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	}
}

// WithArrayResponseType adds an array response with item schema inferred from the provided type
func WithArrayResponseType[T any](statusCode, description string, _ T) RouteOption {
	return func(m *RouteMetadata) {
		if m.Responses == nil {
			m.Responses = make(map[string]Response)
		}

		t := reflect.TypeOf((*T)(nil)).Elem()
		itemSchema := SchemaFromType(t)

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
	spec := &Spec{
		OpenAPI: "3.0.0",
		Info:    g.info,
		Paths:   make(map[string]PathItem),
		Components: &Components{
			SecuritySchemes: g.securitySchemes,
		},
	}

	if len(g.servers) > 0 {
		spec.Servers = g.servers
	}

	for _, route := range routes {
		pathItem, ok := spec.Paths[route.Path]
		if !ok {
			pathItem = PathItem{}
		}

		operation := &Operation{
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

	delete(spec.Paths, "/swagger.json")

	return spec
}
