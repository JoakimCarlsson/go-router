package router

import (
	"github.com/joakimcarlsson/go-router/metadata"
)

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

// WithJSONResponse adds a JSON response with schema to the route
func WithJSONResponse(statusCode int, description string, schema metadata.Schema) RouteOption {
	return func(m *metadata.RouteMetadata) {
		code := metadata.StatusCodeToString(statusCode)
		if m.Responses == nil {
			m.Responses = make(map[string]metadata.Response)
		}
		m.Responses[code] = metadata.Response{
			Description: description,
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
