package router

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// Test types for binding tests
type User struct {
	ID    int    `json:"id" xml:"id" form:"id"`
	Name  string `json:"name" xml:"name" form:"name"`
	Email string `json:"email" xml:"email" form:"email"`
}

type FileUpload struct {
	File        *multipart.FileHeader `form:"file" file:"true"`
	Name        string                `form:"name"`
	Description string                `form:"description"`
}

func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/?name=john&age=25&active=true", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	// Test basic query retrieval
	query := ctx.Query()
	if query.Get("name") != "john" {
		t.Errorf("Expected name to be 'john', got '%s'", query.Get("name"))
	}
}

func TestContext_QueryDefault(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "existing parameter",
			url:          "/?name=john",
			key:          "name",
			defaultValue: "default",
			expected:     "john",
		},
		{
			name:         "missing parameter",
			url:          "/?other=value",
			key:          "name",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty parameter value",
			url:          "/?name=",
			key:          "name",
			defaultValue: "default",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result := ctx.QueryDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestContext_QueryInt(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		key         string
		expected    int
		expectError bool
	}{
		{
			name:        "valid integer",
			url:         "/?age=25",
			key:         "age",
			expected:    25,
			expectError: false,
		},
		{
			name:        "invalid integer",
			url:         "/?age=abc",
			key:         "age",
			expected:    0,
			expectError: true,
		},
		{
			name:        "missing parameter",
			url:         "/?other=value",
			key:         "age",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result, err := ctx.QueryInt(tt.key)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestContext_QueryIntDefault(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			url:          "/?age=25",
			key:          "age",
			defaultValue: 30,
			expected:     25,
		},
		{
			name:         "invalid integer uses default",
			url:          "/?age=abc",
			key:          "age",
			defaultValue: 30,
			expected:     30,
		},
		{
			name:         "missing parameter uses default",
			url:          "/?other=value",
			key:          "age",
			defaultValue: 30,
			expected:     30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result := ctx.QueryIntDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestContext_QueryBool(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		key         string
		expected    bool
		expectError bool
	}{
		{
			name:        "true value",
			url:         "/?active=true",
			key:         "active",
			expected:    true,
			expectError: false,
		},
		{
			name:        "false value",
			url:         "/?active=false",
			key:         "active",
			expected:    false,
			expectError: false,
		},
		{
			name:        "1 value",
			url:         "/?active=1",
			key:         "active",
			expected:    true,
			expectError: false,
		},
		{
			name:        "0 value",
			url:         "/?active=0",
			key:         "active",
			expected:    false,
			expectError: false,
		},
		{
			name:        "invalid boolean",
			url:         "/?active=maybe",
			key:         "active",
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result, err := ctx.QueryBool(tt.key)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestContext_QueryBoolDefault(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		key          string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "valid true",
			url:          "/?active=true",
			key:          "active",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "valid false",
			url:          "/?active=false",
			key:          "active",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "invalid boolean uses default",
			url:          "/?active=maybe",
			key:          "active",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "missing parameter uses default",
			url:          "/?other=value",
			key:          "active",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result := ctx.QueryBoolDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestContext_Param(t *testing.T) {
	// Create a test request with path parameters
	req := httptest.NewRequest("GET", "/users/123", nil)
	req.SetPathValue("id", "123")
	req.SetPathValue("name", "john")

	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	// Test existing parameter
	if ctx.Param("id") != "123" {
		t.Errorf("Expected id to be '123', got '%s'", ctx.Param("id"))
	}

	if ctx.Param("name") != "john" {
		t.Errorf("Expected name to be 'john', got '%s'", ctx.Param("name"))
	}

	// Test missing parameter
	if ctx.Param("missing") != "" {
		t.Errorf("Expected missing param to be empty, got '%s'", ctx.Param("missing"))
	}

	// Test with nil request
	ctx.Request = nil
	if ctx.Param("id") != "" {
		t.Errorf("Expected empty string when request is nil, got '%s'", ctx.Param("id"))
	}
}

func TestContext_ParamInt(t *testing.T) {
	tests := []struct {
		name        string
		paramValue  string
		key         string
		expected    int
		expectError bool
	}{
		{
			name:        "valid integer",
			paramValue:  "123",
			key:         "id",
			expected:    123,
			expectError: false,
		},
		{
			name:        "invalid integer",
			paramValue:  "abc",
			key:         "id",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty parameter",
			paramValue:  "",
			key:         "id",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.SetPathValue(tt.key, tt.paramValue)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result, err := ctx.ParamInt(tt.key)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestContext_ParamIntDefault(t *testing.T) {
	tests := []struct {
		name         string
		paramValue   string
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			paramValue:   "123",
			key:          "id",
			defaultValue: 456,
			expected:     123,
		},
		{
			name:         "invalid integer uses default",
			paramValue:   "abc",
			key:          "id",
			defaultValue: 456,
			expected:     456,
		},
		{
			name:         "empty parameter uses default",
			paramValue:   "",
			key:          "id",
			defaultValue: 456,
			expected:     456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.SetPathValue(tt.key, tt.paramValue)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result := ctx.ParamIntDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestContext_ParamBool(t *testing.T) {
	tests := []struct {
		name        string
		paramValue  string
		key         string
		expected    bool
		expectError bool
	}{
		{
			name:        "true value",
			paramValue:  "true",
			key:         "active",
			expected:    true,
			expectError: false,
		},
		{
			name:        "false value",
			paramValue:  "false",
			key:         "active",
			expected:    false,
			expectError: false,
		},
		{
			name:        "invalid boolean",
			paramValue:  "maybe",
			key:         "active",
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.SetPathValue(tt.key, tt.paramValue)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result, err := ctx.ParamBool(tt.key)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestContext_ParamBoolDefault(t *testing.T) {
	tests := []struct {
		name         string
		paramValue   string
		key          string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "valid true",
			paramValue:   "true",
			key:          "active",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "valid false",
			paramValue:   "false",
			key:          "active",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "invalid boolean uses default",
			paramValue:   "maybe",
			key:          "active",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.SetPathValue(tt.key, tt.paramValue)
			w := httptest.NewRecorder()
			ctx := newContext(w, req, time.Now())

			result := ctx.ParamBoolDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestContext_JSON(t *testing.T) {
	user := User{ID: 1, Name: "John", Email: "john@example.com"}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.JSON(200, user)

	// Check status code
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Check content type
	expectedContentType := "application/json; charset=utf-8"
	if w.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, w.Header().Get("Content-Type"))
	}

	// Check response body
	var responseUser User
	if err := json.Unmarshal(w.Body.Bytes(), &responseUser); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if responseUser != user {
		t.Errorf("Expected %+v, got %+v", user, responseUser)
	}
}

func TestContext_JSON_Error(t *testing.T) {
	// Test with a type that can't be marshaled to JSON (channel)
	invalidData := make(chan int)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.JSON(200, invalidData)

	// Should return 500 status code due to marshal error
	if w.Code != 500 {
		t.Errorf("Expected status code 500, got %d", w.Code)
	}
}

func TestContext_XML(t *testing.T) {
	user := User{ID: 1, Name: "John", Email: "john@example.com"}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.XML(200, user)

	// Check status code
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Check content type
	expectedContentType := "application/xml; charset=utf-8"
	if w.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, w.Header().Get("Content-Type"))
	}

	// Check response body
	var responseUser User
	if err := xml.Unmarshal(w.Body.Bytes(), &responseUser); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if responseUser != user {
		t.Errorf("Expected %+v, got %+v", user, responseUser)
	}
}

func TestContext_Data(t *testing.T) {
	data := []byte("Hello, World!")
	contentType := "text/plain"

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.Data(200, contentType, data)

	// Check status code
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != contentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", contentType, w.Header().Get("Content-Type"))
	}

	// Check response body
	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Errorf("Expected body '%s', got '%s'", string(data), w.Body.String())
	}
}

func TestContext_File(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "Hello, World!"
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.File(tempFile.Name())

	// Check response body
	if w.Body.String() != testContent {
		t.Errorf("Expected body '%s', got '%s'", testContent, w.Body.String())
	}
}

func TestContext_Redirect(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.Redirect(302, "/new-location")

	// Check status code
	if w.Code != 302 {
		t.Errorf("Expected status code 302, got %d", w.Code)
	}

	// Check location header
	expectedLocation := "/new-location"
	if w.Header().Get("Location") != expectedLocation {
		t.Errorf("Expected Location header '%s', got '%s'", expectedLocation, w.Header().Get("Location"))
	}
}

func TestContext_Error(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	errorMessage := "Something went wrong"
	ctx.Error(500, errorMessage)

	// Check status code
	if w.Code != 500 {
		t.Errorf("Expected status code 500, got %d", w.Code)
	}

	// Check response body
	expectedBody := errorMessage + "\n" // http.Error adds a newline
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, w.Body.String())
	}
}

func TestContext_Status(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.Status(201)

	// Check status code
	if w.Code != 201 {
		t.Errorf("Expected status code 201, got %d", w.Code)
	}

	// Check that the context stores the status code
	if ctx.StatusCode != 201 {
		t.Errorf("Expected context status code 201, got %d", ctx.StatusCode)
	}
}

func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	// Test existing headers
	if ctx.GetHeader("Authorization") != "Bearer token123" {
		t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", ctx.GetHeader("Authorization"))
	}

	if ctx.GetHeader("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header 'application/json', got '%s'", ctx.GetHeader("Content-Type"))
	}

	// Test missing header
	if ctx.GetHeader("X-Missing") != "" {
		t.Errorf("Expected missing header to be empty, got '%s'", ctx.GetHeader("X-Missing"))
	}
}

func TestContext_SetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	ctx.SetHeader("X-Custom", "custom-value")
	ctx.SetHeader("Cache-Control", "no-cache")

	// Check that headers were set
	if w.Header().Get("X-Custom") != "custom-value" {
		t.Errorf("Expected X-Custom header 'custom-value', got '%s'", w.Header().Get("X-Custom"))
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Expected Cache-Control header 'no-cache', got '%s'", w.Header().Get("Cache-Control"))
	}
}

func TestContext_BindJSON(t *testing.T) {
	user := User{ID: 1, Name: "John", Email: "john@example.com"}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	var boundUser User
	if err := ctx.BindJSON(&boundUser); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if boundUser != user {
		t.Errorf("Expected %+v, got %+v", user, boundUser)
	}
}

func TestContext_BindJSON_Error(t *testing.T) {
	// Invalid JSON
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	var user User
	if err := ctx.BindJSON(&user); err == nil {
		t.Error("Expected error for invalid JSON, got none")
	}
}

func TestContext_BindXML(t *testing.T) {
	user := User{ID: 1, Name: "John", Email: "john@example.com"}
	xmlData, _ := xml.Marshal(user)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(xmlData))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	var boundUser User
	if err := ctx.BindXML(&boundUser); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if boundUser != user {
		t.Errorf("Expected %+v, got %+v", user, boundUser)
	}
}

func TestContext_BindForm(t *testing.T) {
	form := url.Values{}
	form.Set("id", "1")
	form.Set("name", "John")
	form.Set("email", "john@example.com")

	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	ctx := newContext(w, req, time.Now())

	var user User
	if err := ctx.BindForm(&user); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedUser := User{ID: 1, Name: "John", Email: "john@example.com"}
	if user != expectedUser {
		t.Errorf("Expected %+v, got %+v", expectedUser, user)
	}
}

// Helper function to create a context instance for testing
// This mirrors the newContext function from the router package
func newContext(w http.ResponseWriter, r *http.Request, startTime time.Time) *Context {
	ctx := &Context{
		Writer:             w,
		Request:            r,
		StartTime:          startTime,
		StatusCode:         200,
		store:              make(map[string]interface{}),
		maxMultipartMemory: 32 << 20, // 32 MB
	}
	return ctx
}