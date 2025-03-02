package openapi

import "github.com/joakimcarlsson/go-router/metadata"

// RouteInfo represents information about a route needed for OpenAPI generation
type RouteInfo interface {
	Method() string
	Path() string
	OperationID() string
	Summary() string
	Description() string
	Tags() []string
	Parameters() []metadata.Parameter
	RequestBody() *metadata.RequestBody
	Responses() map[string]metadata.Response
	Security() []metadata.SecurityRequirement
	IsDeprecated() bool
}

// RouteMetadataAdapter adapts the RouteMetadata structure to the RouteInfo interface
type RouteMetadataAdapter struct {
	Metadata metadata.RouteMetadata
}

// Method returns the HTTP method of the route
func (a *RouteMetadataAdapter) Method() string {
	return a.Metadata.Method
}

// Path returns the path pattern of the route
func (a *RouteMetadataAdapter) Path() string {
	return a.Metadata.Path
}

// OperationID returns the operation ID of the route
func (a *RouteMetadataAdapter) OperationID() string {
	return a.Metadata.OperationID
}

// Summary returns the summary of the route
func (a *RouteMetadataAdapter) Summary() string {
	return a.Metadata.Summary
}

// Description returns the description of the route
func (a *RouteMetadataAdapter) Description() string {
	return a.Metadata.Description
}

// Tags returns the tags of the route
func (a *RouteMetadataAdapter) Tags() []string {
	return a.Metadata.Tags
}

// Parameters returns the parameters of the route
func (a *RouteMetadataAdapter) Parameters() []metadata.Parameter {
	return a.Metadata.Parameters
}

// RequestBody returns the request body of the route
func (a *RouteMetadataAdapter) RequestBody() *metadata.RequestBody {
	return a.Metadata.RequestBody
}

// Responses returns the responses of the route
func (a *RouteMetadataAdapter) Responses() map[string]metadata.Response {
	return a.Metadata.Responses
}

// Security returns the security requirements of the route
func (a *RouteMetadataAdapter) Security() []metadata.SecurityRequirement {
	return a.Metadata.Security
}

// IsDeprecated returns whether the route is deprecated
func (a *RouteMetadataAdapter) IsDeprecated() bool {
	return a.Metadata.Deprecated
}

// RouteInfoList is a collection of RouteInfo objects
type RouteInfoList []RouteInfo

// RouteInfoFromMetadata creates a RouteInfo from a RouteMetadata
func RouteInfoFromMetadata(metadata metadata.RouteMetadata) RouteInfo {
	return &RouteMetadataAdapter{
		Metadata: metadata,
	}
}
