package openapi

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func StatusCodeToString(code int) string {
	return strconv.Itoa(code)
}

func StatusCodeFromString(code string) (int, error) {
	return strconv.Atoi(code)
}

const (
	ContentTypeJSON        = "application/json"
	ContentTypeXML         = "application/xml"
	ContentTypeEventStream = "text/event-stream"
	ContentTypeHTML        = "text/html"
	ContentTypeFormData    = "multipart/form-data"
)

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

type SSEEventSchema struct {
	EventName   string
	Description string
	Schema      Schema
}

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

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

type SecurityRequirement map[string][]string

type Example struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Value       interface{} `json:"value,omitempty"`
}

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

type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

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

type Reference struct {
	Ref string `json:"$ref"`
}

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

type Spec struct {
	OpenAPI      string              `json:"openapi"`
	Info         Info                `json:"info"`
	Servers      []Server            `json:"servers,omitempty"`
	Paths        map[string]PathItem `json:"paths"`
	Components   *Components         `json:"components,omitempty"`
	Tags         []Tag               `json:"tags,omitempty"`
	ExternalDocs map[string]string   `json:"externalDocs,omitempty"`
}

type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Version        string   `json:"version"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

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

type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type             string      `json:"type"`
	Scheme           string      `json:"scheme,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Description      string      `json:"description,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type TypeHandler func(reflect.Type) Schema

type typeHandlerRegistry struct {
	handlers map[string]TypeHandler
	mu       sync.RWMutex
}

var globalTypeHandlerRegistry = &typeHandlerRegistry{
	handlers: make(map[string]TypeHandler),
}

func RegisterTypeHandler(typeName string, handler TypeHandler) {
	globalTypeHandlerRegistry.mu.Lock()
	defer globalTypeHandlerRegistry.mu.Unlock()
	globalTypeHandlerRegistry.handlers[typeName] = handler
}

func GetTypeHandler(typeName string) (TypeHandler, bool) {
	globalTypeHandlerRegistry.mu.RLock()
	defer globalTypeHandlerRegistry.mu.RUnlock()
	handler, exists := globalTypeHandlerRegistry.handlers[typeName]
	return handler, exists
}

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

func SanitizeSchemaName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

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
