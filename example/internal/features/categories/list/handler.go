package list

import (
	"net/http"
	"strconv"

	"github.com/joakimcarlsson/go-router/example/internal/features/categories/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(store store.CategoryStore) router.HandlerFunc {
	return func(c *router.Context) {
		limit, _ := strconv.Atoi(c.QueryDefault("limit", "10"))
		offset, _ := strconv.Atoi(c.QueryDefault("offset", "0"))

		req := Request{
			Limit:  limit,
			Offset: offset,
		}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		categories, err := store.List(req.Limit, req.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list categories"})
			return
		}

		c.JSON(http.StatusOK, Response{
			Categories: categories,
			Total:      len(categories),
		})
	}
}
