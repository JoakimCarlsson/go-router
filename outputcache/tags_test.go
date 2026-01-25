package outputcache

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

func TestCache_Tags(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/products/{id}", func(c *router.Context) {
		callCount++
		id := c.Param("id")
		c.JSON(200, map[string]string{"id": id})
	}).WithOutputCache(time.Minute, Tags("products", "product:123"))

	req := httptest.NewRequest("GET", "/products/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if callCount != 1 {
		t.Fatalf("Expected handler to be called once, was called %d times", callCount)
	}

	if rec.Header().Get("X-Cache") == "HIT" {
		t.Fatal("First request should not be a cache hit")
	}

	req2 := httptest.NewRequest("GET", "/products/123", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("Second request should be a cache hit, got X-Cache=%s", rec2.Header().Get("X-Cache"))
	}

	if callCount != 1 {
		t.Errorf("Expected handler to still be called once, was called %d times", callCount)
	}

	cache.InvalidateTag("products")

	req3 := httptest.NewRequest("GET", "/products/123", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice after tag invalidation, was called %d times", callCount)
	}
}

func TestCache_InvalidateTags(t *testing.T) {
	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount1 := 0
	r.GET("/products", func(c *router.Context) {
		callCount1++
		c.JSON(200, map[string]string{"type": "products"})
	}).WithOutputCache(time.Minute, Tags("products"))

	callCount2 := 0
	r.GET("/users", func(c *router.Context) {
		callCount2++
		c.JSON(200, map[string]string{"type": "users"})
	}).WithOutputCache(time.Minute, Tags("users"))

	req1 := httptest.NewRequest("GET", "/products", nil)
	r.ServeHTTP(httptest.NewRecorder(), req1)

	req2 := httptest.NewRequest("GET", "/users", nil)
	r.ServeHTTP(httptest.NewRecorder(), req2)

	if callCount1 != 1 || callCount2 != 1 {
		t.Fatalf("Expected both handlers to be called once")
	}

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/products", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/users", nil))

	if callCount1 != 1 || callCount2 != 1 {
		t.Fatalf("Expected both handlers to still be called once (cache hit)")
	}

	cache.InvalidateTags("products", "users")

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/products", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/users", nil))

	if callCount1 != 2 || callCount2 != 2 {
		t.Errorf("Expected both handlers to be called twice after invalidation, got %d and %d", callCount1, callCount2)
	}
}

func TestMemoryStorage_TagIndex(t *testing.T) {
	storage := NewMemoryStorage()

	response1 := &CachedResponse{
		StatusCode: 200,
		Body:       []byte("product 1"),
		CachedAt:   time.Now(),
		Tags:       []string{"products", "product:1"},
	}

	response2 := &CachedResponse{
		StatusCode: 200,
		Body:       []byte("product 2"),
		CachedAt:   time.Now(),
		Tags:       []string{"products", "product:2"},
	}

	storage.Set("key1", response1, time.Minute)
	storage.Set("key2", response2, time.Minute)

	if _, ok := storage.Get("key1"); !ok {
		t.Fatal("Expected key1 to exist")
	}
	if _, ok := storage.Get("key2"); !ok {
		t.Fatal("Expected key2 to exist")
	}

	storage.InvalidateTag("products")

	if _, ok := storage.Get("key1"); ok {
		t.Error("Expected key1 to be invalidated")
	}
	if _, ok := storage.Get("key2"); ok {
		t.Error("Expected key2 to be invalidated")
	}
}
