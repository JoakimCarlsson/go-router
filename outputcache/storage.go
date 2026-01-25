package outputcache

import (
	"net/http"
	"sync"
	"time"
)

// CachedResponse represents a cached HTTP response with metadata.
type CachedResponse struct {
	StatusCode        int
	Headers           http.Header
	Body              []byte
	CachedAt          time.Time
	ExpiresAt         time.Time
	Tags              []string
	SlidingExpiration bool
	SlidingDuration   time.Duration
	ETag              string
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
	
	// InvalidateTag removes all cached responses with the given tag.
	InvalidateTag(tag string)
	
	// InvalidateTags removes all cached responses with any of the given tags.
	InvalidateTags(tags ...string)
}

// MemoryStorage implements an in-memory cache storage with automatic expiration.
type MemoryStorage struct {
	mu       sync.RWMutex
	cache    map[string]*CachedResponse
	tagIndex map[string][]string
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		cache:    make(map[string]*CachedResponse),
		tagIndex: make(map[string][]string),
	}
}

// Get retrieves a cached response by key.
// Automatically removes expired entries and extends TTL for sliding expiration.
func (m *MemoryStorage) Get(key string) (*CachedResponse, bool) {
	m.mu.RLock()
	response, ok := m.cache[key]
	m.mu.RUnlock()

	if !ok {
		return nil, false
	}

	now := time.Now()
	if now.After(response.ExpiresAt) {
		m.Delete(key)
		return nil, false
	}

	if response.SlidingExpiration {
		m.mu.Lock()
		response.ExpiresAt = now.Add(response.SlidingDuration)
		m.mu.Unlock()
	}

	return response, true
}

// Set stores a cached response with the given TTL.
func (m *MemoryStorage) Set(key string, response *CachedResponse, ttl time.Duration) {
	response.ExpiresAt = response.CachedAt.Add(ttl)

	m.mu.Lock()
	m.cache[key] = response

	for _, tag := range response.Tags {
		m.tagIndex[tag] = append(m.tagIndex[tag], key)
	}
	m.mu.Unlock()
}

// Delete removes a cached response by key.
func (m *MemoryStorage) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	response, ok := m.cache[key]
	if !ok {
		return
	}

	for _, tag := range response.Tags {
		keys := m.tagIndex[tag]
		for i, k := range keys {
			if k == key {
				m.tagIndex[tag] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
		if len(m.tagIndex[tag]) == 0 {
			delete(m.tagIndex, tag)
		}
	}

	delete(m.cache, key)
}

// Clear removes all cached responses.
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	m.cache = make(map[string]*CachedResponse)
	m.tagIndex = make(map[string][]string)
	m.mu.Unlock()
}

// InvalidateTag removes all cached responses with the given tag.
func (m *MemoryStorage) InvalidateTag(tag string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	keys, ok := m.tagIndex[tag]
	if !ok {
		return
	}

	for _, key := range keys {
		delete(m.cache, key)
	}

	delete(m.tagIndex, tag)
}

// InvalidateTags removes all cached responses with any of the given tags.
func (m *MemoryStorage) InvalidateTags(tags ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	keysToDelete := make(map[string]bool)

	for _, tag := range tags {
		if keys, ok := m.tagIndex[tag]; ok {
			for _, key := range keys {
				keysToDelete[key] = true
			}
			delete(m.tagIndex, tag)
		}
	}

	for key := range keysToDelete {
		delete(m.cache, key)
	}
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
