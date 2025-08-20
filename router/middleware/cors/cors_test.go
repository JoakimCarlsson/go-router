package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
)

func TestCORSDefault(t *testing.T) {
	r := router.New()
	r.Use(Default())
	r.GET("/", func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("OK"))
	})

	// Auto-register OPTIONS handlers for all registered routes
	r.AutoRegisterOptions()

	// Test simple request with Origin header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Errorf("Expected no Access-Control-Allow-Credentials header, got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}

	// Test preflight request
	req = httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status code 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if methods := w.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}
}

func TestCORSCustom(t *testing.T) {
	r := router.New()
	r.Use(Handler(Options{
		AllowOrigins:     []string{"https://example.com", "https://sub.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"X-Custom-Header", "Content-Type"},
		ExposeHeaders:    []string{"X-Custom-Response-Header"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	r.GET("/", func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("OK"))
	})

	// Auto-register OPTIONS handlers for all registered routes
	r.AutoRegisterOptions()

	// Test with allowed origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials: true, got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}
	if w.Header().Get("Access-Control-Expose-Headers") != "X-Custom-Response-Header" {
		t.Errorf("Expected Access-Control-Expose-Headers: X-Custom-Response-Header, got %s", w.Header().Get("Access-Control-Expose-Headers"))
	}

	// Test preflight request
	req = httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status code 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Errorf("Expected Access-Control-Allow-Methods: GET, POST, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "X-Custom-Header, Content-Type" {
		t.Errorf("Expected Access-Control-Allow-Headers: X-Custom-Header, Content-Type, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials: true, got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}
	if w.Header().Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("Expected Access-Control-Max-Age: 3600, got %s", w.Header().Get("Access-Control-Max-Age"))
	}

	// Test with not allowed origin
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://not-allowed.com")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSWildcardOrigin(t *testing.T) {
	r := router.New()
	r.Use(Handler(Options{
		AllowOrigins: []string{"https://*.example.com"},
	}))
	r.GET("/", func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("OK"))
	})

	// Test with subdomain that matches wildcard
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://sub.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://sub.example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://sub.example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test with nested subdomain that matches wildcard
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://nested.sub.example.com")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://nested.sub.example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://nested.sub.example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test with admin subdomain pattern
	r2 := router.New()
	r2.Use(Handler(Options{
		AllowOrigins: []string{"https://*.admin.example.com"},
	}))
	r2.GET("/", func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("OK"))
	})

	// Should match subdomain of admin.example.com
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://dashboard.admin.example.com")
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://dashboard.admin.example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://dashboard.admin.example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Should NOT match admin.example.com (no subdomain)
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://admin.example.com")
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Should NOT match sub.example.com (different domain)
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://sub.example.com")
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSOptionsPassthrough(t *testing.T) {
	r := router.New()
	r.Use(Handler(Options{
		AllowOrigins:       []string{"https://example.com"},
		OptionsPassthrough: true,
	}))

	var optionsHandlerCalled bool
	// Use a non-root path to avoid normalization issues
	r.GET("/test", func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("GET handler"))
	})

	r.Handle("OPTIONS /test", func(c *router.Context) {
		optionsHandlerCalled = true
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("OPTIONS handler called"))
	})

	// Test OPTIONS request with OptionsPassthrough enabled
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !optionsHandlerCalled {
		t.Error("Expected OPTIONS handler to be called with OptionsPassthrough enabled")
	}
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	if w.Body.String() != "OPTIONS handler called" {
		t.Errorf("Expected body 'OPTIONS handler called', got %s", w.Body.String())
	}
}

func TestGetAllowOrigin(t *testing.T) {
	testCases := []struct {
		name          string
		origin        string
		allowOrigins  []string
		expectedValue string
	}{
		{
			name:          "Exact match",
			origin:        "https://example.com",
			allowOrigins:  []string{"https://example.com"},
			expectedValue: "https://example.com",
		},
		{
			name:          "Wildcard all",
			origin:        "https://example.com",
			allowOrigins:  []string{"*"},
			expectedValue: "*",
		},
		{
			name:          "Wildcard subdomain match",
			origin:        "https://sub.example.com",
			allowOrigins:  []string{"https://*.example.com"},
			expectedValue: "https://sub.example.com",
		},
		{
			name:          "Wildcard subdomain no match",
			origin:        "https://example.org",
			allowOrigins:  []string{"https://*.example.com"},
			expectedValue: "",
		},
		{
			name:          "Multiple origins - first match",
			origin:        "https://example.com",
			allowOrigins:  []string{"https://foo.com", "https://example.com", "https://bar.com"},
			expectedValue: "https://example.com",
		},
		{
			name:          "No match",
			origin:        "https://example.com",
			allowOrigins:  []string{"https://foo.com", "https://bar.com"},
			expectedValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getAllowOrigin(tc.origin, tc.allowOrigins)
			if result != tc.expectedValue {
				t.Errorf("Expected %q, got %q", tc.expectedValue, result)
			}
		})
	}
}

func TestPreflightRequests(t *testing.T) {
	r := router.New()
	r.Use(Handler(Options{
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut},
		AllowHeaders:     []string{"X-Custom-Header", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	r.GET("/", func(c *router.Context) {
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write([]byte("OK"))
	})

	// Auto-register OPTIONS handlers for all registered routes
	r.AutoRegisterOptions()

	// Test proper preflight request
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status code 204 for preflight request, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT" {
		t.Errorf("Expected Access-Control-Allow-Methods: GET, POST, PUT, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "X-Custom-Header, Content-Type" {
		t.Errorf("Expected Access-Control-Allow-Headers: X-Custom-Header, Content-Type, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}

	// Test OPTIONS request that is not a preflight request (missing Access-Control-Request-Method)
	req = httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	// No Access-Control-Request-Method header
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// It should still have CORS headers and be treated as a preflight in the updated implementation
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin: https://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	// In the updated implementation, all OPTIONS requests get Allow-Methods even without Access-Control-Request-Method
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT" {
		t.Errorf("Expected Access-Control-Allow-Methods: GET, POST, PUT, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
}

// TestOptionsWithoutOrigin tests handling OPTIONS requests without an Origin header
func TestOptionsWithoutOrigin(t *testing.T) {
	r := router.New()
	r.Use(Handler(Options{
		AllowOrigins:     []string{"https://example.com", "https://*.admin.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/test", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "success"})
	})

	r.AutoRegisterOptions()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	// No Origin header set deliberately

	r.ServeHTTP(w, req)

	// Check that we get correct headers even without Origin
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if w.Header().Get("Vary") != "Origin" {
		t.Errorf("Expected Vary: Origin, got %s", w.Header().Get("Vary"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected non-empty Access-Control-Allow-Methods")
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected non-empty Access-Control-Allow-Origin")
	}
}

// TestConsistentOriginHandling ensures the same origin is used in preflight and actual requests
func TestConsistentOriginHandling(t *testing.T) {
	r := router.New()
	origin := "https://test.admin.example.com"

	r.Use(Handler(Options{
		AllowOrigins:     []string{"https://*.admin.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowCredentials: true,
	}))

	r.GET("/test", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "success"})
	})

	r.AutoRegisterOptions()

	// First send OPTIONS request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req1.Header.Set("Origin", origin)

	r.ServeHTTP(w1, req1)

	// Then send GET request
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("Origin", origin)

	r.ServeHTTP(w2, req2)

	// Both should have the same Access-Control-Allow-Origin header
	if w1.Header().Get("Access-Control-Allow-Origin") != origin {
		t.Errorf("Expected OPTIONS response to have origin %s, got %s", origin, w1.Header().Get("Access-Control-Allow-Origin"))
	}
	if w2.Header().Get("Access-Control-Allow-Origin") != origin {
		t.Errorf("Expected GET response to have origin %s, got %s", origin, w2.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestNullOrigin tests handling of requests with null origin
func TestNullOrigin(t *testing.T) {
	r := router.New()
	r.Use(Handler(Options{
		// Restricted CORS - only specific domains allowed
		AllowOrigins:     []string{"https://*.admin.example.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowCredentials: true,
	}))

	r.GET("/secure", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "secure data"})
	})

	r.AutoRegisterOptions()

	// Test OPTIONS with null origin (file:// protocol or similar)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/secure", nil)
	req.Header.Set("Origin", "null") // Browser sends "null" as literal string
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	// Should not get a permissive Access-Control-Allow-Origin
	if w.Header().Get("Access-Control-Allow-Origin") == "*" {
		t.Error("Expected not to get wildcard Access-Control-Allow-Origin for null origin with restricted CORS")
	}

	// Should still set Access-Control-Allow-Methods so browser knows what's possible
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods to be set for null origin")
	}

	// A second router with * allowed should still allow null origins
	r2 := router.New()
	r2.Use(Handler(Options{
		AllowOrigins: []string{"*"},
	}))

	r2.GET("/public", func(c *router.Context) {})
	r2.AutoRegisterOptions()

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodOptions, "/public", nil)
	req2.Header.Set("Origin", "null")
	req2.Header.Set("Access-Control-Request-Method", "GET")
	r2.ServeHTTP(w2, req2)

	// Should get wildcard for permissive configuration
	if w2.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected wildcard Access-Control-Allow-Origin for null origin with permissive CORS, got %s",
			w2.Header().Get("Access-Control-Allow-Origin"))
	}
}
