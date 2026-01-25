package outputcache

import (
	"net/http"
	"time"
)

// Config holds the configuration for the output cache.
type Config struct {
	// DefaultDuration is the default TTL for cached responses.
	// Individual routes can override this with WithOutputCache.
	DefaultDuration time.Duration
	
	// Storage is the cache storage backend to use.
	// If nil, a new MemoryStorage instance will be created.
	Storage Storage
	
	// OnlyStatus specifies which HTTP status codes should be cached.
	// If nil, only 2xx status codes (200-299) are cached.
	OnlyStatus []int
	
	// ExcludeMethods specifies HTTP methods that should never be cached.
	// By default, only GET and HEAD requests are cached.
	ExcludeMethods []string
	
	// CleanupInterval is the interval for cleaning up expired cache entries.
	// Only applies to MemoryStorage. Default is 1 minute.
	CleanupInterval time.Duration
	
	// Profiles contains reusable cache configurations.
	Profiles *Profiles
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DefaultDuration: 5 * time.Minute,
		Storage:         nil,
		OnlyStatus:      nil,
		ExcludeMethods:  []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		CleanupInterval: time.Minute,
	}
}

// shouldCache determines if a request should be cached based on method.
func (c *Config) shouldCache(method string) bool {
	for _, excluded := range c.ExcludeMethods {
		if method == excluded {
			return false
		}
	}
	return true
}

// shouldCacheStatus determines if a response status should be cached.
func (c *Config) shouldCacheStatus(status int) bool {
	if len(c.OnlyStatus) > 0 {
		for _, s := range c.OnlyStatus {
			if s == status {
				return true
			}
		}
		return false
	}

	return status >= 200 && status < 300
}
