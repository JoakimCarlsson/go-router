package recovery_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/router/middleware/recovery"
)

func TestRecovery_Default(t *testing.T) {
	r := router.New()
	r.Use(recovery.Default())

	r.GET("/panic", func(c *router.Context) {
		panic("test panic")
	})

	r.GET("/ok", func(c *router.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("Expected 'Internal Server Error' in response, got: %s", w.Body.String())
	}

	req2 := httptest.NewRequest("GET", "/ok", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("Expected server to continue working after panic, got status %d", w2.Code)
	}
}

func TestRecovery_WithCustomHandler(t *testing.T) {
	r := router.New()

	customHandler := func(w http.ResponseWriter, req *http.Request, err interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "panic_recovered",
			"detail": fmt.Sprintf("%v", err),
		})
	}

	r.Use(recovery.WithHandler(customHandler))

	r.GET("/panic", func(c *router.Context) {
		panic("custom panic message")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["error"] != "panic_recovered" {
		t.Errorf("Expected error 'panic_recovered', got '%s'", response["error"])
	}

	if !strings.Contains(response["detail"], "custom panic message") {
		t.Errorf("Expected detail to contain 'custom panic message', got '%s'", response["detail"])
	}
}

func TestRecovery_WithLogger(t *testing.T) {
	r := router.New()

	var loggedMessage string
	var loggedFields map[string]interface{}

	logger := func(message string, fields map[string]interface{}) {
		loggedMessage = message
		loggedFields = fields
	}

	r.Use(recovery.WithLogger(logger))

	r.GET("/panic", func(c *router.Context) {
		panic("logged panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if loggedMessage == "" {
		t.Error("Expected logger to be called")
	}

	if !strings.Contains(loggedMessage, "Panic recovered") {
		t.Errorf("Expected log message to contain 'Panic recovered', got: %s", loggedMessage)
	}

	if loggedFields == nil {
		t.Fatal("Expected logged fields to be set")
	}

	if loggedFields["error"] == nil {
		t.Error("Expected error field in logged fields")
	}

	if loggedFields["method"] != "GET" {
		t.Errorf("Expected method 'GET', got '%v'", loggedFields["method"])
	}

	if loggedFields["path"] != "/panic" {
		t.Errorf("Expected path '/panic', got '%v'", loggedFields["path"])
	}

	if loggedFields["stack"] == nil {
		t.Error("Expected stack trace in logged fields")
	}
}

func TestRecovery_WithConfig(t *testing.T) {
	r := router.New()

	var logCalled bool
	var handlerCalled bool

	r.Use(recovery.New(recovery.Config{
		EnableStackTrace: true,
		Logger: func(message string, fields map[string]interface{}) {
			logCalled = true
		},
		Handler: func(w http.ResponseWriter, req *http.Request, err interface{}) {
			handlerCalled = true
			w.WriteHeader(http.StatusTeapot)
			fmt.Fprintf(w, "Custom: %v", err)
		},
	}))

	r.GET("/panic", func(c *router.Context) {
		panic("test")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !logCalled {
		t.Error("Expected logger to be called")
	}

	if !handlerCalled {
		t.Error("Expected custom handler to be called")
	}

	if w.Code != http.StatusTeapot {
		t.Errorf("Expected status %d, got %d", http.StatusTeapot, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Custom: test") {
		t.Errorf("Expected custom message in response, got: %s", w.Body.String())
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	r := router.New()

	var logCalled bool

	r.Use(recovery.New(recovery.Config{
		Logger: func(message string, fields map[string]interface{}) {
			logCalled = true
		},
	}))

	r.GET("/ok", func(c *router.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if logCalled {
		t.Error("Logger should not be called when no panic occurs")
	}

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRecovery_MultipleRequests(t *testing.T) {
	r := router.New()
	r.Use(recovery.Default())

	panicCount := 0
	r.GET("/panic", func(c *router.Context) {
		panicCount++
		panic(fmt.Sprintf("panic #%d", panicCount))
	})

	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 500 {
			t.Errorf("Request %d: Expected status 500, got %d", i, w.Code)
		}
	}

	if panicCount != 5 {
		t.Errorf("Expected 5 panics to be recovered, got %d", panicCount)
	}
}

func TestRecovery_StringPanic(t *testing.T) {
	r := router.New()
	r.Use(recovery.Default())

	r.GET("/panic", func(c *router.Context) {
		panic("string panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRecovery_ErrorPanic(t *testing.T) {
	r := router.New()
	r.Use(recovery.Default())

	r.GET("/panic", func(c *router.Context) {
		panic(fmt.Errorf("error panic"))
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func BenchmarkRecovery_NoPanic(b *testing.B) {
	r := router.New()
	r.Use(recovery.Default())

	r.GET("/test", func(c *router.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRecovery_WithPanic(b *testing.B) {
	r := router.New()
	r.Use(recovery.Default())

	r.GET("/panic", func(c *router.Context) {
		panic("benchmark panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
