package main

import (
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/router/middleware/cors"
)

func main() {
	r := router.New()

	r.Use(cors.Handler(cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	r.Group("/", func(public *router.Router) {
		public.GET("/", func(c *router.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"message": "Hello from go-router with default CORS middleware!",
			})
		})

		public.GET("/public", func(c *router.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"message": "This is a public endpoint with default CORS settings",
			})
		})
	})

	r.AutoRegisterOptions()

	log.Fatal(http.ListenAndServe(":8080", r))
}
