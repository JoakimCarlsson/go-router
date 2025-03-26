package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/go-router/router"
)

// Options defines the configuration options for CORS middleware.
type Options struct {
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

// DefaultOptions returns a configuration with sensible defaults.
func DefaultOptions() Options {
	return Options{
		AllowOrigins:       []string{"*"},
		AllowMethods:       []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodPatch},
		AllowHeaders:       []string{},
		ExposeHeaders:      []string{},
		AllowCredentials:   false,
		MaxAge:             0,
		OptionsPassthrough: false,
	}
}

// Default returns a middleware that handles Cross-Origin Resource Sharing with default options.
func Default() router.MiddlewareFunc {
	return Handler(DefaultOptions())
}

// Handler returns a CORS middleware with the specified options.
func Handler(options Options) router.MiddlewareFunc {
	// Normalize configurations
	if len(options.AllowOrigins) == 0 {
		options.AllowOrigins = []string{"*"}
	}
	if len(options.AllowMethods) == 0 {
		options.AllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodPatch}
	}

	// Convert config to header value strings
	allowMethods := strings.Join(options.AllowMethods, ", ")
	allowHeaders := strings.Join(options.AllowHeaders, ", ")
	exposeHeaders := strings.Join(options.ExposeHeaders, ", ")
	maxAge := ""
	if options.MaxAge > 0 {
		maxAge = strconv.Itoa(options.MaxAge)
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			origin := c.GetHeader("Origin")

			// Always set the Vary header
			c.SetHeader("Vary", "Origin")

			// Skip if request has no Origin header, unless it's an OPTIONS request
			// OPTIONS without Origin might still be a preflight sent by browser
			isPreflight := c.Request.Method == http.MethodOptions

			if origin == "" && !isPreflight {
				next(c)
				return
			}

			// Process the allowed origin based on origin header
			// For preflight without origin, we'll use the first allowed origin
			// This ensures consistent behavior
			allowOrigin := ""
			if origin != "" {
				allowOrigin = getAllowOrigin(origin, options.AllowOrigins)
			} else if isPreflight && len(options.AllowOrigins) > 0 {
				// Special case: OPTIONS with no origin or null origin
				// For security, only use "*" if it was explicitly configured
				// Otherwise, set empty to respect the route's restriction
				if contains(options.AllowOrigins, "*") {
					allowOrigin = "*"
				} else {
					// Don't respond with a specific origin for null origins for security reasons
					// but do set methods and headers to allow browser to see what's allowed
					allowOrigin = "null"
				}
			}

			// Handle preflight OPTIONS requests directly
			if isPreflight {
				// Set response headers for preflight request
				c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
				c.SetHeader("Access-Control-Allow-Methods", allowMethods)

				// Handle headers
				reqHeaders := c.GetHeader("Access-Control-Request-Headers")
				if len(options.AllowHeaders) > 0 && options.AllowHeaders[0] == "*" {
					// If wildcard headers, mirror the requested headers
					if reqHeaders != "" {
						c.SetHeader("Access-Control-Allow-Headers", reqHeaders)
					} else {
						c.SetHeader("Access-Control-Allow-Headers", "*")
					}
				} else if len(options.AllowHeaders) > 0 {
					c.SetHeader("Access-Control-Allow-Headers", allowHeaders)
				} else if reqHeaders != "" {
					c.SetHeader("Access-Control-Allow-Headers", reqHeaders)
				}

				if options.AllowCredentials {
					c.SetHeader("Access-Control-Allow-Credentials", "true")
				}

				if maxAge != "" {
					c.SetHeader("Access-Control-Max-Age", maxAge)
				}

				// End preflight request if not passing through
				if !options.OptionsPassthrough {
					c.Status(http.StatusNoContent)
					return
				}
			}

			// Set response headers for actual request
			if origin != "" {
				c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
			}

			if exposeHeaders != "" {
				c.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
			}

			if options.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			}

			next(c)
		}
	}
}

// getAllowOrigin returns the allowed origin based on the Origin header and configuration
func getAllowOrigin(origin string, allowOrigins []string) string {
	for _, allowOrigin := range allowOrigins {
		// Exact match
		if allowOrigin == origin {
			return origin
		}

		// Wildcard match (*)
		if allowOrigin == "*" {
			return "*"
		}

		// Protocol wildcard match (https://*.example.com)
		if strings.Contains(allowOrigin, "://*.") {
			parts := strings.SplitN(allowOrigin, "://*.", 2)
			if len(parts) == 2 {
				protocol := parts[0] + "://"
				domain := "." + parts[1] // Add dot to ensure we match domain boundary

				if strings.HasPrefix(origin, protocol) && strings.HasSuffix(origin, domain) {
					// Check that there's at least one character between protocol and domain
					if len(origin) > len(protocol)+len(domain) {
						return origin
					}
				}
			}
		}
	}

	return ""
}

// contains checks if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
