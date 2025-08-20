package metadata

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
)

// Common content-type constants to avoid magic strings
const (
	ContentTypeJSON        = "application/json"
	ContentTypeXML         = "application/xml"
	ContentTypeEventStream = "text/event-stream"
	ContentTypeHTML        = "text/html"
	ContentTypeFormData    = "multipart/form-data"
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

// MarshalJSON implements json.Marshaler for Parameter to ensure all fields are serialized
func (p Parameter) MarshalJSON() ([]byte, error) {
	// Create a custom struct to ensure all Parameter fields are included
	type ParameterJSON struct {
		Name        string      `json:"name"`
		In          string      `json:"in"`
		Required    bool        `json:"required,omitempty"`
		Description string      `json:"description,omitempty"`
		Schema      Schema      `json:"schema"`
		Example     interface{} `json:"example,omitempty"`
	}

	return json.Marshal(ParameterJSON(p))
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
	Schema    Schema      `json:"schema,omitempty"`
	Example   interface{} `json:"example,omitempty"`
	SchemaRef *Reference  `json:"-"` // Not directly serialized to JSON
}

// MarshalJSON implements json.Marshaler for MediaType to handle SchemaRef
func (m MediaType) MarshalJSON() ([]byte, error) {
	// If SchemaRef is set, serialize with the Reference instead of Schema
	if m.SchemaRef != nil {
		return json.Marshal(struct {
			Schema  *Reference  `json:"schema"`
			Example interface{} `json:"example,omitempty"`
		}{
			Schema:  m.SchemaRef,
			Example: m.Example,
		})
	}

	// Otherwise, use the default serialization (Schema)
	return json.Marshal(struct {
		Schema  Schema      `json:"schema,omitempty"`
		Example interface{} `json:"example,omitempty"`
	}{
		Schema:  m.Schema,
		Example: m.Example,
	})
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
	Pattern              string            `json:"pattern,omitempty"`
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

// Reference is a JSON reference to another component in the OpenAPI document
type Reference struct {
	Ref string `json:"$ref"`
}

// SchemaOrReference can be either a Schema object or a Reference to a schema
type SchemaOrReference struct {
	Schema    *Schema    `json:"-"` // The schema object to reference
	Reference *Reference `json:"-"` // The reference to a schema
}

// MarshalJSON implements the json.Marshaler interface for SchemaOrReference
func (s SchemaOrReference) MarshalJSON() ([]byte, error) {
	if s.Reference != nil {
		return json.Marshal(s.Reference)
	}
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	return json.Marshal(nil)
}

// Spec represents the OpenAPI 3.0.0 specification
type Spec struct {
	OpenAPI      string              `json:"openapi"`
	Info         Info                `json:"info"`
	Servers      []Server            `json:"servers,omitempty"`
	Paths        map[string]PathItem `json:"paths"`
	Components   *Components         `json:"components,omitempty"`
	Tags         []Tag               `json:"tags,omitempty"`
	ExternalDocs map[string]string   `json:"externalDocs,omitempty"`
}

// Info represents OpenAPI info object
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Version        string   `json:"version"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Contact information for the API
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information for the API
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server information for the API
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable defines a variable for server URL template substitution
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem describes operations available on a single API endpoint
type PathItem struct {
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

// Operation describes a single API operation on a path
type Operation struct {
	OperationID string                `json:"operationId,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
}

// Components holds reusable OpenAPI objects
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme defines security mechanism for API
type SecurityScheme struct {
	Type             string      `json:"type"`
	Scheme           string      `json:"scheme,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Description      string      `json:"description,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows is the configuration container for the supported OAuth Flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow configuration details for a specific OAuth Flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Tag represents a tag
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// TypeHandler is a function that can generate a Schema for a specific Go type.
// Custom type handlers can be registered to support specialized type conversions.
type TypeHandler func(reflect.Type) Schema

// typeHandlerRegistry stores custom type handlers for Schema generation
type typeHandlerRegistry struct {
	handlers map[string]TypeHandler
	mu       sync.RWMutex
}

// global instance of the type handler registry
var globalTypeHandlerRegistry *typeHandlerRegistry

// init creates and initializes the global type handler registry
func init() {
	globalTypeRegistry = &typeRegistry{
		types: make(map[string]*TypeRegistryEntry),
	}

	globalTypeHandlerRegistry = &typeHandlerRegistry{
		handlers: make(map[string]TypeHandler),
	}
}

// RegisterTypeHandler adds a custom type handler for a specific type.
// The typeName should be in the format "package.TypeName" (e.g., "time.Time").
// The handler function will be called when SchemaFromType encounters this type.
// This allows users to customize how Schemas are generated for their own types.
func RegisterTypeHandler(typeName string, handler TypeHandler) {
	globalTypeHandlerRegistry.mu.Lock()
	defer globalTypeHandlerRegistry.mu.Unlock()

	globalTypeHandlerRegistry.handlers[typeName] = handler
}

// GetTypeHandler retrieves a type handler for a given type name if registered.
// Returns the handler function and a boolean indicating if a handler was found.
func GetTypeHandler(typeName string) (TypeHandler, bool) {
	globalTypeHandlerRegistry.mu.RLock()
	defer globalTypeHandlerRegistry.mu.RUnlock()

	handler, exists := globalTypeHandlerRegistry.handlers[typeName]
	return handler, exists
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
	types map[string]*TypeRegistryEntry // Map of type name to its Registry entry
	mu    sync.RWMutex                  // Mutex to ensure concurrent access is synchronized
}

// global registry instance
var globalTypeRegistry *typeRegistry // Singleton instance of typeRegistry

// init initializes the global type registry
func init() {
	globalTypeRegistry = &typeRegistry{
		types: make(map[string]*TypeRegistryEntry),
	}
}

// RegisterType adds a type to the registry and returns a non-colliding name
func RegisterType(t reflect.Type) string {
	globalTypeRegistry.mu.Lock()         // Acquire write lock
	defer globalTypeRegistry.mu.Unlock() // Release write lock when done

	name := t.Name()               // Base name
	pkgPath := t.PkgPath()         // Package path
	fullID := pkgPath + "." + name // Full ID of the type

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

// EnsureResponsesMap initializes the Responses map if it's nil.
// This is a utility function to avoid repeated nil checking and map creation.
func EnsureResponsesMap(m *RouteMetadata) {
	if m.Responses == nil {
		m.Responses = make(map[string]Response)
	}
}

// AddResponse is a universal helper for adding responses to RouteMetadata.
// This consolidates the repeated pattern of creating Response structs.
func AddResponse(m *RouteMetadata, statusCode int, description string, content map[string]MediaType) {
	EnsureResponsesMap(m)
	m.Responses[StatusCodeToString(statusCode)] = Response{
		Description: description,
		Content:     content,
	}
}

// AddSimpleResponse adds a response without any content.
func AddSimpleResponse(m *RouteMetadata, statusCode int, description string) {
	AddResponse(m, statusCode, description, nil)
}

// AddJSONResponse adds a response with JSON content.
func AddJSONResponse(m *RouteMetadata, statusCode int, description string, schema Schema) {
	content := map[string]MediaType{
		ContentTypeJSON: {Schema: schema},
	}
	AddResponse(m, statusCode, description, content)
}

// AddJSONResponseWithRef adds a response with JSON content using a schema reference.
func AddJSONResponseWithRef(m *RouteMetadata, statusCode int, description string, ref *Reference) {
	content := map[string]MediaType{
		ContentTypeJSON: {SchemaRef: ref},
	}
	AddResponse(m, statusCode, description, content)
}

// AddParameter is a universal helper for adding parameters to RouteMetadata.
func AddParameter(m *RouteMetadata, name, in, typ string, required bool, description string, example interface{}) {
	schema := Schema{
		Type:    typ,
		Example: example,
	}
	m.Parameters = append(m.Parameters, Parameter{
		Name:        name,
		In:          in,
		Required:    required,
		Description: description,
		Schema:      schema,
	})
}

// AddParameterWithSchema adds a parameter with a full schema specification.
func AddParameterWithSchema(m *RouteMetadata, name, in string, required bool, description string, schema Schema) {
	m.Parameters = append(m.Parameters, Parameter{
		Name:        name,
		In:          in,
		Required:    required,
		Description: description,
		Schema:      schema,
	})
}

// AddArrayJSONResponse adds a JSON response for array types with schema references.
func AddArrayJSONResponse(m *RouteMetadata, statusCode int, description string, itemSchemaRef string) {
	content := map[string]MediaType{
		ContentTypeJSON: {
			Schema: Schema{
				Type: "array",
				Items: &Schema{
					Ref: "#/components/schemas/" + itemSchemaRef,
				},
			},
		},
	}
	AddResponse(m, statusCode, description, content)
}

// AddJSONRequestBody adds a JSON request body to the route metadata.
func AddJSONRequestBody(m *RouteMetadata, description string, required bool, schema Schema) {
	m.RequestBody = &RequestBody{
		Description: description,
		Required:    required,
		Content: map[string]MediaType{
			ContentTypeJSON: {Schema: schema},
		},
	}
}
