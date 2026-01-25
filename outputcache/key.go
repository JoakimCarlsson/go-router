package outputcache

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
)

// CacheKeyGenerator generates cache keys based on request attributes.
type CacheKeyGenerator struct {
	varyByPath      bool
	varyByQuery     []string
	varyByHeaders   []string
	customFunc      func(*http.Request) string
	tags            []string
	cacheWhenFunc   func(int, http.Header) bool
	slidingExp      bool
	withRevalidation bool
}

// newCacheKeyGenerator creates a new cache key generator with the given options.
func newCacheKeyGenerator(opts []interface{}) *CacheKeyGenerator {
	gen := &CacheKeyGenerator{
		varyByPath: false,
	}
	
	for _, opt := range opts {
		if fn, ok := opt.(func(*CacheKeyGenerator)); ok {
			fn(gen)
		}
	}
	
	return gen
}

// GenerateKey generates a cache key for the given request.
func (g *CacheKeyGenerator) GenerateKey(r *http.Request) string {
	var parts []string

	parts = append(parts, r.Method)
	parts = append(parts, r.URL.Path)

	if g.varyByPath {
	}

	if len(g.varyByQuery) > 0 {
		query := r.URL.Query()
		queryParts := make([]string, 0, len(g.varyByQuery))

		for _, param := range g.varyByQuery {
			if values := query[param]; len(values) > 0 {
				sortedValues := make([]string, len(values))
				copy(sortedValues, values)
				sort.Strings(sortedValues)
				queryParts = append(queryParts, param+"="+strings.Join(sortedValues, ","))
			}
		}

		if len(queryParts) > 0 {
			sort.Strings(queryParts)
			parts = append(parts, "q:"+strings.Join(queryParts, "&"))
		}
	}

	if len(g.varyByHeaders) > 0 {
		headerParts := make([]string, 0, len(g.varyByHeaders))

		for _, header := range g.varyByHeaders {
			if value := r.Header.Get(header); value != "" {
				headerParts = append(headerParts, header+":"+value)
			}
		}

		if len(headerParts) > 0 {
			sort.Strings(headerParts)
			parts = append(parts, "h:"+strings.Join(headerParts, "|"))
		}
	}

	if g.customFunc != nil {
		customValue := g.customFunc(r)
		if customValue != "" {
			parts = append(parts, "c:"+customValue)
		}
	}

	combined := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VaryByPath returns an option that includes path parameters in the cache key.
// This is useful for routes with path parameters like /users/{id}.
func VaryByPath() interface{} {
	return func(g *CacheKeyGenerator) {
		g.varyByPath = true
	}
}

// VaryByQuery returns an option that includes specific query parameters in the cache key.
// Only the specified parameters will affect caching.
//
// Example:
//
//	VaryByQuery("q", "page", "sort")
func VaryByQuery(params ...string) interface{} {
	return func(g *CacheKeyGenerator) {
		g.varyByQuery = append(g.varyByQuery, params...)
	}
}

// VaryByHeader returns an option that includes specific request headers in the cache key.
// This is useful for varying cache based on Accept-Language, Accept-Encoding, etc.
//
// Example:
//
//	VaryByHeader("Accept-Language", "Accept-Encoding")
func VaryByHeader(headers ...string) interface{} {
	return func(g *CacheKeyGenerator) {
		for _, h := range headers {
			g.varyByHeaders = append(g.varyByHeaders, http.CanonicalHeaderKey(h))
		}
	}
}

// VaryByCustom returns an option that uses a custom function to generate part of the cache key.
// This is useful for multi-tenant applications, user roles, or custom business logic.
//
// Example:
//
//	VaryByCustom(func(r *http.Request) string {
//	    return getUserRole(r) + ":" + getTenant(r)
//	})
func VaryByCustom(fn func(*http.Request) string) interface{} {
	return func(g *CacheKeyGenerator) {
		g.customFunc = fn
	}
}

// Tags returns an option that associates cache tags with the cached response.
// Tags allow for group-based cache invalidation.
//
// Example:
//
//	Tags("products", "product:123")
func Tags(tags ...string) interface{} {
	return func(g *CacheKeyGenerator) {
		g.tags = append(g.tags, tags...)
	}
}

// SlidingExpiration returns an option that enables sliding expiration for the cached response.
// The TTL is extended on each cache hit.
func SlidingExpiration() interface{} {
	return func(g *CacheKeyGenerator) {
		g.slidingExp = true
	}
}

// WithRevalidation returns an option that enables ETag-based cache revalidation.
// Supports 304 Not Modified responses for bandwidth savings.
func WithRevalidation() interface{} {
	return func(g *CacheKeyGenerator) {
		g.withRevalidation = true
	}
}

// CacheWhen returns an option that conditionally caches responses based on status and headers.
//
// Example:
//
//	CacheWhen(func(status int, headers http.Header) bool {
//	    return status == 200 && headers.Get("X-No-Cache") == ""
//	})
func CacheWhen(fn func(int, http.Header) bool) interface{} {
	return func(g *CacheKeyGenerator) {
		g.cacheWhenFunc = fn
	}
}

// VaryByEncoding returns an option that varies the cache by Accept-Encoding header.
// This ensures separate cache entries for different compression formats (gzip, br, etc).
func VaryByEncoding() interface{} {
	return VaryByHeader("Accept-Encoding")
}
