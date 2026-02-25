package outputcache

import (
	"net/http"
	"testing"
	"time"
)

func TestMemoryStorage_GetSet(t *testing.T) {
	storage := NewMemoryStorage()

	response := &CachedResponse{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"message": "test"}`),
		CachedAt:   time.Now(),
	}

	storage.Set("test-key", response, time.Minute)

	cached, ok := storage.Get("test-key")
	if !ok {
		t.Fatal("Expected to find cached response")
	}

	if cached.StatusCode != response.StatusCode {
		t.Errorf("Expected status %d, got %d", response.StatusCode, cached.StatusCode)
	}

	if string(cached.Body) != string(response.Body) {
		t.Errorf("Expected body %s, got %s", response.Body, cached.Body)
	}
}

func TestMemoryStorage_Expiration(t *testing.T) {
	storage := NewMemoryStorage()

	response := &CachedResponse{
		StatusCode: 200,
		Body:       []byte("test"),
		CachedAt:   time.Now(),
	}

	storage.Set("test-key", response, 10*time.Millisecond)

	if _, ok := storage.Get("test-key"); !ok {
		t.Fatal("Expected to find cached response")
	}

	time.Sleep(20 * time.Millisecond)

	if _, ok := storage.Get("test-key"); ok {
		t.Fatal("Expected cached response to be expired")
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	storage := NewMemoryStorage()

	response := &CachedResponse{
		StatusCode: 200,
		Body:       []byte("test"),
		CachedAt:   time.Now(),
	}

	storage.Set("test-key", response, time.Minute)

	if _, ok := storage.Get("test-key"); !ok {
		t.Fatal("Expected to find cached response")
	}

	storage.Delete("test-key")

	if _, ok := storage.Get("test-key"); ok {
		t.Fatal("Expected cached response to be deleted")
	}
}

func TestMemoryStorage_Clear(t *testing.T) {
	storage := NewMemoryStorage()

	response := &CachedResponse{
		StatusCode: 200,
		Body:       []byte("test"),
		CachedAt:   time.Now(),
	}

	storage.Set("key1", response, time.Minute)
	storage.Set("key2", response, time.Minute)
	storage.Set("key3", response, time.Minute)

	for _, key := range []string{"key1", "key2", "key3"} {
		if _, ok := storage.Get(key); !ok {
			t.Fatalf("Expected to find cached response for %s", key)
		}
	}

	storage.Clear()

	for _, key := range []string{"key1", "key2", "key3"} {
		if _, ok := storage.Get(key); ok {
			t.Fatalf("Expected cached response for %s to be cleared", key)
		}
	}
}

func TestMemoryStorage_Concurrent(t *testing.T) {
	storage := NewMemoryStorage()

	response := &CachedResponse{
		StatusCode: 200,
		Body:       []byte("test"),
		CachedAt:   time.Now(),
	}

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(n int) {
			key := string(rune(n))
			storage.Set(key, response, time.Minute)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	for i := 0; i < 100; i++ {
		go func(n int) {
			key := string(rune(n))
			storage.Get(key)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
