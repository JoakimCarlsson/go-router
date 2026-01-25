package openapi

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// StatusCodeToString converts an HTTP status code to its string representation.
func StatusCodeToString(code int) string {
	return strconv.Itoa(code)
}

// StatusCodeFromString converts a string to an HTTP status code.
func StatusCodeFromString(code string) (int, error) {
	return strconv.Atoi(code)
}

// Content type constants for HTTP responses.
const (
	ContentTypeJSON        = "application/json"
	ContentTypeXML         = "application/xml"
	ContentTypeEventStream = "text/event-stream"
	ContentTypeHTML        = "text/html"
	ContentTypeFormData    = "multipart/form-data"
)

// RouteMetadata contains all metadata for a route used in OpenAPI generation.
type RouteMetadata struct {
	Method          string
	Path            string
	OperationID     string
	Summary         string
	Description     string
	Tags            []string
	Deprecated      bool
	Parameters      []Parameter
	RequestBody     *RequestBody
	Responses       map[string]Response
	Security        []SecurityRequirement
	IsSSE           bool
	SSEEvents       []SSEEventSchema
	ExcludeFromDocs bool
}

// SSEEventSchema defines the schema for a Server-Sent Event type.
type SSEEventSchema struct {
	EventName   string
	Description string
	Schema      Schema
}

// Parameter represents an OpenAPI parameter definition.
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      Schema      `json:"schema"`
	Example     interface{} `json:"example,omitempty"`
}

func (p Parameter) MarshalJSON() ([]byte, error) {
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

// RequestBody represents an OpenAPI request body definition.
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents an OpenAPI response definition.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

// SecurityRequirement represents an OpenAPI security requirement.
type SecurityRequirement map[string][]string

// Example represents an OpenAPI example.
type Example struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Value       interface{} `json:"value,omitempty"`
}

// MediaType represents an OpenAPI media type definition.
type MediaType struct {
	Schema    Schema             `json:"schema,omitempty"`
	Example   interface{}        `json:"example,omitempty"`
	Examples  map[string]Example `json:"examples,omitempty"`
	SchemaRef *Reference         `json:"-"`
}

func (m MediaType) MarshalJSON() ([]byte, error) {
	if m.SchemaRef != nil {
		return json.Marshal(struct {
			Schema   *Reference         `json:"schema"`
			Example  interface{}        `json:"example,omitempty"`
			Examples map[string]Example `json:"examples,omitempty"`
		}{
			Schema:   m.SchemaRef,
			Example:  m.Example,
			Examples: m.Examples,
		})
	}
	return json.Marshal(struct {
		Schema   Schema             `json:"schema,omitempty"`
		Example  interface{}        `json:"example,omitempty"`
		Examples map[string]Example `json:"examples,omitempty"`
	}{
		Schema:   m.Schema,
		Example:  m.Example,
		Examples: m.Examples,
	})
}

// Header represents an OpenAPI header definition.
type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

// Schema represents an OpenAPI schema definition.
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

// Reference represents an OpenAPI reference to another schema.
type Reference struct {
	Ref string `json:"$ref"`
}

// SchemaOrReference represents either an inline schema or a reference.
type SchemaOrReference struct {
	Schema    *Schema
	Reference *Reference
}

func (s SchemaOrReference) MarshalJSON() ([]byte, error) {
	if s.Reference != nil {
		return json.Marshal(s.Reference)
	}
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	return json.Marshal(nil)
}

// Spec represents a complete OpenAPI 3.0 specification.
type Spec struct {
	OpenAPI      string              `json:"openapi"`
	Info         Info                `json:"info"`
	Servers      []Server            `json:"servers,omitempty"`
	Paths        map[string]PathItem `json:"paths"`
	Components   *Components         `json:"components,omitempty"`
	Tags         []Tag               `json:"tags,omitempty"`
	ExternalDocs map[string]string   `json:"externalDocs,omitempty"`
}

// Info represents the OpenAPI info object with API metadata.
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Version        string   `json:"version"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Contact represents contact information for the API.
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License represents license information for the API.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents a server URL for the API.
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a variable for server URL template substitution.
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem represents operations available on a single path.
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

// Operation represents a single API operation on a path.
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

// Components holds a set of reusable objects for the OpenAPI specification.
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents an OpenAPI security scheme definition.
type SecurityScheme struct {
	Type             string      `json:"type"`
	Scheme           string      `json:"scheme,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Description      string      `json:"description,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents the configuration for OAuth2 flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents configuration for a specific OAuth2 flow.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Tag represents an OpenAPI tag for grouping operations.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// TypeHandler is a function that generates a Schema for a custom Go type.
type TypeHandler func(reflect.Type) Schema

type typeHandlerRegistry struct {
	handlers map[string]TypeHandler
	mu       sync.RWMutex
}

var globalTypeHandlerRegistry = &typeHandlerRegistry{
	handlers: make(map[string]TypeHandler),
}

// RegisterTypeHandler registers a custom type handler for schema generation.
func RegisterTypeHandler(typeName string, handler TypeHandler) {
	globalTypeHandlerRegistry.mu.Lock()
	defer globalTypeHandlerRegistry.mu.Unlock()
	globalTypeHandlerRegistry.handlers[typeName] = handler
}

// GetTypeHandler retrieves a custom type handler by type name.
func GetTypeHandler(typeName string) (TypeHandler, bool) {
	globalTypeHandlerRegistry.mu.RLock()
	defer globalTypeHandlerRegistry.mu.RUnlock()
	handler, exists := globalTypeHandlerRegistry.handlers[typeName]
	return handler, exists
}

// TypeRegistryEntry represents a registered type in the global type registry.
type TypeRegistryEntry struct {
	Name      string
	PkgPath   string
	Count     int
	FinalName string
}

type typeRegistry struct {
	types map[string]*TypeRegistryEntry
	mu    sync.RWMutex
}

var globalTypeRegistry = &typeRegistry{
	types: make(map[string]*TypeRegistryEntry),
}

// RegisterType registers a Go type and returns its schema name.
// It handles name collisions by qualifying names with package paths.
func RegisterType(t reflect.Type) string {
	globalTypeRegistry.mu.Lock()
	defer globalTypeRegistry.mu.Unlock()

	name := t.Name()
	pkgPath := t.PkgPath()
	fullID := pkgPath + "." + name

	if entry, exists := globalTypeRegistry.types[fullID]; exists {
		entry.Count++
		return entry.FinalName
	}

	if entry, exists := globalTypeRegistry.types[name]; exists {
		if entry.Count == 1 && entry.FinalName == name {
			origFullID := entry.PkgPath + "." + entry.Name
			origQualifiedName := SanitizeSchemaName(entry.PkgPath + "_" + entry.Name)
			entry.FinalName = origQualifiedName
			globalTypeRegistry.types[origFullID] = entry
			delete(globalTypeRegistry.types, name)
		}

		qualifiedName := SanitizeSchemaName(pkgPath + "_" + name)
		globalTypeRegistry.types[fullID] = &TypeRegistryEntry{
			Name:      name,
			PkgPath:   pkgPath,
			Count:     1,
			FinalName: qualifiedName,
		}
		return qualifiedName
	}

	globalTypeRegistry.types[name] = &TypeRegistryEntry{
		Name:      name,
		PkgPath:   pkgPath,
		Count:     1,
		FinalName: name,
	}
	globalTypeRegistry.types[fullID] = globalTypeRegistry.types[name]
	return name
}

// SanitizeSchemaName converts a type name to a valid OpenAPI schema name.
func SanitizeSchemaName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// OAuth2Config holds OAuth2 configuration for the OpenAPI specification.
type OAuth2Config struct {
	ClientID                                  string
	ClientSecret                              string
	Realm                                     string
	AppName                                   string
	ScopeSeparator                            string
	Scopes                                    []string
	AdditionalQueryParams                     map[string]string
	UseBasicAuthenticationWithAccessCodeGrant bool
	UsePkceWithAuthorizationCodeGrant         bool
	OAuth2RedirectUrl                         string
}

// NewOAuth2Config creates a new OAuth2Config with default values.
func NewOAuth2Config() *OAuth2Config {
	return &OAuth2Config{
		ScopeSeparator:                            " ",
		AdditionalQueryParams:                     make(map[string]string),
		UseBasicAuthenticationWithAccessCodeGrant: false,
		UsePkceWithAuthorizationCodeGrant:         true,
	}
}

func (c *OAuth2Config) WithClientID(clientID string) *OAuth2Config {
	c.ClientID = clientID
	return c
}

func (c *OAuth2Config) WithClientSecret(clientSecret string) *OAuth2Config {
	c.ClientSecret = clientSecret
	return c
}

func (c *OAuth2Config) WithRealm(realm string) *OAuth2Config {
	c.Realm = realm
	return c
}

func (c *OAuth2Config) WithAppName(appName string) *OAuth2Config {
	c.AppName = appName
	return c
}

func (c *OAuth2Config) WithScopeSeparator(separator string) *OAuth2Config {
	c.ScopeSeparator = separator
	return c
}

func (c *OAuth2Config) WithScopes(scopes ...string) *OAuth2Config {
	c.Scopes = scopes
	return c
}

func (c *OAuth2Config) WithAdditionalQueryParam(key, value string) *OAuth2Config {
	c.AdditionalQueryParams[key] = value
	return c
}

func (c *OAuth2Config) WithBasicAuthentication(use bool) *OAuth2Config {
	c.UseBasicAuthenticationWithAccessCodeGrant = use
	return c
}

func (c *OAuth2Config) WithPKCE(use bool) *OAuth2Config {
	c.UsePkceWithAuthorizationCodeGrant = use
	return c
}

func (c *OAuth2Config) WithOAuth2RedirectUrl(url string) *OAuth2Config {
	c.OAuth2RedirectUrl = url
	return c
}

// EnsureResponsesMap initializes the Responses map if it is nil.
func EnsureResponsesMap(m *RouteMetadata) {
	if m.Responses == nil {
		m.Responses = make(map[string]Response)
	}
}

func AddResponse(m *RouteMetadata, statusCode int, description string, content map[string]MediaType) {
	EnsureResponsesMap(m)
	m.Responses[StatusCodeToString(statusCode)] = Response{
		Description: description,
		Content:     content,
	}
}

func AddSimpleResponse(m *RouteMetadata, statusCode int, description string) {
	AddResponse(m, statusCode, description, nil)
}

func AddJSONResponse(m *RouteMetadata, statusCode int, description string, schema Schema) {
	content := map[string]MediaType{
		ContentTypeJSON: {Schema: schema},
	}
	AddResponse(m, statusCode, description, content)
}

func AddJSONResponseWithRef(m *RouteMetadata, statusCode int, description string, ref *Reference) {
	content := map[string]MediaType{
		ContentTypeJSON: {SchemaRef: ref},
	}
	AddResponse(m, statusCode, description, content)
}

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

func AddParameterWithSchema(m *RouteMetadata, name, in string, required bool, description string, schema Schema) {
	m.Parameters = append(m.Parameters, Parameter{
		Name:        name,
		In:          in,
		Required:    required,
		Description: description,
		Schema:      schema,
	})
}

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

func AddJSONRequestBody(m *RouteMetadata, description string, required bool, schema Schema) {
	m.RequestBody = &RequestBody{
		Description: description,
		Required:    required,
		Content: map[string]MediaType{
			ContentTypeJSON: {Schema: schema},
		},
	}
}
