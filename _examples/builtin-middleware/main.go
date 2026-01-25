package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/router/middleware/recovery"
	"github.com/joakimcarlsson/go-router/router/middleware/security"
)

func main() {
	r := router.New()

	r.Use(recovery.New(recovery.Config{
		EnableStackTrace: true,
		Logger: func(message string, fields map[string]interface{}) {
			log.Printf("[PANIC] %s: %v", message, fields["error"])
		},
		Handler: func(w http.ResponseWriter, r *http.Request, err interface{}) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"internal_server_error","timestamp":%d}`, time.Now().Unix())
		},
	}))

	r.Use(security.Headers(security.Config{
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		FrameOptions:          "DENY",
		ContentTypeNosniff:    true,
		XSSProtection:         true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	r.GET("/", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	r.GET("/panic", func(c *router.Context) {
		panic("intentional panic")
	})

	r.GET("/panic/nil", func(c *router.Context) {
		var ptr *string
		c.JSON(http.StatusOK, map[string]string{"value": *ptr})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
