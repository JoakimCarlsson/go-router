package get_all

import (
	"net/http"
	"strconv"

	"github.com/joakimcarlsson/go-router/_examples/basic-api-with-docs-handlers/store"
	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// GetAllProductResponse model for this handler's API responses
type GetAllProductResponse struct {
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
	GetProducts() []store.Product
}

// convertStoreProduct converts a store.Product to this handler's Product type
func convertStoreProduct(p store.Product) GetAllProductResponse {
	createdAt := p.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	updatedAt := p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	return GetAllProductResponse{
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

// Handler lists all products with optional filtering
func Handler(store StoreProductAdapter) router.HandlerFunc {
	return func(c *router.Context) {
		// Get all products from store
		storeProducts := store.GetProducts()

		// Convert store products to handler product type
		products := make([]GetAllProductResponse, 0, len(storeProducts))
		for _, p := range storeProducts {
			products = append(products, convertStoreProduct(p))
		}

		// Apply category filter if provided
		if category := c.QueryDefault("category", ""); category != "" {
			filtered := make([]GetAllProductResponse, 0)
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
				filtered := make([]GetAllProductResponse, 0)
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
}

// RouteOptions returns the route options for this handler
func RouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Products"),
		docs.WithSummary("List all products"),
		docs.WithDescription("Returns a list of all products in the catalog"),
		docs.WithQueryParam("category", "string", false, "Filter products by category", "Electronics"),
		docs.WithQueryParam("inStock", "boolean", false, "Filter by stock availability", true),
		docs.WithResponse(200, "Products retrieved successfully"),
		docs.WithJSONResponse[[]GetAllProductResponse](200, "List of products"),
	}
}
