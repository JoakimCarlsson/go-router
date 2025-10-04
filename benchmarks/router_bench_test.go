package benchmarks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
)

// BenchmarkRouter_StaticRoutes benchmarks routing to static paths
func BenchmarkRouter_StaticRoutes(b *testing.B) {
	r := router.New()

	// Add static routes
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
	r.GET("/api/v1/status", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("status"))
	})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_ParamRoutes benchmarks routing with path parameters
func BenchmarkRouter_ParamRoutes(b *testing.B) {
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

// BenchmarkRouter_MultipleParams benchmarks routing with multiple parameters
func BenchmarkRouter_MultipleParams(b *testing.B) {
	r := router.New()

	r.GET(
		"/users/{userID}/posts/{postID}/comments/{commentID}",
		func(c *router.Context) {
			userID := c.Param("userID")
			postID := c.Param("postID")
			commentID := c.Param("commentID")
			c.Status(200)
			c.Writer.Write([]byte(userID + "-" + postID + "-" + commentID))
		},
	)

	req := httptest.NewRequest("GET", "/users/123/posts/456/comments/789", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_Middleware benchmarks routing with middleware
func BenchmarkRouter_Middleware(b *testing.B) {
	r := router.New()

	// Add middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Middleware", "1")
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

// BenchmarkRouter_LargeRoutingTable benchmarks with many routes
func BenchmarkRouter_LargeRoutingTable(b *testing.B) {
	r := router.New()

	// Add many routes
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/route%d", i)
		r.GET(path, func(c *router.Context) {
			c.Status(200)
			c.Writer.Write([]byte("ok"))
		})
	}

	// Test a route in the middle
	req := httptest.NewRequest("GET", "/route500", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_JSONResponse benchmarks JSON response generation
func BenchmarkRouter_JSONResponse(b *testing.B) {
	r := router.New()

	type Response struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Active bool   `json:"active"`
	}

	r.GET("/json", func(c *router.Context) {
		resp := Response{
			ID:     123,
			Name:   "John Doe",
			Email:  "john@example.com",
			Active: true,
		}
		c.JSON(200, resp)
	})

	req := httptest.NewRequest("GET", "/json", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_RouteGroups benchmarks nested route groups
func BenchmarkRouter_RouteGroups(b *testing.B) {
	r := router.New()

	r.Group("/api", func(api *router.Router) {
		api.Group("/v1", func(v1 *router.Router) {
			v1.Group("/users", func(users *router.Router) {
				users.GET("/{id}", func(c *router.Context) {
					c.Status(200)
					c.Writer.Write([]byte("user"))
				})
			})
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_Concurrent benchmarks concurrent requests
func BenchmarkRouter_Concurrent(b *testing.B) {
	r := router.New()

	r.GET("/concurrent", func(c *router.Context) {
		c.Status(200)
		c.Writer.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/concurrent", nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// BenchmarkRouter_MemoryAllocation benchmarks memory allocation patterns
func BenchmarkRouter_MemoryAllocation(b *testing.B) {
	r := router.New()

	r.GET("/users/{id}/profile", func(c *router.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"user_id": id, "profile": "data"})
	})

	req := httptest.NewRequest("GET", "/users/12345/profile", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
