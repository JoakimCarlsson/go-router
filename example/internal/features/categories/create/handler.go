package create

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/models"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(store store.CategoryStore) router.HandlerFunc {
	return func(c *router.Context) {
		var req Request
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		now := time.Now()
		category := &models.Category{
			ID:          uuid.New().String(),
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := store.Create(category); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
			return
		}

		c.JSON(http.StatusCreated, Response{Category: category})
	}
}
