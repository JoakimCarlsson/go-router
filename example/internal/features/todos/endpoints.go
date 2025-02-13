package todos

import (
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/create"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/delete"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/get"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/list"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/store"
	"github.com/joakimcarlsson/go-router/example/internal/features/todos/update"
	"github.com/joakimcarlsson/go-router/pkg/router"
)

func RegisterRoutes(r *router.Router) {
	todoStore := store.NewInMemoryStore()

	r.Group("/api/v1", func(r *router.Router) {
		r.Group("/todos", func(r *router.Router) {
			r.GET("/", list.NewHandler(todoStore))
			r.POST("/", create.NewHandler(todoStore))
			r.GET("/{id}", get.NewHandler(todoStore))
			r.PATCH("/{id}", update.NewHandler(todoStore))
			r.DELETE("/{id}", delete.NewHandler(todoStore))
		})
	})
}
