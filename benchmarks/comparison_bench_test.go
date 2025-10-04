package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
)

// Standard library router for comparison
func newStdLibRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("home"))
	})
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("users"))
	})
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("products"))
	})
	mux.HandleFunc(
		"/api/v1/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		},
	)

	return mux
}

// BenchmarkComparison_GoRouter_Static benchmarks our router with static routes
func BenchmarkComparison_GoRouter_Static(b *testing.B) {
	r := router.New()

	r.GET("/", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("home"))
	})
	r.GET("/users", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("users"))
	})
	r.GET("/products", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("products"))
	})
	r.GET("/api/v1/health", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_StdLib_Static benchmarks standard library router
func BenchmarkComparison_StdLib_Static(b *testing.B) {
	mux := newStdLibRouter()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_GoRouter_Params benchmarks our router with parameters
func BenchmarkComparison_GoRouter_Params(b *testing.B) {
	r := router.New()

	r.GET("/users/{id}", func(c *router.Context) {
		id := c.Param("id")
		c.Status(200)
		c.Writer.Write([]byte("user: " + id))
	})

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_StdLib_Params benchmarks standard library with manual param extraction
func BenchmarkComparison_StdLib_Params(b *testing.B) {
	mux := http.NewServeMux()

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		// Manual parameter extraction (simplified)
		path := r.URL.Path
		if len(path) > 7 { // "/users/"
			id := path[7:] // Extract ID
			w.Write([]byte("user: " + id))
		}
	})

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_GoRouter_Middleware benchmarks our router with middleware
func BenchmarkComparison_GoRouter_Middleware(b *testing.B) {
	r := router.New()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Custom", "1")
			next.ServeHTTP(w, req)
		})
	})

	r.GET("/test", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_StdLib_Middleware benchmarks standard library with middleware
func BenchmarkComparison_StdLib_Middleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Custom", "1")
			next.ServeHTTP(w, req)
		})
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_GoRouter_LargeTable benchmarks our router with many routes
func BenchmarkComparison_GoRouter_LargeTable(b *testing.B) {
	r := router.New()

	// Add 100 routes
	for i := 0; i < 100; i++ {
		path := "/route" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10))
		r.GET(path, func(c *router.Context) {
			c.Status(200)
			c.Writer.Write([]byte("ok"))
		})
	}

	req := httptest.NewRequest("GET", "/route50", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_StdLib_LargeTable benchmarks standard library with many routes
func BenchmarkComparison_StdLib_LargeTable(b *testing.B) {
	mux := http.NewServeMux()

	// Add 100 routes
	for i := 0; i < 100; i++ {
		path := "/route" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10))
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})
	}

	req := httptest.NewRequest("GET", "/route50", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_GoRouter_JSON benchmarks JSON response
func BenchmarkComparison_GoRouter_JSON(b *testing.B) {
	r := router.New()

	type Response struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	r.GET("/json", func(c *router.Context) {
		c.JSON(200, Response{Message: "success", Code: 200})
	})

	req := httptest.NewRequest("GET", "/json", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkComparison_StdLib_JSON benchmarks standard library JSON response
func BenchmarkComparison_StdLib_JSON(b *testing.B) {
	mux := http.NewServeMux()

	type Response struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Manual JSON encoding would be here
		w.Write([]byte(`{"message":"success","code":200}`))
	})

	req := httptest.NewRequest("GET", "/json", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}
}
