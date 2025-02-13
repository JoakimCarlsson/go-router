package list

import "github.com/joakimcarlsson/go-router/example/internal/features/todos/models"

type Response struct {
	Todos []models.Todo `json:"todos"`
	Total int          `json:"total"`
}