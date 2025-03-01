package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

type UserProfile struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"isActive"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Simple middleware that checks for OAuth2 bearer token and extracts scopes
func oauthMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized - missing token"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized - invalid token format"})
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized - empty token"})
			return
		}

		c.Set(userIDKey, "user_123")
		next(c)
	}
}

func main() {
	r := router.New()

	// Set up OpenAPI generator with basic info
	info := openapi.Info{
		Title:       "OAuth2 API Example",
		Description: "Example API demonstrating OAuth2 integration with go-router",
		Version:     "1.0.0",
		Contact: openapi.Contact{
			Name:  "API Support",
			Email: "support@example.com",
		},
	}
	generator := openapi.NewGenerator(info)

	// Add server information
	generator.WithServer("http://localhost:8080", "Local development")

	// Add OAuth2 security schemes
	// 1. Implicit flow
	generator.WithOAuth2ImplicitFlow(
		"oauth2-implicit",
		"OAuth2 Implicit Flow",
		"https://auth.example.com/oauth2/authorize",
		map[string]string{
			"read:profile":  "Read user profile",
			"write:profile": "Update user profile",
			"admin":         "Admin access",
		},
	)

	// 2. Authorization Code flow
	generator.WithOAuth2AuthorizationCodeFlow(
		"oauth2-authcode",
		"OAuth2 Authorization Code Flow",
		"https://auth.example.com/oauth2/authorize",
		"https://auth.example.com/oauth2/token",
		map[string]string{
			"read:profile":  "Read user profile",
			"write:profile": "Update user profile",
			"admin":         "Admin access",
		},
	)

	// 3. Client Credentials flow
	generator.WithOAuth2ClientCredentialsFlow(
		"oauth2-client-credentials",
		"OAuth2 Client Credentials Flow",
		"https://auth.example.com/oauth2/token",
		map[string]string{
			"api:read":  "API read access",
			"api:write": "API write access",
		},
	)

	// Set up API routes with OAuth2 security
	r.Group("/api", func(api *router.Router) {

		api.Use(oauthMiddleware)
		api.GET("/profile", getProfile,
			openapi.WithOperationID("getProfile"),
			openapi.WithSummary("Get user profile"),
			openapi.WithDescription("Returns the authenticated user's profile information"),
			openapi.WithResponseType(http.StatusOK, "User profile", UserProfile{}),
			openapi.WithResponseType(http.StatusUnauthorized, "Unauthorized", ErrorResponse{}),
			openapi.WithSecurity(map[string][]string{"oauth2-authcode": {"read:profile"}}),
		)

		api.PUT("/profile", updateProfile,
			openapi.WithOperationID("updateProfile"),
			openapi.WithSummary("Update user profile"),
			openapi.WithDescription("Updates the authenticated user's profile information"),
			openapi.WithRequestBody("Updated profile information", true, UserProfile{}),
			openapi.WithResponseType(http.StatusOK, "Updated profile", UserProfile{}),
			openapi.WithResponseType(http.StatusUnauthorized, "Unauthorized", ErrorResponse{}),
			openapi.WithSecurity(map[string][]string{"oauth2-authcode": {"write:profile"}}),
		)

		api.GET("/admin/users", listUsers,
			openapi.WithOperationID("listUsers"),
			openapi.WithSummary("List all users"),
			openapi.WithDescription("Returns a list of all users in the system (admin only)"),
			openapi.WithResponseType(http.StatusOK, "List of users", []UserProfile{}),
			openapi.WithResponseType(http.StatusUnauthorized, "Unauthorized", ErrorResponse{}),
			openapi.WithResponseType(http.StatusForbidden, "Forbidden", ErrorResponse{}),
			openapi.WithSecurity(map[string][]string{"oauth2-authcode": {"admin"}}),
		)

		api.GET("/stats", getStats,
			openapi.WithOperationID("getStats"),
			openapi.WithSummary("Get API statistics"),
			openapi.WithDescription("Returns statistics about API usage (service accounts only)"),
			openapi.WithResponseType(http.StatusOK, "API statistics", map[string]interface{}{}),
			openapi.WithResponseType(http.StatusUnauthorized, "Unauthorized", ErrorResponse{}),
			openapi.WithSecurity(map[string][]string{"oauth2-client-credentials": {"api:read"}}),
		)
	})

	r.GET("/health", healthCheck,
		openapi.WithOperationID("healthCheck"),
		openapi.WithSummary("Health check"),
		openapi.WithDescription("Check if the API is functioning properly"),
		openapi.WithResponseType(http.StatusOK, "Health status", map[string]interface{}{}),
	)

	r.GET("/openapi.json", r.ServeOpenAPI(generator))

	swaggerConfig := router.DefaultSwaggerUIConfig()

	swaggerConfig.OAuth2Config = router.NewOAuth2Config("my-client-id")
	swaggerConfig.OAuth2Config.WithAppName("Go Router Example")
	swaggerConfig.OAuth2Config.WithScopes("read:profile", "write:profile")
	swaggerConfig.OAuth2Config.WithAdditionalQueryParam("audience", "https://api.example.com")
	swaggerConfig.OAuth2Config.WithPKCE(true)

	r.GET("/docs", r.ServeSwaggerUI(swaggerConfig))

	fmt.Println("Server started on http://localhost:8080")
	fmt.Println("API Documentation available at:")
	fmt.Println("  - OpenAPI spec: http://localhost:8080/openapi.json")
	fmt.Println("  - Swagger UI: http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// API handlers
func getProfile(c *router.Context) {
	userID, _ := c.GetString(userIDKey)
	profile := UserProfile{
		ID:       userID,
		Name:     "John Doe",
		Email:    "john@example.com",
		Roles:    []string{"user"},
		IsActive: true,
	}
	c.JSON(http.StatusOK, profile)
}

func updateProfile(c *router.Context) {
	var profile UserProfile
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	userID, _ := c.GetString(userIDKey)
	profile.ID = userID // Ensure ID is not changed

	c.JSON(http.StatusOK, profile)
}

func listUsers(c *router.Context) {
	users := []UserProfile{
		{ID: "user_123", Name: "John Doe", Email: "john@example.com", Roles: []string{"user"}, IsActive: true},
		{ID: "user_456", Name: "Jane Smith", Email: "jane@example.com", Roles: []string{"admin"}, IsActive: true},
		{ID: "user_789", Name: "Bob Johnson", Email: "bob@example.com", Roles: []string{"user"}, IsActive: false},
	}
	c.JSON(http.StatusOK, users)
}

func getStats(c *router.Context) {
	stats := map[string]interface{}{
		"totalRequests": 12345,
		"activeUsers":   42,
		"apiVersion":    "1.0.0",
		"serverTime":    "2023-05-10T15:04:05Z",
	}
	c.JSON(http.StatusOK, stats)
}

func healthCheck(c *router.Context) {
	status := map[string]interface{}{
		"status":  "UP",
		"version": "1.0.0",
		"uptime":  "2d 4h 32m",
	}
	c.JSON(http.StatusOK, status)
}
