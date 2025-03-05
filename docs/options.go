package docs

import (
	"reflect"
	"strings"

	"github.com/joakimcarlsson/go-router/metadata"
)

// RouteOption configures route metadata for API documentation.
// Route options are functions that modify a RouteMetadata object
// to add documentation details like parameters, responses, etc.
type RouteOption func(*metadata.RouteMetadata)

// WithOperationID sets the operationId for the route.
// The operationId is a unique identifier for the operation and is used in
// generated client libraries as the function name.
func WithOperationID(operationId string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.OperationID = operationId
	}
}

// WithSummary sets the route summary.
// The summary is a short description of what the operation does.
func WithSummary(summary string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Summary = summary
	}
}

// WithDescription sets the route description.
// The description provides more detailed information about the operation.
func WithDescription(description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Description = description
	}
}

// WithTags adds tags to the route.
// Tags are used to group operations by logical groups in the API documentation.
func WithTags(tags ...string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Tags = append(m.Tags, tags...)
	}
}

// WithParameter adds a parameter to the route.
// This is a generic function that can add any type of parameter (path, query, header, etc.).
//
// Parameters:
//   - name: The parameter name
//   - in: The parameter location (path, query, header, cookie)
//   - typ: The parameter type (string, integer, boolean, etc.)
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
//   - name: The parameter name
//   - typ: The parameter type (string, integer, boolean, etc.)
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithQueryParam(name, typ string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "query", typ, required, description, example)
}

// WithPathParam adds a path parameter to the route.
// Path parameters are part of the URL path and are denoted by a colon prefix in the route pattern.
//
// Parameters:
//   - name: The parameter name (without the colon)
//   - typ: The parameter type (string, integer, boolean, etc.)
//   - required: Whether the parameter is required (typically true for path parameters)
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithPathParam(name, typ string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "path", typ, required, description, example)
}

// WithHeaderParam adds a header parameter to the route.
// Header parameters are sent in the HTTP request headers.
//
// Parameters:
//   - name: The header name
//   - required: Whether the header is required
//   - description: A description of the header
//   - example: An example value for the header
func WithHeaderParam(name string, required bool, description string, example interface{}) RouteOption {
	return WithParameter(name, "header", "string", required, description, example)
}

// WithRequestBody adds a request body with a specific content type.
// This defines the schema and requirements for the request body.
//
// Parameters:
//   - contentType: The media type of the request body (e.g., "application/json")
//   - schema: The schema describing the request body structure
//   - required: Whether the request body is required
//   - description: A description of the request body
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

// WithJSONRequestBody adds a JSON request body with schema inferred from the provided type.
// This uses Go's reflect package to generate a schema from the type parameter T.
//
// Type Parameters:
//   - T: The Go type to use for the request body schema
//
// Parameters:
//   - required: Whether the request body is required
//   - description: A description of the request body
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

// WithMultipartFormData adds a multipart form data request body to the route.
// This is useful for file uploads and form submissions with files.
//
// Parameters:
//   - required: Whether the request body is required
//   - description: A description of the request body
//   - formFields: A map where keys are field names and values are field descriptions
func WithMultipartFormData(required bool, description string, formFields map[string]string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		properties := make(map[string]metadata.Schema)
		requiredFields := make([]string, 0)

		for fieldName, fieldDesc := range formFields {
			if strings.HasSuffix(fieldName, "[]") {
				baseName := strings.TrimSuffix(fieldName, "[]")
				properties[baseName] = metadata.Schema{
					Type: "array",
					Items: &metadata.Schema{
						Type:        "string",
						Format:      "binary",
						Description: fieldDesc,
					},
				}
				requiredFields = append(requiredFields, baseName)
			} else {
				properties[fieldName] = metadata.Schema{
					Type:        "string",
					Format:      "binary",
					Description: fieldDesc,
				}
				requiredFields = append(requiredFields, fieldName)
			}
		}

		schema := metadata.Schema{
			Type:       "object",
			Properties: properties,
			Required:   requiredFields,
		}

		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required:    required,
			Content: map[string]metadata.MediaType{
				"multipart/form-data": {Schema: schema},
			},
		}
	}
}

// WithResponse adds a response to the route.
// This defines a response without any content schema.
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

// WithJSONResponse adds a JSON response with schema inferred from the provided type.
// This uses Go's reflect package to generate a schema from the type parameter T.
//
// Type Parameters:
//   - T: The Go type to use for the response schema
//
// Parameters:
//   - statusCode: The HTTP status code for the response
//   - description: A description of the response
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

// WithDeprecated marks a route as deprecated.
// Deprecated routes will be clearly marked in the API documentation.
//
// Parameters:
//   - message: An optional message explaining why the route is deprecated and
//     what to use instead
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

// WithBasicAuth adds basic authentication requirement to a route.
// This adds a security requirement for HTTP Basic authentication.
func WithBasicAuth() RouteOption {
	return WithSecurity(map[string][]string{"basicAuth": {}})
}

// WithBearerAuth adds bearer token authentication requirement to a route.
// This adds a security requirement for HTTP Bearer token authentication.
func WithBearerAuth() RouteOption {
	return WithSecurity(map[string][]string{"bearerAuth": {}})
}

// WithAPIKey adds API key authentication requirement to a route.
// This adds a security requirement for API key authentication.
func WithAPIKey() RouteOption {
	return WithSecurity(map[string][]string{"apiKey": {}})
}

// WithOAuth2Scopes adds OAuth2 authentication requirement with specific scopes.
// This adds a security requirement for OAuth2 authentication with the specified scopes.
//
// Parameters:
//   - scopes: The OAuth2 scopes required for the operation
func WithOAuth2Scopes(scopes ...string) RouteOption {
	return WithSecurity(map[string][]string{"oauth2": scopes})
}
