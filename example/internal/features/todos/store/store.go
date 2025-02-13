package store

import (
	"errors"

	"github.com/joakimcarlsson/go-router/example/internal/features/todos/models"
)

// ErrNotFound is returned when a todo is not found
var ErrNotFound = errors.New("todo not found")

type TodoStore interface {
	Create(todo *models.Todo) error
	Get(id string) (*models.Todo, error)
	List(limit, offset int, done *bool) ([]models.Todo, error)
	Update(todo *models.Todo) error
	Delete(id string) error
}
