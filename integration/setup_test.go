package integration

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joakimcarlsson/go-router/router"
)

func TestDefaultSetupOptions(t *testing.T) {
	opts := DefaultSetupOptions()

	// Test default values
	if opts.Title != "API Documentation" {
		t.Errorf("Expected default title 'API Documentation', got '%s'", opts.Title)
	}

	if opts.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", opts.Version)
	}

	if opts.Description != "API documentation powered by OpenAPI and Swagger UI" {
		t.Errorf("Expected default description, got '%s'", opts.Description)
	}

	if opts.SpecPath != "/openapi.json" {
		t.Errorf("Expected default spec path '/openapi.json', got '%s'", opts.SpecPath)
	}

	if opts.DocsPath != "/docs" {
		t.Errorf("Expected default docs path '/docs', got '%s'", opts.DocsPath)
	}

	if opts.DarkMode != false {
		t.Errorf("Expected default dark mode to be false, got %t", opts.DarkMode)
	}

	// Test that boolean security options default to false
	if opts.UseBasicAuth != false {
		t.Errorf("Expected UseBasicAuth to default to false, got %t", opts.UseBasicAuth)
	}

	if opts.UseBearerAuth != false {
		t.Errorf("Expected UseBearerAuth to default to false, got %t", opts.UseBearerAuth)
	}

	if opts.UseAPIKey != false {
		t.Errorf("Expected UseAPIKey to default to false, got %t", opts.UseAPIKey)
	}
}

func TestSetup_ValidOptions(t *testing.T) {
	r := router.New()

	opts := SetupOptions{
		Title:       "Test API",
		Version:     "2.0.0",
		Description: "Test API Description",
		SpecPath:    "/api/spec",
		DocsPath:    "/api/docs",
		DarkMode:    true,
		UITitle:     "Custom UI Title",
	}

	err := Setup(r, opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test that routes are properly configured
	// We can't easily test the actual route registration without exposing internals,
	// but we can test that the function completes without error
}

func TestSetup_DefaultPaths(t *testing.T) {
	r := router.New()

	opts := SetupOptions{
		Title:   "Test API",
		Version: "1.0.0",
		// Leave SpecPath and DocsPath empty to test defaults
	}

	err := Setup(r, opts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// The function should use default paths when empty strings are provided
}

func TestSetup_PathConflict(t *testing.T) {
	r := router.New()

	opts := SetupOptions{
		Title:    "Test API",
		Version:  "1.0.0",
		SpecPath: "/same-path",
		DocsPath: "/same-path", // Same as spec path - should cause error
	}

	err := Setup(r, opts)
	if err == nil {
		t.Error("Expected error for conflicting paths, got none")
	}

	expectedError := "spec path and docs path cannot be the same: /same-path"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSetup_WithSecuritySchemes(t *testing.T) {
	r := router.New()

	opts := SetupOptions{
		Title:         "Test API",
		Version:       "1.0.0",
		UseBasicAuth:  true,
		UseBearerAuth: true,
		UseAPIKey:     true,
	}

	err := Setup(r, opts)
	if err != nil {
		t.Errorf("Expected no error with security schemes, got: %v", err)
	}

	// The function should complete successfully with all security schemes enabled
}

func TestSetup_CustomUITitle(t *testing.T) {
	r := router.New()

	opts := SetupOptions{
		Title:   "API Title",
		Version: "1.0.0",
		UITitle: "Custom UI Title",
	}

	err := Setup(r, opts)
	if err != nil {
		t.Errorf("Expected no error with custom UI title, got: %v", err)
	}
}

func TestSetup_NoUITitle(t *testing.T) {
	r := router.New()

	opts := SetupOptions{
		Title:   "API Title",
		Version: "1.0.0",
		// UITitle is empty, should default to Title
	}

	err := Setup(r, opts)
	if err != nil {
		t.Errorf("Expected no error without UI title, got: %v", err)
	}
}

func TestSetupOptions_Validation(t *testing.T) {
	tests := []struct {
		name        string
		opts        SetupOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid options",
			opts: SetupOptions{
				Title:    "Test API",
				Version:  "1.0.0",
				SpecPath: "/spec",
				DocsPath: "/docs",
			},
			expectError: false,
		},
		{
			name: "conflicting paths",
			opts: SetupOptions{
				Title:    "Test API",
				Version:  "1.0.0",
				SpecPath: "/same",
				DocsPath: "/same",
			},
			expectError: true,
			errorMsg:    "spec path and docs path cannot be the same: /same",
		},
		{
			name: "empty title and version",
			opts: SetupOptions{
				Title:   "",
				Version: "",
			},
			expectError: false, // The setup function doesn't validate required fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.New()
			err := Setup(r, tt.opts)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

// Integration test to verify that the setup actually works with a simple router
func TestSetup_Integration(t *testing.T) {
	r := router.New()

	// Add a simple route to test with
	r.GET("/users/{id}", func(c *router.Context) {
		c.JSON(200, map[string]string{"id": c.Param("id")})
	})

	opts := SetupOptions{
		Title:       "Integration Test API",
		Version:     "1.0.0",
		Description: "Test API for integration testing",
		SpecPath:    "/openapi.json",
		DocsPath:    "/docs",
	}

	err := Setup(r, opts)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test that the main route still works
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"id":"123"`) {
		t.Errorf("Expected response to contain user ID, got: %s", w.Body.String())
	}
}

// Test that Setup works with complex configuration
func TestSetup_ComplexConfiguration(t *testing.T) {
	r := router.New()

	// Add multiple routes with different methods and patterns
	r.GET("/users", func(c *router.Context) {
		c.JSON(200, []string{"user1", "user2"})
	})

	r.POST("/users", func(c *router.Context) {
		c.JSON(201, map[string]string{"status": "created"})
	})

	r.PUT("/users/{id}", func(c *router.Context) {
		c.JSON(200, map[string]string{"id": c.Param("id"), "status": "updated"})
	})

	r.DELETE("/users/{id}", func(c *router.Context) {
		c.Status(204)
	})

	opts := SetupOptions{
		Title:         "Complex API",
		Version:       "2.1.0",
		Description:   "A complex API with multiple endpoints and security",
		SpecPath:      "/api-spec.json",
		DocsPath:      "/api-docs",
		DarkMode:      true,
		UITitle:       "Complex API Documentation",
		UseBasicAuth:  true,
		UseBearerAuth: true,
		UseAPIKey:     true,
	}

	err := Setup(r, opts)
	if err != nil {
		t.Fatalf("Setup with complex configuration failed: %v", err)
	}

	// Test that all original routes still work
	testCases := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{"GET", "/users", 200},
		{"POST", "/users", 201},
		{"PUT", "/users/456", 200},
		{"DELETE", "/users/789", 204},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d", tc.expectedCode, w.Code)
			}
		})
	}
}

// Test edge cases and error conditions
func TestSetup_EdgeCases(t *testing.T) {
	t.Run("nil router", func(t *testing.T) {
		// This would panic in the actual Setup function, which is expected behavior
		// We don't test this as it would crash the test, but it's worth noting
		// that Setup expects a valid router instance
	})

	t.Run("empty options", func(t *testing.T) {
		r := router.New()
		opts := SetupOptions{} // All fields empty

		err := Setup(r, opts)
		if err != nil {
			t.Errorf("Expected Setup to work with empty options (using defaults), got error: %v", err)
		}
	})

	t.Run("very long paths", func(t *testing.T) {
		r := router.New()
		longPath := "/" + strings.Repeat("a", 1000)

		opts := SetupOptions{
			Title:    "Test API",
			Version:  "1.0.0",
			SpecPath: longPath + "1",
			DocsPath: longPath + "2",
		}

		err := Setup(r, opts)
		if err != nil {
			t.Errorf("Expected Setup to work with long paths, got error: %v", err)
		}
	})

	t.Run("special characters in paths", func(t *testing.T) {
		r := router.New()

		opts := SetupOptions{
			Title:    "Test API",
			Version:  "1.0.0",
			SpecPath: "/api-spec.json",
			DocsPath: "/api_docs",
		}

		err := Setup(r, opts)
		if err != nil {
			t.Errorf("Expected Setup to work with special characters in paths, got error: %v", err)
		}
	})
}
