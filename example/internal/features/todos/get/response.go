package get

import "github.com/joakimcarlsson/go-router/example/internal/features/todos/models"

type Response struct {
	Todo *models.Todo `json:"todo"`
}