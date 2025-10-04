/*
Package recovery provides panic recovery middleware for the go-router.

The recovery middleware catches panics in HTTP handlers and prevents server crashes,
allowing the server to continue handling requests gracefully after a panic occurs.

# Basic Usage

The simplest way to use the recovery middleware is with the Default() function:

	r := router.New()
	r.Use(recovery.Default())

	r.GET("/api/endpoint", func(c *router.Context) {
		// If this handler panics, the middleware will catch it
		// and return a 500 Internal Server Error
		panic("something went wrong")
	})

# Custom Logger

You can add logging to track when panics occur:

	r.Use(recovery.WithLogger(func(message string, fields map[string]interface{}) {
		log.Printf("[PANIC] %s: %v", message, fields)
	}))

The logger receives structured fields including:
  - error: the panic value
  - method: HTTP method of the request
  - path: URL path of the request
  - remote_addr: client IP address
  - stack: stack trace (if EnableStackTrace is true)

# Custom Error Handler

You can customize the error response format:

	r.Use(recovery.WithHandler(func(w http.ResponseWriter, r *http.Request, err interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Internal Server Error",
			"message": fmt.Sprintf("%v", err),
			"timestamp": time.Now().Unix(),
		})
	}))

# Full Configuration

For complete control, use the New() function with a Config:

	r.Use(recovery.New(recovery.Config{
		EnableStackTrace: true,
		EnablePrintStack: false, // Set to true for development
		Logger: func(message string, fields map[string]interface{}) {
			// Your structured logging implementation
			logger.Error(message, fields)
		},
		Handler: func(w http.ResponseWriter, r *http.Request, err interface{}) {
			// Your custom error response
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Custom error response"))
		},
	}))

# Integration with Built-in Recovery

The router has built-in panic recovery in the handlerWrapper that prevents context
pool corruption. This middleware provides an additional layer that allows you to:
  - Log panics with custom loggers
  - Customize error responses
  - Add panic monitoring/alerting
  - Collect panic statistics

The built-in recovery ensures safety; this middleware adds observability and control.

# Best Practices

 1. Always use recovery middleware in production to prevent unexpected panics from
    crashing your server.

 2. Enable stack traces in development but consider disabling EnablePrintStack in
    production if using structured logging.

 3. Use a custom logger to send panic information to your monitoring system
    (e.g., Sentry, DataDog, CloudWatch).

 4. Keep custom error handlers simple - they run in a recovered panic state,
    so avoid operations that might panic again.

 5. Place recovery middleware early in the middleware chain so it can catch
    panics from other middleware and handlers.

# Performance

The recovery middleware has minimal overhead when no panics occur:
  - Uses defer/recover which is optimized by the Go runtime
  - Stack trace collection only happens when a panic is caught
  - Logger and handler are only called on actual panics

Benchmark results show negligible impact on normal request handling.

# Example: Production Setup

	package main

	import (
		"log"
		"net/http"

		"github.com/joakimcarlsson/go-router/router"
		"github.com/joakimcarlsson/go-router/router/middleware/recovery"
	)

	func main() {
		r := router.New()

		// Add recovery middleware with custom logger
		r.Use(recovery.New(recovery.Config{
			EnableStackTrace: true,
			Logger: func(message string, fields map[string]interface{}) {
				// Send to your logging service
				log.Printf("[PANIC] %s", message)
				for k, v := range fields {
					log.Printf("  %s: %v", k, v)
				}
				// Also send to alerting system
				// alerting.SendAlert("panic_recovered", fields)
			},
		}))

		// Your routes
		r.GET("/api/users", getUsersHandler)

		http.ListenAndServe(":8080", r)
	}
*/
package recovery
