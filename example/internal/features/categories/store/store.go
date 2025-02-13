package store

import (
	"errors"

	"github.com/joakimcarlsson/go-router/example/internal/features/categories/models"
)

var ErrNotFound = errors.New("category not found")

type CategoryStore interface {
	Create(category *models.Category) error
	Get(id string) (*models.Category, error)
	List(limit, offset int) ([]models.Category, error)
	Update(category *models.Category) error
	Delete(id string) error
}
