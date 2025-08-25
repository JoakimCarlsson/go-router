package benchmarks

import (
	"net/http/httptest"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
)

// BenchmarkMinimalAllocation tests the absolute minimum allocation path
func BenchmarkMinimalAllocation(b *testing.B) {
	r := router.New()
	
	r.GET("/test", func(c *router.Context) {
		// Minimal response - no JSON, no string operations
		c.Status(200)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkWithoutRouter tests just the httptest overhead
func BenchmarkWithoutRouter(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		_ = w
	}
}