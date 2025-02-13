package get

import (
	"net/http"
	"errors"

	"github.com/joakimcarlsson/go-router/example/internal/features/todos/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(todoStore store.TodoStore) router.HandlerFunc {
	return func(c *router.Context) {
		id := c.Param("id")
		req := Request{ID: id}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		todo, err := todoStore.Get(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, map[string]string{"error": "todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get todo"})
			return
		}

		c.JSON(http.StatusOK, Response{Todo: todo})
	}
}
