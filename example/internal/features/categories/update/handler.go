package update

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

		category, err := categoryStore.Get(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get category"})
			return
		}

		var req Request
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if req.Name != nil {
			category.Name = *req.Name
		}
		if req.Description != nil {
			category.Description = *req.Description
		}
		category.UpdatedAt = time.Now()

		if err := categoryStore.Update(category); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update category"})
			return
		}

		c.JSON(http.StatusOK, Response{Category: category})
	}
}
