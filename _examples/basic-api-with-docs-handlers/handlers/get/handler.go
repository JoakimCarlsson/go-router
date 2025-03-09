package get

import (
	"fmt"
	"net/http"

	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/store"
	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// GetProductResponse model for this handler's API responses
type GetProductResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price" validate:"min=0.01"`
	Category    string  `json:"category"`
	InStock     bool    `json:"inStock"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// StoreProductAdapter interfaces with the store
type StoreProductAdapter interface {
	GetProduct(id string) (store.Product, bool)
}

// convertStoreProduct converts a store.Product to this handler's Product type
func convertStoreProduct(p store.Product) GetProductResponse {
	createdAt := p.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	updatedAt := p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	return GetProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Category:    p.Category,
		InStock:     p.InStock,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// Handler retrieves a product by ID
func Handler(store StoreProductAdapter) router.HandlerFunc {
	return func(c *router.Context) {
		id := c.Param("id")

		// Get product from store
		storeProduct, found := store.GetProduct(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]string{
				"error": fmt.Sprintf("Product with ID '%s' not found", id),
			})
			return
		}

		// Convert to handler's product type
		product := convertStoreProduct(storeProduct)
		c.JSON(http.StatusOK, product)
	}
}

// RouteOptions returns the route options for this handler
func RouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Products"),
		docs.WithSummary("Get product by ID"),
		docs.WithDescription("Returns a specific product by its ID"),
		docs.WithPathParam("id", "string", true, "Product ID", "1"),
		docs.WithResponse(200, "Product found"),
		docs.WithJSONResponse[GetProductResponse](200, "Product details"),
		docs.WithResponse(404, "Product not found"),
	}
}
