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
	varyByPath    bool
	varyByQuery   []string
	varyByHeaders []string
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
