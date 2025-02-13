package list

import "github.com/joakimcarlsson/go-router/example/internal/features/categories/models"

type Response struct {
	Categories []models.Category `json:"categories"`
	Total      int               `json:"total"`
}
