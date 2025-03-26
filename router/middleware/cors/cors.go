package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/go-router/router"
)

// Options defines the configuration options for CORS middleware.
type Options struct {
	AllowOrigins       []string // List of allowed origins. Use "*" to allow all.
	AllowMethods       []string // List of allowed HTTP methods.
	AllowHeaders       []string // List of allowed headers. Use "*" to allow all.
	ExposeHeaders      []string // List of headers that are safe to expose.
	AllowCredentials   bool     // Allow cookies and credentials.
	MaxAge             int      // How long (seconds) preflight results can be cached.
	OptionsPassthrough bool     // Forward OPTIONS requests to handlers.
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

// Default returns a middleware that handles CORS with default options.
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

	// Precompute header values
	allowMethods := strings.Join(options.AllowMethods, ", ")
	allowHeaders := strings.Join(options.AllowHeaders, ", ")
	exposeHeaders := strings.Join(options.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(options.MaxAge)
	if options.MaxAge <= 0 {
		maxAge = ""
	}

	// Check if wildcard is allowed
	allowWildcard := contains(options.AllowOrigins, "*")

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			// Always set Vary header
			c.SetHeader("Vary", "Origin")

			origin := c.GetHeader("Origin")
			isPreflight := c.Request.Method == http.MethodOptions

			// Skip if no origin and not preflight
			if origin == "" && !isPreflight {
				next(c)
				return
			}

			// Determine allowed origin
			allowOrigin := ""
			if origin != "" {
				allowOrigin = getAllowOrigin(origin, options.AllowOrigins)
			} else if isPreflight && allowWildcard {
				allowOrigin = "*"
			} else if isPreflight {
				allowOrigin = "null"
			}

			// Process preflight requests
			if isPreflight {
				// Set CORS headers
				c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
				c.SetHeader("Access-Control-Allow-Methods", allowMethods)

				// Handle headers
				reqHeaders := c.GetHeader("Access-Control-Request-Headers")
				if len(options.AllowHeaders) > 0 && options.AllowHeaders[0] == "*" && reqHeaders != "" {
					c.SetHeader("Access-Control-Allow-Headers", reqHeaders)
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

				if exposeHeaders != "" {
					c.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
				}

				if options.AllowCredentials {
					c.SetHeader("Access-Control-Allow-Credentials", "true")
				}
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

				if strings.HasPrefix(origin, protocol) &&
					strings.HasSuffix(origin, domain) &&
					len(origin) > len(protocol)+len(domain) {
					return origin
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
