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

	profiles := outputcache.NewProfiles()
	profiles.Add("aggressive", time.Hour, outputcache.VaryByPath(), outputcache.SlidingExpiration())
	profiles.Add("api", 5*time.Minute, outputcache.WithRevalidation())

	cache := outputcache.New(outputcache.Config{
		DefaultDuration: 5 * time.Minute,
		Profiles:        profiles,
	})

	r.Use(outputcache.WithRouter(r))
	r.Use(cache.Middleware())

	r.GET("/", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Output Cache Example",
			"info":    "Try the various endpoints to see different caching strategies",
		})
	})

	r.GET("/products", listProducts).
		WithOutputCache(time.Minute, outputcache.Tags("products", "product:list"))

	r.GET("/products/{id}", getProduct).
		WithOutputCache(time.Hour,
			outputcache.Tags("products", "product:detail"),
			outputcache.VaryByPath(),
			outputcache.WithRevalidation())

	r.GET("/users/{id}", getUser).
		WithProfile("aggressive")

	r.GET("/search", search).
		WithOutputCache(30*time.Second,
			outputcache.VaryByQuery("q", "page"),
			outputcache.VaryByHeader("Accept-Language"))

	r.GET("/prices", getPrices).
		WithOutputCache(time.Minute,
			outputcache.VaryByCustom(func(req *http.Request) string {
				role := req.Header.Get("X-User-Role")
				if role == "" {
					role = "guest"
				}
				return role
			}))

	r.GET("/conditional", getConditional).
		WithOutputCache(time.Minute,
			outputcache.CacheWhen(func(status int, headers http.Header) bool {
				return status == 200 && headers.Get("X-No-Cache") == ""
			}))

	r.GET("/compressed", getCompressed).
		WithOutputCache(time.Minute, outputcache.VaryByEncoding())

	r.POST("/products", createProduct)

	r.POST("/invalidate", func(c *router.Context) {
		cache.InvalidateTag("products")
		c.JSON(http.StatusOK, map[string]string{"message": "Products cache invalidated"})
	})

	r.GET("/no-cache", noCacheHandler)

	log.Println("Server starting on :8080")
	log.Println("Endpoints:")
	log.Println("  GET  /products           - Cached with tags")
	log.Println("  GET  /products/{id}      - Cached with tags, ETag, varies by path")
	log.Println("  GET  /users/{id}         - Uses 'aggressive' profile (sliding expiration)")
	log.Println("  GET  /search?q=...       - Varies by query and language")
	log.Println("  GET  /prices             - Varies by X-User-Role header (custom)")
	log.Println("  GET  /conditional        - Conditionally cached")
	log.Println("  GET  /compressed         - Varies by Accept-Encoding")
	log.Println("  POST /products           - Not cached (POST method)")
	log.Println("  POST /invalidate         - Invalidates products tag")
	log.Println("  GET  /no-cache           - Not cached")
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

func getProduct(c *router.Context) {
	id := c.Param("id")
	log.Printf("Handler: getProduct called for ID %s\n", id)

	product := Product{
		ID:    1,
		Name:  "Product " + id,
		Price: 999,
	}
	c.JSON(http.StatusOK, product)
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

func getPrices(c *router.Context) {
	role := c.Request.Header.Get("X-User-Role")
	if role == "" {
		role = "guest"
	}
	log.Printf("Handler: getPrices called for role %s\n", role)

	prices := map[string]int{
		"guest":   100,
		"member":  80,
		"premium": 60,
	}

	price := prices[role]
	if price == 0 {
		price = 100
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"role":  role,
		"price": price,
	})
}

func getConditional(c *router.Context) {
	log.Println("Handler: getConditional called")

	shouldNotCache := c.Query().Get("nocache")
	if shouldNotCache == "true" {
		c.SetHeader("X-No-Cache", "true")
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data":      "conditional",
		"timestamp": time.Now().Unix(),
	})
}

func getCompressed(c *router.Context) {
	log.Println("Handler: getCompressed called")
	data := make([]byte, 1000)
	for i := range data {
		data[i] = 'A'
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"data": string(data),
	})
}

func noCacheHandler(c *router.Context) {
	log.Println("Handler: noCacheHandler called")
	c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "This endpoint is not cached",
		"timestamp": time.Now().Unix(),
	})
}
