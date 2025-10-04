package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/router/middleware/recovery"
)

func main() {
	r := router.New()

	r.Use(recovery.New(recovery.Config{
		EnableStackTrace: true,
		Logger: func(message string, fields map[string]interface{}) {
			log.Printf("[PANIC RECOVERED] %s", message)
			log.Printf("  Method: %v", fields["method"])
			log.Printf("  Path: %v", fields["path"])
			log.Printf("  Error: %v", fields["error"])
			if stack, ok := fields["stack"].(string); ok {
				log.Printf("  Stack trace:\n%s", stack)
			}
		},
		Handler: func(w http.ResponseWriter, r *http.Request, err interface{}) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"internal_server_error","message":"Something went wrong","timestamp":%d}`, time.Now().Unix())
		},
	}))

	r.GET("/", func(c *router.Context) {
		c.JSON(200, map[string]string{
			"message": "Welcome! Try these endpoints:",
			"panic":   "GET /panic - This will panic and be recovered",
			"divide":  "GET /divide/0 - Division by zero panic",
			"ok":      "GET /ok - Normal endpoint",
		})
	})

	r.GET("/ok", func(c *router.Context) {
		c.JSON(200, map[string]string{
			"status":  "ok",
			"message": "This endpoint works normally",
		})
	})

	r.GET("/panic", func(c *router.Context) {
		panic("This is an intentional panic to demonstrate recovery!")
	})

	r.GET("/divide/{n}", func(c *router.Context) {
		n, _ := c.ParamInt("n")
		result := 100 / n
		c.JSON(200, map[string]interface{}{
			"result": result,
		})
	})

	r.GET("/nil-pointer", func(c *router.Context) {
		var ptr *string
		//lint:ignore nilness Intentional nil dereference to demonstrate panic recovery
		c.JSON(200, map[string]string{"value": *ptr})
	})

	log.Println("Server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  GET http://localhost:8080/")
	log.Println("  GET http://localhost:8080/ok")
	log.Println("  GET http://localhost:8080/panic")
	log.Println("  GET http://localhost:8080/divide/10")
	log.Println("  GET http://localhost:8080/divide/0  (causes panic)")
	log.Println("  GET http://localhost:8080/nil-pointer (causes panic)")
	log.Println()
	log.Println("The server will continue running even after panics!")

	log.Fatal(http.ListenAndServe(":8080", r))
}
