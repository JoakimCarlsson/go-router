package delete

import (
	"errors"
	"net/http"

	"github.com/joakimcarlsson/go-router/example/internal/features/categories/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(categoryStore store.CategoryStore) router.HandlerFunc {
	return func(c *router.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "missing category ID"})
			return
		}

		err := categoryStore.Delete(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete category"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
