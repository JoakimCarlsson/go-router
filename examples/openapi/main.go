package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/joakimcarlsson/go-router/pkg/http/openapi"
	"github.com/joakimcarlsson/go-router/pkg/http/router"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	r := router.New()

	r.GET("/users", listUsers,
		openapi.WithTags("Users"),
		openapi.WithSummary("List all users"),
		openapi.WithDescription("Returns a list of all users in the system"),
		openapi.WithParameter("limit", "query", "integer", false, "Maximum number of users to return"),
		openapi.WithArrayResponseType("200", "Successfully retrieved users", User{}),
	)

	r.GET("/users/{id}", getUser,
		openapi.WithTags("Users"),
		openapi.WithSummary("Get a user by ID"),
		openapi.WithDescription("Returns a single user by their ID"),
		openapi.WithParameter("id", "path", "integer", true, "User ID"),
		openapi.WithResponseType("200", "User found", User{}),
		openapi.WithResponseType("404", "User not found", ErrorResponse{}),
	)

	r.POST("/users", createUser,
		openapi.WithTags("Users"),
		openapi.WithSummary("Create a new user"),
		openapi.WithDescription("Creates a new user in the system"),
		openapi.WithRequestBody("User information to create", true, User{}),
		openapi.WithResponseType("201", "User created", User{}),
		openapi.WithResponseType("400", "Invalid request", ErrorResponse{}),
	)

	// Serve the OpenAPI documentation at /swagger.json
	info := openapi.Info{
		Title:       "User Management API",
		Description: "API for managing users in the system",
		Version:     "1.0.0",
		Contact: openapi.Contact{
			Name:  "API Support",
			Email: "support@example.com",
		},
	}
	r.GET("/swagger.json", r.ServeOpenAPI(info))

	// Start the server
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("OpenAPI documentation available at http://localhost:8080/swagger.json")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func listUsers(c *router.Context) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	c.JSON(200, users)
}

func getUser(c *router.Context) {
	// Get user ID from path parameter
	userID := c.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid user ID"})
		return
	}
	// In a real application, you would fetch the user from a database
	user := User{ID: id, Name: "Alice"}
	c.JSON(200, user)
}

func createUser(c *router.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid request body"})
		return
	}

	// In a real application, you would save the user to a database
	c.JSON(201, newUser)
}
