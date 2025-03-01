package router

import (
	"github.com/joakimcarlsson/go-router/metadata"
)

// WithOperationID sets the operationId for the route.
// OperationIDs should be unique across all operations in the API.
// They are used as function names in generated client libraries.
func WithOperationID(operationId string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.OperationID = operationId
	}
}

// WithSummary sets the route summary.
// The summary is a short description of the operation and is displayed
// in the OpenAPI documentation.
func WithSummary(summary string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Summary = summary
	}
}

// WithDescription sets the route description.
// The description provides more detailed information about the operation
// and is displayed in the OpenAPI documentation.
func WithDescription(description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Description = description
	}
}

// WithTags adds tags to the route.
// Tags are used to group operations in the OpenAPI documentation.
func WithTags(tags ...string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Tags = append(m.Tags, tags...)
	}
}

// WithParameter adds a parameter to the route.
// Parameters can be path parameters, query parameters, header parameters, etc.
//
// Parameters:
//   - name: The name of the parameter
//   - in: The location of the parameter (path, query, header, cookie)
//   - typ: The data type of the parameter (string, integer, boolean, etc.)
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
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

// WithQueryParam adds a query parameter to the route.
// Query parameters are appended to the URL after a question mark.
//
// Parameters:
//   - name: The name of the parameter
//   - typ: The data type of the parameter (string, integer, boolean, etc.)
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithQueryParam(name, typ string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "query", typ, required, description, example)
}

// WithPathParam adds a path parameter to the route.
// Path parameters are part of the URL path and are defined using a colon prefix in the route path.
//
// Parameters:
//   - name: The name of the parameter
//   - typ: The data type of the parameter (string, integer, boolean, etc.)
//   - required: Whether the parameter is required (typically true for path parameters)
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithPathParam(name, typ string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "path", typ, required, description, example)
}

// WithDeprecated marks a route as deprecated.
// Deprecated routes will be marked as such in the OpenAPI documentation.
//
// Parameters:
//   - message: An optional message explaining why the route is deprecated
//     and what to use instead
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

// WithResponse adds a response to the route.
// Responses are used to document the possible outcomes of an operation.
//
// Parameters:
//   - statusCode: The HTTP status code for the response
//   - description: A description of the response
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

// WithJSONResponse adds a JSON response with schema to the route.
// This is used to document responses that return JSON data.
//
// Parameters:
//   - statusCode: The HTTP status code for the response
//   - description: A description of the response
//   - schema: The schema of the response body
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

// WithSecurity adds security requirements to a route.
// Security requirements define the authentication methods that can be used
// to access the route.
//
// Parameters:
//   - requirements: Maps of security scheme names to required scopes
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
