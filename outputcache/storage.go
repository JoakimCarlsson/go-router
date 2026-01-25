package outputcache

import (
	"net/http"
	"sync"
	"time"
)

// CachedResponse represents a cached HTTP response with metadata.
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	CachedAt   time.Time
	ExpiresAt  time.Time
}

// Storage defines the interface for cache storage backends.
// Implementations must be thread-safe.
type Storage interface {
	// Get retrieves a cached response by key.
	// Returns the cached response and true if found, nil and false otherwise.
	Get(key string) (*CachedResponse, bool)
	
	// Set stores a cached response with the given TTL.
	// The response will be automatically expired after the TTL duration.
	Set(key string, response *CachedResponse, ttl time.Duration)
	
	// Delete removes a cached response by key.
	Delete(key string)
	
	// Clear removes all cached responses.
	Clear()
}

// MemoryStorage implements an in-memory cache storage with automatic expiration.
type MemoryStorage struct {
	mu    sync.RWMutex
	cache map[string]*CachedResponse
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		cache: make(map[string]*CachedResponse),
	}
}

// Get retrieves a cached response by key.
// Automatically removes expired entries.
func (m *MemoryStorage) Get(key string) (*CachedResponse, bool) {
	m.mu.RLock()
	response, ok := m.cache[key]
	m.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Now().After(response.ExpiresAt) {
		m.Delete(key)
		return nil, false
	}

	return response, true
}

// Set stores a cached response with the given TTL.
func (m *MemoryStorage) Set(key string, response *CachedResponse, ttl time.Duration) {
	response.ExpiresAt = response.CachedAt.Add(ttl)
	
	m.mu.Lock()
	m.cache[key] = response
	m.mu.Unlock()
}

// Delete removes a cached response by key.
func (m *MemoryStorage) Delete(key string) {
	m.mu.Lock()
	delete(m.cache, key)
	m.mu.Unlock()
}

// Clear removes all cached responses.
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	m.cache = make(map[string]*CachedResponse)
	m.mu.Unlock()
}

// StartCleanup starts a background goroutine that periodically removes expired entries.
// Returns a function that can be called to stop the cleanup goroutine.
func (m *MemoryStorage) StartCleanup(interval time.Duration) func() {
	done := make(chan struct{})
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.cleanup()
			case <-done:
				return
			}
		}
	}()
	
	return func() {
		close(done)
	}
}

func (m *MemoryStorage) cleanup() {
	now := time.Now()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for key, response := range m.cache {
		if now.After(response.ExpiresAt) {
			delete(m.cache, key)
		}
	}
}
