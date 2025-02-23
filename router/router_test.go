package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_BasicRouting(t *testing.T) {
	r := New()
	called := false

	r.GET("/test", func(c *Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
}

func TestRouter_MethodHandlers(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			r := New()
			called := false

			handler := func(c *Context) {
				called = true
				c.Status(http.StatusOK)
			}

			switch method {
			case "GET":
				r.GET("/test", handler)
			case "POST":
				r.POST("/test", handler)
			case "PUT":
				r.PUT("/test", handler)
			case "DELETE":
				r.DELETE("/test", handler)
			case "PATCH":
				r.PATCH("/test", handler)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/test", nil)
			r.ServeHTTP(w, req)

			if !called {
				t.Error("handler was not called")
			}
			if w.Code != http.StatusOK {
				t.Errorf("expected status OK; got %v", w.Code)
			}
		})
	}
}

func TestRouter_Groups(t *testing.T) {
	r := New()
	var handlerCalled bool

	r.Group("/api", func(g *Router) {
		g.Group("/v1", func(v1 *Router) {
			v1.GET("/test", func(c *Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			})
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	r.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("group handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
}

func TestRouter_Middleware(t *testing.T) {
	r := New()
	middlewareCalled := false
	handlerCalled := false

	middleware := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			middlewareCalled = true
			next(c)
		}
	}

	r.Use(middleware)
	r.GET("/test", func(c *Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("middleware was not called")
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestRouter_Tags(t *testing.T) {
	r := New()
	r.WithTags("test-tag")

	r.GET("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	routes := r.collectRoutesRecursively()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if len(routes[0].Tags) != 1 || routes[0].Tags[0] != "test-tag" {
		t.Errorf("expected route to have tag 'test-tag', got %v", routes[0].Tags)
	}
}

func TestRouter_Security(t *testing.T) {
	r := New()
	r.WithSecurity(map[string][]string{
		"bearerAuth": {"read", "write"},
	})

	r.GET("/secure", func(c *Context) {
		c.Status(http.StatusOK)
	})

	routes := r.collectRoutesRecursively()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if len(routes[0].Security) != 1 {
		t.Fatalf("expected 1 security requirement, got %d", len(routes[0].Security))
	}
}

func TestRouter_Static(t *testing.T) {
	r := New()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid static path")
		}
	}()
	r.Static("/static", "./testdata") // Should panic without /*filepath suffix
}

func TestRouter_WithTagsAndSecurity(t *testing.T) {
	r := New()
	r.WithTags("test-tag")
	r.WithSecurity(map[string][]string{
		"bearerAuth": {"read"},
	})

	r.GET("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	routes := r.collectRoutesRecursively()
	if len(routes) == 0 {
		t.Fatal("expected at least one route")
	}

	route := routes[0]
	if len(route.Tags) == 0 || route.Tags[0] != "test-tag" {
		t.Errorf("expected route to have tag 'test-tag', got %v", route.Tags)
	}

	if len(route.Security) == 0 {
		t.Error("expected route to have security requirements")
	}
}

func TestRouter_PathNormalization(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/test/", "/test"},
		{"test", "/test"},
		{"/test//path/", "/test/path"},
		{"///test", "/test"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := normalizePath(tt.path)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q; want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func BenchmarkRouter_SimpleRoute(b *testing.B) {
	r := New()
	r.GET("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_GroupedRoute(b *testing.B) {
	r := New()
	r.Group("/api", func(g *Router) {
		g.GET("/test", func(c *Context) {
			c.Status(http.StatusOK)
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_WithMiddleware(b *testing.B) {
	r := New()
	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			next(c)
		}
	})

	r.GET("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_DeeplyNestedGroup(b *testing.B) {
	r := New()
	r.Group("/api", func(g1 *Router) {
		g1.Group("/v1", func(g2 *Router) {
			g2.Group("/users", func(g3 *Router) {
				g3.Group("/profiles", func(g4 *Router) {
					g4.GET("/details", func(c *Context) {
						c.Status(http.StatusOK)
					})
				})
			})
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/profiles/details", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_MultipleMiddlewares(b *testing.B) {
	r := New()

	// Add multiple middlewares
	for i := 0; i < 5; i++ {
		r.Use(func(next HandlerFunc) HandlerFunc {
			return func(c *Context) {
				next(c)
			}
		})
	}

	r.GET("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_WithMetadata(b *testing.B) {
	r := New()
	r.WithTags("tag1", "tag2")
	r.WithSecurity(map[string][]string{
		"bearerAuth": {"read", "write"},
	})

	r.GET("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_ComplexSetup(b *testing.B) {
	r := New()

	// Add global middlewares
	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			c.SetHeader("X-Request-ID", "test-id")
			next(c)
		}
	})

	// Create a complex routing structure
	r.Group("/api", func(api *Router) {
		api.WithTags("api")
		api.WithSecurity(map[string][]string{"bearerAuth": {"read"}})

		api.Group("/v1", func(v1 *Router) {
			v1.WithTags("v1")

			v1.Group("/users", func(users *Router) {
				users.WithTags("users")
				users.Use(func(next HandlerFunc) HandlerFunc {
					return func(c *Context) {
						next(c)
					}
				})

				users.GET("/profile", func(c *Context) {
					c.JSON(http.StatusOK, map[string]string{"status": "ok"})
				})
			})
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_ParallelRequests(b *testing.B) {
	r := New()
	r.GET("/test", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	b.RunParallel(func(pb *testing.PB) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		for pb.Next() {
			w.Body.Reset()
			r.ServeHTTP(w, req)
		}
	})
}

func BenchmarkRouter_StaticRoute(b *testing.B) {
	r := New()
	tempDir := b.TempDir()

	r.Static("/static/*filepath", tempDir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/static/test.txt", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_DynamicRoutes(b *testing.B) {
	r := New()
	r.GET("/users/{id}", func(c *Context) { c.Status(http.StatusOK) })
	r.GET("/users/{id}/posts/{postId}", func(c *Context) { c.Status(http.StatusOK) })
	r.GET("/articles/{category}/{id}/{slug}", func(c *Context) { c.Status(http.StatusOK) })

	paths := []string{
		"/users/123",
		"/users/456/posts/789",
		"/articles/tech/42/how-to-benchmark",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			for _, path := range paths {
				req := httptest.NewRequest("GET", path, nil)
				r.ServeHTTP(w, req)
				w.Body.Reset()
			}
		}
	})
}

func BenchmarkRouter_ComplexMiddlewareChain(b *testing.B) {
	r := New()

	// Add 10 middleware functions that simulate real-world scenarios
	for i := 0; i < 10; i++ {
		i := i // Capture loop variable
		r.Use(func(next HandlerFunc) HandlerFunc {
			return func(c *Context) {
				c.SetHeader(fmt.Sprintf("X-Middleware-%d", i), "processed")
				next(c)
			}
		})
	}

	r.GET("/test", func(c *Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"data":   make([]int, 100), // Simulate payload
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_GroupInheritance(b *testing.B) {
	b.ReportAllocs()
	r := New()

	// Create a deep group hierarchy with inherited middleware and metadata
	setupGroup := func() {
		r.WithTags("api")
		r.WithSecurity(map[string][]string{"bearerAuth": {"read"}})
		r.Use(func(next HandlerFunc) HandlerFunc {
			return func(c *Context) {
				c.SetHeader("X-API-Version", "1.0")
				next(c)
			}
		})

		r.Group("/api", func(api *Router) {
			api.WithTags("v1")
			api.Use(func(next HandlerFunc) HandlerFunc {
				return func(c *Context) {
					c.SetHeader("X-API-Group", "main")
					next(c)
				}
			})

			api.Group("/users", func(users *Router) {
				users.WithTags("users")
				users.GET("/:id", func(c *Context) {
					c.JSON(http.StatusOK, map[string]string{"id": c.Param("id")})
				})
			})
		})
	}

	setupGroup()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/users/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_WildcardRoutes(b *testing.B) {
	r := New()
	r.Static("/assets/*filepath", b.TempDir())
	r.GET("/api/*path", func(c *Context) {
		c.Status(http.StatusOK)
	})

	paths := []string{
		"/assets/css/style.css",
		"/assets/js/app.js",
		"/api/v1/users",
		"/api/v2/posts/123",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		w := httptest.NewRecorder()
		for pb.Next() {
			for _, path := range paths {
				req := httptest.NewRequest("GET", path, nil)
				r.ServeHTTP(w, req)
				w.Body.Reset()
			}
		}
	})
}
