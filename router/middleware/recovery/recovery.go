package recovery

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// HandlerFunc defines a function that handles recovered panics.
// It receives the http.ResponseWriter, *http.Request, and the panic error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, err interface{})

// LoggerFunc defines a function for logging panic information.
// It receives a message and optional fields for structured logging.
type LoggerFunc func(message string, fields map[string]interface{})

// Config configures the recovery middleware behavior.
type Config struct {
	// Handler is called when a panic is recovered.
	// If nil, uses the default handler that returns 500 with "Internal Server Error".
	Handler HandlerFunc

	// EnableStackTrace includes stack traces in logs when panics occur.
	// Default: true
	EnableStackTrace bool

	// Logger is called to log panic information.
	// If nil, no logging occurs (panics are silently recovered).
	Logger LoggerFunc

	// EnablePrintStack prints the stack trace to stderr.
	// Useful for development but should be disabled in production if using structured logging.
	// Default: false
	EnablePrintStack bool
}

// DefaultConfig returns the default recovery configuration.
func DefaultConfig() Config {
	return Config{
		Handler:          nil,
		EnableStackTrace: true,
		Logger:           nil,
		EnablePrintStack: false,
	}
}

// New creates a new recovery middleware with the provided configuration.
// The middleware catches panics in handlers and prevents server crashes.
//
// Example usage:
//
//	r := router.New()
//	r.Use(recovery.New())
//
//	// With custom configuration
//	r.Use(recovery.New(recovery.Config{
//		EnableStackTrace: true,
//		Logger: func(message string, fields map[string]interface{}) {
//			log.Printf("%s: %v", message, fields)
//		},
//	}))
func New(config ...Config) func(http.Handler) http.Handler {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Handler == nil {
		cfg.Handler = defaultPanicHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					fields := make(map[string]interface{})
					fields["error"] = err
					fields["method"] = r.Method
					fields["path"] = r.URL.Path
					fields["remote_addr"] = r.RemoteAddr

					if cfg.EnableStackTrace {
						stack := debug.Stack()
						fields["stack"] = string(stack)

						if cfg.EnablePrintStack {
							fmt.Printf("Panic recovered: %v\nStack trace:\n%s\n", err, stack)
						}
					}

					if cfg.Logger != nil {
						cfg.Logger("Panic recovered in HTTP handler", fields)
					}

					cfg.Handler(w, r, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Default returns recovery middleware with default configuration.
// This is the recommended way to use the recovery middleware for most cases.
//
// Example:
//
//	r := router.New()
//	r.Use(recovery.Default())
func Default() func(http.Handler) http.Handler {
	return New()
}

// defaultPanicHandler is the default handler that returns a 500 error.
func defaultPanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// WithLogger creates recovery middleware with a custom logger.
// This is a convenience function for the common case of only wanting to add logging.
//
// Example:
//
//	r.Use(recovery.WithLogger(func(message string, fields map[string]interface{}) {
//		log.Printf("%s: %v", message, fields)
//	}))
func WithLogger(logger LoggerFunc) func(http.Handler) http.Handler {
	return New(Config{
		EnableStackTrace: true,
		Logger:           logger,
	})
}

// WithHandler creates recovery middleware with a custom panic handler.
// This is useful when you want to customize the error response format.
//
// Example:
//
//	r.Use(recovery.WithHandler(func(w http.ResponseWriter, r *http.Request, err interface{}) {
//		w.Header().Set("Content-Type", "application/json")
//		w.WriteHeader(http.StatusInternalServerError)
//		json.NewEncoder(w).Encode(map[string]string{
//			"error": "Internal Server Error",
//			"detail": fmt.Sprintf("%v", err),
//		})
//	}))
func WithHandler(handler HandlerFunc) func(http.Handler) http.Handler {
	return New(Config{
		Handler:          handler,
		EnableStackTrace: true,
	})
}
