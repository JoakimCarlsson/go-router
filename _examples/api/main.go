package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/router"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id" validate:"required"`
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=0,max=150"`
}

func main() {
	r := router.New()

	r.GET("/users", listUsers,
		docs.WithTags("Users"),
		docs.WithSummary("List all users"),
		docs.WithDescription("Returns a list of all users in the system"),
		docs.WithResponse(200, "List of users retrieved successfully"),
		docs.WithJSONResponse[[]User](200, "List of users"),
	)

	r.POST("/users", createUser,
		docs.WithTags("Users"),
		docs.WithSummary("Create a new user"),
		docs.WithDescription("Creates a new user in the system"),
		docs.WithJSONRequestBody[User](true, "User information"),
		docs.WithResponse(201, "User created successfully"),
		docs.WithResponse(400, "Invalid user data"),
		docs.WithBearerAuth(),
	)

	r.GET("/users/{id}", getUser,
		docs.WithTags("Users"),
		docs.WithSummary("Get user by ID"),
		docs.WithDescription("Retrieves a user's details by their unique identifier"),
		docs.WithPathParam("id", "string", true, "The user's unique identifier", nil),
		docs.WithResponse(200, "User found"),
		docs.WithJSONResponse[User](200, "User details"),
		docs.WithResponse(404, "User not found"),
		docs.WithBearerAuth(),
	)

	err := integration.Setup(r, integration.SetupOptions{
		Title:         "User Management API",
		Version:       "1.0.0",
		Description:   "An example API demonstrating the refactored router components",
		UseBearerAuth: true,
	})
	if err != nil {
		log.Fatal("Failed to set up API documentation:", err)
	}

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handler implementations
func listUsers(c *router.Context) {
	users := []User{
		{ID: "1", Username: "john_doe", Email: "john@example.com", Age: 30},
		{ID: "2", Username: "jane_doe", Email: "jane@example.com", Age: 28},
	}
	c.JSON(http.StatusOK, users)
}

func createUser(c *router.Context) {
	var user User
	if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	// In a real application, validate and save the user
	c.JSON(http.StatusCreated, user)
}

func getUser(c *router.Context) {
	id := c.Param("id")
	// In a real application, fetch the user from a database
	user := User{
		ID:       id,
		Username: "john_doe",
		Email:    "john@example.com",
		Age:      30,
	}
	c.JSON(http.StatusOK, user)
}
