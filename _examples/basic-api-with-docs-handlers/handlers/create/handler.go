package create

import (
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/store"
	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// CreateProductResponse model for this handler's API responses
type CreateProductResponse struct {
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

// StoreProductAdapter interfaces with the store
type StoreProductAdapter interface {
	AddProduct(product store.Product) string
	GetProduct(id string) (store.Product, bool)
}

// convertToStoreProduct converts a request to store.Product
func convertToStoreProduct(req NewProductRequest) store.Product {
	return store.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		InStock:     req.InStock,
	}
}

// convertFromStoreProduct converts store.Product to this handler's Product type
func convertFromStoreProduct(p store.Product) CreateProductResponse {
	return CreateProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Category:    p.Category,
		InStock:     p.InStock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// Handler creates a new product
func Handler(store StoreProductAdapter) router.HandlerFunc {
	return func(c *router.Context) {
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

		// Convert request to store product and add it
		storeProduct := convertToStoreProduct(request)
		id := store.AddProduct(storeProduct)

		// Get the created product from store
		createdStoreProduct, _ := store.GetProduct(id)

		// Convert to handler's product type
		createdProduct := convertFromStoreProduct(createdStoreProduct)

		c.JSON(http.StatusCreated, createdProduct)
	}
}

// RouteOptions returns the route options for this handler
func RouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Products"),
		docs.WithSummary("Create a product"),
		docs.WithDescription("Creates a new product in the catalog"),
		docs.WithJSONRequestBody[NewProductRequest](true, "Product information"),
		docs.WithResponse(201, "Product created successfully"),
		docs.WithJSONResponse[CreateProductResponse](201, "Created product"),
		docs.WithResponse(400, "Invalid product data"),
	}
}
