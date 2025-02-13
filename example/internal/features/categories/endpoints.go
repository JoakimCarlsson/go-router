package categories

import (
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/create"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/delete"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/get"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/list"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/store"
	"github.com/joakimcarlsson/go-router/example/internal/features/categories/update"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func RegisterRoutes(r *router.Router) {
	categoryStore := store.NewInMemoryStore()

	r.Group("/api/v1", func(r *router.Router) {
		r.Group("/categories", func(r *router.Router) {
			r.GET("/", list.NewHandler(categoryStore))
			r.POST("/", create.NewHandler(categoryStore))
			r.GET("/{id}", get.NewHandler(categoryStore))
			r.PATCH("/{id}", update.NewHandler(categoryStore))
			r.DELETE("/{id}", delete.NewHandler(categoryStore))
		})
	})
}
