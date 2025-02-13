package get

import (
	"errors"
	"net/http"

	"github.com/joakimcarlsson/go-router/example/internal/features/categories/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(categoryStore store.CategoryStore) router.HandlerFunc {
	return func(c *router.Context) {
		id := c.Param("id")
		req := Request{ID: id}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		category, err := categoryStore.Get(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get category"})
			return
		}

		c.JSON(http.StatusOK, Response{Category: category})
	}
}
