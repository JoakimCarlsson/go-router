package metadata

// RouteMetadata contains documentation and configuration for a route
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

// Parameter represents an API parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, path, header, cookie
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      Schema      `json:"schema"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents an API response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

// SecurityRequirement represents security requirements for an operation
type SecurityRequirement map[string][]string

// MediaType represents the structure of request/response content
type MediaType struct {
	Schema  Schema      `json:"schema"`
	Example interface{} `json:"example,omitempty"`
}

// Header represents a response header
type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

// Schema represents a type schema
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

// OAuth2Config holds OAuth2 configuration for API authentication
type OAuth2Config struct {
	// ClientId is the OAuth2 client ID
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

// NewOAuth2Config creates a new OAuth2 configuration with default values
func NewOAuth2Config() *OAuth2Config {
	return &OAuth2Config{
		ScopeSeparator:                            " ",
		AdditionalQueryParams:                     make(map[string]string),
		UseBasicAuthenticationWithAccessCodeGrant: false,
		UsePkceWithAuthorizationCodeGrant:         true,
	}
}

// WithClientID sets the client ID
func (c *OAuth2Config) WithClientID(clientID string) *OAuth2Config {
	c.ClientID = clientID
	return c
}

// WithClientSecret sets the client secret
func (c *OAuth2Config) WithClientSecret(clientSecret string) *OAuth2Config {
	c.ClientSecret = clientSecret
	return c
}

// WithRealm sets the realm
func (c *OAuth2Config) WithRealm(realm string) *OAuth2Config {
	c.Realm = realm
	return c
}

// WithAppName sets the application name
func (c *OAuth2Config) WithAppName(appName string) *OAuth2Config {
	c.AppName = appName
	return c
}

// WithScopeSeparator sets the scope separator
func (c *OAuth2Config) WithScopeSeparator(separator string) *OAuth2Config {
	c.ScopeSeparator = separator
	return c
}

// WithScopes sets the scopes
func (c *OAuth2Config) WithScopes(scopes ...string) *OAuth2Config {
	c.Scopes = scopes
	return c
}

// WithAdditionalQueryParam adds a query parameter to the OAuth2 flow
func (c *OAuth2Config) WithAdditionalQueryParam(key, value string) *OAuth2Config {
	c.AdditionalQueryParams[key] = value
	return c
}

// WithBasicAuthentication sets whether to use basic authentication with access code grant
func (c *OAuth2Config) WithBasicAuthentication(use bool) *OAuth2Config {
	c.UseBasicAuthenticationWithAccessCodeGrant = use
	return c
}

// WithPKCE sets whether to use PKCE with authorization code grant
func (c *OAuth2Config) WithPKCE(use bool) *OAuth2Config {
	c.UsePkceWithAuthorizationCodeGrant = use
	return c
}
