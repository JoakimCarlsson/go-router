package delete

import (
	"errors"
	"net/http"

	"github.com/joakimcarlsson/go-router/example/internal/features/todos/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(todoStore store.TodoStore) router.HandlerFunc {
	return func(c *router.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "missing todo ID"})
			return
		}

		err := todoStore.Delete(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, map[string]string{"error": "todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete todo"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
