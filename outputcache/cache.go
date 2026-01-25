package outputcache

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

// contextKey is used to store cache-related data in request context.
type contextKey string

const (
	cacheRouterKey contextKey = "outputcache:router"
)

// Cache manages HTTP response caching with configurable storage and policies.
type Cache struct {
	config  Config
	storage Storage
	cleanup func()
}

// New creates a new Cache instance with the given configuration.
// If config.Storage is nil, a new MemoryStorage instance will be created.
func New(config Config) *Cache {
	if config.Storage == nil {
		config.Storage = NewMemoryStorage()
	}
	
	if config.DefaultDuration == 0 {
		config.DefaultDuration = 5 * time.Minute
	}
	
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Minute
	}
	
	if len(config.ExcludeMethods) == 0 {
		config.ExcludeMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	}
	
	cache := &Cache{
		config:  config,
		storage: config.Storage,
	}

	if ms, ok := config.Storage.(*MemoryStorage); ok {
		cache.cleanup = ms.StartCleanup(config.CleanupInterval)
	}

	return cache
}

// Middleware returns an HTTP middleware that handles output caching.
// It must be registered before routes are defined to properly capture the router reference.
func (c *Cache) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !c.config.shouldCache(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			var rtr *router.Router
			if routerVal := r.Context().Value(cacheRouterKey); routerVal != nil {
				rtr = routerVal.(*router.Router)
			}

			var cacheConfig *router.CacheConfig
			if rtr != nil {
				cacheConfig = rtr.GetCacheConfig(r.Method, r.URL.Path)
			}

			if cacheConfig == nil {
				next.ServeHTTP(w, r)
				return
			}

			keyGen := newCacheKeyGenerator(cacheConfig.Options)
			cacheKey := keyGen.GenerateKey(r)

			if cached, ok := c.storage.Get(cacheKey); ok {
				c.serveCachedResponse(w, cached)
				return
			}

			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			if c.config.shouldCacheStatus(recorder.statusCode) {
				cached := &CachedResponse{
					StatusCode: recorder.statusCode,
					Headers:    recorder.Header().Clone(),
					Body:       recorder.body.Bytes(),
					CachedAt:   time.Now(),
				}

				duration := cacheConfig.Duration
				if duration == 0 {
					duration = c.config.DefaultDuration
				}

				c.storage.Set(cacheKey, cached, duration)
			}
		})
	}
}

// serveCachedResponse writes a cached response to the ResponseWriter.
func (c *Cache) serveCachedResponse(w http.ResponseWriter, cached *CachedResponse) {
	for key, values := range cached.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("X-Cache", "HIT")
	age := time.Since(cached.CachedAt).Seconds()
	w.Header().Set("Age", string(rune(int(age))))

	w.WriteHeader(cached.StatusCode)
	_, _ = w.Write(cached.Body)
}

// Clear removes all cached responses.
func (c *Cache) Clear() {
	c.storage.Clear()
}

// Close stops any background cleanup goroutines.
func (c *Cache) Close() {
	if c.cleanup != nil {
		c.cleanup()
	}
}

// WithRouter returns an HTTP middleware that injects the router into the request context.
// This must be used before the Cache middleware for the cache to access route configurations.
func WithRouter(r *router.Router) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), cacheRouterKey, r)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// responseRecorder captures HTTP responses for caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) ReadFrom(src io.Reader) (int64, error) {
	buf := &bytes.Buffer{}
	n, err := io.Copy(io.MultiWriter(r.ResponseWriter, buf), src)
	r.body.Write(buf.Bytes())
	return n, err
}
