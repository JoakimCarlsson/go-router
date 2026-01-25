package outputcache

import (
	"net/http/httptest"
	"testing"
)

func TestCacheKeyGenerator_Basic(t *testing.T) {
	gen := newCacheKeyGenerator(nil)

	req := httptest.NewRequest("GET", "/products", nil)
	key1 := gen.GenerateKey(req)

	req2 := httptest.NewRequest("GET", "/products", nil)
	key2 := gen.GenerateKey(req2)

	if key1 != key2 {
		t.Errorf("Expected same key for identical requests, got %s and %s", key1, key2)
	}

	req3 := httptest.NewRequest("GET", "/users", nil)
	key3 := gen.GenerateKey(req3)

	if key1 == key3 {
		t.Error("Expected different key for different paths")
	}

	req4 := httptest.NewRequest("POST", "/products", nil)
	key4 := gen.GenerateKey(req4)

	if key1 == key4 {
		t.Error("Expected different key for different methods")
	}
}

func TestCacheKeyGenerator_VaryByQuery(t *testing.T) {
	gen := newCacheKeyGenerator([]interface{}{VaryByQuery("q", "page")})

	req1 := httptest.NewRequest("GET", "/search?q=test&page=1", nil)
	key1 := gen.GenerateKey(req1)

	req2 := httptest.NewRequest("GET", "/search?q=test&page=1", nil)
	key2 := gen.GenerateKey(req2)

	if key1 != key2 {
		t.Error("Expected same key for same query params")
	}

	req3 := httptest.NewRequest("GET", "/search?q=test&page=2", nil)
	key3 := gen.GenerateKey(req3)

	if key1 == key3 {
		t.Error("Expected different key for different query param values")
	}

	req4 := httptest.NewRequest("GET", "/search?page=1&q=test", nil)
	key4 := gen.GenerateKey(req4)

	if key1 != key4 {
		t.Error("Expected same key for query params in different order")
	}

	req5 := httptest.NewRequest("GET", "/search?q=test&page=1&extra=value", nil)
	key5 := gen.GenerateKey(req5)

	if key1 != key5 {
		t.Error("Expected same key when non-specified query params are present")
	}
}

func TestCacheKeyGenerator_VaryByHeader(t *testing.T) {
	gen := newCacheKeyGenerator([]interface{}{VaryByHeader("Accept-Language")})

	req1 := httptest.NewRequest("GET", "/products", nil)
	req1.Header.Set("Accept-Language", "en-US")
	key1 := gen.GenerateKey(req1)

	req2 := httptest.NewRequest("GET", "/products", nil)
	req2.Header.Set("Accept-Language", "en-US")
	key2 := gen.GenerateKey(req2)

	if key1 != key2 {
		t.Error("Expected same key for same header")
	}

	req3 := httptest.NewRequest("GET", "/products", nil)
	req3.Header.Set("Accept-Language", "fr-FR")
	key3 := gen.GenerateKey(req3)

	if key1 == key3 {
		t.Error("Expected different key for different header values")
	}

	req4 := httptest.NewRequest("GET", "/products", nil)
	key4 := gen.GenerateKey(req4)

	if key1 == key4 {
		t.Error("Expected different key when header is missing")
	}

	req5 := httptest.NewRequest("GET", "/products", nil)
	req5.Header.Set("Accept-Language", "en-US")
	req5.Header.Set("User-Agent", "TestAgent")
	key5 := gen.GenerateKey(req5)

	if key1 != key5 {
		t.Error("Expected same key when non-specified headers are present")
	}
}

func TestCacheKeyGenerator_VaryByPath(t *testing.T) {
	gen := newCacheKeyGenerator([]interface{}{VaryByPath()})

	req1 := httptest.NewRequest("GET", "/users/123", nil)
	key1 := gen.GenerateKey(req1)

	req2 := httptest.NewRequest("GET", "/users/456", nil)
	key2 := gen.GenerateKey(req2)

	if key1 == key2 {
		t.Error("Expected different key for different path parameters")
	}
}

func TestCacheKeyGenerator_Multiple(t *testing.T) {
	gen := newCacheKeyGenerator([]interface{}{
		VaryByQuery("q"),
		VaryByHeader("Accept-Language"),
	})

	req1 := httptest.NewRequest("GET", "/search?q=test", nil)
	req1.Header.Set("Accept-Language", "en-US")
	key1 := gen.GenerateKey(req1)

	req2 := httptest.NewRequest("GET", "/search?q=test", nil)
	req2.Header.Set("Accept-Language", "en-US")
	key2 := gen.GenerateKey(req2)

	if key1 != key2 {
		t.Error("Expected same key for identical requests")
	}

	req3 := httptest.NewRequest("GET", "/search?q=other", nil)
	req3.Header.Set("Accept-Language", "en-US")
	key3 := gen.GenerateKey(req3)

	if key1 == key3 {
		t.Error("Expected different key for different query")
	}

	req4 := httptest.NewRequest("GET", "/search?q=test", nil)
	req4.Header.Set("Accept-Language", "fr-FR")
	key4 := gen.GenerateKey(req4)

	if key1 == key4 {
		t.Error("Expected different key for different header")
	}
}
