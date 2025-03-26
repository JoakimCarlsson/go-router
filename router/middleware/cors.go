package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/go-router/router"
)

// CORSConfig defines the configuration options for CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is ["*"]
	AllowOrigins []string

	// AllowMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH)
	AllowMethods []string

	// AllowHeaders is a list of non-simple headers the client is allowed to use with
	// cross-domain requests. If the special "*" value is present, all headers will be allowed.
	// Default value is [] which means that no custom headers are allowed.
	AllowHeaders []string

	// ExposeHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification. Default value is [] which means no headers are exposed.
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	// Default value is false.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. Default value is 0 which means no caching.
	MaxAge int

	// OptionsPassthrough instructs the middleware to forward the OPTIONS request to
	// the next handler in the chain if no CORS headers are to be added.
	// Default value is false.
	OptionsPassthrough bool
}

// DefaultCORSConfig returns a configuration with sensible defaults.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:       []string{"*"},
		AllowMethods:       []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodPatch},
		AllowHeaders:       []string{},
		ExposeHeaders:      []string{},
		AllowCredentials:   false,
		MaxAge:             0,
		OptionsPassthrough: false,
	}
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
func CORS() router.MiddlewareFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with configuration.
func CORSWithConfig(config CORSConfig) router.MiddlewareFunc {
	// Normalize configurations
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = []string{"*"}
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodPatch}
	}

	// Convert config to header value strings
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := ""
	if config.MaxAge > 0 {
		maxAge = strconv.Itoa(config.MaxAge)
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			origin := c.GetHeader("Origin")

			// Skip if request has no Origin header
			if origin == "" {
				next(c)
				return
			}

			// Handle preflight requests
			if c.Request.Method == http.MethodOptions {
				// Set preflight response headers
				c.SetHeader("Access-Control-Allow-Origin", getAllowOrigin(origin, config.AllowOrigins))

				if allowMethods != "" {
					c.SetHeader("Access-Control-Allow-Methods", allowMethods)
				}

				if len(config.AllowHeaders) > 0 {
					c.SetHeader("Access-Control-Allow-Headers", allowHeaders)
				} else {
					reqHeaders := c.GetHeader("Access-Control-Request-Headers")
					if reqHeaders != "" {
						c.SetHeader("Access-Control-Allow-Headers", reqHeaders)
					}
				}

				if config.AllowCredentials {
					c.SetHeader("Access-Control-Allow-Credentials", "true")
				}

				if maxAge != "" {
					c.SetHeader("Access-Control-Max-Age", maxAge)
				}

				if !config.OptionsPassthrough {
					c.Status(http.StatusNoContent)
					return
				}
			}

			// Set response headers for actual request
			c.SetHeader("Access-Control-Allow-Origin", getAllowOrigin(origin, config.AllowOrigins))

			if exposeHeaders != "" {
				c.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
			}

			if config.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			}

			next(c)
		}
	}
}

// getAllowOrigin returns the allowed origin based on the Origin header and configuration
func getAllowOrigin(origin string, allowOrigins []string) string {
	// If "*" is specified, return "*" or origin based on whether credentials are allowed
	for _, allowOrigin := range allowOrigins {
		if allowOrigin == "*" {
			return "*"
		}
		if allowOrigin == origin {
			return origin
		}
		// Handle wildcards in origins like "https://*.example.com"
		if strings.HasPrefix(allowOrigin, "*") {
			if strings.HasSuffix(origin, allowOrigin[1:]) {
				return origin
			}
		}
	}

	return ""
}
