package router

import (
	"github.com/joakimcarlsson/go-router/openapi"
)

// RouteMetadata holds the OpenAPI metadata for a route
type RouteMetadata = openapi.RouteMetadata

// RouteOption is a function that configures route metadata
type RouteOption = openapi.RouteOption

// Parameter represents an OpenAPI parameter
type Parameter = openapi.Parameter

// Response represents an OpenAPI response
type Response = openapi.Response

// SecurityRequirement represents an OpenAPI security requirement
type SecurityRequirement = openapi.SecurityRequirement

// RequestBody represents an OpenAPI request body
type RequestBody = openapi.RequestBody
