package docs

import (
	"reflect"

	"github.com/joakimcarlsson/go-router/metadata"
)

// RouteOption configures route metadata for API documentation
type RouteOption func(*metadata.RouteMetadata)

// WithOperationID sets the operationId for the route
func WithOperationID(operationId string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.OperationID = operationId
	}
}

// WithSummary sets the route summary
func WithSummary(summary string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Summary = summary
	}
}

// WithDescription sets the route description
func WithDescription(description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Description = description
	}
}

// WithTags adds tags to the route
func WithTags(tags ...string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Tags = append(m.Tags, tags...)
	}
}

// WithParameter adds a parameter to the route
func WithParameter(name, in, typ string, required bool, description string, example interface{}) RouteOption {
	return func(m *metadata.RouteMetadata) {
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

// WithHeaderParam adds a header parameter to the route
func WithHeaderParam(name string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "header", "string", required, description, example)
}

// WithRequestBody adds a request body with a specific content type
func WithRequestBody(contentType string, schema metadata.Schema, required bool, description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required:    required,
			Content: map[string]metadata.MediaType{
				contentType: {Schema: schema},
			},
		}
	}
}

// WithJSONRequestBody adds a JSON request body
func WithJSONRequestBody[T any](required bool, description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := SchemaFromType(t)

		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required:    required,
			Content: map[string]metadata.MediaType{
				"application/json": {Schema: schema},
			},
		}
	}
}

// WithResponse adds a response to the route
func WithResponse(statusCode int, description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		code := metadata.StatusCodeToString(statusCode)
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}
		m.Responses[code] = metadata.Response{
			Description: description,
		}
	}
}

// WithJSONResponse adds a JSON response with schema
func WithJSONResponse[T any](statusCode int, description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		t := reflect.TypeOf((*T)(nil)).Elem()
		schema := SchemaFromType(t)

		code := metadata.StatusCodeToString(statusCode)
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}
		m.Responses[code] = metadata.Response{
			Description: description,
			Content: map[string]metadata.MediaType{
				"application/json": {Schema: schema},
			},
		}
	}
}

// WithDeprecated marks a route as deprecated
func WithDeprecated(message string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Deprecated = true
		if message != "" {
			if m.Description != "" {
				m.Description += "\n\n"
			}
			m.Description += "DEPRECATED: " + message
		}
	}
}

// WithSecurity adds security requirements to a route
func WithSecurity(requirements ...map[string][]string) RouteOption {
	return func(m *metadata.RouteMetadata) {
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

// WithBasicAuth adds basic authentication requirement to a route
func WithBasicAuth() RouteOption {
	return WithSecurity(map[string][]string{"basicAuth": {}})
}

// WithBearerAuth adds bearer token authentication requirement to a route
func WithBearerAuth() RouteOption {
	return WithSecurity(map[string][]string{"bearerAuth": {}})
}

// WithAPIKey adds API key authentication requirement to a route
func WithAPIKey() RouteOption {
	return WithSecurity(map[string][]string{"apiKey": {}})
}

// WithOAuth2Scopes adds OAuth2 authentication requirement with specific scopes
func WithOAuth2Scopes(scopes ...string) RouteOption {
	return WithSecurity(map[string][]string{"oauth2": scopes})
}
