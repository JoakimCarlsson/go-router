// Package security provides security-related HTTP middleware.
package security

import (
	"fmt"
	"net/http"
)

// Config defines security header options.
type Config struct {
	// HSTSMaxAge sets the Strict-Transport-Security max-age directive in seconds.
	// Set to 0 to disable HSTS header. Default: 31536000 (1 year).
	HSTSMaxAge int

	// HSTSIncludeSubdomains includes subdomains in HSTS policy.
	HSTSIncludeSubdomains bool

	// HSTSPreload allows the domain to be included in browser preload lists.
	HSTSPreload bool

	// FrameOptions sets X-Frame-Options header.
	// Common values: "DENY", "SAMEORIGIN". Empty string disables the header.
	FrameOptions string

	// ContentTypeNosniff enables X-Content-Type-Options: nosniff header.
	ContentTypeNosniff bool

	// XSSProtection enables X-XSS-Protection: 1; mode=block header.
	// Note: This header is deprecated in modern browsers but may help older browsers.
	XSSProtection bool

	// ReferrerPolicy sets the Referrer-Policy header.
	// Common values: "no-referrer", "strict-origin-when-cross-origin".
	ReferrerPolicy string

	// ContentSecurityPolicy sets the Content-Security-Policy header.
	// This is application-specific and should be configured per-app.
	ContentSecurityPolicy string

	// PermissionsPolicy sets the Permissions-Policy header (formerly Feature-Policy).
	PermissionsPolicy string
}

// DefaultConfig returns a secure default configuration.
func DefaultConfig() Config {
	return Config{
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
		FrameOptions:          "DENY",
		ContentTypeNosniff:    true,
		XSSProtection:         true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// Headers returns a middleware that sets security headers.
func Headers(config Config) func(http.Handler) http.Handler {
	hstsValue := ""
	if config.HSTSMaxAge > 0 {
		hstsValue = fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
		if config.HSTSIncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		if config.HSTSPreload {
			hstsValue += "; preload"
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hstsValue != "" {
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			if config.ContentTypeNosniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			if config.XSSProtection {
				w.Header().Set("X-XSS-Protection", "1; mode=block")
			}

			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Default returns a middleware with default security headers.
func Default() func(http.Handler) http.Handler {
	return Headers(DefaultConfig())
}
