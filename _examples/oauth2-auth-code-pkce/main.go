package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

type UserProfile struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
}

type TodoItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type NewTodoRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type TokenInfo struct {
	UserID string
	Scopes []string
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type contextKey string

const tokenInfoKey contextKey = "tokenInfo"

func main() {
	r := router.New()

	r.GET("/health", healthCheck,
		openapi.WithTags("System"),
		openapi.WithSummary("Health check endpoint"),
		openapi.WithResponse(http.StatusOK, "API is healthy"),
	)

	r.Group("/api", func(authRouter *router.Router) {
		authRouter.Use(authMiddleware)

		authRouter.GET("/profile", getProfile,
			openapi.WithTags("Profile"),
			openapi.WithSummary("Get user profile"),
			openapi.WithJSONResponse[UserProfile](http.StatusOK, "User profile"),
			openapi.WithOAuth2Scopes("profile:read"),
		)

		authRouter.GET("/todos", listTodos,
			openapi.WithTags("Todos"),
			openapi.WithSummary("List todos"),
			openapi.WithJSONResponse[[]TodoItem](http.StatusOK, "Todo items"),
			openapi.WithOAuth2Scopes("todos:read"),
		)

		authRouter.POST("/todos", createTodo,
			openapi.WithTags("Todos"),
			openapi.WithSummary("Create todo"),
			openapi.WithJSONRequestBody[NewTodoRequest](true, "Todo details"),
			openapi.WithJSONResponse[TodoItem](http.StatusCreated, "Created todo item"),
			openapi.WithOAuth2Scopes("todos:write"),
		)

		authRouter.GET("/todos/{id}", getTodo,
			openapi.WithTags("Todos"),
			openapi.WithSummary("Get todo"),
			openapi.WithPathParam("id", "string", true, "Todo item ID", nil),
			openapi.WithJSONResponse[TodoItem](http.StatusOK, "Todo item"),
			openapi.WithOAuth2Scopes("todos:read"),
		)
	})

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Todo API with OAuth2 Authorization Code & PKCE",
		Version:     "1.0.0",
		Description: "API demonstrating OAuth2 Authorization Code Flow with PKCE",
	})

	generator.WithOAuth2AuthorizationCodeFlow(
		"oauth2",
		"OAuth2 Authorization",
		"https://your-auth-server.com/authorize",
		"https://your-auth-server.com/token",
		map[string]string{
			"profile:read": "Read your profile information",
			"todos:read":   "Read your todo items",
			"todos:write":  "Create and edit your todo items",
		},
	)

	uiConfig := swaggerui.DefaultUIConfig()
	uiConfig.Title = "Todo API Documentation"
	uiConfig.TryItOutEnabled = true
	uiConfig.OAuth2Config = &swaggerui.OAuth2Config{
		ClientID:                          "your-client-id",
		AppName:                           "Todo App",
		Scopes:                            `"profile:read todos:read todos:write"`,
		UsePkceWithAuthorizationCodeGrant: true,
	}

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(uiConfig)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Status: http.StatusUnauthorized, Message: "Authorization header required", Error: "unauthorized"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid authorization format", Error: "invalid_format"})
			return
		}

		tokenInfo, err := validateToken(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Status: http.StatusUnauthorized, Message: err.Error(), Error: "invalid_token"})
			return
		}

		ctx := context.WithValue(r.Context(), tokenInfoKey, tokenInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateToken(token string) (*TokenInfo, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	return &TokenInfo{UserID: "user-123", Scopes: []string{"profile:read", "todos:read", "todos:write"}}, nil
}

func tokenInfoFromContext(ctx context.Context) (*TokenInfo, bool) {
	info, ok := ctx.Value(tokenInfoKey).(*TokenInfo)
	return info, ok
}

func hasScope(c *router.Context, requiredScope string) bool {
	tokenInfo, ok := tokenInfoFromContext(c.Request.Context())
	if !ok {
		return false
	}
	for _, scope := range tokenInfo.Scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "healthy", "time": time.Now().Format(time.RFC3339)})
}

func getProfile(c *router.Context) {
	if !hasScope(c, "profile:read") {
		c.JSON(http.StatusForbidden, ErrorResponse{Status: http.StatusForbidden, Message: "Missing scope: profile:read", Error: "insufficient_scope"})
		return
	}
	tokenInfo, _ := tokenInfoFromContext(c.Request.Context())
	c.JSON(http.StatusOK, UserProfile{ID: tokenInfo.UserID, Username: "johndoe", Email: "john@example.com", Name: "John Doe", Roles: []string{"user"}})
}

func listTodos(c *router.Context) {
	if !hasScope(c, "todos:read") {
		c.JSON(http.StatusForbidden, ErrorResponse{Status: http.StatusForbidden, Message: "Missing scope: todos:read", Error: "insufficient_scope"})
		return
	}
	c.JSON(http.StatusOK, []TodoItem{
		{ID: "todo-1", Title: "Implement OAuth2", Description: "Add OAuth2 flow", Completed: true, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		{ID: "todo-2", Title: "Write tests", Description: "Add tests", Completed: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	})
}

func createTodo(c *router.Context) {
	if !hasScope(c, "todos:write") {
		c.JSON(http.StatusForbidden, ErrorResponse{Status: http.StatusForbidden, Message: "Missing scope: todos:write", Error: "insufficient_scope"})
		return
	}
	var request NewTodoRequest
	if err := c.BindJSON(&request); err != nil || request.Title == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid request", Error: "invalid_request"})
		return
	}
	now := time.Now()
	c.JSON(http.StatusCreated, TodoItem{ID: fmt.Sprintf("todo-%d", now.Unix()), Title: request.Title, Description: request.Description, CreatedAt: now, UpdatedAt: now})
}

func getTodo(c *router.Context) {
	if !hasScope(c, "todos:read") {
		c.JSON(http.StatusForbidden, ErrorResponse{Status: http.StatusForbidden, Message: "Missing scope: todos:read", Error: "insufficient_scope"})
		return
	}
	id := c.Param("id")
	if id == "todo-1" {
		c.JSON(http.StatusOK, TodoItem{ID: "todo-1", Title: "Implement OAuth2", Description: "Add OAuth2 flow", Completed: true, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()})
		return
	}
	c.JSON(http.StatusNotFound, ErrorResponse{Status: http.StatusNotFound, Message: fmt.Sprintf("Todo '%s' not found", id), Error: "not_found"})
}
