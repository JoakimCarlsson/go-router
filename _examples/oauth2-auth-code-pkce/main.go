package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// UserProfile represents user profile information
type UserProfile struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
}

// TodoItem represents a todo item
type TodoItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewTodoRequest represents a request to create a todo item
type NewTodoRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

// TokenInfo represents information from a validated token
type TokenInfo struct {
	UserID string
	Scopes []string
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func main() {
	r := router.New()

	// Public endpoints
	r.GET("/health", healthCheck,
		docs.WithTags("System"),
		docs.WithSummary("Health check endpoint"),
		docs.WithDescription("Check if the API is healthy"),
		docs.WithResponse(http.StatusOK, "API is healthy"),
	)

	// Auth required endpoints
	r.Group("/api", func(authRouter *router.Router) {
		authRouter.Use(authMiddleware)

		// Profile endpoints
		authRouter.GET("/profile", getProfile,
			docs.WithTags("Profile"),
			docs.WithSummary("Get user profile"),
			docs.WithDescription("Returns the authenticated user's profile information"),
			docs.WithResponse(http.StatusOK, "Profile retrieved successfully"),
			docs.WithJSONResponse[UserProfile](http.StatusOK, "User profile"),
			docs.WithResponse(http.StatusUnauthorized, "Unauthorized"),
			docs.WithJSONResponse[ErrorResponse](http.StatusUnauthorized, "Authentication error"),
			docs.WithResponse(http.StatusForbidden, "Forbidden - insufficient permissions"),
			docs.WithJSONResponse[ErrorResponse](http.StatusForbidden, "Missing required scope"),
			docs.WithOAuth2Scopes("profile:read"),
		)

		// Todo endpoints
		authRouter.GET("/todos", listTodos,
			docs.WithTags("Todos"),
			docs.WithSummary("List todos"),
			docs.WithDescription("Returns the authenticated user's todo items"),
			docs.WithResponse(http.StatusOK, "Todos retrieved successfully"),
			docs.WithJSONResponse[[]TodoItem](http.StatusOK, "Todo items"),
			docs.WithResponse(http.StatusUnauthorized, "Unauthorized"),
			docs.WithJSONResponse[ErrorResponse](http.StatusUnauthorized, "Authentication error"),
			docs.WithResponse(http.StatusForbidden, "Forbidden - insufficient permissions"),
			docs.WithJSONResponse[ErrorResponse](http.StatusForbidden, "Missing required scope"),
			docs.WithOAuth2Scopes("todos:read"),
		)

		authRouter.POST("/todos", createTodo,
			docs.WithTags("Todos"),
			docs.WithSummary("Create todo"),
			docs.WithDescription("Creates a new todo item for the authenticated user"),
			docs.WithJSONRequestBody[NewTodoRequest](true, "Todo details"),
			docs.WithResponse(http.StatusCreated, "Todo created successfully"),
			docs.WithJSONResponse[TodoItem](http.StatusCreated, "Created todo item"),
			docs.WithResponse(http.StatusBadRequest, "Invalid request"),
			docs.WithJSONResponse[ErrorResponse](http.StatusBadRequest, "Invalid request details"),
			docs.WithResponse(http.StatusUnauthorized, "Unauthorized"),
			docs.WithJSONResponse[ErrorResponse](http.StatusUnauthorized, "Authentication error"),
			docs.WithResponse(http.StatusForbidden, "Forbidden - insufficient permissions"),
			docs.WithJSONResponse[ErrorResponse](http.StatusForbidden, "Missing required scope"),
			docs.WithOAuth2Scopes("todos:write"),
		)

		authRouter.GET("/todos/{id}", getTodo,
			docs.WithTags("Todos"),
			docs.WithSummary("Get todo"),
			docs.WithDescription("Returns a specific todo item by ID"),
			docs.WithPathParam("id", "string", true, "Todo item ID", nil),
			docs.WithResponse(http.StatusOK, "Todo retrieved successfully"),
			docs.WithJSONResponse[TodoItem](http.StatusOK, "Todo item"),
			docs.WithResponse(http.StatusUnauthorized, "Unauthorized"),
			docs.WithJSONResponse[ErrorResponse](http.StatusUnauthorized, "Authentication error"),
			docs.WithResponse(http.StatusForbidden, "Forbidden - insufficient permissions"),
			docs.WithJSONResponse[ErrorResponse](http.StatusForbidden, "Missing required scope"),
			docs.WithResponse(http.StatusNotFound, "Todo not found"),
			docs.WithJSONResponse[ErrorResponse](http.StatusNotFound, "Todo not found details"),
			docs.WithOAuth2Scopes("todos:read"),
		)
	})

	// Configure OpenAPI documentation
	generator := openapi.NewGenerator(openapi.Info{
		Title:       "Todo API with OAuth2 Authorization Code & PKCE",
		Version:     "1.0.0",
		Description: "API demonstrating OAuth2 Authorization Code Flow with PKCE for secure authentication",
	})

	// Configure OAuth2 Authorization Code flow with PKCE
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

	// Configure OAuth2 for Swagger UI
	oauth2Config := metadata.NewOAuth2Config().
		WithClientID("your-client-id").
		WithScopes("profile:read", "todos:read", "todos:write").
		WithAppName("Todo App").
		// PKCE is enabled by default with NewOAuth2Config
		WithPKCE(true)

	// Configure Swagger UI
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.OAuth2Config = oauth2Config
	uiConfig.Title = "Todo API Documentation"
	uiConfig.TryItOutEnabled = true
	uiConfig.PersistAuthorization = true

	// Set up Swagger UI integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Middleware for authentication and authorization
func authMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		// Extract the bearer token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Authorization header is missing",
				Error:   "unauthorized",
			})
			return
		}

		// Check format (should be "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid authorization format, expected 'Bearer <token>'",
				Error:   "invalid_format",
			})
			return
		}

		// Get the token
		token := parts[1]

		// Validate token (in a real app, you'd verify with your auth server)
		tokenInfo, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: err.Error(),
				Error:   "invalid_token",
			})
			return
		}

		// Store token info in the context for handlers to access
		c.Set("tokenInfo", tokenInfo)

		// Continue to the next middleware or handler
		next(c)
	}
}

// In a real application, this would validate the token with your auth server
func validateToken(token string) (*TokenInfo, error) {
	// This is just a mock implementation for the example
	// In a real app, you would:
	// 1. Validate the JWT signature
	// 2. Check expiration and other claims
	// 3. Extract user information and scopes

	if token == "" {
		return nil, errors.New("empty token")
	}

	// For demo purposes, accept any non-empty token
	// and assign some default scopes
	return &TokenInfo{
		UserID: "user-123",
		Scopes: []string{"profile:read", "todos:read", "todos:write"},
	}, nil
}

// Check if the token has the required scope
func hasScope(c *router.Context, requiredScope string) bool {
	tokenInfoAny, exists := c.Get("tokenInfo")
	if !exists {
		return false
	}

	tokenInfo, ok := tokenInfoAny.(*TokenInfo)
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

// Handler implementations
func healthCheck(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func getProfile(c *router.Context) {
	// Check if the token has the required scope
	if (!hasScope(c, "profile:read")) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Status:  http.StatusForbidden,
			Message: "Missing required scope: profile:read",
			Error:   "insufficient_scope",
		})
		return
	}

	// In a real app, you'd fetch the user's profile from a database
	// using the user ID from the token
	tokenInfo, _ := c.Get("tokenInfo")
	userID := tokenInfo.(*TokenInfo).UserID

	profile := UserProfile{
		ID:       userID,
		Username: "johndoe",
		Email:    "john.doe@example.com",
		Name:     "John Doe",
		Roles:    []string{"user"},
	}

	c.JSON(http.StatusOK, profile)
}

func listTodos(c *router.Context) {
	// Check if the token has the required scope
	if (!hasScope(c, "todos:read")) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Status:  http.StatusForbidden,
			Message: "Missing required scope: todos:read",
			Error:   "insufficient_scope",
		})
		return
	}

	// In a real app, you'd fetch the user's todos from a database
	// using the user ID from the token
	todos := []TodoItem{
		{
			ID:          "todo-1",
			Title:       "Implement OAuth2",
			Description: "Add OAuth2 authorization code flow with PKCE",
			Completed:   true,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
		},
		{
			ID:          "todo-2",
			Title:       "Write tests",
			Description: "Add unit and integration tests",
			Completed:   false,
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, todos)
}

func createTodo(c *router.Context) {
	// Check if the token has the required scope
	if (!hasScope(c, "todos:write")) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Status:  http.StatusForbidden,
			Message: "Missing required scope: todos:write",
			Error:   "insufficient_scope",
		})
		return
	}

	// Parse request body
	var request NewTodoRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
			Error:   "invalid_request",
		})
		return
	}

	// Validate request (simple validation)
	if request.Title == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Title is required",
			Error:   "invalid_request",
		})
		return
	}

	// In a real app, you'd save the todo to a database
	// and return the created item
	now := time.Now()
	todo := TodoItem{
		ID:          fmt.Sprintf("todo-%d", time.Now().Unix()),
		Title:       request.Title,
		Description: request.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	c.JSON(http.StatusCreated, todo)
}

func getTodo(c *router.Context) {
	// Check if the token has the required scope
	if (!hasScope(c, "todos:read")) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Status:  http.StatusForbidden,
			Message: "Missing required scope: todos:read",
			Error:   "insufficient_scope",
		})
		return
	}

	// Get the todo ID from the URL parameter
	id := c.Param("id")

	// In a real app, you'd fetch the todo from a database
	// For this example, we'll just return a mock todo
	if id == "todo-1" {
		todo := TodoItem{
			ID:          "todo-1",
			Title:       "Implement OAuth2",
			Description: "Add OAuth2 authorization code flow with PKCE",
			Completed:   true,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
		}
		c.JSON(http.StatusOK, todo)
		return
	}

	// Return a 404 if not found
	c.JSON(http.StatusNotFound, ErrorResponse{
		Status:  http.StatusNotFound,
		Message: fmt.Sprintf("Todo with ID '%s' not found", id),
		Error:   "not_found",
	})
}
