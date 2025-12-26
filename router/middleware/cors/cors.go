package cors

import (
	"net/http"
	"strconv"
	"strings"
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
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPatch,
		},
		AllowHeaders:       []string{},
		ExposeHeaders:      []string{},
		AllowCredentials:   false,
		MaxAge:             0,
		OptionsPassthrough: false,
	}
}

// Default returns a middleware that handles CORS with default options.
func Default() func(http.Handler) http.Handler {
	return Handler(DefaultOptions())
}

// Handler returns a CORS middleware with the specified options.
func Handler(options Options) func(http.Handler) http.Handler {
	// Normalize configurations
	if len(options.AllowOrigins) == 0 {
		options.AllowOrigins = []string{"*"}
	}
	if len(options.AllowMethods) == 0 {
		options.AllowMethods = []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPatch,
		}
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

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always set Vary header
			w.Header().Set("Vary", "Origin")

			origin := r.Header.Get("Origin")
			isPreflight := r.Method == http.MethodOptions

			// Skip if no origin and not preflight
			if origin == "" && !isPreflight {
				next.ServeHTTP(w, r)
				return
			}

			// Determine allowed origin
			allowOrigin := ""
			if origin != "" {
				allowOrigin = getAllowOrigin(origin, options.AllowOrigins, options.AllowCredentials)
			} else if isPreflight && allowWildcard {
				allowOrigin = "*"
			} else if isPreflight {
				allowOrigin = "null"
			}

			// Process preflight requests
			if isPreflight {
				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)

				// Handle headers
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if len(options.AllowHeaders) > 0 &&
					options.AllowHeaders[0] == "*" &&
					reqHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				} else if len(options.AllowHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				} else if reqHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				}

				if options.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if maxAge != "" {
					w.Header().Set("Access-Control-Max-Age", maxAge)
				}

				// End preflight request if not passing through
				if !options.OptionsPassthrough {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			// Set response headers for actual request
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)

				if exposeHeaders != "" {
					w.Header().
						Set("Access-Control-Expose-Headers", exposeHeaders)
				}

				if options.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getAllowOrigin returns the allowed origin based on the Origin header and configuration
func getAllowOrigin(origin string, allowOrigins []string, allowCredentials bool) string {
	for _, allowOrigin := range allowOrigins {
		// Exact match
		if allowOrigin == origin {
			return origin
		}

		// Wildcard match (*)
		if allowOrigin == "*" {
			if allowCredentials && origin != "" {
				return origin
			}
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
