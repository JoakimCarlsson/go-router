package router_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
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
		c.Writer.Write([]byte("Hello, World!"))
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
	// Setup simple middlewares
	loggingMiddleware := func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			// Just call next handler
			next(c)
		}
	}

	authMiddleware := func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			// Simulate auth check
			if c.GetHeader("Authorization") != "" {
				c.Set("user", "authenticated")
			}
			next(c)
		}
	}

	b.Run("NoMiddleware", func(b *testing.B) {
		r := router.New()
		r.GET("/hello", func(c *router.Context) {
			c.Writer.WriteHeader(200)
			c.Writer.Write([]byte("Hello, World!"))
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
		r.Use(loggingMiddleware)
		r.GET("/hello", func(c *router.Context) {
			c.Writer.WriteHeader(200)
			c.Writer.Write([]byte("Hello, World!"))
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
		r.Use(loggingMiddleware)
		r.Use(authMiddleware)
		r.Use(loggingMiddleware) // Add a third middleware
		r.GET("/hello", func(c *router.Context) {
			c.Writer.WriteHeader(200)
			c.Writer.Write([]byte("Hello, World!"))
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
		r.Use(loggingMiddleware)

		r.Group("/api", func(api *router.Router) {
			api.Use(authMiddleware)

			api.GET("/hello", func(c *router.Context) {
				c.Writer.WriteHeader(200)
				c.Writer.Write([]byte("Hello, World!"))
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
		r.Use(func(next router.HandlerFunc) router.HandlerFunc {
			return func(c *router.Context) {
				c.Set("key1", "value1")
				c.Set("key2", 123)
				c.Set("key3", true)
				next(c)
			}
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
	r.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) {
			c.SetHeader("X-Response-Time", "0.1ms") // Simulated response time
			next(c)
		}
	})

	// Product API
	r.Group("/api/products", func(api *router.Router) {
		// Auth middleware
		api.Use(func(next router.HandlerFunc) router.HandlerFunc {
			return func(c *router.Context) {
				auth := c.GetHeader("Authorization")
				if auth != "" {
					c.Set("userID", "user-123")
				}
				next(c)
			}
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
