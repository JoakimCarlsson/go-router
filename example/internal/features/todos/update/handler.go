package update

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

		todo, err := todoStore.Get(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, map[string]string{"error": "todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get todo"})
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

		if req.Title != nil {
			todo.Title = *req.Title
		}
		if req.Description != nil {
			todo.Description = *req.Description
		}
		if req.Completed != nil {
			todo.Completed = *req.Completed
		}
		todo.UpdatedAt = time.Now()

		if err := todoStore.Update(todo); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update todo"})
			return
		}

		c.JSON(http.StatusOK, Response{Todo: todo})
	}
}
