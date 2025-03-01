package router

// NewOAuth2Config creates a basic OAuth2 configuration for Swagger UI
func NewOAuth2Config(clientID string) *OAuth2Config {
	return &OAuth2Config{
		ClientID:                                  clientID,
		AdditionalQueryParams:                     make(map[string]string),
		UsePkceWithAuthorizationCodeGrant:         true,
		UseBasicAuthenticationWithAccessCodeGrant: false,
	}
}

// WithClientSecret sets the client secret for OAuth2
func (c *OAuth2Config) WithClientSecret(clientSecret string) *OAuth2Config {
	c.ClientSecret = clientSecret
	return c
}

// WithRealm sets the realm for OAuth2
func (c *OAuth2Config) WithRealm(realm string) *OAuth2Config {
	c.Realm = realm
	return c
}

// WithAppName sets the application name for OAuth2
func (c *OAuth2Config) WithAppName(appName string) *OAuth2Config {
	c.AppName = appName
	return c
}

// WithScopeSeparator sets the scope separator character (default is space)
func (c *OAuth2Config) WithScopeSeparator(separator string) *OAuth2Config {
	c.ScopeSeparator = separator
	return c
}

// WithScopes sets the predefined scopes for OAuth2
func (c *OAuth2Config) WithScopes(scopes ...string) *OAuth2Config {
	c.Scopes = scopes
	return c
}

// WithAdditionalQueryParam adds a query parameter to the OAuth2 requests
func (c *OAuth2Config) WithAdditionalQueryParam(key, value string) *OAuth2Config {
	if c.AdditionalQueryParams == nil {
		c.AdditionalQueryParams = make(map[string]string)
	}
	c.AdditionalQueryParams[key] = value
	return c
}

// WithBasicAuthentication enables using HTTP Basic authentication for the access code grant flow
func (c *OAuth2Config) WithBasicAuthentication(enabled bool) *OAuth2Config {
	c.UseBasicAuthenticationWithAccessCodeGrant = enabled
	return c
}

// WithPKCE enables or disables Proof Key for Code Exchange for the authorization code grant flow
func (c *OAuth2Config) WithPKCE(enabled bool) *OAuth2Config {
	c.UsePkceWithAuthorizationCodeGrant = enabled
	return c
}
