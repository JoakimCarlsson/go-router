/*
Package metadata provides shared type definitions for API documentation and configuration
that are used across the router, OpenAPI, and Swagger UI packages.

The package defines core types for:
  - Route metadata (parameters, responses, security requirements)
  - OpenAPI schemas and components
  - OAuth2 configuration
  - Common utilities for status codes and type conversion

These types serve as the foundation for:
  - Documenting API routes and their behavior
  - Generating OpenAPI specifications
  - Configuring Swagger UI components
  - Managing API authentication and authorization

The types in this package are designed to be framework-agnostic and can be used
independently of the router implementation.
*/
package metadata
