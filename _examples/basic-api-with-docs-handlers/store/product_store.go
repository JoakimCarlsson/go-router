package store

import (
	"strconv"
	"sync"
	"time"
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
	description := "High quality wireless earbuds with noise cancellation"
	store.AddProduct(Product{
		ID:          "1",
		Name:        "Wireless Earbuds",
		Description: &description,
		Price:       129.99,
		Category:    "Electronics",
		InStock:     true,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-15 * 24 * time.Hour),
	})

	description = "Comfortable running shoes with excellent support"
	store.AddProduct(Product{
		ID:          "2",
		Name:        "Running Shoes",
		Description: &description,
		Price:       89.99,
		Category:    "Footwear",
		InStock:     true,
		CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
	})

	description = "Premium coffee maker with timer and temperature control"
	store.AddProduct(Product{
		ID:          "3",
		Name:        "Coffee Maker",
		Description: &description,
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
