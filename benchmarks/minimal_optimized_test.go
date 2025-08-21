package benchmarks

import (
	"net/http/httptest"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
)

// BenchmarkOptimizedMinimal tests with minimal operations only
func BenchmarkOptimizedMinimal(b *testing.B) {
	r := router.New()
	
	// Ultra-minimal handler that avoids any potential allocations
	r.GET("/test", func(c *router.Context) {
		// Just set status - no response body
		c.StatusCode = 200
		c.Writer.WriteHeader(200)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}