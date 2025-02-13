package list

import (
	"net/http"
	"strconv"

	"github.com/joakimcarlsson/go-router/example/internal/features/todos/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(store store.TodoStore) router.HandlerFunc {
	return func(c *router.Context) {
		limit, _ := strconv.Atoi(c.QueryDefault("limit", "10"))
		offset, _ := strconv.Atoi(c.QueryDefault("offset", "0"))
		var done *bool
		if doneStr := c.QueryDefault("done", ""); doneStr != "" {
			isDone := doneStr == "true"
			done = &isDone
		}

		req := Request{
			Limit:  limit,
			Offset: offset,
		}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		todos, err := store.List(req.Limit, req.Offset, done)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list todos"})
			return
		}
		
		c.JSON(http.StatusOK, Response{
			Todos: todos,
			Total: len(todos),
		})
	}
}
