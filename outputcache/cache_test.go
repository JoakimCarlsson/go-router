package outputcache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

func TestCache_Basic(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/test", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(time.Minute)

	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", callCount)
	}

	if rec1.Header().Get("X-Cache") == "HIT" {
		t.Error("Expected cache miss on first request")
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once total, was called %d times", callCount)
	}

	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Error("Expected cache hit on second request")
	}

	if rec1.Body.String() != rec2.Body.String() {
		t.Error("Expected same response body from cache")
	}
}

func TestCache_NoCacheForPost(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.POST("/test", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(time.Minute)

	req1 := httptest.NewRequest("POST", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest("POST", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice for POST, was called %d times", callCount)
	}
}

func TestCache_NoCacheWithoutConfig(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/test", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"message": "test"})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice without cache config, was called %d times", callCount)
	}
}

func TestCache_Expiration(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/test", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(50 * time.Millisecond)

	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", callCount)
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to still be called once, was called %d times", callCount)
	}

	time.Sleep(100 * time.Millisecond)

	req3 := httptest.NewRequest("GET", "/test", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice after expiration, was called %d times", callCount)
	}
}

func TestCache_VaryByQuery(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/search", func(c *router.Context) {
		callCount++
		query := c.Query().Get("q")
		c.JSON(200, map[string]string{"query": query})
	}).WithOutputCache(time.Minute, VaryByQuery("q"))

	req1 := httptest.NewRequest("GET", "/search?q=test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", callCount)
	}

	req2 := httptest.NewRequest("GET", "/search?q=test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to still be called once, was called %d times", callCount)
	}

	req3 := httptest.NewRequest("GET", "/search?q=other", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice for different query, was called %d times", callCount)
	}
}

func TestCache_Clear(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/test", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(time.Minute)

	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once before clear, was called %d times", callCount)
	}

	cache.Clear()

	req3 := httptest.NewRequest("GET", "/test", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice after clear, was called %d times", callCount)
	}
}

func TestConfig_ShouldCacheStatus(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		status      int
		shouldCache bool
	}{
		{"200 OK default", DefaultConfig(), 200, true},
		{"201 Created default", DefaultConfig(), 201, true},
		{"204 No Content default", DefaultConfig(), 204, true},
		{"400 Bad Request default", DefaultConfig(), 400, false},
		{"404 Not Found default", DefaultConfig(), 404, false},
		{"500 Server Error default", DefaultConfig(), 500, false},
		{"200 OK with OnlyStatus", Config{OnlyStatus: []int{200, 404}}, 200, true},
		{"404 with OnlyStatus", Config{OnlyStatus: []int{200, 404}}, 404, true},
		{"201 with OnlyStatus", Config{OnlyStatus: []int{200, 404}}, 201, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.shouldCacheStatus(tt.status)
			if result != tt.shouldCache {
				t.Errorf("shouldCacheStatus(%d) = %v, want %v", tt.status, result, tt.shouldCache)
			}
		})
	}
}

func TestWithRouter(t *testing.T) {
	r := router.New()
	middleware := WithRouter(r)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rtr := req.Context().Value(cacheRouterKey)
		if rtr == nil {
			t.Error("Expected router in context")
			return
		}
		if rtr != r {
			t.Error("Expected correct router in context")
		}
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)
}

func TestCache_WithoutRouterContext(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/test", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(time.Minute)

	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice without router context, was called %d times", callCount)
	}
}

func BenchmarkCache_Hit(b *testing.B) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	r.GET("/test", func(c *router.Context) {
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(time.Minute)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkCache_Miss(b *testing.B) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	r.GET("/test", func(c *router.Context) {
		c.JSON(200, map[string]string{"message": "test"})
	}).WithOutputCache(time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?n="+string(rune(i)), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}
