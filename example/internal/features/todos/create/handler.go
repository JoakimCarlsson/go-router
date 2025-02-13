package create

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/models"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/store"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func NewHandler(store store.TodoStore) router.HandlerFunc {
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
		todo := &models.Todo{
			ID:          uuid.New().String(),
			Title:       req.Title,
			Description: req.Description,
			Completed:   false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := store.Create(todo); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create todo"})
			return
		}

		c.JSON(http.StatusCreated, Response{Todo: todo})
	}
}
