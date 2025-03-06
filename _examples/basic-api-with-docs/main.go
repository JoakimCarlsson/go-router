package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// Product represents a product in the catalog
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description *string   `json:"description"`
	Price       float64   `json:"price" validate:"min=0.01"`
	Category    string    `json:"category"`
	InStock     bool      `json:"inStock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewProductRequest represents a request to create a product
type NewProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price" validate:"min=0.01"`
	Category    string  `json:"category"`
	InStock     bool    `json:"inStock"`
}

// ProductStore is a simple in-memory store for products
type ProductStore struct {
	mu       sync.RWMutex
	products map[string]Product
	counter  int
}

// NewProductStore creates a new product store with sample data
func NewProductStore() *ProductStore {
	store := &ProductStore{
		products: make(map[string]Product),
		counter:  100,
	}

	// Add some sample products
	store.AddProduct(Product{
		ID:          "1",
		Name:        "Wireless Earbuds",
		Description: nil,
		Price:       129.99,
		Category:    "Electronics",
		InStock:     true,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-15 * 24 * time.Hour),
	})

	store.AddProduct(Product{
		ID:          "2",
		Name:        "Running Shoes",
		Description: nil,
		Price:       89.99,
		Category:    "Footwear",
		InStock:     true,
		CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
	})

	store.AddProduct(Product{
		ID:          "3",
		Name:        "Coffee Maker",
		Description: nil,
		Price:       74.50,
		Category:    "Kitchen",
		InStock:     false,
		CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour),
	})

	return store
}

// GetProducts returns all products
func (s *ProductStore) GetProducts() []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]Product, 0, len(s.products))
	for _, product := range s.products {
		products = append(products, product)
	}
	return products
}

// GetProduct returns a product by ID
func (s *ProductStore) GetProduct(id string) (Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, found := s.products[id]
	return product, found
}

// AddProduct adds a product to the store
func (s *ProductStore) AddProduct(product Product) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if product.ID == "" {
		s.counter++
		product.ID = strconv.Itoa(s.counter)
	}

	now := time.Now()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	product.UpdatedAt = now

	s.products[product.ID] = product
	return product.ID
}

// UpdateProduct updates a product in the store
func (s *ProductStore) UpdateProduct(product Product) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[product.ID]; !exists {
		return false
	}

	// Preserve creation time
	product.CreatedAt = s.products[product.ID].CreatedAt
	product.UpdatedAt = time.Now()

	s.products[product.ID] = product
	return true
}

// DeleteProduct deletes a product from the store
func (s *ProductStore) DeleteProduct(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return false
	}

	delete(s.products, id)
	return true
}

func main() {
	store := NewProductStore()
	r := router.New()

	// Logger middleware
	r.Use(loggerMiddleware)

	// Public endpoints
	r.GET("/health", healthCheck,
		docs.WithTags("System"),
		docs.WithSummary("Health check endpoint"),
		docs.WithDescription("Returns the health status of the API"),
		docs.WithResponse(200, "API is healthy"),
	)

	// Products endpoints
	r.GET("/products", func(c *router.Context) { listProducts(c, store) },
		docs.WithTags("Products"),
		docs.WithSummary("List all products"),
		docs.WithDescription("Returns a list of all products in the catalog"),
		docs.WithQueryParam("category", "string", false, "Filter products by category", "Electronics"),
		docs.WithQueryParam("inStock", "boolean", false, "Filter by stock availability", true),
		docs.WithResponse(200, "Products retrieved successfully"),
		docs.WithJSONResponse[[]Product](200, "List of products"),
	)

	r.POST("/products", func(c *router.Context) { createProduct(c, store) },
		docs.WithTags("Products"),
		docs.WithSummary("Create a product"),
		docs.WithDescription("Creates a new product in the catalog"),
		docs.WithJSONRequestBody[NewProductRequest](true, "Product information"),
		docs.WithResponse(201, "Product created successfully"),
		docs.WithJSONResponse[Product](201, "Created product"),
		docs.WithResponse(400, "Invalid product data"),
	)

	r.GET("/products/{id}", func(c *router.Context) { getProduct(c, store) },
		docs.WithTags("Products"),
		docs.WithSummary("Get product by ID"),
		docs.WithDescription("Returns a specific product by its ID"),
		docs.WithPathParam("id", "string", true, "Product ID", "1"),
		docs.WithResponse(200, "Product found"),
		docs.WithJSONResponse[Product](200, "Product details"),
		docs.WithResponse(404, "Product not found"),
	)

	r.PUT("/products/{id}", func(c *router.Context) { updateProduct(c, store) },
		docs.WithTags("Products"),
		docs.WithSummary("Update product"),
		docs.WithDescription("Updates an existing product"),
		docs.WithPathParam("id", "string", true, "Product ID", "1"),
		docs.WithJSONRequestBody[NewProductRequest](true, "Updated product information"),
		docs.WithResponse(200, "Product updated successfully"),
		docs.WithJSONResponse[Product](200, "Updated product"),
		docs.WithResponse(400, "Invalid product data"),
		docs.WithResponse(404, "Product not found"),
	)

	r.DELETE("/products/{id}", func(c *router.Context) { deleteProduct(c, store) },
		docs.WithTags("Products"),
		docs.WithSummary("Delete product"),
		docs.WithDescription("Deletes a product from the catalog"),
		docs.WithPathParam("id", "string", true, "Product ID", "1"),
		docs.WithResponse(204, "Product deleted successfully"),
		docs.WithResponse(404, "Product not found"),
	)

	// Create OpenAPI generator
	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Product Catalog API",
		Version:     "1.0.0",
		Description: "A sample product catalog API built with go-router",
	})

	// Configure Swagger UI with specific settings for pointer fields
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.DefaultModelRendering = "example"
	uiConfig.Title = "Product Catalog API"

	// Set up the integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Middleware for logging requests
func loggerMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		start := time.Now()

		// Process request
		next(c)

		// Log after request is processed
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s - %d (%v)\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.StatusCode,
			duration,
		)
	}
}

// Handler implementations
func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func listProducts(c *router.Context, store *ProductStore) {
	// Get all products
	products := store.GetProducts()

	// Apply category filter if provided
	if category := c.QueryDefault("category", ""); category != "" {
		filtered := make([]Product, 0)
		for _, product := range products {
			if product.Category == category {
				filtered = append(filtered, product)
			}
		}
		products = filtered
	}

	// Apply inStock filter if provided
	if inStockParam := c.QueryDefault("inStock", ""); inStockParam != "" {
		inStock, err := strconv.ParseBool(inStockParam)
		if err == nil {
			filtered := make([]Product, 0)
			for _, product := range products {
				if product.InStock == inStock {
					filtered = append(filtered, product)
				}
			}
			products = filtered
		}
	}

	c.JSON(http.StatusOK, products)
}

func getProduct(c *router.Context, store *ProductStore) {
	id := c.Param("id")
	product, found := store.GetProduct(id)

	if !found {
		c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Product with ID '%s' not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func createProduct(c *router.Context, store *ProductStore) {
	var request NewProductRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Simple validation
	if request.Name == "" {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name is required",
		})
		return
	}

	if request.Price <= 0 {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Price must be greater than 0",
		})
		return
	}

	product := Product{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
		InStock:     request.InStock,
	}

	id := store.AddProduct(product)
	createdProduct, _ := store.GetProduct(id)

	c.JSON(http.StatusCreated, createdProduct)
}

func updateProduct(c *router.Context, store *ProductStore) {
	id := c.Param("id")

	_, found := store.GetProduct(id)

	if !found {
		c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Product with ID '%s' not found", id),
		})
		return
	}

	var request NewProductRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Simple validation
	if request.Name == "" {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name is required",
		})
		return
	}

	if request.Price <= 0 {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Price must be greater than 0",
		})
		return
	}

	product := Product{
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
		InStock:     request.InStock,
	}

	store.UpdateProduct(product)
	updatedProduct, _ := store.GetProduct(id)

	c.JSON(http.StatusOK, updatedProduct)
}

func deleteProduct(c *router.Context, store *ProductStore) {
	id := c.Param("id")

	found := store.DeleteProduct(id)

	if !found {
		c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Product with ID '%s' not found", id),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
