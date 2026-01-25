package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

// Product represents a product in the catalog
type Product struct {
	ID          uuid.UUID `json:"id"`
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
	products map[uuid.UUID]Product
	counter  int
}

// NewProductStore creates a new product store with sample data
func NewProductStore() *ProductStore {
	store := &ProductStore{
		products: make(map[uuid.UUID]Product),
		counter:  100,
	}

	store.AddProduct(Product{
		ID:          uuid.New(),
		Name:        "Wireless Earbuds",
		Description: nil,
		Price:       129.99,
		Category:    "Electronics",
		InStock:     true,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-15 * 24 * time.Hour),
	})

	store.AddProduct(Product{
		ID:          uuid.New(),
		Name:        "Running Shoes",
		Description: nil,
		Price:       89.99,
		Category:    "Footwear",
		InStock:     true,
		CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
	})

	store.AddProduct(Product{
		ID:          uuid.New(),
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
func (s *ProductStore) GetProduct(id uuid.UUID) (Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, found := s.products[id]
	return product, found
}

// AddProduct adds a product to the store
func (s *ProductStore) AddProduct(product Product) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	if product.ID == uuid.Nil {
		product.ID = uuid.New()
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

	product.CreatedAt = s.products[product.ID].CreatedAt
	product.UpdatedAt = time.Now()

	s.products[product.ID] = product
	return true
}

// DeleteProduct deletes a product from the store
func (s *ProductStore) DeleteProduct(id uuid.UUID) bool {
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

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, req)

			duration := time.Since(start)
			fmt.Printf("[%s] %s %s - (%v)\n",
				time.Now().Format("2006-01-02 15:04:05"),
				req.Method,
				req.URL.Path,
				duration,
			)
		})
	})

	r.GET("/health", healthCheck,
		openapi.WithTags("System"),
		openapi.WithSummary("Health check endpoint"),
		openapi.WithDescription("Returns the health status of the API"),
		openapi.WithResponse(200, "API is healthy"),
	)

	r.GET("/products", func(c *router.Context) { listProducts(c, store) },
		openapi.WithTags("Products"),
		openapi.WithSummary("List all products"),
		openapi.WithDescription("Returns a list of all products in the catalog"),
		openapi.WithQueryParam("category", "string", false, "Filter products by category", "Electronics"),
		openapi.WithQueryParam("inStock", "boolean", false, "Filter by stock availability", true),
		openapi.WithResponse(200, "Products retrieved successfully"),
		openapi.WithJSONResponse[[]Product](200, "List of products"),
	)

	r.POST("/products", func(c *router.Context) { createProduct(c, store) },
		openapi.WithTags("Products"),
		openapi.WithSummary("Create a product"),
		openapi.WithDescription("Creates a new product in the catalog"),
		openapi.WithJSONRequestBody[NewProductRequest](true, "Product information"),
		openapi.WithResponse(201, "Product created successfully"),
		openapi.WithJSONResponse[Product](201, "Created product"),
		openapi.WithResponse(400, "Invalid product data"),
	)

	r.POST("/products/batch", func(c *router.Context) { createProductBatch(c, store) },
		openapi.WithTags("Products"),
		openapi.WithSummary("Creates a batch of products"),
		openapi.WithDescription("Creates multiple products in the catalog"),
		openapi.WithJSONRequestBody[[]NewProductRequest](true, "Product information"),
		openapi.WithResponse(201, "Product created successfully"),
		openapi.WithJSONResponse[[]Product](201, "Created product"),
		openapi.WithResponse(400, "Invalid product data"),
	)

	r.GET("/products/{id}", func(c *router.Context) { getProduct(c, store) },
		openapi.WithTags("Products"),
		openapi.WithSummary("Get product by ID"),
		openapi.WithDescription("Returns a specific product by its ID"),
		openapi.WithPathParam("id", "string", true, "Product ID", "6B29FC40-CA47-1067-B31D-00DD010662DA"),
		openapi.WithResponse(200, "Product found"),
		openapi.WithJSONResponse[Product](200, "Product details"),
		openapi.WithResponse(404, "Product not found"),
	)

	r.PUT("/products/{id}", func(c *router.Context) { updateProduct(c, store) },
		openapi.WithTags("Products"),
		openapi.WithSummary("Update product"),
		openapi.WithDescription("Updates an existing product"),
		openapi.WithPathParam("id", "string", true, "Product ID", "6B29FC40-CA47-1067-B31D-00DD010662DA"),
		openapi.WithJSONRequestBody[NewProductRequest](true, "Updated product information"),
		openapi.WithResponse(200, "Product updated successfully"),
		openapi.WithJSONResponse[Product](200, "Updated product"),
		openapi.WithResponse(400, "Invalid product data"),
		openapi.WithResponse(404, "Product not found"),
	)

	r.DELETE("/products/{id}", func(c *router.Context) { deleteProduct(c, store) },
		openapi.WithTags("Products"),
		openapi.WithSummary("Delete product"),
		openapi.WithDescription("Deletes a product from the catalog"),
		openapi.WithPathParam("id", "string", true, "Product ID", "6B29FC40-CA47-1067-B31D-00DD010662DA"),
		openapi.WithResponse(204, "Product deleted successfully"),
		openapi.WithResponse(404, "Product not found"),
	)

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Product Catalog API",
		Version:     "1.0.0",
		Description: "A sample product catalog API built with go-router",
	})

	uiConfig := swaggerui.DefaultUIConfig()
	uiConfig.DefaultModelRendering = "example"
	uiConfig.Title = "Product Catalog API"

	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func createProductBatch(c *router.Context, store *ProductStore) {
	var requests []NewProductRequest
	if err := c.BindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	products := make([]Product, 0, len(requests))
	for _, request := range requests {
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
		products = append(products, createdProduct)
	}

	c.JSON(http.StatusCreated, products)
}

func listProducts(c *router.Context, store *ProductStore) {
	products := store.GetProducts()

	if category := c.QueryDefault("category", ""); category != "" {
		filtered := make([]Product, 0)
		for _, product := range products {
			if product.Category == category {
				filtered = append(filtered, product)
			}
		}
		products = filtered
	}

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
	uuid := uuid.MustParse(id)
	product, found := store.GetProduct(uuid)

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
	uuid := uuid.MustParse(id)

	_, found := store.GetProduct(uuid)

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
		ID:          uuid,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
		InStock:     request.InStock,
	}

	store.UpdateProduct(product)
	updatedProduct, _ := store.GetProduct(uuid)

	c.JSON(http.StatusOK, updatedProduct)
}

func deleteProduct(c *router.Context, store *ProductStore) {
	id := c.Param("id")
	uuid := uuid.MustParse(id)

	found := store.DeleteProduct(uuid)

	if !found {
		c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Product with ID '%s' not found", id),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
