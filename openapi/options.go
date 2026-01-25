package openapi

import (
	"reflect"

	"github.com/joakimcarlsson/go-router/router"
)

// OperationIDOption sets the operationId for a route.
type OperationIDOption struct{ OperationID string }

// SummaryOption sets the summary for a route.
type SummaryOption struct{ Summary string }

// DescriptionOption sets the description for a route.
type DescriptionOption struct{ Description string }

// TagsOption adds tags to a route.
type TagsOption struct{ Tags []string }

// DeprecatedOption marks a route as deprecated.
type DeprecatedOption struct{ Message string }

// ExcludeFromDocsOption excludes a route from OpenAPI documentation.
type ExcludeFromDocsOption struct{}

// ParameterOption adds a parameter to a route.
type ParameterOption struct {
	Name        string
	In          string
	Type        string
	Required    bool
	Description string
	Example     interface{}
}

// ParameterWithSchemaOption adds a parameter with a full schema to a route.
type ParameterWithSchemaOption struct {
	Name        string
	In          string
	Required    bool
	Description string
	Schema      Schema
}

// RequestBodyOption adds a request body with a specific content type.
type RequestBodyOption struct {
	ContentType string
	Schema      Schema
	Required    bool
	Description string
}

// JSONRequestBodyOption adds a JSON request body with schema inferred from a type.
type JSONRequestBodyOption struct {
	Type        reflect.Type
	Required    bool
	Description string
}

// MultipartFormDataOption adds a multipart form data request body.
type MultipartFormDataOption struct {
	Description string
	FormFields  map[string]FormFieldSpec
}

// MultipartFormStructOption adds a multipart form data request body from a struct type.
type MultipartFormStructOption struct {
	Type        reflect.Type
	Description string
}

// ResponseOption adds a response to a route.
type ResponseOption struct {
	StatusCode  int
	Description string
}

// JSONResponseOption adds a JSON response with schema inferred from a type.
type JSONResponseOption struct {
	StatusCode  int
	Description string
	Type        reflect.Type
}

// SecurityOption adds security requirements to a route.
type SecurityOption struct {
	Requirements []map[string][]string
}

// SSEResponseOption marks a route as returning Server-Sent Events.
type SSEResponseOption struct {
	Description string
}

// SSEEventOption documents a specific SSE event type.
type SSEEventOption struct {
	EventName   string
	Description string
	Type        reflect.Type
}

// SSEEventsOption documents multiple SSE event types.
type SSEEventsOption struct {
	Description string
	Events      []SSEEventSpec
}

// FormFieldSpec defines the specification for a form field.
type FormFieldSpec struct {
	Description string
	Required    bool
	Type        string
}

// SSEEventSpec defines the specification for an SSE event type.
type SSEEventSpec struct {
	Name        string
	Description string
	Schema      Schema
}

// WithOperationID sets the operationId for the route.
func WithOperationID(operationId string) router.RouteOption {
	return OperationIDOption{OperationID: operationId}
}

// WithSummary sets the route summary.
func WithSummary(summary string) router.RouteOption {
	return SummaryOption{Summary: summary}
}

// WithDescription sets the route description.
func WithDescription(description string) router.RouteOption {
	return DescriptionOption{Description: description}
}

// WithTags adds tags to the route.
func WithTags(tags ...string) router.RouteOption {
	return TagsOption{Tags: tags}
}

// WithParameter adds a parameter to the route.
func WithParameter(name, in, typ string, required bool, description string, example interface{}) router.RouteOption {
	return ParameterOption{
		Name:        name,
		In:          in,
		Type:        typ,
		Required:    required,
		Description: description,
		Example:     example,
	}
}

// WithQueryParam adds a query parameter to the route.
func WithQueryParam(name, typ string, required bool, description string, example interface{}) router.RouteOption {
	return WithParameter(name, "query", typ, required, description, example)
}

// WithPathParam adds a path parameter to the route.
func WithPathParam(name, typ string, required bool, description string, example interface{}) router.RouteOption {
	return WithParameter(name, "path", typ, required, description, example)
}

// WithFormattedPathParam adds a path parameter with a specific format.
func WithFormattedPathParam(name, format string, required bool, description string, example interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "path",
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "string",
			Format:  format,
			Example: example,
		},
	}
}

// WithRegexPathParam adds a path parameter with a regex pattern.
func WithRegexPathParam(name, pattern string, required bool, description string, example interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "path",
		Required:    required,
		Description: description + "\n\nMust match pattern: `" + pattern + "`",
		Schema: Schema{
			Type:    "string",
			Pattern: pattern,
			Example: example,
		},
	}
}

// WithNumericPathParam adds a numeric path parameter with optional range constraints.
func WithNumericPathParam(name string, required bool, description string, example float64, minimum, maximum *float64) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "path",
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "number",
			Example: example,
			Minimum: minimum,
			Maximum: maximum,
		},
	}
}

// WithIntegerPathParam adds an integer path parameter with optional range constraints.
func WithIntegerPathParam(name string, required bool, description string, example int64, minimum, maximum *float64) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "path",
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "integer",
			Example: example,
			Minimum: minimum,
			Maximum: maximum,
		},
	}
}

// WithEnumPathParam adds a path parameter with enumerated allowed values.
func WithEnumPathParam(name string, required bool, description string, example interface{}, values []interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "path",
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "string",
			Enum:    values,
			Example: example,
		},
	}
}

// WithFormattedQueryParam adds a query parameter with a specific format.
func WithFormattedQueryParam(name, format string, required bool, description string, example interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "query",
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "string",
			Format:  format,
			Example: example,
		},
	}
}

// WithRegexQueryParam adds a query parameter with a regex pattern.
func WithRegexQueryParam(name, pattern string, required bool, description string, example interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "query",
		Required:    required,
		Description: description + "\n\nMust match pattern: `" + pattern + "`",
		Schema: Schema{
			Type:    "string",
			Pattern: pattern,
			Example: example,
		},
	}
}

// WithEnumQueryParam adds a query parameter with enumerated allowed values.
func WithEnumQueryParam(name string, required bool, description string, example interface{}, values []interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          "query",
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "string",
			Enum:    values,
			Example: example,
		},
	}
}

// WithFormattedParam adds a parameter with a specific format.
func WithFormattedParam(name, in, format string, required bool, description string, example interface{}) router.RouteOption {
	return ParameterWithSchemaOption{
		Name:        name,
		In:          in,
		Required:    required,
		Description: description,
		Schema: Schema{
			Type:    "string",
			Format:  format,
			Example: example,
		},
	}
}

// WithUUIDPathParam adds a UUID path parameter.
func WithUUIDPathParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedPathParam(name, "uuid", required, description, example)
}

// WithDatePathParam adds a date path parameter.
func WithDatePathParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedPathParam(name, "date", required, description, example)
}

// WithDateTimePathParam adds a date-time path parameter.
func WithDateTimePathParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedPathParam(name, "date-time", required, description, example)
}

// WithEmailPathParam adds an email path parameter.
func WithEmailPathParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedPathParam(name, "email", required, description, example)
}

// WithUUIDQueryParam adds a UUID query parameter.
func WithUUIDQueryParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedQueryParam(name, "uuid", required, description, example)
}

// WithDateQueryParam adds a date query parameter.
func WithDateQueryParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedQueryParam(name, "date", required, description, example)
}

// WithDateTimeQueryParam adds a date-time query parameter.
func WithDateTimeQueryParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedQueryParam(name, "date-time", required, description, example)
}

// WithEmailQueryParam adds an email query parameter.
func WithEmailQueryParam(name string, required bool, description string, example string) router.RouteOption {
	return WithFormattedQueryParam(name, "email", required, description, example)
}

// WithHeaderParam adds a header parameter.
func WithHeaderParam(name string, required bool, description string, example interface{}) router.RouteOption {
	return WithParameter(name, "header", "string", required, description, example)
}

// WithRequestBody adds a request body with a specific content type.
func WithRequestBody(contentType string, schema Schema, required bool, description string) router.RouteOption {
	return RequestBodyOption{
		ContentType: contentType,
		Schema:      schema,
		Required:    required,
		Description: description,
	}
}

// WithJSONRequestBody adds a JSON request body with schema inferred from the type.
func WithJSONRequestBody[T any](required bool, description string) router.RouteOption {
	return JSONRequestBodyOption{
		Type:        reflect.TypeOf((*T)(nil)).Elem(),
		Required:    required,
		Description: description,
	}
}

// WithMultipartFormData adds a multipart form data request body.
func WithMultipartFormData(description string, formFields map[string]FormFieldSpec) router.RouteOption {
	return MultipartFormDataOption{
		Description: description,
		FormFields:  formFields,
	}
}

// WithMultipartFormStruct adds a multipart form data request body using a struct type.
func WithMultipartFormStruct[T any](description string) router.RouteOption {
	return MultipartFormStructOption{
		Type:        reflect.TypeOf((*T)(nil)).Elem(),
		Description: description,
	}
}

// WithResponse adds a response to the route.
func WithResponse(statusCode int, description string) router.RouteOption {
	return ResponseOption{
		StatusCode:  statusCode,
		Description: description,
	}
}

// WithJSONResponse adds a JSON response with schema inferred from the type.
func WithJSONResponse[T any](statusCode int, description string) router.RouteOption {
	return JSONResponseOption{
		StatusCode:  statusCode,
		Description: description,
		Type:        reflect.TypeOf((*T)(nil)).Elem(),
	}
}

// WithDeprecated marks a route as deprecated.
func WithDeprecated(message string) router.RouteOption {
	return DeprecatedOption{Message: message}
}

// WithSecurity adds security requirements to a route.
func WithSecurity(requirements ...map[string][]string) router.RouteOption {
	return SecurityOption{Requirements: requirements}
}

// WithBasicAuth adds basic authentication requirement.
func WithBasicAuth() router.RouteOption {
	return WithSecurity(map[string][]string{"basicAuth": {}})
}

// WithBearerAuth adds bearer token authentication requirement.
func WithBearerAuth() router.RouteOption {
	return WithSecurity(map[string][]string{"bearerAuth": {}})
}

// WithAPIKey adds API key authentication requirement.
func WithAPIKey() router.RouteOption {
	return WithSecurity(map[string][]string{"apiKey": {}})
}

// WithOAuth2Scopes adds OAuth2 authentication requirement with specific scopes.
func WithOAuth2Scopes(scopes ...string) router.RouteOption {
	return WithSecurity(map[string][]string{"oauth2": scopes})
}

// WithOAuth2Security adds OAuth2 authentication requirement without specific scopes.
func WithOAuth2Security() router.RouteOption {
	return WithSecurity(map[string][]string{"oauth2": {}})
}

// ExcludeFromDocs marks a route to be excluded from OpenAPI documentation.
func ExcludeFromDocs() router.RouteOption {
	return ExcludeFromDocsOption{}
}

// WithSSEResponse marks an endpoint as returning Server-Sent Events.
func WithSSEResponse(description string) router.RouteOption {
	return SSEResponseOption{Description: description}
}

// WithSSEEvent documents a specific event type that the SSE endpoint can emit.
func WithSSEEvent[T any](eventName, description string) router.RouteOption {
	return SSEEventOption{
		EventName:   eventName,
		Description: description,
		Type:        reflect.TypeOf((*T)(nil)).Elem(),
	}
}

// WithSSEEvents documents multiple SSE event types using manual specifications.
func WithSSEEvents(description string, events ...SSEEventSpec) router.RouteOption {
	return SSEEventsOption{
		Description: description,
		Events:      events,
	}
}
