// Package outputcache provides smart HTTP response caching for go-router applications.
//
// This package implements output caching similar to ASP.NET's output caching system,
// allowing you to cache HTTP responses to improve performance by avoiding expensive
// handler executions for repeated requests.
//
// # Basic Usage
//
// Create a cache instance and register it as middleware:
//
//	cache := outputcache.New(outputcache.Config{
//	    DefaultDuration: 5 * time.Minute,
//	})
//	r.Use(cache.Middleware())
//
// Then enable caching on specific routes:
//
//	r.GET("/products", listProducts).
//	    WithOutputCache(time.Minute)
//
// # Cache Key Variation
//
// Control how cache keys are generated with VaryBy options:
//
//	r.GET("/users/{id}", getUser).
//	    WithOutputCache(time.Hour, outputcache.VaryByPath())
//
//	r.GET("/search", search).
//	    WithOutputCache(30*time.Second,
//	        outputcache.VaryByQuery("q", "page"),
//	        outputcache.VaryByHeader("Accept-Language"))
//
// # Storage Backends
//
// The package supports pluggable storage backends via the Storage interface.
// By default, an in-memory storage is used, but you can implement your own:
//
//	type Storage interface {
//	    Get(key string) (*CachedResponse, bool)
//	    Set(key string, response *CachedResponse, ttl time.Duration)
//	    Delete(key string)
//	}
//
// # Thread Safety
//
// All cache operations are thread-safe and optimized for concurrent access.
package outputcache
