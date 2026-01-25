package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/outputcache"
	"github.com/joakimcarlsson/go-router/router"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	r := router.New()

	cache := outputcache.New(outputcache.Config{
		DefaultDuration: 5 * time.Minute,
	})

	r.Use(outputcache.WithRouter(r))
	r.Use(cache.Middleware())

	r.GET("/", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Output Cache Example",
			"info":    "Try the /products, /users/{id}, and /search endpoints",
		})
	})

	r.GET("/products", listProducts).
		WithOutputCache(time.Minute)

	r.GET("/users/{id}", getUser).
		WithOutputCache(time.Hour, outputcache.VaryByPath())

	r.GET("/search", search).
		WithOutputCache(30*time.Second,
			outputcache.VaryByQuery("q", "page"),
			outputcache.VaryByHeader("Accept-Language"))

	r.POST("/products", createProduct)

	r.GET("/no-cache", noCacheHandler)

	log.Println("Server starting on :8080")
	log.Println("Endpoints:")
	log.Println("  GET  /products          - Cached for 1 minute")
	log.Println("  GET  /users/{id}        - Cached for 1 hour, varies by ID")
	log.Println("  GET  /search?q=...      - Cached for 30 seconds, varies by query and language")
	log.Println("  POST /products          - Not cached (POST method)")
	log.Println("  GET  /no-cache          - Not cached (no WithOutputCache)")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func listProducts(c *router.Context) {
	log.Println("Handler: listProducts called")
	products := []Product{
		{ID: 1, Name: "Laptop", Price: 999},
		{ID: 2, Name: "Mouse", Price: 25},
		{ID: 3, Name: "Keyboard", Price: 75},
	}
	c.JSON(http.StatusOK, products)
}

func getUser(c *router.Context) {
	id := c.Param("id")
	log.Printf("Handler: getUser called for ID %s\n", id)

	user := User{
		ID:       1,
		Username: "user" + id,
		Email:    "user" + id + "@example.com",
	}
	c.JSON(http.StatusOK, user)
}

func search(c *router.Context) {
	query := c.Query().Get("q")
	page := c.Query().Get("page")
	lang := c.Request.Header.Get("Accept-Language")

	log.Printf("Handler: search called with q=%s, page=%s, lang=%s\n", query, page, lang)

	c.JSON(http.StatusOK, map[string]interface{}{
		"query":    query,
		"page":     page,
		"language": lang,
		"results":  []string{"result1", "result2", "result3"},
	})
}

func createProduct(c *router.Context) {
	log.Println("Handler: createProduct called")
	var product Product
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

func noCacheHandler(c *router.Context) {
	log.Println("Handler: noCacheHandler called")
	c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "This endpoint is not cached",
		"timestamp": time.Now().Unix(),
	})
}
