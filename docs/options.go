package docs

import (
	"reflect"

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
func WithParameter(
	name, in, typ string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		metadata.AddParameter(m, name, in, typ, required, description, example)
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
func WithQueryParam(
	name, typ string,
	required bool,
	description string,
	example interface{},
) RouteOption {
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
func WithPathParam(
	name, typ string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return WithParameter(name, "path", typ, required, description, example)
}

// WithFormattedPathParam adds a path parameter with a specific format to the route.
// The format field specifies the format of the parameter value, such as "uuid", "date", etc.
//
// Parameters:
//   - name: The parameter name
//   - format: The format of the parameter (uuid, date-time, date, email, uri, etc.)
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithFormattedPathParam(
	name, format string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		schema := metadata.Schema{
			Type:    "string",
			Format:  format,
			Example: example,
		}
		metadata.AddParameterWithSchema(
			m,
			name,
			"path",
			required,
			description,
			schema,
		)
	}
}

// WithRegexPathParam adds a path parameter with a regex pattern to the route.
// This documents that the parameter must match the specified pattern.
//
// Parameters:
//   - name: The parameter name
//   - pattern: The regex pattern the parameter must match
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithRegexPathParam(
	name, pattern string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		schema := metadata.Schema{
			Type:    "string",
			Pattern: pattern,
			Example: example,
		}
		metadata.AddParameterWithSchema(
			m,
			name,
			"path",
			required,
			description+"\n\nMust match pattern: `"+pattern+"`",
			schema,
		)
	}
}

// WithNumericPathParam adds a numeric path parameter to the route with optional range constraints.
//
// Parameters:
//   - name: The parameter name
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
//   - minimum: Minimum value (optional, set to nil if not needed)
//   - maximum: Maximum value (optional, set to nil if not needed)
func WithNumericPathParam(
	name string,
	required bool,
	description string,
	example float64,
	minimum, maximum *float64,
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		schema := metadata.Schema{
			Type:    "number",
			Example: example,
			Minimum: minimum,
			Maximum: maximum,
		}
		metadata.AddParameterWithSchema(
			m,
			name,
			"path",
			required,
			description,
			schema,
		)
	}
}

// WithIntegerPathParam adds an integer path parameter to the route with optional range constraints.
//
// Parameters:
//   - name: The parameter name
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
//   - minimum: Minimum value (optional, set to nil if not needed)
//   - maximum: Maximum value (optional, set to nil if not needed)
func WithIntegerPathParam(
	name string,
	required bool,
	description string,
	example int64,
	minimum, maximum *float64,
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		schema := metadata.Schema{
			Type:    "integer",
			Example: example,
		}

		if minimum != nil {
			schema.Minimum = minimum
		}
		if maximum != nil {
			schema.Maximum = maximum
		}

		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          "path",
			Required:    required,
			Description: description,
			Schema:      schema,
		})
	}
}

// WithEnumPathParam adds a path parameter with enumerated allowed values to the route.
//
// Parameters:
//   - name: The parameter name
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
//   - values: Array of allowed values for the parameter
func WithEnumPathParam(
	name string,
	required bool,
	description string,
	example interface{},
	values []interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          "path",
			Required:    required,
			Description: description,
			Schema: metadata.Schema{
				Type:    "string",
				Enum:    values,
				Example: example,
			},
		})
	}
}

// WithFormattedQueryParam adds a query parameter with a specific format to the route.
// The format field specifies the format of the parameter value, such as "uuid", "date", etc.
//
// Parameters:
//   - name: The parameter name
//   - format: The format of the parameter (uuid, date-time, date, email, uri, etc.)
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithFormattedQueryParam(
	name, format string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          "query",
			Required:    required,
			Description: description,
			Schema: metadata.Schema{
				Type:    "string",
				Format:  format,
				Example: example,
			},
		})
	}
}

// WithRegexQueryParam adds a query parameter with a regex pattern to the route.
// This documents that the parameter must match the specified pattern.
//
// Parameters:
//   - name: The parameter name
//   - pattern: The regex pattern the parameter must match
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithRegexQueryParam(
	name, pattern string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          "query",
			Required:    required,
			Description: description + "\n\nMust match pattern: `" + pattern + "`",
			Schema: metadata.Schema{
				Type:    "string",
				Pattern: pattern,
				Example: example,
			},
		})
	}
}

// WithEnumQueryParam adds a query parameter with enumerated allowed values to the route.
//
// Parameters:
//   - name: The parameter name
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
//   - values: Array of allowed values for the parameter
func WithEnumQueryParam(
	name string,
	required bool,
	description string,
	example interface{},
	values []interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          "query",
			Required:    required,
			Description: description,
			Schema: metadata.Schema{
				Type:    "string",
				Enum:    values,
				Example: example,
			},
		})
	}
}

// WithFormattedParam is a generic function that adds a parameter with a specific format to the route.
// The format field specifies the format of the parameter value, such as "uuid", "date", etc.
//
// Parameters:
//   - name: The parameter name
//   - in: The parameter location (path, query, header, cookie)
//   - format: The format of the parameter (uuid, date-time, date, email, uri, etc.)
//   - required: Whether the parameter is required
//   - description: A description of the parameter
//   - example: An example value for the parameter
func WithFormattedParam(
	name, in, format string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.Parameters = append(m.Parameters, metadata.Parameter{
			Name:        name,
			In:          in,
			Required:    required,
			Description: description,
			Schema: metadata.Schema{
				Type:    "string",
				Format:  format,
				Example: example,
			},
		})
	}
}

// WithUUIDPathParam adds a UUID path parameter to the route.
func WithUUIDPathParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedPathParam(name, "uuid", required, description, example)
}

// WithDatePathParam adds a date path parameter to the route.
func WithDatePathParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedPathParam(name, "date", required, description, example)
}

// WithDateTimePathParam adds a date-time path parameter to the route.
func WithDateTimePathParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedPathParam(
		name,
		"date-time",
		required,
		description,
		example,
	)
}

// WithEmailPathParam adds an email path parameter to the route.
func WithEmailPathParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedPathParam(name, "email", required, description, example)
}

// WithUUIDQueryParam adds a UUID query parameter to the route.
func WithUUIDQueryParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedQueryParam(name, "uuid", required, description, example)
}

// WithDateQueryParam adds a date query parameter to the route.
func WithDateQueryParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedQueryParam(name, "date", required, description, example)
}

// WithDateTimeQueryParam adds a date-time query parameter to the route.
func WithDateTimeQueryParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedQueryParam(
		name,
		"date-time",
		required,
		description,
		example,
	)
}

// WithEmailQueryParam adds an email query parameter to the route.
func WithEmailQueryParam(
	name string,
	required bool,
	description string,
	example string,
) RouteOption {
	return WithFormattedQueryParam(
		name,
		"email",
		required,
		description,
		example,
	)
}

// WithHeaderParam adds a header parameter to the route.
// Header parameters are sent in the HTTP request headers.
//
// Parameters:
//   - name: The header name
//   - required: Whether the header is required
//   - description: A description of the header
//   - example: An example value for the header
func WithHeaderParam(
	name string,
	required bool,
	description string,
	example interface{},
) RouteOption {
	return WithParameter(
		name,
		"header",
		"string",
		required,
		description,
		example,
	)
}

// WithRequestBody adds a request body with a specific content type.
// This defines the schema and requirements for the request body.
//
// Parameters:
//   - contentType: The media type of the request body (e.g., metadata.ContentTypeJSON)
//   - schema: The schema describing the request body structure
//   - required: Whether the request body is required
//   - description: A description of the request body
func WithRequestBody(
	contentType string,
	schema metadata.Schema,
	required bool,
	description string,
) RouteOption {
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
		t := GetTypeFromGeneric[T]()
		schema := SchemaFromType(t)
		metadata.AddJSONRequestBody(m, description, required, schema)
	}
}

// FormFieldSpec defines the specification for a form field
type FormFieldSpec struct {
	Description string
	Required    bool
	Type        string // "file", "file[]", or "string"
}

// WithMultipartFormData adds a multipart form data request body to the route.
// This is useful for file uploads and form submissions with files.
//
// Parameters:
//   - description: A description of the request body
//   - formFields: A map where keys are field names and values are field specifications
func WithMultipartFormData(
	description string,
	formFields map[string]FormFieldSpec,
) RouteOption {
	return func(m *metadata.RouteMetadata) {
		properties := make(map[string]metadata.Schema)
		requiredFields := make([]string, 0)

		for fieldName, spec := range formFields {
			switch spec.Type {
			case "file[]":
				// Array of files
				properties[fieldName] = metadata.Schema{
					Type: "array",
					Items: &metadata.Schema{
						Type:        "string",
						Format:      "binary",
						Description: spec.Description,
					},
				}
			case "file":
				// Single file field
				properties[fieldName] = metadata.Schema{
					Type:        "string",
					Format:      "binary",
					Description: spec.Description,
				}
			default:
				// Regular form field (string)
				properties[fieldName] = metadata.Schema{
					Type:        "string",
					Description: spec.Description,
				}
			}

			if spec.Required {
				requiredFields = append(requiredFields, fieldName)
			}
		}

		schema := metadata.Schema{
			Type:       "object",
			Properties: properties,
		}

		// Only add required fields if there are any
		if len(requiredFields) > 0 {
			schema.Required = requiredFields
		}

		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required: len(
				requiredFields,
			) > 0, // RequestBody is required if any field is required
			Content: map[string]metadata.MediaType{
				metadata.ContentTypeFormData: {Schema: schema},
			},
		}
	}
}

// WithMultipartFormStruct adds a multipart form data request body to the route
// using a struct type to define the form fields and their requirements.
// Field tags are used to configure the form:
//   - `form:"name"` defines the form field name
//   - `file:"true"` indicates a file upload field
//   - `required:"true"` marks the field as required
//   - `description:"text"` provides field description for the docs
//
// Example:
//
//	type Upload struct {
//	    File        *multipart.FileHeader `form:"file" file:"true" required:"true" description:"The file to upload"`
//	    Name        string                `form:"name" description:"Optional name for the file"`
//	    Description string                `form:"description" description:"File description"`
//	}
func WithMultipartFormStruct[T any](description string) RouteOption {
	return func(m *metadata.RouteMetadata) {
		t := GetTypeFromGeneric[T]()
		properties := make(map[string]metadata.Schema)
		requiredFields := make([]string, 0)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Get form field name
			formTag := field.Tag.Get("form")
			if formTag == "" {
				continue
			}

			// Check if field is a file upload
			isFile := field.Tag.Get("file") == "true"
			isRequired := field.Tag.Get("required") == "true"
			fieldDesc := field.Tag.Get("description")
			if fieldDesc == "" {
				fieldDesc = formTag // Use form name as fallback description
			}

			// Determine if it's an array of files
			isFileArray := isFile && field.Type.Kind() == reflect.Slice

			var schema metadata.Schema
			if isFile {
				if isFileArray {
					schema = metadata.Schema{
						Type: "array",
						Items: &metadata.Schema{
							Type:        "string",
							Format:      "binary",
							Description: fieldDesc,
						},
					}
				} else {
					schema = metadata.Schema{
						Type:        "string",
						Format:      "binary",
						Description: fieldDesc,
					}
				}
			} else {
				schema = metadata.Schema{
					Type:        "string",
					Description: fieldDesc,
				}
			}

			properties[formTag] = schema
			if isRequired {
				requiredFields = append(requiredFields, formTag)
			}
		}

		schema := metadata.Schema{
			Type:       "object",
			Properties: properties,
		}

		if len(requiredFields) > 0 {
			schema.Required = requiredFields
		}

		m.RequestBody = &metadata.RequestBody{
			Description: description,
			Required:    len(requiredFields) > 0,
			Content: map[string]metadata.MediaType{
				metadata.ContentTypeFormData: {Schema: schema},
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
		metadata.AddSimpleResponse(m, statusCode, description)
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
		t := GetTypeFromGeneric[T]()
		schema := SchemaFromType(t)
		metadata.AddJSONResponse(m, statusCode, description, schema)
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

// ExcludeFromDocs marks a route to be excluded from OpenAPI documentation.
// This is useful for internal routes like health checks, documentation endpoints,
// or any routes that should not appear in the public API documentation.
func ExcludeFromDocs() RouteOption {
	return func(m *metadata.RouteMetadata) {
		m.ExcludeFromDocs = true
	}
}
