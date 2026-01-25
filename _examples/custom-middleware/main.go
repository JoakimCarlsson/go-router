package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

type contextKey string

const (
	requestIDKey contextKey = "requestID"
	userIDKey    contextKey = "userID"
)

func main() {
	r := router.New()

	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware)

	r.GET("/", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message":   "Hello!",
			"requestID": getRequestID(c.Request.Context()),
		})
	})

	r.Group("/api", func(api *router.Router) {
		api.Use(authMiddleware)

		api.GET("/profile", func(c *router.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"userID": getUserID(c.Request.Context()),
				"name":   "John Doe",
			})
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	var counter int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = time.Now().Format("20060102150405") + "-" + string(rune(counter))
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, "user-"+parts[1][:min(8, len(parts[1]))])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func getUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}
