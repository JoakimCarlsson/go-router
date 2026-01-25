package openapi

// RouteInfo represents information about a route needed for OpenAPI generation.
type RouteInfo interface {
	Method() string
	Path() string
	OperationID() string
	Summary() string
	Description() string
	Tags() []string
	Parameters() []Parameter
	RequestBody() *RequestBody
	Responses() map[string]Response
	Security() []SecurityRequirement
	IsDeprecated() bool
	IsSSE() bool
	SSEEvents() []SSEEventSchema
}

// RouteMetadataAdapter adapts the RouteMetadata structure to the RouteInfo interface.
type RouteMetadataAdapter struct {
	Metadata RouteMetadata
}

// Method returns the HTTP method of the route.
func (a *RouteMetadataAdapter) Method() string {
	return a.Metadata.Method
}

// Path returns the path pattern of the route.
func (a *RouteMetadataAdapter) Path() string {
	return a.Metadata.Path
}

// OperationID returns the operation ID of the route.
func (a *RouteMetadataAdapter) OperationID() string {
	return a.Metadata.OperationID
}

// Summary returns the summary of the route.
func (a *RouteMetadataAdapter) Summary() string {
	return a.Metadata.Summary
}

// Description returns the description of the route.
func (a *RouteMetadataAdapter) Description() string {
	return a.Metadata.Description
}

// Tags returns the tags of the route.
func (a *RouteMetadataAdapter) Tags() []string {
	return a.Metadata.Tags
}

// Parameters returns the parameters of the route.
func (a *RouteMetadataAdapter) Parameters() []Parameter {
	return a.Metadata.Parameters
}

// RequestBody returns the request body of the route.
func (a *RouteMetadataAdapter) RequestBody() *RequestBody {
	return a.Metadata.RequestBody
}

// Responses returns the responses of the route.
func (a *RouteMetadataAdapter) Responses() map[string]Response {
	return a.Metadata.Responses
}

// Security returns the security requirements of the route.
func (a *RouteMetadataAdapter) Security() []SecurityRequirement {
	return a.Metadata.Security
}

// IsDeprecated returns whether the route is deprecated.
func (a *RouteMetadataAdapter) IsDeprecated() bool {
	return a.Metadata.Deprecated
}

// IsSSE returns whether this route returns Server-Sent Events.
func (a *RouteMetadataAdapter) IsSSE() bool {
	return a.Metadata.IsSSE
}

// SSEEvents returns the SSE event types this endpoint can emit.
func (a *RouteMetadataAdapter) SSEEvents() []SSEEventSchema {
	return a.Metadata.SSEEvents
}

// RouteInfoList is a collection of RouteInfo objects.
type RouteInfoList []RouteInfo

// RouteInfoFromMetadata creates a RouteInfo from a RouteMetadata.
func RouteInfoFromMetadata(m RouteMetadata) RouteInfo {
	return &RouteMetadataAdapter{
		Metadata: m,
	}
}
