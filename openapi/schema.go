package openapi

import (
	"github.com/joakimcarlsson/go-router/metadata"
)

// SchemaFromMetadataSchema converts a metadata Schema to an OpenAPI Schema
func SchemaFromMetadataSchema(s metadata.Schema) metadata.Schema {
	// Since types are now the same, we don't need conversion
	return s
}

// ParameterFromMetadataParameter converts a metadata Parameter to an OpenAPI Parameter
func ParameterFromMetadataParameter(p metadata.Parameter) metadata.Parameter {
	// Since types are now the same, we don't need conversion
	return p
}

// ResponseFromMetadataResponse converts a metadata Response to an OpenAPI Response
func ResponseFromMetadataResponse(r metadata.Response) metadata.Response {
	// Since types are now the same, we don't need conversion
	return r
}

// RequestBodyFromMetadataRequestBody converts a metadata RequestBody to an OpenAPI RequestBody
func RequestBodyFromMetadataRequestBody(r *metadata.RequestBody) *metadata.RequestBody {
	// Since types are now the same, we don't need conversion
	return r
}
