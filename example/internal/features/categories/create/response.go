package create

import "github.com/joakimcarlsson/go-router/example/internal/features/categories/models"

type Response struct {
	Category *models.Category `json:"category"`
}
