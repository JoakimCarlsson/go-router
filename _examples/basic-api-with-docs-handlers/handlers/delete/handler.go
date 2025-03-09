package delete

import (
	"fmt"
	"net/http"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/router"
)

// StoreProductAdapter interfaces with the store
type StoreProductAdapter interface {
	DeleteProduct(id string) bool
}

// Handler deletes a product by ID
func Handler(store StoreProductAdapter) router.HandlerFunc {
	return func(c *router.Context) {
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
}

// RouteOptions returns the route options for this handler
func RouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Products"),
		docs.WithSummary("Delete product"),
		docs.WithDescription("Deletes a product from the catalog"),
		docs.WithPathParam("id", "string", true, "Product ID", "1"),
		docs.WithResponse(204, "Product deleted successfully"),
		docs.WithResponse(404, "Product not found"),
	}
}
