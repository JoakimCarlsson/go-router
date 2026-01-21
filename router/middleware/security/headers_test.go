package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaders_Default(t *testing.T) {
	handler := Default()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	tests := []struct {
		header   string
		expected string
	}{
		{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tc := range tests {
		t.Run(tc.header, func(t *testing.T) {
			got := rec.Header().Get(tc.header)
			if got != tc.expected {
				t.Errorf("%s: expected %q, got %q", tc.header, tc.expected, got)
			}
		})
	}
}

func TestHeaders_HSTS(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "basic HSTS",
			config: Config{
				HSTSMaxAge: 3600,
			},
			expected: "max-age=3600",
		},
		{
			name: "HSTS with subdomains",
			config: Config{
				HSTSMaxAge:            3600,
				HSTSIncludeSubdomains: true,
			},
			expected: "max-age=3600; includeSubDomains",
		},
		{
			name: "HSTS with preload",
			config: Config{
				HSTSMaxAge:            31536000,
				HSTSIncludeSubdomains: true,
				HSTSPreload:           true,
			},
			expected: "max-age=31536000; includeSubDomains; preload",
		},
		{
			name: "HSTS disabled",
			config: Config{
				HSTSMaxAge: 0,
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := Headers(tc.config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("Strict-Transport-Security")
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestHeaders_FrameOptions(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"DENY", "DENY", "DENY"},
		{"SAMEORIGIN", "SAMEORIGIN", "SAMEORIGIN"},
		{"disabled", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := Headers(Config{FrameOptions: tc.value})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("X-Frame-Options")
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestHeaders_ContentTypeNosniff(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected string
	}{
		{"enabled", true, "nosniff"},
		{"disabled", false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := Headers(Config{ContentTypeNosniff: tc.enabled})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("X-Content-Type-Options")
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestHeaders_CustomCSP(t *testing.T) {
	csp := "default-src 'self'; script-src 'self' 'unsafe-inline'"

	handler := Headers(Config{
		ContentSecurityPolicy: csp,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("Content-Security-Policy")
	if got != csp {
		t.Errorf("expected %q, got %q", csp, got)
	}
}

func TestHeaders_PermissionsPolicy(t *testing.T) {
	policy := "geolocation=(), microphone=()"

	handler := Headers(Config{
		PermissionsPolicy: policy,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("Permissions-Policy")
	if got != policy {
		t.Errorf("expected %q, got %q", policy, got)
	}
}

func TestHeaders_ReferrerPolicy(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"no-referrer", "no-referrer", "no-referrer"},
		{"strict-origin-when-cross-origin", "strict-origin-when-cross-origin", "strict-origin-when-cross-origin"},
		{"disabled", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := Headers(Config{ReferrerPolicy: tc.value})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("Referrer-Policy")
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
