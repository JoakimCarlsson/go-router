package metadata

import (
	"reflect"
	"strings"
	"sync"
)

// RouteMetadata contains documentation and configuration for a route.
// This structure is used for generating OpenAPI documentation and provides
// all the information needed to describe an API endpoint.
type RouteMetadata struct {
	// Core routing information
	Method string `json:"-"`
	Path   string `json:"-"`

	// Documentation
	OperationID string   `json:"operationId,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Deprecated  bool     `json:"deprecated,omitempty"`

	// API Documentation (OpenAPI specific)
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
}

// Parameter represents an API parameter such as path, query, header, or cookie parameters.
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, path, header, cookie
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      Schema      `json:"schema"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody represents a request body for an API operation.
// It contains information about the content type, schema, and whether the body is required.
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents an API response for an operation.
// It includes a description, content schema by media type, and optional headers.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

// SecurityRequirement represents security requirements for an operation.
// The map keys are security scheme names and the values are required scopes.
type SecurityRequirement map[string][]string

// MediaType represents the structure of request/response content.
// It includes a schema and an optional example.
type MediaType struct {
	Schema  Schema      `json:"schema"`
	Example interface{} `json:"example,omitempty"`
}

// Header represents a response header.
// It includes a description and schema for the header value.
type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

// Schema represents a type schema used in OpenAPI specifications.
// It defines the structure of request/response data.
type Schema struct {
	Type                 string            `json:"type,omitempty"`
	Ref                  string            `json:"$ref,omitempty"`
	Format               string            `json:"format,omitempty"`
	Description          string            `json:"description,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	Example              interface{}       `json:"example,omitempty"`
	Required             []string          `json:"required,omitempty"`
	MinLength            *int              `json:"minLength,omitempty"`
	MaxLength            *int              `json:"maxLength,omitempty"`
	Minimum              *float64          `json:"minimum,omitempty"`
	Maximum              *float64          `json:"maximum,omitempty"`
	Enum                 []interface{}     `json:"enum,omitempty"`
	AllOf                []Schema          `json:"allOf,omitempty"`
	OneOf                []Schema          `json:"oneOf,omitempty"`
	AnyOf                []Schema          `json:"anyOf,omitempty"`
	Nullable             bool              `json:"nullable,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty"`
	TypeName             string            `json:"-"`
}

// TypeRegistryEntry stores information about a registered type
type TypeRegistryEntry struct {
	Name      string
	PkgPath   string
	Count     int
	FinalName string
}

// typeRegistry tracks registered types and detects name collisions
type typeRegistry struct {
	types map[string]*TypeRegistryEntry
	mu    sync.RWMutex
}

// global registry instance
var globalTypeRegistry *typeRegistry

// init initializes the global type registry
func init() {
	globalTypeRegistry = &typeRegistry{
		types: make(map[string]*TypeRegistryEntry),
	}
}

// RegisterType adds a type to the registry and returns a non-colliding name
func RegisterType(t reflect.Type) string {
	globalTypeRegistry.mu.Lock()
	defer globalTypeRegistry.mu.Unlock()

	name := t.Name()
	pkgPath := t.PkgPath()
	fullID := pkgPath + "." + name

	// Check if we've seen this exact type before (same name and package)
	if entry, exists := globalTypeRegistry.types[fullID]; exists {
		entry.Count++
		// Return the name we've already assigned to this type
		return entry.FinalName
	}

	// Check if we've seen this base name before but with a different package
	if entry, exists := globalTypeRegistry.types[name]; exists {
		// This is a collision - we need qualified names for both

		// If this is the first collision with this name, we need to rename the original entry
		if entry.Count == 1 && entry.FinalName == name {
			// Update the original entry to use a qualified name
			origFullID := entry.PkgPath + "." + entry.Name

			// Calculate the original's qualified name
			origQualifiedName := SanitizeSchemaName(entry.PkgPath + "_" + entry.Name)
			entry.FinalName = origQualifiedName

			// Update map to point to the same entry with full ID
			globalTypeRegistry.types[origFullID] = entry
			delete(globalTypeRegistry.types, name)
		}

		// Register this new type with its package-qualified name
		qualifiedName := SanitizeSchemaName(pkgPath + "_" + name)
		globalTypeRegistry.types[fullID] = &TypeRegistryEntry{
			Name:      name,
			PkgPath:   pkgPath,
			Count:     1,
			FinalName: qualifiedName,
		}

		// Return the qualified name when there's a collision
		return qualifiedName
	}

	// First time seeing this name, register with the simple name
	globalTypeRegistry.types[name] = &TypeRegistryEntry{
		Name:      name,
		PkgPath:   pkgPath,
		Count:     1,
		FinalName: name, // Initially use the simple name
	}

	// Also register with the full ID for exact lookups
	globalTypeRegistry.types[fullID] = globalTypeRegistry.types[name]

	// Return simple name when there's no collision
	return name
}

// SanitizeSchemaName converts a fully qualified type name to a valid schema name
// by removing invalid characters and normalizing the format
func SanitizeSchemaName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// OAuth2Config holds OAuth2 configuration for API authentication.
// This is used in Swagger UI to configure OAuth2 flows.
type OAuth2Config struct {
	// ClientID is the OAuth2 client ID
	ClientID string
	// ClientSecret is the OAuth2 client secret (typically only used in password, implicit, or access code flows)
	ClientSecret string
	// Realm is the realm query parameter
	Realm string
	// AppName is the application name for OAuth2 authorization
	AppName string
	// ScopeSeparator is the separator used when passing multiple scopes
	ScopeSeparator string
	// Scopes is a predefined list of scopes to be used
	Scopes []string
	// AdditionalQueryParams allows adding query params to the OAuth2 flow
	AdditionalQueryParams map[string]string
	// UseBasicAuthenticationWithAccessCodeGrant requires sending client credentials via header
	UseBasicAuthenticationWithAccessCodeGrant bool
	// UsePkceWithAuthorizationCodeGrant uses PKCE when available
	UsePkceWithAuthorizationCodeGrant bool
}

// NewOAuth2Config creates a new OAuth2 configuration with default values.
// It initializes sensible defaults for scope separator and security features.
func NewOAuth2Config() *OAuth2Config {
	return &OAuth2Config{
		ScopeSeparator:                            " ",
		AdditionalQueryParams:                     make(map[string]string),
		UseBasicAuthenticationWithAccessCodeGrant: false,
		UsePkceWithAuthorizationCodeGrant:         true,
	}
}

// WithClientID sets the client ID for OAuth2 configuration.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithClientID(clientID string) *OAuth2Config {
	c.ClientID = clientID
	return c
}

// WithClientSecret sets the client secret for OAuth2 configuration.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithClientSecret(clientSecret string) *OAuth2Config {
	c.ClientSecret = clientSecret
	return c
}

// WithRealm sets the realm for OAuth2 configuration.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithRealm(realm string) *OAuth2Config {
	c.Realm = realm
	return c
}

// WithAppName sets the application name for OAuth2 configuration.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithAppName(appName string) *OAuth2Config {
	c.AppName = appName
	return c
}

// WithScopeSeparator sets the scope separator for OAuth2 configuration.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithScopeSeparator(separator string) *OAuth2Config {
	c.ScopeSeparator = separator
	return c
}

// WithScopes sets the scopes for OAuth2 configuration.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithScopes(scopes ...string) *OAuth2Config {
	c.Scopes = scopes
	return c
}

// WithAdditionalQueryParam adds a query parameter to the OAuth2 flow.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithAdditionalQueryParam(key, value string) *OAuth2Config {
	c.AdditionalQueryParams[key] = value
	return c
}

// WithBasicAuthentication sets whether to use basic authentication with access code grant.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithBasicAuthentication(use bool) *OAuth2Config {
	c.UseBasicAuthenticationWithAccessCodeGrant = use
	return c
}

// WithPKCE sets whether to use PKCE with authorization code grant.
// PKCE (Proof Key for Code Exchange) provides additional security for public clients.
// Returns the OAuth2Config for method chaining.
func (c *OAuth2Config) WithPKCE(use bool) *OAuth2Config {
	c.UsePkceWithAuthorizationCodeGrant = use
	return c
}
