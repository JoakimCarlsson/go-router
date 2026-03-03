package outputcache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joakimcarlsson/go-router/router/v2"
)

func TestCache_ETag(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/data", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"data": "test"})
	}).WithOutputCache(time.Minute, WithRevalidation())

	req1 := httptest.NewRequest("GET", "/data", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", callCount)
	}

	etag := rec1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("Expected ETag header to be set")
	}

	req2 := httptest.NewRequest("GET", "/data", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != 304 {
		t.Errorf("Expected 304 Not Modified, got %d", rec2.Code)
	}

	if callCount != 1 {
		t.Errorf("Expected handler to still be called once (304 from cache), was called %d times", callCount)
	}

	if rec2.Body.Len() != 0 {
		t.Error("Expected empty body for 304 response")
	}
}

func TestGenerateETag(t *testing.T) {
	body1 := []byte("test data")
	etag1 := generateETag(body1)

	if etag1 == "" {
		t.Fatal("Expected non-empty ETag")
	}

	body2 := []byte("test data")
	etag2 := generateETag(body2)

	if etag1 != etag2 {
		t.Error("Expected same ETag for identical content")
	}

	body3 := []byte("different data")
	etag3 := generateETag(body3)

	if etag1 == etag3 {
		t.Error("Expected different ETag for different content")
	}
}

func TestShouldRevalidate(t *testing.T) {
	tests := []struct {
		name        string
		ifNoneMatch string
		etag        string
		shouldMatch bool
	}{
		{"exact match", `"abc123"`, `"abc123"`, true},
		{"wildcard", "*", `"abc123"`, true},
		{"no match", `"different"`, `"abc123"`, false},
		{"empty if-none-match", "", `"abc123"`, false},
		{"empty etag", `"abc123"`, "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRevalidate(tt.ifNoneMatch, tt.etag)
			if result != tt.shouldMatch {
				t.Errorf("shouldRevalidate(%q, %q) = %v, want %v", tt.ifNoneMatch, tt.etag, result, tt.shouldMatch)
			}
		})
	}
}

func TestCache_SlidingExpiration(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/data", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"data": "test"})
	}).WithOutputCache(100*time.Millisecond, SlidingExpiration())

	req1 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req1)

	if callCount != 1 {
		t.Fatalf("Expected handler to be called once")
	}

	time.Sleep(60 * time.Millisecond)

	req2 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req2)

	if callCount != 1 {
		t.Fatalf("Expected handler to still be called once (within sliding window)")
	}

	time.Sleep(60 * time.Millisecond)

	req3 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req3)

	if callCount != 1 {
		t.Fatalf("Expected handler to still be called once (sliding extended TTL)")
	}

	time.Sleep(120 * time.Millisecond)

	req4 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req4)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice after full expiration, was called %d times", callCount)
	}
}

func TestCache_VaryByCustom(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/data", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"data": "test"})
	}).WithOutputCache(time.Minute, VaryByCustom(func(req *http.Request) string {
		return req.Header.Get("X-Tenant-ID")
	}))

	req1 := httptest.NewRequest("GET", "/data", nil)
	req1.Header.Set("X-Tenant-ID", "tenant1")
	r.ServeHTTP(httptest.NewRecorder(), req1)

	if callCount != 1 {
		t.Fatalf("Expected handler to be called once")
	}

	req2 := httptest.NewRequest("GET", "/data", nil)
	req2.Header.Set("X-Tenant-ID", "tenant1")
	r.ServeHTTP(httptest.NewRecorder(), req2)

	if callCount != 1 {
		t.Fatalf("Expected handler to still be called once (cache hit for same tenant)")
	}

	req3 := httptest.NewRequest("GET", "/data", nil)
	req3.Header.Set("X-Tenant-ID", "tenant2")
	r.ServeHTTP(httptest.NewRecorder(), req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice (different tenant), was called %d times", callCount)
	}
}

func TestCache_CacheWhen(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	shouldCache := true
	r.GET("/data", func(c *router.Context) {
		callCount++
		if shouldCache {
			c.JSON(200, map[string]string{"data": "test"})
		} else {
			c.SetHeader("X-No-Cache", "true")
			c.JSON(200, map[string]string{"data": "nocache"})
		}
	}).WithOutputCache(time.Minute, CacheWhen(func(status int, headers http.Header) bool {
		return headers.Get("X-No-Cache") == ""
	}))

	req1 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req1)

	if callCount != 1 {
		t.Fatalf("Expected handler to be called once")
	}

	req2 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req2)

	if callCount != 1 {
		t.Fatalf("Expected handler to still be called once (cache hit)")
	}

	cache.Clear()
	shouldCache = false

	req3 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req3)

	if callCount != 2 {
		t.Fatalf("Expected handler to be called twice (miss after clear)")
	}

	req4 := httptest.NewRequest("GET", "/data", nil)
	r.ServeHTTP(httptest.NewRecorder(), req4)

	if callCount != 3 {
		t.Errorf("Expected handler to be called three times (not cached due to condition), was called %d times", callCount)
	}
}

func TestCache_VaryByEncoding(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/data", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"data": "test"})
	}).WithOutputCache(time.Minute, VaryByEncoding())

	req1 := httptest.NewRequest("GET", "/data", nil)
	req1.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(httptest.NewRecorder(), req1)

	if callCount != 1 {
		t.Fatalf("Expected handler to be called once")
	}

	req2 := httptest.NewRequest("GET", "/data", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(httptest.NewRecorder(), req2)

	if callCount != 1 {
		t.Fatalf("Expected handler to still be called once (cache hit for gzip)")
	}

	req3 := httptest.NewRequest("GET", "/data", nil)
	req3.Header.Set("Accept-Encoding", "br")
	r.ServeHTTP(httptest.NewRecorder(), req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice (different encoding), was called %d times", callCount)
	}
}
