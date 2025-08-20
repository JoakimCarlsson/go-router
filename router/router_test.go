package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// Context key types to avoid staticcheck warnings
type contextKey string

const (
	userKey     contextKey = "user"
	userIDKey   contextKey = "userID"
	userIDKey2  contextKey = "user_id"
	usernameKey contextKey = "username"
	key1        contextKey = "key1"
	key2        contextKey = "key2"
	key3        contextKey = "key3"
)

// BenchmarkRouteRegistration measures the performance of registering routes
func BenchmarkRouteRegistration(b *testing.B) {
	b.ReportAllocs()

	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := router.New()
			r.GET("/", func(c *router.Context) {})
			r.GET("/users", func(c *router.Context) {})
			r.POST("/users", func(c *router.Context) {})
			r.GET("/users/{id}", func(c *router.Context) {})
			r.PUT("/users/{id}", func(c *router.Context) {})
			r.DELETE("/users/{id}", func(c *router.Context) {})
		}
	})

	b.Run("WithGroups", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := router.New()
			r.Group("/api", func(api *router.Router) {
				api.GET("/health", func(c *router.Context) {})

				api.Group("/v1", func(v1 *router.Router) {
					v1.GET("/users", func(c *router.Context) {})
					v1.POST("/users", func(c *router.Context) {})
					v1.GET("/users/{id}", func(c *router.Context) {})
					v1.PUT("/users/{id}", func(c *router.Context) {})
					v1.DELETE("/users/{id}", func(c *router.Context) {})
				})
			})
		}
	})

	b.Run("WithMetadata", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := router.New()
			r.GET("/users", func(c *router.Context) {},
				docs.WithSummary("List users"),
				docs.WithDescription("Get a list of all users"),
				docs.WithTags("users"),
				docs.WithResponse(200, "List of users"),
			)
			r.POST("/users", func(c *router.Context) {},
				docs.WithSummary("Create user"),
				docs.WithDescription("Create a new user"),
				docs.WithTags("users"),
				docs.WithResponse(201, "User created"),
				docs.WithResponse(400, "Invalid request"),
			)
			r.GET("/users/{id}", func(c *router.Context) {},
				docs.WithSummary("Get user"),
				docs.WithDescription("Get a user by ID"),
				docs.WithTags("users"),
				docs.WithPathParam("id", "string", true, "User ID", nil),
				docs.WithResponse(200, "User found"),
				docs.WithResponse(404, "User not found"),
			)
		}
	})
}

// BenchmarkRequestHandling measures the performance of handling HTTP requests
func BenchmarkRequestHandling(b *testing.B) {
	// Setup handlers
	helloHandler := func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("Hello, World!"))
	}

	jsonHandler := func(c *router.Context) {
		c.JSON(200, map[string]string{"message": "Hello, World!"})
	}

	paramHandler := func(c *router.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"id": id})
	}

	b.Run("StaticRoute", func(b *testing.B) {
		r := router.New()
		r.GET("/hello", helloHandler)

		req := httptest.NewRequest("GET", "/hello", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("JSONResponse", func(b *testing.B) {
		r := router.New()
		r.GET("/api/json", jsonHandler)

		req := httptest.NewRequest("GET", "/api/json", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("WithPathParams", func(b *testing.B) {
		r := router.New()
		r.GET("/users/{id}", paramHandler)

		req := httptest.NewRequest("GET", "/users/123", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("MultipleRoutes", func(b *testing.B) {
		r := router.New()
		// Add lots of routes to test routing performance with a larger route table
		for i := 0; i < 100; i++ {
			path := "/path" + strconv.Itoa(i)
			r.GET(path, helloHandler)
		}
		r.GET("/target", helloHandler)

		req := httptest.NewRequest("GET", "/target", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// BenchmarkMiddleware measures the performance of middleware execution
func BenchmarkMiddleware(b *testing.B) {
	// Setup simple middlewares using standard HTTP middleware pattern
	stdLoggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Just call next handler
			next.ServeHTTP(w, r)
		})
	}

	stdAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate auth check
			if r.Header.Get("Authorization") != "" {
				// We can't directly set user context in standard middleware
				ctx := r.Context()
				ctx = context.WithValue(ctx, userKey, "authenticated")
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}

	b.Run("NoMiddleware", func(b *testing.B) {
		r := router.New()
		r.GET("/hello", func(c *router.Context) {
			c.Writer.WriteHeader(200)
			_, _ = c.Writer.Write([]byte("Hello, World!"))
		})

		req := httptest.NewRequest("GET", "/hello", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("SingleMiddleware", func(b *testing.B) {
		r := router.New()
		r.Use(stdLoggingMiddleware)
		r.GET("/hello", func(c *router.Context) {
			c.Writer.WriteHeader(200)
			_, _ = c.Writer.Write([]byte("Hello, World!"))
		})

		req := httptest.NewRequest("GET", "/hello", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("MultipleMiddlewares", func(b *testing.B) {
		r := router.New()
		r.Use(stdLoggingMiddleware)
		r.Use(stdAuthMiddleware)
		r.Use(stdLoggingMiddleware) // Add a third middleware
		r.GET("/hello", func(c *router.Context) {
			c.Writer.WriteHeader(200)
			_, _ = c.Writer.Write([]byte("Hello, World!"))
		})

		req := httptest.NewRequest("GET", "/hello", nil)
		req.Header.Set("Authorization", "Bearer token")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("GroupMiddlewares", func(b *testing.B) {
		r := router.New()
		r.Use(stdLoggingMiddleware)

		r.Group("/api", func(api *router.Router) {
			api.Use(stdAuthMiddleware)

			api.GET("/hello", func(c *router.Context) {
				c.Writer.WriteHeader(200)
				_, _ = c.Writer.Write([]byte("Hello, World!"))
			})
		})

		req := httptest.NewRequest("GET", "/api/hello", nil)
		req.Header.Set("Authorization", "Bearer token")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// BenchmarkContextOperations measures the performance of Context operations
func BenchmarkContextOperations(b *testing.B) {
	handler := func(c *router.Context) {
		c.JSON(200, map[string]string{"message": "Hello, World!"})
	}

	r := router.New()
	r.GET("/test", handler)

	b.Run("ParamExtraction", func(b *testing.B) {
		r := router.New()
		r.GET("/users/{id}", func(c *router.Context) {
			id := c.Param("id")
			_ = id // Use the parameter to avoid optimization
		})

		req := httptest.NewRequest("GET", "/users/123", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("QueryParams", func(b *testing.B) {
		r := router.New()
		r.GET("/search", func(c *router.Context) {
			query := c.QueryDefault("q", "")
			limit := c.QueryIntDefault("limit", 10)
			_ = query // Use to avoid optimization
			_ = limit
		})

		req := httptest.NewRequest("GET", "/search?q=test&limit=20", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("ContextStore", func(b *testing.B) {
		r := router.New()
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Store values in request context
				ctx := r.Context()
				ctx = context.WithValue(ctx, key1, "value1")
				ctx = context.WithValue(ctx, key2, 123)
				ctx = context.WithValue(ctx, key3, true)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})

		r.GET("/test", func(c *router.Context) {
			val1, _ := c.Get("key1")
			val2, _ := c.Get("key2")
			val3, _ := c.Get("key3")
			_ = val1
			_ = val2
			_ = val3
		})

		req := httptest.NewRequest("GET", "/test", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("JSONResponse", func(b *testing.B) {
		r := router.New()
		r.GET("/json", func(c *router.Context) {
			data := map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
				"address": map[string]string{
					"street":  "123 Main St",
					"city":    "Anytown",
					"country": "USA",
				},
				"hobbies": []string{"reading", "gaming", "hiking"},
			}
			c.JSON(200, data)
		})

		req := httptest.NewRequest("GET", "/json", nil)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// BenchmarkContentNegotiation measures the performance of content negotiation
func BenchmarkContentNegotiation(b *testing.B) {
	type User struct {
		ID    string `json:"id" xml:"id"`
		Name  string `json:"name" xml:"name"`
		Email string `json:"email" xml:"email"`
	}

	user := User{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	r := router.New()
	r.GET("/user", func(c *router.Context) {
		switch c.Negotiate("application/json", "application/xml") {
		case "application/xml":
			c.XML(200, user)
		default:
			c.JSON(200, user)
		}
	})

	b.Run("JSON", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Accept", "application/json")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("XML", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Accept", "application/xml")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("Wildcard", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Accept", "*/*")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Helper functions for real-world scenario benchmarks
func setupProductAPI() *router.Router {
	r := router.New()

	// Middleware for all routes
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Response-Time", "0.1ms") // Simulated response time
			next.ServeHTTP(w, r)
		})
	})

	// Product API
	r.Group("/api/products", func(api *router.Router) {
		// Auth middleware
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "" {
					ctx := r.Context()
					ctx = context.WithValue(ctx, userIDKey, "user-123")
					r = r.WithContext(ctx)
				}
				next.ServeHTTP(w, r)
			})
		})

		// Routes
		api.GET("", listProducts)
		api.GET("/{id}", getProduct)
		api.POST("", createProduct)
		api.PUT("/{id}", updateProduct)
		api.DELETE("/{id}", deleteProduct)
	})

	return r
}

// Product handlers
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Categories  []string `json:"categories"`
	InStock     bool     `json:"inStock"`
}

func listProducts(c *router.Context) {
	// Simulate filtering
	category := c.QueryDefault("category", "")
	inStock := c.QueryBoolDefault("inStock", false)

	// Return dummy data
	products := []Product{
		{ID: "1", Name: "Product 1", Price: 29.99, InStock: true},
		{ID: "2", Name: "Product 2", Price: 39.99, InStock: true},
		{ID: "3", Name: "Product 3", Price: 49.99, InStock: inStock},
	}

	// Filter if category provided
	if category != "" {
		filtered := []Product{}
		for _, p := range products {
			for _, c := range p.Categories {
				if c == category {
					filtered = append(filtered, p)
					break
				}
			}
		}
		products = filtered
	}

	c.JSON(200, products)
}

func getProduct(c *router.Context) {
	id := c.Param("id")

	product := Product{
		ID:          id,
		Name:        "Product " + id,
		Description: "This is product " + id,
		Price:       29.99,
		Categories:  []string{"electronics", "gadgets"},
		InStock:     true,
	}

	c.JSON(200, product)
}

func createProduct(c *router.Context) {
	var product Product
	if err := c.BindJSON(&product); err != nil {
		c.JSON(400, map[string]string{"error": "Invalid request body"})
		return
	}

	// Simulate ID generation
	product.ID = "new-id"

	c.JSON(201, product)
}

func updateProduct(c *router.Context) {
	id := c.Param("id")

	var product Product
	if err := c.BindJSON(&product); err != nil {
		c.JSON(400, map[string]string{"error": "Invalid request body"})
		return
	}

	// Ensure ID matches path parameter
	product.ID = id

	c.JSON(200, product)
}

func deleteProduct(c *router.Context) {
	// Just return a success status
	c.Status(204)
}

// BenchmarkRealWorldScenario measures performance in a realistic API scenario
func BenchmarkRealWorldScenario(b *testing.B) {
	r := setupProductAPI()
	sampleProduct := Product{
		Name:        "New Product",
		Description: "Product description",
		Price:       59.99,
		Categories:  []string{"electronics", "new"},
		InStock:     true,
	}

	jsonBody, _ := json.Marshal(sampleProduct)

	b.Run("ListProducts", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/api/products", nil)
		req.Header.Set("Authorization", "Bearer token")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("FilteredProducts", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/api/products?category=electronics&inStock=true", nil)
		req.Header.Set("Authorization", "Bearer token")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("GetProduct", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/api/products/123", nil)
		req.Header.Set("Authorization", "Bearer token")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("CreateProduct", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/products", createReaderFromBytes(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer token")
			r.ServeHTTP(w, req)
		}
	})
}

func createReaderFromBytes(b []byte) io.Reader {
	return bytes.NewReader(b)
}

// Test middleware functionality
func TestRouter_Use(t *testing.T) {
	t.Run("single middleware", func(t *testing.T) {
		r := router.New()

		// Middleware that adds a header
		headerMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Middleware", "applied")
				next.ServeHTTP(w, req)
			})
		}

		r.Use(headerMiddleware)
		r.GET("/test", func(c *router.Context) {
			c.JSON(200, map[string]string{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Header().Get("X-Middleware") != "applied" {
			t.Errorf("Expected middleware header to be set")
		}

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("multiple middleware", func(t *testing.T) {
		r := router.New()

		// Multiple middleware that modify request/response
		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Middleware-1", "first")
				next.ServeHTTP(w, req)
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Middleware-2", "second")
				next.ServeHTTP(w, req)
			})
		}

		middleware3 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Middleware-3", "third")
				next.ServeHTTP(w, req)
			})
		}

		r.Use(middleware1, middleware2, middleware3)
		r.GET("/test", func(c *router.Context) {
			c.JSON(200, map[string]string{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// All middleware should have been applied
		if w.Header().Get("X-Middleware-1") != "first" {
			t.Errorf("Expected first middleware header to be set")
		}
		if w.Header().Get("X-Middleware-2") != "second" {
			t.Errorf("Expected second middleware header to be set")
		}
		if w.Header().Get("X-Middleware-3") != "third" {
			t.Errorf("Expected third middleware header to be set")
		}
	})

	t.Run("middleware execution order", func(t *testing.T) {
		r := router.New()

		var executionOrder []string

		// Middleware that tracks execution order
		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				executionOrder = append(executionOrder, "before-1")
				next.ServeHTTP(w, req)
				executionOrder = append(executionOrder, "after-1")
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				executionOrder = append(executionOrder, "before-2")
				next.ServeHTTP(w, req)
				executionOrder = append(executionOrder, "after-2")
			})
		}

		r.Use(middleware1, middleware2)
		r.GET("/test", func(c *router.Context) {
			executionOrder = append(executionOrder, "handler")
			c.JSON(200, map[string]string{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		expectedOrder := []string{"before-1", "before-2", "handler", "after-2", "after-1"}
		if len(executionOrder) != len(expectedOrder) {
			t.Fatalf("Expected %d execution steps, got %d", len(expectedOrder), len(executionOrder))
		}

		for i, expected := range expectedOrder {
			if executionOrder[i] != expected {
				t.Errorf("Expected execution order[%d] to be '%s', got '%s'", i, expected, executionOrder[i])
			}
		}
	})

	t.Run("middleware can modify request", func(t *testing.T) {
		r := router.New()

		// Middleware that adds authentication info to context
		authMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ctx := context.WithValue(req.Context(), userIDKey2, "123")
				ctx = context.WithValue(ctx, usernameKey, "testuser")
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		}

		r.Use(authMiddleware)
		r.GET("/profile", func(c *router.Context) {
			userID := c.Request.Context().Value(userIDKey2)
			username := c.Request.Context().Value(usernameKey)

			c.JSON(200, map[string]interface{}{
				"user_id":  userID,
				"username": username,
			})
		})

		req := httptest.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		if response["user_id"] != "123" {
			t.Errorf("Expected user_id to be '123', got %v", response["user_id"])
		}

		if response["username"] != "testuser" {
			t.Errorf("Expected username to be 'testuser', got %v", response["username"])
		}
	})

	t.Run("middleware can short-circuit request", func(t *testing.T) {
		r := router.New()

		// Middleware that blocks unauthorized requests
		authMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				authHeader := req.Header.Get("Authorization")
				if authHeader != "Bearer valid-token" {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte("Unauthorized"))
					return // Short-circuit, don't call next handler
				}
				next.ServeHTTP(w, req)
			})
		}

		r.Use(authMiddleware)
		r.GET("/protected", func(c *router.Context) {
			c.JSON(200, map[string]string{"message": "Access granted"})
		})

		// Test without valid token
		req1 := httptest.NewRequest("GET", "/protected", nil)
		w1 := httptest.NewRecorder()

		r.ServeHTTP(w1, req1)

		if w1.Code != 401 {
			t.Errorf("Expected status 401, got %d", w1.Code)
		}

		if w1.Body.String() != "Unauthorized" {
			t.Errorf("Expected 'Unauthorized' response, got '%s'", w1.Body.String())
		}

		// Test with valid token
		req2 := httptest.NewRequest("GET", "/protected", nil)
		req2.Header.Set("Authorization", "Bearer valid-token")
		w2 := httptest.NewRecorder()

		r.ServeHTTP(w2, req2)

		if w2.Code != 200 {
			t.Errorf("Expected status 200, got %d", w2.Code)
		}

		if !strings.Contains(w2.Body.String(), "Access granted") {
			t.Errorf("Expected successful response, got '%s'", w2.Body.String())
		}
	})
}

// Test route group functionality
func TestRouter_Group(t *testing.T) {
	t.Run("simple group", func(t *testing.T) {
		r := router.New()

		r.Group("/api", func(api *router.Router) {
			api.GET("/users", func(c *router.Context) {
				c.JSON(200, []string{"user1", "user2"})
			})

			api.POST("/users", func(c *router.Context) {
				c.JSON(201, map[string]string{"status": "created"})
			})
		})

		// Test GET /api/users
		req1 := httptest.NewRequest("GET", "/api/users", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Code != 200 {
			t.Errorf("Expected status 200, got %d", w1.Code)
		}

		// Test POST /api/users
		req2 := httptest.NewRequest("POST", "/api/users", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != 201 {
			t.Errorf("Expected status 201, got %d", w2.Code)
		}
	})

	t.Run("nested groups", func(t *testing.T) {
		r := router.New()

		r.Group("/api", func(api *router.Router) {
			api.GET("/health", func(c *router.Context) {
				c.JSON(200, map[string]string{"status": "ok"})
			})

			api.Group("/v1", func(v1 *router.Router) {
				v1.GET("/users", func(c *router.Context) {
					c.JSON(200, []string{"user1", "user2"})
				})

				v1.Group("/admin", func(admin *router.Router) {
					admin.GET("/settings", func(c *router.Context) {
						c.JSON(200, map[string]string{"role": "admin"})
					})
				})
			})

			api.Group("/v2", func(v2 *router.Router) {
				v2.GET("/users", func(c *router.Context) {
					c.JSON(200, []string{"user1", "user2", "user3"})
				})
			})
		})

		// Test /api/health
		req1 := httptest.NewRequest("GET", "/api/health", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Code != 200 {
			t.Errorf("Expected status 200 for /api/health, got %d", w1.Code)
		}

		// Test /api/v1/users
		req2 := httptest.NewRequest("GET", "/api/v1/users", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != 200 {
			t.Errorf("Expected status 200 for /api/v1/users, got %d", w2.Code)
		}

		// Test /api/v1/admin/settings
		req3 := httptest.NewRequest("GET", "/api/v1/admin/settings", nil)
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)

		if w3.Code != 200 {
			t.Errorf("Expected status 200 for /api/v1/admin/settings, got %d", w3.Code)
		}

		// Test /api/v2/users (different from v1)
		req4 := httptest.NewRequest("GET", "/api/v2/users", nil)
		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, req4)

		if w4.Code != 200 {
			t.Errorf("Expected status 200 for /api/v2/users, got %d", w4.Code)
		}

		// Verify that v2 returns different data than v1
		var v2Response []string
		if err := json.Unmarshal(w4.Body.Bytes(), &v2Response); err != nil {
			t.Errorf("Failed to parse v2 response: %v", err)
		} else if len(v2Response) != 3 {
			t.Errorf("Expected v2 to return 3 users, got %d", len(v2Response))
		}
	})

	t.Run("groups inherit parent middleware", func(t *testing.T) {
		r := router.New()

		// Add global middleware
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Global", "global")
				next.ServeHTTP(w, req)
			})
		})

		r.Group("/api", func(api *router.Router) {
			// Add API-specific middleware
			api.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("X-API", "api")
					next.ServeHTTP(w, req)
				})
			})

			api.GET("/test", func(c *router.Context) {
				c.JSON(200, map[string]string{"status": "ok"})
			})

			api.Group("/v1", func(v1 *router.Router) {
				// Add v1-specific middleware
				v1.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.Header().Set("X-V1", "v1")
						next.ServeHTTP(w, req)
					})
				})

				v1.GET("/nested", func(c *router.Context) {
					c.JSON(200, map[string]string{"version": "v1"})
				})
			})
		})

		// Test /api/test (should have global + api middleware)
		req1 := httptest.NewRequest("GET", "/api/test", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Header().Get("X-Global") != "global" {
			t.Errorf("Expected global middleware header")
		}

		if w1.Header().Get("X-API") != "api" {
			t.Errorf("Expected API middleware header")
		}

		// Test /api/v1/nested (should have global + api + v1 middleware)
		req2 := httptest.NewRequest("GET", "/api/v1/nested", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Header().Get("X-Global") != "global" {
			t.Errorf("Expected global middleware header in nested route")
		}

		if w2.Header().Get("X-API") != "api" {
			t.Errorf("Expected API middleware header in nested route")
		}

		if w2.Header().Get("X-V1") != "v1" {
			t.Errorf("Expected V1 middleware header in nested route")
		}
	})

	t.Run("group with path parameters", func(t *testing.T) {
		r := router.New()

		r.Group("/api/v1", func(v1 *router.Router) {
			v1.GET("/users/{id}", func(c *router.Context) {
				userID := c.Param("id")
				c.JSON(200, map[string]string{
					"user_id": userID,
					"version": "v1",
				})
			})

			v1.GET("/users/{id}/posts/{postId}", func(c *router.Context) {
				userID := c.Param("id")
				postID := c.Param("postId")
				c.JSON(200, map[string]string{
					"user_id": userID,
					"post_id": postID,
				})
			})
		})

		// Test single parameter
		req1 := httptest.NewRequest("GET", "/api/v1/users/123", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Code != 200 {
			t.Errorf("Expected status 200, got %d", w1.Code)
		}

		var response1 map[string]string
		if err := json.Unmarshal(w1.Body.Bytes(), &response1); err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}

		if response1["user_id"] != "123" {
			t.Errorf("Expected user_id to be '123', got '%s'", response1["user_id"])
		}

		// Test multiple parameters
		req2 := httptest.NewRequest("GET", "/api/v1/users/456/posts/789", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != 200 {
			t.Errorf("Expected status 200, got %d", w2.Code)
		}

		var response2 map[string]string
		if err := json.Unmarshal(w2.Body.Bytes(), &response2); err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}

		if response2["user_id"] != "456" {
			t.Errorf("Expected user_id to be '456', got '%s'", response2["user_id"])
		}

		if response2["post_id"] != "789" {
			t.Errorf("Expected post_id to be '789', got '%s'", response2["post_id"])
		}
	})

	t.Run("empty group prefix", func(t *testing.T) {
		r := router.New()

		r.Group("", func(group *router.Router) {
			group.GET("/test", func(c *router.Context) {
				c.JSON(200, map[string]string{"group": "empty"})
			})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("group with trailing slash", func(t *testing.T) {
		r := router.New()

		r.Group("/api/", func(api *router.Router) {
			api.GET("/users", func(c *router.Context) {
				c.JSON(200, map[string]string{"path": "with_trailing_slash"})
			})
		})

		// Both /api/users and /api/users/ should work due to path normalization
		req1 := httptest.NewRequest("GET", "/api/users", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Code != 200 {
			t.Errorf("Expected status 200 for /api/users, got %d", w1.Code)
		}

		req2 := httptest.NewRequest("GET", "/api/users/", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != 200 {
			t.Errorf("Expected status 200 for /api/users/, got %d", w2.Code)
		}
	})
}
