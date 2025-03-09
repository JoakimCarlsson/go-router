package edit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/store"
	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// EditProductResponse model for this handler's API responses
type EditProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description *string   `json:"description"`
	Price       float64   `json:"price" validate:"min=0.01"`
	Category    string    `json:"category"`
	InStock     bool      `json:"inStock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price" validate:"min=0.01"`
	Category    string  `json:"category"`
	InStock     bool    `json:"inStock"`
}

// StoreProductAdapter interfaces with the store
type StoreProductAdapter interface {
	GetProduct(id string) (store.Product, bool)
	UpdateProduct(product store.Product) bool
}

// convertToStoreProduct converts a request to store.Product
func convertToStoreProduct(id string, req UpdateProductRequest) store.Product {
	return store.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		InStock:     req.InStock,
	}
}

// convertFromStoreProduct converts store.Product to this handler's Product type
func convertFromStoreProduct(p store.Product) EditProductResponse {
	return EditProductResponse{
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

// Handler updates an existing product
func Handler(store StoreProductAdapter) router.HandlerFunc {
	return func(c *router.Context) {
		id := c.Param("id")

		// Check if product exists
		_, found := store.GetProduct(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]string{
				"error": fmt.Sprintf("Product with ID '%s' not found", id),
			})
			return
		}

		var request UpdateProductRequest
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

		// Convert to store product and update
		storeProduct := convertToStoreProduct(id, request)
		store.UpdateProduct(storeProduct)

		// Get updated product
		updatedStoreProduct, _ := store.GetProduct(id)
		updatedProduct := convertFromStoreProduct(updatedStoreProduct)

		c.JSON(http.StatusOK, updatedProduct)
	}
}

// RouteOptions returns the route options for this handler
func RouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Products"),
		docs.WithSummary("Update product"),
		docs.WithDescription("Updates an existing product"),
		docs.WithPathParam("id", "string", true, "Product ID", "1"),
		docs.WithJSONRequestBody[UpdateProductRequest](true, "Updated product information"),
		docs.WithResponse(200, "Product updated successfully"),
		docs.WithJSONResponse[EditProductResponse](200, "Updated product"),
		docs.WithResponse(400, "Invalid product data"),
		docs.WithResponse(404, "Product not found"),
	}
}
