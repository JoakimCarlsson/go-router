// Package cors provides Cross-Origin Resource Sharing (CORS) middleware for go-router.
//
// # Usage
//
//	import (
//	    "github.com/joakimcarlsson/go-router/router"
//	    "github.com/joakimcarlsson/go-router/router/middleware/cors"
//	)
//
//	func main() {
//	    r := router.New()
//
//	    // Use default CORS settings (allow all origins)
//	    r.Use(cors.Default())
//
//	    // Or use custom CORS configuration
//	    r.Use(cors.Handler(cors.Options{
//	        AllowOrigins:     []string{"https://example.com", "https://*.trusted-domain.com"},
//	        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut},
//	        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
//	        ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
//	        AllowCredentials: true,
//	        MaxAge:           86400, // Cache preflight response for 24 hours
//	    }))
//
//	    // Different CORS settings for specific route groups
//	    r.Group("/api", func(api *router.Router) {
//	        api.Use(cors.Handler(cors.Options{
//	            AllowOrigins: []string{"https://api.example.com"},
//	            // Other options...
//	        }))
//
//	        // API routes...
//	    })
//	}
//
// # Configuration Options
//
// The Options struct provides the following configuration options:
//
// AllowOrigins []string: List of allowed origins. Can contain wildcards like "https://*.example.com".
// Use "*" to allow all origins. Default value is ["*"].
//
// For domain wildcards like "https://*.example.com":
//   - This matches "https://sub.example.com", "https://nested.sub.example.com", etc.
//   - This does NOT match "https://example.com" itself (the parent domain)
//
// For a pattern like "https://*.admin.example.com":
//   - This matches "https://dashboard.admin.example.com", "https://user.admin.example.com"
//   - This does NOT match "https://admin.example.com" or "https://example.com"
//   - This does NOT match origins on different domains like "https://sub.example.com"
//
// AllowMethods []string: HTTP methods allowed for CORS requests.
// Default value is [GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH].
//
// AllowHeaders []string: HTTP headers clients can use in requests. Use "*" to allow all headers.
// Default value is [] (none).
//
// ExposeHeaders []string: HTTP headers that should be exposed to clients.
// Default value is [] (none).
//
// AllowCredentials bool: Allow cookies and credentials to be sent with requests.
// Default value is false.
//
// MaxAge int: How long (in seconds) the results of a preflight request can be cached.
// Default value is 0 (no caching).
//
// OptionsPassthrough bool: Pass OPTIONS requests to handlers instead of responding automatically.
// Default value is false.
//
// # Security Considerations
//
//  1. Avoid using AllowOrigins: []string{"*"} with AllowCredentials: true as this is
//     not secure and not allowed by browsers.
//
// 2. Specify exact origins rather than using wildcard "*" in production environments.
//
// 3. Only expose headers that are necessary for your client applications.
//
// 4. Limit allowed methods to only those your API actually supports.
package cors
