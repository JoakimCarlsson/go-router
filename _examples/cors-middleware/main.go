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
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://myapp.example.com",
			"https://*.example.com",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	r.GET("/", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Hello from go-router with CORS!",
		})
	})

	r.GET("/api/data", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"items": []string{"item1", "item2", "item3"},
		})
	})

	r.POST("/api/data", func(c *router.Context) {
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		c.JSON(http.StatusCreated, body)
	})

	r.AutoRegisterOptions()

	log.Fatal(http.ListenAndServe(":8080", r))
}
