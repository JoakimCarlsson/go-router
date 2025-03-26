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

			// Skip if request has no Origin header
			if origin == "" {
				next(c)
				return
			}

			// Handle preflight requests
			if c.Request.Method == http.MethodOptions {
				// Set preflight response headers
				c.SetHeader("Access-Control-Allow-Origin", getAllowOrigin(origin, options.AllowOrigins))

				if allowMethods != "" {
					c.SetHeader("Access-Control-Allow-Methods", allowMethods)
				}

				// Only add allow-headers if there are custom headers or if "*" is specified
				if len(options.AllowHeaders) > 0 {
					c.SetHeader("Access-Control-Allow-Headers", allowHeaders)
				} else {
					reqHeaders := c.GetHeader("Access-Control-Request-Headers")
					if reqHeaders != "" {
						c.SetHeader("Access-Control-Allow-Headers", reqHeaders)
					}
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
			c.SetHeader("Access-Control-Allow-Origin", getAllowOrigin(origin, options.AllowOrigins))

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
