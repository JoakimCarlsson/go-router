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
	config   Config
	storage  Storage
	cleanup  func()
	profiles *Profiles
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
		config:   config,
		storage:  config.Storage,
		profiles: config.Profiles,
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
			var profileConfig *ProfileConfig

			if rtr != nil {
				routeMethod := r.Method
				routePath := r.URL.Path

				if method, ok := r.Context().Value(router.RouteMethodKey).(string); ok && method != "" {
					routeMethod = method
				}
				if path, ok := r.Context().Value(router.RoutePathKey).(string); ok && path != "" {
					routePath = path
				}

				cacheConfig = rtr.GetCacheConfig(routeMethod, routePath)

				if cacheConfig == nil {
					profileName := rtr.GetCacheProfile(routeMethod, routePath)
					if profileName != "" && c.profiles != nil {
						profileConfig = c.profiles.Get(profileName)
					}
				}
			}

			if cacheConfig == nil && profileConfig == nil {
				next.ServeHTTP(w, r)
				return
			}

			var duration time.Duration
			var opts []interface{}

			if profileConfig != nil {
				duration = profileConfig.Duration
				opts = profileConfig.Options
			} else {
				duration = cacheConfig.Duration
				opts = cacheConfig.Options
			}

			keyGen := newCacheKeyGenerator(opts)
			cacheKey := keyGen.GenerateKey(r)

			if cached, ok := c.storage.Get(cacheKey); ok {
				if keyGen.withRevalidation && cached.ETag != "" {
					if shouldRevalidate(r.Header.Get("If-None-Match"), cached.ETag) {
						w.WriteHeader(http.StatusNotModified)
						return
					}
				}
				c.serveCachedResponse(w, cached)
				return
			}

			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			if keyGen.cacheWhenFunc != nil {
				if !keyGen.cacheWhenFunc(recorder.statusCode, recorder.Header()) {
					return
				}
			}

			if c.config.shouldCacheStatus(recorder.statusCode) {
				if duration == 0 {
					duration = c.config.DefaultDuration
				}

				body := recorder.body.Bytes()
				etag := ""

				if keyGen.withRevalidation {
					etag = generateETag(body)
					w.Header().Set("ETag", etag)
				}

				cached := &CachedResponse{
					StatusCode:        recorder.statusCode,
					Headers:           recorder.Header().Clone(),
					Body:              body,
					CachedAt:          time.Now(),
					Tags:              keyGen.tags,
					SlidingExpiration: keyGen.slidingExp,
					SlidingDuration:   duration,
					ETag:              etag,
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

	w.WriteHeader(cached.StatusCode)
	_, _ = w.Write(cached.Body)
}

// Clear removes all cached responses.
func (c *Cache) Clear() {
	c.storage.Clear()
}

// InvalidateTag removes all cached responses with the given tag.
func (c *Cache) InvalidateTag(tag string) {
	c.storage.InvalidateTag(tag)
}

// InvalidateTags removes all cached responses with any of the given tags.
func (c *Cache) InvalidateTags(tags ...string) {
	c.storage.InvalidateTags(tags...)
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
